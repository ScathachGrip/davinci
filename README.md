## Dockerized

`ghcr.io/scathachgrip/davinci:latest`

### 1. Run on VPS using Docker Compose

Create a `docker-compose.yml`:

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
