# Claude Desktop 项目文档

## 项目概述

Claude Desktop 是一个基于 Wails 框架开发的桌面应用程序，集成了 Claude AI 命令行工具，提供图形化的对话界面和项目管理功能。

### 技术栈

**后端:**
- Go 1.24+
- Wails v2.11.0 (Go 桌面应用框架)
- 标准库: encoding/json, os/exec, sync, context 等

**前端:**
- Vue 3 (Composition API)
- TypeScript
- Pinia (状态管理)
- Vite (构建工具)

### 项目结构

```
claude_desktop/
├── backend/                    # Go 后端代码
│   ├── app/                   # 应用层 (Wails 绑定)
│   │   └── app.go            # 主应用结构体和 API 导出
│   ├── detector/              # 环境检测模块
│   │   ├── manager.go        # 检测管理器
│   │   ├── claude_detector.go
│   │   ├── git_detector.go
│   │   ├── network_detector.go
│   │   ├── node_detector.go
│   │   └── npm_detector.go
│   ├── manager/              # 业务管理器
│   │   ├── conversation/    # 对话管理
│   │   │   ├── conversation.go
│   │   │   └── storage.go
│   │   └── workspace/       # 工作区管理
│   │       └── workspace.go
│   ├── models/              # 数据模型
│   │   ├── environment.go
│   │   └── workspace.go
│   ├── service/             # 服务层
│   │   └── claude_service.go
│   └── main.go             # 程序入口
├── frontend/              # Vue 前端代码
│   ├── src/
│   │   ├── components/    # 组件
│   │   ├── stores/       # Pinia 状态管理
│   │   │   ├── env.ts    # 环境检测状态
│   │   │   └── workspace.ts  # 工作区状态
│   │   ├── types/        # TypeScript 类型定义
│   │   ├── views/        # 页面视图
│   │   │   ├── MainView.vue  # 主界面
│   │   │   └── WelcomeView.vue
│   │   ├── App.vue
│   │   └── main.ts
│   └── package.json
├── build/               # 构建输出
├── wails.json          # Wails 配置
├── main.go            # Wails 入口
└── go.mod             # Go 依赖管理
```

---

## 核心架构设计

### 1. 分层架构

```
┌─────────────────────────────────────────┐
│          Frontend (Vue 3)               │
│  - UI Components                        │
│  - State Management (Pinia)             │
│  - Wails Runtime Bindings               │
└────────────────┬────────────────────────┘
                 │ Wails IPC
┌────────────────▼────────────────────────┐
│          App Layer (Go)                 │
│  - API Export                           │
│  - Event Emitting                       │
│  - Context Management                   │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         Service Layer                   │
│  - ClaudeService (AI Integration)       │
│  - ConversationManager                  │
│  - WorkspaceManager                     │
│  - EnvironmentDetector                  │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         Data Layer                      │
│  - JSONStorage (Conversations)          │
│  - FileStorage (Workspaces)             │
│  - Claude CLI (External Process)        │
└─────────────────────────────────────────┘
```

### 2. 模块职责

#### App Layer (`backend/app/`)
**职责:** Wails 绑定层，导出 Go 函数到前端

**核心方法:**
```go
// 环境检测 API
EnvDetectAll() (*models.EnvironmentInfo, error)
EnvDetectByName(name string) (*models.DetectionResult, error)
EnvGetStatus() (*models.EnvironmentInfo, error)
EnvClearCache() error

// 工作区管理 API
WorkspaceOpen(path string) (*models.WorkspaceInfo, error)
WorkspaceClose()
WorkspaceGetCurrent() string
WorkspaceList() []*models.WorkspaceInfo
WorkspaceListFiles() ([]*models.FileInfo, error)
WorkspaceReadFile(relativePath string) (string, error)
WorkspaceWriteFile(relativePath, content string) error
WorkspaceCreateFile(relativePath, content string) error
WorkspaceDeleteFile(relativePath string) error
WorkspaceRenameFile(oldPath, newPath string) error
WorkspaceCopyFile(srcPath, destPath string) error
WorkspaceMoveFile(srcPath, destPath string) error
WorkspaceCreateDirectory(relativePath string) error
WorkspaceSetActiveConversation(convID string) error
WorkspaceGetActiveConversation() string

// 对话管理 API
ConversationCreate(title, projectPath string) (*conversation.Conversation, error)
ConversationDelete(id string) error
ConversationInfo(id string) (*conversation.Conversation, error)
ConversationList() ([]*conversation.Conversation, error)
ConversationUpdate(conv *conversation.Conversation) error
ConversationSendWithEvents(convID, content string) error

// 系统操作 API
SystemOpenFile(relativePath string) error
SystemOpenTerminal(relativePath string) error
SystemRevealInFinder(relativePath string) error
```

