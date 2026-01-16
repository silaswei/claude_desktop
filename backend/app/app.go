package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"claude_desktop/backend/detector"
	"claude_desktop/backend/manager/conversation"
	"claude_desktop/backend/manager/workspace"
	"claude_desktop/backend/models"
	"claude_desktop/backend/service"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	envManager       *detector.Manager
	envConfig        *models.EnvironmentConfig
	workspaceManager *workspace.Manager
	convManager      *service.ConversationManager
	storage          conversation.Storage
}

// NewApp creates a new App application struct
// NewApp 创建一个新的 App 应用程序
func NewApp() *App {
	// 创建环境配置
	envConfig := models.DefaultEnvironmentConfig()

	// 创建环境检测管理器
	envManager := detector.NewManager(envConfig)

	// 获取用户主目录
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".claude-terminal")

	// 创建存储服务
	storage, _ := conversation.NewJSONStorage(dataDir)

	// 创建工作区管理器
	workspaceManager := workspace.NewManager()

	// 创建对话管理器
	convManager := service.NewConversationManager(storage)

	return &App{
		envConfig:        envConfig,
		envManager:       envManager,
		workspaceManager: workspaceManager,
		convManager:      convManager,
		storage:          storage,
	}
}

// Startup is called at application startup
// Startup 在应用程序启动时调用
func (a *App) Startup(ctx context.Context) {
	// Perform your setup here
	// 在这里执行初始化设置
	a.ctx = ctx

	// 调整窗口大小为屏幕的 3/4
	a.resizeWindowToThreeQuarters()
}

// resizeWindowToThreeQuarters 调整窗口大小为屏幕的 3/4
func (a *App) resizeWindowToThreeQuarters() {
	// 获取所有屏幕信息
	screens, err := runtime.ScreenGetAll(a.ctx)

	if err != nil || len(screens) == 0 {
		// 如果无法获取屏幕信息，使用默认值
		runtime.WindowSetSize(a.ctx, 1200, 800)
		runtime.WindowCenter(a.ctx)
		return
	}

	// 使用主屏幕（第一个屏幕）
	primaryScreen := screens[0]
	screenWidth := primaryScreen.Width
	screenHeight := primaryScreen.Height

	// 计算屏幕的 3/4
	targetWidth := int(float64(screenWidth) * 0.75)
	targetHeight := int(float64(screenHeight) * 0.75)

	// 设置窗口大小
	runtime.WindowSetSize(a.ctx, targetWidth, targetHeight)

	// 将窗口居中
	runtime.WindowCenter(a.ctx)
}

// DomReady is called after the front-end dom has been loaded
// DomReady 在前端Dom加载完毕后调用
func (a *App) DomReady(ctx context.Context) {
	// Add your action here
	// 在这里添加你的操作
}

// BeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue,
// false will continue shutdown as normal.
// BeforeClose在单击窗口关闭按钮或调用runtime.Quit即将退出应用程序时被调用.
// 返回 true 将导致应用程序继续，false 将继续正常关闭。
func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	return false
}

// Shutdown is called at application termination
// 在应用程序终止时被调用
func (a *App) Shutdown(ctx context.Context) {
	// Perform your teardown here
	// 在此处做一些资源释放的操作
}

// ==================== 环境检测相关 API ====================

// EnvDetectAll 执行所有环境检测
func (a *App) EnvDetectAll() (*models.EnvironmentInfo, error) {
	return a.envManager.DetectAll(a.ctx)
}

// EnvDetectByName 执行指定名称的检测
func (a *App) EnvDetectByName(name string) (*models.DetectionResult, error) {
	return a.envManager.DetectByName(a.ctx, name)
}

// EnvGetStatus 获取环境状态摘要
func (a *App) EnvGetStatus() (*models.EnvironmentInfo, error) {
	return a.envManager.GetStatus(a.ctx)
}

// EnvClearCache 清除环境检测缓存
func (a *App) EnvClearCache() error {
	return a.envManager.ClearCache()
}

// EnvGetDetectorNames 获取所有检测器名称
func (a *App) EnvGetDetectorNames() []string {
	return a.envManager.GetAllDetectors()
}

