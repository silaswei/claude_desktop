import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { EnvironmentInfo, DetectionResult } from "@/types";
import {
  EnvDetectAll,
  EnvDetectByName,
  EnvGetStatus,
  EnvClearCache,
} from "../../wailsjs/go/app/App";

export const useEnvStore = defineStore("env", () => {
  // ==================== 状态 ====================
  const envInfo = ref<EnvironmentInfo | null>(null);
  const isDetecting = ref(false);
  const error = ref<string | null>(null);

  // ==================== 计算属性 ====================

  // 所有检测是否通过
  const allPassed = computed<boolean>(() => {
    return envInfo.value?.status === "success";
  });

  // 是否有必需项失败
  const hasRequiredFailed = computed<boolean>(() => {
    if (!envInfo.value?.results) return false;
    return envInfo.value.results.some(
      (r: DetectionResult) => r.required && r.status === "failed"
    );
  });

  // 必需项通过数量
  const requiredPassed = computed<number>(() => {
    if (!envInfo.value) return 0;
    return envInfo.value.totalRequiredPassed;
  });

  // 必需项总数
  const requiredTotal = computed<number>(() => {
    if (!envInfo.value) return 0;
    return envInfo.value.totalRequired;
  });

  // ==================== 方法 ====================

  /**
   * 执行所有环境检测
   */
  async function detectAll(): Promise<EnvironmentInfo> {
    isDetecting.value = true;
    error.value = null;

    try {
      const result = await EnvDetectAll();
      envInfo.value = result;
      return result;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      throw err;
    } finally {
      isDetecting.value = false;
    }
  }

  /**
   * 执行指定名称的检测
   */
  async function detectByName(name: string): Promise<DetectionResult> {
    try {
      const result = await EnvDetectByName(name);

      // 更新对应的检测结果
      if (envInfo.value?.results) {
        const index = envInfo.value.results.findIndex(
          (r: DetectionResult) => r.name === name
        );
        if (index !== -1) {
          envInfo.value.results[index] = result;
          // 重新计算整体状态
          updateEnvironmentStatus();
        }
      }

      return result;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      throw err;
    }
  }

  /**
   * 获取环境状态（从缓存或执行新检测）
   */
  async function getStatus(): Promise<EnvironmentInfo> {
    try {
      const result = await EnvGetStatus();
      envInfo.value = result;
      return result;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      throw err;
    }
  }

  /**
   * 清除缓存
   */
  async function clearCache(): Promise<void> {
    try {
      await EnvClearCache();
      envInfo.value = null;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      error.value = errorMessage;
      throw err;
    }
  }

  /**
   * 更新整体环境状态
   */
  function updateEnvironmentStatus(): void {
    if (!envInfo.value?.results) return;

    const results = envInfo.value.results;
    let totalRequired = 0;
    let totalRequiredPassed = 0;

    for (const result of results) {
      if (result.required) {
        totalRequired++;
        if (result.status === "success") {
          totalRequiredPassed++;
        }
      }
    }

    envInfo.value.totalRequired = totalRequired;
    envInfo.value.totalRequiredPassed = totalRequiredPassed;

    // 确定整体状态
    if (totalRequiredPassed === totalRequired) {
      envInfo.value.status = "success";
    } else if (totalRequiredPassed > 0) {
      envInfo.value.status = "partial";
    } else {
      envInfo.value.status = "failed";
    }
  }

  /**
   * 重置状态
   */
  function reset(): void {
    envInfo.value = null;
    isDetecting.value = false;
    error.value = null;
  }

  // ==================== 返回 ====================
  return {
    // 状态
    envInfo,
    isDetecting,
    error,

    // 计算属性
    allPassed,
    hasRequiredFailed,
    requiredPassed,
    requiredTotal,

    // 方法
    detectAll,
    detectByName,
    getStatus,
    clearCache,
    reset,
  };
});
