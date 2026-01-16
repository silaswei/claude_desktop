# Claude 工作区对话助手 - 产品设计文档

## 一、产品定位

### 1.1 核心价值
**基于本地文件夹的 AI 对话工作台**，用户可以打开任意文件夹作为工作区，通过与 Claude 对话来处理、分析、操作工作区中的文件（Excel、图片、文档、代码等）。

### 1.2 目标用户
- 数据分析师：处理 Excel、CSV 数据文件
- 内容创作者：批量处理图片、文档
- 开发者：快速查看、编辑代码文件
- 办公人员：自动化处理文档、报表

### 1.3 使用场景
1. **数据分析**：打开数据文件夹，让 Claude 分析 Excel、生成图表
2. **批量处理**：选择图片文件夹，让 Claude 批量调整、转换图片
3. **文档处理**：打开文档文件夹，让 Claude 总结、翻译、提取信息
4. **代码辅助**：打开代码项目，让 Claude 解释、修改代码

---

## 二、核心功能

### 2.1 工作区管理
- 打开任意文件夹作为工作区
- 显示工作区文件列表
- 快速预览文件内容
- 记录最近使用的工作区

### 2.2 AI 对话
- 与 Claude 进行多轮对话
- 支持上下文记忆
- 流式输出响应
- Markdown 渲染（代码高亮、表格、公式）

### 2.3 文件处理
- 上传文件到对话
- 在对话中引用工作区文件
- 文件内容预览
- 文件格式转换
- 批量文件操作

### 2.4 文件操作
- 读取文件内容
- 写入/修改文件
- 创建新文件
- 删除文件
- 重命名文件

---

## 三、界面设计

### 3.1 整体布局

```
┌─────────────────────────────────────────────────────────┐
│  Claude Desktop                 [📁 工作区] [⚙️ 设置]  │
├──────────────────┬──────────────────────────────────────┤
│  历史对话        │                                       │
│  ─────────────   │  💬 对话区域                          │
│  📊 数据分析     │                                       │
│  📝 报告生成     │  👤 用户: 帮我分析这个 Excel 表格      │
│  🖼️ 图片处理     │                                       │
│  💻 代码项目     │  🤖 Claude: 我来帮您分析...           │
│                  │                                       │
│  ─────────────   │  [📎 附件] [📁 引用文件]              │
│  工作区:         │  [输入消息...]            [发送 ↗]    │
│  📂 ~/Documents  │                                       │
│  📄 sales.xlsx   │                                       │
│  📄 report.docx  │                                       │
│  🖼️ product.png │                                       │
└──────────────────┴──────────────────────────────────────┘
```

### 3.2 界面元素

#### 左侧边栏（可折叠）
1. **历史对话列表**
   - 对话标题
   - 时间戳
   - 关联工作区
   - 新建对话按钮

2. **工作区信息**
   - 当前工作区路径
   - 文件列表（树形结构）
   - 文件预览

#### 主对话区
1. **消息列表**
   - 用户消息（右对齐）
   - Claude 响应（左对齐）
   - Markdown 渲染
   - 代码高亮
   - 表格展示

2. **输入面板**
   - 文本输入框（多行）
   - 附件上传按钮
   - 引用工作区文件按钮
   - 发送按钮

#### 工具栏
1. **工作区切换**
   - 打开文件夹
   - 最近工作区
   - 清空工作区

2. **设置**
   - API 配置
   - 界面主题
   - 快捷键

---

## 四、交互流程

### 4.1 首次使用流程
```
启动应用
    ↓
打开文件夹（选择工作区）
    ↓
进入对话界面
    ↓
开始对话
```

### 4.2 对话流程
```
用户输入消息
    ↓
可选择：上传文件 / 引用工作区文件
    ↓
发送给 Claude
    ↓
Claude 处理（可能读取文件）
    ↓
流式返回响应
    ↓
保存对话历史
```

### 4.3 文件处理流程
```
方式1: 上传文件
用户点击附件 → 选择文件 → 文件上传到临时目录 → 发送给 Claude

方式2: 引用工作区文件
用户点击引用文件 → 浏览工作区 → 选择文件 → Claude 读取文件
```

---

## 五、数据模型

### 5.1 对话（Conversation）
```typescript
{
  id: string;              // 对话 ID
  title: string;           // 对话标题（自动生成或用户编辑）
  workspacePath: string;   // 关联的工作区路径
  createdAt: string;       // 创建时间
  updatedAt: string;       // 更新时间
  messages: Message[];     // 消息列表
}
```

