# 阶段性交付包：用户交互中文简体化

## 1. 原始需求

甲方希望将项目中机器人与用户交互显示的繁体中文改为简体中文，同时明确范围只限用户可见交互文案，不扩大到历史文档、部署文档或阶段性交付记录。

本阶段目标：

- 保持现有功能逻辑不变；
- 只调整 Telegram bot 回复、按钮、进度、错误提示等用户可见文案；
- 只调整 WebApp 中用户可见的少量中文文案；
- 不影响后续同步上游，只把本地化差异控制在少量运行时文案文件中。

## 2. 验收结论

本阶段已完成并通过甲方人工验证。

甲方验收反馈：

```text
经验证，没问题，满意
```

## 3. 最终方案

### 3.1 Telegram bot 用户可见文案

已将 bot 运行时与用户交互的中文文案从繁体改为简体，涉及：

```text
core/message.go
core/init.go
core/states.go
core/sticker.go
core/sticker_download.go
core/shutdown.go
```

覆盖类型包括：

- `/start` 欢迎语；
- `/command_list` 指令说明；
- `/about`、`/faq`、`/privacy`、`/changelog` 文案；
- 导入、下载、创建、管理贴图包流程提示；
- emoji/关键字设置提示；
- 转换、上传、排队、处理进度提示；
- 错误、失败、重试、重启中断等用户提示；
- inline button 中文部分。

### 3.2 WebApp 用户可见文案

已将 WebApp 中用户可见中文改为简体，涉及：

```text
web/webapp3/src/App.js
web/webapp3/src/Edit.js
web/webapp3/src/Export.js
```

覆盖类型包括：

- WebApp 打开方式提示；
- 编辑页拖拽排序提示；
- 导出按钮与预览文案。

### 3.3 保持不变的范围

本阶段没有修改：

- 数据库 schema；
- Docker Compose / Nginx / 宝塔部署配置；
- `.tgs -> gif` 转换逻辑；
- Webhook / WebApp API 逻辑；
- README、部署文档、历史阶段交付文档；
- secrets、tokens、runtime 数据、生成缓存。

## 4. 新增能力

项目现在具备面向简体中文用户的主要交互文案：

- Telegram bot 的主要中文回复已为简体中文；
- WebApp 的中文提示已为简体中文；
- 本地化改动集中在运行时文案文件中，后续同步上游时冲突范围可控。

## 5. 上游同步影响

本阶段属于本地化文案改动，不改变业务逻辑。

后续同步上游时，只有当上游也修改同一批用户文案行时，才可能产生文本冲突。预计冲突集中在：

```text
core/message.go
core/init.go
core/states.go
core/sticker.go
core/sticker_download.go
core/shutdown.go
web/webapp3/src/App.js
web/webapp3/src/Edit.js
web/webapp3/src/Export.js
```

这类冲突通常只需要保留上游逻辑变化，同时保留本地简体中文文案即可。

## 6. 已执行验证

智能体已执行：

```bash
gofmt
```

```bash
go test ./...
```

结果：

```text
ok   github.com/star-39/moe-sticker-bot/core
ok   github.com/star-39/moe-sticker-bot/pkg/msbimport
```

```bash
docker compose config
```

结果：配置校验通过。

```bash
git diff --check
```

结果：无 whitespace 问题。

还执行了运行时文件繁体残留扫描，范围为：

```text
core/**/*.go
web/webapp3/src/**/*.js
```

结果：未发现目标繁体残留。

WebApp 使用临时 Node Docker 环境执行：

```bash
npm ci && npm run build
```

结果：构建成功。构建过程中出现原项目已有的 React/ESLint warning 与 npm audit 提示，不阻断构建，也不是本阶段文案修改引入的失败。

## 7. 人类验证

甲方已在实际使用中验证：

- bot 与 WebApp 用户交互文案无问题；
- 当前阶段结果满意。

## 8. 踩坑边界

1. 不建议对整个仓库做全量繁简转换，否则会污染历史文档和阶段交付记录。
2. 不应转换代码标识符、URL、常量名，例如 `KAKAO`、`FID_KAKAO_SHARE_LINK`、`WARN_KAKAO_PREFER_SHARE_LINK`。
3. 后续从 upstream 合并时，如果上游改同一段英文/中文双语文案，需要手动保留简体中文结果。
4. WebApp 构建时的 ESLint warning 和 npm audit 输出属于现有依赖/代码提示，不等同于本阶段失败。

## 9. 接手说明

下一任接手用户文案问题时：

- 优先查看 `core/message.go`；
- 其次查看 `core/init.go`、`core/states.go`、`core/sticker.go`、`core/sticker_download.go`、`core/shutdown.go`；
- WebApp 用户文案查看 `web/webapp3/src/App.js`、`web/webapp3/src/Edit.js`、`web/webapp3/src/Export.js`；
- 不要把 docs/phase-deliveries 下的历史交付记录当成本阶段必须继续转换的运行时文案。
