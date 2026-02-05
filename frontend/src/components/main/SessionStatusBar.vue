<script setup lang="ts">
import type { WorkspaceInfo } from "@/types/workspace";

interface Props {
  workspace: WorkspaceInfo | null;
  conversationId: string;
}

defineProps<Props>();
</script>

<template>
  <div v-if="workspace" class="session-status">
    <span class="workspace-name">{{ workspace.name }}</span>
    <div class="session-actions">
      <span class="status-indicator" :class="{ active: conversationId }">
        <span class="status-dot"></span>
        <span class="status-text">{{
          conversationId ? "会话中" : "未连接"
        }}</span>
      </span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.session-status {
  padding: 12px 16px;
  background: #f5f5f5;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;

  .workspace-name {
    font-size: 14px;
    font-weight: 600;
    color: #333;
  }

  .session-actions {
    display: flex;
    align-items: center;
  }

  .status-indicator {
    font-size: 12px;
    color: #999;
    padding: 4px 10px;
    border-radius: 12px;
    background: #e0e0e0;
    display: flex;
    align-items: center;
    gap: 6px;

    &.active {
      background: #e8f5e9;
      color: #4caf50;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #9e9e9e;
      transition: background 0.3s;
    }

    &.active .status-dot {
      background: #4caf50;
      box-shadow: 0 0 6px rgba(76, 175, 80, 0.5);
    }

    .status-text {
      font-size: 12px;
    }
  }
}
</style>
