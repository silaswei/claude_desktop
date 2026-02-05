<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import type { WorkspaceInfo } from "@/types/workspace";

interface Props {
  workspaces: WorkspaceInfo[];
  selectedPath: string;
  selectedWorkspaceConversationId: string;
  selectedWorkspaceMessageCount: number;
}

interface Emits {
  (e: "select", workspace: WorkspaceInfo): void;
  (e: "toggle-sidebar"): void;
  (e: "clear-messages"): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

const openedMenuPath = ref<string | null>(null);

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

function toggleMenu(path: string, event: Event) {
  event.stopPropagation();
  if (openedMenuPath.value === path) {
    openedMenuPath.value = null;
  } else {
    openedMenuPath.value = path;
  }
}

function closeMenu() {
  openedMenuPath.value = null;
}

function handleClearMessages(event: Event) {
  event.stopPropagation();
  emit("clear-messages");
  closeMenu();
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (!target.closest(".workspace-actions")) {
    closeMenu();
  }
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<template>
  <div class="section workspace-list-section">
    <div class="section-header">
      <h3>工作区列表</h3>
      <button
        class="collapse-btn"
        @click="$emit('toggle-sidebar')"
        title="收起侧边栏"
      >
        ◀
      </button>
    </div>
    <div class="workspace-list">
      <div
        v-for="ws in workspaces"
        :key="ws.path"
        class="workspace-item"
        :class="{ active: selectedPath === ws.path }"
        @click="$emit('select', ws)"
      >
        <div class="workspace-item-content">
          <div class="workspace-info">
            <div class="workspace-name">{{ ws.name }}</div>
            <div class="workspace-meta">
              <span class="workspace-time">{{
                formatTime(ws.lastOpened)
              }}</span>
              <span v-if="ws.isOpen" class="workspace-status">● 当前</span>
            </div>
          </div>
        </div>
        <div class="workspace-actions">
          <div class="menu-container">
            <button
              class="more-btn"
              @click="toggleMenu(ws.path, $event)"
              title="更多操作"
            >
              ⋮
            </button>
            <div v-if="openedMenuPath === ws.path" class="dropdown-menu">
              <button
                v-if="
                  selectedPath === ws.path &&
                  selectedWorkspaceConversationId &&
                  selectedWorkspaceMessageCount > 0
                "
                class="menu-item"
                @click="handleClearMessages($event)"
              >
                清空聊天记录
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="workspaces.length === 0" class="empty-state">
        暂无工作区，点击右上角"打开文件夹"添加
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.workspace-list-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-bottom: 1px solid #e0e0e0;

  .section-header {
    padding: 10px 12px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #f9f9f9;

    h3 {
      margin: 0;
      font-size: 13px;
      color: #333;
      font-weight: 600;
    }

    .collapse-btn {
      background: none;
      border: none;
      font-size: 12px;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 4px;
      transition: background 0.2s;
      opacity: 0.6;

      &:hover {
        background: #e0e0e0;
        opacity: 1;
      }
    }
  }

  .workspace-list {
    flex: 1;
    overflow-y: auto;
    padding: 6px;

    &::-webkit-scrollbar {
      width: 6px;
    }

    &::-webkit-scrollbar-thumb {
      background: #ccc;
      border-radius: 3px;
    }
  }

  .workspace-item {
    padding: 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.2s;
    position: relative;
    margin-bottom: 6px;
    background: #fff;
    border: 1px solid #e0e0e0;

    &:hover {
      background: #f5f5f5;

      .workspace-actions {
        opacity: 1;
      }
    }

    &.active {
      background: #e8f5e9;
      border-color: #4caf50;
    }

    .workspace-item-content {
      display: flex;
      align-items: center;
    }

    .workspace-info {
      flex: 1;
      min-width: 0;

      .workspace-name {
        font-size: 13px;
        color: #333;
        font-weight: 500;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .workspace-meta {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-top: 4px;

        .workspace-time {
          font-size: 11px;
          color: #999;
        }

        .workspace-status {
          font-size: 11px;
          color: #4caf50;
          font-weight: 500;
        }
      }
    }

    .workspace-actions {
      position: absolute;
      top: 8px;
      right: 8px;
      display: flex;
      flex-direction: column;
      gap: 4px;
      opacity: 0;
      transition: opacity 0.2s;
      align-items: center;
    }

    .menu-container {
      position: relative;
    }

    .more-btn {
      background: none;
      border: none;
      color: #999;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 4px;
      transition: all 0.2s;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      font-weight: 300;
      line-height: 1;
      min-width: 28px;

      &:hover {
        background: rgba(0, 0, 0, 0.06);
      }
    }

    .remove-btn {
      background: none;
      border: none;
      color: #999;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 4px;
      transition: all 0.2s;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 13px;
      font-weight: 400;
      line-height: 1;
      min-width: 28px;

      &:hover {
        background: rgba(0, 0, 0, 0.06);
        color: #f44336;
      }
    }

    .dropdown-menu {
      position: absolute;
      top: 100%;
      right: 0;
      margin-top: 4px;
      background: #fff;
      border: 1px solid #e0e0e0;
      border-radius: 6px;
      box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
      min-width: 120px;
      z-index: 1000;
      overflow: hidden;

      .menu-item {
        width: 100%;
        padding: 8px 12px;
        background: none;
        border: none;
        text-align: left;
        font-size: 12px;
        color: #333;
        cursor: pointer;
        transition: background 0.2s;

        &:hover {
          background: #f5f5f5;
          color: #ff9800;
        }
      }
    }
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 12px;
    text-align: center;
    padding: 20px;
  }
}
</style>
