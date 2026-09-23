# integration-test-platform

基于 Mocha 的一体化测试平台，主要用于 Chrome 内核测试、网页测试，以及测试过程录制。

## 技术栈

- **后端**：Go + Gin（`server/`）
- **前端**：Vue 3 + Vue Router + Element Plus + Vite（`web/`）
- **脚本编辑器**：CodeMirror 6（`web/src/views/scripts/`）
- **菜单**：侧栏由 Vue Router 的 `meta.menu` 定义（见 `web/src/router/menu.js`）；[`config/menu.json`](config/menu.json) 保留与路由一致的条目，供服务端校验

布局：顶部栏 + 左侧边栏 + 主内容区（Hash 路由 SPA）。开发模式下 Go 服务托管 `web/dist/` 构建产物；发布构建将前端嵌入可执行文件。

> 前端已由 jQuery + Bootstrap（`web/pages/` 片段 + CDN）重构为上述 Vue 技术栈，旧的多页 HTML 片段方案已移除。

## 开发

需要 **Go 1.22+** 与 **Node.js 18+**（前端构建）。

### 环境要求

- 在仓库根目录工作（存在 `go.mod`）

```bash
go version
node -v
```

### 安装依赖

Go：

```bash
go mod download
```

或整理并补全依赖：

```bash
go mod tidy
```

`go.sum` 已在仓库中时，一般 `go mod download` 即可。

前端：

```bash
cd web
npm install
```

### 使用中国镜像（Go 模块）

国内访问 `proxy.golang.org` 经常超时，拉依赖前设置 `GOPROXY`。

当前终端临时生效（Git Bash / bash）：

```bash
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn
go mod download
```

Windows PowerShell：

```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "sum.golang.google.cn"
go mod download
```

长期写入 Go 配置（推荐）：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

常用国内源：

| 用途 | 地址 |
|------|------|
| 七牛 / goproxy.cn | `https://goproxy.cn,direct` |
| 阿里云 | `https://mirrors.aliyun.com/goproxy/,direct` |
| 腾讯云 | `https://mirrors.cloud.tencent.com/go/,direct` |

末尾的 `direct` 表示镜像没有的模块再回源。查看当前配置：

```bash
go env GOPROXY GOSUMDB
```

npm 安装较慢时，可临时使用国内 registry，例如：

```bash
npm config set registry https://registry.npmmirror.com
```

### 前端依赖与构建

```bash
cd web
npm install
npm run build
```

构建输出在 `web/dist/`，Go 开发模式会从这里提供静态资源。

开发时可单独启动 Vite（`/api` 代理到 `:8080`）：

```bash
cd web
npm run dev
```

浏览器访问 [http://127.0.0.1:5173](http://127.0.0.1:5173)，需同时运行 Go 服务。

### 启动

在项目根目录（需已执行 `npm run build` 生成 `web/dist`）：

```bash
go run ./server -config-dir ./config
```

或：

```bash
make run
```

默认监听 `:8080`，浏览器打开 [http://127.0.0.1:8080](http://127.0.0.1:8080)。

改 Vue 源码后需重新 `npm run build`，再刷新 `:8080`；或使用 `npm run dev` + API 代理进行前端热更新。

可选参数：

```bash
go run ./server -config-dir ./config -addr :8080
```

配置目录也可用环境变量 `ITP_CONFIG_DIR`。

### 热更新（Air）

开发时用 [Air](https://github.com/air-verse/air) 监听 **Go 源码**变更并自动重新编译、重启服务。仓库根目录已有 [`.air.toml`](.air.toml)：构建 `./server`，启动参数为 `-config-dir ./config`，监听 `server/` 下 `.go`、`.json` 等；**不包含** `web/` 源码。

前端改动请任选其一：

- 在 `web/` 运行 `npm run dev`（推荐，HMR）
- 或修改后执行 `npm run build`，再刷新浏览器

国内安装前先设置模块代理（与上文相同，推荐 goproxy.cn）：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

安装 Air：

```bash
go install github.com/air-verse/air@latest
```

确认 `$(go env GOPATH)/bin`（Windows 为 `%USERPROFILE%\go\bin`）在 `PATH` 中，然后在项目根目录启动：

```bash
air
```

若还没有配置文件，可先生成默认配置再按需修改：

```bash
air init
```

本仓库已提交 `.air.toml`，一般直接执行 `air` 即可。默认仍监听 `:8080`。

### 新增页面与菜单项

1. 在 `web/src/views/` 添加页面组件，并在 `web/src/router/menu.js`（及 `web/src/router/index.js`）注册路由与 `meta.menu`。
2. 在 `config/menu.json` 的 `items` 中增加对应 `"path"`（须与路由 path 一致，如 `/scripts`）。
3. 重新 `npm run build` 后重启 Go 服务（或使用 Air 监听后端 / `config` 变更）。

## 构建（单文件可执行程序）

发布构建会通过 [`scripts/sync-web.sh`](scripts/sync-web.sh) / [`scripts/sync-web.ps1`](scripts/sync-web.ps1) 在 `web/` 执行 `npm ci && npm run build`，将 `web/dist` 复制到 `server/embedded/web/`，并以 `-tags release` 嵌入二进制：

```bash
make build          # 当前操作系统/架构 -> dist/itp-<os>-<arch>[.exe]
make build-all      # linux / windows / darwin 全平台
make release        # build-all，并打包 zip（含可执行文件与 config/menu.json）
```

Windows PowerShell：

```powershell
./scripts/build.ps1
./scripts/build.ps1 -Mode release
```

在发布目录中运行打包后的程序，将 `config/menu.json` 与可执行文件放在同一目录（或使用 `-config-dir`）。

## 项目结构

```
server/              Go 入口与 src/（API、静态资源托管）
server/embedded/web/ 发布构建时由 sync-web 写入的前端 dist（嵌入二进制）
web/                 前端源码（Vue + Vite）
  src/
    api/             后端 API 封装
    components/      通用组件
    layouts/         布局（MainLayout）
    router/          路由与菜单 meta
    theme/           主题切换
    views/           页面（如 scripts/）
  dist/              npm run build 输出（开发模式由 Go 读取）
config/
  menu.json          菜单 path 校验
  data/              业务数据（如 scripts.json、脚本文件）
scripts/             sync-web、跨平台构建
```
