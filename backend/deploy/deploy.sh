#!/usr/bin/env bash
# Pulls latest code, rebuilds backend + frontend, redeploys both, restarts
# the service. This is also what a future CI/CD pipeline will SSH in and run.
#
# Usage: sudo ./deploy.sh
set -euo pipefail

APP_USER="${SUDO_USER:-ammad}"
APP_HOME="$(getent passwd "$APP_USER" | cut -d: -f6)"
APP_DIR="$APP_HOME/go-web"
BACKEND_DIR="$APP_DIR/backend"
FRONTEND_DIR="$APP_DIR/frontend"
WEB_ROOT="/var/www/go-web"
GO_BIN="/usr/local/go/bin/go"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this as root (sudo ./deploy.sh)" >&2
  exit 1
fi

echo "==> Pulling latest code"
sudo -u "$APP_USER" git -C "$APP_DIR" pull

echo "==> Building backend"
sudo -u "$APP_USER" bash -c "cd '$BACKEND_DIR' && CGO_ENABLED=0 '$GO_BIN' build -o bin/server ./cmd/server"

echo "==> Restarting backend service"
systemctl restart chatapp-backend
sleep 1
systemctl --no-pager --full status chatapp-backend || true

echo "==> Building frontend"
sudo -u "$APP_USER" bash -c "cd '$FRONTEND_DIR' && npm install && npm run build"

echo "==> Redeploying frontend to $WEB_ROOT"
rm -rf "${WEB_ROOT:?}"/*
cp -r "$FRONTEND_DIR/dist/"* "$WEB_ROOT/"
chown -R www-data:www-data "$WEB_ROOT"

echo ""
echo "==> Deployed. Backend logs: sudo journalctl -u chatapp-backend -f"