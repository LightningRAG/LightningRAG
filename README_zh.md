# LightningRAG

[English](./README.md) | 简体中文

## 重要提示

1. 本项目需要您有一定的 golang 和 Vue 基础。

2. 如果您将此项目用于商业用途，请遵守 Apache 2.0 协议并保留作者技术支持声明。您需保留如下版权声明信息，以及日志和代码中所包含的版权声明信息。所需保留信息均为文案性质，不会影响任何业务内容，如决定商用【产生收益的商业行为均在商用行列】或者必须剔除请[购买授权](https://plugin.LightningRAG.com/licenseindex.html)。

## 1. 基本介绍

### 1.1 项目介绍

> LightningRAG 是一个基于 [vue](https://vuejs.org) 和 [gin](https://gin-gonic.com) 开发的全栈前后端分离的开发基础平台，集成jwt鉴权，动态路由，动态菜单，casbin鉴权，表单生成器，代码生成器等功能，提供多种示例文件，让您把更多时间专注在业务开发上。

[在线预览](https://demo.LightningRAG.com): https://demo.LightningRAG.com

测试用户名：admin

测试密码：123456

### 1.2 贡献指南
Hi! 首先感谢你使用 LightningRAG。

LightningRAG 是一套为快速研发准备的一整套前后端分离架构式的开源框架，旨在快速搭建中小型项目。

LightningRAG 的成长离不开大家的支持，如果你愿意为 LightningRAG 贡献代码或提供建议，请阅读以下内容。

#### 1.2.1 Issue 规范
- issue 仅用于提交 Bug 或 Feature 以及设计相关的内容，其它内容可能会被直接关闭。
									      
- 在提交 issue 之前，请搜索相关内容是否已被提出。

#### 1.2.2 Pull Request 规范
- 请先 fork 一份到自己的项目下，不要直接在仓库下建分支。

- commit 信息要以`[文件名]: 描述信息` 的形式填写，例如 `README.md: fix xxx bug`。

- 如果是修复 bug，请在 PR 中给出描述信息。

- 合并代码需要两名维护人员参与：一人进行 review 后 approve，另一人再次 review，通过后即可合并。

## 2. 使用说明

```
- node版本 > v18.16.0
- golang版本 >= v1.22
- IDE推荐：Goland
```

### 2.1 server项目

使用 `Goland` 等编辑工具，打开server目录，不可以打开 LightningRAG 根目录

```bash

# 克隆项目
git clone https://github.com/LightningRAG/LightningRAG.git
# 进入server文件夹
cd server

# 使用 go mod 并安装go依赖包
go generate

# 运行
go run . 

```

### 2.2 web项目

```bash
# 进入web文件夹
cd web

# 安装依赖
npm install

# 启动web项目
npm run serve
```

### 2.3 swagger自动化API文档

#### 2.3.1 安装 swagger

``` shell
go install github.com/swaggo/swag/cmd/swag@latest
```

#### 2.3.2 生成API文档

```` shell
cd server
swag init
````

> 执行上面的命令后，server目录下会出现docs文件夹里的 `docs.go`, `swagger.json`, `swagger.yaml` 三个文件更新，启动go服务之后, 在浏览器输入 [http://localhost:8888/swagger/index.html](http://localhost:8888/swagger/index.html) 即可查看swagger文档

### 2.4 VSCode工作区

#### 2.4.1 开发

使用`VSCode`打开根目录下的工作区文件`LightningRAG.code-workspace`，在边栏可以看到三个虚拟目录：`backend`、`frontend`、`root`。

#### 2.4.2 运行/调试

在运行和调试中也可以看到三个task：`Backend`、`Frontend`、`Both (Backend & Frontend)`。运行`Both (Backend & Frontend)`可以同时启动前后端项目。

#### 2.4.3 settings

在工作区配置文件中有`go.toolsEnvVars`字段，是用于`VSCode`自身的go工具环境变量。此外在多go版本的系统中，可以通过`gopath`、`go.goroot`指定运行版本。

```json
    "go.gopath": null,
    "go.goroot": null,
```

### 2.5 其他说明文档

- [第三方平台对话接入（Webhook / 渠道连接器）](docs/THIRD_PARTY_CHANNEL_CONNECTORS_zh.md)

### 2.6 单二进制嵌入前端（go:embed）

将 `web` 构建产物同步到 `server/webui/webdist` 后再编译 Go，可把静态前端打进 **同一个可执行文件**，便于只发布一个二进制。

1. **构建并编译（推荐，在仓库根目录）**

   ```bash
   make build-server-embed-local
   ```

   或：

   ```bash
   bash scripts/build-server-with-embed.sh
   ```

   上述流程会执行 `build-web-local`（`yarn install` + `yarn build`）、`scripts/sync-web-dist.sh`，再在 `server/` 下 `go build`。

2. **仅同步已有 `web/dist`（已手动 `yarn build` 时）**

   ```bash
   make sync-web-dist
   # 或
   bash scripts/sync-web-dist.sh
   ```

3. **启用内置前端**

   在 `config.yaml`（或对应环境的配置文件）中将 `system.embed-web-ui` 设为 `true`。默认 `false`，与「Nginx 托管静态资源 + 后端 API」的分离部署保持一致。

   当 `embed-web-ui: true` 且 `system.router-prefix` 为空时，在 **HTTP 层**（Gin 匹配路由之前）将 `/api/...` 改为 `/...`，与 `web/.env.production` 的 `VITE_BASE_API=/api` 及 Nginx `rewrite ^/api/(.*)$ /$1` 一致（若仅用 Gin 中间件改写 Path，会在未匹配路由时无效）。若 `router-prefix` 非空，不会自动剥离 `/api`，需自行对齐前端 `VITE_BASE_API`。

4. **运行**

   使用带嵌入资源编译出的 `server` 二进制，在 `embed-web-ui: true` 的配置下启动；浏览器访问后端监听端口即可打开管理端（例如 `http://127.0.0.1:8888`），Swagger 仍为 `http://127.0.0.1:8888/swagger/index.html`。

### 2.7 GoReleaser 多平台发布（GitHub Releases）

使用仓库根目录的 [GoReleaser](https://goreleaser.com/) 配置 **`.goreleaser.yaml`**，在发布流程中自动完成「前端构建 → 同步到 `webui/webdist` → 交叉编译」，与上一节的手动嵌入流程一致。

| 步骤 | 说明 |
|------|------|
| 编译前钩子 | 在 `web/` 执行 `npm install` 与 `npm run build`，再执行 `scripts/sync-web-dist.sh`，将产物写入 `server/webui/webdist` 供 `go:embed` |
| Go 模块 | `go.mod` 位于 `server/`，配置中通过 `gomod.dir: server` 指定，避免在仓库根目录执行 `go list -m` 失败 |
| 构建参数 | `CGO_ENABLED=0`、`-trimpath`、`-ldflags "-s -w"`；目标平台见 `.goreleaser.yaml`（含 Linux / Windows / macOS / FreeBSD 及常见 amd64、arm64、386、Linux armv7、Windows arm64 等） |
| 发布包内容 | 每个压缩包内含可执行文件 `lightningrag`、根目录 **`config.yaml`**（由 `server/config.docker.yaml` 复制而来）、以及 **`resource/`**（来自 `server/resource`） |
| 自动化 | 推送 **`v*`** 语义化标签时由 **`.github/workflows/goreleaser.yml`** 执行 `goreleaser release --clean` 并发布到 GitHub Releases（依赖 `GITHUB_TOKEN` 的 `contents: write`） |

**本地试打包（不上传 Release）：**

```bash
goreleaser release --snapshot --clean --skip=publish
```

产物默认在 **`dist/`** 目录（建议勿提交到 Git）。

**正式发布：** 打标签并推送，例如 `git tag v2.9.1 && git push origin v2.9.1`。

更多说明见 **`.github/workflows/README_zh.md`** 中「GoReleaser」一节。

## 3. 技术选型

- 前端：用基于 [Vue](https://vuejs.org) 的 [Element](https://github.com/ElemeFE/element) 构建基础页面。
- 后端：用 [Gin](https://gin-gonic.com/) 快速搭建基础restful风格API，[Gin](https://gin-gonic.com/) 是一个go语言编写的Web框架。
- 数据库：采用`MySql` > (5.7) 版本 数据库引擎 InnoDB，使用 [gorm](http://gorm.cn) 实现对数据库的基本操作。
- 缓存：使用`Redis`实现记录当前活跃用户的`jwt`令牌并实现多点登录限制。
- API文档：使用`Swagger`构建自动化文档。
- 配置文件：使用 [fsnotify](https://github.com/fsnotify/fsnotify) 和 [viper](https://github.com/spf13/viper) 实现`yaml`格式的配置文件。
- 日志：使用 [zap](https://github.com/uber-go/zap) 实现日志记录。

## 4. 项目架构

### 4.1 目录结构

```
    ├── server
        ├── api             (api层)
        │   └── v1          (v1版本接口)
        ├── config          (配置包)
        ├── core            (核心文件)
        ├── docs            (swagger文档目录)
        ├── global          (全局对象)                    
        ├── initialize      (初始化)                        
        │   └── internal    (初始化内部函数)                            
        ├── middleware      (中间件层)                        
        ├── model           (模型层)                    
        │   ├── request     (入参结构体)                        
        │   └── response    (出参结构体)                            
        ├── packfile        (静态文件打包)                        
        ├── resource        (静态资源文件夹)                        
        │   ├── excel       (excel导入导出默认路径)                        
        │   ├── page        (表单生成器)                        
        │   └── template    (模板)                            
        ├── router          (路由层)                    
        ├── service         (service层)                    
        ├── source          (source层)                    
        └── utils           (工具包)                    
            ├── timer       (定时器接口封装)                        
            └── upload      (oss接口封装)                        
    
            web
        ├── babel.config.js
        ├── Dockerfile
        ├── favicon.ico
        ├── index.html                 -- 主页面
        ├── limit.js                   -- 助手代码
        ├── package.json               -- 包管理器代码
        ├── src                        -- 源代码
        │   ├── api                    -- api 组
        │   ├── App.vue                -- 主页面
        │   ├── assets                 -- 静态资源
        │   ├── components             -- 全局组件
        │   ├── core                   -- lrag 组件包
        │   │   ├── config.js          -- lrag网站配置文件
        │   │   ├── lightningrag.js   -- 注册欢迎文件
        │   │   └── global.js          -- 统一导入文件
        │   ├── directive              -- v-auth 注册文件
        │   ├── main.js                -- 主文件
        │   ├── permission.js          -- 路由中间件
        │   ├── pinia                  -- pinia 状态管理器，取代vuex
        │   │   ├── index.js           -- 入口文件
        │   │   └── modules            -- modules
        │   │       ├── dictionary.js
        │   │       ├── router.js
        │   │       └── user.js
        │   ├── router                 -- 路由声明文件
        │   │   └── index.js
        │   ├── style                  -- 全局样式
        │   │   ├── base.scss
        │   │   ├── basics.scss
        │   │   ├── element_visiable.scss  -- 此处可以全局覆盖 element-plus 样式
        │   │   ├── iconfont.css           -- 顶部几个icon的样式文件
        │   │   ├── main.scss
        │   │   ├── mobile.scss
        │   │   └── newLogin.scss
        │   ├── utils                  -- 方法包库
        │   │   ├── asyncRouter.js     -- 动态路由相关
        │   │   ├── btnAuth.js         -- 动态权限按钮相关
        │   │   ├── bus.js             -- 全局mitt声明文件
        │   │   ├── date.js            -- 日期相关
        │   │   ├── dictionary.js      -- 获取字典方法 
        │   │   ├── downloadImg.js     -- 下载图片方法
        │   │   ├── format.js          -- 格式整理相关
        │   │   ├── image.js           -- 图片相关方法
        │   │   ├── page.js            -- 设置页面标题
        │   │   ├── request.js         -- 请求
        │   │   └── stringFun.js       -- 字符串文件
        |   ├── view -- 主要view代码
        |   |   ├── about -- 关于我们
        |   |   ├── dashboard -- 面板
        |   |   ├── error -- 错误
        |   |   ├── example --上传案例
        |   |   ├── iconList -- icon列表
        |   |   ├── init -- 初始化数据  
        |   |   |   ├── index -- 新版本
        |   |   |   ├── init -- 旧版本
        |   |   ├── layout  --  layout约束页面 
        |   |   |   ├── aside 
        |   |   |   ├── bottomInfo     -- bottomInfo
        |   |   |   ├── screenfull     -- 全屏设置
        |   |   |   ├── setting        -- 系统设置
        |   |   |   └── index.vue      -- base 约束
        |   |   ├── login              --登录 
        |   |   ├── person             --个人中心 
        |   |   ├── superAdmin         -- 超级管理员操作
        |   |   ├── system             -- 系统检测页面
        |   |   ├── systemTools        -- 系统配置相关页面
        |   |   └── routerHolder.vue   -- page 入口页面 
        ├── vite.config.js             -- vite 配置文件
        └── yarn.lock

```

## 5. 主要功能

- 权限管理：基于`jwt`和`casbin`实现的权限管理。
- 文件上传下载：实现基于`七牛云`, `阿里云`, `腾讯云` 的文件上传操作(请开发自己去各个平台的申请对应 `token` 或者对应`key`)。
- 分页封装：前端使用 `mixins` 封装分页，分页方法调用 `mixins` 即可。
- 用户管理：系统管理员分配用户角色和角色权限。
- 角色管理：创建权限控制的主要对象，可以给角色分配不同api权限和菜单权限。
- 菜单管理：实现用户动态菜单配置，实现不同角色不同菜单。
- api管理：不同用户可调用的api接口的权限不同。
- 配置管理：配置文件可前台修改(在线体验站点不开放此功能)。
- 条件搜索：增加条件搜索示例。
- restful示例：可以参考用户管理模块中的示例API。
	- 前端文件参考: [web/src/view/superAdmin/api/api.vue](https://github.com/LightningRAG/LightningRAG/blob/master/web/src/view/superAdmin/api/api.vue)
    - 后台文件参考: [server/router/sys_api.go](https://github.com/LightningRAG/LightningRAG/blob/master/server/router/sys_api.go)
- 多点登录限制：需要在`config.yaml`中把`system`中的`use-multipoint`修改为true(需要自行配置Redis和Config中的Redis参数，测试阶段，有bug请及时反馈)。
- 分片上传：提供文件分片上传和大文件分片上传功能示例。
- 表单生成器：表单生成器借助 [@Variant Form](https://github.com/vform666/variant-form) 。
- 代码生成器：后台基础逻辑以及简单curd的代码生成器。

## 6. 贡献者

感谢您对 LightningRAG 的贡献。完整列表见 [GitHub Contributors](https://github.com/LightningRAG/LightningRAG/graphs/contributors)。

## 7. 注意事项

请严格遵守Apache 2.0协议并保留作品声明，去除版权信息请务必[获取授权](https://plugin.LightningRAG.com/license)  
未授权去除版权信息将依法追究法律责任
