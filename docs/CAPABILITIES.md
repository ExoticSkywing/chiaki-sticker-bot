# 项目能力清单

本文件沉淀项目已经具备的能力，作为长期能力地图。它不是阶段性交付包的复印件；每条能力只保留可被下一团队快速理解和复用的核心信息。

## 已稳定能力

<!-- LPD:CAPABILITIES:START -->
- Docker Compose + 宝塔完整部署：当前生产基线使用 `network_mode: host`，宝塔反代 `/` 到 `http://127.0.0.1:18080`，WebApp 和 webhook 均已跑通；详情见 `docs/phase-deliveries/2026-07-03-docker-compose-bt-webapp-deployment.md`。
- 本地/宝塔 MySQL schema 初始化：可用 `scripts/init-db-schema.sh` 从 `.env` 读取 `DB_ADDR`、`DB_USER`、`DB_NAME`、`DB_PASS` 并幂等初始化所需表和 `DB_VER=7`；详情见 `docs/phase-deliveries/2026-07-07-db-schema-init-script.md`。
- TGS 透明 GIF 转换：`.tgs -> gif` 默认 `MSB_TGS_GIF_BACKEND=auto`，优先 lottie-converter + gifski，失败回退 rlottie-python，已解决官方 TGS 转 GIF 黑底问题；详情见 `docs/phase-deliveries/2026-07-07-tgs-transparent-gif-backend.md`。
- 用户交互中文简体化：Telegram bot 回复/按钮/进度/错误提示与 WebApp 用户可见中文已切换为简体，功能逻辑不变；详情见 `docs/phase-deliveries/2026-07-08-user-facing-simplified-chinese.md`。
<!-- LPD:CAPABILITIES:END -->

## 待验证能力

- 暂无

## 已知限制

- 暂无

## 维护规则

1. 用户验收一个阶段后，把本阶段新增或稳定的能力补到本文件；
2. 每条能力尽量保持一行到数行，说明“现在能做什么、状态如何、详情见哪里”；
3. 详细过程放在 `docs/phase-deliveries/`，不要复制到这里；
4. 不记录 secrets、tokens、runtime 数据、生成缓存。
