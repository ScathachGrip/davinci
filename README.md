# DaVinci GitHub Bot (Go Rewrite)

A native Go rewrite of the experimental DaVinci GitHub App bot. The bot listens to webhook events and automatically welcomes contributors with labels and anime reaction GIFs.

This version runs fully in Go, compiles to a single native binary, has zero Node.js/npm dependencies, and utilizes the free and keyless **nekos.best** and **waifu.pics** APIs.

---

## 🚀 Features

- **Automated Issue Triage:** Welcomes issues with a randomized anime greeting GIF (such as `pout`, `smug`, `smile`) and applies the `triage` label.
- **Automated PR Pending Tracking:** Welcomes pull requests with a randomized reaction GIF (such as `pat`, `happy`, `nom`) and applies a `pr:pending` label.
- **Clean Label Removal:** Automatically removes the `pr:pending` label once the pull request is closed or merged.
- **Zero Node/npm Dependencies:** Executes entirely as a compiled native binary.
- **Local Webhook Proxying:** Can be debugged locally without using Node-based `smee.io`.

---

## 🛠️ Prerequisites

- **Golang:** Make sure Go (v1.22+) is installed. Run `go version` to check.
- **GitHub App:** A registered GitHub App with the following permissions:
  - **Issues:** Read & Write
  - **Pull Requests:** Read & Write
  - **Metadata:** Read-only (required by default)
  - **Subscribed Events:** Issues (`opened`), Pull requests (`opened`, `closed`)

---

## ⚙️ Setup & Configuration

1. **Clone and Navigate:** Open the workspace directory containing the Go files.
2. **Environment Variables:** Create a `.env` file based on the template:
   ```cmd
   cp .env.example .env
   ```
   Configure the variables:
   - `APP_ID`: The ID of your GitHub App (found in General settings).
   - `WEBHOOK_SECRET`: The webhook secret you configured in your GitHub App (e.g. `development`).
   - `PORT`: Optional local HTTP port (defaults to `8080`, configures `6667` in the example).
3. **Private Key:** Download the private key (`.pem`) file from your GitHub App's dashboard and place it directly in the project root folder. The Go bot will auto-detect and load it automatically!

---

## 🖥️ Running Locally

1. **Verify & Run Tests:**
   Make sure everything is working as expected:
   ```cmd
   go test -v ./...
   ```
2. **Run Server:**
   Start the webhook listener:
   ```cmd
   go run .
   ```
   The application will start on `http://localhost:8080`. You can visit `http://localhost:8080/` in your browser to verify it is online.

---

## 🌐 Receiving Webhooks Locally (No Smee/Node/npx Required)

Because GitHub cannot send webhooks directly to `localhost`, you must tunnel traffic to your local server.

We recommend using **Cloudflare Tunnel (`cloudflared`)**, which is a native Go tool and completely free (no account signup required):

1. **Download:** Download the standalone `cloudflared` binary from [Cloudflare](https://github.com/cloudflare/cloudflared/releases) or install it via your system package manager.
2. **Start Tunnel:** Expose your local port:
   ```cmd
   cloudflared tunnel --url http://localhost:8080
   ```
3. **Configure GitHub App:**
   Copy the generated HTTPS URL (e.g., `https://xxxxxx.trycloudflare.com`) and paste it as the **Webhook URL** under your GitHub App settings (General > Webhooks).

Now, whenever you open an issue or pull request in a repository where the App is installed, the webhooks will route directly to your local Go server!

---

## 🐳 Docker Deployment (VPS)

Since the image is automatically built and pushed to GitHub Container Registry (`ghcr.io/scathachgrip/davinci:latest`) by the CI workflow, you only need to run it on your VPS using Docker Compose.

### 1. Run on VPS using Docker Compose

Create a `docker-compose.yml` on your VPS:

```yaml
version: "3.8"

services:
  da-vinci-bot:
    image: ghcr.io/scathachgrip/davinci:latest
    container_name: da-vinci-bot
    restart: unless-stopped
    ports:
      - "6667:6667"
    environment:
      - PORT=6667
      - APP_ID=YOUR_APP_ID
      - WEBHOOK_SECRET=YOUR_WEBHOOK_SECRET
    volumes:
      # Mount the private key file on your VPS into the container
      - ./da-vinci-bot.2026-07-25.private-key.pem:/app/private-key.pem:ro
```

Place your private key file `da-vinci-bot.2026-07-25.private-key.pem` next to the `docker-compose.yml` file, then run:

### 2. Run the Containers

Run the following commands in the same directory as your `docker-compose.yml`:

```bash
# Pull latest images
docker compose pull

# Optional: Sync and fresh memory init
sync; echo 3 | sudo tee /proc/sys/vm/drop_caches;free -h

# Start all services in background
docker compose up -d
```

### 3. Logs all services

Log after starting the services:

```bash
docker compose logs -f
```

### 4/?. Stop all services

Stop all services and fresh memory init:

```bash
docker compose down
```

### 5/?. Purge all inactive containers

Purge all inactive containers:

```bash
docker system prune -a --volumes
```

### 6/?. Fresh memory

Optional for performance boost:

```bash
sync; echo 3 | sudo tee /proc/sys/vm/drop_caches;free -h
```

The bot will start in detached mode, listening on port `6667`.
