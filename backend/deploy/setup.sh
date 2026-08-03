#!/usr/bin/env bash
# One-time server setup: installs Go/Node, clones the repo into the current
# user's home, builds backend + frontend, installs systemd service + nginx.
# Safe to re-run.
#
# Usage (on the Debian machine):
#   sudo ./setup.sh
set -euo pipefail

REPO_URL="https://github.com/lazy-myst/go-web.git"
APP_USER="${SUDO_USER:-ammad}"
APP_HOME="$(getent passwd "$APP_USER" | cut -d: -f6)"
APP_DIR="$APP_HOME/go-web"
BACKEND_DIR="$APP_DIR/backend"
FRONTEND_DIR="$APP_DIR/frontend"
WEB_ROOT="/var/www/go-web"
GO_VERSION="1.24.13"
GO_BIN="/usr/local/go/bin/go"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this as root (sudo ./setup.sh)" >&2
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "==> Installing git"
  apt-get update && apt-get install -y git
fi

if [ ! -x "$GO_BIN" ]; then
  echo "==> Installing Go ${GO_VERSION}"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >/etc/profile.d/go.sh
  chmod +x /etc/profile.d/go.sh
else
  echo "==> Go already installed: $($GO_BIN version)"
fi

if ! command -v node >/dev/null 2>&1; then
  echo "==> Installing Node.js LTS"
  curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -
  apt-get install -y nodejs
fi

if [ -d "$APP_DIR/.git" ]; then
  echo "==> Repo already cloned at $APP_DIR, pulling latest"
  sudo -u "$APP_USER" git -C "$APP_DIR" pull
else
  echo "==> Cloning $REPO_URL into $APP_DIR"
  sudo -u "$APP_USER" git clone "$REPO_URL" "$APP_DIR"
fi

echo "==> Building backend"
sudo -u "$APP_USER" bash -c "cd '$BACKEND_DIR' && CGO_ENABLED=0 '$GO_BIN' build -o bin/server ./cmd/server"

if [ ! -f "$BACKEND_DIR/.env" ]; then
  echo "==> No .env found — copying template. YOU MUST EDIT THIS with real values:"
  cp "$BACKEND_DIR/deploy/.env.example" "$BACKEND_DIR/.env"
  chown "$APP_USER":"$APP_USER" "$BACKEND_DIR/.env"
  chmod 600 "$BACKEND_DIR/.env"
  echo "    sudo nano $BACKEND_DIR/.env"
fi

echo "==> Building frontend"
sudo -u "$APP_USER" bash -c "cd '$FRONTEND_DIR' && npm install && npm run build"

echo "==> Deploying frontend to $WEB_ROOT"
mkdir -p "$WEB_ROOT"
rm -rf "${WEB_ROOT:?}"/*
cp -r "$FRONTEND_DIR/dist/"* "$WEB_ROOT/"
chown -R www-data:www-data "$WEB_ROOT"

echo "==> Installing systemd service"
BACKEND_PORT="$(grep -E '^PORT=' "$BACKEND_DIR/.env" | cut -d= -f2)"
cat > /etc/systemd/system/chatapp-backend.service <<EOF
[Unit]
Description=Chat backend (Go/Fiber REST + gRPC-Web chat)
After=network.target

[Service]
Type=simple
User=$APP_USER
Group=$APP_USER
WorkingDirectory=$BACKEND_DIR
EnvironmentFile=$BACKEND_DIR/.env
ExecStart=$BACKEND_DIR/bin/server
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable chatapp-backend
systemctl restart chatapp-backend

if command -v nginx >/dev/null 2>&1; then
  echo "==> Installing nginx config (backend port: ${BACKEND_PORT:-3001})"
  cat > /etc/nginx/sites-available/chatapp <<EOF
server {
    listen 80;
    server_name _;

    root $WEB_ROOT;
    index index.html;

    # Backend REST API
    location /api/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT:-3001};
        proxy_http_version 1.1;
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    # gRPC-Web chat service — path is the fully-qualified service name,
    # not a prefix you chose (see ChatService_ServiceDesc.ServiceName).
    location /chat.ChatService/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT:-3001};
        proxy_http_version 1.1;
        proxy_buffering off;
        proxy_request_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Connection "";
    }

    # Frontend SPA
    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
  ln -sf /etc/nginx/sites-available/chatapp /etc/nginx/sites-enabled/chatapp
  rm -f /etc/nginx/sites-enabled/default
  nginx -t
  systemctl reload nginx
else
  echo "==> nginx not found, installing"
  apt-get install -y nginx
fi

echo ""
echo "==> Done."
echo "    Backend status: sudo systemctl status chatapp-backend"
echo "    Backend logs:   sudo journalctl -u chatapp-backend -f"