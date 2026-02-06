<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted, watch, nextTick } from "vue";
import { useWorkspaceStore } from "@/stores/workspace";
import { useEnvStore } from "@/stores/env";
import type { FileInfo, WorkspaceInfo } from "@/types/workspace";
import {
  DialogOpenDirectory,
  ConversationCreate,
  ConversationSendWithEvents,
  ConversationGetByProjectPath,
  WorkspaceSetActiveConversation,
  WorkspaceGetActiveConversation,
  WorkspaceDeleteFile,
  WorkspaceRenameFile,
  SystemOpenFile,
  SystemOpenTerminal,
  SystemRevealInFinder,
  LogFrontend,
  SystemOpenClaudeTerminal,
} from "../../wailsjs/go/app/App";
import { EventsOn, EventsOff } from "../../wailsjs/runtime";

const workspaceStore = useWorkspaceStore();
const envStore = useEnvStore();

// 消息列表容器引用
const messageListRef = ref<HTMLElement | null>(null);

// 输入框引用
const messageInputRef = ref<HTMLTextAreaElement | null>(null);

// UI 状态
const showSidebar = ref(true);
const sidebarWidth = ref(280); // 侧边栏宽度
const isResizing = ref(false); // 是否正在调整宽度
const selectedWorkspace = ref<WorkspaceInfo | null>(null);
const messageInput = ref("");
const isSending = ref(false);
const isThinking = ref(false); // 思考状态
const conversationId = ref("");
const messages = ref<
  Array<{
    id: string;
    role: "user" | "assistant";
    content: string;
    timestamp: string;
  }>
>([]);

// 流式输出：当前正在更新的消息对象
let streamingMessage: {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: string;
} | null = null;

// 思考中消息
let thinkingMessageId: string | null = null;

// 文件树展开状态
const expandedFolders = ref<Set<string>>(new Set());
const currentFolderFilter = ref<string | null>(null);

// 右键菜单状态
const contextMenu = ref({
  visible: false,
  x: 0,
  y: 0,
  file: null as FileInfo | null,
});

// 计算属性：过滤后的文件列表
const filteredFiles = computed(() => {
  const allFiles = workspaceStore.files || [];

  if (!currentFolderFilter.value || !selectedWorkspace.value) {
    // 显示根目录下的文件
    return allFiles.filter((file) => {
      // 只显示第一层文件（不包含 / 的）
      return !file.path.includes("/");
    });
  }

  // 计算当前过滤器的相对路径
  const filterRelative = currentFolderFilter.value.replace(
    selectedWorkspace.value.path + "/",
    ""
  );

  // 显示特定文件夹下的直接子文件和子文件夹
  return allFiles.filter((file) => {
    // 文件必须以当前过滤路径开头
    if (!file.path.startsWith(filterRelative + "/")) {
      return false;
    }

    // 获取相对路径的剩余部分
    const remainingPath = file.path.substring(filterRelative.length + 1);

    // 只显示直接子项（剩余部分不包含 /）
    return !remainingPath.includes("/");
  });
});

// 计算面包屑路径
const breadcrumbPath = computed(() => {
  if (!currentFolderFilter.value || !selectedWorkspace.value) {
    return [];
  }

  const relativePath = currentFolderFilter.value.replace(
    selectedWorkspace.value.path + "/",
    ""
  );
  const parts = relativePath.split("/");
  const breadcrumbs = [];
  let currentPath = selectedWorkspace.value.path;

  for (const part of parts) {
    currentPath += "/" + part;
    breadcrumbs.push({
      name: part,
      path: currentPath,
    });
  }

  return breadcrumbs;
});

// 组件挂载时加载数据
onMounted(async () => {
  try {
    // 加载工作区列表
    await workspaceStore.loadWorkspaces();

    // 如果有当前工作区，选中它
    if (workspaceStore.currentPath) {
      const current = workspaceStore.workspaces.find(
        (ws) => ws.path === workspaceStore.currentPath
      );
      if (current) {
        selectedWorkspace.value = current;
      }
    }

    // 监听 Claude 响应事件
    LogFrontend("注册事件监听器...");
    EventsOn("claude:response", handleClaudeResponse);
    EventsOn("claude:thinking", handleClaudeThinking);
    EventsOn("claude:complete", handleClaudeComplete);
    EventsOn("claude:error", handleClaudeError);
    LogFrontend("事件监听器注册完成");

    // 点击页面其他地方关闭右键菜单
    window.addEventListener("click", closeContextMenu);
  } catch (error) {
    console.error("加载数据失败:", error);
  }
});