### 5.2 消息（Message）
```typescript
{
  id: string;              // 消息 ID
  role: 'user' | 'assistant';
  content: string;         // 消息内容
  timestamp: string;       // 时间戳
  attachments?: Attachment[];  // 附件列表
  fileReferences?: FileRef[];  // 引用的文件
}
```

### 5.3 工作区（Workspace）
```typescript
{
  path: string;            // 工作区路径
  name: string;            // 文件夹名称
  files: FileInfo[];       // 文件列表
  lastOpened: string;      // 最后打开时间
}
```

### 5.4 文件信息（FileInfo）
```typescript
{
  path: string;            // 文件路径
  name: string;            // 文件名
  type: string;            // 文件类型
  size: number;            // 文件大小
  modifiedAt: string;      // 修改时间
}
```

---

## 六、技术实现

### 6.1 后端（Go）

#### 工作区管理
```go
type WorkspaceManager struct {
    currentPath string
    fileScanner *FileScanner
}

// 打开工作区
func (m *WorkspaceManager) Open(path string) error

// 获取文件列表
func (m *WorkspaceManager) ListFiles() []FileInfo

// 读取文件内容
func (m *WorkspaceManager) ReadFile(path string) (string, error)

// 写入文件
func (m *WorkspaceManager) WriteFile(path, content string) error
```

#### 对话管理
```go
type ConversationManager struct {
    storage Storage
    claude  *ClaudeService
}

// 创建对话
func (m *ConversationManager) Create(title, workspacePath string) (*Conversation, error)

// 发送消息（支持文件附件）
func (m *ConversationManager) Send(convID, content string, files []string) (*Conversation, error)
```

#### 文件处理
```go
type FileService struct {
    workspace *WorkspaceManager
}

// 支持的文件类型
const (
    FileImage = "image/*"
    FileExcel = "application/vnd.ms-excel"
    FileWord  = "application/vnd.ms-word"
    FilePDF   = "application/pdf"
    FileText  = "text/*"
)

// 读取文件
func (s *FileService) Read(path string) ([]byte, error)

// 获取文件预览
func (s *FileService) Preview(path string) (string, error)
```

### 6.2 前端（Vue 3）

#### 状态管理
```typescript
// workspace.ts
export const useWorkspaceStore = defineStore('workspace', () => {
  const currentPath = ref<string>('');
  const files = ref<FileInfo[]>([]);

  async function openFolder(path: string) { }
  async function loadFiles() { }
  async function readFile(path: string) { }

  return { currentPath, files, openFolder, loadFiles, readFile };
});
```

#### 界面组件
- `WorkspaceSidebar.vue` - 工作区边栏
- `ConversationList.vue` - 对话列表
- `MessageList.vue` - 消息列表
- `MessageItem.vue` - 消息项
- `InputPanel.vue` - 输入面板
- `FileUpload.vue` - 文件上传
- `FileBrowser.vue` - 文件浏览器

---

## 七、功能优先级

### P0（核心功能）
- ✅ 打开文件夹作为工作区
- ✅ AI 对话（基本功能）
- ✅ 工作区文件列表
- ✅ 对话历史管理

### P1（重要功能）
- ⏳ 文件上传
- ⏳ 引用工作区文件
- ⏳ 文件预览
- ⏳ Markdown 渲染

### P2（增强功能）
- ⏳ 文件操作（编辑、创建、删除）
- ⏳ 批量文件处理
- ⏳ 导出对话
- ⏳ 搜索对话

### P3（优化功能）
- ⏳ 快捷键
- ⏳ 主题切换
- ⏳ 多工作区切换
- ⏳ 数据统计

---

## 八、与原设计的差异

| 原设计 | 新设计 |
|--------|--------|
| 项目管理（扫描代码项目） | 工作区管理（打开任意文件夹） |
| 按编程语言分类 | 不分类，显示所有文件 |
| 强调代码项目 | 支持各种文件类型（Excel、图片、文档） |
| 项目扫描器 | 文件浏览器 |
| 代码语言检测 | 文件类型识别 |

---

## 九、开发计划

### 阶段一：核心功能（当前）
1. ✅ 修改后端 - 移除项目扫描，改为工作区管理
2. ⏳ 修改前端 - 简化界面，以对话为主
3. ⏳ 实现文件列表显示

### 阶段二：文件处理
1. ⏳ 文件上传功能
2. ⏳ 文件预览功能
3. ⏳ 引用工作区文件

### 阶段三：优化体验
1. ⏳ Markdown 渲染
2. ⏳ 代码高亮
3. ⏳ 快捷键支持

---

**更新时间**: 2026-01-09
**版本**: v2.0 - 工作区对话助手
