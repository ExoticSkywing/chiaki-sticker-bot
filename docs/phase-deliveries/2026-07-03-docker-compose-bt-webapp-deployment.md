# 阶段性交付包：Docker Compose + 宝塔完整部署

## 1. 原始需求

甲方希望将 `chiaki-sticker-bot` 从低内存 Fly.io 取向的部署方式，整理为适合当前服务器的完整功能部署方案：

- 使用 Docker Compose 部署；
- 使用宝塔面板/Nginx 做 HTTPS 和反向代理；
- 保留完整功能，尤其是 WebApp；
- 使用本机/宝塔 MySQL；
- 按当前服务器配置释放性能；
- 部署跑通后整理文档，形成可接棒状态。

## 2. 验收结论

本阶段甲方已表示满意。

当前已形成稳定基线：

- Docker Compose 容器可启动；
- bot 健康检查通过；
- 宝塔公网 HTTPS 访问通过；
- Telegram webhook 可达；
- MySQL 初始化完成并可用；
- WebApp 配置保留并启用；
- Telegram 命令菜单已注册；
- 部署文档和状态文档已整理。

## 3. 最终方案

### 3.1 部署结构

```text
Telegram / Browser
        |
        v
https://stickerbot.tsmoe.com
        |
        v
宝塔 Nginx / SSL / 反向代理
        |
        v
http://127.0.0.1:18080
        |
        v
Docker Compose: chiaki-sticker-bot
        |
        +-- container nginx: /webhook /health /webapp/*
        +-- Go bot HTTP server: :8081
        +-- ffmpeg / ImageMagick / Python tools
        +-- /data persistent volume
        |
        v
宿主机 MySQL: 127.0.0.1:3306
```

### 3.2 Compose 网络模式

当前使用：

```yaml
network_mode: host
```

原因：

- 直接访问宿主机 MySQL：`127.0.0.1:3306`；
- 避免 Docker bridge 网关和宝塔/系统防火墙带来的额外不确定性；
- 由容器内部 nginx 直接监听宿主机 `18080`。

### 3.3 宝塔反向代理

宝塔站点 `stickerbot.tsmoe.com` 配置：

```text
代理目录: /
目标 URL: http://127.0.0.1:18080
```

宝塔/云厂商需要放行公网 `443`。

## 4. 当前状态

### 4.1 关键文件

- `docker-compose.yml`：Docker Compose 部署定义；
- `.env.example`：环境变量模板；
- `start-bot.sh`：容器启动脚本，已改为环境变量驱动；
- `web/nginx/fly.conf`：容器内部 nginx 配置，当前监听 `18080`；
- `DEPLOY_DOCKER_COMPOSE.md`：宝塔 + Docker Compose 操作手册；
- `DEPLOYMENT_STATUS.md`：当前已验证部署状态；
- `docs/HANDOFF.md`：项目接棒入口。

### 4.2 当前 `.env` 关键项

真实 `.env` 不应提交。当前模板含义如下：

```dotenv
WEBHOOK_URL=https://stickerbot.tsmoe.com/webhook
WEBAPP_URL=https://stickerbot.tsmoe.com/webapp
DB_ADDR=127.0.0.1:3306
DB_NAME=stickerbot
DB_TLS_CONFIG=false
HOST_PORT=18080
```

### 4.3 MySQL 状态

当前使用宿主机 MySQL：

```dotenv
DB_ADDR=127.0.0.1:3306
DB_NAME=stickerbot
DB_TLS_CONFIG=false
```

`DB_TLS_CONFIG=false` 是为了适配本地/宝塔 MySQL。此前 TLS 校验失败，因为本地 MySQL 证书不能验证 `127.0.0.1`。

数据库 `stickerbot` 已初始化需要的表和 `DB_VER=7`。

## 5. 新增能力

### 5.1 Docker Compose 部署

新增 `docker-compose.yml`，包含：

- 构建本地镜像；
- host 网络；
- `/data` 持久化；
- healthcheck；
- 当前服务器性能环境变量。

### 5.2 WebApp 完整启用

通过：

