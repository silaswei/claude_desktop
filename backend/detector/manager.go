package detector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"claude_desktop/backend/models"
)

// Manager 环境检测管理器
type Manager struct {
	detectors []Detector
	config    *models.EnvironmentConfig
	cache     *models.EnvironmentInfo
	cachePath string
	mu        sync.RWMutex
}

// NewManager 创建环境检测管理器
func NewManager(config *models.EnvironmentConfig) *Manager {
	// 获取缓存路径
	homeDir, _ := os.UserHomeDir()
	cachePath := filepath.Join(homeDir, ".claude-desktop", "cache", "env_check.json")

	m := &Manager{
		config:    config,
		cachePath: cachePath,
	}

	// 初始化检测器
	m.initDetectors()

	return m
}

// initDetectors 初始化所有检测器
func (m *Manager) initDetectors() {
	m.detectors = []Detector{
		NewNodeDetector(m.config.NodeMinVersion),
		NewNpmDetector(),
		NewClaudeDetector(m.config.ClaudeMinVersion),
		NewNetworkDetector(m.config.NetworkTimeout, m.config.NetworkRetryCount),
		NewGitDetector(),
	}
}

// DetectAll 执行所有环境检测
func (m *Manager) DetectAll(ctx context.Context) (*models.EnvironmentInfo, error) {
	// 如果启用缓存且缓存有效，直接返回缓存结果
	if m.config.EnableCache {
		if cached := m.loadFromCache(); cached != nil {
			return cached, nil
		}
	}

	// 执行检测
	results := make([]models.DetectionResult, len(m.detectors))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并行执行所有检测
	for i, detector := range m.detectors {
		wg.Add(1)
		go func(index int, d Detector) {
			defer wg.Done()

			result, err := d.Detect(ctx)
			if err != nil {
				// 检测出错，创建失败结果
				result = &models.DetectionResult{
					Name:     d.Name(),
					Status:   "failed",
					Message:  fmt.Sprintf("检测出错: %v", err),
					Required: d.Required(),
				}
			}

			mu.Lock()
			results[index] = *result
			mu.Unlock()
		}(i, detector)
	}

	wg.Wait()

	// 构建环境信息
	envInfo := m.buildEnvironmentInfo(results)

	// 保存到缓存
	if m.config.EnableCache {
		m.saveToCache(envInfo)
	}

	// 更新内存缓存
	m.mu.Lock()
	m.cache = envInfo
	m.mu.Unlock()

	return envInfo, nil
}

// DetectByName 执行指定名称的检测
func (m *Manager) DetectByName(ctx context.Context, name string) (*models.DetectionResult, error) {
	for _, detector := range m.detectors {
		if detector.Name() == name {
			result, err := detector.Detect(ctx)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("未找到名为 %s 的检测器", name)
}

// GetStatus 获取环境状态摘要（从缓存或执行新检测）
func (m *Manager) GetStatus(ctx context.Context) (*models.EnvironmentInfo, error) {
	m.mu.RLock()
	if m.cache != nil {
		m.mu.RUnlock()
		return m.cache, nil
	}
	m.mu.RUnlock()

	return m.DetectAll(ctx)
}

// buildEnvironmentInfo 构建环境信息
func (m *Manager) buildEnvironmentInfo(results []models.DetectionResult) *models.EnvironmentInfo {
	totalRequired := 0
	totalRequiredPassed := 0
	allPassed := true

	for _, result := range results {
		if result.Required {
			totalRequired++
			if result.Status == "success" {
				totalRequiredPassed++
			} else {
				allPassed = false
			}
		}
	}

	// 确定整体状态
	var status string
	if allPassed {
		status = "success"
	} else if totalRequiredPassed > 0 {
		status = "partial"
	} else {
		status = "failed"
	}

	return &models.EnvironmentInfo{
		Status:              status,
		TotalRequired:       totalRequired,
		TotalRequiredPassed: totalRequiredPassed,
		Results:             results,
		LastCheck:           time.Now(),
	}
}

// loadFromCache 从缓存加载检测结果
func (m *Manager) loadFromCache() *models.EnvironmentInfo {
	data, err := os.ReadFile(m.cachePath)
	if err != nil {
		return nil
	}

	var envInfo models.EnvironmentInfo
	if err := json.Unmarshal(data, &envInfo); err != nil {
		return nil
	}

	// 检查缓存是否过期
	expiryTime := envInfo.LastCheck.Add(time.Duration(m.config.CacheExpiry) * time.Hour)
	if time.Now().After(expiryTime) {
		return nil
	}

	return &envInfo
}

// saveToCache 保存检测结果到缓存
func (m *Manager) saveToCache(envInfo *models.EnvironmentInfo) error {
	// 确保缓存目录存在
	cacheDir := filepath.Dir(m.cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(envInfo, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.cachePath, data, 0644)
}

// ClearCache 清除缓存
func (m *Manager) ClearCache() error {
	m.mu.Lock()
	m.cache = nil
	m.mu.Unlock()

	if _, err := os.Stat(m.cachePath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(m.cachePath)
}

// GetAllDetectors 获取所有检测器名称
func (m *Manager) GetAllDetectors() []string {
	names := make([]string, len(m.detectors))
	for i, detector := range m.detectors {
		names[i] = detector.Name()
	}
	return names
}