// 组件卸载时清理
onUnmounted(() => {
  EventsOff("claude:response");
  EventsOff("claude:thinking");
  EventsOff("claude:complete");
  EventsOff("claude:error");
  window.removeEventListener("click", closeContextMenu);
});

// 处理 Claude 响应（真正的流式输出）
function handleClaudeResponse(data: any) {
  // 快速提取内容
  let content = "";
  if (typeof data === "string") {
    content = data;
  } else if (data?.content) {
    content = data.content;
  }

  if (!content) return;

  // 移除思考中消息（第一次收到任何内容时）
  if (thinkingMessageId) {
    const thinkingIndex = messages.value.findIndex(
      (m) => m.id === thinkingMessageId
    );
    if (thinkingIndex !== -1) {
      messages.value.splice(thinkingIndex, 1);
    }
    thinkingMessageId = null;
    isThinking.value = false;
  }

  // 查找或创建流式消息对象
  if (!streamingMessage) {
    const lastMessage = messages.value[messages.value.length - 1];
    if (lastMessage?.role === "assistant") {
      streamingMessage = lastMessage;
    } else {
      streamingMessage = {
        id: `msg-${Date.now()}`,
        role: "assistant" as const,
        content: "",
        timestamp: new Date().toISOString(),
      };
      messages.value.push(streamingMessage);
    }
  }

  // 立即追加内容并创建新对象替换
  streamingMessage.content += content;

  // 创建新对象来强制 Vue 重新渲染（解构创建新对象）
  const newMessage = { ...streamingMessage };
  const index = messages.value.findIndex(m => m.id === streamingMessage!.id);
  if (index !== -1) {
    messages.value.splice(index, 1, newMessage);
    streamingMessage = newMessage; // 更新引用
  }

  // 如果在底部，就滚动
  if (isNearBottom()) {
    nextTick(() => {
      scrollToBottom();
    });
  }
}

// 处理 Claude 开始思考事件
function handleClaudeThinking() {
  isThinking.value = true;

  // 创建思考中消息
  thinkingMessageId = `msg-thinking-${Date.now()}`;
  messages.value.push({
    id: thinkingMessageId,
    role: "assistant" as const,
    content: "思考中",
    timestamp: new Date().toISOString(),
  });

  // 如果当前在底部，思考中消息显示后滚动到底部
  if (isNearBottom()) {
    nextTick(() => {
      scrollToBottom();
    });
  }
}

// 处理 Claude 完成事件
function handleClaudeComplete(data: any) {
  // 如果思考动画还在，说明没有收到任何实际内容
  if (thinkingMessageId) {
    const thinkingIndex = messages.value.findIndex(
      (m) => m.id === thinkingMessageId
    );
    if (thinkingIndex !== -1) {
      messages.value.splice(thinkingIndex, 1);
    }
    thinkingMessageId = null;

    messages.value.push({
      id: `msg-error-${Date.now()}`,
      role: "assistant" as const,
      content: "抱歉，没有收到任何响应。请检查 Claude CLI 是否正确配置。",
      timestamp: new Date().toISOString(),
    });
  }

  isThinking.value = false;
  streamingMessage = null;
}

// 处理 Claude 错误事件
function handleClaudeError(data: any) {
  console.error("收到错误事件:", data);
  const errorMsg = data?.error || "未知错误";

  // 移除思考中消息
  if (thinkingMessageId) {
    const thinkingIndex = messages.value.findIndex(
      (m) => m.id === thinkingMessageId
    );
    if (thinkingIndex !== -1) {
      messages.value.splice(thinkingIndex, 1);
    }
    thinkingMessageId = null;
  }

  // 添加错误消息
  messages.value.push({
    id: `msg-error-${Date.now()}`,
    role: "assistant" as const,
    content: `发生错误: ${errorMsg}`,
    timestamp: new Date().toISOString(),
  });

  isThinking.value = false;
}

// 停止思考（占位函数，实际需要后端支持）
async function handleStopThinking() {
  // TODO: 实现停止功能，需要后端添加对应的 API
  isThinking.value = false;
  isSending.value = false;
  streamingMessage = null;
}

