<script setup lang="ts">
import { computed } from "vue";
import { useEnvStore } from "@/stores/env";
import type { DetectionResult } from "@/types/env";

// 定义检测项接口
interface DetectionItem {
  name: string;
  status: "pending" | "success" | "failed";
  version: string;
  required: boolean;
  message?: string; // 添加消息字段
}

// 定义 emit
const emit = defineEmits(["enter-main"]);

const envStore = useEnvStore();

// 默认的待检测项目列表
const defaultDetectionItems: DetectionItem[] = [
  { name: "Node.js", status: "pending", version: "", required: true },
  { name: "npm", status: "pending", version: "", required: true },
  { name: "Claude Code CLI", status: "pending", version: "", required: true },
  { name: "网络连通性", status: "pending", version: "", required: true },
  { name: "Git", status: "pending", version: "", required: false },
];

// 检测项列表
const detectionItems = computed<DetectionItem[]>(() => {
  if (!envStore.envInfo || !envStore.envInfo.results) {
    return defaultDetectionItems;
  }

  return envStore.envInfo.results.map(
    (r: DetectionResult): DetectionItem => ({
      name: r.name,
      status: r.status as "pending" | "success" | "failed",
      version: r.version || "",
      required: r.required,
      message: r.message, // 映射消息字段
    })
  );
});

// 检测进度
const progress = computed(() => {
  const items = detectionItems.value;
  const completed = items.filter((i) => i.status !== "pending").length;
  return (completed / items.length) * 100;
});

// 是否显示进入主页面按钮（检测完成且全部通过）
const showEnterButton = computed(() => {
  return envStore.envInfo !== null && envStore.allPassed;
});

// 进入主页面
const handleEnterMain = () => {
  emit("enter-main");
};
</script>

<template>
  <div class="launch-screen">
    <div class="content">
      <div class="logo">
        <div class="logo-icon">🔍</div>
        <div class="logo-text">Claude Desktop</div>
        <div class="logo-subtitle">环境检测中...</div>
      </div>

      <div class="progress-section">
        <div class="progress-text">检测进度: {{ Math.round(progress) }}%</div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: `${progress}%` }"></div>
        </div>
      </div>

      <!-- 环境检测结果列表 -->
      <div class="detection-list">
        <div class="list-header">
          <span>环境项</span>
          <span>状态</span>
          <span>版本信息</span>
        </div>

        <div
          v-for="item in detectionItems"
          :key="item.name"
          class="list-item"
          :class="`status-${item.status}`"
        >
          <div class="item-name">
            <span class="item-icon">
              <span v-if="item.status === 'pending'">⏳</span>
              <span v-else-if="item.status === 'success'">✓</span>
              <span v-else>✗</span>
            </span>
            {{ item.name }}
            <span v-if="item.required" class="required-badge">必需</span>
            <span v-else class="optional-badge">可选</span>
          </div>

          <div class="item-status">
            <span v-if="item.status === 'pending'" class="status-pending"
              >检测中</span
            >
            <span v-else-if="item.status === 'success'" class="status-success"
              >通过</span
            >
            <span v-else class="status-failed">失败</span>
          </div>

          <div class="item-version">
            <div
              v-if="item.status === 'failed' && item.message"
              class="version-error"
            >
              {{ item.message }}
            </div>
            <div v-else-if="item.version">
              {{ item.version }}
            </div>
            <div
              v-else-if="item.status === 'pending'"
              class="version-detecting"
            >
              检测中...
            </div>
            <div v-else class="version-unknown">未知</div>
          </div>
        </div>
      </div>

      <!-- 检测摘要 -->
      <div v-if="envStore.envInfo" class="summary-section">
        <div class="summary-title">检测摘要</div>
        <div class="summary-content">
          <div class="summary-item">
            <span class="summary-label">必需项通过:</span>
            <span class="summary-value"
              >{{ envStore.requiredPassed }}/{{ envStore.requiredTotal }}</span
            >
          </div>
          <div class="summary-item">
            <span class="summary-label">整体状态:</span>
            <span
              class="summary-value"
              :class="{
                'status-good': envStore.allPassed,
                'status-bad': envStore.hasRequiredFailed,
                'status-warning':
                  !envStore.allPassed && !envStore.hasRequiredFailed,
              }"
            >
              {{ envStore.allPassed ? "全部通过 ✓" : "部分失败 ✗" }}
            </span>
          </div>
        </div>
      </div>

      <!-- 进入主页面按钮 -->
      <button
        v-if="showEnterButton"
        class="enter-main-btn"
        @click="handleEnterMain"
      >
        进入主页面
      </button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.launch-screen {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  overflow: hidden;
}

