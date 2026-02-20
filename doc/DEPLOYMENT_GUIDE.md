# Nagare Deployment Guide: Production & Staging

Moving from local development to a production environment requires specific security and performance configurations. This guide explains how to deploy Nagare as a stable, high-performance service.

---

## 🏗️ Deployment Architectures
1. **Single Node (Recommended)**: Both Backend and Frontend on one server. Use Nginx as a reverse proxy.
2. **Distributed Node**: Backend, Frontend, and MySQL on separate servers for high availability.

---

## ⚙️ Core Configuration (`nagare_config.json`)

Nagare looks for its configuration in `configs/nagare_config.json`. Below are the critical production settings:

```json
{
  "system": {
    "port": 8080,
    "mode": "release",
    "log_level": "info",
    "jwt_secret": "YOUR_STRONG_SECRET_HERE"
  },
  "database": {
    "dsn": "user:pass@tcp(localhost:3306)/nagare?charset=utf8mb4&parseTime=True&loc=Local"
  },
  "ai": {
    "gemini_api_key": "YOUR_GEMINI_KEY",
    "default_model": "gemini-1.5-pro"
  }
}
```

---

## 🔒 Security Best Practices
- **JWT Secret**: Never use the default. Generate a 64-character random string.
- **HTTPS**: Always run Nagare behind a reverse proxy (Nginx/Caddy) with SSL (Let's Encrypt).
- **SSH Credentials**: Nagare stores SSH credentials for WebSSH. Ensure your database is only accessible from the Nagare server's IP.
- **API Whitelisting**: Use the `X-Tunnel-Skip-AntiPhishing-Page: true` header if deploying behind a tunnel service like Microsoft Dev Tunnel.

---

## 📦 Deployment Steps (Manual)

### 1. Build the Backend
```bash
cd backend
go build -o nagare-web-server cmd/server/main.go
./nagare-web-server
```

### 2. Build the Frontend
```bash
cd frontend
npm install
npm run build
```
Copy the contents of `frontend/dist` to your Nginx web root.

## 🚀 Running as a System Service (systemd)
Create `/etc/systemd/system/nagare-backend.service`:

```ini
[Unit]
Description=Nagare Backend Service
After=network.target mysql.service

[Service]
Type=simple
User=nagare
WorkingDirectory=/opt/nagare/backend
ExecStart=/opt/nagare/backend/nagare-web-server
Restart=always
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

---

## 📡 Port Mapping & Reverse Proxy (Nginx)
Nagare requires WebSocket support for Site Messages and WebSSH.

```nginx
server {
    listen 80;
    server_name nagare.yourdomain.com;

    location / {
        root /var/www/nagare-frontend;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

---

## 🛠️ Troubleshooting Production
- **Logs**: Check `backend/logs/` or use `journalctl -u nagare-backend`.
- **Health Check**: Visit `http://your-server:8080/health`. A response of `{"status": "UP"}` means the core brain is healthy.
- **DB Migration**: Nagare uses GORM AutoMigrate. It will update the database schema automatically on startup.