#### Service Layer (`backend/service/`)
**职责:** Claude CLI 集成和流式输出处理

**核心方法:**
```go
// 设置项目路径
SetProjectPath(path string)

// 流式发送消息 (推荐)
StreamMessage(ctx context.Context, content string, onChunk func(string)) error

// 验证环境
ValidateEnvironment(ctx context.Context) error
```

**流式输出实现:**
```go
// 使用 Claude CLI 的 stream-json 格式
cmd := exec.CommandContext(ctx, "claude", "--print", content,
    "--output-format", "stream-json",
    "--verbose",
    "--include-partial-messages")

// 解析 JSON 流
scanner := bufio.NewScanner(stdout)
scanner.Buffer(buf, 1024*1024) // 1MB 缓冲区

for scanner.Scan() {
    var raw map[string]interface{}
    json.Unmarshal([]byte(line), &raw)

    // 只处理 content_block_delta 事件
    if eventType == "stream_event" {
        if eventStr == "content_block_delta" {
            onChunk(text)  // 回调函数发送到前端
        }
    }
}
```

#### Workspace Manager (`backend/manager/workspace/`)
**职责:** 工作区生命周期管理和持久化

**数据结构:**
```go
type Workspace struct {
    Path                 string    // 工作区绝对路径
    Name                 string    // 文件夹名称
    LastOpened           time.Time // 最后打开时间
    ActiveConversationID string    // 当前活跃会话 ID
}

type Manager struct {
    mu          sync.RWMutex
    workspaces  []*Workspace
    currentPath string         // 当前工作区路径
    storageFile string         // 持久化文件路径
}
```

**持久化:**
- 文件位置: `~/.claude-desktop/workspaces.json`
- 格式: JSON
- 触发时机: 打开/关闭/选择工作区、设置活跃会话时自动保存

#### Conversation Manager (`backend/manager/conversation/`)
**职责:** 对话历史管理和存储

**数据结构:**
```go
type Conversation struct {
    ID          string
    Title       string
    ProjectPath string
    Messages    []Message
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Message struct {
    ID        string
    Role      string  // "user" | "assistant"
    Content   string
    Timestamp time.Time
}
```

**存储:**
- 文件位置: `~/.claude-desktop/conversations/{id}.json`
- 格式: JSON
- 一个对话一个文件

#### Environment Detector (`backend/detector/`)
**职责:** 检测系统环境是否满足运行要求

**检测器列表:**
1. **ClaudeDetector** - 检查 Claude CLI 是否安装及版本
2. **NodeDetector** - 检查 Node.js 版本
3. **NpmDetector** - 检查 npm 是否可用
4. **GitDetector** - 检查 Git 是否安装
5. **NetworkDetector** - 检查网络连接

**检测流程:**
```go
// 1. 并行执行所有检测
for i, detector := range m.detectors {
    wg.Add(1)
    go func(index int, d Detector) {
        defer wg.Done()
        result, err := d.Detect(ctx)
        results[index] = *result
    }(i, detector)
}
wg.Wait()

// 2. 构建环境信息
envInfo := m.buildEnvironmentInfo(results)

// 3. 保存到缓存 (可选)
if m.config.EnableCache {
    m.saveToCache(envInfo)
}
```

---

## 核心功能详解

### 1. 流式对话功能

#### 工作流程

```
用户输入消息
    ↓
前端发送事件
    ↓
后端 ConversationSendWithEvents()
    ↓
发送 "claude:thinking" 事件
    ↓
前端显示"思考中..."动画
    ↓
ClaudeService.StreamMessage() 调用 Claude CLI
    ↓
解析 stream-json 输出
    ↓
通过 onChunk 回调发送 "claude:response" 事件
    ↓
前端实时更新消息内容
    ↓
发送 "claude:complete" 事件
    ↓
前端刷新缓冲区，结束思考状态
```

#### 关键代码

**后端 - 事件发送:**
```go
func (a *App) ConversationSendWithEvents(convID, content string) error {
    // 1. 发送思考开始事件
    runtime.EventsEmit(a.ctx, "claude:thinking", map[string]interface{}{
        "convID": convID,
    })

    // 2. 流式发送响应
    hasContent := false
    _, err := a.convManager.SendMessageWithCallback(convID, content, func(chunk string) {
        if strings.TrimSpace(chunk) != "" {
            hasContent = true
        }
        runtime.EventsEmit(a.ctx, "claude:response", map[string]interface{}{
            "content": chunk,
            "convID":  convID,
        })
    })

    if err != nil {
        // 发送错误事件
        runtime.EventsEmit(a.ctx, "claude:error", map[string]interface{}{
            "convID": convID,
            "error":  err.Error(),
        })
        return err
    }

    // 3. 发送完成事件
    runtime.EventsEmit(a.ctx, "claude:complete", map[string]interface{}{
        "convID":     convID,
        "hasContent": hasContent,
    })

    return nil
}
```