// 加载工作区的历史对话
async function loadWorkspaceConversation(projectPath: string) {
  try {
    // 首先尝试获取存储的活跃会话ID
    const storedConvID = await WorkspaceGetActiveConversation();
    console.log("存储的会话ID:", storedConvID);

    let conv = null;
    if (storedConvID) {
      // 如果有存储的会话ID，直接使用该会话
      // 这里我们需要添加一个新的API来通过ID获取会话
      // 暂时先使用 GetByProjectPath 作为备选方案
    }

    // 如果没有存储的会话ID或加载失败，通过项目路径查找最新会话
    if (!conv) {
      conv = await ConversationGetByProjectPath(projectPath);
    }

    if (conv && conv.messages && conv.messages.length > 0) {
      // 转换消息格式
      messages.value = conv.messages.map((msg: any) => ({
        id: msg.id || `msg-${Date.now()}-${Math.random()}`,
        role: msg.role,
        content: msg.content,
        timestamp: msg.timestamp || new Date().toISOString(),
      }));
      conversationId.value = conv.id;
      console.log(
        "加载历史对话成功，消息数:",
        messages.value.length,
        "会话ID:",
        conv.id
      );

      // 确保活跃会话ID已设置
      await WorkspaceSetActiveConversation(conv.id);

      // 加载历史对话后滚动到底部（等待 DOM 更新）
      await nextTick();
      await nextTick(); // 双重 nextTick 确保 DOM 完全渲染
      scrollToBottom();
    } else {
      // 没有历史对话，清空消息
      messages.value = [];
      conversationId.value = "";
    }
  } catch (error) {
    // 没有历史对话或其他错误，清空消息
    console.log("没有历史对话:", error);
    messages.value = [];
    conversationId.value = "";
  }
}

// 滚动到底部
function scrollToBottom() {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight;
  }
}

// 检查是否在底部（50px以内）
function isNearBottom(): boolean {
  if (!messageListRef.value) return false;
  const el = messageListRef.value;
  const threshold = 50;
  return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
}

// 打开文件夹
async function handleOpenFolder() {
  try {
    const path = await DialogOpenDirectory();
    if (path) {
      const wsInfo = await workspaceStore.openFolder(path);
      selectedWorkspace.value = wsInfo;

      // 尝试加载历史对话
      await loadWorkspaceConversation(path);

      // 如果没有历史对话，创建新会话
      if (!conversationId.value) {
        const conv = await ConversationCreate(wsInfo.name, path);
        conversationId.value = conv.id;
        messages.value = [];
      }
    }
  } catch (error) {
    console.error("打开文件夹失败:", error);
    alert("打开文件夹失败: " + error);
  }
}

// 打开 Claude 终端
async function handleOpenClaudeTerminal() {
  if (!selectedWorkspace.value) {
    alert("请先选择工作区");
    return;
  }

  try {
    await SystemOpenClaudeTerminal();
  } catch (error) {
    console.error("打开 Claude 终端失败:", error);
    alert("打开 Claude 终端失败: " + error);
  }
}

// 选择工作区
async function handleSelectWorkspace(ws: WorkspaceInfo) {
  try {
    await workspaceStore.selectWorkspace(ws.path);
    selectedWorkspace.value = ws;

    // 尝试加载历史对话
    await loadWorkspaceConversation(ws.path);

    // 如果没有历史对话，创建新会话
    if (!conversationId.value) {
      const conv = await ConversationCreate(ws.name, ws.path);
      conversationId.value = conv.id;
      messages.value = [];
    }
  } catch (error) {
    console.error("选择工作区失败:", error);
  }
}

// 移除工作区
async function handleRemoveWorkspace(path: string, event: Event) {
  event.stopPropagation(); // 阻止事件冒泡

  if (confirm("确定要移除这个工作区吗？")) {
    try {
      await workspaceStore.removeWorkspace(path);

      // 如果移除的是当前工作区
      if (selectedWorkspace.value?.path === path) {
        selectedWorkspace.value = null;
        conversationId.value = "";
        messages.value = [];
      }
    } catch (error) {
      console.error("移除工作区失败:", error);
    }
  }
}

