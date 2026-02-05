package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"claude_desktop/backend/models"
)

// Workspace 工作区
type Workspace struct {
	Path                 string
	Name                 string
	LastOpened           time.Time
	ActiveConversationID string
}

// Manager 工作区管理器
type Manager struct {
	mu          sync.RWMutex
	workspaces  []*Workspace // 所有工作区列表
	currentPath string       // 当前选中的工作区路径
	storageFile string       // 持久化文件路径
}

// NewManager 创建工作区管理器
func NewManager() *Manager {
	// 获取用户主目录
	homeDir, _ := os.UserHomeDir()
	storageDir := filepath.Join(homeDir, ".claude-desktop")

	// 确保目录存在
	os.MkdirAll(storageDir, 0755)

	storageFile := filepath.Join(storageDir, "workspaces.json")

	m := &Manager{
		workspaces:  make([]*Workspace, 0),
		storageFile: storageFile,
	}

	// 加载持久化的工作区数据
	m.loadFromStorage()

	return m
}

// loadFromStorage 从文件加载工作区数据
func (m *Manager) loadFromStorage() {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.storageFile)
	if err != nil {
		// 文件不存在，使用空列表
		return
	}

	var storageList []struct {
		Path                 string    `json:"path"`
		Name                 string    `json:"name"`
		LastOpened           time.Time `json:"lastOpened"`
		ActiveConversationID string    `json:"activeConversationId"`
	}

	if err := json.Unmarshal(data, &storageList); err != nil {
		fmt.Printf("加载工作区数据失败: %v\n", err)
		return
	}

	// 转换为 Workspace 对象
	m.workspaces = make([]*Workspace, 0, len(storageList))
	for _, item := range storageList {
		// 检查路径是否仍然存在
		if _, err := os.Stat(item.Path); err == nil {
			m.workspaces = append(m.workspaces, &Workspace{
				Path:                 item.Path,
				Name:                 item.Name,
				LastOpened:           item.LastOpened,
				ActiveConversationID: item.ActiveConversationID,
			})
		}
	}
}

// saveToStorage 保存工作区数据到文件
func (m *Manager) saveToStorage() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	storageList := make([]struct {
		Path                 string    `json:"path"`
		Name                 string    `json:"name"`
		LastOpened           time.Time `json:"lastOpened"`
		ActiveConversationID string    `json:"activeConversationId"`
	}, len(m.workspaces))

	for i, ws := range m.workspaces {
		storageList[i] = struct {
			Path                 string    `json:"path"`
			Name                 string    `json:"name"`
			LastOpened           time.Time `json:"lastOpened"`
			ActiveConversationID string    `json:"activeConversationId"`
		}{
			Path:                 ws.Path,
			Name:                 ws.Name,
			LastOpened:           ws.LastOpened,
			ActiveConversationID: ws.ActiveConversationID,
		}
	}

	data, err := json.MarshalIndent(storageList, "", "  ")
	if err != nil {
		fmt.Printf("序列化工作区数据失败: %v\n", err)
		return
	}

	if err := os.WriteFile(m.storageFile, data, 0644); err != nil {
		fmt.Printf("保存工作区数据失败: %v\n", err)
	}
}

// Open 打开工作区（如果不存在则创建新的）
func (m *Manager) Open(path string) (*Workspace, error) {
	m.mu.Lock()

	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		m.mu.Unlock()
		return nil, err
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		m.mu.Unlock()
		return nil, err
	}

	// 检查是否已在列表中
	for _, ws := range m.workspaces {
		if ws.Path == absPath {
			ws.LastOpened = time.Now()
			m.currentPath = absPath
			m.mu.Unlock()
			// 异步保存，避免阻塞
			go m.saveToStorage()
			return ws, nil
		}
	}

	// 创建新工作区
	name := filepath.Base(absPath)
	workspace := &Workspace{
		Path:       absPath,
		Name:       name,
		LastOpened: time.Now(),
	}

	m.workspaces = append(m.workspaces, workspace)
	m.currentPath = absPath
	m.mu.Unlock()

	// 异步保存，避免阻塞
	go m.saveToStorage()

	return workspace, nil
}

// GetCurrent 获取当前工作区路径
func (m *Manager) GetCurrent() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentPath
}

// IsOpen 检查是否已打开工作区
func (m *Manager) IsOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentPath != ""
}

// Close 关闭当前工作区
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentPath = ""
}

