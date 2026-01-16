package detector

import (
	"claude_desktop/backend/models"
	"context"
)

// Detector 环境检测器接口
type Detector interface {
	// Detect 执行检测
	Detect(ctx context.Context) (*models.DetectionResult, error)

	// Name 获取检测器名称
	Name() string

	// Required 是否必需
	Required() bool
}

// BaseDetector 基础检测器，提供通用功能
type BaseDetector struct {
	name     string
	required bool
}

// NewBaseDetector 创建基础检测器
func NewBaseDetector(name string, required bool) *BaseDetector {
	return &BaseDetector{
		name:     name,
		required: required,
	}
}

// Name 获取检测器名称
func (d *BaseDetector) Name() string {
	return d.name
}

// Required 是否必需
func (d *BaseDetector) Required() bool {
	return d.required
}

// CreatePendingResult 创建待定状态的检测结果
func (d *BaseDetector) CreatePendingResult() *models.DetectionResult {
	return &models.DetectionResult{
		Name:     d.name,
		Status:   "pending",
		Required: d.required,
	}
}

// CreateSuccessResult 创建成功状态的检测结果
func (d *BaseDetector) CreateSuccessResult(version, message string) *models.DetectionResult {
	return &models.DetectionResult{
		Name:     d.name,
		Status:   "success",
		Version:  version,
		Message:  message,
		Required: d.required,
	}
}

// CreateFailedResult 创建失败状态的检测结果
func (d *BaseDetector) CreateFailedResult(message, fixCommand string) *models.DetectionResult {
	return &models.DetectionResult{
		Name:       d.name,
		Status:     "failed",
		Message:    message,
		FixCommand: fixCommand,
		Required:   d.required,
	}
}
