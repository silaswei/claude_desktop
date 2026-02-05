import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { FileInfo, WorkspaceInfo } from "@/types/workspace";
import {
  WorkspaceOpen,
  WorkspaceClose,
  WorkspaceSelect,
  WorkspaceRemove,
  WorkspaceList,
  WorkspaceListFiles,
  WorkspaceReadFile,
  WorkspaceWriteFile,
  WorkspaceDeleteFile,
  WorkspaceCreateFile,
  WorkspaceGetInfo,
} from "../../wailsjs/go/app/App";

export const useWorkspaceStore = defineStore("workspace", () => {
  // ==================== 状态 ====================

  // 工作区列表
  const workspaces = ref<WorkspaceInfo[]>([]);

  // 当前选中的工作区路径
  const currentPath = ref<string>("");

  // 当前工作区信息
  const workspaceInfo = ref<WorkspaceInfo | null>(null);

  // 当前工作区的文件列表
  const files = ref<FileInfo[]>([]);

  // 加载状态
  const loading = ref(false);

  // 错误信息
  const error = ref<string | null>(null);

  // ==================== 计算属性 ====================

  // 是否已打开工作区
  const isOpen = computed(() => currentPath.value !== "");

  // 当前工作区名称
  const workspaceName = computed(() => {
    return workspaceInfo.value?.name || "";
  });

  // 当前工作区对象
  const currentWorkspace = computed(() => {
    return workspaces.value.find((ws) => ws.path === currentPath.value) || null;
  });

  // ==================== 方法 ====================

  /**
   * 加载工作区列表
   */
  async function loadWorkspaces(): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const result = await WorkspaceList();
      workspaces.value = result || [];
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 打开工作区（如果不存在则创建新的）
   */
  async function openFolder(path: string): Promise<WorkspaceInfo> {
    loading.value = true;
    error.value = null;

    try {
      // 打开工作区
      const wsInfo = await WorkspaceOpen(path);

      // 更新当前工作区
      currentPath.value = path;
      workspaceInfo.value = wsInfo;

      // 重新加载工作区列表
      await loadWorkspaces();

      // 加载当前工作区的文件
      await loadFiles();

      return wsInfo;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 选择工作区
   */
  async function selectWorkspace(path: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceSelect(path);
      currentPath.value = path;

      // 获取工作区信息
      const info = await WorkspaceGetInfo();
      workspaceInfo.value = info;

      // 加载文件列表
      await loadFiles();

      // 更新工作区列表的打开状态
      workspaces.value.forEach((ws) => {
        ws.isOpen = ws.path === path;
      });
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 移除工作区
   */
  async function removeWorkspace(path: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceRemove(path);

      // 从列表中移除
      workspaces.value = workspaces.value.filter((ws) => ws.path !== path);

      // 如果移除的是当前工作区，清空当前状态
      if (currentPath.value === path) {
        currentPath.value = "";
        workspaceInfo.value = null;
        files.value = [];
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 关闭当前工作区
   */
  async function closeFolder(): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceClose();
      currentPath.value = "";
      workspaceInfo.value = null;
      files.value = [];

      // 更新工作区列表的打开状态
      workspaces.value.forEach((ws) => {
        ws.isOpen = false;
      });
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 加载文件列表
   */
  async function loadFiles(): Promise<void> {
    if (!currentPath.value) {
      files.value = [];
      return;
    }

    loading.value = true;
    error.value = null;

    try {
      const result = await WorkspaceListFiles();
      files.value = result || [];
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 读取文件内容
   */
  async function readFile(path: string): Promise<string> {
    loading.value = true;
    error.value = null;

    try {
      const content = await WorkspaceReadFile(path);
      return content;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 写入文件
   */
  async function writeFile(path: string, content: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceWriteFile(path, content);
      await loadFiles(); // 刷新文件列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 删除文件
   */
  async function deleteFile(path: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceDeleteFile(path);
      await loadFiles(); // 刷新文件列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 创建文件
   */
  async function createFile(path: string, content: string): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      await WorkspaceCreateFile(path, content);
      await loadFiles(); // 刷新文件列表
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 刷新工作区信息
   */
  async function refreshInfo(): Promise<void> {
    loading.value = true;
    error.value = null;

    try {
      const info = await WorkspaceGetInfo();
      workspaceInfo.value = info;
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err);
      throw err;
    } finally {
      loading.value = false;
    }
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
    workspaces,
    currentPath,
    workspaceInfo,
    files,
    loading,
    error,

    // 计算属性
    isOpen,
    workspaceName,
    currentWorkspace,

    // 方法
    loadWorkspaces,
    openFolder,
    selectWorkspace,
    removeWorkspace,
    closeFolder,
    loadFiles,
    readFile,
    writeFile,
    deleteFile,
    createFile,
    refreshInfo,
    clearError,
  };
});