**前端 - 事件处理:**
```typescript
// 组件挂载时监听事件
onMounted(() => {
  EventsOn('claude:response', handleClaudeResponse);
  EventsOn('claude:thinking', handleClaudeThinking);
  EventsOn('claude:complete', handleClaudeComplete);
  EventsOn('claude:error', handleClaudeError);
});

// 处理思考事件
function handleClaudeThinking() {
  isThinking.value = true;
  thinkingMessageId = `msg-thinking-${Date.now()}`;
  messages.value.push({
    id: thinkingMessageId,
    role: 'assistant',
    content: '思考中',
    timestamp: new Date().toISOString()
  });
}

// 处理响应事件
function handleClaudeResponse(data: any) {
  const content = data?.content || '';
  if (!content.trim()) return;  // 忽略空白内容

  // 移除思考中消息
  if (thinkingMessageId) {
    const index = messages.value.findIndex(m => m.id === thinkingMessageId);
    if (index !== -1) {
      messages.value.splice(index, 1);
    }
    thinkingMessageId = null;
  }

  // 追加到流式消息
  streamingBuffer += content;

  // 16ms 防抖更新 (60fps)
  if (streamingTimer === null) {
    streamingTimer = setTimeout(() => {
      streamingMessage.content += streamingBuffer;
      streamingBuffer = '';
      streamingTimer = null;
    }, 16);
  }
}

// 处理完成事件
function handleClaudeComplete(data: any) {
  flushStreamingMessage();

  // 只有在收到内容后才结束思考状态
  if (!thinkingMessageId) {
    isThinking.value = false;
  }
  // 如果思考中消息还在，保持 isThinking = true
}
```

#### 思考中动画优化

**问题:** 思考中动画会提前消失，但实际内容还没到达

**解决方案:**
1. 前端判断: 只有收到真实文本内容才移除思考中消息
2. 事件优化: complete 事件携带 `hasContent` 标志
3. 状态保持: 如果思考中消息还在，保持 `isThinking = true`

### 2. 工作区管理

#### 工作区生命周期

```
打开工作区 (WorkspaceOpen)
    ↓
检查路径是否存在
    ↓
转换为绝对路径
    ↓
是否已在列表?
    ├─ 是 → 更新 LastOpened，设为当前
    └─ 否 → 创建新 Workspace，添加到列表
    ↓
异步保存到 ~/.claude-desktop/workspaces.json
    ↓
返回工作区信息
```

#### 文件树扫描

```go
func (m *Manager) scanDir(ctx context.Context, rootPath, relativePath string) ([]*models.FileInfo, error) {
    fullPath := filepath.Join(rootPath, relativePath)
    entries, _ := os.ReadDir(fullPath)

    for _, entry := range entries {
        // 跳过隐藏文件
        if strings.HasPrefix(entry.Name(), ".") {
            continue
        }

        fileInfo := &models.FileInfo{
            Path: relPath,
            Name: entry.Name(),
            Type: getFileType(entry.Name()),
            Icon: getFileIcon(fileInfo.Type),
        }

        if entry.IsDir() {
            // 递归扫描子目录
            subFiles, _ := m.scanDir(ctx, rootPath, relPath)
            files = append(files, subFiles...)
        }
    }

    // 排序: 目录优先，按名称排序
    sort.Slice(files, func(i, j int) bool {
        // 深度小的在前
        // 同一深度，目录在前
        // 同一类型，按名称排序
    })

    return files
}
```

#### 会话上下文持久化

**目标:** 重新打开应用后，继续之前的对话而不是只显示历史

**实现:**

1. **存储活跃会话ID:**
```go
// 创建新会话时保存
func (a *App) ConversationSendWithEvents(convID, content string) error {
    // ... 发送消息

    // 保存活跃会话ID到工作区
    a.workspaceManager.SetActiveConversationID(convID)
}
```

2. **恢复会话:**
```typescript
// 加载工作区对话
async function loadWorkspaceConversation(projectPath: string) {
    // 1. 获取存储的活跃会话ID
    const storedConvID = await WorkspaceGetActiveConversation();

    // 2. 通过项目路径查找最新会话
    const conv = await ConversationGetByProjectPath(projectPath);

    // 3. 恢复消息历史
    messages.value = conv.messages.map(...);
    conversationId.value = conv.id;

    // 4. 确保活跃会话ID已设置
    await WorkspaceSetActiveConversation(conv.id);
}
```

### 3. 文件操作

#### 支持的操作

