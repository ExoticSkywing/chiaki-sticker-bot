<!-- LPD:CLAUDE:RULES:START -->
# loop-phase-delivery 强制规则

本项目使用 loop-phase-delivery 轻量阶段交付与接棒系统。

## 接手项目先读

Claude Code 接手本项目的非 quick task 时，必须先读：

1. `docs/HANDOFF.md`；
2. `docs/CAPABILITIES.md`；
3. 最新 1-2 个 `docs/phase-deliveries/*.md`，如果已经存在；
4. `docs/PROTOCOL.md`，当你不确定阶段收口规则时。

不要跳过“先读项目规则”直接执行任务。

## 验收触发

用户出现以下语义时，视为上一阶段已验收：

- “满意”；
- “OK”；
- “可以了”；
- “验证通过”；
- “先把前面的落盘”；
- “在加之前，我对之前所做的感到满意”；
- 其他明确表示当前阶段已经达到预期的说法。

一旦触发验收，必须先收口上一阶段，不要只做口头记录。

## 强制执行顺序

如果用户一边表达满意、一边提出下一步需求，顺序必须是：

1. 先生成 / 补齐上一阶段简洁交付包；
2. 更新 `docs/HANDOFF.md` 的最近阶段索引；
3. 更新 / 补齐 `docs/CAPABILITIES.md` 的能力清单；
4. 汇报交付包路径、HANDOFF 更新、能力清单更新和验证结果；
5. 再继续下一阶段需求。

首选可复制命令（不依赖全局安装）：

```bash
npx --yes github:ExoticSkywing/loop-phase-delivery status --tool claude
npx --yes github:ExoticSkywing/loop-phase-delivery close --title "<阶段名>" --summary "<验收结论>"
```

如果已经全局安装，也可以使用：

```bash
lpd status --tool claude
lpd close --title "<阶段名>" --summary "<验收结论>"
```
<!-- LPD:CLAUDE:RULES:END -->
