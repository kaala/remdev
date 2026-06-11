# remdev — Web IDE 服务端

Go 单体 HTTP 服务器，内嵌单文件 HTML 前端。纯 tiling 窗口管理器，多工作区，标准 WebDAV 文件服务。

## 项目结构

```
src/
├── main.go              # 入口: CLI 解析, 配置加载, 首次运行创建默认配置
├── go.mod / go.sum
├── config/
│   └── config.go        # 配置结构, JSON 加载/写回, CLI 合并, --option
├── embed/
│   └── index.html       # 单文件前端 (go:embed)
├── handler/
│   ├── auth.go          # Bearer token 认证中间件
│   ├── config.go        # /api/config 和 /api/config/full
│   ├── serverinfo.go    # /api/serverinfo (hostname + IP)
│   ├── terminal.go      # /ws/terminal (302 redirect + WS upgrade + PTY)
│   └── webdav.go        # WebDAV handler (golang.org/x/net/webdav)
└── pty/
    ├── pty.go           # PTY 封装 (creack/pty)
    └── manager.go       # UUID → PTY 生命周期管理
```

## CLI

```
remdev --root /path [--config ~/.config/remdev/remdev.json] [--port 7000] [--host 0.0.0.0]
     [--option key=value]...
```

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--root` | `./` | 文件服务根目录 (必须 CLI 指定) |
| `--config` | `~/.config/remdev/remdev.json` | 配置文件路径 |
| `--port` | `7000` | 监听端口 |
| `--host` | `0.0.0.0` | 监听地址 |
| `--option k=v` | - | 覆盖配置字段 (可多次) |

优先级: CLI > `--option` > 配置文件 > 默认值

## 配置文件 `~/.config/remdev/remdev.json`

```json
{
  "port": 7000,
  "host": "0.0.0.0",
  "token": "",
  "theme": "light",
  "font_family": "Ubuntu Sans Mono",
  "font_size": 14
}
```

首次运行自动创建。`root` 不在配置文件中，必须 CLI 指定。

## 后端 API

### WebDAV — `/dav/{path}` (全方法, 标准 DAV)

PROPFIND / GET / PUT / DELETE / MKCOL / MOVE / COPY / LOCK / UNLOCK / OPTIONS

### 配置

| Method | Path | 说明 |
|---|---|---|
| GET | `/api/config` | 前端配置 `{theme, font_family, font_size}` |
| PUT | `/api/config` | 更新配置 |
| GET | `/api/config/full` | 完整配置 (编辑器用) |
| PUT | `/api/config/full` | 保存完整配置 |

### 服务器信息

| GET | `/api/serverinfo` | `{"hostname":"...","addrs":["192.168.x.x",...]}` |

### 终端

| Path | 说明 |
|---|---|
| `GET /ws/terminal` | 302 → `/ws/terminal?uuid=<uuid>` |
| `GET /ws/terminal?uuid=<uuid>` | WS upgrade (带 Upgrade header 时) |

WS 协议: `input` / `resize` / `output` / `exit` / `title`

### 认证

Query `?token=...` 或 header `Authorization: Bearer ...`。token 为空则跳过。

## 依赖

| 包 | 用途 |
|---|---|
| `github.com/gorilla/websocket` | WebSocket |
| `github.com/creack/pty` | PTY |
| `golang.org/x/net/webdav` | WebDAV |

## 前端 (单文件 HTML)

### 技术栈

- **编辑器**: CodeMirror 5 (CDN, 12 种语言)
- **终端**: xterm.js 5 + FitAddon (CDN)
- **无构建工具**: 纯 `<script>` 标签加载

### 交互

- 纯 tiling，无侧边栏
- 窗口类型: Editor 或 Terminal
- 焦点窗口: 蓝色高亮标题栏 (#0af)；非焦点: 暗色

### 快捷键 (全部 Alt)

| 快捷键 | 行为 |
|---|---|
| Alt+B | 打开文件选择器 (WebDAV PROPFIND) |
| Alt+N | 新建 Terminal |
| Alt+W | 关闭窗口 (Editor 未保存提示) |
| Alt+S | 保存文件 (Editor) |
| Alt+R | 重新读取文件 |
| Alt+= | 窗口放大 (+5%) |
| Alt+- | 窗口缩小 (-5%) |
| Alt+1-4 | 切换工作区 |
| Alt+↑↓←→ | 焦点导航 |

### Tiling (二叉树)

- 树节点: `{ type: 'split', dir: 'v'|'h', ratio, a, b }` 或 `{ type: 'leaf', win }`
- 新建: 拆分焦点 leaf 为 split (50/50)，方向按容器宽高比
- 关闭: parent split 被 sibling 替代
- 缩放: 调整 direct parent split 的 ratio，固定步长 ±5%，最小 5%
- 布局: 递归计算 px 坐标，窗口 absolute 定位

### 工作区

- 4 个工作区，独立窗口组和 tiling 树
- 切换时隐藏/显示 DOM，不自动创建终端
- 状态条左侧 [1][2][3][4] 按钮

### 文件选择器 (Alt+B)

- 模态弹窗，左侧目录树 + 右侧文件列表 + 文件名输入
- 双击或 Enter 打开文件到 Editor 窗口
- 文件已打开则聚焦已有窗口

### 配置编辑器 (cfg 按钮)

- 以特殊路径 `__remdev_config__` 打开 Editor
- 读取 `/api/config/full`，保存到 `/api/config/full`
- 编辑器中 theme 变更即时生效

### UI 风格

- 终端风格：暗色背景 `#0a0a0a`，等宽字体 14px
- 细边框 1px，蓝色焦点 `#0af`
- 状态条: [工作区] `hostname (ip,...)` [cfg]
- 脏标记: `[modified]` 文本

## 构建

```bash
make
# 或
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C src -o dist/remdev-linux-amd64 .
# ... 6 平台: linux/darwin/windows × amd64/arm64
```

产物: `dist/remdev-{platform}` + `dist/index.html`