.content {
  background: rgba(255, 255, 255, 0.98);
  border-radius: 20px;
  padding: 40px;
  max-width: 1200px;
  width: 100%;
  height: calc(100% - 40px);
  max-height: calc(100% - 40px);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.logo {
  text-align: center;
  margin-bottom: 15px;

  .logo-icon {
    font-size: 40px;
    margin-bottom: 8px;
  }

  .logo-text {
    font-size: 22px;
    font-weight: bold;
    color: #333;
    margin-bottom: 4px;
  }

  .logo-subtitle {
    font-size: 13px;
    color: #666;
  }
}

.progress-section {
  margin-bottom: 15px;

  .progress-text {
    font-size: 12px;
    color: #666;
    margin-bottom: 5px;
    text-align: center;
  }

  .progress-bar {
    width: 100%;
    height: 4px;
    background: #e0e0e0;
    border-radius: 2px;
    overflow: hidden;

    .progress-fill {
      height: 100%;
      background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
      transition: width 0.3s ease;
    }
  }
}

.detection-list {
  margin-bottom: 12px;
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  max-height: calc(100vh - 500px);
  overflow-y: auto;
  overflow-x: visible;
  flex: 1;
  min-height: 0;
  padding-right: 8px;

  /* 自定义滚动条样式 */
  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 3px;
  }

  &::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 3px;

    &:hover {
      background: #a8a8a8;
    }
  }

  .list-header {
    display: grid;
    grid-template-columns: 2fr 1fr 1.5fr;
    gap: 10px;
    padding: 8px 12px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 6px 6px 0 0;
    color: white;
    font-size: 12px;
    font-weight: 600;
  }

  .list-item {
    display: grid;
    grid-template-columns: 2fr 1fr 1.5fr;
    gap: 10px;
    padding: 10px 12px;
    background: #f8f9fa;
    border-bottom: 1px solid #e0e0e0;
    align-items: center;
    transition: all 0.2s ease;

    &:last-child {
      border-radius: 0 0 8px 8px;
      border-bottom: none;
    }

    &.status-success {
      background: #e8f5e9;

      .item-icon span {
        color: #4caf50;
      }

      .status-success {
        color: #4caf50;
        font-weight: 600;
      }
    }

    &.status-failed {
      background: #ffebee;

      .item-icon span {
        color: #f44336;
      }

      .status-failed {
        color: #f44336;
        font-weight: 600;
      }
    }

    .item-name {
      font-size: 13px;
      color: #333;
      display: flex;
      align-items: center;
      gap: 8px;

      .item-icon {
        font-size: 16px;
      }

      .required-badge {
        font-size: 9px;
        padding: 2px 5px;
        background: #ff5722;
        color: white;
        border-radius: 3px;
        font-weight: 500;
      }

      .optional-badge {
        font-size: 9px;
        padding: 2px 5px;
        background: #9e9e9e;
        color: white;
        border-radius: 3px;
        font-weight: 500;
      }
    }

    .item-status {
      font-size: 12px;

      .status-pending {
        color: #ff9800;
      }
    }

    .item-version {
      font-size: 12px;
      color: #666;
      font-family: "JetBrains Mono", monospace;

      .version-error {
        color: #f44336;
        font-weight: 500;
        line-height: 1.4;
      }

      .version-detecting {
        color: #ff9800;
        font-style: italic;
      }

      .version-unknown {
        color: #9e9e9e;
      }
    }
  }
}

.summary-section {
  padding: 12px 15px;
  background: #f5f5f5;
  border-radius: 8px;
  flex-shrink: 0;

  .summary-title {
    font-size: 13px;
    font-weight: 600;
    color: #333;
    margin-bottom: 8px;
  }

  .summary-content {
    display: flex;
    gap: 25px;

    .summary-item {
      display: flex;
      align-items: center;
      gap: 6px;

      .summary-label {
        font-size: 12px;
        color: #666;
      }

      .summary-value {
        font-size: 14px;
        font-weight: 600;
        color: #333;

        &.status-good {
          color: #4caf50;
        }

        &.status-bad {
          color: #f44336;
        }

        &.status-warning {
          color: #ff9800;
        }
      }
    }
  }
}

.enter-main-btn {
  padding: 12px 35px;
  font-size: 15px;
  font-weight: 600;
  color: white;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  margin-top: 12px;
  flex-shrink: 0;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
  }

  &:active {
    transform: translateY(0);
  }
}
</style>
