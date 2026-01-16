<script setup lang="ts">
import { computed } from 'vue';
import { useEnvStore } from '../../stores/env';
import type { DetectionResult } from '../../types/env';

// 定义 emit - 使用数组语法避免 TypeScript 类型错误
const emit = defineEmits(['retry', 'skip']);

const envStore = useEnvStore();

// 失败的检测项
const failedItems = computed<DetectionResult[]>(() => {
  if (!envStore.envInfo || !envStore.envInfo.results) {
    return [];
  }
  return envStore.envInfo.results.filter((r: DetectionResult) => r.status === 'failed');
});

// 是否有必需项失败
const hasRequiredFailed = computed(() => {
  return failedItems.value.some((r: DetectionResult) => r.required);
});

// 复制命令到剪贴板
const copyCommand = async (command: string) => {
  try {
    await navigator.clipboard.writeText(command);
    alert('命令已复制到剪贴板');
  } catch (err) {
    console.error('复制失败:', err);
  }
};
</script>

<template>
  <div class="failure-guide">
    <div class="content">
      <div class="header">
        <div class="icon">⚠️</div>
        <div class="title">环境检测未通过</div>
        <div class="subtitle">
          {{
            hasRequiredFailed ? "部分必需环境未正确配置" : "部分可选环境未安装"
          }}
        </div>
      </div>

      <!-- 所有检测结果列表 -->
      <div class="all-results">
        <div class="results-title">所有检测结果</div>
        <div class="results-list">
          <div
            v-for="item in envStore.envInfo?.results || []"
            :key="item.name"
            class="result-item"
            :class="`status-${item.status}`"
          >
            <div class="result-header">
              <div class="result-name">
                <span class="result-icon">
                  <span v-if="item.status === 'success'">✓</span>
                  <span v-else>✗</span>
                </span>
                {{ item.name }}
                <span v-if="item.required" class="required-badge">必需</span>
                <span v-else class="optional-badge">可选</span>
              </div>
              <div class="result-status">
                <span v-if="item.status === 'success'" class="status-success">通过</span>
                <span v-else class="status-failed">失败</span>
              </div>
            </div>

            <div class="result-details">
              <div v-if="item.version" class="detail-row">
                <span class="detail-label">版本:</span>
                <span class="detail-value">{{ item.version }}</span>
              </div>
              <div v-if="item.message" class="detail-row">
                <span class="detail-label">说明:</span>
                <span class="detail-value">{{ item.message }}</span>
              </div>
              <div v-if="item.fixCommand" class="detail-row detail-fix">
                <div class="fix-label">修复命令:</div>
                <div class="fix-command" @click="copyCommand(item.fixCommand)">
                  <code>{{ item.fixCommand }}</code>
                  <span class="copy-hint">📋 点击复制</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="actions">
        <button class="btn btn-retry" @click="emit('retry')">
          <span class="btn-icon">🔄</span>
          重新检测
        </button>
        <button
          v-if="!hasRequiredFailed"
          class="btn btn-skip"
          @click="emit('skip')"
        >
          跳过（有风险）
        </button>
      </div>

      <div v-if="hasRequiredFailed" class="warning">
        <div class="warning-text">
          ⚠️ 必需环境未配置可能会影响应用正常使用，请按照上述提示完成安装后重新检测。
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.failure-guide {
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

.header {
  text-align: center;
  margin-bottom: 25px;

  .icon {
    font-size: 50px;
    margin-bottom: 15px;
  }

  .title {
    font-size: 24px;
    font-weight: bold;
    color: #333;
    margin-bottom: 10px;
  }

  .subtitle {
    font-size: 14px;
    color: #666;
  }
}

.all-results {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 20px;
  padding-right: 5px;

  &::-webkit-scrollbar {
    width: 8px;
  }

  &::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 4px;
  }

  &::-webkit-scrollbar-thumb {
    background: #888;
    border-radius: 4px;

    &:hover {
      background: #555;
    }
  }

  .results-title {
    font-size: 16px;
    font-weight: 600;
    color: #333;
    margin-bottom: 15px;
    padding-bottom: 10px;
    border-bottom: 2px solid #e0e0e0;
    position: sticky;
    top: 0;
    background: rgba(255, 255, 255, 0.98);
    z-index: 10;
  }

  .results-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .result-item {
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid #e0e0e0;

    &.status-success {
      border-color: #4caf50;
      background: #e8f5e9;
    }

    &.status-failed {
      border-color: #f44336;
      background: #ffebee;
    }

    .result-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 12px 15px;
      background: rgba(255, 255, 255, 0.5);

      .result-name {
        font-size: 14px;
        font-weight: 600;
        color: #333;
        display: flex;
        align-items: center;
        gap: 8px;

        .result-icon {
          font-size: 18px;
        }

        .required-badge {
          font-size: 10px;
          padding: 2px 6px;
          background: #ff5722;
          color: white;
          border-radius: 4px;
          font-weight: 500;
        }

        .optional-badge {
          font-size: 10px;
          padding: 2px 6px;
          background: #9e9e9e;
          color: white;
          border-radius: 4px;
          font-weight: 500;
        }
      }

      .result-status {
        font-size: 13px;
        font-weight: 600;

        .status-success {
          color: #4caf50;
        }

        .status-failed {
          color: #f44336;
        }
      }
    }

    .result-details {
      padding: 12px 15px;
      border-top: 1px solid rgba(0, 0, 0, 0.05);

      .detail-row {
        display: flex;
        margin-bottom: 8px;
        font-size: 13px;

        &:last-child {
          margin-bottom: 0;
        }

        .detail-label {
          font-weight: 500;
          color: #666;
          min-width: 60px;
          margin-right: 10px;
        }

        .detail-value {
          color: #333;
          flex: 1;
        }
      }

      .detail-fix {
        margin-top: 10px;
        flex-direction: column;

        .fix-label {
          font-weight: 500;
          color: #666;
          font-size: 12px;
          margin-bottom: 6px;
        }

        .fix-command {
          background: #263238;
          color: #aed581;
          padding: 10px;
          border-radius: 4px;
          cursor: pointer;
          position: relative;
          transition: all 0.2s ease;

          &:hover {
            background: #37474f;

            .copy-hint {
              opacity: 1;
            }
          }

          code {
            font-family: 'JetBrains Mono', monospace;
            font-size: 12px;
            display: block;
            line-height: 1.5;
          }

          .copy-hint {
            position: absolute;
            right: 10px;
            top: 50%;
            transform: translateY(-50%);
            font-size: 11px;
            color: #fff;
            opacity: 0;
            transition: opacity 0.2s ease;
          }
        }
      }
    }
  }
}

.actions {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  flex-shrink: 0;

  .btn {
    flex: 1;
    padding: 12px 20px;
    border: none;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: all 0.2s ease;

    .btn-icon {
      font-size: 16px;
    }

    &.btn-retry {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
      }
    }

    &.btn-skip {
      background: #f5f5f5;
      color: #666;

      &:hover {
        background: #e0e0e0;
      }
    }
  }
}

.warning {
  padding: 15px;
  background: #fff3e0;
  border-radius: 8px;
  border-left: 4px solid #ff9800;
  flex-shrink: 0;

  .warning-text {
    font-size: 13px;
    color: #e65100;
    line-height: 1.5;
  }
}
</style>
