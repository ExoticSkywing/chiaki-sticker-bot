# 项目接棒入口

## 1. 项目定位

`chiaki-sticker-bot` 是一个自托管 Telegram sticker bot，用于：

- 导入 LINE / Kakao 贴纸包到 Telegram；
- 创建、下载、管理 Telegram sticker set / CustomEmoji；
- 通过 WebApp 做可视化管理；
- 可选使用 MySQL 记录和搜索贴纸数据。

当前项目 fork 原始 `moe-sticker-bot`，早期目标偏低内存 Fly.io 部署；当前生产基线已调整为 Docker Compose + 宝塔面板部署。

## 2. 当前阶段

最近一个已验收阶段：

- 合并上游 Custom Emoji 自动拆包修复阶段
- 交付包：`docs/phase-deliveries/2026-07-13-upstream-custom-emoji-split-merge.md`

当前稳定基线：

- Docker Compose 运行 `chiaki-sticker-bot`；
- `network_mode: host`；
- 容器内部 nginx 监听宿主机 `127.0.0.1:18080`；
- 宝塔反代 `/` 到 `http://127.0.0.1:18080`；
- 公网域名：`https://stickerbot.tsmoe.com`；
- Webhook：`https://stickerbot.tsmoe.com/webhook`；
- WebApp：`https://stickerbot.tsmoe.com/webapp`；
- 宿主机 MySQL：`127.0.0.1:3306`；
- 当前数据库名：`stickerbot`；
- 本地 MySQL 使用 `DB_TLS_CONFIG=false`；
- Telegram 命令菜单已注册；
- `.tgs -> gif` 默认使用 `MSB_TGS_GIF_BACKEND=auto`，优先 lottie-converter + gifski，失败回退 rlottie-python；
- Telegram bot 与 WebApp 面向用户的中文交互文案已切换为简体中文。
- 新建超过 200 个项目的 Custom Emoji 时，会按每包最多 200 个自动拆分并返回全部分包。

更完整状态见：`DEPLOYMENT_STATUS.md`。

## 3. 目录结构速览

```text
README.md                         项目入口和推荐部署入口
DEPLOY_DOCKER_COMPOSE.md          Docker Compose + 宝塔部署手册
DEPLOYMENT_STATUS.md              当前已验证部署状态
.env.example                      环境变量模板，不含真实 secrets
docker-compose.yml                当前 Compose 部署定义
start-bot.sh                      容器启动脚本
scripts/init-db-schema.sh          数据库 schema 初始化脚本
third_party/lottie-converter/       TGS 透明 GIF 新后端 vendored 源码
web/nginx/fly.conf                容器内部 nginx 配置
docs/HANDOFF.md                   当前接棒入口
docs/PROTOCOL.md                  阶段性交付包协议
docs/phase-deliveries/            已验收阶段的正式交付包
docs/evidence/                    验证证据，可选
docs/runbooks/                    人工复现步骤，可选
docs/decisions/                   长期决策，可选
```

## 4. 已完成阶段性交付包

- `2026-07-03-docker-compose-bt-webapp-deployment.md`：完成 Docker Compose + 宝塔 + 本地 MySQL + WebApp 部署，建立当前稳定生产基线。
- `2026-07-07-db-schema-init-script.md`：新增幂等数据库 schema 初始化脚本，替代部署文档中的手写 SQL。
- `2026-07-07-tgs-transparent-gif-backend.md`：新增 `.tgs -> gif` 双线路后端，优先 lottie-converter + gifski，解决官方 TGS 转 GIF 黑底问题。
- `2026-07-08-user-facing-simplified-chinese.md`：将 Telegram bot 与 WebApp 用户交互中文文案从繁体切换为简体，保持功能逻辑不变。
- `2026-07-13-upstream-custom-emoji-split-merge.md`：合并上游 Custom Emoji 超过 200 个项目时自动拆包的修复，并保留本项目简体中文提示。

## 5. 当前待解决问题

- 需要甲方在 Telegram 客户端人工确认：
  - `/start` 是否稳定回复；
  - 左下角命令菜单是否刷新出现；
  - `/manage` 打开的 WebApp 是否能正常使用。
- 数据库 schema 初始化脚本已补充并验收：`scripts/init-db-schema.sh`。
- TGS 透明 GIF 后端已补充并验收：`MSB_TGS_GIF_BACKEND=auto`。
- 用户交互中文简体化已补充并验收：`docs/phase-deliveries/2026-07-08-user-facing-simplified-chinese.md`。
- Custom Emoji 自动拆包修复已合并并通过全仓库 Go 测试；仍需甲方在真实 Telegram 环境验证超过 200 个项目的导入。
- 后续可补充 smoke test 脚本，自动验证 health、webhook info、commands、WebApp 静态资源。

## 6. 接棒契约

任何新单智能体 / 专家团队 / 下一任乙方接手后，必须遵守：

1. 甲方不满意时，继续推进；
2. 智能体只承担自己力所能及、可稳定执行的验证；
3. 人类更好操作或更适合判断的验证，应明确交给人类；
4. 甲方满意时，生成阶段性交付包；
5. 阶段性交付包必须说明原始需求、验收结论、最终方案、当前状态、新增能力、操作步骤、验证方式、踩坑边界和接手说明；
6. 更新本接棒入口；
7. 不把中间混乱过程当成交付物；
8. 不把 runtime 数据、secrets、tokens、生成缓存提交进仓库。

## 7. 按需阅读索引

新团队不需要一开始读完所有材料。建议顺序：

1. 想了解当前生产状态：读 `DEPLOYMENT_STATUS.md`；
2. 想复现部署：读 `DEPLOY_DOCKER_COMPOSE.md`；
3. 想了解最近交付结果：读 `docs/phase-deliveries/2026-07-13-upstream-custom-emoji-split-merge.md`；
4. 想了解框架契约：读 `docs/PROTOCOL.md`；
5. 想继续开发：从 `cmd/moe-sticker-bot/main.go`、`core/init.go`、`core/webapp.go`、`pkg/msbimport/` 开始。
