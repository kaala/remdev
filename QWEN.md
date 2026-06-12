# remdev — Web IDE 服务端

Go 单体 HTTP 服务器，内嵌单文件 HTML 前端。纯 tiling 窗口管理器，多工作区，标准 WebDAV 文件服务。
前端配置 (theme, font) 存储在浏览器 localStorage，无服务端配置文件。

## 项目结构

```
src/
├── main.go              # 入口: CLI 解析, 启动 HTTP 服务
├── go.mod / go.sum
├── config/
│   └── config.go        # CLI 参数解析 (--root, --port, --host, --option)
├── embed/
│   └── index.html       # 单文件前端 (go:embed)
├── handler/
│   ├── auth.go          # Bearer token 认证中间件
│   ├── serverinfo.go    # /api/info (hostname + IP)
│   ├── terminal.go      # /ws/terminal (302 redirect + WS upgrade + PTY)
│   └── webdav.go        # WebDAV handler (golang.org/x/net/webdav)
└── pty/
    ├── pty.go           # PTY 封装 (creack/pty), New / NewWithSize
    └── manager.go       # UUID → PTY 生命周期, Create / CreateWithSize
```

## CLI

```
remdev --root /path [--port 7000] [--host 0.0.0.0] [--option key=value]...
```

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--root` | *(必需)* | 文件服务根目录 |
| `--port` | `7000` | 监听端口 |
| `--host` | `0.0.0.0` | 监听地址 |
| `--option k=v` | - | 覆盖配置 (支持: port, host, token) |

所有参数仅通过 CLI 指定，无配置文件。

## 后端 API

### WebDAV — `/dav/{path}` (全方法, 标准 DAV)

PROPFIND / GET / PUT / DELETE / MKCOL / MOVE / COPY / LOCK / UNLOCK / OPTIONS

### 服务器信息

| GET | `/api/info` | `{"hostname":"...","addrs":["192.168.x.x",...]}` |

### 终端

| Path | 说明 |
|---|---|
| `GET /ws/terminal` | 302 → `/ws/terminal?uuid=<uuid>` |
| `GET /ws/terminal?uuid=<uuid>` | WS upgrade (带 Upgrade header 时) |

WS 协议: `input` / `resize` / `output` / `exit` / `title`

PTY 在收到首个 `resize` 消息后才创建，确保 shell 以正确尺寸启动，避免多余换行。

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
- **配置存储**: localStorage (`rdev_config` key)

### 交互

- 纯 tiling，无侧边栏
- 窗口类型: Editor 或 Terminal
- 焦点窗口: 蓝色高亮标题栏 (#0af)；非焦点: 暗色

### 快捷键 (全部 Alt)

| 快捷键 | 行为 |
|---|---|
| Alt+B | 打开文件选择器 (WebDAV PROPFIND) |
| Alt+, | 打开配置编辑器 |
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

### 配置编辑器 (Alt+,)

- 以特殊路径 `__remdev_config__` 打开 Editor
- 编辑 theme, font_family, font_size (JSON 格式)
- 保存到 localStorage，theme 变更即时生效

### UI 风格

- 终端风格：暗色背景 `#0a0a0a`，等宽字体 14px
- 细边框 1px，蓝色焦点 `#0af`
- 状态条: [工作区] `hostname (ip,...)`
- 脏标记: `[modified]` 文本

## 构建

```bash
make
# 或
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C src -o dist/remdev-linux-amd64 .
# ... 6 平台: linux/darwin/windows × amd64/arm64
```

产物: `dist/remdev-{platform}` + `dist/index.html`