// SelectWorkspace 选择工作区
func (m *Manager) SelectWorkspace(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查找工作区
	for _, ws := range m.workspaces {
		if ws.Path == path {
			m.currentPath = path
			return nil
		}
	}

	return os.ErrNotExist
}

// RemoveWorkspace 移除工作区
func (m *Manager) RemoveWorkspace(path string) {
	m.mu.Lock()

	for i, ws := range m.workspaces {
		if ws.Path == path {
			// 从列表中移除
			m.workspaces = append(m.workspaces[:i], m.workspaces[i+1:]...)

			// 如果移除的是当前工作区，清空当前路径
			if m.currentPath == path {
				m.currentPath = ""
			}
			m.mu.Unlock()

			// 异步保存，避免阻塞
			go m.saveToStorage()
			return
		}
	}

	m.mu.Unlock()
}

// GetWorkspaces 获取所有工作区列表（按最后打开时间排序）
func (m *Manager) GetWorkspaces() []*Workspace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 复制列表
	result := make([]*Workspace, len(m.workspaces))
	copy(result, m.workspaces)

	// 按最后打开时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastOpened.After(result[j].LastOpened)
	})

	return result
}

// GetCurrentWorkspace 获取当前工作区
func (m *Manager) GetCurrentWorkspace() *Workspace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentPath == "" {
		return nil
	}

	for _, ws := range m.workspaces {
		if ws.Path == m.currentPath {
			return ws
		}
	}

	return nil
}

// ListFiles 获取工作区文件列表
func (m *Manager) ListFiles(ctx context.Context) ([]*models.FileInfo, error) {
	m.mu.RLock()
	path := m.currentPath
	m.mu.RUnlock()

	if path == "" {
		return nil, nil
	}

	return m.scanDir(ctx, path, "")
}

// scanDir 扫描目录（递归扫描所有文件和文件夹）
func (m *Manager) scanDir(ctx context.Context, rootPath, relativePath string) ([]*models.FileInfo, error) {
	fullPath := filepath.Join(rootPath, relativePath)

	// 读取目录
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []*models.FileInfo

	for _, entry := range entries {
		// 检查上下文
		select {
		case <-ctx.Done():
			return files, ctx.Err()
		default:
		}

		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		relPath := filepath.Join(relativePath, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := &models.FileInfo{
			Path:       relPath,
			Name:       entry.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		}

		if entry.IsDir() {
			fileInfo.Type = "directory"
			fileInfo.Icon = "📁"

			// 递归扫描子目录
			subFiles, err := m.scanDir(ctx, rootPath, relPath)
			if err == nil {
				files = append(files, subFiles...)
			}
		} else {
			fileInfo.Type = getFileType(entry.Name())
			fileInfo.Icon = getFileIcon(fileInfo.Type)
		}

		files = append(files, fileInfo)
	}

	// 排序：目录在前，然后按名称排序
	sort.Slice(files, func(i, j int) bool {
		// 先按路径深度排序，深度小的在前
		depthI := strings.Count(files[i].Path, string(filepath.Separator))
		depthJ := strings.Count(files[j].Path, string(filepath.Separator))

		if depthI != depthJ {
			return depthI < depthJ
		}

		// 同一深度，目录在前
		if files[i].Type == "directory" && files[j].Type != "directory" {
			return true
		}
		if files[i].Type != "directory" && files[j].Type == "directory" {
			return false
		}

		// 同一类型，按名称排序
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	return files, nil
}

// ReadFile 读取文件内容
func (m *Manager) ReadFile(relativePath string) (string, error) {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return "", os.ErrNotExist
	}

	fullPath := filepath.Join(basePath, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// WriteFile 写入文件
func (m *Manager) WriteFile(relativePath, content string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	fullPath := filepath.Join(basePath, relativePath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// DeleteFile 删除文件或目录
func (m *Manager) DeleteFile(relativePath string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	fullPath := filepath.Join(basePath, relativePath)

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", relativePath)
	}
	if err != nil {
		return err
	}

	// 如果是目录，递归删除
	if info.IsDir() {
		return os.RemoveAll(fullPath)
	}

	// 删除文件
	return os.Remove(fullPath)
}

// CreateFile 创建新文件
func (m *Manager) CreateFile(relativePath, content string) error {
	return m.WriteFile(relativePath, content)
}

// CreateDirectory 创建目录
func (m *Manager) CreateDirectory(relativePath string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	fullPath := filepath.Join(basePath, relativePath)
	return os.MkdirAll(fullPath, 0755)
}

// RenameFile 重命名文件或目录
func (m *Manager) RenameFile(oldPath, newPath string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	oldFullPath := filepath.Join(basePath, oldPath)
	newFullPath := filepath.Join(basePath, newPath)

	// 检查源文件是否存在
	if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", oldPath)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(newFullPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s", newPath)
	}

	// 直接重命名，不创建新目录
	return os.Rename(oldFullPath, newFullPath)
}

// CopyFile 复制文件或目录
func (m *Manager) CopyFile(srcPath, destPath string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	srcFullPath := filepath.Join(basePath, srcPath)
	destFullPath := filepath.Join(basePath, destPath)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return err
	}

	// 获取源文件信息
	srcInfo, err := os.Stat(srcFullPath)
	if err != nil {
		return err
	}

	// 如果是目录，递归复制
	if srcInfo.IsDir() {
		return m.copyDirectory(srcFullPath, destFullPath)
	}

	// 复制文件
	return m.copyFile(srcFullPath, destFullPath)
}

// copyFile 复制单个文件
func (m *Manager) copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dest, srcInfo.Mode())
}

