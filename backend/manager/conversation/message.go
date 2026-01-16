package conversation

import (
	"time"
)

// Message 消息实体
type Message struct {
	ID        string    `json:"id"`        // 消息 ID
	Role      string    `json:"role"`      // 角色: user/assistant/system
	Content   string    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
	ToolCalls []ToolCall `json:"toolCalls,omitempty"` // 工具调用
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string                 `json:"id"`       // 工具调用 ID
	Name     string                 `json:"name"`     // 工具名称
	Input    map[string]interface{} `json:"input"`    // 输入参数
	Output   string                 `json:"output"`   // 输出结果
	Status   string                 `json:"status"`   // 状态: pending/success/failed
}

// NewMessage 创建新消息
func NewMessage(role, content string) *Message {
	return &Message{
		ID:        "msg-" + time.Now().Format("20060102150405000"),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		ToolCalls: make([]ToolCall, 0),
	}
}

// AddToolCall 添加工具调用
func (m *Message) AddToolCall(toolCall ToolCall) {
	m.ToolCalls = append(m.ToolCalls, toolCall)
}
