# Stage 1: Build React WebApp
FROM node:18-bookworm-slim AS webapp-builder

WORKDIR /webapp
COPY web/webapp3/package.json web/webapp3/package-lock.json ./
RUN npm ci
COPY web/webapp3/ ./
RUN PUBLIC_URL=/webapp npm run build

# Stage 2: Build Go binary
FROM golang:1.22-bookworm AS go-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /moe-sticker-bot ./cmd/moe-sticker-bot/main.go

# Stage 3: Build gifski for higher-quality TGS -> GIF conversion
FROM rust:1.95-bookworm AS gifski-builder
RUN cargo install --version 1.32.0 gifski

# Stage 4: Build lottie_to_png used by lottie_to_gif.sh
FROM gcc:15-bookworm AS lottie-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    python3 \
    python3-pip \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN pip3 install --break-system-packages conan==2.17.0

WORKDIR /lottie-converter
RUN conan profile detect
COPY third_party/lottie-converter/conanfile.txt ./
RUN conan install . --build=missing -s build_type=Release
COPY third_party/lottie-converter/CMakeLists.txt ./
COPY third_party/lottie-converter/src ./src
RUN cmake -DCMAKE_BUILD_TYPE=Release -DLOTTIE_MODULE=OFF CMakeLists.txt && cmake --build . --config Release

# Stage 5: Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    imagemagick \
    libarchive-tools \
    ffmpeg \
    cpulimit \
    curl \
    gifsicle \
    bash \
    python3 \
    python3-pip \
    ca-certificates \
    nginx \
    && rm -rf /var/lib/apt/lists/*

RUN pip3 install --break-system-packages rlottie-python emoji pillow

COPY tools/msb_emoji.py /usr/local/bin/msb_emoji.py
COPY tools/msb_kakao_decrypt.py /usr/local/bin/msb_kakao_decrypt.py
COPY tools/msb_rlottie.py /usr/local/bin/msb_rlottie.py
RUN chmod +x /usr/local/bin/msb_emoji.py /usr/local/bin/msb_kakao_decrypt.py /usr/local/bin/msb_rlottie.py

COPY --from=go-builder /moe-sticker-bot /usr/local/bin/moe-sticker-bot
COPY --from=webapp-builder /webapp/build /webapp/build
COPY --from=gifski-builder /usr/local/cargo/bin/gifski /usr/local/bin/gifski
COPY --from=lottie-builder /lottie-converter/bin/lottie_to_png /usr/local/bin/lottie_to_png
COPY third_party/lottie-converter/bin/lottie_common.sh /usr/local/bin/lottie_common.sh
COPY third_party/lottie-converter/bin/lottie_to_gif.sh /usr/local/bin/lottie_to_gif.sh
RUN chmod +x /usr/local/bin/lottie_common.sh /usr/local/bin/lottie_to_gif.sh

COPY web/nginx/fly.conf /etc/nginx/conf.d/default.conf
RUN rm -f /etc/nginx/sites-enabled/default

COPY start-bot.sh /usr/local/bin/start-bot.sh
RUN chmod +x /usr/local/bin/start-bot.sh

VOLUME ["/data"]

EXPOSE 8080

CMD ["/usr/local/bin/start-bot.sh"]
