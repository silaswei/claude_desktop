package models

import "time"

// DetectionResult 环境检测结果
type DetectionResult struct {
	Name       string    `json:"name"`       // 检测项名称
	Status     string    `json:"status"`     // pending/success/failed
	Version    string    `json:"version"`    // 版本信息
	Message    string    `json:"message"`    // 提示信息
	FixCommand string    `json:"fixCommand"` // 修复命令
	Required   bool      `json:"required"`   // 是否必需
	Timestamp  time.Time `json:"timestamp"`  // 检测时间
}

// EnvironmentInfo 环境信息摘要
type EnvironmentInfo struct {
	Status              string           `json:"status"`              // overall status
	TotalRequired       int              `json:"totalRequired"`       // 总必需项数量
	TotalRequiredPassed int              `json:"totalRequiredPassed"` // 已通过的必需项数量
	Results             []DetectionResult `json:"results"`             // 所有检测结果
	LastCheck           time.Time        `json:"lastCheck"`           // 最后检测时间
}

// EnvironmentConfig 环境配置
type EnvironmentConfig struct {
	NodeMinVersion       string  `json:"nodeMinVersion"`       // Node.js 最低版本
	ClaudeMinVersion     string  `json:"claudeMinVersion"`     // Claude CLI 最低版本
	NetworkTimeout       int     `json:"networkTimeout"`       // 网络检测超时时间（秒）
	NetworkRetryCount    int     `json:"networkRetryCount"`    // 网络检测重试次数
	CacheExpiry          int     `json:"cacheExpiry"`          // 缓存过期时间（小时）
	EnableCache          bool    `json:"enableCache"`          // 是否启用缓存
	SkipOptionalCheck    bool    `json:"skipOptionalCheck"`    // 是否跳过可选检测
}
