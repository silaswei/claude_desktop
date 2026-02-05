<script setup lang="ts">
import { computed } from "vue";
import type { FileInfo, WorkspaceInfo } from "@/types/workspace";

interface Props {
  workspace: WorkspaceInfo | null;
  files: FileInfo[];
  currentFolderFilter: string | null;
}

interface Emits {
  (e: "enter-folder", path: string): void;
  (e: "navigate-to-folder", path: string): void;
  (e: "navigate-to-root"): void;
  (e: "file-click", file: FileInfo): void;
  (e: "file-dblclick", file: FileInfo): void;
  (e: "context-menu", event: MouseEvent, file: FileInfo): void;
}

const props = defineProps<Props>();
defineEmits<Emits>();

// 计算面包屑路径
const breadcrumbPath = computed(() => {
  if (!props.currentFolderFilter || !props.workspace) {
    return [];
  }

  const relativePath = props.currentFolderFilter.replace(
    props.workspace.path + "/",
    ""
  );
  const parts = relativePath.split("/");
  const breadcrumbs = [];
  let currentPath = props.workspace.path;

  for (const part of parts) {
    currentPath += "/" + part;
    breadcrumbs.push({
      name: part,
      path: currentPath,
    });
  }

  return breadcrumbs;
});

// 计算过滤后的文件列表
const filteredFiles = computed(() => {
  const allFiles = props.files || [];

  if (!props.currentFolderFilter || !props.workspace) {
    // 显示根目录下的文件
    return allFiles.filter((file) => {
      return !file.path.includes("/");
    });
  }

  // 计算当前过滤器的相对路径
  const filterRelative = props.currentFolderFilter.replace(
    props.workspace.path + "/",
    ""
  );

  // 显示特定文件夹下的直接子文件和子文件夹
  return allFiles.filter((file) => {
    if (!file.path.startsWith(filterRelative + "/")) {
      return false;
    }

    const remainingPath = file.path.substring(filterRelative.length + 1);
    return !remainingPath.includes("/");
  });
});

// 获取文件图标
function getFileIcon(file: FileInfo): string {
  if (file.type === "directory") {
    return "📁";
  }

  const ext = file.name.split(".").pop()?.toLowerCase() || "";

  if (
    [
      "js",
      "ts",
      "jsx",
      "tsx",
      "vue",
      "go",
      "py",
      "java",
      "c",
      "cpp",
      "h",
      "hpp",
      "rs",
      "rb",
      "php",
    ].includes(ext)
  ) {
    return "📄";
  }
  if (["css", "scss", "sass", "less"].includes(ext)) {
    return "🎨";
  }
  if (["json", "yaml", "yml", "toml", "ini", "conf", "config"].includes(ext)) {
    return "⚙️";
  }
  if (["md", "markdown"].includes(ext)) {
    return "📝";
  }
  if (["png", "jpg", "jpeg", "gif", "svg", "ico", "webp"].includes(ext)) {
    return "🖼️";
  }
  if (["mp3", "wav", "ogg", "flac"].includes(ext)) {
    return "🎵";
  }
  if (["mp4", "avi", "mov", "wmv", "flv"].includes(ext)) {
    return "🎬";
  }
  if (["zip", "rar", "7z", "tar", "gz"].includes(ext)) {
    return "📦";
  }
  if (["txt", "log"].includes(ext)) {
    return "📃";
  }

  return "📄";
}
</script>

<template>
  <div class="section file-tree-section">
    <div class="section-header">
      <h3>文件树</h3>
    </div>

    <!-- 面包屑导航 -->
    <div v-if="currentFolderFilter" class="breadcrumb-nav">
      <span class="breadcrumb-item" @click="$emit('navigate-to-root')">
        🏠 根目录
      </span>
      <span class="breadcrumb-separator">/</span>
      <template v-for="(crumb, index) in breadcrumbPath" :key="crumb.path">
        <span
          v-if="index === breadcrumbPath.length - 1"
          class="breadcrumb-item current"
        >
          {{ crumb.name }}
        </span>
        <span
          v-else
          class="breadcrumb-item"
          @click="$emit('navigate-to-folder', crumb.path)"
        >
          {{ crumb.name }}
        </span>
        <span
          v-if="index < breadcrumbPath.length - 1"
          class="breadcrumb-separator"
          >/</span
        >
      </template>
    </div>

    <div v-if="workspace" class="file-tree">
      <div
        v-for="file in filteredFiles"
        :key="file.path"
        class="file-item"
        :class="`file-type-${file.type}`"
        @click="$emit('file-click', file)"
        @dblclick="$emit('file-dblclick', file)"
        @contextmenu="$emit('context-menu', $event, file)"
      >
        <div class="file-item-row">
          <span class="file-icon">
            {{ file.type === "directory" ? "📁" : getFileIcon(file) }}
          </span>
          <div class="file-info">
            <div class="file-name">{{ file.name }}</div>
          </div>
        </div>
      </div>
      <div v-if="filteredFiles.length === 0" class="empty-state-small">
        {{ currentFolderFilter ? "文件夹为空" : "工作区为空" }}
      </div>
    </div>
    <div v-else class="empty-state-small">未选择工作区</div>
  </div>
</template>

<style lang="scss" scoped>
.file-tree-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  .section-header {
    padding: 10px 12px;
    border-bottom: 1px solid #e0e0e0;
    display: flex;
    align-items: center;
    background: #f9f9f9;

    h3 {
      margin: 0;
      font-size: 13px;
      color: #333;
      font-weight: 600;
    }
  }

  .breadcrumb-nav {
    padding: 8px 12px;
    border-bottom: 1px solid #e0e0e0;
    background: #f5f5f5;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
    font-size: 12px;

    .breadcrumb-item {
      cursor: pointer;
      color: #667eea;
      transition: color 0.2s;

      &:hover {
        color: #764ba2;
        text-decoration: underline;
      }

      &.current {
        color: #333;
        cursor: default;
        font-weight: 500;

        &:hover {
          text-decoration: none;
        }
      }
    }

    .breadcrumb-separator {
      color: #999;
      margin: 0 2px;
    }
  }

  .file-tree {
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

  .file-item {
    margin-bottom: 2px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s ease;

    &:hover {
      background: #f0f4ff;
      transform: translateX(2px);
    }

    .file-item-row {
      display: flex;
      align-items: center;
      padding: 6px 8px;
      gap: 8px;
    }

    .file-icon {
      font-size: 14px;
      flex-shrink: 0;
    }

    .file-info {
      flex: 1;
      min-width: 0;

      .file-name {
        font-size: 13px;
        color: #2c3e50;
        font-weight: 500;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    &.file-type-directory {
      .file-name {
        color: #2980b9;
        font-weight: 600;
      }

      &:hover {
        background: #e3f2fd;
      }
    }
  }

  .empty-state-small {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 12px;
    text-align: center;
    padding: 12px;
  }
}
</style>
