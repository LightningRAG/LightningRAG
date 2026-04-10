# 如果需要用 cicd ，请设置环境变量：
# variables:
#     DOCKER_BUILDKIT: 1

FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
COPY . /app
WORKDIR /app


FROM base AS prod-deps
# --ignore-scripts：postinstall 依赖 dev 的 patch-package，生产依赖安装阶段会失败
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --ignore-scripts

FROM base AS build
COPY --from=prod-deps /app/node_modules /app/node_modules
# DOCKER_BUILD=1 触发 vite 使用 esbuild 压缩 + 降低 Rollup 并行，避免 terser 在容器内 OOM(137)
# 堆上限与 Docker Desktop 分配给 builder 的内存匹配；仍 OOM 时可改为 2048 或增大 Desktop 内存
ENV DOCKER_BUILD=1
ENV NODE_OPTIONS="--max-old-space-size=3072"
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install && pnpm run build


FROM nginx:alpine
LABEL MAINTAINER="bypanghu@163.com"
COPY --from=base  /app/.docker-compose/nginx/conf.d/my.conf /etc/nginx/conf.d/my.conf
COPY --from=build  /app/dist /usr/share/nginx/html
RUN ls -al /usr/share/nginx/html