```dotenv
WEBAPP_URL=https://stickerbot.tsmoe.com/webapp
WEBAPP_DATA_DIR=/data/webapp
```

启用 `/manage` 的 WebApp 管理能力。

### 5.3 数据库灵活配置

代码支持：

- `DB_NAME`：指定已有数据库名；
- `DB_TLS_CONFIG=false|0|disable`：本地 MySQL 禁用 TLS。

### 5.4 性能释放

新增：

```dotenv
MSB_WEBM_WORKER_CONCURRENCY=2
```

替代原本 hardcoded `ants.NewPoolWithFunc(1, ...)`，使动画 WEBM 转换并发适配当前 2 vCPU 服务器。

当前性能默认值：

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

### 5.5 Telegram 命令菜单

程序启动时注册 Telegram 命令菜单：

- `/start`
- `/import`
- `/download`
- `/create`
- `/manage`
- `/search`
- `/help`
- `/about`

已用 Telegram API `getMyCommands` 验证。

## 6. 操作步骤

### 6.1 启动/重建

```bash
cd /root/data/lty/repo/chiaki-sticker-bot
docker compose up -d --build
```

### 6.2 强制重建

`.env`、Compose 或代码变动后：

```bash
docker compose up -d --build --force-recreate
```

### 6.3 查看状态

```bash
docker compose ps
```

### 6.4 查看日志

```bash
docker compose logs -f
```

### 6.5 健康检查

本地：

```bash
curl -fsS http://127.0.0.1:18080/health
```

公网：

```bash
curl -fsS https://stickerbot.tsmoe.com/health
```

预期：

```text
ok
```

## 7. 已执行验证

智能体已执行并观察到：

- `docker compose ps` 显示容器 healthy；
- `curl http://127.0.0.1:18080/health` 返回 `ok`；
- `curl https://stickerbot.tsmoe.com/health` 返回 `ok`；
- bot 日志出现：
  - `MariaDB OK.`
  - `Bot OK.`
- Telegram `getMyCommands` 返回已注册命令；
- Telegram webhook 此前因 Oracle 443 未开放超时，开放后公网健康检查已恢复。

## 8. 需要人类执行的验证

以下更适合甲方手动确认：

1. Telegram 客户端中 `/start` 是否稳定回复；
2. 左下角命令菜单是否在客户端刷新后出现；
3. `/manage` 打开的 WebApp 是否能正常编辑/管理贴纸；
4. 实际导入 LINE/Kakao 动态贴纸包时，速度和资源占用是否符合预期；
5. 宝塔 SSL 续期和反代配置长期是否稳定。

## 9. 踩坑边界

1. Docker bridge 下 `127.0.0.1` 指容器自身，不是宿主机；当前改用 host 网络避免该问题。
2. 本地 MySQL TLS 会因为 `127.0.0.1` 证书 SAN 不匹配失败，需 `DB_TLS_CONFIG=false`。
3. 数据库用户没有 `CREATE DATABASE` 权限时，需要指定已有 `DB_NAME` 并初始化 schema。
4. Telegram webhook 依赖公网 `443`，Oracle 安全组/宝塔/系统防火墙任一处不通都会导致 bot 无响应。
5. Telegram 客户端命令菜单刷新可能有延迟；API 注册成功不一定立刻在所有客户端 UI 出现。
6. `.env` 含 token/password，不能提交。

## 10. 接手说明

下一任接手时优先阅读：

1. `docs/HANDOFF.md`
2. 本交付包
3. `DEPLOY_DOCKER_COMPOSE.md`
4. `DEPLOYMENT_STATUS.md`
5. `.env.example`

不要从 Fly.io 配置判断当前生产状态；当前生产部署以 Docker Compose + 宝塔 + host network 为准。

## 11. 后续建议

1. 等甲方手动验证 `/start`、命令菜单、`/manage` WebApp 后，再进入下一阶段。
2. 后续可补充自动化 smoke test：健康检查、webhook info、命令菜单、WebApp 静态资源。
3. 后续可把数据库 schema 初始化做成脚本，避免手动 SQL。
4. 如未来迁移到非宝塔环境，再重新拆分 Nginx/Compose 网络模式文档。