// 发送消息
async function handleSendMessage() {
  if (!messageInput.value.trim() || isSending.value || isThinking.value) {
    return;
  }

  if (!selectedWorkspace.value) {
    alert("请先选择一个工作区");
    return;
  }

  // 保存消息内容
  const messageContent = messageInput.value;

  // 立即清空输入框
  messageInput.value = "";

  isSending.value = true;

  try {
    // 创建会话（如果还没有）
    if (!conversationId.value) {
      const conv = await ConversationCreate(
        selectedWorkspace.value.name,
        selectedWorkspace.value.path
      );
      conversationId.value = conv.id;
      // 保存活跃会话ID到工作区
      await WorkspaceSetActiveConversation(conv.id);
    }

    // 立即添加用户消息到界面（立即显示）
    const userMessage = {
      id: `msg-${Date.now()}`,
      role: "user" as const,
      content: messageContent,
      timestamp: new Date().toISOString(),
    };
    messages.value.push(userMessage);

    // 如果在底部，立即滚动
    if (isNearBottom()) {
      nextTick(() => {
        scrollToBottom();
      });
    }

    // 发送到后端（使用事件流式接收响应）
    await ConversationSendWithEvents(conversationId.value, messageContent);

    // 发送完成后重新加载对话以确保同步
    await loadWorkspaceConversation(selectedWorkspace.value.path);

    // 重新加载后滚动到底部
    nextTick(() => {
      scrollToBottom();
    });
  } catch (error) {
    console.error("发送消息失败:", error);
    alert("发送消息失败: " + error);
  } finally {
    // 重置流式状态
    streamingMessage = null;

    isSending.value = false;
    isThinking.value = false;
  }
}

// 切换侧边栏显示
function toggleSidebar() {
  showSidebar.value = !showSidebar.value;
}

// 开始拖动调整宽度
function startResizing(event: MouseEvent) {
  event.preventDefault();
  isResizing.value = true;

  // 添加全局鼠标移动和释放监听
  document.addEventListener("mousemove", handleResizing);
  document.addEventListener("mouseup", stopResizing);
}

// 拖动中
function handleResizing(event: MouseEvent) {
  if (!isResizing.value) return;

  const newWidth = event.clientX;

  // 限制最小和最大宽度
  if (newWidth >= 200 && newWidth <= 600) {
    sidebarWidth.value = newWidth;
  }
}

// 停止拖动
function stopResizing() {
  isResizing.value = false;
  document.removeEventListener("mousemove", handleResizing);
  document.removeEventListener("mouseup", stopResizing);
}

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

// 处理文件点击
function handleFileClick(file: FileInfo) {
  console.log("点击文件:", file);
  // 如果是文件夹，进入该文件夹
  if (file.type === "directory") {
    enterFolder(file.path);
  }
}

// 进入文件夹
function enterFolder(folderPath: string) {
  currentFolderFilter.value = folderPath;
  expandedFolders.value.add(folderPath);
  console.log("进入文件夹:", folderPath);
}

// 返回上级文件夹
function navigateToFolder(folderPath: string) {
  currentFolderFilter.value = folderPath;
  console.log("导航到文件夹:", folderPath);
}

// 返回根目录
function navigateToRoot() {
  currentFolderFilter.value = null;
  console.log("返回根目录");
}

// 处理右键菜单
function handleContextMenu(event: MouseEvent, file: FileInfo) {
  event.preventDefault();

  // 获取窗口尺寸
  const windowWidth = window.innerWidth;
  const windowHeight = window.innerHeight;

  // 菜单尺寸（估算）
  const menuWidth = 200;
  const menuHeight = 300;

  // 计算菜单位置，确保不会被遮挡
  let x = event.clientX;
  let y = event.clientY;

  // 如果右边空间不足，向左显示
  if (x + menuWidth > windowWidth) {
    x = windowWidth - menuWidth - 10;
  }

  // 如果下边空间不足，向上显示
  if (y + menuHeight > windowHeight) {
    y = windowHeight - menuHeight - 10;
  }

  // 确保不会超出左边界和上边界
  x = Math.max(10, x);
  y = Math.max(10, y);

  contextMenu.value = {
    visible: true,
    x,
    y,
    file,
  };
}

// 关闭右键菜单
function closeContextMenu() {
  contextMenu.value.visible = false;
}