// copyDirectory 递归复制目录
func (m *Manager) copyDirectory(src, dest string) error {
	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// 读取源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 递归复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := m.copyDirectory(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := m.copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// MoveFile 移动文件或目录
func (m *Manager) MoveFile(srcPath, destPath string) error {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return os.ErrNotExist
	}

	srcFullPath := filepath.Join(basePath, srcPath)
	destFullPath := filepath.Join(basePath, destPath)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return err
	}

	return os.Rename(srcFullPath, destFullPath)
}

// GetFullPath 获取文件的完整路径
func (m *Manager) GetFullPath(relativePath string) (string, error) {
	m.mu.RLock()
	basePath := m.currentPath
	m.mu.RUnlock()

	if basePath == "" {
		return "", os.ErrNotExist
	}

	return filepath.Join(basePath, relativePath), nil
}

// getFileType 根据文件扩展名获取文件类型
func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp":
		return "image"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx":
		return "word"
	case ".xls", ".xlsx", ".csv":
		return "excel"
	case ".ppt", ".pptx":
		return "powerpoint"
	case ".txt", ".md", ".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg":
		return "text"
	case ".js", ".ts", ".jsx", ".tsx", ".vue", ".html", ".css", ".scss", ".less":
		return "code-web"
	case ".py", ".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".php", ".rb", ".swift", ".kt":
		return "code"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	case ".mp3", ".wav", ".flac", ".ogg":
		return "audio"
	case ".mp4", ".avi", "mkv", "mov", "wmv":
		return "video"
	default:
		return "unknown"
	}
}

// getFileIcon 获取文件图标
func getFileIcon(fileType string) string {
	switch fileType {
	case "directory":
		return "📁"
	case "image":
		return "🖼️"
	case "pdf":
		return "📕"
	case "word":
		return "📘"
	case "excel":
		return "📊"
	case "powerpoint":
		return "📙"
	case "text", "code", "code-web":
		return "📄"
	case "archive":
		return "🗜️"
	case "audio":
		return "🎵"
	case "video":
		return "🎬"
	default:
		return "📄"
	}
}

// GetWorkspaceInfo 获取工作区信息
func (m *Manager) GetWorkspaceInfo() *models.WorkspaceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ws := m.GetCurrentWorkspace()
	if ws == nil {
		return nil
	}

	return &models.WorkspaceInfo{
		Path:                 ws.Path,
		Name:                 ws.Name,
		IsOpen:               true,
		LastOpened:           ws.LastOpened,
		ActiveConversationID: ws.ActiveConversationID,
	}
}

// SetActiveConversationID 设置当前工作区的活跃会话ID
func (m *Manager) SetActiveConversationID(convID string) error {
	m.mu.Lock()

	if m.currentPath == "" {
		m.mu.Unlock()
		return fmt.Errorf("没有打开的工作区")
	}

	for _, ws := range m.workspaces {
		if ws.Path == m.currentPath {
			ws.ActiveConversationID = convID
			m.mu.Unlock()

			// 异步保存，避免阻塞
			go m.saveToStorage()
			return nil
		}
	}

	m.mu.Unlock()
	return fmt.Errorf("当前工作区不在列表中")
}

// GetActiveConversationID 获取当前工作区的活跃会话ID
func (m *Manager) GetActiveConversationID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentPath == "" {
		return ""
	}

	for _, ws := range m.workspaces {
		if ws.Path == m.currentPath {
			return ws.ActiveConversationID
		}
	}

	return ""
}
