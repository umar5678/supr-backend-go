# Setup Guide for Go Backend Application on a New Machine

This guide assumes you're starting with a fresh Linux server (e.g., Ubuntu 22.04 or similar). We'll cover everything from initial server access and prerequisite installations to full deployment, configuration, and monitoring. The application is a Go backend (supr-backend-go) using PostgreSQL, served via Nginx, and managed with systemd. The target domain is api.pittapizzahusrev.be, and the app listens on port 3000 internally.

## Prerequisites

- A new VPS or server with root access.
- SSH key pair (private key on your local machine).
- Domain pointed to the server's IP ( your IP).
- Basic familiarity with Linux commands.

## Step 1: Initial Server Access and Setup

### SSH into the new machine

Use your SSH key for secure access. Replace `~/.ssh/hostinger_vps_access` with your private key path and update the IP if needed.

```bash
ssh -i ~/.ssh/vps_access.key root@<your-server-ip>
```

### Update system packages

```bash
apt update && apt upgrade -y
```

### Set up firewall (UFW)

Allow SSH, HTTP, HTTPS.

```bash
apt install ufw -y
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
ufw status
```

### Install essential tools

```bash
apt install git curl nano wget unzip -y
```

## Step 2: Install Prerequisites

### Install Go

Download and install the latest Go version (e.g., 1.21 or higher).

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=/root/go' >> ~/.bashrc
source ~/.bashrc
go version
```

### Create GOPATH directory

```bash
mkdir -p /root/go/bin
```

### Install PostgreSQL

```bash
apt install postgresql postgresql-contrib -y
systemctl start postgresql
systemctl enable postgresql
systemctl status postgresql
```

### Install Nginx

```bash
apt install nginx -y
systemctl start nginx
systemctl enable nginx
systemctl status nginx
```

### Install Migrate Tool (for database migrations)

Install from source or binary.

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
mv migrate /root/go/bin/
chmod +x /root/go/bin/migrate
migrate --version
```

## Step 3: Set Up Database

### Create database user and database

Switch to postgres user and create the app's DB user (use a secure password, e.g., goPass_Secure123!).

```bash
sudo -u postgres psql
```

Inside psql:

```sql
CREATE USER go_backend_admin WITH PASSWORD 'goPass_Secure123!';
CREATE DATABASE go_backend OWNER go_backend_admin;
GRANT ALL PRIVILEGES ON DATABASE go_backend TO go_backend_admin;
\q
```

### Test database connection

```bash
psql -U go_backend_admin -d go_backend -h localhost
```

Enter password when prompted. Run `\dt` to list tables (should be empty initially). Exit with `\q`.

## Step 4: Clone Repository and Set Up Application

### Create application directory

```bash
mkdir -p /var/www/go-backend/supr-backend-go
cd /var/www/go-backend/supr-backend-go
```

### Clone the repository

Assuming it's a private repo; use your Git credentials or SSH key setup:

```bash
git clone <your-repo-url> .
```

### Set up .env file

Create or copy .env with necessary configs (e.g., DB credentials, server port).

```bash
nano .env
```

Example content (adjust as needed):

```text
DB_USER=go_backend_admin
DB_PASSWORD=goPass_Secure123!
DB_NAME=go_backend
DB_HOST=localhost
DB_PORT=5432
SERVER_PORT=3000
# Add other env vars like CORS, JWT secrets, etc.
```

### Update dependencies

```bash
go mod download
go mod tidy
```

### Run initial migrations

Ensure migrations directory exists.

```bash
ls -la migrations/
/root/go/bin/migrate -path ./migrations -database 'postgres://go_backend_admin:goPass_Secure123!@localhost:5432/go_backend?sslmode=disable' up
```

### Verify tables

```bash
psql -U go_backend_admin -d go_backend -h localhost -c "\dt"
```

### Build the application

```bash
go build -o bin/go-backend ./cmd/api
chmod +x bin/go-backend
```

### Quick test

```bash
./bin/go-backend  # Press Ctrl+C to stop after verifying it starts
```

## Step 5: Set Up Systemd Service

### Create service file

```bash
nano /etc/systemd/system/go-backend.service
```

Add:

```text
[Unit]
Description=Go Backend Service
After=network.target

[Service]
User=root
WorkingDirectory=/var/www/go-backend/supr-backend-go
ExecStart=/var/www/go-backend/supr-backend-go/bin/go-backend
Restart=always
StandardOutput=append:/var/log/go-backend/app.log
StandardError=append:/var/log/go-backend/app.log

[Install]
WantedBy=multi-user.target
```

### Create log directory

```bash
mkdir -p /var/log/go-backend
chmod 755 /var/log/go-backend
```

### Reload and enable service

```bash
systemctl daemon-reload
systemctl enable go-backend
```

## Step 6: Configure Nginx

### Create Nginx site config

```bash
nano /etc/nginx/sites-available/pizza-husrev
```

Add (adjust for your domain; this proxies to port 3000 and handles WebSockets):