// 发送文件路径到输入框
async function sendFilePathToInput(file: FileInfo) {
  if (!selectedWorkspace.value) {
    alert("请先选择工作区");
    return;
  }

  // 计算相对路径
  const relativePath = file.path.replace(
    selectedWorkspace.value.path + "/",
    ""
  );
  const pathMessage = `@${relativePath} `; // 路径后加空格

  // 添加到输入框
  messageInput.value += (messageInput.value ? "\n" : "") + pathMessage;

  // 关闭右键菜单
  closeContextMenu();

  // 激活输入框并聚焦
  await nextTick();
  if (messageInputRef.value) {
    messageInputRef.value.focus();
    // 将光标移动到文本末尾
    messageInputRef.value.setSelectionRange(
      messageInput.value.length,
      messageInput.value.length
    );
  }
}

// 打开文件/文件夹
async function openFile(file: FileInfo) {
  try {
    if (file.type === "directory") {
      // 如果是目录，进入该目录
      enterFolder(file.path);
    } else {
      // 如果是文件，使用系统默认应用打开
      await SystemOpenFile(file.path);
    }
  } catch (error) {
    console.error("打开失败:", error);
    alert("打开失败: " + error);
  }
  closeContextMenu();
}

// 重命名文件
async function renameFile(file: FileInfo) {
  const newName = prompt("请输入新名称:", file.name);
  if (!newName || newName === file.name) {
    closeContextMenu();
    return;
  }

  // 验证新名称
  if (newName.includes("/") || newName.includes("\\")) {
    alert("文件名不能包含斜杠");
    closeContextMenu();
    return;
  }

  try {
    // 计算新路径
    const lastSlashIndex = file.path.lastIndexOf("/");
    let newPath: string;

    if (lastSlashIndex === -1) {
      // 根目录下的文件
      newPath = newName;
    } else {
      // 子目录下的文件
      const dir = file.path.substring(0, lastSlashIndex);
      newPath = `${dir}/${newName}`;
    }

    console.log(`重命名: ${file.path} -> ${newPath}`);

    await WorkspaceRenameFile(file.path, newPath);

    console.log("重命名成功");

    // 重新加载文件树
    await workspaceStore.loadFiles();

    alert("重命名成功！");
  } catch (error) {
    console.error("重命名失败:", error);
    alert("重命名失败: " + error);
  }
  closeContextMenu();
}

// 删除文件
async function deleteFile(file: FileInfo) {
  const typeText = file.type === "directory" ? "文件夹" : "文件";
  if (
    !confirm(`确定要删除${typeText} "${file.name}" 吗？\n\n此操作不可恢复！`)
  ) {
    closeContextMenu();
    return;
  }

  try {
    console.log(`删除: ${file.path} (${file.type})`);

    await WorkspaceDeleteFile(file.path);

    console.log("删除成功");

    // 重新加载文件树
    await workspaceStore.loadFiles();

    alert("删除成功！");
  } catch (error) {
    console.error("删除失败:", error);
    alert("删除失败: " + error);
  }
  closeContextMenu();
}

// 在终端中打开
async function openInTerminal(file: FileInfo) {
  try {
    await SystemOpenTerminal(file.path);
  } catch (error) {
    console.error("打开终端失败:", error);
    alert("打开终端失败: " + error);
  }
  closeContextMenu();
}

// 在Finder中显示
async function revealInFinder(file: FileInfo) {
  try {
    await SystemRevealInFinder(file.path);
  } catch (error) {
    console.error("在Finder中显示失败:", error);
    alert("在Finder中显示失败: " + error);
  }
  closeContextMenu();
}

// 获取文件图标
function getFileIcon(file: FileInfo): string {
  if (file.type === "directory") {
    return "📁";
  }

  const ext = file.name.split(".").pop()?.toLowerCase() || "";

  // 代码文件
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

  // 样式文件
  if (["css", "scss", "sass", "less"].includes(ext)) {
    return "🎨";
  }

  // 配置文件
  if (["json", "yaml", "yml", "toml", "ini", "conf", "config"].includes(ext)) {
    return "⚙️";
  }

  // Markdown
  if (["md", "markdown"].includes(ext)) {
    return "📝";
  }

  // 图片
  if (["png", "jpg", "jpeg", "gif", "svg", "ico", "webp"].includes(ext)) {
    return "🖼️";
  }

  // 音频
  if (["mp3", "wav", "ogg", "flac"].includes(ext)) {
    return "🎵";
  }

  // 视频
  if (["mp4", "avi", "mov", "wmv", "flv"].includes(ext)) {
    return "🎬";
  }

  // 压缩文件
  if (["zip", "rar", "7z", "tar", "gz"].includes(ext)) {
    return "📦";
  }

  // 文本文件
  if (["txt", "log"].includes(ext)) {
    return "📃";
  }

  // 默认文件图标
  return "📄";
}
</script>

