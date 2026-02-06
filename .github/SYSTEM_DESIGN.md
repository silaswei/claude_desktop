# Claude Code 对话管理器 - 整体设计文档

## 一、系统架构设计

### 1.1 技术栈
- **后端**: Go (Wails v2)
- **前端**: Vue 3 + TypeScript + Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **UI框架**: 自定义组件 + Tailwind CSS
- **国际化**: Vue I18n

### 1.2 应用架构
```
┌─────────────────────────────────────────────────┐
│               前端层 (Vue 3)                     │
├─────────────────────────────────────────────────┤
│  启动检测页  │  主界面  │  对话管理  │  设置页    │
└─────────────────────────────────────────────────┘
                    ↕ (Wails Binding)
┌─────────────────────────────────────────────────┐
│            业务逻辑层 (Go)                       │
├─────────────────────────────────────────────────┤
│  环境检测器  │  对话管理器  │  项目管理器         │
└─────────────────────────────────────────────────┘
                    ↕
┌─────────────────────────────────────────────────┐
│            服务层 (Go)                           │
├─────────────────────────────────────────────────┤
│  Claude API  │  本地存储  │  终端命令执行         │
└─────────────────────────────────────────────────┘
```

## 二、目录结构设计

```
claude_terminal/
├── main.go                          # 应用入口
├── app.go                           # 核心应用结构
├── design.md                        # 原始需求文档
├── SYSTEM_DESIGN.md                 # 系统设计文档（本文档）
│
├── backend/                         # Go 后端代码
│   ├── app/
│   │   ├── app.go                   # 应用主结构
│   │   └── events.go                # 事件处理
│   ├── detector/                    # 环境检测模块
│   │   ├── detector.go              # 检测器接口定义
│   │   ├── node_detector.go         # Node.js 检测
│   │   ├── npm_detector.go          # npm 检测
│   │   ├── claude_detector.go       # Claude CLI 检测
│   │   ├── network_detector.go      # 网络连通性检测
│   │   └── git_detector.go          # Git 检测（可选）
│   ├── manager/                     # 管理器模块
│   │   ├── conversation/            # 对话管理
│   │   │   ├── conversation.go      # 对话实体
│   │   │   ├── message.go           # 消息实体
│   │   │   └── storage.go           # 本地存储
│   │   └── project/                 # 项目管理
│   │       ├── project.go           # 项目实体
│   │       └── scanner.go           # 项目扫描器
│   ├── service/                     # 服务层
│   │   ├── claude_service.go        # Claude API 服务
│   │   └── terminal_service.go      # 终端命令执行服务
│   └── models/                      # 数据模型
│       ├── detection_result.go      # 检测结果模型
│       ├── environment.go           # 环境信息模型
│       └── config.go                # 配置模型
│
├── frontend/                        # Vue 前端代码
│   ├── src/
│   │   ├── App.vue                  # 根组件
│   │   ├── main.ts                  # 前端入口
│   │   │
│   │   ├── views/                   # 页面视图
│   │   │   ├── LaunchView.vue       # 启动检测页
│   │   │   ├── MainView.vue         # 主界面
│   │   │   ├── ConversationView.vue # 对话管理页
│   │   │   └── SettingsView.vue     # 设置页
│   │   │
│   │   ├── components/              # 组件
│   │   │   ├── launcher/            # 启动相关组件
│   │   │   │   ├── LaunchScreen.vue       # 启动画面
│   │   │   │   ├── DetectionProgress.vue  # 检测进度
│   │   │   │   ├── FailureGuide.vue       # 失败引导页
│   │   │   │   └── EnvironmentCard.vue    # 环境信息卡片
│   │   │   ├── layout/              # 布局组件
│   │   │   │   ├── Sidebar.vue            # 侧边栏
│   │   │   │   ├── Header.vue             # 顶部栏
│   │   │   │   └── StatusIndicator.vue    # 状态指示器
│   │   │   ├── conversation/        # 对话相关组件
│   │   │   │   ├── ConversationList.vue   # 对话列表
│   │   │   │   ├── MessageList.vue        # 消息列表
│   │   │   │   ├── MessageItem.vue        # 消息项
│   │   │   │   └── InputPanel.vue         # 输入面板
│   │   │   └── project/             # 项目相关组件
│   │   │       ├── ProjectSelector.vue    # 项目选择器
│   │   │       └── ProjectCard.vue        # 项目卡片
│   │   │
│   │   ├── stores/                  # Pinia 状态管理
│   │   │   ├── env.ts               # 环境状态
│   │   │   ├── conversation.ts      # 对话状态
│   │   │   ├── project.ts           # 项目状态
│   │   │   └── ui.ts                # UI 状态
│   │   │
│   │   ├── router/                  # 路由配置
│   │   │   └── index.ts
│   │   │
│   │   ├── api/                     # Wails 生成的 API
│   │   │   └── wailsjs/
│   │   │
│   │   ├── types/                   # TypeScript 类型定义
│   │   │   ├── env.ts               # 环境类型
│   │   │   ├── conversation.ts      # 对话类型
│   │   │   └── project.ts           # 项目类型
│   │   │
│   │   ├── i18n/                    # 国际化
│   │   │   └── locales/
│   │   │       ├── zh-Hans.json
│   │   │       └── en.json
│   │   │
│   │   ├── assets/                  # 静态资源
│   │   │   ├── css/
│   │   │   └── images/
│   │   │
│   │   └── utils/                   # 工具函数
│   │       ├── formatter.ts
│   │       └── validator.ts
│   │
│   ├── dist/                        # 构建输出
│   └── package.json
│
└── wailsjson/                       # Wails 生成的绑定代码
```