| 操作 | 后端方法 | 前端触发 |
|------|---------|---------|
| 读取 | `WorkspaceReadFile(path)` | 双击文件 |
| 写入 | `WorkspaceWriteFile(path, content)` | 编辑器保存 |
| 创建 | `WorkspaceCreateFile(path, content)` | 新建文件 |
| 删除 | `WorkspaceDeleteFile(path)` | 右键菜单 |
| 重命名 | `WorkspaceRenameFile(old, new)` | 右键菜单 |
| 复制 | `WorkspaceCopyFile(src, dest)` | 右键菜单 |
| 移动 | `WorkspaceMoveFile(src, dest)` | 拖拽(未实现) |
| 创建目录 | `WorkspaceCreateDirectory(path)` | 新建文件夹 |

#### 系统集成

```go
// 在系统默认应用中打开
func (a *App) SystemOpenFile(relativePath string) error {
    fullPath, _ := a.workspaceManager.GetFullPath(relativePath)
    cmd := exec.Command("open", fullPath)  // macOS
    return cmd.Run()
}

// 在终端中打开
func (a *App) SystemOpenTerminal(relativePath string) error {
    fullPath, _ := a.workspaceManager.GetFullPath(relativePath)

    // 如果是文件，获取其目录
    if !info.IsDir() {
        dirPath = filepath.Dir(fullPath)
    }

    cmd := exec.Command("open", "-a", "Terminal", dirPath)
    return cmd.Run()
}

// 在 Finder 中显示
func (a *App) SystemRevealInFinder(relativePath string) error {
    fullPath, _ := a.workspaceManager.GetFullPath(relativePath)
    cmd := exec.Command("open", "-R", fullPath)
    return cmd.Run()
}
```

### 4. 环境检测

#### 检测器接口

```go
type Detector interface {
    // Detect 执行检测
    Detect(ctx context.Context) (*DetectionResult, error)

    // Name 返回检测器名称
    Name() string

    // Required 是否必需项
    Required() bool
}

type DetectionResult struct {
    Name     string    // 检测项名称
    Status   string    // "success" | "failed" | "warning"
    Message  string    // 详细信息
    Version  string    // 版本号 (可选)
    Required bool      // 是否必需
    CheckedAt time.Time // 检测时间
}
```

#### 检测流程

```
用户触发检测
    ↓
检查缓存是否启用且有效?
    ├─ 是 → 直接返回缓存结果
    └─ 否 → 继续
    ↓
并行执行所有检测器
    ├─ ClaudeDetector
    ├─ NodeDetector
    ├─ NpmDetector
    ├─ GitDetector
    └─ NetworkDetector
    ↓
收集所有检测结果
    ↓
计算整体状态
    ├─ success: 所有必需项通过
    ├─ partial: 部分必需项通过
    └─ failed: 无必需项通过
    ↓
保存到缓存 (~/.claude-desktop/cache/env_check.json)
    ↓
返回环境信息
```

#### Claude 检测器实现

```go
type ClaudeDetector struct {
    minVersion string
}

func (d *ClaudeDetector) Detect(ctx context.Context) (*DetectionResult, error) {
    // 执行 claude --version
    cmd := exec.CommandContext(ctx, "claude", "--version")
    output, err := cmd.Output()

    if err != nil {
        return &DetectionResult{
            Name:     "Claude CLI",
            Status:   "failed",
            Message:  "Claude CLI 未安装",
            Required: true,
        }, nil
    }

    // 解析版本
    version := strings.TrimSpace(string(output))

    // 检查版本是否满足要求
    if !d.checkVersion(version) {
        return &DetectionResult{
            Name:     "Claude CLI",
            Status:   "failed",
            Version:  version,
            Message:  fmt.Sprintf("版本过低，需要 %s 以上", d.minVersion),
            Required: true,
        }, nil
    }

    return &DetectionResult{
        Name:     "Claude CLI",
        Status:   "success",
        Version:  version,
        Message:  "已安装",
        Required: true,
    }, nil
}
```

---

## 前端架构

### 1. 组件结构

```
App.vue (根组件)
    ├─ WelcomeView (欢迎页)
    │   ├─ 环境检测卡片
    │   ├─ 检测按钮
    │   └─ 状态显示
    │
    └─ MainView (主界面)
        ├─ 侧边栏
        │   ├─ 文件树
        │   ├─ 文件右键菜单
        │   └─ 工作区切换
        │
        ├─ 聊天区域
        │   ├─ 消息列表
        │   │   ├─ 用户消息
        │   │   ├─ AI 消息
        │   │   └─ 思考中动画
        │   ├─ 输入框
        │   └─ 发送/停止按钮
        │
        └─ 欢迎界面 (无消息时显示)
```

### 2. 状态管理 (Pinia)

#### Env Store (`frontend/src/stores/env.ts`)

```typescript
export const useEnvStore = defineStore('env', () => {
  // 状态
  const envInfo = ref<EnvironmentInfo | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // 方法
  async function detectAll(): Promise<void>
  async function detectByName(name: string): Promise<void>
  async function getStatus(): Promise<void>
  function clearError(): void

  return {
    envInfo, loading, error,
    detectAll, detectByName, getStatus, clearError
  };
});
```

