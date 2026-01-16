// 对话相关类型定义

// 消息角色
export type MessageRole = 'user' | 'assistant' | 'system';

// 工具调用状态
export type ToolCallStatus = 'pending' | 'success' | 'failed';

// 工具调用
export interface ToolCall {
  id: string;
  name: string;
  input: Record<string, any>;
  output: string;
  status: ToolCallStatus;
}

// 附件
export interface Attachment {
  id: string;
  name: string;
  type: string;
  size: number;
  path: string;
}

// 文件引用
export interface FileRef {
  path: string;
  name: string;
  type: string;
}

// 消息
export interface Message {
  id: string;
  role: MessageRole;
  content: string;
  timestamp: string;
  toolCalls?: ToolCall[];
  attachments?: Attachment[];
  fileReferences?: FileRef[];
}

// 对话
export interface Conversation {
  id: string;
  title: string;
  projectPath: string;
  createdAt: string;
  updatedAt: string;
  messages: Message[];
}

// 创建对话请求
export interface CreateConversationRequest {
  title: string;
  workspacePath: string; // 前端使用 workspacePath，传递给后端时会映射到 projectPath
}

// 发送消息请求
export interface SendMessageRequest {
  conversationId: string;
  content: string;
  attachments?: string[]; // 文件路径列表
  fileReferences?: string[]; // 引用的文件路径
}

