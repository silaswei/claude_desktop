import { defineStore } from "pinia";
import { ref } from "vue";

// 定义页面类型
type PageType = "launch" | "main" | "settings";

// 定义启动状态类型
type LaunchStateType = "detecting" | "success" | "failed" | "idle";

export const useUiStore = defineStore("ui", () => {
  // ==================== 状态 ====================

  // 当前页面
  const currentPage = ref<PageType>("launch");

  // 启动页状态
  const launchState = ref<LaunchStateType>("idle");

  // ==================== 方法 ====================

  /**
   * 设置当前页面
   */
  function setCurrentPage(page: PageType): void {
    currentPage.value = page;
  }

  /**
   * 设置启动状态
   */
  function setLaunchState(state: LaunchStateType): void {
    launchState.value = state;
  }

  /**
   * 重置状态
   */
  function reset(): void {
    currentPage.value = "launch";
    launchState.value = "idle";
  }

  // ==================== 返回 ====================
  return {
    currentPage,
    launchState,
    setCurrentPage,
    setLaunchState,
    reset,
  };
});