// ==================== 工作区管理相关 API ====================

// DialogOpenDirectory 打开系统文件夹选择对话框
func (a *App) DialogOpenDirectory() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择工作区文件夹",
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

// WorkspaceOpen 打开工作区（如果不存在则创建新的）
func (a *App) WorkspaceOpen(path string) (*models.WorkspaceInfo, error) {
	ws, err := a.workspaceManager.Open(path)
	if err != nil {
		return nil, err
	}

	return &models.WorkspaceInfo{
		Path:       ws.Path,
		Name:       ws.Name,
		IsOpen:     true,
		LastOpened: ws.LastOpened,
	}, nil
}

// WorkspaceClose 关闭工作区
func (a *App) WorkspaceClose() {
	a.workspaceManager.Close()
}

// WorkspaceGetCurrent 获取当前工作区路径
func (a *App) WorkspaceGetCurrent() string {
	return a.workspaceManager.GetCurrent()
}

// WorkspaceIsOpen 检查是否已打开工作区
func (a *App) WorkspaceIsOpen() bool {
	return a.workspaceManager.IsOpen()
}

// WorkspaceListFiles 获取工作区文件列表
func (a *App) WorkspaceListFiles() ([]*models.FileInfo, error) {
	return a.workspaceManager.ListFiles(a.ctx)
}

// WorkspaceReadFile 读取文件内容
func (a *App) WorkspaceReadFile(relativePath string) (string, error) {
	return a.workspaceManager.ReadFile(relativePath)
}

// WorkspaceWriteFile 写入文件
func (a *App) WorkspaceWriteFile(relativePath, content string) error {
	return a.workspaceManager.WriteFile(relativePath, content)
}

// WorkspaceDeleteFile 删除文件
func (a *App) WorkspaceDeleteFile(relativePath string) error {
	return a.workspaceManager.DeleteFile(relativePath)
}

// WorkspaceCreateFile 创建新文件
func (a *App) WorkspaceCreateFile(relativePath, content string) error {
	return a.workspaceManager.CreateFile(relativePath, content)
}

// WorkspaceGetInfo 获取工作区信息
func (a *App) WorkspaceGetInfo() *models.WorkspaceInfo {
	return a.workspaceManager.GetWorkspaceInfo()
}

// WorkspaceList 获取所有工作区列表
func (a *App) WorkspaceList() []*models.WorkspaceInfo {
	workspaces := a.workspaceManager.GetWorkspaces()
	result := make([]*models.WorkspaceInfo, len(workspaces))
	for i, ws := range workspaces {
		result[i] = &models.WorkspaceInfo{
			Path:       ws.Path,
			Name:       ws.Name,
			IsOpen:     ws.Path == a.workspaceManager.GetCurrent(),
			LastOpened: ws.LastOpened,
		}
	}
	return result
}

// WorkspaceSelect 选择工作区
func (a *App) WorkspaceSelect(path string) error {
	return a.workspaceManager.SelectWorkspace(path)
}

// WorkspaceRemove 移除工作区
func (a *App) WorkspaceRemove(path string) {
	a.workspaceManager.RemoveWorkspace(path)
}

// WorkspaceRenameFile 重命名文件或目录
func (a *App) WorkspaceRenameFile(oldPath, newPath string) error {
	return a.workspaceManager.RenameFile(oldPath, newPath)
}

// WorkspaceCopyFile 复制文件或目录
func (a *App) WorkspaceCopyFile(srcPath, destPath string) error {
	return a.workspaceManager.CopyFile(srcPath, destPath)
}

// WorkspaceMoveFile 移动文件或目录
func (a *App) WorkspaceMoveFile(srcPath, destPath string) error {
	return a.workspaceManager.MoveFile(srcPath, destPath)
}

// WorkspaceCreateDirectory 创建目录
func (a *App) WorkspaceCreateDirectory(relativePath string) error {
	return a.workspaceManager.CreateDirectory(relativePath)
}

// WorkspaceGetFullPath 获取文件完整路径
func (a *App) WorkspaceGetFullPath(relativePath string) (string, error) {
	return a.workspaceManager.GetFullPath(relativePath)
}

// ==================== 系统操作相关 API ====================

