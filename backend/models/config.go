package models

import "encoding/json"

// DefaultEnvironmentConfig 默认环境配置
func DefaultEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		NodeMinVersion:    "18.0.0",
		ClaudeMinVersion:  "1.0.0",  // 支持所有 1.x 和 2.x 版本
		NetworkTimeout:    10,
		NetworkRetryCount: 3,
		CacheExpiry:       24,
		EnableCache:       true,
		SkipOptionalCheck: false,
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(data []byte) (*EnvironmentConfig, error) {
	config := DefaultEnvironmentConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

// SaveConfig 保存配置到文件
func (c *EnvironmentConfig) SaveConfig() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}
