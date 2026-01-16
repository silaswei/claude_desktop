package conversation

import (
	"time"
)

// Conversation 对话实体
type Conversation struct {
	ID          string    `json:"id"`          // 对话 ID
	Title       string    `json:"title"`       // 对话标题
	ProjectPath string    `json:"projectPath"` // 关联项目路径（可为空）
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间
	Messages    []Message `json:"messages"`    // 消息列表
}

// NewConversation 创建新对话
func NewConversation(title, projectPath string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:          generateID(),
		Title:       title,
		ProjectPath: projectPath,
		CreatedAt:   now,
		UpdatedAt:   now,
		Messages:    make([]Message, 0),
	}
}

// AddMessage 添加消息
func (c *Conversation) AddMessage(msg Message) {
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// GetLastMessage 获取最后一条消息
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return &c.Messages[len(c.Messages)-1]
}

// generateID 生成唯一 ID
func generateID() string {
	return "conv-" + time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