<template>
  <div class="main-view">
    <!-- 顶部栏 -->
    <div class="header">
      <div class="header-left">
        <!-- 左侧留空 -->
      </div>

      <div class="header-center">
        <h1 class="app-title">Claude Desktop</h1>
      </div>

      <div class="header-right">
        <!-- 环境状态 -->
        <div class="env-status">
          <span class="status-label">环境:</span>
          <span
            class="status-badge"
            :class="{
              success: envStore.allPassed,
              failed: envStore.hasRequiredFailed,
              partial: !envStore.allPassed && !envStore.hasRequiredFailed,
            }"
          >
            {{ envStore.allPassed ? "正常" : "异常" }}
          </span>
        </div>

        <!-- 打开工作区按钮 -->
        <button
          class="workspace-btn"
          @click="handleOpenFolder"
          title="打开工作区"
        >
          📁 打开文件夹
        </button>

        <!-- 打开 Claude 终端按钮 -->
        <button
          v-if="selectedWorkspace"
          class="workspace-btn"
          @click="handleOpenClaudeTerminal"
          title="在项目目录中打开 Claude 终端"
        >
          💬 打开 Claude
        </button>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="content">
      <!-- 左侧边栏展开指示器（小凸起） -->
      <div
        v-if="!showSidebar"
        class="sidebar-tab"
        @click="toggleSidebar"
        title="展开侧边栏"
      >
        <span class="tab-icon">▶</span>
      </div>

      <!-- 左侧边栏 -->
      <div
        v-if="showSidebar"
        class="sidebar"
        :style="{ width: sidebarWidth + 'px' }"
      >
        <!-- 工作区列表 -->
        <div class="section workspace-list-section">
          <div class="section-header">
            <h3>工作区列表</h3>
            <button
              class="collapse-btn"
              @click="toggleSidebar"
              title="收起侧边栏"
            >
              ◀
            </button>
          </div>
          <div class="workspace-list">
            <div
              v-for="ws in workspaceStore.workspaces"
              :key="ws.path"
              class="workspace-item"
              :class="{ active: selectedWorkspace?.path === ws.path }"
              @click="handleSelectWorkspace(ws)"
            >
              <div class="workspace-item-content">
                <div class="workspace-info">
                  <div class="workspace-name">{{ ws.name }}</div>
                  <div class="workspace-meta">
                    <span class="workspace-time">{{
                      formatTime(ws.lastOpened)
                    }}</span>
                    <span v-if="ws.isOpen" class="workspace-status"
                      >● 当前</span
                    >
                  </div>
                </div>
              </div>
              <button
                class="remove-btn"
                @click="(e: MouseEvent) => handleRemoveWorkspace(ws.path, e)"
                title="移除工作区"
              >
                ✕
              </button>
            </div>
            <div
              v-if="workspaceStore.workspaces.length === 0"
              class="empty-state"
            >
              暂无工作区，点击右上角"打开文件夹"添加
            </div>
          </div>
        </div>

        <!-- 文件树 -->
        <div class="section file-tree-section">
          <div class="section-header">
            <h3>文件树</h3>
          </div>

          <!-- 面包屑导航 -->
          <div v-if="currentFolderFilter" class="breadcrumb-nav">
            <span class="breadcrumb-item" @click="navigateToRoot">
              🏠 根目录
            </span>
            <span class="breadcrumb-separator">/</span>
            <template
              v-for="(crumb, index) in breadcrumbPath"
              :key="crumb.path"
            >
              <span
                v-if="index === breadcrumbPath.length - 1"
                class="breadcrumb-item current"
              >
                {{ crumb.name }}
              </span>
              <span
                v-else
                class="breadcrumb-item"
                @click="navigateToFolder(crumb.path)"
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

          <div v-if="workspaceStore.isOpen" class="file-tree">
            <div
              v-for="file in filteredFiles"
              :key="file.path"
              class="file-item"
              :class="`file-type-${file.type}`"
              @click="handleFileClick(file)"
              @dblclick="sendFilePathToInput(file)"
              @contextmenu="handleContextMenu($event, file)"
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

        <!-- 右键菜单 -->
        <div
          v-if="contextMenu.visible"
          class="context-menu"
          :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
          @click.stop
        >
          <div class="context-menu-item" @click="openFile(contextMenu.file!)">
            📂 打开
          </div>
          <div class="context-menu-divider"></div>
          <div
            class="context-menu-item"
            @click="sendFilePathToInput(contextMenu.file!)"
          >
            📎 发送路径到输入框
          </div>
          <div class="context-menu-divider"></div>
          <div class="context-menu-item" @click="renameFile(contextMenu.file!)">
            ✏️ 重命名
          </div>
          <div
            class="context-menu-item danger"
            @click="deleteFile(contextMenu.file!)"
          >
            🗑️ 删除
          </div>
          <div class="context-menu-divider"></div>
          <div
            class="context-menu-item"
            @click="openInTerminal(contextMenu.file!)"
          >
            💻 在终端中打开
          </div>
          <div
            class="context-menu-item"
            @click="revealInFinder(contextMenu.file!)"
          >
            👁️ 在Finder中显示
          </div>
        </div>
      </div>

      <!-- 拖动条 -->
      <div v-if="showSidebar" class="resizer" @mousedown="startResizing"></div>

      <!-- 主对话区 -->
      <div class="main-area">
        <!-- 会话状态指示 -->
        <div v-if="selectedWorkspace" class="session-status">
          <span class="workspace-name">{{ selectedWorkspace.name }}</span>
          <span class="status-indicator" :class="{ active: conversationId }">
            <span class="status-dot"></span>
            <span class="status-text">{{
              conversationId ? "会话中" : "未连接"
            }}</span>
          </span>
        </div>

        <!-- 消息列表 -->
        <div
          v-if="messages.length > 0"
          ref="messageListRef"
          class="message-list"
        >
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

        <!-- 欢迎界面 -->
        <div v-else-if="selectedWorkspace" class="welcome-screen">
          <div class="welcome-icon">💬</div>
          <h2 class="welcome-title">{{ selectedWorkspace.name }}</h2>
          <p class="welcome-desc">当前工作区: {{ selectedWorkspace.path }}</p>
          <div class="welcome-hint">
            <p>💡 在下方输入消息开始与 Claude 对话</p>
          </div>
        </div>

        <!-- 未选择工作区 -->
        <div v-else class="welcome-screen">
          <div class="welcome-icon">👋</div>
          <h2 class="welcome-title">欢迎使用 Claude Desktop</h2>
          <p class="welcome-desc">选择或打开一个工作区开始使用</p>
        </div>

        <!-- 输入区域 -->
        <div v-if="selectedWorkspace" class="input-panel">
          <textarea
            ref="messageInputRef"
            v-model="messageInput"
            class="message-input"
            placeholder="输入消息... (Shift+Enter 换行, Enter 发送)"
            rows="3"
            @keydown.enter.exact.prevent="
              isThinking ? handleStopThinking() : handleSendMessage()
            "
          ></textarea>
          <div class="input-actions">
            <span class="input-hint">{{ messageInput.length }} 字符</span>
            <!-- 发送按钮 / 停止按钮 -->
            <button
              v-if="!isThinking"
              class="send-btn"
              :disabled="!messageInput.trim() || isSending"
              @click="handleSendMessage"
            >
              {{ isSending ? "发送中..." : "发送" }}
            </button>
            <button v-else class="stop-btn-inline" @click="handleStopThinking">
              <span class="stop-icon">⏹</span>
              <span class="stop-text">停止</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.main-view {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  overflow: hidden;
}