## 三、核心模块设计

### 3.1 环境检测模块

#### 3.1.1 检测器接口
```go
type Detector interface {
    // 执行检测
    Detect(ctx context.Context) (*DetectionResult, error)
    // 获取检测器名称
    Name() string
    // 是否必需
    Required() bool
}
```

#### 3.1.2 检测结果模型
```go
type DetectionResult struct {
    Name        string    `json:"name"`         // 检测项名称
    Status      string    `json:"status"`       // pending/success/failed
    Version     string    `json:"version"`      // 版本信息
    Message     string    `json:"message"`      // 提示信息
    FixCommand  string    `json:"fixCommand"`   // 修复命令
    Required    bool      `json:"required"`     // 是否必需
}
```

#### 3.1.3 检测器实现
- **NodeDetector**: 检测 Node.js 版本（要求 >= 18.0.0）
- **NpmDetector**: 检测 npm 版本
- **ClaudeDetector**: 检测 claude-code 命令是否可用
- **NetworkDetector**: 检测 Claude API 连通性
- **GitDetector**: 检测 Git（可选）

### 3.2 对话管理模块

#### 3.2.1 对话实体
```go
type Conversation struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    ProjectPath string    `json:"projectPath"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    Messages    []Message `json:"messages"`
}
```

#### 3.2.2 消息实体
```go
type Message struct {
    ID        string    `json:"id"`
    Role      string    `json:"role"`      // user/assistant/system
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    ToolCalls []ToolCall `json:"toolCalls,omitempty"`
}
```

#### 3.2.3 本地存储
- 使用 JSON 文件存储对话历史
- 路径: `~/.claude-desktop/conversations/`
- 支持导入/导出功能

### 3.3 项目管理模块

#### 3.3.1 项目实体
```go
type Project struct {
    Path         string    `json:"path"`
    Name         string    `json:"name"`
    Language     string    `json:"language"`
    LastOpened   time.Time `json:"lastOpened"`
    Icon         string    `json:"icon"`
}
```

#### 3.3.2 项目扫描器
- 自动扫描常用开发目录
- 支持手动添加项目
- 记录最近打开的项目

## 四、数据流程设计

### 4.1 启动检测流程
```
应用启动
    ↓
显示启动画面
    ↓
并行执行环境检测
    ↓
收集检测结果
    ↓
