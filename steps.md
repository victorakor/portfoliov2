# Railway Setup Guide — Victor Akor Portfolio v3

Complete step-by-step instructions to get the portfolio live on Railway.

---

## Step 1: Create a Railway Account

1. Go to [https://railway.app](https://railway.app)
2. Sign up with your GitHub account
3. Authorize Railway to access your repositories

---

## Step 2: Create a New Project

1. Click **"New Project"** on the Railway dashboard
2. Select **"Deploy from GitHub repo"**
3. Find and select your portfolio repository
4. Railway will detect the `Dockerfile` automatically

---

## Step 3: Add a PostgreSQL Database

1. Inside your Railway project, click **"+ New"**
2. Select **"Database"** → **"Add PostgreSQL"**
3. Railway will provision a PostgreSQL instance and add it to your project
4. Click on the PostgreSQL service → go to the **"Variables"** tab
5. Copy the value of **`DATABASE_URL`** — you will need it in Step 4

---

## Step 4: Set Environment Variables

Go to your **web service** (not the database) → click **"Variables"** tab → add each variable below.

> ⚠️ Replace every placeholder value with your real credentials before deploying.

---

### Required — Server

| Variable | Value | Notes |
|----------|-------|-------|
| `ENV` | `production` | Change from `development` to `production` |
| `PORT` | `8080` | Railway sets this automatically — you can skip it |

---

### Required — Database

You have two options. Use **Option A** (recommended for Railway):

**Option A — Single connection string (recommended)**

| Variable | Value |
|----------|-------|
| `DATABASE_URL` | Paste the `DATABASE_URL` you copied from the PostgreSQL service in Step 3 |

**Option B — Individual variables (use if not using DATABASE_URL)**

| Variable | Value | Notes |
|----------|-------|-------|
| `DB_HOST` | Get from PostgreSQL service Variables tab | e.g. `monorail.proxy.rlwy.net` |
| `DB_PORT` | Get from PostgreSQL service Variables tab | e.g. `12345` |
| `DB_USER` | Get from PostgreSQL service Variables tab | e.g. `postgres` |
| `DB_PASSWORD` | Get from PostgreSQL service Variables tab | Your actual DB password |
| `DB_NAME` | Get from PostgreSQL service Variables tab | e.g. `railway` |
| `DB_SSLMODE` | `require` | Must be `require` on Railway (not `disable`) |

> ℹ️ If you set `DATABASE_URL`, the app will use it and ignore the individual `DB_*` variables.

---

### Required — Authentication

| Variable | Value | Notes |
|----------|-------|-------|
| `JWT_SECRET` | Generate a strong random string | Must be at least 32 characters. Example: `openssl rand -hex 32` in your terminal |

**To generate a secure JWT secret, run this in your terminal:**
```bash
openssl rand -hex 32
```
Copy the output and paste it as the value.

---

### Required — Domain

| Variable | Value | Notes |
|----------|-------|-------|
| `ALLOWED_ORIGIN` | `https://your-app.up.railway.app` | Your Railway domain. Find it under Settings → Networking → Public Domain |

Once you add a custom domain, update this to `https://yourdomain.com`.

---

### Optional — Email (Resend)

Used for sending email notifications when a new lead submits the contact form.

1. Go to [https://resend.com](https://resend.com) and create a free account
2. Go to **API Keys** → **Create API Key**
3. Copy the key

| Variable | Value |
|----------|-------|
| `RESEND_API_KEY` | `re_xxxxxxxxxxxxxxxxxxxx` |

---

### Optional — Image Storage (Cloudinary)

Used for uploading blog post cover images and project images via the admin panel.

1. Go to [https://cloudinary.com](https://cloudinary.com) and create a free account
2. Go to **Dashboard** → copy your **Cloud Name**, **API Key**, and **API Secret**
3. Format the URL as: `cloudinary://API_KEY:API_SECRET@CLOUD_NAME`

| Variable | Value |
|----------|-------|
| `CLOUDINARY_URL` | `cloudinary://123456789:abcdefghijk@yourcloudname` |

---

### Optional — Analytics (Google Analytics)

1. Go to [https://analytics.google.com](https://analytics.google.com)
2. Create a new property for your portfolio
3. Copy the **Measurement ID** (format: `G-XXXXXXXXXX`)

| Variable | Value |
|----------|-------|
| `GA_MEASUREMENT_ID` | `G-XXXXXXXXXX` |

---

## Step 5: Run the Database Migration

After your first deploy, you need to create the database tables.

1. In Railway, click on your **PostgreSQL** service
2. Go to the **"Query"** tab (or connect via a database client)
3. Copy the entire contents of `migrations/001_init.sql` from your repo
4. Paste and run it in the query editor

**Or connect from your terminal:**
```bash
# Get the connection string from Railway PostgreSQL → Variables → DATABASE_URL
psql "your_DATABASE_URL_here" -f migrations/001_init.sql
```

---

## Step 6: Create Your Admin User

You need to create the first admin account manually. Run this in the Railway PostgreSQL query tab:

```sql
-- First generate a bcrypt hash of your password
-- Use this site: https://bcrypt-generator.com (rounds: 10)
-- OR run in your terminal: htpasswd -bnBC 10 "" yourpassword | tr -d ':\n'

INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES (
  uuid_generate_v4(),
  'victorakor04@gmail.com',
  '$2y$10$BSlMuato1ITok0puPurr0OeM8FvWGYtHOe5dl5.DGNP3WvQ7i/6uO',
  'admin',
  NOW(),
  NOW()
);
```

**To generate a bcrypt hash from your terminal:**
```bash
# Install htpasswd if needed: sudo apt install apache2-utils
htpasswd -bnBC 10 "" yourpassword | tr -d ':\n' | sed 's/^://'
```

---

## Step 7: Set Your Public Domain

1. In Railway, go to your web service → **Settings** → **Networking**
2. Click **"Generate Domain"** to get a free `*.up.railway.app` domain
3. Or click **"Custom Domain"** to add your own domain and follow the DNS instructions
4. Update the `ALLOWED_ORIGIN` environment variable to match your domain

---

## Step 8: Verify the Deployment

Once deployed, check these URLs:

| URL | Expected |
|-----|----------|
| `https://yourdomain.com` | Portfolio homepage |
| `https://yourdomain.com/about` | About page |
| `https://yourdomain.com/contact` | 6-step contact form |
| `https://yourdomain.com/admin/login` | Admin login page |
| `https://yourdomain.com/admin` | Admin dashboard (after login) |

---

## Summary Checklist

- [ ] Railway account created
- [ ] Project created from GitHub repo
- [ ] PostgreSQL database added
- [ ] `DATABASE_URL` or individual `DB_*` variables set
- [ ] `ENV` set to `production`
- [ ] `JWT_SECRET` set to a strong 32+ character random string
- [ ] `ALLOWED_ORIGIN` set to your Railway domain
- [ ] `DB_SSLMODE` set to `require` (not `disable`)
- [ ] Migration SQL run on the database
- [ ] Admin user created in the database
- [ ] Site loads at your public domain
- [ ] Admin login works at `/admin/login`

---

## Troubleshooting

**Build fails with "directory not found"**
→ Make sure all files are committed and pushed to GitHub. Run `git status` to check.

**App crashes on start**
→ Check Railway logs. Most likely a missing environment variable or database connection issue.

**Database connection refused**
→ Make sure `DB_SSLMODE=require` (not `disable`) when connecting to Railway PostgreSQL.

**Admin login returns 401**
→ The admin user hasn't been created yet. Follow Step 6.

**CORS errors in browser**
→ Update `ALLOWED_ORIGIN` to exactly match your domain including `https://`.
