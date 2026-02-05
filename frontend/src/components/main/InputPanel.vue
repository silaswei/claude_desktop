<script setup lang="ts">
import { ref, nextTick } from "vue";

interface Props {
  isThinking: boolean;
  isSending: boolean;
}

interface Emits {
  (e: "send"): void;
  (e: "stop-thinking"): void;
}

defineProps<Props>();
defineEmits<Emits>();

const messageInput = ref("");
const messageInputRef = ref<HTMLTextAreaElement | null>(null);

defineExpose({
  messageInput,
  messageInputRef,
  focus: () => {
    nextTick(() => {
      if (messageInputRef.value) {
        messageInputRef.value.focus();
      }
    });
  },
  clear: () => {
    messageInput.value = "";
  },
  getText: () => messageInput.value,
  setText: (text: string) => {
    messageInput.value = text;
  },
});
</script>

<template>
  <div class="input-panel">
    <textarea
      ref="messageInputRef"
      v-model="messageInput"
      class="message-input"
      placeholder="输入消息... (Shift+Enter 换行, Enter 发送)"
      rows="3"
      @keydown.enter.exact.prevent="
        isThinking ? $emit('stop-thinking') : $emit('send')
      "
    ></textarea>
    <div class="input-actions">
      <span class="input-hint">{{ messageInput.length }} 字符</span>
      <!-- 发送按钮 / 停止按钮 -->
      <button
        v-if="!isThinking"
        class="send-btn"
        :disabled="!messageInput.trim() || isSending"
        @click="$emit('send')"
      >
        {{ isSending ? "发送中..." : "发送" }}
      </button>
      <button v-else class="stop-btn-inline" @click="$emit('stop-thinking')">
        <span class="stop-icon">⏹</span>
        <span class="stop-text">停止</span>
      </button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.input-panel {
  border-top: 1px solid #e0e0e0;
  padding: 12px;
  background: #fafafa;

  .message-input {
    width: 100%;
    padding: 10px;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    font-size: 14px;
    font-family: inherit;
    resize: none;
    outline: none;
    transition: border-color 0.2s;

    &:focus {
      border-color: #667eea;
    }
  }

  .input-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 8px;

    .input-hint {
      font-size: 12px;
      color: #666;
    }

    .send-btn {
      padding: 6px 16px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      border: none;
      border-radius: 6px;
      font-size: 13px;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.2s;

      &:hover:not(:disabled) {
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
      }

      &:disabled {
        opacity: 0.5;
        cursor: not-allowed;
      }
    }

    .stop-btn-inline {
      padding: 6px 16px;
      background: linear-gradient(135deg, #ff5252 0%, #ff1744 100%);
      color: white;
      border: none;
      border-radius: 6px;
      font-size: 13px;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.2s;
      display: flex;
      align-items: center;
      gap: 6px;

      .stop-icon {
        font-size: 14px;
      }

      .stop-text {
        font-size: 13px;
      }

      &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(255, 82, 82, 0.4);
      }

      &:active {
        transform: translateY(0);
      }
    }
  }
}
</style>
