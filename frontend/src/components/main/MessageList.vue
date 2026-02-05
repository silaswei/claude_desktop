<script setup lang="ts">
import { ref, nextTick, watch } from "vue";

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: string;
}

interface Props {
  messages: Message[];
}

const props = defineProps<Props>();

const messageListRef = ref<HTMLElement | null>(null);

// 格式化时间
function formatTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  if (diff < 60000) return "刚刚";
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`;
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`;

  return date.toLocaleDateString("zh-CN");
}

// 滚动到底部
function scrollToBottom() {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
  }
}

// 检查是否在底部（100px以内）
function isNearBottom(): boolean {
  if (!messageListRef.value) return false;
  const el = messageListRef.value;
  const threshold = 100;
  return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
}

// 监听消息变化，自动滚动到底部
watch(
  () => props.messages,
  async () => {
    if (isNearBottom()) {
      await nextTick();
      scrollToBottom();
    }
  },
  { deep: true }
);

defineExpose({
  scrollToBottom,
  messageListRef,
});
</script>

<template>
  <div v-if="messages.length > 0" ref="messageListRef" class="message-list">
    <div
      v-for="msg in messages"
      :key="msg.id"
      v-show="
        msg.content.trim() !== '' ||
        msg.role === 'user' ||
        msg.id.includes('thinking')
      "
      class="message-item"
      :class="msg.role"
    >
      <div class="message-header">
        <span class="message-role">
          {{ msg.role === "user" ? "用户" : "Claude" }}
        </span>
        <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
      </div>
      <!-- 思考中消息显示动画 -->
      <div
        v-if="msg.id.includes('thinking')"
        class="message-content thinking-content"
      >
        <span class="thinking-text">思考中</span>
        <span class="thinking-dots">
          <span class="dot"></span>
          <span class="dot"></span>
          <span class="dot"></span>
        </span>
      </div>
      <!-- 普通消息内容 -->
      <div v-else class="message-content">{{ msg.content }}</div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-thumb {
    background: #ccc;
    border-radius: 3px;
  }

  .message-item {
    margin-bottom: 16px;
    padding: 12px;
    border-radius: 8px;
    background: #f9f9f9;

    &.user {
      background: #e3f2fd;
    }

    &.assistant {
      background: #f5f5f5;
    }

    .message-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;

      .message-role {
        font-size: 12px;
        font-weight: 500;
        color: #333;
      }

      .message-time {
        font-size: 11px;
        color: #666;
      }
    }

    .message-content {
      font-size: 14px;
      line-height: 1.6;
      color: #333;
      white-space: pre-wrap;
      word-wrap: break-word;

      &.thinking-content {
        display: flex;
        align-items: center;
        gap: 8px;

        .thinking-text {
          font-size: 14px;
          color: #666;
        }

        .thinking-dots {
          display: flex;
          align-items: center;
          gap: 4px;

          .dot {
            width: 6px;
            height: 6px;
            border-radius: 50%;
            background: #667eea;
            animation: thinking-bounce 1.4s ease-in-out infinite;

            &:nth-child(1) {
              animation-delay: 0s;
            }

            &:nth-child(2) {
              animation-delay: 0.2s;
            }

            &:nth-child(3) {
              animation-delay: 0.4s;
            }
          }
        }
      }
    }
  }
}
</style>

<!-- 思考动画关键帧 -->
<style>
@keyframes thinking-bounce {
  0%,
  60%,
  100% {
    transform: translateY(0);
    opacity: 0.3;
  }
  30% {
    transform: translateY(-8px);
    opacity: 1;
  }
}
</style>
