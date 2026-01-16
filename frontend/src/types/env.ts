// 环境检测相关类型定义

export interface DetectionResult {
  name: string;          // 检测项名称
  status: string;        // pending/success/failed
  version: string;       // 版本信息
  message: string;       // 提示信息
  fixCommand: string;    // 修复命令
  required: boolean;     // 是否必需
  timestamp: string;     // 检测时间
}

export interface EnvironmentInfo {
  status: string;                // overall status: success/partial/failed
  totalRequired: number;         // 总必需项数量
  totalRequiredPassed: number;   // 已通过的必需项数量
  results: DetectionResult[];    // 所有检测结果
  lastCheck: string;             // 最后检测时间
}

export interface EnvironmentConfig {
  nodeMinVersion: string;        // Node.js 最低版本
  claudeMinVersion: string;      // Claude CLI 最低版本
  networkTimeout: number;        // 网络检测超时时间（秒）
  networkRetryCount: number;     // 网络检测重试次数
  cacheExpiry: number;           // 缓存过期时间（小时）
  enableCache: boolean;          // 是否启用缓存
  skipOptionalCheck: boolean;    // 是否跳过可选检测
}

// 检测状态枚举
export enum DetectionStatus {
  Pending = 'pending',
  Success = 'success',
  Failed = 'failed'
}

// 整体状态枚举
export enum EnvironmentStatus {
  Success = 'success',
  Partial = 'partial',
  Failed = 'failed'
}