// ==================== 顶部栏 ====================
.header {
  background: rgba(255, 255, 255, 0.98);
  padding: 12px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;

  .header-left {
    flex: 1;
    /* 左侧留空 */
  }

  .header-center {
    flex: 0 0 auto;
    display: flex;
    justify-content: center;
    align-items: center;

    .app-title {
      margin: 0;
      font-size: 18px;
      font-weight: 600;
      color: #333;
    }
  }

  .header-right {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 15px;

    .env-status {
      display: flex;
      align-items: center;
      gap: 6px;

      .status-label {
        font-size: 12px;
        color: #666;
      }

      .status-badge {
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 500;

        &.success {
          background: #e8f5e9;
          color: #4caf50;
        }

        &.failed {
          background: #ffebee;
          color: #f44336;
        }

        &.partial {
          background: #fff3e0;
          color: #ff9800;
        }
      }
    }

    .workspace-btn {
      padding: 6px 12px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      border: none;
      border-radius: 6px;
      font-size: 13px;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
      }
    }
  }
}

// ==================== 主内容区 ====================
.content {
  flex: 1;
  display: flex;
  overflow: hidden;
  padding: 10px;
  gap: 6px;
  position: relative;
}

// ==================== 侧边栏展开指示器（小凸起） ====================
.sidebar-tab {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 80px;
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid #e0e0e0;
  border-radius: 0 8px 8px 0;
  border-left: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
  z-index: 100;

  &:hover {
    background: #f5f5f5;
    width: 24px;
  }

  .tab-icon {
    font-size: 12px;
    color: #666;
  }
}