#### Workspace Store (`frontend/src/stores/workspace.ts`)

```typescript
export const useWorkspaceStore = defineStore('workspace', () => {
  // 状态
  const workspaces = ref<WorkspaceInfo[]>([]);
  const currentPath = ref<string>('');
  const workspaceInfo = ref<WorkspaceInfo | null>(null);
  const files = ref<FileInfo[]>([]);
  const loading = ref(false);

  // 计算属性
  const isOpen = computed(() => currentPath.value !== '');
  const workspaceName = computed(() => workspaceInfo.value?.name || '');
  const currentWorkspace = computed(() =>
    workspaces.value.find(ws => ws.path === currentPath.value)
  );

  // 方法
  async function loadWorkspaces(): Promise<void>
  async function openFolder(path: string): Promise<WorkspaceInfo>
  async function selectWorkspace(path: string): Promise<void>
  async function removeWorkspace(path: string): Promise<void>
  async function closeFolder(): Promise<void>
  async function loadFiles(): Promise<void>
  async function readFile(path: string): Promise<string>
  async function writeFile(path: string, content: string): Promise<void>
  async function deleteFile(path: string): Promise<void>
  async function createFile(path: string, content: string): Promise<void>
  async function refreshInfo(): Promise<void>
  function clearError(): void

  return {
    workspaces, currentPath, workspaceInfo, files, loading, error,
    isOpen, workspaceName, currentWorkspace,
    loadWorkspaces, openFolder, selectWorkspace, removeWorkspace,
    closeFolder, loadFiles, readFile, writeFile, deleteFile,
    createFile, refreshInfo, clearError
  };
});
```

### 3. 文件树交互

#### 双击文件

```vue
<div @dblclick="sendFilePathToInput(file)">
  {{ file.name }}
</div>
```

```typescript
async function sendFilePathToInput(file: FileInfo) {
  if (!selectedWorkspace.value) {
    alert('请先选择工作区');
    return;
  }

  // 计算相对路径
  const relativePath = file.path.replace(selectedWorkspace.value.path + '/', '');
  const pathMessage = `@${relativePath} `;  // 路径后加空格

  // 添加到输入框
  messageInput.value += (messageInput.value ? '\n' : '') + pathMessage;

  // 关闭右键菜单
  closeContextMenu();

  // 激活输入框并聚焦
  await nextTick();
  if (messageInputRef.value) {
    messageInputRef.value.focus();
    // 将光标移动到文本末尾
    messageInputRef.value.setSelectionRange(
      messageInput.value.length,
      messageInput.value.length
    );
  }
}
```

#### 右键菜单

```typescript
function handleContextMenu(event: MouseEvent, file: FileInfo) {
  event.preventDefault();
  contextMenu.value = {
    visible: true,
    x: event.clientX,
    y: event.clientY,
    file: file
  };
}

function closeContextMenu() {
  contextMenu.value = {
    visible: false,
    x: 0,
    y: 0,
    file: null
  };
}
```

```vue
<div v-if="contextMenu.visible"
     class="context-menu"
     :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
  <div class="context-menu-item" @click="sendFilePathToInput(contextMenu.file!)">
    📎 发送路径到输入框
  </div>
  <div class="context-menu-item" @click="renameFile(contextMenu.file!)">
    ✏️ 重命名
  </div>
  <div class="context-menu-item danger" @click="deleteFile(contextMenu.file!)">
    🗑️ 删除
  </div>
</div>
```

### 4. 智能滚动

**需求:** 只有当用户在消息列表底部时，才自动滚动

**实现:**

```typescript
// 检查是否在底部 (100px 阈值内)
function isNearBottom(): boolean {
  if (!messageListRef.value) return false;
  const el = messageListRef.value;
  const threshold = 100;
  return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
}

// 监听消息变化
watch(() => messages.value, async () => {
  if (isNearBottom()) {
    await nextTick();
    scrollToBottom();
  }
}, { deep: true });

function scrollToBottom() {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
  }
}
```

### 5. 消息显示优化

**模板条件渲染:**

```vue
<div
  v-show="msg.content.trim() !== '' || msg.role === 'user' || msg.id.includes('thinking')"
  class="message-item"
  :class="msg.role"
>
  <!-- 思考中动画 -->
  <div v-if="msg.id.includes('thinking')" class="message-content thinking-content">
    <span class="thinking-text">思考中</span>
    <span class="thinking-dots">
      <span class="dot"></span>
      <span class="dot"></span>
      <span class="dot"></span>
    </span>
  </div>

  <!-- 普通消息 -->
  <div v-else class="message-content">
    {{ msg.content }}
  </div>
</div>
```