判断是否全部通过
    ├─ 是 → 进入主界面
    └─ 否 → 显示引导页 → 用户修复 → 重新检测
```

### 4.2 对话流程
```
用户输入消息
    ↓
添加到消息列表（前端）
    ↓
调用 Claude API（后端）
    ↓
流式返回响应
    ↓
更新消息列表（前端）
    ↓
保存到本地存储
```

## 五、接口设计

### 5.1 环境检测接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/env/detect/all` | GET | 执行所有环境检测 |
| `/env/detect/{name}` | GET | 执行单个检测 |
| `/env/fix/{name}` | POST | 尝试自动修复 |
| `/env/status` | GET | 获取环境状态摘要 |

### 5.2 对话管理接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/conversation/create` | POST | 创建新对话 |
| `/conversation/delete` | DELETE | 删除对话 |
| `/conversation/update` | PUT | 更新对话信息 |
| `/conversation/info` | GET | 获取对话详情 |
| `/conversation/list` | GET | 获取对话列表 |
| `/conversation/send` | POST | 发送消息 |

### 5.3 项目管理接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/project/scan` | GET | 扫描项目 |
| `/project/open` | POST | 打开项目 |
| `/project/recent` | GET | 获取最近项目 |

## 六、UI/UX 设计

### 6.1 启动检测页
- **风格**: 毛玻璃效果背景，深色主题
- **动画**: 检测进度条，状态图标动画
- **交互**: 实时显示检测进度，失败时显示修复指引

### 6.2 主界面
- **布局**:
  - 左侧：项目列表 + 对话列表
  - 中间：消息显示区域
  - 右侧：工具调用面板（可折叠）
- **状态栏**: 显示环境状态指示器

### 6.3 主题配色
```typescript
const theme = {
  primary: '#ab7edc',      // 主色调（紫色）
  secondary: '#d7a8d8',    // 次要色
  success: '#52c41a',      // 成功（绿色）
  warning: '#faad14',      // 警告（黄色）
  error: '#f5222d',        // 错误（红色）
  background: 'rgba(219, 188, 239, 0.9)', // 背景毛玻璃
}
```

## 七、技术实现要点

### 7.1 检测缓存机制
- 首次检测完成后缓存结果
- 缓存有效期: 24小时
- 手动刷新时强制重新检测

### 7.2 平台适配
```go
// 获取平台特定的安装命令
func getPlatformFixCommand(detectorName string) string {
    switch runtime.GOOS {
    case "darwin":
        return getMacOSCommand(detectorName)
    case "windows":
        return getWindowsCommand(detectorName)
    case "linux":
        return getLinuxCommand(detectorName)
    default:
        return ""
    }
}
```

### 7.3 网络检测策略
- 超时时间: 10秒
- 重试次数: 3次
- 失败时提供离线模式选项

### 7.4 权限处理
- 检测命令是否需要 sudo/admin 权限
- 提示用户授权
- macOS: 使用 osascript 提示授权
- Windows: 检测管理员权限

### 7.5 Claude API 集成
- 调用本地 claude-code 命令
- 使用流式输出
- 支持 Tool Calls 显示

## 八、开发顺序建议

1. **阶段一**: 环境检测模块
   - 实现检测器接口
   - 实现各个检测器
   - 完成启动检测页 UI

2. **阶段二**: 项目管理模块
   - 实现项目扫描器
   - 完成项目选择器 UI
   - 集成到主界面

3. **阶段三**: 对话管理模块
   - 实现对话实体和存储
   - 集成 Claude API
   - 完成对话界面

4. **阶段四**: 优化和完善
   - 添加设置页
   - 完善国际化
   - 性能优化

## 九、文件修改规则

根据你的要求：
- **API 文件修改**: 只修改 `backend/` 目录下的 Go 文件
- **路由格式**: 统一使用 `xxx/create`, `xxx/delete`, `xxx/update`, `xxx/info`, `xxx/list` 格式
- **GET 方法**: info, list
- **POST 方法**: create
- **PUT 方法**: update
- **DELETE 方法**: delete
