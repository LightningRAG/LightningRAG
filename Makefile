SHELL = /bin/bash

#SCRIPT_DIR         = $(shell pwd)/etc/script
# Go version should match go.mod (currently 1.24)
BUILD_IMAGE_SERVER  = golang:1.24
# Node version should match CI (currently 20)
BUILD_IMAGE_WEB     = node:20
#项目名称
PROJECT_NAME        = github.com/LightningRAG/LightningRAG/server
#配置文件目录
CONFIG_FILE         = config.yaml
#镜像仓库命名空间
IMAGE_NAME          = lightningrag
#镜像地址
REPOSITORY          = registry.cn-hangzhou.aliyuncs.com/${IMAGE_NAME}
#镜像版本
TAGS_OPT           ?= latest
PLUGIN             ?= email

#容器环境前后端共同打包
build: build-web build-server
	docker run --name build-local --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE_SERVER} make build-local

#容器环境打包前端
build-web:
	docker run --name build-web-local --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE_WEB} make build-web-local

#容器环境打包后端
build-server:
	docker run --name build-server-local --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE_SERVER} make build-server-local

#构建web镜像
build-image-web:
	@cd web/ && docker build -t ${REPOSITORY}/web:${TAGS_OPT} .

#构建server镜像
build-image-server:
	@cd server/ && docker build -t ${REPOSITORY}/server:${TAGS_OPT} .

#本地环境打包前后端
build-local:
	if [ -d "build" ];then rm -rf build; else echo "OK!"; fi \
	&& if [ -f "/.dockerenv" ];then echo "OK!"; else  make build-web-local && make build-server-local; fi \
	&& mkdir build && cp -r web/dist build/ && cp server/server build/ && cp -r server/resource build/resource

# Build frontend locally (uses npm; CI uses pnpm — both read package-lock.json / pnpm-lock.yaml)
build-web-local:
	@cd web/ && if [ -d "dist" ];then rm -rf dist; else echo "OK!"; fi \
	&& npm install --no-audit --no-fund && npm run build

# Build backend locally; set GOPROXY env var to override the default proxy if needed
build-server-local:
	@cd server/ && if [ -f "server" ];then rm -rf server; else echo "OK!"; fi \
	&& CGO_ENABLED=0 go mod tidy \
	&& CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -B 0x$(shell head -c8 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${TAGS_OPT}" -v

# 将 web/dist 同步到 server/webui/webdist（供 go:embed，需在 yarn build 之后执行）
sync-web-dist:
	@bash scripts/sync-web-dist.sh

# 构建前端并同步后编译后端（单二进制内含前端静态资源）
build-server-embed-local: build-web-local sync-web-dist
	@cd server/ && if [ -f "server" ];then rm -rf server; else echo "OK!"; fi \
	&& CGO_ENABLED=0 go mod tidy \
	&& CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -B 0x$(shell head -c8 /dev/urandom|od -An -tx1|tr -d ' \n') -X main.Version=${TAGS_OPT}" -v

#打包前后端二合一镜像
image: build
	docker build -t ${REPOSITORY}/lightningrag:${TAGS_OPT} -f deploy/docker/Dockerfile .

#尝鲜版
images: build build-image-web build-image-server
	docker build -t ${REPOSITORY}/all:${TAGS_OPT} -f deploy/docker/Dockerfile .

#swagger 文档生成
doc:
	@cd server && swag init

#插件快捷打包： make plugin PLUGIN="这里是插件文件夹名称,默认为email"
plugin:
	if [ -d ".plugin" ];then rm -rf .plugin ; else echo "OK!"; fi && mkdir -p .plugin/${PLUGIN}/{server/plugin,web/plugin} \
	&& if [ -d "server/plugin/${PLUGIN}" ];then cp -r server/plugin/${PLUGIN} .plugin/${PLUGIN}/server/plugin/ ; else echo "OK!"; fi \
	&& if [ -d "web/src/plugin/${PLUGIN}" ];then cp -r web/src/plugin/${PLUGIN} .plugin/${PLUGIN}/web/plugin/ ; else echo "OK!"; fi \
	&& cd .plugin && zip -r ${PLUGIN}.zip ${PLUGIN} && mv ${PLUGIN}.zip ../ && cd ..
