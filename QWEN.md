# remdev — Web IDE 服务端

Go 单体 HTTP 服务器，内嵌单文件 HTML 前端。纯 tiling 窗口管理器，多工作区，标准 WebDAV 文件服务。
前端配置 (theme, font) 存储在浏览器 localStorage，无服务端配置文件。

## 项目结构

```
src/
├── main.go              # 入口: CLI 解析, 启动 HTTP 服务 (go:embed 前端)
├── go.mod / go.sum
├── config/
│   └── config.go        # CLI 参数解析 (--root, --port, --host, --token)
├── embed/
│   └── index.html       # 单文件前端 (go:embed)
├── handler/
│   ├── auth.go          # Bearer token 认证中间件 (header 或 query)
│   ├── serverinfo.go    # /api/info (hostname + non-loopback IPv4)
│   ├── terminal.go      # /ws/terminal (302 redirect + WS upgrade + PTY)
│   └── webdav.go        # WebDAV handler, rootFS 限制在 --root 内
└── pty/
    ├── pty.go           # PTY 封装 (creack/pty), NewWithSize
    └── manager.go       # UUID → PTY 生命周期, CreateWithSize / Remove
```

## CLI

```
remdev --root /path [--port 7000] [--host 0.0.0.0] [--token <token>]
```

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--root` | *(必需)* | 文件服务根目录 |
| `--port` | `7000` | 监听端口 |
| `--host` | `0.0.0.0` | 监听地址 |
| `--token` | - | Bearer token 认证 (空则跳过) |

所有参数仅通过 CLI 指定。

## 后端 API

### WebDAV — `/dav/{path}`

PROPFIND / GET / PUT / DELETE / MKCOL / MOVE / COPY / LOCK / UNLOCK / OPTIONS

所有操作限制在 `--root` 目录内 (rootFS.isSafe 检查)。

### 服务器信息 — `GET /api/info`

```json
{"hostname":"...","addrs":["192.168.x.x",...]}
```

### 终端 — `/ws/terminal`

| Path | 说明 |
|---|---|
| `GET /ws/terminal` | 302 → `/ws/terminal?uuid=<uuid>` |
| `GET /ws/terminal?uuid=<uuid>` | WS upgrade (带 Upgrade header 时) |

WS 协议: `input`(base64) / `resize`(cols+rows) / `output`(base64) / `exit`(code) / `title`(text)

PTY 在收到首个 `resize` 消息后才创建，确保 shell 以正确尺寸启动。

### 认证

Query `?token=...` 或 header `Authorization: Bearer ...`。token 为空则跳过所有检查。

## 依赖

| 包 | 用途 |
|---|---|
| `github.com/gorilla/websocket` | WebSocket |
| `github.com/creack/pty` | PTY |
| `golang.org/x/net/webdav` | WebDAV |

## 前端 (单文件 HTML)

### 技术栈

- **编辑器**: CodeMirror 5 (CDN, 11 种语言模式)
- **终端**: xterm.js 5 + FitAddon (CDN)
- **文件选择器**: 原生 `<dialog>` 元素
- **无构建工具**: 纯 `<script>` / `<style>` 标签
- **配置存储**: localStorage (`rdev_config` key)

### DOM 结构

```
body[data-theme]
  #status-bar
    #workspaces       (workspace-dot × N)
    #hostname
  #main
    .workspace#ws-N   (absolute, 懒加载)
      .window#wN      (absolute, tiling 定位)
        .window-title  (::before = 焦点指示点)
          .title-text
          .modified
        .window-body   (CodeMirror 或 xterm)
  dialog#picker       (文件选择器)
    input#filename
    .dialog-files#filelist
  .toast              (通知, 动态创建)
```

### 设计令牌 (CSS 自定义属性)

亮色主题 (`[data-theme="light"]`) / 暗色主题 (`:root`):
`--bg`, `--surface`, `--border`, `--accent`, `--accent-dim`, `--text`, `--text-muted`, `--overlay`, `--ws-dot`, `--shadow`, `--focus-ring`, `--radius`

UI 组件 (标题栏、对话框、按钮) 使用 `system-ui, sans-serif`；内容区 (编辑器、终端、文件列表) 使用 `monospace`。

### Tiling (二叉树)

- 节点: `{ type: 'split', dir: 'v'|'h', ratio, a, b }` 或 `{ type: 'leaf', win }`
- 新建窗口: 拆分焦点 leaf 为 split (50/50)，方向按容器宽高比
- 关闭窗口:父 split 被 sibling 替代
- 缩放: 调整父 split 的 ratio，步长 ±10%，最小 5%
- 窗口间 2px 内边距 (G=2)，视觉间距 4px
- 布局: applyLayout() 递归计算 px 坐标，absolute 定位

### 工作区

- 最多 10 个 (MAX_WORKSPACES)，懒加载创建 (ensureWorkspace)
- 首次创建 workspace 0，Alt+N (N=5-0) 自动创建到 N
- 切换时显示/隐藏对应 workspace 的窗口，自动聚焦最后活跃窗口
- 状态条圆点: 8px，当前高亮 (accent 色, scale 1.25)

### 文件选择器 (Alt+B)

- 原生 `<dialog>` + `showModal()`，backdrop 模糊遮罩
- 输入框即为路径导航栏: 显示当前路径，光标在末尾
- 输入文本追加到路径末尾，实时过滤文件列表
- `/` 输入: 取 `/` 前文本，匹配目录后自动进入
- Backspace: 在路径末尾时退回上级目录
- Tab: 自动补全第一个匹配文件名
- ↑↓: 移动高亮项，Enter 打开高亮项
- 点击文件直接打开，点击目录进入
- 点击 backdrop 或 Escape 关闭
- 过滤无匹配时 Enter: 将输入作为绝对/相对路径尝试打开

### 配置编辑器 (Alt+,)

- 以特殊路径 `__rdev_config__` 打开 Editor
- 初始内容为 `{}`，用户添加所需字段
- 支持的 key: `theme`, `font_family`, `font_size`
- Ctrl/Cmd+S 保存到 localStorage，theme 即时生效

### 快捷键 (全部 Alt)

| 快捷键 | 行为 |
|---|---|
| Alt+B | 打开文件选择器 |
| Alt+, | 打开 rdev.json 配置编辑器 |
| Alt+N | 新建 Terminal 窗口 |
| Alt+W | 关闭当前窗口 (Editor 未保存提示) |
| Alt+S | 保存 Editor 内容 |
| Alt+R | 重新读取 Editor 文件 |
| Alt+= | 当前窗口放大 |
| Alt+- | 当前窗口缩小 |
| Alt+1–0 | 切换工作区 (0 = workspace 10) |
| Alt+↑↓←→ | 焦点导航 |

### 终端交互

- 选中文本后 Cmd/Ctrl+C 复制到剪贴板
- Cmd/Ctrl+V 从剪贴板粘贴到终端
- 终端退出后窗口自动关闭

## 构建

```bash
make clean && make
```

产物: `dist/remdev-{linux|darwin|windows}-{amd64|arm64}[.exe]` (6 个目标)
