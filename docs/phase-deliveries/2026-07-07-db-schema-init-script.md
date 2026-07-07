# 阶段性交付包：数据库 schema 初始化脚本

## 1. 原始需求

甲方指定下一阶段只补一个能力：数据库 schema 初始化脚本。

目标是替代部署文档中的手写 SQL，让本地/宝塔 MySQL 的数据库初始化可重复执行、可接棒、可文档化。

## 2. 验收结论

本阶段完成：

- 新增 `scripts/init-db-schema.sh`；
- 默认读取 `.env`；
- 自动创建/确认数据库和必要表；
- 写入/更新 `properties.DB_VER=7`；
- 写入/保留 `properties.last_line_dedup_index=-1`；
- 脚本已在当前数据库上连续执行两次验证幂等；
- 部署文档、状态文档和接棒入口已更新。

## 3. 最终方案

新增脚本：

```text
scripts/init-db-schema.sh
```

默认用法：

```bash
./scripts/init-db-schema.sh
```

指定 env 文件：

```bash
ENV_FILE=/path/to/.env ./scripts/init-db-schema.sh
```

脚本读取：

```dotenv
DB_ADDR=127.0.0.1:3306
DB_USER=stickerbot
DB_NAME=stickerbot
DB_PASS=...
```

## 4. 新增能力

项目现在具备独立的数据库 schema 初始化能力：

- 可由部署人员直接运行；
- 可重复执行；
- 不依赖复制粘贴多行 SQL；
- 与当前 `.env` 配置一致；
- 适配宝塔/宿主机 MySQL。

## 5. 操作步骤

```bash
cd /root/data/lty/repo/chiaki-sticker-bot
./scripts/init-db-schema.sh
```

成功时输出类似：

```text
name                    value
DB_VER                  7
last_line_dedup_index   -1
Database schema initialized: stickerbot on 127.0.0.1:3306
```

## 6. 已执行验证

智能体已执行：

```bash
./scripts/init-db-schema.sh
./scripts/init-db-schema.sh
```

两次均成功，确认脚本幂等。

## 7. 需要人类执行的验证

无强制人工验证项。

如果未来换数据库或迁移服务器，建议部署人员在新环境执行一次：

```bash
./scripts/init-db-schema.sh
```

并确认 bot 日志出现：

```text
MariaDB OK.
Bot OK.
```

## 8. 踩坑边界

1. 脚本需要本机安装 `mysql` client。
2. 如果 `DB_USER` 没有 `CREATE DATABASE` 权限，需要先在宝塔/MySQL 面板创建 `DB_NAME` 并授权。
3. `.env` 含密码，不能提交。
4. 当前脚本使用 shell 读取 `.env`，因此 `.env` 应保持简单的 `KEY=value` 格式。

## 9. 接手说明

下一任接手时：

- 数据库初始化优先使用 `scripts/init-db-schema.sh`；
- 不再从旧文档复制手写 SQL；
- 当前部署细节见 `DEPLOY_DOCKER_COMPOSE.md` 和 `DEPLOYMENT_STATUS.md`。
