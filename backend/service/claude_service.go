package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"claude_desktop/backend/manager/conversation"
)

// ClaudeService Claude API 服务
type ClaudeService struct {
	mu          sync.Mutex
	projectPath string
}

// NewClaudeService 创建 Claude 服务实例
func NewClaudeService() *ClaudeService {
	return &ClaudeService{}
}

// SetProjectPath 设置当前项目路径
func (s *ClaudeService) SetProjectPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projectPath = path
}

// SendRequest 发送消息到 Claude
func (s *ClaudeService) SendRequest(ctx context.Context, messages []conversation.Message, onChunk func(string)) error {
	s.mu.Lock()
	projectPath := s.projectPath
	s.mu.Unlock()

	// 构建输入消息
	var inputContent string
	for _, msg := range messages {
		if msg.Role == "user" {
			inputContent += msg.Content + "\n"
		}
	}

	// 构建 claude 命令（使用 --print 非交互模式）
	cmd := exec.CommandContext(ctx, "claude", "--print", inputContent)

	// 设置工作目录为项目路径
	cmd.Dir = projectPath

	// 创建标准输出和错误管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude command: %w", err)
	}

	// 读取响应
	var wg sync.WaitGroup
	wg.Add(2)

	// 处理标准输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if onChunk != nil {
				onChunk(line + "\n")
			}
		}
	}()

	// 处理错误输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			// 可以记录错误日志
			fmt.Printf("Claude stderr: %s\n", line)
		}
	}()

	// 等待输出完成
	wg.Wait()

	// 等待命令结束
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude command failed: %w", err)
	}

	return nil
}

// SendMessage 发送单个消息并获取完整响应
func (s *ClaudeService) SendMessage(ctx context.Context, content string, onChunk func(string)) (string, error) {
	s.mu.Lock()
	projectPath := s.projectPath
	s.mu.Unlock()

	// 构建 claude 命令（使用 --print 非交互模式）
	cmd := exec.CommandContext(ctx, "claude", "--print", content)

	// 设置工作目录为项目路径
	cmd.Dir = projectPath

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 运行命令
	err := cmd.Run()

	// 流式输出
	output := stdout.String()
	if onChunk != nil && output != "" {
		// 按行流式输出
		scanner := bufio.NewScanner(strings.NewReader(output))
		for scanner.Scan() {
			onChunk(scanner.Text() + "\n")
		}
	}

	if err != nil {
		return "", fmt.Errorf("claude command failed: %w, stderr: %s", err, stderr.String())
	}

	return output, nil
}

// StreamEvent represents a stream event from Claude CLI
type StreamEvent struct {
	Type      string                 `json:"type"`
	EventType string                 `json:"event_type"` // for legacy format
	Event     map[string]interface{} `json:"event"`
}

// StreamMessage 流式发送消息
func (s *ClaudeService) StreamMessage(ctx context.Context, content string, onChunk func(string)) error {
	s.mu.Lock()
	projectPath := s.projectPath
	s.mu.Unlock()

	// 构建 claude 命令（使用 --print 非交互模式 + 流式 JSON 输出）
	cmd := exec.CommandContext(ctx, "claude", "--print", content,
		"--output-format", "stream-json",
		"--verbose",
		"--include-partial-messages")

	// 设置工作目录为项目路径
	cmd.Dir = projectPath

	// 创建管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude command: %w", err)
	}

	// 读取输出
	var wg sync.WaitGroup
	wg.Add(2)

	// 处理标准输出 - 逐行解析 JSON（优化版）
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		// 增加缓冲区大小以处理长 JSON 行
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// 解析 JSON
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}

			// 处理不同类型的事件
			switch eventType := raw["type"].(string); eventType {
			case "stream_event":
				// 处理流式事件
				if event, ok := raw["event"].(map[string]interface{}); ok {
					switch eventStr := event["type"].(string); eventStr {
					case "content_block_delta":
						// 文本内容增量
						if delta, ok := event["delta"].(map[string]interface{}); ok {
							if text, ok := delta["text"].(string); ok && text != "" && onChunk != nil {
								onChunk(text)
							}
						}
					case "tool_use_delta":
						// 工具调用增量（显示工具调用过程）
						if delta, ok := event["delta"].(map[string]interface{}); ok {
							if name, ok := delta["name"].(string); ok && name != "" && onChunk != nil {
								onChunk(fmt.Sprintf("\n🔧 使用工具: %s\n", name))
							}
							if input, ok := delta["input"].(string); ok && input != "" && onChunk != nil {
								onChunk(fmt.Sprintf("   参数: %s\n", input))
							}
						}
					}
				}
			case "tool_use":
				// 显示工具调用信息
				if name, ok := raw["name"].(string); ok && onChunk != nil {
					onChunk(fmt.Sprintf("\n🔧 调用工具: %s\n", name))
				}
			case "tool_result":
				// 显示工具结果
				if onChunk != nil {
					onChunk("\n✓ 工具执行完成\n")
				}
			}
		}
	}()

	// 处理错误输出（静默）
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// 静默处理，不输出日志
		}
	}()

	// 等待完成
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude command failed: %w", err)
	}

	return nil
}

