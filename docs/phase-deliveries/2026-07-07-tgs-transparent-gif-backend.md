# 阶段性交付包：TGS 透明 GIF 转换后端

## 1. 原始需求

甲方在使用过程中发现：下载 Telegram 官方 sticker set 后，原始压缩包内是 `.tgs` 文件，但转换后的 GIF 背景变成黑色，期望在保持现有功能不变的基础上，为 `.tgs -> gif` 增加一条更高质量的转换线路。

甲方提出的方向：

- 保持现有一切能力不变；
- 在关键 `.tgs -> gif` 转换点启用两条线路；
- `.tgs` 优先使用 `lottie-converter/gifski`；
- 失败时仍能回退到原有转换方式。

## 2. 验收结论

本阶段甲方已表示满意。

实际验证结果：

- 新后端成功将 `.tgs` 转为透明 GIF；
- 用户确认“直接变成了透明 gif”；
- bot 容器使用新镜像重建后保持 healthy；
- 本地 health check 返回 `ok`。

## 3. 最终方案

### 3.1 双线路后端

`.tgs -> gif` 入口仍保持：

```go
msbimport.RlottieToGIF(f)
```

内部新增后端优先级：

```text
MSB_TGS_GIF_BACKEND=auto
    先使用 lottie-converter + gifski
    失败后 fallback 到旧的 msb_rlottie.py

MSB_TGS_GIF_BACKEND=lottie-converter
    只使用 lottie-converter + gifski，失败直接报错

MSB_TGS_GIF_BACKEND=rlottie-python
    只使用旧的 rlottie-python 路径
```

默认推荐：

```dotenv
MSB_TGS_GIF_BACKEND=auto
```

### 3.2 vendored lottie-converter

新增：

```text
third_party/lottie-converter/
```

包含必要源码、脚本和 MIT LICENSE。

不依赖 `/root/data/lty/repo/test/lottie-converter` 这个本地测试目录，保证 Docker 构建可复现。

### 3.3 Docker 构建

Dockerfile 新增构建阶段：

- `gifski-builder`：构建 `gifski`；
- `lottie-builder`：构建 `lottie_to_png`；
- runtime 镜像安装：
  - `gifski`
  - `lottie_to_png`
  - `lottie_to_gif.sh`
  - `lottie_common.sh`

运行镜像中可用：

```text
/usr/local/bin/lottie_to_gif.sh
/usr/local/bin/lottie_to_png
/usr/local/bin/gifski
```

## 4. 新增能力

项目现在具备高质量 `.tgs -> gif` 转换能力：

```text
.tgs -> C++ rlottie -> RGBA PNG frames -> gifski -> transparent GIF
```

相比旧路径：

```text
.tgs -> rlottie-python -> ffmpeg GIF -> gifsicle
```

新路径对透明通道更友好，解决了 Telegram 官方 `.tgs` 转 GIF 黑底问题。

## 5. 可调参数

新增环境变量：

```dotenv
MSB_TGS_GIF_BACKEND=auto
MSB_TGS_GIF_WIDTH=512
MSB_TGS_GIF_HEIGHT=512
MSB_TGS_GIF_FPS=30
MSB_TGS_GIF_QUALITY=90
MSB_TGS_GIF_THREADS=2
```

调优建议：

- 2 vCPU 机器上 CPU 100% 是正常现象，`gifski` 和 rlottie 渲染都是 CPU 密集型；
- 如果要更稳，降低线程：

```dotenv
MSB_TGS_GIF_THREADS=1
```

- 如果要更丝滑，可以自行尝试：

```dotenv
MSB_TGS_GIF_FPS=50
```

- 不建议默认 60fps，因为会显著增加 CPU 时间和 GIF 文件体积。

## 6. 已执行验证

智能体已执行：

1. Docker 镜像构建成功；
2. 镜像内工具存在并可执行：

```bash
command -v lottie_to_gif.sh
command -v lottie_to_png
command -v gifski
```

3. 使用 lottie-converter 自带 `.tgs` 样本实际转换成功；
4. `gifsicle -I output.gif` 显示每帧包含透明信息，例如：

```text
+ image #0 256x256 transparent 0
```

5. bot 容器重建成功；
6. `docker compose ps` 显示 healthy；
7. `curl -fsS http://127.0.0.1:18080/health` 返回：

```text
ok
```

## 7. 人类验证

甲方已用实际问题贴纸验证：

- 转换结果从黑底变为透明 GIF；
- 当前工作结果满意。

后续甲方会自行调节 FPS、quality、threads 参数。

## 8. 踩坑边界

1. `gifski 1.32.0` 需要较新的 Rust/Cargo；`rust:1.83` 构建失败，已改为 `rust:1.95-bookworm`。
2. vendored `lottie_to_gif.sh` 使用 bash 的 `source`，必须有 bash shebang 并确保 runtime 安装 bash。
3. GIF 只有 1-bit 透明，边缘效果无法完全等同原始 TGS 的 alpha。
4. 60fps 可以尝试，但在 2 vCPU 服务器上会明显增加 CPU 时间和文件体积。
5. 默认 `auto` 会 fallback 到旧路径；如果需要暴露失败，使用 `MSB_TGS_GIF_BACKEND=lottie-converter`。

## 9. 接手说明

下一任接手 `.tgs -> gif` 问题时优先查看：

- `pkg/msbimport/convert_lottie.go`
- `third_party/lottie-converter/`
- `Dockerfile`
- `.env.example`
- `DEPLOY_DOCKER_COMPOSE.md`

当前阶段不要再回退到单一路径；保留 `auto` 后端可以兼顾质量和可用性。
