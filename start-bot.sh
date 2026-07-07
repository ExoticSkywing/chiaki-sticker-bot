#!/bin/sh
set -eu

DATA_DIR=${DATA_DIR:-/data}
LOG_LEVEL=${LOG_LEVEL:-info}
ADMIN_UID=${ADMIN_UID:--1}
WEBAPP_DATA_DIR=${WEBAPP_DATA_DIR:-/data/webapp}

nginx -g "daemon off;" &
exec moe-sticker-bot \
    --data_dir="$DATA_DIR" \
    --log_level="$LOG_LEVEL" \
    --bot_token="$BOT_TOKEN" \
    --db_addr="${DB_ADDR:-}" \
    --db_user="${DB_USER:-}" \
    --db_pass="${DB_PASS:-}" \
    --admin_uid="$ADMIN_UID" \
    --webapp_url="${WEBAPP_URL:-}" \
    --webapp_data_dir="$WEBAPP_DATA_DIR" \
    --webhook_url="${WEBHOOK_URL:-}" \
    --webhook_secret="${WEBHOOK_SECRET:-}"
