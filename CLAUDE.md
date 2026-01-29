# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Wails v2 框架开发的 Claude Code 对话管理器桌面应用。使用 Go 语言作为后端，Vue 3 + TypeScript 作为前端，用于管理 Claude Code CLI 的对话和工作区。

## 开发命令

### 构建和运行

```bash
# 开发模式运行（前端热重载）
wails dev

# 构建生产版本
wails build

# 检查 Wails 环境
wails doctor
```

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发服务器（独立运行，不使用 Wails）
npm run dev

# 构建前端
npm run build

# 类型检查
npm run type-check

# 代码检查和修复
npm run lint
```

### 后端开发

```bash
# 格式化 Go 代码
go fmt ./...

# 运行测试
go test ./...

# 添加依赖
go get <package>
go mod tidy
```

## 架构概述

### 技术栈
- **后端**: Go 1.22+ with Wails v2.11
- **前端**: Vue 3 + TypeScript + Vite
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **UI**: TailwindCSS + 自定义组件
- **国际化**: Vue I18n

### 三层架构

```
┌─────────────────────────────────────┐
│  前端层 (Vue 3)                      │
│  - 环境检测启动页                     │
│  - 工作区和对话管理界面                │
└─────────────────────────────────────┘
            ↕ Wails Bindings
┌─────────────────────────────────────┐
│  业务逻辑层 (Go)                     │
│  - 环境检测器 (Detector)             │
│  - 工作区管理器 (Workspace Manager)  │
│  - 对话管理器 (Conversation Manager)  │
└─────────────────────────────────────┘
            ↕
┌─────────────────────────────────────┐
│  服务层 (Go)                        │
│  - Claude API 服务                   │
│  - JSON 文件存储                     │
│  - 系统命令执行                      │
└─────────────────────────────────────┘
```

## 核心模块

### 1. 环境检测模块 (`backend/detector/`)

环境检测器在应用启动时自动运行，检查开发环境是否满足运行 Claude Code 的条件。

**检测器接口** (`detector.go`):
- `Detect(ctx) -> DetectionResult`: 执行检测
- `Name() -> string`: 获取检测器名称
- `Required() -> bool`: 是否必需

**实现的检测器**:
- `NodeDetector`: 检测 Node.js 版本（要求 >= 18.0.0）
- `NpmDetector`: 检测 npm 包管理器
- `ClaudeDetector`: 检测 claude-code CLI 命令
- `NetworkDetector`: 检测 Claude API 网络连通性
- `GitDetector`: 检测 Git（可选）

**检测结果状态**:
- `pending`: 检测中
- `success`: 检测通过
- `failed`: 检测失败

### 2. 工作区管理模块 (`backend/manager/workspace/`)

管理项目工作区，提供文件操作和工作区切换功能。

**核心功能**:
- 打开/关闭工作区
- 文件系统操作（读取、写入、删除、重命名）
- 最近使用的工作区列表
- 工作区与对话的关联

**关键 API** (`backend/app/app.go`):
- `WorkspaceOpen(path)`: 打开工作区
- `WorkspaceGetCurrent()`: 获取当前工作区路径
- `WorkspaceListFiles()`: 列出工作区文件
- `WorkspaceReadFile(path)`: 读取文件内容
- `WorkspaceWriteFile(path, content)`: 写入文件

### 3. 对话管理模块 (`backend/manager/conversation/`)

管理 Claude Code 对话历史，支持本地存储和流式响应。

**数据结构**:
- `Conversation`: 对话实体（ID、标题、项目路径、消息列表）
- `Message`: 消息实体（ID、角色、内容、时间戳）

**存储**:
- 使用 JSON 文件存储在 `~/.claude-terminal/conversations/`
- 每个对话一个独立的 JSON 文件

**流式响应**:
- 通过 Wails Events 实时推送响应片段
- 事件名称: `claude:thinking`, `claude:response`, `claude:complete`, `claude:error`

### 4. 服务层 (`backend/service/`)

**ClaudeService**: Claude CLI 调用服务
- 使用 `claude --print` 命令进行非交互式调用
- 支持流式 JSON 输出 (`--output-format stream-json`)
- 在指定项目目录下执行命令
- 方法: `SendRequest()`, `SendMessage()`, `StreamMessage()`, `ValidateEnvironment()`

**ConversationManager**: 对话管理服务
- 包装存储层，提供完整的 CRUD 操作
- 整合 ClaudeService 处理消息发送
- 支持流式响应回调 (`onChunk`)
- 方法: `CreateConversation()`, `DeleteConversation()`, `GetConversation()`, `ListConversations()`, `UpdateConversation()`, `SendMessageWithCallback()`

### 5. 应用主结构 (`backend/app/app.go`)

`App` 结构体是 Wails 绑定的主要入口，包含:
- 环境检测 API（Env 前缀）
- 工作区管理 API（Workspace 前缀）
- 对话管理 API（Conversation 前缀）
- 系统操作 API（System 前缀）

## 路由和 API 规范

根据项目要求，所有增删改查操作统一使用以下格式:

| 功能 | 路由格式 | HTTP 方法 |
|------|---------|----------|
| 创建 | `xxx/create` | POST |
| 删除 | `xxx/delete` | DELETE |
| 更新 | `xxx/update` | PUT |
| 详情 | `xxx/info` | GET |
| 列表 | `xxx/list` | GET |

**实际实现**:
由于使用 Wails 框架（而非传统 HTTP API），这些路由格式体现在 Go 方法命名上:
- `ConversationCreate()`: 创建新对话
- `ConversationDelete()`: 删除对话
- `ConversationUpdate()`: 更新对话
- `ConversationInfo()`: 获取对话详情
- `ConversationList()`: 获取对话列表

## 数据模型 (`backend/models/`)

- `DetectionResult`: 环境检测结果（状态、版本、修复命令、时间戳）
- `EnvironmentConfig`: 环境检测配置（版本要求、超时设置、缓存策略）
- `EnvironmentInfo`: 环境信息摘要（总体状态、通过率、检测时间）
- `Workspace`: 工作区实体（路径、名称、最后打开时间、关联对话）
- `Conversation`: 对话实体（ID、标题、项目路径、消息列表、时间戳）
- `Message`: 消息实体（ID、角色、内容、时间戳、工具调用）

## 前端状态管理

使用 Pinia 进行状态管理，主要 stores:

- `stores/env.ts`: 环境检测状态
- `stores/workspace.ts`: 工作区状态
- `stores/conversation.ts`: 对话状态
- `stores/ui.ts`: UI 状态（主题、语言等）

## 重要文件位置

- `main.go`: 应用入口，定义窗口配置
- `backend/app/app.go`: 核心应用结构和所有 API 绑定
- `backend/detector/`: 环境检测器实现
- `backend/manager/`: 工作区和对话管理器
- `backend/models/`: 数据模型定义
- `frontend/src/views/`: 页面视图组件
- `frontend/src/stores/`: Pinia 状态管理
- `SYSTEM_DESIGN.md`: 详细的系统设计文档
- `design.md`: 原始需求文档

## Wails 事件系统

应用使用 Wails Events 进行前后端通信，特别是流式响应:

**后端发送事件** (`runtime.EventsEmit`):
```go
runtime.EventsEmit(ctx, "claude:response", map[string]interface{}{
    "content": chunk,
    "convID":  convID,
})
```

**前端监听事件**:
```typescript
import { EventsOn } from '../../wailsjs/runtime/runtime'