```text
server {
    listen 80;
    listen 443 ssl;
    server_name api.pittapizzahusrev.be;

    # SSL certs (set up with Certbot or similar)
    # ssl_certificate /etc/letsencrypt/live/api.pittapizzahusrev.be/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/api.pittapizzahusrev.be/privkey.pem;

    location /go/ {
        proxy_pass http://localhost:3000/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }

    access_log /var/log/nginx/api_access.log;
    error_log /var/log/nginx/api_error.log;
}
```

### Enable site and test

```bash
ln -s /etc/nginx/sites-available/pizza-husrev /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

### Set up SSL (optional but recommended)

```bash
apt install certbot python3-certbot-nginx -y
certbot --nginx -d api.pittapizzahusrev.be
```

## Step 7: Start Service and Test Deployment

### Start the service

```bash
systemctl start go-backend
systemctl status go-backend
journalctl -u go-backend -f  # View real-time logs
```

### Test endpoints

```bash
curl http://localhost:3000/health
curl https://api.pittapizzahusrev.be/go/health
ss -tulpn | grep 3000  # Check port listening
tail -50 /var/log/go-backend/app.log  # View app logs
```

## Service Management

- **Restart:** `systemctl restart go-backend`
- **Stop:** `systemctl stop go-backend`
- **Start:** `systemctl start go-backend`
- **Status:** `systemctl status go-backend`
- **Logs:** `journalctl -u go-backend -f` or `journalctl -u go-backend -n 100` or `tail -f /var/log/go-backend/app.log`

## Database Operations

- **Connect:** `psql -U go_backend_admin -d go_backend -h localhost`
- **Run migrations:** `/root/go/bin/migrate -path ./migrations -database 'postgres://go_backend_admin:goPass_Secure123!@localhost:5432/go_backend?sslmode=disable' up`
- **Check version:** `/root/go/bin/migrate ... version`
- **Rollback last:** `/root/go/bin/migrate ... down 1`
- **Backup:** `pg_dump -U go_backend_admin -h localhost go_backend > /var/www/backups/db_backup_$(date +%Y%m%d).sql` (Create `/var/www/backups` first)

## Testing

- **Endpoints:** `curl https://api.pittapizzahusrev.be/go/health` and `curl https://api.pittapizzahusrev.be/go/api/health`
- **Ports:** `ss -tulpn | grep -E "3000|8080"`
- **WebSockets:** `ss -tulpn | grep 3000`
- **Logs:** `tail -f /var/log/go-backend/app.log /var/log/nginx/api_error.log`
- **External:** `curl -I https://api.pittapizzahusrev.be/go/health`

## Nginx Management

- **Test config:** `nginx -t`
- **Reload:** `systemctl reload nginx`
- **Restart:** `systemctl restart nginx`
- **Logs:** `tail -f /var/log/nginx/api_access.log` or `tail -f /var/log/nginx/api_error.log`

## Troubleshooting

### Service Won't Start

- **Logs:** `journalctl -u go-backend -xe`
- **Port check:** `ss -tulpn | grep 3000`
- **Test binary:** `cd /var/www/go-backend/supr-backend-go && ./bin/go-backend`
- **Env check:** `cat .env | grep -E "SERVER_PORT|DB_"`

### WebSocket Not Connecting

- **Listening check:** `ss -tulpn | grep 3000`
- **Nginx logs:** `tail -f /var/log/nginx/api_error.log`
- **Test upgrade:** `curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" https://api.pittapizzahusrev.be/go/ws`
- **CORS in .env**

### Database Connection Issues

- **Test connect:** `psql -U go_backend_admin -d go_backend -h localhost`
- **Status:** `systemctl status postgresql`
- **Logs:** `tail -f /var/log/postgresql/postgresql-16-main.log`
- **Credentials:** `cat .env | grep DB_`

### 502 Bad Gateway

- **Backend status:** `systemctl status go-backend` and `curl http://localhost:3000/health`
- **Nginx logs:** `tail -f /var/log/nginx/api_error.log`
- **Restart:** `systemctl restart go-backend && systemctl reload nginx`

## Rollback Procedure

1. **Stop service:** `systemctl stop go-backend`
2. **List backups:** `ls -lh /var/www/backups/` (Create backups regularly, e.g., `tar -czf /var/www/backups/backup_$(date +%Y%m%d).tar.gz /var/www/go-backend`)
3. **Restore:** `cd /var/www/go-backend && tar -xzf /var/www/backups/backup_TIMESTAMP.tar.gz`
4. **Start:** `systemctl start go-backend`
5. **Status:** `systemctl status go-backend`

## Monitoring

### Set Up Log Rotation

```bash
nano /etc/logrotate.d/go-backend
```

Add:

```text
/var/log/go-backend/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        systemctl reload go-backend > /dev/null 2>&1 || true
    endscript
}
```

### Set Up Monitoring Script

```bash
nano /root/monitor-go-backend.sh
```

Add:

```bash
#!/bin/bash
if ! systemctl is-active --quiet go-backend; then
    echo "Go backend is down! Restarting..."
    systemctl restart go-backend
    echo "Go backend restarted at $(date)" >> /var/log/go-backend/monitor.log
fi
```

Then:

```bash
chmod +x /root/monitor-go-backend.sh
crontab -e  # Add: */5 * * * * /root/monitor-go-backend.sh
```