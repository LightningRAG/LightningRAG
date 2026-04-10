FROM golang:alpine as builder

WORKDIR /go/src/github.com/LightningRAG/LightningRAG/server
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:latest

LABEL MAINTAINER="newfarry@126.com"
# 设置时区
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd wget \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /go/src/github.com/LightningRAG/LightningRAG/server

COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/server ./
COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/resource ./resource/
COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/config.docker.yaml ./

# 挂载目录：如果使用了sqlite数据库，容器命令示例：docker run -d -v /宿主机路径/lrag.db:/go/src/github.com/LightningRAG/LightningRAG/server/lrag.db -p 8888:8888 --name lightningrag-server-v1 lightningrag-server:1.0
# VOLUME ["/go/src/github.com/LightningRAG/LightningRAG/server"]

EXPOSE 8888
ENTRYPOINT ./server -c config.docker.yaml
