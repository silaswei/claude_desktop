# Claude Desktop

<div align="center">

![Claude Desktop](../docs/logo.png)

**一个现代化的 Claude Code CLI 对话管理桌面应用**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8E?logo=go)](https://golang.org/)
[![Vue 3](https://img.shields.io/badge/Vue-3.x-brightgreen.svg)](https://vuejs.org/)

</div>

## ✨ 功能特性

- 🚀 **环境检测** - 自动检测开发环境（Node.js、npm、Claude CLI 等）
- 📁 **工作区管理** - 便捷的项目文件管理和浏览
- 💬 **对话管理** - 完整的 Claude 对话历史记录和查看
- 📝 **流式响应** - 实时显示 Claude 的流式输出
- 🎨 **现代化界面** - 基于 Vue 3 + TypeScript 的响应式设计
- 🔧 **文件操作** - 在应用内直接查看、编辑项目文件
- 🖥️ **终端集成** - 一键打开项目目录的 Claude 终端

## 📸 界面预览

<div align="center">

![Logo](docs/logo.png)

</div>

> 截图准备中... 更多界面展示请参考 [GitHub Issues](https://github.com/yourusername/claude_desktop/issues)

## 🛠️ 技术栈

### 后端
- **框架**: [Wails v2](https://wails.io/) - Go 桌面应用框架
- **语言**: Go 1.22+
- **CLI 集成**: Claude Code CLI
- **存储**: JSON 文件存储（~/.claude-desktop/）

### 前端
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite
- **状态管理**: Pinia
- **UI 框架**: TailwindCSS
- **国际化**: Vue I18n

## 📋 前置要求

### 必需

1. **[Go](https://golang.org/dl/)** 1.22 或更高版本
2. **[Node.js](https://nodejs.org/)** 18.0 或更高版本
3. **[npm](https://www.npmjs.com/)**（随 Node.js 一起安装）
4. **[Claude Code CLI](https://github.com/anthropics/claude-code)** - 必须安装并配置

### 安装 Claude CLI

```bash
npm install -g @anthropics/claude-code
```

### 验证环境

```bash
# 验证安装
claude --version

# 测试 Claude（需要在项目目录下）
claude "hello"
```

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/claude_desktop.git
cd claude_desktop
```

### 2. 安装依赖

```bash
# 安装前端依赖
cd frontend
npm install

# 返回项目根目录
cd ..
```

### 3. 开发模式运行

```bash
# 启动开发服务器（支持热重载）
wails dev
```

应用会自动启动并打开主窗口。

### 4. 生产构建

```bash
# 构建 macOS 应用
wails build
```

构建完成后，可执行文件位于：
```
build/bin/claude_desktop.app/Contents/MacOS/claude_desktop
```

## 📖 使用指南

### 首次使用

1. **环境检测**
   - 启动应用后会自动检测开发环境
   - 确保所有必需的工具（Node.js、npm、Claude CLI）已安装
   - 如果检测失败，会显示详细的修复指引

2. **打开工作区**
   - 点击右上角的 **"📁 打开文件夹"** 按钮
   - 选择你的项目目录
   - 应用会自动加载该目录的文件

3. **开始对话**

   **方式一：在应用内聊天**（推荐）
   - 在底部输入框输入消息
   - 点击发送或按 Enter
   - Claude 的响应会以流式方式实时显示

   **方式二：打开 Claude 终端**
   - 点击顶部的 **"💬 打开 Claude"** 按钮
   - 会在项目目录中打开终端并启动 Claude CLI
   - 可以使用所有 Claude 功能（创建文件、运行命令等）

### 工作区功能

- **文件浏览**: 查看项目文件结构
- **文件编辑**: 点击文件可在编辑器中打开（需配置默认编辑器）
- **右键菜单**: 支持多种文件操作
- **最近使用**: 自动记录最近打开的工作区

### 对话管理

- **历史记录**: 所有对话自动保存
- **多项目切换**: 不同项目保持独立的对话历史
- **导入导出**: 支持对话历史的导入导出（开发中）

## 🏗️ 项目结构

```
claude_desktop/
├── backend/           # Go 后端
│   ├── app/            # 应用主逻辑和 API
│   ├── detector/      # 环境检测器
│   ├── manager/       # 管理器（工作区、对话）
│   ├── models/        # 数据模型
│   ├── service/       # Claude CLI 服务
│   └── logger/        # 日志系统
├── frontend/          # Vue 3 前端
│   ├── src/
│   │   ├── components/  # Vue 组件
│   │   ├── stores/      # Pinia 状态管理
│   │   ├── types/       # TypeScript 类型定义
│   │   ├── views/       # 页面视图
│   │   └── assets/      # 静态资源
│   └── public/         # 公共文件
├── docs/              # 文档和截图
├── build/             # 构建输出
└── main.go            # 应用入口
```

详细架构设计请参阅：
- [系统设计文档](SYSTEM_DESIGN.md)
- [项目开发文档](PROJECT_DOCUMENTATION.md)

## 🔧 配置

### 环境检测配置

环境检测配置位于 `backend/models/environment.go`：

```go
var DefaultEnvironmentConfig = &EnvironmentConfig{
    NodeVersion:      "18.0.0",
    NpmRequired:     true,
    ClaudeRequired:  true,
    CacheDuration:   24 * time.Hour,
}
```

### 数据存储位置

- **对话历史**: `~/.claude-desktop/conversations/`
- **工作区记录**: `~/.claude-desktop/workspaces.json`
- **应用日志**: `~/.claude-desktop/logs/`

## 🐛 故障排除

### 应用无法启动

1. 检查 Go 版本：`go version`
2. 检查 Node.js 版本：`node -v`
3. 清理并重新安装：`cd frontend && rm -rf node_modules && npm install`

### Claude CLI 无响应

1. 验证安装：`claude --version`
2. 检查 API 密钥配置：`claude auth status`
3. 查看应用日志：`~/.claude-desktop/logs/app-*.log`

### 无法创建文件

**开发模式**：无需特殊配置，可直接使用

**生产模式**：需要手动授予权限
1. 系统偏好设置 → 隐私与安全性 → 文件和文件夹
2. 添加 `claude_desktop` 并开启权限
3. 或在 完全磁盘访问权限中添加应用

### 流式输出卡顿

- 检查网络连接
- 查看 Claude CLI 版本
- 确认项目索引已构建

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境设置

```bash
# 安装依赖
cd frontend && npm install

# 启动开发服务器
wails dev
```

### 代码规范

- 后端：遵循 Go 语言规范
- 前端：使用 ESLint + Prettier
- 提交前运行：`npm run lint`

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [Wails](https://wails.io/) - 优秀的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Claude Code](https://github.com/anthropics/claude-code) - 强大的 AI 编程助手
- [TailwindCSS](https://tailwindcss.com/) - 实用优先的 CSS 框架

## 📮 联系方式

- **作者**: silaswei
- **邮箱**: silaswei.com
- **问题反馈**: [GitHub Issues](https://github.com/yourusername/claude_desktop/issues)

---

<div align="center">

**Made with ❤️ by Claude Desktop Team**

</div>
