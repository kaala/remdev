remdev — 基于浏览器的轻量级 Web IDE。

Go 编写的单体 HTTP 服务器，内嵌单文件 HTML 前端。通过 WebDAV 提供文件浏览/编辑，
通过 WebSocket + PTY 提供终端。纯 tiling 窗口管理器布局，支持多工作区。

## 核心能力

1. **文件管理** — 通过 WebDAV 协议 (PROPFIND/GET/PUT/DELETE/MKCOL 等)
2. **终端** — WebSocket + PTY，支持 resize、input/output、title 同步
3. **代码编辑** — CodeMirror 5，多语言语法高亮

## 技术选型

- **后端**: Go，跨平台静态编译 (linux/darwin/windows × amd64/arm64)
- **前端**: 单文件 HTML，无构建工具，纯 `<script>` 标签加载
- **编辑器**: CodeMirror 5 (CDN，11 种语言)
- **终端**: xterm.js 5 + FitAddon (CDN)
- **文件服务**: golang.org/x/net/webdav
- **配置**: 浏览器 localStorage (`rdev_config` key)，无服务端配置

## 布局

- 纯 tiling 窗口管理器，无侧边栏
- 窗口类型: Editor / Terminal
- 二叉树分割 (v/h)，新建沿长边 50/50 分割
- 窗口间 4px 间距
- 支持窗口缩放 (Alt+= / Alt+-)，步长 ±10%

## 工作区

- 最多 10 个工作区 (Alt+1–0)，懒加载创建
- 顶部状态条以 8px 圆点指示，当前工作区高亮
- 状态条右侧显示 hostname (IP)

## 文件选择器 (Alt+B)

- 原生 `<dialog>` 元素，路径栏输入框在上，文件列表在下
- 输入框即为路径：输入 `/` 进入子目录，Backspace 退回上级
- 实时过滤：路径后缀自动过滤文件名
- Tab 自动补全，↑↓ 移动高亮，Enter 打开
- 点击文件直接打开，点击目录进入
- 点击背景或 Escape 关闭

## 快捷键 (全部 Alt)

| 快捷键 | 行为 |
|---|---|
| Alt+B | 打开文件选择器 |
| Alt+, | 打开配置编辑器 (`rdev.json`) |
| Alt+N | 新建终端 |
| Alt+W | 关闭窗口 (Editor 未保存提示) |
| Alt+S | 保存文件 |
| Alt+R | 重新读取文件 |
| Alt+= | 窗口放大 |
| Alt+- | 窗口缩小 |
| Alt+1–0 | 切换工作区 (0=10) |
| Alt+↑↓←→ | 焦点导航 |

## 主题

- 亮色 / 暗色双主题，默认亮色
- CSS 自定义属性切换 (`[data-theme]`)
- 终端和编辑器主题随系统切换
- 等宽字体由浏览器默认 (可配置)
