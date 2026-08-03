# Backend deployment (Debian, no CI/CD yet)

Runs as a native binary under systemd, with nginx reverse-proxying to it.
MongoDB stays on Atlas (or wherever `MONGO_URI` already points) — nothing to
install for that. This covers the backend only; the frontend isn't part of
this pass.

## Layout on the server

```
/opt/chatapp/            <- full repo clone, owned by the `chatapp` user
/opt/chatapp/backend/bin/server   <- compiled binary
/opt/chatapp/backend/.env         <- secrets, created by hand, never in git
/etc/systemd/system/chatapp-backend.service
/etc/nginx/sites-available/chatapp
```

## One-time setup

1. Copy this `deploy/` folder's scripts to the server, or just clone the repo
   there first — `setup.sh` will happily reuse an existing clone at
   `/opt/chatapp` if one's already there.

2. On the Debian machine:

   ```bash
   chmod +x setup.sh deploy.sh   # git on Windows won't have set the exec bit
   sudo ./setup.sh
   ```

   This installs Go 1.24 (official tarball, only if missing), creates a
   dedicated `chatapp` system user, clones the repo to `/opt/chatapp`, builds
   `backend/bin/server`, installs the systemd unit, and installs the nginx
   site config — all idempotent, safe to re-run.

3. `setup.sh` will stop and tell you to fill in `/opt/chatapp/backend/.env`
   the first time (copied from `.env.example`) — it has no way to know your
   real Mongo URI / JWT secret. Edit it:

   ```bash
   sudo nano /opt/chatapp/backend/.env
   ```

   Then restart:

   ```bash
   sudo systemctl restart chatapp-backend
   ```

4. Firewall — only nginx should be internet-facing, not the raw backend port:

   ```bash
   sudo ufw allow OpenSSH
   sudo ufw allow 80/tcp
   sudo ufw enable   # if not already on
   ```

   Do **not** open 3001 externally — the backend only needs to be reachable
   from nginx on localhost, which it already is.

## Verify

```bash
sudo systemctl status chatapp-backend     # should be "active (running)"
sudo journalctl -u chatapp-backend -f     # watch it connect to Mongo etc.

curl -i http://localhost/api/users        # expect 401 (no token) — proves
                                           # nginx -> backend is wired up
```

If you're testing from another machine, replace `localhost` with the
server's IP: `http://<server-ip>/api/users`.

## Deploying an update

Every time you push new backend code:

```bash
sudo /opt/chatapp/backend/deploy/deploy.sh
```

Pulls latest `main`, rebuilds the binary, restarts the service. This is also
exactly what a future CI/CD job would SSH in and run — see "Later" below.

## Notes / things you'll want to revisit

- **CORS_ORIGINS**: set this in `.env` to wherever the frontend actually gets
  served from (e.g. `http://<server-ip>` or `http://<server-ip>:5173`). The
  backend won't accept browser requests from an origin that isn't listed
  here — this is the #1 thing that'll look like a bug if forgotten.
- **No domain / TLS yet**: nginx listens on plain port 80. Once you have a
  domain pointed at this box, `sudo apt install certbot python3-certbot-nginx`
  and `sudo certbot --nginx` will get you HTTPS in about a minute — ask and
  I'll wire that up along with the `wss://`-equivalent gRPC-Web considerations.
- **go.sum isn't committed** (it's in `.gitignore`), so `go build` on the
  server re-resolves and re-verifies module checksums from the Go proxy on
  first build rather than using a pinned lockfile. Works fine as long as the
  server has internet access, just flagging it as non-standard.
- **CI/CD**: deferred for now. When you're ready, the shape will likely be a
  GitHub Actions workflow that SSHes in on push to `main` and runs
  `deploy.sh` — happy to set that up whenever.
