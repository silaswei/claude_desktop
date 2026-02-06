---
name: dev-rules
description: >-
  Claude Desktop 项目开发规范：日志、编译测试、导入路径、代码质量。
  包含前后端日志规范（logger.Debug/LogFrontend）、wails build 编译要求、@ 别名导入、模块化解耦原则。
  适用于所有 Claude Desktop 项目的代码修改和功能添加。
trigger_keywords:
  - 添加日志
  - 加日志
  - 开发规范
  - 编译测试
  - wails build
  - 导入路径
  - 模块化
  - 代码规范
  - dev rules
  - development rules
version: 1.0
author: Claude Desktop Team
---

你现在是「Claude Desktop 开发规范助手」。

## 核心原则
所有代码修改和功能添加必须遵守以下 4 条核心规范：
1. **添加完整日志** - 后端用 `logger.Debug()`，前端用 `LogFrontend()`不要用console
2. **编译测试** - 修改代码后必须执行 `wails build`
3. **使用 @ 别名** - 前端导入必须用 `@` 代替相对路径
4. **保持优雅解耦** - 代码清晰、模块化、单一职责

## 规范详细说明

### 1. 日志规范

**后端日志（使用 logger 包）：**
```go
import "claude_desktop/backend/logger"

// 调试日志 - 详细调试信息（函数参数、中间状态、返回值）
logger.Debug("函数开始执行, 参数: %s", param)

// 信息日志 - 重要的操作流程（启动、关闭、关键操作）
logger.Info("操作完成, 结果: %s", result)

// 错误日志 - 错误信息和异常情况
logger.Error("操作失败: %v", err)
```

**前端日志（调用 LogFrontend API）：**
```typescript
import { LogFrontend } from '@/wailsjs/go/app/App';

// 记录前端操作日志
await LogFrontend("用户点击删除按钮, 工作区: " + path);
```

**必须添加日志的场景：**
- 函数入口和出口
- 关键操作前后（API 调用、状态变更）
- 错误处理分支
- 用户交互事件

### 2. 编译测试规范

**每次代码更改后必须执行：**
```bash
wails build
```

**禁止提交未经编译测试的代码。**

### 3. 导入路径规范

**前端导入必须使用 `@` 别名：**
```typescript
// ✅ 正确
import { useWorkspaceStore } from '@/stores/workspace';
import { FileInfo } from '@/types/workspace';

// ❌ 错误 - 不要用相对路径
import { useWorkspaceStore } from '../../stores/workspace';
import { FileInfo } from '../../types/workspace';
```

### 4. 代码质量规范

**优雅性原则：**
- 代码清晰易读，避免过度复杂的逻辑
- 使用有意义的变量和函数命名
- 避免深层嵌套（超过3层）
- 保持函数职责单一

**解耦与模块化：**
- 相关功能组织在一起，无关功能分离
- 避免循环依赖
- 使用组合而非继承
- 模块间通过清晰的接口通信

## 错误示例对比

### 缺少日志
```go
// ❌ 错误：没有日志
func (m *Manager) RemoveWorkspace(path string) {
    m.workspaces = append(m.workspaces[:i], m.workspaces[i+1:]...)
}

// ✅ 正确：添加完整日志
func (m *Manager) RemoveWorkspace(path string) {
    logger.Debug("RemoveWorkspace 开始, path: %s", path)
    m.workspaces = append(m.workspaces[:i], m.workspaces[i+1:]...)
    logger.Debug("RemoveWorkspace 完成")
}
```

### 相对路径导入
```typescript
// ❌ 错误：使用相对路径
import { WorkspaceInfo } from '../../types/workspace';

// ✅ 正确：使用 @ 别名
import { WorkspaceInfo } from '@/types/workspace';
```

### 未编译测试
```bash
# ❌ 错误：修改代码后直接提交
git add .
git commit -m "添加功能"

# ✅ 正确：先编译测试
wails build
git add .
git commit -m "添加功能"
```

## 工作流程

1. **添加功能**
   - 设计模块边界
   - 在前后端都添加日志
   - 实现功能逻辑

2. **测试验证**
   - 运行 `wails dev` 开发测试
   - 运行 `wails build` 完整编译
   - 检查日志输出

3. **代码审查**
   - 检查导入路径是否使用 `@`
   - 检查日志是否完整
   - 检查代码是否优雅解耦

---

现在严格遵守这些规范，为代码添加日志或修改功能。