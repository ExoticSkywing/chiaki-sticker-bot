# Deployment Status

Date: 2026-07-03

## Verified deployment

The bot is deployed with Docker Compose and Baota/BT panel reverse proxy.

- Compose file: `docker-compose.yml`
- Env file: `.env` (not committed)
- Deployment guide: `DEPLOY_DOCKER_COMPOSE.md`
- Public domain: `https://stickerbot.tsmoe.com`
- Public health endpoint verified: `https://stickerbot.tsmoe.com/health` -> `ok`
- Local health endpoint verified: `http://127.0.0.1:18080/health` -> `ok`
- Container status verified healthy with `docker compose ps`
- Bot logs verified:
  - `MariaDB OK.`
  - `Bot OK.`

## Actual network shape

The container uses host networking:

```yaml
network_mode: host
```

The container's internal nginx listens on host port `18080` via `web/nginx/fly.conf`:

```nginx
listen 18080;
```

BT panel reverse proxy should point `/` to:

```text
http://127.0.0.1:18080
```

Oracle Cloud / BT firewall must allow public `443` for Telegram webhook and WebApp.

## Database

The current local DB settings are:

```dotenv
DB_ADDR=127.0.0.1:3306
DB_NAME=stickerbot
DB_TLS_CONFIG=false
```

`DB_TLS_CONFIG=false` is required for the local BT/MySQL setup because MySQL TLS verification failed against `127.0.0.1`.

The `stickerbot` DB schema is initialized by `scripts/init-db-schema.sh`, including `properties` with `DB_VER=7`. The script is idempotent and reads `.env` by default.

Code now supports:

- `DB_NAME` to select an existing database instead of forcing `{botName}_db`.
- `DB_TLS_CONFIG=false|0|disable` to disable DB TLS for local MySQL.

## WebApp

WebApp is enabled through:

```dotenv
WEBAPP_URL=https://stickerbot.tsmoe.com/webapp
WEBAPP_DATA_DIR=/data/webapp
```

The container serves WebApp static files and API paths through its internal nginx.

## Command menu

Telegram command menu registration was added in `core/init.go` and verified with `getMyCommands`.

Registered commands:

- `/start`
- `/import`
- `/download`
- `/create`
- `/manage`
- `/search`
- `/help`
- `/about`

## TGS transparent GIF backend

`.tgs -> gif` now defaults to a dual-backend strategy:

```dotenv
MSB_TGS_GIF_BACKEND=auto
MSB_TGS_GIF_WIDTH=512
MSB_TGS_GIF_HEIGHT=512
MSB_TGS_GIF_FPS=30
MSB_TGS_GIF_QUALITY=90
MSB_TGS_GIF_THREADS=2
```

`auto` tries lottie-converter + gifski first, then falls back to the legacy `rlottie-python` path. This has been verified by the user to turn the previously black-background official TGS GIF output into transparent GIF output.

CPU can reach 100% during conversion on the current 2 vCPU server. For smoother bot responsiveness, reduce `MSB_TGS_GIF_THREADS` to `1`; for smoother GIF output, the user may tune FPS/quality in `.env`.

## Performance tuning

Defaults are tuned for the current 2 vCPU / ~11 GiB RAM server:

```dotenv
GOMEMLIMIT=6GiB
MSB_IMPORT_CONCURRENCY=3
MSB_FFMPEG_CONCURRENCY=2
MSB_WEBM_WORKER_CONCURRENCY=2
MSB_IM_MEMORY_LIMIT=1GiB
MSB_IM_MAP_LIMIT=2GiB
MSB_IM_OOM_MEMORY_LIMIT=512MiB
MSB_IM_OOM_MAP_LIMIT=1GiB
MSB_KAKAO_FAST_PIPE=1
MSB_CONVERT_TIMEOUT_SECONDS=240
MSB_IMPORT_QUEUE_TIMEOUT_SECONDS=1800
```

`MSB_WEBM_WORKER_CONCURRENCY` was added to replace the hardcoded animated WEBM worker pool size of `1`.