**CSS 动画:**

```css
.thinking-dots {
  display: flex;
  gap: 4px;
}

.thinking-dots .dot {
  width: 8px;
  height: 8px;
  background-color: #666;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}

.thinking-dots .dot:nth-child(1) {
  animation-delay: -0.32s;
}

.thinking-dots .dot:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}
```

---

## 数据流详解

### 1. 发送消息流程

```
用户输入 "帮我分析这个文件"
    ↓
点击发送按钮
    ↓
前端 handleSendMessage()
    ↓
创建用户消息对象
messages.value.push({ role: 'user', content: '...' })
    ↓
判断是否有会话ID?
    ├─ 否 → 创建新会话 ConversationCreate()
    │   ↓
    │   保存会话ID到工作区 WorkspaceSetActiveConversation()
    │
    └─ 是 → 使用现有会话
    ↓
调用 ConversationSendWithEvents(convID, content)
    ↓
后端发送 "claude:thinking" 事件
    ↓
前端显示思考中动画
    ↓
后端调用 Claude CLI
    ↓
解析流式 JSON 输出
    ↓
逐个发送 "claude:response" 事件
    ↓
前端实时更新消息内容
    ↓
CLI 执行完成
    ↓
后端保存对话到文件
    ↓
后端发送 "claude:complete" 事件
    ↓
前端刷新缓冲区，结束思考状态
```

### 2. 打开工作区流程

```
用户选择工作区文件夹
    ↓
前端调用 WorkspaceOpen(path)
    ↓
后端检查路径是否存在
    ↓
转换为绝对路径
    ↓
是否已在工作区列表?
    ├─ 是 → 更新 LastOpened，设为当前工作区
    └─ 否 → 创建新 Workspace 对象，添加到列表
    ↓
异步保存到 ~/.claude-desktop/workspaces.json
    ↓
返回工作区信息
    ↓
前端更新状态
    ↓
扫描文件树
    ↓
前端显示文件列表
    ↓
加载该工作区的历史对话
    ↓
调用 WorkspaceGetActiveConversation() 获取活跃会话ID
    ↓
调用 ConversationGetByProjectPath() 恢复消息历史
    ↓
前端显示历史消息
```

### 3. 环境检测流程

```
用户点击"检测环境"按钮
    ↓
前端调用 EnvDetectAll()
    ↓
后端检查缓存
    ↓
缓存有效?
    ├─ 是 → 直接返回缓存结果
    └─ 否 → 继续执行检测
        ↓
并行执行所有检测器 (goroutine)
    ├─ ClaudeDetector → 执行 claude --version
    ├─ NodeDetector → 执行 node --version
    ├─ NpmDetector → 执行 npm --version
    ├─ GitDetector → 执行 git --version
    └─ NetworkDetector → 请求 GitHub API
    ↓
收集所有检测结果
    ↓
计算整体状态 (success/partial/failed)
    ↓
保存到缓存 (~/.claude-desktop/cache/env_check.json)
    ↓
返回环境信息
    ↓
前端更新 UI 显示检测结果
```

---

## 关键技术点

### 1. 流式输出优化

**问题:** 逐字符更新导致性能问题和卡顿

**解决方案: 防抖 + 批量更新**

```typescript
let streamingBuffer = '';
let streamingTimer: number | null = null;

function handleClaudeResponse(data: any) {
  const content = data?.content || '';
  if (!content.trim()) return;

  // 立即追加到缓冲区
  streamingBuffer += content;

  // 16ms 防抖 (约 60fps)
  if (streamingTimer === null) {
    streamingTimer = setTimeout(() => {
      // 批量更新到消息对象
      streamingMessage.content += streamingBuffer;
      streamingBuffer = '';
      streamingTimer = null;
    }, 16);
  }
}
```

### 2. 并发安全

**后端互斥锁:**

```go
type ClaudeService struct {
    mu          sync.Mutex
    projectPath string
}

func (s *ClaudeService) SetProjectPath(path string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.projectPath = path
}
```

```go
type WorkspaceManager struct {
    mu     sync.RWMutex
    // ...
}

func (m *Manager) Open(path string) (*Workspace, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    // 修改工作区列表
}

func (m *Manager) GetCurrent() string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.currentPath
}
```

**前端状态管理:**
- 使用 Vue 3 的 ref/reactive 确保响应式
- 事件监听器正确注册和注销
- 避免内存泄漏

### 3. 错误处理

**后端错误处理:**

```go
func (a *App) ConversationSendWithEvents(convID, content string) error {
    _, err := a.convManager.SendMessageWithCallback(convID, content, func(chunk string) {
        runtime.EventsEmit(a.ctx, "claude:response", ...)
    })

    if err != nil {
        // 发送错误事件到前端
        runtime.EventsEmit(a.ctx, "claude:error", map[string]interface{}{
            "convID": convID,
            "error":  err.Error(),
        })
        return err
    }

    return nil
}
```