EventsOn("claude:response", (data) => {
    // 处理响应片段
})
```

## 开发注意事项

1. **修改 API 文件时**: 只允许修改 `backend/` 目录下的 Go 文件，不允许修改其他文件

2. **环境检测缓存**: 检测结果会被缓存（24小时），开发时可能需要调用 `EnvClearCache()` 清除缓存

3. **macOS 特定代码**: 某些系统命令（如打开文件、终端）使用了 macOS 特定的命令（`open`, `-a Terminal`）

4. **窗口大小**: 应用启动时会自动调整为屏幕尺寸的 3/4（在 `app.go:resizeWindowToThreeQuarters()`）

5. **数据目录**: 对话数据存储在 `~/.claude-terminal/` 目录

6. **前端构建**: 修改前端代码后需要运行 `wails build` 或在 `wails dev` 模式下会自动热重载

## Claude CLI 集成

应用通过执行本地 `claude` 命令与 Claude Code CLI 交互。

**关键参数**:
- `--print`: 非交互模式，直接输出响应
- `--output-format stream-json`: 流式 JSON 输出格式，返回 `stream_event` 类型
- `--verbose`: 详细输出
- `--include-partial-messages`: 包含部分消息

**流式事件处理**:
后端解析 `stream_event` 类型的事件，提取 `content_block_delta` 中的文本增量：
```go
switch eventStr := event["type"].(string); eventStr {
case "content_block_delta":
    // 文本内容增量
    if delta, ok := event["delta"].(map[string]interface{}); ok {
        if text, ok := delta["text"].(string); ok {
            onChunk(text)
        }
    }
}
```

**工作目录**: 命令在选定的工作区路径下执行（`cmd.Dir = projectPath`），确保 Claude 可以访问项目文件。

**版本验证**: 使用 `claude --version` 验证环境是否可用。

## 主题和样式

- 使用 TailwindCSS 进行样式开发
- 支持明暗主题切换
- 主题配置: `frontend/src/stores/ui.ts`
- 主要配色: 紫色系 (#ab7edc, #d7a8d8)
- 毛玻璃效果背景

## 国际化

- 使用 Vue I18n
- 语言文件: `frontend/src/locales/`
- 支持中文（zh-Hans）和英文（en）
