#!/usr/bin/env bash
# Pulls the latest code, rebuilds the backend binary, and restarts the
# service. Run this for every update after the initial setup.sh. This is
# also the script a future CI/CD pipeline would SSH in and run.
#
# Usage (on the Debian machine, as root):
#   sudo ./deploy.sh
set -euo pipefail

APP_DIR="/opt/chatapp"
APP_USER="chatapp"
GO_BIN="/usr/local/go/bin/go"

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this as root (sudo ./deploy.sh)" >&2
  exit 1
fi

echo "==> Pulling latest code"
sudo -u "$APP_USER" git -C "$APP_DIR" pull

echo "==> Building backend"
sudo -u "$APP_USER" bash -c "cd '$APP_DIR/backend' && CGO_ENABLED=0 '$GO_BIN' build -o bin/server ./cmd/server"

echo "==> Restarting service"
systemctl restart chatapp-backend
sleep 1
systemctl --no-pager --full status chatapp-backend || true

echo ""
echo "==> Deployed. Tail logs with: sudo journalctl -u chatapp-backend -f"