// ==================== 侧边栏 ====================
.sidebar {
  width: 280px; // 默认宽度，会被动态样式覆盖
  background: rgba(255, 255, 255, 0.98);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0; // 防止被压缩

  .section {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-bottom: 1px solid #e0e0e0;

    &:last-child {
      border-bottom: none;
    }

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

      .icon-btn,
      .collapse-btn {
        background: none;
        border: none;
        font-size: 12px;
        cursor: pointer;
        padding: 4px 8px;
        border-radius: 4px;
        transition: background 0.2s;

        &:hover {
          background: #e0e0e0;
        }
      }

      .collapse-btn {
        opacity: 0.6;

        &:hover {
          opacity: 1;
        }
      }
    }

    // 面包屑导航
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

    .workspace-list,
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

        .remove-btn {
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

      .remove-btn {
        position: absolute;
        top: 8px;
        right: 8px;
        background: none;
        border: none;
        font-size: 14px;
        color: #999;
        cursor: pointer;
        opacity: 0;
        transition: opacity 0.2s;
        padding: 4px;
        line-height: 1;

        &:hover {
          color: #f44336;
        }
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

    .empty-state,
    .empty-state-small {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #999;
      font-size: 12px;
      text-align: center;
      padding: 20px;
    }

    .empty-state-small {
      padding: 12px;
    }
  }
}

// ==================== 右键菜单 ====================
.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #d0d0d0;
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25), 0 0 1px rgba(0, 0, 0, 0.3);
  z-index: 9999;
  min-width: 154px;
  max-width: 210px;
  overflow: hidden;

  .context-menu-item {
    padding: 9px 14px;
    cursor: pointer;
    font-size: 12px;
    color: #333;
    transition: background 0.15s;
    user-select: none;
    display: flex;
    align-items: center;
    gap: 8px;

    &:hover {
      background: #f0f0f0;
    }

    &:first-child {
      border-radius: 6px 6px 0 0;
    }

    &:last-child {
      border-radius: 0 0 6px 6px;
    }

    &.danger {
      color: #f44336;

      &:hover {
        background: #ffebee;
      }
    }
  }

  .context-menu-divider {
    height: 1px;
    background: #e0e0e0;
    margin: 4px 0;
  }
}

// ==================== 主对话区 ====================
.main-area {
  flex: 1;
  background: rgba(255, 255, 255, 0.98);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;

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

  .welcome-screen {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #666;
    padding: 40px 20px;

    .welcome-icon {
      font-size: 64px;
      margin-bottom: 20px;
      opacity: 0.5;
    }

    .welcome-title {
      font-size: 24px;
      color: #333;
      margin-bottom: 12px;
    }

    .welcome-desc {
      font-size: 14px;
      margin-bottom: 24px;
    }

    .welcome-hint {
      padding: 12px 20px;
      background: #fff3e0;
      border-radius: 8px;
      border-left: 4px solid #ff9800;

      p {
        margin: 0;
        font-size: 13px;
        color: #e65100;
      }
    }
  }

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
}

// ==================== 拖动条 ====================
.resizer {
  width: 6px;
  background: rgba(0, 0, 0, 0.05);
  cursor: col-resize;
  flex-shrink: 0;
  transition: background 0.2s;
  position: relative;

  &:hover {
    background: rgba(102, 126, 234, 0.5);
  }

  // 添加拖动时的视觉反馈
  &::after {
    content: "";
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 2px;
    height: 40px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 1px;
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
