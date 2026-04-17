# syntax=docker/dockerfile:1

FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml ./

FROM base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --ignore-scripts

FROM base AS build
COPY --from=prod-deps /app/node_modules /app/node_modules
COPY . .
ENV DOCKER_BUILD=1
ENV NODE_OPTIONS="--max-old-space-size=3072"
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install && pnpm run build

FROM nginx:alpine
LABEL maintainer="newfarry@126.com"
COPY --from=build /app/.docker-compose/nginx/conf.d/my.conf /etc/nginx/conf.d/my.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
