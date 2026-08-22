#!/usr/bin/env bash
# Victor Akor Portfolio v3 — one-shot provisioner for a fresh Oracle Cloud
# Always Free VM (Ubuntu 24.04, aarch64).
#
#   curl -fsSL https://raw.githubusercontent.com/victorakor/portfoliov2/main/deploy/bootstrap.sh | bash
#
# Or, from a clone:  bash deploy/bootstrap.sh
#
# Idempotent — safe to re-run. Secrets are generated once and preserved in .env
# on re-runs, so an existing database keeps working.

set -euo pipefail

REPO="${REPO:-https://github.com/victorakor/portfoliov2.git}"
APP_DIR="${APP_DIR:-$HOME/portfolio}"

log() { printf '\n\033[1;36m==> %s\033[0m\n' "$1"; }

# ---------------------------------------------------------------- 1. packages
log "Installing Docker and tools"
sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq git curl openssl >/dev/null

if ! command -v docker >/dev/null 2>&1; then
	curl -fsSL https://get.docker.com | sudo sh >/dev/null
	sudo usermod -aG docker "$USER"
fi

# ---------------------------------------------------------------- 2. firewall
# Oracle images ship an iptables ruleset that REJECTs everything except SSH.
# Opening the VCN security list alone is not enough. Inserting at the head of
# the chain works regardless of how the default rules are ordered.
log "Opening ports 80 and 443 on the host firewall"
for port in 80 443; do
	if ! sudo iptables -C INPUT -p tcp --dport "$port" -j ACCEPT 2>/dev/null; then
		sudo iptables -I INPUT 1 -p tcp --dport "$port" -j ACCEPT
	fi
done

echo iptables-persistent iptables-persistent/autosave_v4 boolean true | sudo debconf-set-selections
echo iptables-persistent iptables-persistent/autosave_v6 boolean true | sudo debconf-set-selections
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq iptables-persistent >/dev/null 2>&1 || true
sudo netfilter-persistent save >/dev/null 2>&1 || true

# ---------------------------------------------------------------- 3. source
log "Fetching application source"
if [ -d "$APP_DIR/.git" ]; then
	git -C "$APP_DIR" pull --ff-only
else
	git clone --depth 1 "$REPO" "$APP_DIR"
fi
cd "$APP_DIR"

# ---------------------------------------------------------------- 4. config
# sslip.io resolves <ip>.sslip.io to <ip>, which gives Caddy a real resolvable
# name to get a Let's Encrypt certificate for — no domain purchase needed.
# Override with:  DOMAIN=yourdomain.com bash deploy/bootstrap.sh
PUBLIC_IP="$(curl -fsS https://api.ipify.org)"
DOMAIN="${DOMAIN:-${PUBLIC_IP}.sslip.io}"

if [ -f .env ]; then
	log "Reusing existing .env (secrets preserved)"
	# shellcheck disable=SC1091
	set -a; . ./.env; set +a
else
	log "Generating .env with fresh secrets"
	cat > .env <<EOF
DOMAIN=${DOMAIN}
DB_PASSWORD=$(openssl rand -hex 24)
JWT_SECRET=$(openssl rand -hex 48)

# Fill these in later, then: docker compose up -d
RESEND_API_KEY=
CLOUDINARY_URL=
GA_MEASUREMENT_ID=
OLLAMA_API_KEY=
OLLAMA_MODEL=
EOF
	chmod 600 .env
fi

# The www redirect in the Caddyfile cannot get a certificate for an sslip.io
# host, and a failing cert blocks startup. Drop it for IP-based deploys.
if [[ "$DOMAIN" == *.sslip.io ]] && grep -q 'www.{$DOMAIN}' Caddyfile; then
	log "Removing www redirect (not valid for sslip.io)"
	sed -i '/^# Send the apex/,$d' Caddyfile
fi

# ---------------------------------------------------------------- 5. start
log "Building and starting the stack (first build takes a few minutes)"
sudo docker compose up -d --build

log "Waiting for the app to become healthy"
for _ in $(seq 1 60); do
	if sudo docker compose logs app 2>/dev/null | grep -q "Server starting"; then
		break
	fi
	sleep 3
done

# ---------------------------------------------------------------- 6. admin
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@${DOMAIN}}"
EXISTING="$(sudo docker compose exec -T postgres \
	psql -U portfolio -d portfolio -tAc "SELECT count(*) FROM users;" 2>/dev/null || echo 0)"

if [ "${EXISTING//[[:space:]]/}" = "0" ]; then
	ADMIN_PASS="${ADMIN_PASS:-$(openssl rand -base64 18)}"
	log "Seeding admin user"
	HASH="$(sudo docker run --rm -v "$PWD":/src -w /src golang:1.25-alpine \
		go run ./cmd/hashpw "$ADMIN_PASS" 2>/dev/null | tail -1)"
	sudo docker compose exec -T postgres psql -U portfolio -d portfolio -c \
		"INSERT INTO users (email, password_hash, role) VALUES ('${ADMIN_EMAIL}', '${HASH}', 'admin');"
	CREDS="  email:    ${ADMIN_EMAIL}
  password: ${ADMIN_PASS}"
else
	CREDS="  (admin user already exists — credentials unchanged)"
fi

# ---------------------------------------------------------------- done
cat <<EOF

$(printf '\033[1;32m')Deployment complete.$(printf '\033[0m')

  Site:   https://${DOMAIN}
  Admin:  https://${DOMAIN}/admin/login

${CREDS}

Save that password now — it is not stored anywhere in plaintext.

Logs:     cd ${APP_DIR} && docker compose logs -f
Restart:  cd ${APP_DIR} && docker compose restart
Update:   cd ${APP_DIR} && git pull && docker compose up -d --build

Certificates take up to ~60s to issue on first boot. If HTTPS is not up yet,
confirm ports 80/443 are also open in the VCN security list (Networking → VCN →
subnet → Security List → Add Ingress Rules, source 0.0.0.0/0).
EOF
