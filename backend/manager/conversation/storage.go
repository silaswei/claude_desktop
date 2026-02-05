package conversation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage 存储接口
type Storage interface {
	// 对话存储
	SaveConversation(conv *Conversation) error
	LoadConversation(id string) (*Conversation, error)
	DeleteConversation(id string) error
	ListConversations() ([]*Conversation, error)
}

// JSONStorage JSON 文件存储实现
type JSONStorage struct {
	mu      sync.RWMutex
	baseDir string
	convDir string
}

// NewJSONStorage 创建 JSON 存储实例
func NewJSONStorage() (*JSONStorage, error) {
	// 获取用户主目录
	homeDir, _ := os.UserHomeDir()
	baseDir := filepath.Join(homeDir, ".claude-desktop")

	// 创建基础目录
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// 创建对话目录
	convDir := filepath.Join(baseDir, "conversations")
	if err := os.MkdirAll(convDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create conversations directory: %w", err)
	}

	return &JSONStorage{
		baseDir: baseDir,
		convDir: convDir,
	}, nil
}

// ==================== 对话存储 ====================

// SaveConversation 保存对话
func (s *JSONStorage) SaveConversation(conv *Conversation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.convDir, conv.ID+".json")
	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadConversation 加载对话
func (s *JSONStorage) LoadConversation(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := filepath.Join(s.convDir, id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read conversation file: %w", err)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return &conv, nil
}

// DeleteConversation 删除对话
func (s *JSONStorage) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := filepath.Join(s.convDir, id+".json")
	return os.Remove(filename)
}

// ListConversations 列出所有对话
func (s *JSONStorage) ListConversations() ([]*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.convDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read conversations directory: %w", err)
	}

	conversations := make([]*Conversation, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 从文件名提取 ID（移除 .json 后缀）
		id := entry.Name()[:len(entry.Name())-5]
		conv, err := s.LoadConversation(id)
		if err != nil {
			continue // 跳过损坏的文件
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}
