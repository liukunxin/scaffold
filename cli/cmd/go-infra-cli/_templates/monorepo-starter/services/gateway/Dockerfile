# 构建上下文必须是仓库根目录，而不是本目录 —— 构建要能看见 go.work 与 packages/。
#
#   docker build -f services/gateway/Dockerfile -t app:dev .
#   docker run --rm -p 8080:8080 -e env=docker app:dev
#   docker run --rm -p 9091:9091 -e env=docker app:dev /app/app-grpc
#
# HTTP 与 gRPC 是两个独立入口（cmd/http、cmd/grpc），镜像里两个二进制都带着，
# 默认起 HTTP，要起 gRPC 就覆盖 ENTRYPOINT 参数。
#
# 配置按 env 分层加载：configs/config.yml 打底，configs/config.docker.yml 覆盖，
# 容器内默认 env=docker；要换环境就 -e env=<local|test|gray|prod>。
#
# 上下文是整个仓库，仓库根的 .dockerignore 负责把无关目录挡在构建缓存之外。
#
# 注意：go-infra-cli mono add 复用本文件作为新 Project 的 Dockerfile 模板，
# 生成时只替换「本服务所在目录」这一段路径前缀。改这里等于改所有新生成的 Project。

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY . .
RUN cd services/gateway \
 && CGO_ENABLED=0 go build -trimpath -o /out/app-http ./cmd/http \
 && CGO_ENABLED=0 go build -trimpath -o /out/app-grpc ./cmd/grpc

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/app-http /app/app-http
COPY --from=build /out/app-grpc /app/app-grpc
COPY services/gateway/configs /app/configs
EXPOSE 8080 9091
ENV env=docker
ENTRYPOINT ["/app/app-http"]
