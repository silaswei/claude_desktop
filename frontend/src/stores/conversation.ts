import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { CreateConversationRequest } from '../types/conversation';
import type { conversation } from '../../wailsjs/go/models';
import {
  ConversationCreate,
  ConversationDelete,
  ConversationInfo,
  ConversationList,
  ConversationUpdate,
  ConversationSend,
  ConversationSendWithCallback
} from '../../wailsjs/go/app/App';

// 使用 Wails 生成的类型
type Conversation = conversation.Conversation;

export const useConversationStore = defineStore('conversation', () => {
  // ==================== 状态 ====================

  // 对话列表
  const conversations = ref<Conversation[]>([]);

  // 当前对话
  const currentConversation = ref<Conversation | null>(null);

  // 加载状态
  const loading = ref(false);

  // 错误信息
  const error = ref<string | null>(null);

  // ==================== 计算属性 ====================

  // 按工作区分组的对话
  const conversationsByWorkspace = computed(() => {
    const groups: Record<string, Conversation[]> = {};
    conversations.value.forEach((conv) => {
      const key = conv.projectPath || '未分类';
      if (!groups[key]) {
        groups[key] = [];
      }
      groups[key].push(conv);
    });
    return groups;
  });

  // 最近对话
  const recentConversations = computed(() => {
    return [...conversations.value]
      .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
      .slice(0, 10);
  });

  // 当前对话的消息列表
  const currentMessages = computed(() => {
    return currentConversation.value?.messages || [];
  });

  // ==================== 方法 ====================

  /**
   * 加载对话列表
   */
  async function loadConversations(): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const result = await ConversationList();
      conversations.value = result || [];
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 创建新对话
   */
  async function createConversation(request: CreateConversationRequest): Promise<Conversation> {
    loading.value = true;
    error.value = null;

    try {
      const result = await ConversationCreate(request.title, request.workspacePath);
      conversations.value.push(result);
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 删除对话
   */
  async function deleteConversation(id: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await ConversationDelete(id);

      // 从列表中移除
      conversations.value = conversations.value.filter((c) => c.id !== id);

      // 如果是当前对话，清空当前对话
      if (currentConversation.value?.id === id) {
        currentConversation.value = null;
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 获取对话详情
   */
  async function getConversation(id: string): Promise<Conversation> {
    loading.value = true;
    error.value = null;

    try {
      const result = await ConversationInfo(id);
      currentConversation.value = result;
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 更新对话
   */
  async function updateConversation(conversation: Conversation): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await ConversationUpdate(conversation);

      // 更新列表中的对话
      const index = conversations.value.findIndex((c) => c.id === conversation.id);
      if (index !== -1) {
        conversations.value[index] = conversation;
      }

      // 如果是当前对话，也更新
      if (currentConversation.value?.id === conversation.id) {
        currentConversation.value = conversation;
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 发送消息
   */
  async function sendMessage(
    conversationId: string,
    content: string,
    onChunk?: (chunk: string) => void
  ): Promise<Conversation> {
    loading.value = true;
    error.value = null;

    try {
      let result: Conversation;
      if (onChunk) {
        result = await ConversationSendWithCallback(conversationId, content, onChunk);
      } else {
        result = await ConversationSend(conversationId, content);
      }

      // 更新对话
      const index = conversations.value.findIndex((c) => c.id === result.id);
      if (index !== -1) {
        conversations.value[index] = result;
      }
      if (currentConversation.value?.id === result.id) {
        currentConversation.value = result;
      }
      return result;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 设置当前对话
   */
  function setCurrentConversation(conversation: Conversation | null): void {
    currentConversation.value = conversation;
  }

  /**
   * 清空错误
   */
  function clearError(): void {
    error.value = null;
  }

  // ==================== 返回 ====================

  return {
    // 状态
    conversations,
    currentConversation,
    loading,
    error,

    // 计算属性
    conversationsByWorkspace,
    recentConversations,
    currentMessages,

    // 方法
    loadConversations,
    createConversation,
    deleteConversation,
    getConversation,
    updateConversation,
    sendMessage,
    setCurrentConversation,
    clearError,
  };
});