**前端错误处理:**

```typescript
function handleClaudeError(data: any) {
  const errorMsg = data?.error || '未知错误';

  // 移除思考中消息
  if (thinkingMessageId) {
    const index = messages.value.findIndex(m => m.id === thinkingMessageId);
    if (index !== -1) {
      messages.value.splice(index, 1);
    }
    thinkingMessageId = null;
  }

  // 添加错误消息
  messages.value.push({
    id: `msg-error-${Date.now()}`,
    role: 'assistant',
    content: `发生错误: ${errorMsg}`,
    timestamp: new Date().toISOString()
  });

  isThinking.value = false;
}
```

### 4. 持久化策略

**工作区持久化:**
- 时机: 打开/关闭/选择工作区、设置活跃会话时
- 方式: 异步保存 (goroutine)
- 文件: `~/.claude-desktop/workspaces.json`

```go
func (m *Manager) saveToStorage() {
    go func() {
        m.mu.RLock()
        data := json.Marshal(m.workspaces)
        m.mu.RUnlock()

        os.WriteFile(m.storageFile, data, 0644)
    }()
}
```

**对话持久化:**
- 时机: 每次发送消息后
- 方式: 同步保存
- 文件: `~/.claude-desktop/conversations/{id}.json`

```go
func (s *JSONStorage) SaveConversation(conv *Conversation) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    filename := filepath.Join(s.convDir, conv.ID+".json")
    data, _ := json.MarshalIndent(conv, "", "  ")
    return os.WriteFile(filename, data, 0644)
}
```

---

## 开发指南

### 1. 环境准备

**必需软件:**
- Go 1.24+
- Node.js 18+
- Claude CLI
- Wails CLI

**安装步骤:**

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆项目
git clone <repo-url>
cd claude_desktop

# 安装前端依赖
cd frontend
npm install

# 返回项目根目录
cd ..
```

### 2. 开发模式

**启动开发服务器:**

```bash
wails dev
```

这会:
1. 启动后端 Go 程序
2. 启动前端 Vite 开发服务器
3. 启用热重载
4. 打开浏览器 (http://localhost:34115)

**调试:**
- 后端: 使用 Go IDE (如 GoLand) 的调试功能
- 前端: 使用浏览器开发者工具
- 日志: 后端日志在终端，前端日志在浏览器控制台

### 3. 构建发布

**开发版本:**

```bash
wails build
```

**生产版本:**

```bash
wails build -production
```

**输出:**
- macOS: `build/bin/claude_desktop.app`
- Windows: `build/bin/claude_desktop.exe`
- Linux: `build/bin/claude_desktop`

### 4. 添加新功能

**步骤:**

1. **后端 API:**
   ```go
   // backend/app/app.go
   func (a *App) NewFeature(param string) (Result, error) {
       // 实现逻辑
       return result, nil
   }
   ```

2. **重新生成绑定:**
   ```bash
   wails generate module
   ```

3. **前端调用:**
   ```typescript
   import { NewFeature } from '../../wailsjs/go/app/App';

   async function handleNewFeature() {
       const result = await NewFeature('param');
       console.log(result);
   }
   ```

4. **前端 UI:**
   ```vue
   <button @click="handleNewFeature">执行功能</button>
   ```

### 5. 添加新检测器

**创建检测器:**

```go
// backend/detector/my_detector.go
package detector

import (
    "context"
    "fmt"
)

type MyDetector struct {
    config Config
}

func NewMyDetector(config Config) *MyDetector {
    return &MyDetector{config: config}
}

func (d *MyDetector) Detect(ctx context.Context) (*DetectionResult, error) {
    // 执行检测逻辑
    cmd := exec.CommandContext(ctx, "my-tool", "--version")
    output, err := cmd.Output()

    if err != nil {
        return &DetectionResult{
            Name:     "My Tool",
            Status:   "failed",
            Message:  "未安装",
            Required: false,
        }, nil
    }

    return &DetectionResult{
        Name:     "My Tool",
        Status:   "success",
        Version:  string(output),
        Message:  "已安装",
        Required: false,
    }, nil
}

func (d *MyDetector) Name() string {
    return "MyTool"
}

func (d *MyDetector) Required() bool {
    return false
}
```

**注册检测器:**

```go
// backend/detector/manager.go
func (m *Manager) initDetectors() {
    m.detectors = []Detector{
        // ... 现有检测器
        NewMyDetector(m.config),
    }
}
```

### 6. 调试技巧

**查看 Claude CLI 输出:**

```go
// backend/service/claude_service.go
fmt.Printf("Claude stdout: %s\n", line)
```

**查看前端事件:**

```typescript
EventsOn('claude:response', (data) => {
  console.log('收到响应事件:', data);
});
```

**查看工作区文件:**

```bash
# macOS
cat ~/.claude-desktop/workspaces.json

