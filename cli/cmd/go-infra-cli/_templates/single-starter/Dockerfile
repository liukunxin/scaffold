# 单服务项目的运行镜像。构建上下文 = 本项目根目录（这里没有 go.work，也没有同仓模块）。
#
#   docker build -t myapp:dev .
#   docker run --rm -p 8080:8080 -e env=docker myapp:dev
#   docker run --rm -p 9091:9091 -e env=docker myapp:dev /app/app-grpc
#
# HTTP 与 gRPC 是两个独立入口（cmd/http、cmd/grpc），镜像里两个二进制都带着，
# 默认起 HTTP，要起 gRPC 就覆盖 ENTRYPOINT 参数。
#
# 配置按 env 分层加载：configs/config.yml 打底，configs/config.docker.yml 覆盖，
# 所以容器内默认 env=docker；要换环境就 -e env=<local|test|gray|prod>。

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app-http ./cmd/http \
 && CGO_ENABLED=0 go build -trimpath -o /out/app-grpc ./cmd/grpc

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/app-http /app/app-http
COPY --from=build /out/app-grpc /app/app-grpc
COPY configs /app/configs
EXPOSE 8080 9091
ENV env=docker
ENTRYPOINT ["/app/app-http"]
