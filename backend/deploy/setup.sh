#!/usr/bin/env bash
# One-time server setup: installs Go, clones the repo, builds the backend,
# and installs the systemd service + nginx reverse proxy. Safe to re-run —
# each step checks whether it's already done.
#
# Usage (on the Debian machine, as root):
#   sudo ./setup.sh
set -euo pipefail

REPO_URL="https://github.com/lazy-myst/go-web.git"
APP_DIR="/opt/chatapp"
APP_USER="chatapp"
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

if ! id "$APP_USER" >/dev/null 2>&1; then
  echo "==> Creating system user '$APP_USER'"
  useradd --system --create-home --home-dir "$APP_DIR" --shell /usr/sbin/nologin "$APP_USER"
fi

if [ -d "$APP_DIR/.git" ]; then
  echo "==> Repo already cloned at $APP_DIR, pulling latest"
  sudo -u "$APP_USER" git -C "$APP_DIR" pull
else
  echo "==> Cloning $REPO_URL into $APP_DIR"
  mkdir -p "$APP_DIR"
  chown "$APP_USER":"$APP_USER" "$APP_DIR"
  sudo -u "$APP_USER" git clone "$REPO_URL" "$APP_DIR"
fi

echo "==> Building backend"
sudo -u "$APP_USER" bash -c "cd '$APP_DIR/backend' && CGO_ENABLED=0 '$GO_BIN' build -o bin/server ./cmd/server"

if [ ! -f "$APP_DIR/backend/.env" ]; then
  echo "==> No .env found — copying the template. YOU MUST EDIT THIS with real values:"
  cp "$APP_DIR/backend/deploy/.env.example" "$APP_DIR/backend/.env"
  chown "$APP_USER":"$APP_USER" "$APP_DIR/backend/.env"
  chmod 600 "$APP_DIR/backend/.env"
  echo "    sudo nano $APP_DIR/backend/.env"
  echo "    then: sudo systemctl restart chatapp-backend"
fi

echo "==> Installing systemd service"
cp "$APP_DIR/backend/deploy/chatapp-backend.service" /etc/systemd/system/chatapp-backend.service
systemctl daemon-reload
systemctl enable chatapp-backend
systemctl restart chatapp-backend

if command -v nginx >/dev/null 2>&1; then
  echo "==> Installing nginx config"
  cp "$APP_DIR/backend/deploy/nginx-chatapp.conf" /etc/nginx/sites-available/chatapp
  ln -sf /etc/nginx/sites-available/chatapp /etc/nginx/sites-enabled/chatapp
  nginx -t
  systemctl reload nginx
else
  echo "==> nginx not found on this machine, skipping reverse proxy setup"
fi

echo ""
echo "==> Done."
echo "    Status: sudo systemctl status chatapp-backend"
echo "    Logs:   sudo journalctl -u chatapp-backend -f"
