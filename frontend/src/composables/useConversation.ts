import { ref } from "vue";
import {
  ConversationCreate,
  ConversationGetByProjectPath,
  ConversationSendWithEvents,
  WorkspaceSetActiveConversation,
} from "../../wailsjs/go/app/App";
import type { WorkspaceInfo } from "@/types/workspace";

export function useConversation() {
  const conversationId = ref("");
  const messages = ref<
    Array<{
      id: string;
      role: "user" | "assistant";
      content: string;
      timestamp: string;
    }>
  >([]);

  // 流式输出优化：批量更新
  let streamingMessage: {
    id: string;
    role: "user" | "assistant";
    content: string;
    timestamp: string;
  } | null = null;
  let streamingBuffer = "";
  let streamingTimer: number | null = null;

  // 思考中消息
  let thinkingMessageId: string | null = null;

  // 刷新流式输出显示
  function flushStreamingMessage() {
    if (streamingMessage && streamingBuffer) {
      streamingMessage.content += streamingBuffer;
      streamingBuffer = "";
    }
    if (streamingTimer !== null) {
      clearTimeout(streamingTimer);
      streamingTimer = null;
    }
  }

  // 加载工作区的历史对话
  async function loadWorkspaceConversation(projectPath: string) {
    try {
      const conv = await ConversationGetByProjectPath(projectPath);

      if (conv && conv.messages && conv.messages.length > 0) {
        messages.value = conv.messages.map((msg: any) => ({
          id: msg.id || `msg-${Date.now()}-${Math.random()}`,
          role: msg.role,
          content: msg.content,
          timestamp: msg.timestamp || new Date().toISOString(),
        }));
        conversationId.value = conv.id;

        await WorkspaceSetActiveConversation(conv.id);
      } else {
        messages.value = [];
        conversationId.value = "";
      }
    } catch (error) {
      messages.value = [];
      conversationId.value = "";
    }
  }

  // 创建新会话
  async function createConversation(workspace: WorkspaceInfo) {
    const conv = await ConversationCreate(workspace.name, workspace.path);
    conversationId.value = conv.id;
    messages.value = [];
    await WorkspaceSetActiveConversation(conv.id);
    return conv;
  }

  // 发送消息
  async function sendMessage(content: string) {
    const userMessage = {
      id: `msg-${Date.now()}`,
      role: "user" as const,
      content,
      timestamp: new Date().toISOString(),
    };
    messages.value.push(userMessage);

    // 如果没有会话，返回 null 表示需要先创建会话
    if (!conversationId.value) {
      return null;
    }

    try {
      await ConversationSendWithEvents(conversationId.value, content);
    } finally {
      flushStreamingMessage();
      streamingMessage = null;
      streamingBuffer = "";
      streamingTimer = null;
    }

    return true;
  }

  // 处理 Claude 响应
  function handleClaudeResponse(data: any) {
    let content = "";
    if (typeof data === "string") {
      content = data;
    } else if (data?.content) {
      content = data.content;
    }

    if (!content) return;

    const trimmedContent = content.trim();
    if (!trimmedContent) {
      return;
    }

    // 移除思考中消息
    if (thinkingMessageId) {
      const thinkingIndex = messages.value.findIndex(
        (m) => m.id === thinkingMessageId
      );
      if (thinkingIndex !== -1) {
        messages.value.splice(thinkingIndex, 1);
      }
      thinkingMessageId = null;
    }

    // 查找或创建流式消息对象
    if (!streamingMessage) {
      const lastMessage = messages.value[messages.value.length - 1];
      if (lastMessage?.role === "assistant") {
        streamingMessage = lastMessage;
      } else {
        streamingMessage = {
          id: `msg-${Date.now()}`,
          role: "assistant" as const,
          content: "",
          timestamp: new Date().toISOString(),
        };
        messages.value.push(streamingMessage);
      }
    }

    streamingBuffer += content;

    if (streamingTimer === null) {
      streamingTimer = window.setTimeout(() => {
        flushStreamingMessage();
      }, 16);
    }
  }

  // 处理 Claude 开始思考
  function handleClaudeThinking() {
    thinkingMessageId = `msg-thinking-${Date.now()}`;
    messages.value.push({
      id: thinkingMessageId,
      role: "assistant" as const,
      content: "思考中",
      timestamp: new Date().toISOString(),
    });
  }

  // 处理 Claude 完成
  function handleClaudeComplete(data: any) {
    flushStreamingMessage();
    const hasContent = data?.hasContent ?? true;

    if (thinkingMessageId && !hasContent) {
      const thinkingIndex = messages.value.findIndex(
        (m) => m.id === thinkingMessageId
      );
      if (thinkingIndex !== -1) {
        messages.value.splice(thinkingIndex, 1);
      }
      thinkingMessageId = null;

      messages.value.push({
        id: `msg-error-${Date.now()}`,
        role: "assistant" as const,
        content: "抱歉，没有收到任何响应。请检查 Claude CLI 是否正确配置。",
        timestamp: new Date().toISOString(),
      });
    }
  }

  // 处理 Claude 错误
  function handleClaudeError(data: any) {
    const errorMsg = data?.error || "未知错误";

    if (thinkingMessageId) {
      const thinkingIndex = messages.value.findIndex(
        (m) => m.id === thinkingMessageId
      );
      if (thinkingIndex !== -1) {
        messages.value.splice(thinkingIndex, 1);
      }
      thinkingMessageId = null;
    }

    messages.value.push({
      id: `msg-error-${Date.now()}`,
      role: "assistant" as const,
      content: `发生错误: ${errorMsg}`,
      timestamp: new Date().toISOString(),
    });
  }

  return {
    conversationId,
    messages,
    loadWorkspaceConversation,
    createConversation,
    sendMessage,
    handleClaudeResponse,
    handleClaudeThinking,
    handleClaudeComplete,
    handleClaudeError,
  };
}
