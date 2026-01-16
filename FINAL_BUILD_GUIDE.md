# 最终构建指南

## 已完成的修改

### 1. ✅ 实现系统文件选择器
- 在 `app.go` 中添加了 `DialogOpenDirectory()` 函数
- 在 `MainView.vue` 中更新了 `handleOpenFolder()` 使用系统文件对话框
- 导入了 `DialogOpenDirectory` API

### 2. ✅ 启用真实环境检测
- 更新 `env.ts` 使用真实的 Wails API 调用（移除 mock 数据）
- 更新 `workspace.ts` 使用真实的 Wails API 调用
- 更新 `conversation.ts` 使用真实的 Wails API 调用

### 3. ✅ 修复 TypeScript 类型错误
- 将前端类型从 `workspacePath` 改为 `projectPath` 以匹配后端
- 更新了所有相关的计算属性和函数

## 构建步骤

请按以下顺序执行：

### 方案 1：使用现有的完整清理脚本（推荐）

```bash
cd /Users/Apple/GolandProjects/OpenSource/claude_desktop
chmod +x full_clean_and_build.sh
./full_clean_and_build.sh
```

这个脚本会自动完成以下步骤：
1. 清理 root 权限的文件
2. 修复前端目录权限
3. 删除旧的 store 文件
4. 清理构建缓存
5. 执行编译

### 方案 2：手动执行

```bash
# 1. 进入项目目录
cd /Users/Apple/GolandProjects/OpenSource/claude_desktop

# 2. 清理 root 权限的文件
sudo rm -rf frontend/wailsjs frontend/dist frontend/node_modules/.vite

# 3. 修复前端目录权限
sudo chown -R $USER:$USER frontend/

# 4. 删除旧的 project store 文件（如果存在）
rm -f frontend/src/stores/project.ts

# 5. 设置自定义临时目录并编译
export TMPDIR=$(pwd)/.tmp
mkdir -p .tmp
wails build
```

## 预期结果

编译成功后，应用程序将位于：
```
build/bin/claude-terminal.app
```

## 功能验证

启动应用后，你应该能使用以下功能：

1. **环境检测**
   - 自动检测 Node.js、npm、Claude CLI、网络、Git
   - 显示真实的检测结果（不是 mock 数据）

2. **打开工作区**
   - 点击"打开文件夹"按钮
   - 系统文件选择对话框会弹出
   - 选择任意文件夹作为工作区
   - 文件列表会显示该文件夹下的所有文件

3. **对话管理**
   - 创建新对话
   - 发送消息（会调用真实的 Claude API）
   - 查看对话历史

## 常见问题

### Q: 仍然出现权限错误
A: 确保使用 `sudo` 执行清理命令，并确保 `$USER` 变量正确指向你的用户名（Apple）

### Q: TypeScript 编译错误
A: 所有类型错误已修复，确保 `frontend/wailsjs` 目录已重新生成

### Q: 找不到 DialogOpenDirectory
A: 需要重新生成 Wails 绑定，确保 `wails build` 完整运行

## 技术变更总结

### 后端变更
- `app.go`: 新增 `DialogOpenDirectory()` 函数

### 前端变更
- `MainView.vue`: 使用系统文件对话框
- `env.ts`: 移除所有 mock 数据，使用真实 API
- `workspace.ts`: 移除所有 mock 数据，使用真实 API
- `conversation.ts`: 移除所有 mock 数据，使用真实 API
- `conversation.ts` 类型: `workspacePath` → `projectPath`
