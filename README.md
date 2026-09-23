# integration-test-platform

基于 Mocha 的一体化测试平台，主要用于 Chrome 内核测试、网页测试，以及测试过程录制。

## 技术栈

- **后端**：Go + Gin（`server/`）
- **前端**：jQuery + Bootstrap 5（`web/`，仅 HTML/CSS/JS）
- **菜单**：集中配置于 [`config/menu.json`](config/menu.json)

布局：顶部栏 + 左侧边栏 + 主内容区。外壳通过 AJAX 从 `web/pages/` 加载页面片段。

## 开发

当前只有 Go 依赖。前端 jQuery / Bootstrap 走 CDN，没有 `npm` / `package.json`。

### 环境要求

- Go **1.22+**
- 在仓库根目录工作（存在 `go.mod`）

```bash
go version
```

### 安装依赖

```bash
go mod download
```

或整理并补全依赖：

```bash
go mod tidy
```

`go.sum` 已在仓库中时，一般 `go mod download` 即可。

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

页面能打开但样式缺失时，多半是前端 CDN（`cdn.jsdelivr.net`）被墙，与 Go 镜像无关。

### 启动

在项目根目录：

```bash
go run ./server -config-dir ./config
```

或：

```bash
make run
```

默认监听 `:8080`，浏览器打开 [http://127.0.0.1:8080](http://127.0.0.1:8080)。

开发模式直接读磁盘上的 `./web`，改 HTML/CSS/JS 后刷新即可，无需同步、无需编译前端。

可选参数：

```bash
go run ./server -config-dir ./config -addr :8080
```

配置目录也可用环境变量 `ITP_CONFIG_DIR`。

### 热更新（Air）

开发时用 [Air](https://github.com/air-verse/air) 监听源码变更并自动重新编译、重启服务。仓库根目录已有 [`.air.toml`](.air.toml)：构建 `./server`，启动参数为 `-config-dir ./config`，并监听 `.go`、`.html`、`.json` 等文件。前端 CSS/JS 仍直接读磁盘，改完刷新浏览器即可。

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

1. 在 `web/pages/your-page.html` 下添加 HTML 片段（不要写完整文档外壳）。
2. 在 `config/menu.json` 中增加一条 `"page": "your-page.html"`。
3. 重启服务端（启动时会校验每个菜单页是否存在）。使用 Air 时保存后会自动重启。

## 构建（单文件可执行程序）

发布构建会将 `web/` 复制到 `server/embedded/web/`，并以 `-tags release` 嵌入二进制。同步步骤会自动执行：

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
server/          Go 入口与 src/
web/             前端源码（不含 .go 文件）
config/          menu.json 及其他配置
scripts/         sync-web、跨平台构建
```
