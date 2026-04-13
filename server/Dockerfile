FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o server .

FROM alpine:3.21

LABEL maintainer="newfarry@126.com"

ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata ca-certificates wget python3 py3-pypdf \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
    && PYPDF_SRC="$(python3 -c 'import site; print(site.getsitepackages()[0])')" \
    && printf 'LIGHTNINGRAG_PYTHON=/usr/bin/python3\nLIGHTNINGRAG_PYPDF_SRC=%s\n' "$PYPDF_SRC" > /etc/lightningrag-pypdf.env

WORKDIR /app

COPY --from=builder /build/server ./server
COPY --from=builder /build/resource ./resource/
COPY --from=builder /build/config.docker.yaml ./config.docker.yaml

EXPOSE 8888
ENTRYPOINT ["/bin/sh", "-c", "set -a && [ -f /etc/lightningrag-pypdf.env ] && . /etc/lightningrag-pypdf.env; set +a && exec ./server -c config.docker.yaml"]