# 查看对话
ls ~/.claude-desktop/conversations/
cat ~/.claude-desktop/conversations/{conversation-id}.json
```

---

## 常见问题

### 1. 思考中动画消失问题

**原因:**
- 后端发送工具调用信息导致动画提前移除
- complete 事件在收到内容前触发

**解决:**
- 后端只发送真正的文本内容
- 前端判断只有收到内容才移除思考中消息
- complete 事件携带 hasContent 标志

### 2. 流式输出卡顿

**原因:**
- 逐字符更新 Vue 组件导致频繁重渲染

**解决:**
- 使用 16ms 防抖批量更新
- 缓冲区累积内容后一次性更新

### 3. 工作区路径问题

**原因:**
- 相对路径和绝对路径混用
- 路径未标准化

**解决:**
- 所有路径转换为绝对路径
- 使用 `filepath.Abs()` 标准化

### 4. 会话上下文丢失

**原因:**
- 只恢复消息历史，未保存活跃会话ID

**解决:**
- 工作区持久化保存 ActiveConversationID
- 重新打开时恢复会话ID

---

## 性能优化

### 1. 文件树扫描优化

**问题:** 递归扫描大项目很慢

**优化:**
- 使用 context 支持取消
- 并行扫描子目录
- 跳过隐藏文件和 node_modules

```go
for _, entry := range entries {
    // 跳过隐藏文件
    if strings.HasPrefix(entry.Name(), ".") {
        continue
    }

    // 跳过 node_modules
    if entry.Name() == "node_modules" {
        continue
    }

    // 检查 context 是否已取消
    select {
    case <-ctx.Done():
        return files, ctx.Err()
    default:
    }
}
```

### 2. 对话列表加载优化

**问题:** 对话多时加载慢

**优化:**
- 分页加载 (未实现)
- 懒加载对话详情
- 缓存最近使用的对话

### 3. 流式输出缓冲区

**优化:**
- 使用 1MB 缓冲区处理长 JSON 行
- 扫描器使用动态缓冲区

```go
scanner := bufio.NewScanner(stdout)
buf := make([]byte, 0, 64*1024)
scanner.Buffer(buf, 1024*1024)  // 最大 1MB
```

---

## 未来改进方向

### 1. 功能增强

- [ ] 支持多会话并发
- [ ] 支持流式输出导出
- [ ] 支持代码高亮显示
- [ ] 支持消息搜索和过滤
- [ ] 支持对话分支管理
- [ ] 支持快捷键
- [ ] 支持拖拽文件
- [ ] 支持主题切换

### 2. 性能优化

- [ ] 虚拟滚动长消息列表
- [ ] Web Worker 处理消息解析
- [ ] IndexedDB 缓存对话
- [ ] 增量文件扫描

### 3. 用户体验

- [ ] 更丰富的错误提示
- [ ] 加载进度指示
- [ ] 离线模式支持
- [ ] 多语言支持
- [ ] 自定义配置

---

## 附录

### A. 配置文件

**wails.json:**
```json
{
  "name": "claude_desktop",
  "outputfilename": "claude_desktop",
  "assetdir": "frontend/dist",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:build": "npm run build-only",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "debounceMS": 500,
  "author": {
    "name": "silaswei",
    "email": "silaswei.com"
  }
}
```

### B. 环境变量

- `HOME`: 用户主目录
- `PATH`: 系统 PATH
- `CLAUDE_API_KEY`: Claude API 密钥 (如果需要)

### C. 数据目录

**macOS/Linux:**
```
~/.claude-desktop/
├── conversations/     # 对话历史
│   ├── {id}.json
│   └── ...
├── cache/             # 缓存
│   └── env_check.json
└── workspaces.json    # 工作区列表
```

### D. 依赖版本

**Go:**
```
github.com/wailsapp/wails/v2 v2.11.0
```

**前端:**
```json
{
  "vue": "^3.3.4",
  "pinia": "^2.1.6",
  "typescript": "^5.2.2",
  "vite": "^4.4.9"
}
```

---

## 总结

Claude Desktop 是一个功能完整的 AI 对话桌面应用，具有以下特点:

1. **流式对话**: 实时显示 Claude 响应，用户体验流畅
2. **工作区管理**: 支持多项目切换，自动保存上下文
3. **环境检测**: 自动检测运行环境，确保功能正常
4. **文件操作**: 集成文件系统操作，方便项目文件管理
5. **状态持久化**: 所有状态自动保存，关闭不丢失

通过本文档，你应该能够:
- 理解项目的整体架构
- 了解各个模块的职责
- 掌握核心功能的实现原理
- 快速上手开发新功能
- 解决常见问题

祝你开发顺利！
