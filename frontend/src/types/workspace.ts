// 工作区相关类型定义

// 文件信息
export interface FileInfo {
  path: string;
  name: string;
  type: string;
  size: number;
  icon: string;
  modifiedAt: string;
}

// 工作区信息
export interface WorkspaceInfo {
  path: string;
  name: string;
  isOpen: boolean;
  lastOpened: string;
  activeConversationId: string;
}
