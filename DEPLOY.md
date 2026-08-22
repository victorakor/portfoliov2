# Deploying to an Oracle Cloud Always Free VM

One VM runs the app, Postgres, and Caddy. Nothing expires and nothing needs a
paid upgrade later — the tradeoff is that TLS, backups, and OS updates are yours.

Files involved: `docker-compose.yml`, `Caddyfile`, `.env.production.example`.
The existing `Dockerfile` is used as-is.

## Fast path

Create the VM (step 1), open ports in the VCN security list (step 2, cloud layer
only), then SSH in and run:

```bash
curl -fsSL https://raw.githubusercontent.com/victorakor/portfoliov2/main/deploy/bootstrap.sh | bash
```

`deploy/bootstrap.sh` handles everything else — Docker, host firewall, clone,
secret generation, TLS, and seeding the admin user — and prints your URL and
admin credentials at the end. It is idempotent, so re-running it to update is
safe and preserves existing secrets and data.

Without a domain it serves on `https://<your-ip>.sslip.io` with a real
Let's Encrypt certificate. Once you have a domain pointed at the VM, re-run with
`DOMAIN=yourdomain.com bash deploy/bootstrap.sh`.

The manual steps below are the same work, spelled out, if you'd rather not pipe
a script to bash.

## 1. Create the instance

In the Oracle Cloud console: **Compute → Instances → Create instance**.

- **Image:** Ubuntu 24.04 (aarch64 build)
- **Shape:** `VM.Standard.A1.Flex`, **2 OCPU / 12 GB**

The Always Free allowance is 4 OCPU and 24 GB of Ampere A1 spread across all
your instances, so 2/12 leaves headroom and is far more than this app needs.

Upload your SSH public key before launching. The login user is `ubuntu`.

> **Expect "Out of host capacity."** A1 capacity is genuinely scarce and this is
> the single most common failure here, not a mistake on your part. Retry, and try
> each availability domain in your region. If it stays unavailable, the AMD
> `VM.Standard.E2.1.Micro` is also Always Free, but at 1 GB RAM you should add a
> swap file before running the stack.

## 2. Open ports — both layers

This is the other classic Oracle trap: there are **two** independent firewalls,
and opening only one leaves the site silently unreachable.

**Cloud layer** — Networking → your VCN → the public subnet → its Security List
→ Add Ingress Rules. Source `0.0.0.0/0`, IP Protocol TCP, destination ports `80`
and `443`.

**Host layer** — Oracle's images ship iptables rules that `DROP` everything
except SSH. On the VM:

```bash
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
sudo netfilter-persistent save
```

Without `netfilter-persistent save` those rules vanish on reboot.

## 3. Point DNS at the box

Add an `A` record for your domain to the instance's public IP (and a second one
for `www` if you kept that block in the `Caddyfile`).

**Do this before step 5.** Caddy proves domain ownership to Let's Encrypt over
port 80, so the name has to already resolve to the server. Verify with
`dig +short yourdomain.com` before continuing.

## 4. Install Docker and the code

```bash
sudo apt-get update && sudo apt-get upgrade -y
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker ubuntu && newgrp docker

git clone <your-repo-url> portfolio && cd portfolio
cp .env.production.example .env
nano .env      # set DOMAIN, DB_PASSWORD, JWT_SECRET, API keys
```

Generate the two secrets with `openssl rand -base64 24` and
`openssl rand -base64 48`.

Ampere is ARM64, and both `golang:1.25-alpine` and `alpine:3.19` are multi-arch,
so the build on the VM is native. No cross-compilation or buildx setup.

## 5. Start it

```bash
docker compose up -d --build
docker compose logs -f
```

First build takes a few minutes. You want to see `Database connected
successfully`, then `Server starting on :8080`. Caddy fetches the certificate on
its own within about 30 seconds; then `https://yourdomain.com` is live.

Postgres applies `migrations/001_init.sql` automatically on first boot, via the
`docker-entrypoint-initdb.d` mount. Nothing in the Go code runs migrations, so
that mount is the only thing creating your schema — see the note in step 8.

## 6. Seed the admin user

There is no signup route, so the first admin row goes in by hand. Mint a bcrypt
hash with the `cmd/hashpw` helper — on the VM, without installing Go:

```bash
docker run --rm -v "$PWD":/src -w /src golang:1.25-alpine \
  go run ./cmd/hashpw 'your-admin-password'
```

Copy the `$2a$...` line it prints, then insert the user:

```bash
docker compose exec postgres psql -U portfolio -d portfolio -c \
  "INSERT INTO users (email, password_hash, role)
   VALUES ('you@example.com', '<paste-hash>', 'admin');"
```

Then log in at `https://yourdomain.com/admin/login`.

## 7. Back up the database

You are the DBA now — there are no managed snapshots. A nightly dump:

```bash
mkdir -p ~/backups
crontab -e
```

```cron
0 3 * * * cd ~/portfolio && docker compose exec -T postgres pg_dump -U portfolio portfolio | gzip > ~/backups/portfolio-$(date +\%F).sql.gz && find ~/backups -name '*.sql.gz' -mtime +14 -delete
```

Pull a copy off the box periodically — a backup that only exists on the VM does
not protect you from losing the VM.

## 8. Operating notes

**Updates.** `git pull && docker compose up -d --build`. Postgres and Caddy
volumes persist across rebuilds.

**Future schema changes will not apply themselves.** The initdb mount only runs
when the `pgdata` volume is first created, so migration `002_*.sql` and later
must be applied manually:

```bash
docker compose exec -T postgres psql -U portfolio -d portfolio < migrations/002_whatever.sql
```

If you'd rather not remember that, move migrations into the Go startup path so
they run on every boot — worth doing before the schema changes again.

**Idle reclamation.** Oracle can reclaim Always Free compute that sits idle for
~7 days. A site with real traffic and periodic certificate renewals normally
stays above the threshold, but it is a real policy and the reason to keep an eye
on uptime. If it worries you, an uptime monitor pinging the site every few
minutes both watches it and keeps it warm.

**Never commit `.env`.** It's already in `.gitignore`.