// ValidateEnvironment 验证 Claude 环境是否可用
func (s *ClaudeService) ValidateEnvironment(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "claude", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("claude command not available: %w", err)
	}

	// 解析版本
	fmt.Printf("Claude version: %s\n", strings.TrimSpace(string(output)))
	return nil
}

// ConversationManager 对话管理器
type ConversationManager struct {
	storage conversation.Storage
	claude  *ClaudeService
}

// NewConversationManager 创建对话管理器
func NewConversationManager(storage conversation.Storage) *ConversationManager {
	return &ConversationManager{
		storage: storage,
		claude:  NewClaudeService(),
	}
}

// CreateConversation 创建新对话
func (m *ConversationManager) CreateConversation(title, projectPath string) (*conversation.Conversation, error) {
	conv := conversation.NewConversation(title, projectPath)
	if err := m.storage.SaveConversation(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// DeleteConversation 删除对话
func (m *ConversationManager) DeleteConversation(id string) error {
	return m.storage.DeleteConversation(id)
}

// GetConversation 获取对话
func (m *ConversationManager) GetConversation(id string) (*conversation.Conversation, error) {
	return m.storage.LoadConversation(id)
}

// ListConversations 列出所有对话
func (m *ConversationManager) ListConversations() ([]*conversation.Conversation, error) {
	return m.storage.ListConversations()
}

// UpdateConversation 更新对话
func (m *ConversationManager) UpdateConversation(conv *conversation.Conversation) error {
	return m.storage.SaveConversation(conv)
}

// GetConversationByProjectPath 根据项目路径获取最近的对话
func (m *ConversationManager) GetConversationByProjectPath(projectPath string) (*conversation.Conversation, error) {
	conversations, err := m.storage.ListConversations()
	if err != nil {
		return nil, err
	}

	// 查找匹配项目路径的对话，返回最近的一个
	var latestConv *conversation.Conversation

	for _, conv := range conversations {
		if conv.ProjectPath == projectPath {
			if latestConv == nil || conv.UpdatedAt.After(latestConv.UpdatedAt) {
				latestConv = conv
			}
		}
	}

	if latestConv == nil {
		return nil, fmt.Errorf("no conversation found for project path: %s", projectPath)
	}

	return latestConv, nil
}

// SendMessage 发送消息并保存
func (m *ConversationManager) SendMessage(convID, content string) (*conversation.Conversation, error) {
	// 加载对话
	conv, err := m.storage.LoadConversation(convID)
	if err != nil {
		return nil, err
	}

	// 添加用户消息
	userMsg := conversation.NewMessage("user", content)
	conv.AddMessage(*userMsg)

	// 设置项目路径
	m.claude.SetProjectPath(conv.ProjectPath)

	// 发送到 Claude 并获取响应
	var responseBuilder strings.Builder
	err = m.claude.StreamMessage(context.Background(), content, func(chunk string) {
		responseBuilder.WriteString(chunk)
	})
	if err != nil {
		return nil, err
	}

	// 添加助手消息
	assistantMsg := conversation.NewMessage("assistant", responseBuilder.String())
	conv.AddMessage(*assistantMsg)

	// 保存对话
	if err := m.storage.SaveConversation(conv); err != nil {
		return nil, err
	}

	return conv, nil
}

// SendMessageWithCallback 发送消息并提供回调
func (m *ConversationManager) SendMessageWithCallback(
	convID, content string,
	onChunk func(string),
) (*conversation.Conversation, error) {
	// 加载对话
	conv, err := m.storage.LoadConversation(convID)
	if err != nil {
		return nil, err
	}

	// 添加用户消息
	userMsg := conversation.NewMessage("user", content)
	conv.AddMessage(*userMsg)

	// 保存用户消息
	if err := m.storage.SaveConversation(conv); err != nil {
		return nil, err
	}

	// 设置项目路径
	m.claude.SetProjectPath(conv.ProjectPath)

	// 发送到 Claude 并流式接收响应
	var responseBuilder strings.Builder
	err = m.claude.StreamMessage(context.Background(), content, func(chunk string) {
		responseBuilder.WriteString(chunk)
		if onChunk != nil {
			onChunk(chunk)
		}
	})
	if err != nil {
		return nil, err
	}

	// 添加助手消息
	assistantMsg := conversation.NewMessage("assistant", responseBuilder.String())
	conv.AddMessage(*assistantMsg)

	// 保存完整对话
	if err := m.storage.SaveConversation(conv); err != nil {
		return nil, err
	}

	return conv, nil
}