// SystemOpenFile 打开文件（使用系统默认应用）
func (a *App) SystemOpenFile(relativePath string) error {
	fullPath, err := a.workspaceManager.GetFullPath(relativePath)
	if err != nil {
		return err
	}

	// 使用 macOS 的 open 命令
	cmd := exec.Command("open", fullPath)
	return cmd.Run()
}

// SystemOpenTerminal 在终端中打开指定路径
func (a *App) SystemOpenTerminal(relativePath string) error {
	fullPath, err := a.workspaceManager.GetFullPath(relativePath)
	if err != nil {
		return err
	}

	// 获取文件信息
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	// 如果是文件，获取其目录
	dirPath := fullPath
	if !info.IsDir() {
		dirPath = filepath.Dir(fullPath)
	}

	// 使用 macOS 的 open 命令打开 Terminal
	cmd := exec.Command("open", "-a", "Terminal", dirPath)
	return cmd.Run()
}

// SystemRevealInFinder 在Finder中显示文件
func (a *App) SystemRevealInFinder(relativePath string) error {
	fullPath, err := a.workspaceManager.GetFullPath(relativePath)
	if err != nil {
		return err
	}

	// 使用 macOS 的 open -R 命令在 Finder 中显示
	cmd := exec.Command("open", "-R", fullPath)
	return cmd.Run()
}

// ==================== 对话管理相关 API ====================

// ConversationCreate 创建新对话
func (a *App) ConversationCreate(title, projectPath string) (*conversation.Conversation, error) {
	conv, err := a.convManager.CreateConversation(title, projectPath)
	if err != nil {
		fmt.Printf("创建会话失败: %v\n", err)
		return nil, err
	}
	fmt.Printf("创建会话成功: ID=%s, Title=%s, Path=%s\n", conv.ID, conv.Title, conv.ProjectPath)
	return conv, nil
}

// ConversationDelete 删除对话
func (a *App) ConversationDelete(id string) error {
	return a.convManager.DeleteConversation(id)
}

// ConversationInfo 获取对话详情
func (a *App) ConversationInfo(id string) (*conversation.Conversation, error) {
	return a.convManager.GetConversation(id)
}

// ConversationList 获取对话列表
func (a *App) ConversationList() ([]*conversation.Conversation, error) {
	return a.convManager.ListConversations()
}

// ConversationUpdate 更新对话信息
func (a *App) ConversationUpdate(conv *conversation.Conversation) error {
	return a.convManager.UpdateConversation(conv)
}

// ConversationGetByProjectPath 根据项目路径获取最近的对话
func (a *App) ConversationGetByProjectPath(projectPath string) (*conversation.Conversation, error) {
	return a.convManager.GetConversationByProjectPath(projectPath)
}

// ConversationSend 发送消息
func (a *App) ConversationSend(convID, content string) (*conversation.Conversation, error) {
	return a.convManager.SendMessage(convID, content)
}

// ConversationSendWithCallback 发送消息并提供流式回调
func (a *App) ConversationSendWithCallback(convID, content string, onChunk func(string)) (*conversation.Conversation, error) {
	return a.convManager.SendMessageWithCallback(convID, content, onChunk)
}

// ConversationSendWithEvents 发送消息并通过 Wails Events 推送响应
func (a *App) ConversationSendWithEvents(convID, content string) error {
	fmt.Printf("发送消息到会话 %s: %s\n", convID, content)

	// 发送思考开始事件
	runtime.EventsEmit(a.ctx, "claude:thinking", map[string]interface{}{
		"convID": convID,
	})

	// 使用 SendMessageWithCallback 并在回调中发送事件
	_, err := a.convManager.SendMessageWithCallback(convID, content, func(chunk string) {
		// 通过 Wails Events 发送响应片段到前端
		runtime.EventsEmit(a.ctx, "claude:response", map[string]interface{}{
			"content": chunk,
			"convID":  convID,
		})
	})

	// 发送完成事件
	runtime.EventsEmit(a.ctx, "claude:complete", map[string]interface{}{
		"convID": convID,
	})

	if err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
	} else {
		fmt.Printf("发送消息成功\n")
	}

	return err
}
