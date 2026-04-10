FROM golang:alpine as builder

WORKDIR /go/src/github.com/LightningRAG/LightningRAG/server
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -o server .

FROM alpine:3.21

LABEL MAINTAINER="newfarry@126.com"
# 设置时区；PDF 默认解析链优先 pypdf（需 Python + pypdf，见 deploy/DOCKER_PYTHON_PYPDF_zh.md）
ENV TZ=Asia/Shanghai
RUN apk update && apk add --no-cache tzdata openntpd wget python3 py3-pypdf \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
    && PYPDF_SRC="$(python3 -c 'import site; print(site.getsitepackages()[0])')" \
    && printf 'LIGHTNINGRAG_PYTHON=/usr/bin/python3\nLIGHTNINGRAG_PYPDF_SRC=%s\n' "$PYPDF_SRC" > /etc/lightningrag-pypdf.env

WORKDIR /go/src/github.com/LightningRAG/LightningRAG/server

COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/server ./
COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/resource ./resource/
COPY --from=0 /go/src/github.com/LightningRAG/LightningRAG/server/config.docker.yaml ./

# 挂载目录：如果使用了sqlite数据库，容器命令示例：docker run -d -v /宿主机路径/lrag.db:/go/src/github.com/LightningRAG/LightningRAG/server/lrag.db -p 8888:8888 --name lightningrag-server-v1 lightningrag-server:1.0
# VOLUME ["/go/src/github.com/LightningRAG/LightningRAG/server"]

EXPOSE 8888
ENTRYPOINT ["/bin/sh", "-c", "set -a && [ -f /etc/lightningrag-pypdf.env ] && . /etc/lightningrag-pypdf.env; set +a && exec ./server -c config.docker.yaml"]
