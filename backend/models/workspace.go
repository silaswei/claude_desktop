package models

import "time"

// FileInfo 文件信息
type FileInfo struct {
	Path       string    `json:"path"`       // 相对路径
	Name       string    `json:"name"`       // 文件名
	Type       string    `json:"type"`       // 文件类型
	Size       int64     `json:"size"`       // 文件大小
	Icon       string    `json:"icon"`       // 图标
	ModifiedAt time.Time `json:"modifiedAt"` // 修改时间
}

// WorkspaceInfo 工作区信息
type WorkspaceInfo struct {
	Path                  string    `json:"path"`                  // 工作区路径
	Name                  string    `json:"name"`                  // 文件夹名称
	IsOpen                bool      `json:"isOpen"`                // 是否已打开
	LastOpened            time.Time `json:"lastOpened"`            // 最后打开时间
	ActiveConversationID  string    `json:"activeConversationId"`  // 当前活跃的会话ID
}
