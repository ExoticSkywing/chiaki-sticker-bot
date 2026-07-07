#!/bin/sh
set -eu

ENV_FILE=${ENV_FILE:-.env}

if [ -f "$ENV_FILE" ]; then
    set -a
    # shellcheck disable=SC1090
    case "$ENV_FILE" in
        */*) . "$ENV_FILE" ;;
        *) . "./$ENV_FILE" ;;
    esac
    set +a
fi

DB_ADDR=${DB_ADDR:-127.0.0.1:3306}
DB_HOST=${DB_ADDR%:*}
DB_PORT=${DB_ADDR##*:}
if [ "$DB_HOST" = "$DB_PORT" ]; then
    DB_PORT=3306
fi
DB_USER=${DB_USER:-}
DB_NAME=${DB_NAME:-}
DB_PASS=${DB_PASS:-}

if [ -z "$DB_USER" ]; then
    echo "DB_USER is required. Set it in $ENV_FILE or the environment." >&2
    exit 1
fi

if [ -z "$DB_NAME" ]; then
    echo "DB_NAME is required. Set it in $ENV_FILE or the environment." >&2
    exit 1
fi

if ! command -v mysql >/dev/null 2>&1; then
    echo "mysql client is required. Install it first, or run this on the DB host." >&2
    exit 1
fi

export MYSQL_PWD="$DB_PASS"

mysql_base() {
    mysql -u"$DB_USER" -h"$DB_HOST" -P"$DB_PORT" "$@"
}

mysql_base <<SQL
CREATE DATABASE IF NOT EXISTS \`$DB_NAME\` CHARACTER SET utf8mb4;
SQL

mysql_base "$DB_NAME" <<'SQL'
CREATE TABLE IF NOT EXISTS line (
    line_id VARCHAR(128),
    tg_id VARCHAR(128),
    tg_title VARCHAR(255),
    line_link VARCHAR(512),
    auto_emoji BOOL
);

CREATE TABLE IF NOT EXISTS properties (
    name VARCHAR(128) PRIMARY KEY,
    value VARCHAR(128)
);

CREATE TABLE IF NOT EXISTS stickers (
    user_id BIGINT,
    tg_id VARCHAR(128),
    tg_title VARCHAR(255),
    timestamp BIGINT
);

CREATE TABLE IF NOT EXISTS events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    action VARCHAR(32) NOT NULL,
    pack_id VARCHAR(128),
    status VARCHAR(16) NOT NULL,
    reason TEXT NOT NULL,
    ts DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_ts (ts)
);

CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(64),
    display_name VARCHAR(128),
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT IGNORE INTO properties (name, value) VALUES ('last_line_dedup_index', '-1');
INSERT INTO properties (name, value) VALUES ('DB_VER', '7') ON DUPLICATE KEY UPDATE value='7';
SQL

mysql_base "$DB_NAME" <<'SQL'
SELECT name, value FROM properties WHERE name IN ('DB_VER', 'last_line_dedup_index') ORDER BY name;
SQL

echo "Database schema initialized: $DB_NAME on $DB_HOST:$DB_PORT"
