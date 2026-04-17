FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o server .

FROM alpine:3.21

LABEL maintainer="newfarry@126.com"

ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata ca-certificates wget \
    && ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=builder /build/server ./server
COPY --from=builder /build/resource ./resource/
COPY --from=builder /build/config.docker.yaml ./config.docker.yaml

EXPOSE 8888
ENTRYPOINT ["./server", "-c", "config.docker.yaml"]
