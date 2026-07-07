# Docker Compose + 宝塔面板部署

这份部署方式适合当前服务器：

- Docker Compose 运行 bot 容器
- 容器使用 `network_mode: host`
- 宝塔面板负责域名、SSL、反向代理
- 宝塔/宿主机 MySQL 提供 `/search` 和使用记录
- WebApp 完整启用在 `/webapp`

当前默认性能参数按这台服务器设置：

- 2 vCPU
- 约 11 GiB RAM
- bot 容器内置 nginx 监听宿主机 `127.0.0.1:18080`

## 1. 准备 `.env`

```bash
cp .env.example .env
```

编辑 `.env`，至少填写：

```dotenv
BOT_TOKEN=123456789:your-real-token
WEBHOOK_URL=https://stickerbot.tsmoe.com/webhook
WEBAPP_URL=https://stickerbot.tsmoe.com/webapp
WEBHOOK_SECRET=replace-with-random-hex
ADMIN_UID=your-telegram-user-id

DB_ADDR=127.0.0.1:3306
DB_USER=stickerbot
DB_NAME=stickerbot
DB_PASS=your-db-password
DB_TLS_CONFIG=false

HOST_PORT=18080
```

生成 webhook secret：

```bash
openssl rand -hex 32
```

注意：

- `WEBHOOK_URL` 必须是公网 HTTPS，并且以 `/webhook` 结尾。
- `WEBAPP_URL` 必须是公网 HTTPS，并且以 `/webapp` 结尾。
- `WEBAPP_URL` 不为空时，`/manage` 的可视化 WebApp 才会启用。
- 当前 Compose 使用 `network_mode: host`，所以 `DB_ADDR=127.0.0.1:3306` 指向宿主机 MySQL。
- 本地宝塔 MySQL 通常没有可被 `127.0.0.1` 验证的 TLS 证书，所以设置 `DB_TLS_CONFIG=false`。
- `HOST_PORT=18080` 用于记录/文档兼容；host 网络模式下实际监听端口由容器内部 nginx 配置决定。

## 2. MySQL 数据库

当前代码支持通过 `DB_NAME` 指定数据库名。初始化 schema 使用脚本：

```bash
./scripts/init-db-schema.sh
```

脚本默认读取当前目录的 `.env`，并使用这些变量：

```dotenv
DB_ADDR=127.0.0.1:3306
DB_USER=stickerbot
DB_NAME=stickerbot
DB_PASS=your-db-password
```

如果要指定其它 env 文件：

```bash
ENV_FILE=/path/to/.env ./scripts/init-db-schema.sh
```

脚本可重复执行，会确保数据库和必要表存在，并写入/更新：

```text
properties.DB_VER = 7
properties.last_line_dedup_index = -1
```

如果数据库用户有 `CREATE DATABASE` 权限，脚本会自动创建 `DB_NAME` 指定的数据库；否则请先在宝塔/MySQL 面板创建数据库并授权给 `DB_USER`。

## 3. 启动 Docker Compose

```bash
docker compose up -d --build
```

查看状态：

```bash
docker compose ps
```

查看日志：

```bash
docker compose logs -f
```

正常日志应包含：

```text
MariaDB OK.
Bot OK.
```

持久化数据目录：

```text
./data
```

## 4. 本地健康检查

在服务器上执行：

```bash
curl -fsS http://127.0.0.1:18080/health
```

正常返回：

```text
ok
```

如果这一步不通，先看容器日志，不要先排查宝塔。

## 5. 宝塔面板配置反向代理

进入宝塔：

```text
网站 -> stickerbot.tsmoe.com -> 反向代理 -> 添加反向代理
```

建议配置：

```text
代理目录: /
目标 URL: http://127.0.0.1:18080
```

这个项目的容器内部 nginx 会自己处理这些路径：

```text
/webhook
/health
/webapp/static
/webapp/edit
/webapp/export
/webapp/data
/webapp/api
```

所以宝塔里直接代理 `/` 到 `127.0.0.1:18080` 就可以，不需要单独拆 `/webhook`、`/webapp`、`/health`。

## 6. 宝塔/云厂商防火墙

Telegram webhook 和 Telegram WebApp 都要求公网 HTTPS。

需要确保：

- 宝塔站点 SSL 已开启；
- 云厂商安全组放行 `443`；
- 宝塔/系统防火墙放行 `443`；
- 域名 DNS 指向当前服务器公网 IP。

公网测试：

```bash
curl -fsS https://stickerbot.tsmoe.com/health
```

正常返回：

```text
ok
```

## 7. Telegram 测试

给 bot 发送：

```text
/start
/manage
```

如果 `.env` 里的 `WEBAPP_URL` 正确，WebApp 按钮应该打开：

```text
https://stickerbot.tsmoe.com/webapp/...
```

命令菜单由程序启动时自动注册，可用下面命令验证：

```bash
set -a
. ./.env
set +a
curl -fsS "https://api.telegram.org/bot${BOT_TOKEN}/getMyCommands"
```

## 性能默认值

`docker-compose.yml` 默认使用：

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

对这台 2 vCPU 服务器，建议先保持这套配置。

如果 CPU 长期没跑满，可以优先尝试：

```dotenv
MSB_IMPORT_CONCURRENCY=4
```

不要一开始就把 `MSB_FFMPEG_CONCURRENCY` 拉太高。ffmpeg / ImageMagick 是重 CPU 任务，2 核机器上重转码并发长期超过 `2` 通常会让单个任务变慢。

## 常用命令

```bash
# 启动 / 重新构建
docker compose up -d --build

# 强制重建容器，使 .env 或 compose 改动生效
docker compose up -d --build --force-recreate

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 停止
docker compose down

# 校验 compose 配置
docker compose config
```
