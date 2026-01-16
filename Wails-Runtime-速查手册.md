# Wails 开发速查手册

> 从 Wails 官方文档整理

## 目录
- [CLI 命令行](#cli-命令行)
- [运行时简介](#运行时简介)
- [Events 事件](#events-事件)
- [Window 窗口](#window-窗口)
- [Dialog 对话框](#dialog-对话框)
- [Browser 浏览器](#browser-浏览器)
- [Log 日志](#log-日志)
- [Menu 菜单](#menu-菜单)
- [Clipboard 剪贴板](#clipboard-剪贴板)

---

## CLI 命令行

所有命令格式：`wails <命令> <标志>`

### wails init - 初始化项目

用于生成新项目。

**常用标志：**

| 标志 | 描述 | 默认值 |
|------|------|--------|
| `-n` | 项目名称（必填） | - |
| `-d` | 项目目录 | 项目名 |
| `-t` | 模板名称或 URL | vanilla |
| `-g` | 初始化 git 仓库 | - |
| `-ide` | 生成 IDE 项目文件（vscode 或 goland） | - |
| `-l` | 列出可用项目模板 | - |
| `-f` | 强制构建应用 | false |
| `-q` | 静默模式 | - |

**示例：**

```bash
# 基础项目
wails init -n myproject

# 完整配置示例
wails init -n test -d mytestproject -g -ide vscode -q

# 使用远程模板
wails init -n test -t https://github.com/leaanthony/testtemplate

# 使用特定版本的远程模板
wails init -n test -t https://github.com/leaanthony/testtemplate@v1.0.0

# 查看可用模板
wails init -l
```

**远程模板支持：**
- 支持托管在 GitHub 上的模板
- 可以使用 `[@version]` 指定版本
- 参考社区模板列表（注意：第三方模板不由 Wails 维护）

---

### wails dev - 开发模式

实时开发模式，自动重新构建和加载。

**功能：**
- 自动更新 go.mod 到与 CLI 相同的 Wails 版本
- 编译并自动运行应用
- 监听 Go 文件变化，自动重新构建
- 启动开发服务器（默认 http://localhost:34115）
- 资源文件修改自动重新加载（不重新构建）
- 生成带 JSDoc 的 JS 包装器和 TypeScript 类型定义
- 生成运行时包装器和 TS 声明

**常用标志：**

| 标志 | 描述 | 默认值 |
|------|------|--------|
| `-browser` | 启动时打开浏览器 | - |
| `-assetdir` | 资源文件目录 | wails.json 配置 |
| `-frontenddevserverurl` | 第三方开发服务器 URL（如 Vite） | - |
| `-devserver` | 开发服务器地址 | localhost:34115 |
| `-debounce` | 资源更改后等待时间（毫秒） | 100 |
| `-reloaddirs` | 额外的重新加载目录 | wails.json 配置 |
| `-wailsjsdir` | 生成 JS 模块的目录 | wails.json 配置 |
| `-loglevel` | 日志级别（Trace/Debug/Info/Warning/Error） | Debug |
| `-race` | 使用竞态检测器 | false |
| `-skipbindings` | 跳过 bindings 生成 | - |
| `-s` | 跳过前端构建 | false |
| `-nosyncgomod` | 不同步 go.mod 中的 Wails 版本 | false |
| `-noreload` | 禁用自动重新加载 | - |
| `-nocolour` | 禁用彩色输出 | - |
| `-v` | 详细级别 (0-2) | 1 |
| `-compiler` | 指定 Go 编译器 | go |
| `-save` | 保存配置到 wails.json | - |

**示例：**

```bash
# 基础开发模式
wails dev

# 使用 Vite 开发服务器
wails dev -frontenddevserverurl http://localhost:3000

# 自定义资源目录并打开浏览器
wails dev -assetdir ./frontend/dist -browser

# 保存配置供后续使用
wails dev -assetdir ./frontend/dist -wailsjsdir ./frontend/src -save
```

---

### wails build - 构建应用

编译为生产可用的二进制文件。

**常用标志：**

| 标志 | 描述 | 默认值 |
|------|------|--------|
| `-clean` | 清理 build/bin 目录 | - |
| `-o` | 输出文件名 | - |
| `-platform` | 目标平台（逗号分隔） | 当前系统 |
| `-ldflags` | 传递给编译器的额外 ldflags | - |
| `-tags` | 构建标签 | - |
| `-trimpath` | 删除所有文件系统路径 | - |
| `-u` | 更新到与 CLI 相同的 Wails 版本 | - |
| `-s` | 跳过前端构建 | - |
| `-skipbindings` | 跳过 bindings 生成 | - |
| `-nopackage` | 不打包应用程序 | - |
| `-nsis` | 为 Windows 生成 NSIS 安装程序 | - |
| `-upx` | 使用 upx 压缩二进制文件 | - |
| `-upxflags` | 传递给 upx 的标志 | - |
| `-garbleargs` | 传递给 garble 的参数 | -literals -tiny -seed=random |
| `-obfuscated` | 使用 garble 混淆应用 | - |
| `-webview2` | WebView2 安装策略（download/embed/browser/error） | download |
| `-debug` | 保留调试信息并显示调试控制台 | - |
| `-devtools` | 允许使用 DevTools（仅调试用） | - |
| `-race` | 使用竞态检测器 | - |
| `-compiler` | 指定 Go 编译器 | go |
| `-nosyncgomod` | 不同步 go.mod | - |
| `-nocolour` | 禁用彩色输出 | - |
| `-v` | 详细级别 (0-2) | 1 |
| `-dryrun` | 打印命令但不执行 | - |
| `-f` | 强制构建 | - |
| `-m` | 编译前跳过 mod tidy | - |
| `-windowsconsole` | Windows 保留控制台窗口 | - |

**支持的平台：**

| 平台 | 描述 |
|------|------|
| darwin | macOS（当前构建机器架构） |
| darwin/amd64 | macOS 10.13+ AMD64 |
| darwin/arm64 | macOS 11.0+ ARM64 |
| darwin/universal | macOS 通用应用（AMD64+ARM64） |
| windows | Windows 10/11（当前架构） |
| windows/amd64 | Windows 10/11 AMD64 |
| windows/arm64 | Windows 10/11 ARM64 |
| linux | Linux（当前架构） |
| linux/amd64 | Linux AMD64 |
| linux/arm64 | Linux ARM64 |

**示例：**

```bash
# 基础构建
wails build

# 清理构建并指定输出名
wails build -clean -o myapp.exe

# 交叉编译到 Windows ARM64
wails build -platform windows/arm64

# 混淆构建
wails build -obfuscated

# 使用 UPX 压缩
wails build -upx

# 生成 NSIS 安装程序
wails build -nsis
```

**重要提示：**
- Mac 上使用 `Info.plist`（开发模式用 `Info.dev.plist`）
- UPX 在 Apple 芯片上有已知问题
- UPX 压缩的文件可能被防病毒软件误报
- `-devtools` 会导致应用无法通过 Mac App Store 审核
- 设置最低 macOS 版本：`CGO_CFLAGS=-mmacosx-version-min=10.15.0 CGO_LDFLAGS=-mmacosx-version-min=10.15.0 wails build`

---

### wails doctor - 诊断检查

运行诊断程序检查系统是否准备就绪。

**示例：**

```bash
wails doctor
```

**输出示例：**
```
Wails CLI v2.0.0
Scanning system...

System
------
OS: Windows 10 Pro
Version: 2009 (Build: 19043)
ID: 21H1

Go Version: go1.18
Platform: windows
Architecture: amd64

Dependency       Package Name    Status      Version
----------       ------------    ------      -------
WebView2         N/A             Installed   93.0.961.52
npm              N/A             Installed   6.14.15
upx              N/A             Installed   upx 3.96 (*Optional)

Diagnosis
---------
Your system is ready for Wails development!
```

---

### wails generate - 生成器

#### 生成模板

```bash
wails generate template -name <模板名称> -frontend <前端路径>
```

| 标志 | 描述 |
|------|------|
| `-name` | 模板名称（必填） |
| `-frontend` | 前端项目路径 |

#### 生成模块

手动生成 wailsjs 目录。

```bash
wails generate module
```

| 标志 | 描述 | 默认值 |
|------|------|--------|
| `-compiler` | 指定 Go 编译器 | go |
| `-tags` | 构建标签 | - |

---

### wails update - 更新 CLI

```bash
# 更新到最新版本
wails update

# 更新到预发布版本
wails update -pre

# 更新到指定版本
wails update -version v2.10.0
```

---

### wails version - 查看版本

```bash
wails version
```

---

## 常用工作流

### 创建新项目

```bash
# 1. 创建项目
wails init -n myapp -g -ide vscode

# 2. 进入目录
cd myapp

# 3. 开发模式
wails dev
```

### 日常开发

```bash
# 启动开发服务器
wails dev -browser

# 如果使用 Vite/React/Vue 等框架
wails dev -frontenddevserverurl http://localhost:3000 -browser
```

### 构建发布

```bash
# 清理并构建
wails build -clean

# 交叉编译
wails build -platform windows/amd64,darwin/arm64

# 压缩并混淆
wails build -upx -obfuscated
```

---

---

## 运行时简介

### 导入和基本使用

**Go 端：**
```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// 所有方法都需要 context 作为第一个参数
// context 应该从应用启动回调或前端 Dom 加载完成回调中获取
```

**JavaScript 端：**
```javascript
// 运行时通过 window.runtime 提供
// 开发模式下会生成 TypeScript 声明，位于 wailsjs 目录
```

### Environment 获取环境信息

```go
// Go
info := runtime.Environment(ctx)
// 返回: EnvironmentInfo{BuildType, Platform, Arch}
```

```javascript
// JavaScript
const info = await runtime.Environment();
// { buildType: string, platform: string, arch: string }
```

---

## Events 事件

Wails 提供统一的事件系统，Go 和 JavaScript 都可以发出或接收事件。

### EventsOn - 添加事件监听器

```go
// Go
runtime.EventsOn(ctx, "eventName", func(optionalData ...interface{}) {
    // 处理事件
})

// 返回取消函数
cancel := runtime.EventsOn(ctx, "event", callback)
cancel() // 取消监听
```

```javascript
// JavaScript
const cancel = runtime.EventsOn("eventName", (optionalData) => {
    // 处理事件
});

cancel(); // 取消监听
```

### EventsOff - 移除事件监听器

```go
// Go
runtime.EventsOff(ctx, "eventName")
runtime.EventsOff(ctx, "eventName", "additionalEvent1", "additionalEvent2")
```

```javascript
// JavaScript
runtime.EventsOff("eventName");
runtime.EventsOff("eventName", "additionalEvent1");
```

### EventsOnce - 只触发一次的监听器

```go
// Go
cancel := runtime.EventsOnce(ctx, "eventName", callback)
```

```javascript
// JavaScript
const cancel = runtime.EventsOnce("eventName", callback);
```

### EventsOnMultiple - 限制触发次数

```go
// Go - 最多触发 5 次
cancel := runtime.EventsOnMultiple(ctx, "eventName", callback, 5)
```

```javascript
// JavaScript
const cancel = runtime.EventsOnMultiple("eventName", callback, 5);
```

### EventsEmit - 触发事件

```go
// Go
runtime.EventsEmit(ctx, "eventName")
runtime.EventsEmit(ctx, "eventName", data1, data2)
```

```javascript
// JavaScript
runtime.EventsEmit("eventName");
runtime.EventsEmit("eventName", data1, data2);
```

---

## Window 窗口

### 窗口标题

```go
// Go
runtime.WindowSetTitle(ctx, "新窗口标题")
```

```javascript
// JavaScript
runtime.WindowSetTitle("新窗口标题");
```

### 窗口全屏

```go
// Go
runtime.WindowFullscreen(ctx)
runtime.WindowUnfullscreen(ctx)
isFullscreen := runtime.WindowIsFullscreen(ctx)
```

```javascript
// JavaScript
runtime.WindowFullscreen();
runtime.WindowUnfullscreen();
const isFullscreen = await runtime.WindowIsFullscreen();
```

### 窗口位置和尺寸

```go
// Go
runtime.WindowCenter(ctx)              // 居中
runtime.SetSize(ctx, 800, 600)         // 设置尺寸
w, h := runtime.WindowGetSize(ctx)     // 获取尺寸
runtime.WindowSetPosition(ctx, 100, 200) // 设置位置
x, y := runtime.WindowGetPosition(ctx) // 获取位置
```

```javascript
// JavaScript
runtime.WindowCenter();
runtime.WindowSetSize(800, 600);
const size = await runtime.WindowGetSize(); // {w: number, h: number}
runtime.WindowSetPosition(100, 200);
const pos = await runtime.WindowGetPosition(); // {x: number, y: number}
```

### 窗口状态

```go
// Go
runtime.WindowMaximise(ctx)
runtime.WindowUnmaximise(ctx)
runtime.WindowToggleMaximise(ctx)
isMaximised := runtime.WindowIsMaximised(ctx)

runtime.WindowMinimise(ctx)
runtime.WindowUnminimise(ctx)
isMinimised := runtime.WindowIsMinimised(ctx)

isNormal := runtime.WindowIsNormal(ctx) // 是否正常状态（非最小化、最大化、全屏）
```

```javascript
// JavaScript
runtime.WindowMaximise();
runtime.WindowUnmaximise();
runtime.WindowToggleMaximise();
const isMaximised = await runtime.WindowIsMaximised();

runtime.WindowMinimise();
runtime.WindowUnminimise();
const isMinimised = await runtime.WindowIsMinimised();

const isNormal = await runtime.WindowIsNormal();
```

### 显示/隐藏窗口

```go
// Go
runtime.WindowShow(ctx)
runtime.WindowHide(ctx)
```

```javascript
// JavaScript
runtime.WindowShow();
runtime.WindowHide();
```

### 窗口主题（仅 Windows）

```go
// Go
runtime.WindowSetSystemDefaultTheme(ctx)
runtime.WindowSetLightTheme(ctx)
runtime.WindowSetDarkTheme(ctx)
```

```javascript
// JavaScript
runtime.WindowSetSystemDefaultTheme();
runtime.WindowSetLightTheme();
runtime.WindowSetDarkTheme();
```

### 窗口尺寸限制

```go
// Go
runtime.WindowSetMinSize(ctx, 400, 300)  // 最小尺寸
runtime.WindowSetMaxSize(ctx, 1920, 1080) // 最大尺寸
// 设置 0,0 禁用约束
```

```javascript
// JavaScript
runtime.WindowSetMinSize(400, 300);
runtime.WindowSetMaxSize(1920, 1080);
```

### 窗口置顶

```go
// Go
runtime.WindowSetAlwaysOnTop(ctx, true)
```

```javascript
// JavaScript
runtime.WindowSetAlwaysOnTop(true);
```

### 窗口背景色

```go
// Go
// RGBA 值 0-255，Windows 上 alpha 只支持 0 或 255
runtime.WindowSetBackgroundColour(ctx, 255, 255, 255, 255)
```

```javascript
// JavaScript
runtime.WindowSetBackgroundColour(255, 255, 255, 255);
```

### 其他窗口操作

```go
// Go
runtime.WindowReload(ctx)      // 重新加载页面
runtime.WindowReloadApp(ctx)   // 重新加载应用前端
runtime.WindowExecJS(ctx, "alert('hello')") // 执行 JS
runtime.WindowPrint(ctx)       // 打开打印对话框
```

```javascript
// JavaScript
runtime.WindowReload();
runtime.WindowReloadApp();
runtime.WindowPrint();
```

---

## Dialog 对话框

> 注意：JavaScript 运行时不支持对话框

### OpenDirectoryDialog - 选择目录

```go
result, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
    Title:              "选择目录",
    DefaultDirectory:   "/home/user",
    CanCreateDirectories: true,
})
// 返回：所选目录路径或空字符串（用户取消）
```

### OpenFileDialog - 选择文件

```go
result, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
    Title:            "选择文件",
    DefaultDirectory: "/home/user",
    DefaultFilename:  "test.txt",
    Filters: []runtime.FileFilter{
        {DisplayName: "图片文件", Pattern: "*.png;*.jpg;*.jpeg"},
        {DisplayName: "文本文件", Pattern: "*.txt"},
    },
    ShowHiddenFiles: true,
})
// 返回：所选文件路径或空字符串（用户取消）
```

### OpenMultipleFilesDialog - 选择多个文件

```go
files, err := runtime.OpenMultipleFilesDialog(ctx, runtime.OpenDialogOptions{
    Title:  "选择多个文件",
    Filters: []runtime.FileFilter{...},
})
// 返回：文件路径数组或 nil（用户取消）
```

### SaveFileDialog - 保存文件

```go
result, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
    Title:            "保存文件",
    DefaultDirectory: "/home/user",
    DefaultFilename:  "document.txt",
    Filters: []runtime.FileFilter{...},
    CanCreateDirectories: true,
})
// 返回：保存路径或空字符串（用户取消）
```

### MessageDialog - 消息对话框

```go
result, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
    Type:          runtime.QuestionDialog,
    Title:         "确认",
    Message:       "确定要继续吗？",
    Buttons:       []string{"是", "否", "取消"},
    DefaultButton: "否",  // 绑定回车键
    CancelButton:  "取消", // 绑定 ESC 键
})
// 返回：所选按钮的文本

// 对话框类型
const (
    InfoDialog     DialogType = "info"
    WarningDialog  DialogType = "warning"
    ErrorDialog    DialogType = "error"
    QuestionDialog DialogType = "question"
)
```

**平台差异：**

- **Windows**：标准对话框类型，按钮不可自定义，返回预定义值
- **Linux**：标准对话框类型，按钮不可定制
- **Mac**：最多 4 个自定义按钮

### FileFilter 结构

```go
type FileFilter struct {
    DisplayName string // "Image Files (*.jpg, *.png)"
    Pattern     string // "*.jpg;*.png"
}
```

---

## Browser 浏览器

### BrowserOpenURL - 打开 URL

```go
// Go
runtime.BrowserOpenURL(ctx, "https://www.example.com")
```

```javascript
// JavaScript
runtime.BrowserOpenURL("https://www.example.com");
```

---

## Log 日志

### 日志级别

- **Trace**（追踪）
- **Debug**（调试）
- **Info**（信息）
- **Warning**（警告）
- **Error**（错误）
- **Fatal**（致命）

### 基础日志

```go
// Go
runtime.LogPrint(ctx, "原始消息")
runtime.LogPrintf(ctx, "原始消息: %s", data)

runtime.LogTrace(ctx, "追踪消息")
runtime.LogTracef(ctx, "追踪: %s", data)

runtime.LogDebug(ctx, "调试消息")
runtime.LogDebugf(ctx, "调试: %s", data)

runtime.LogInfo(ctx, "信息消息")
runtime.LogInfof(ctx, "信息: %s", data)

runtime.LogWarning(ctx, "警告消息")
runtime.LogWarningf(ctx, "警告: %s", data)

runtime.LogError(ctx, "错误消息")
runtime.LogErrorf(ctx, "错误: %s", data)

runtime.LogFatal(ctx, "致命消息")
runtime.LogFatalf(ctx, "致命: %s", data)
```

```javascript
// JavaScript
runtime.LogPrint("原始消息");
runtime.LogTrace("追踪消息");
runtime.LogDebug("调试消息");
runtime.LogInfo("信息消息");
runtime.LogWarning("警告消息");
runtime.LogError("错误消息");
runtime.LogFatal("致命消息");
```

### 设置日志级别

```go
// Go
runtime.LogSetLogLevel(ctx, logger.InfoLevel)
```

```javascript
// JavaScript
// 级别对应: 1=Trace, 2=Debug, 3=Info, 4=Warning, 5=Error
runtime.LogSetLogLevel(3); // Info 级别
```

---

## Menu 菜单

> 注意：JavaScript 运行时不支持菜单

```go
// Go
import "github.com/wailsapp/wails/v2/pkg/menu"

// 创建菜单
appMenu := menu.NewMenu()
fileMenu := appMenu.AddSubmenu("文件")
fileMenu.AddText("新建", menu.Keys("cmd+n"), func(_ *menu.CallbackData) {
    runtime.WindowSetTitle(ctx, "新建文件")
})
fileMenu.AddText("打开", menu.Keys("cmd+o"), func(_ *menu.CallbackData) {
    // 打开文件
})

// 设置应用菜单
runtime.MenuSetApplicationMenu(ctx, appMenu)

// 更新应用菜单（获取任意更改）
runtime.MenuUpdateApplicationMenu(ctx)
```

---

## Clipboard 剪贴板

> 当前实现仅处理文本

### ClipboardGetText - 获取文本

```go
// Go
text, err := runtime.ClipboardGetText(ctx)
// 返回：剪贴板文本或空字符串（剪贴板为空）
```

```javascript
// JavaScript
const text = await runtime.ClipboardGetText();
// 返回：剪贴板文本或空字符串
```

### ClipboardSetText - 设置文本

```go
// Go
err := runtime.ClipboardSetText(ctx, "要复制的文本")
```

```javascript
// JavaScript
const success = await runtime.ClipboardSetText("要复制的文本");
// 返回：true（成功）或 false（失败）
```

---

## TypeScript 类型定义

### EnvironmentInfo
```typescript
interface EnvironmentInfo {
  buildType: string;
  platform: string;
  arch: string;
}
```

### Position
```typescript
interface Position {
  x: number;
  y: number;
}
```

### Size
```typescript
interface Size {
  w: number;
  h: number;
}
```

---

## 重要提示

1. **Context 使用**：Go 端所有 runtime 方法都需要 context，应从应用启动回调或前端 Dom 加载完成回调中获取
2. **JavaScript 限制**：Dialog 和 Menu 功能在 JavaScript 端不支持
3. **平台差异**：部分功能在不同平台上有差异（如窗口主题仅支持 Windows）
4. **异步操作**：JavaScript 中的大多数方法返回 Promise

## 参考链接

- Wails 官方文档：https://wails.io/zh-Hans/docs/reference/runtime/intro
- GitHub：https://github.com/wailsapp/wails
