package detector

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"claude_desktop/backend/models"
)

// NpmDetector npm 包管理器检测器
type NpmDetector struct {
	*BaseDetector
}

// NewNpmDetector 创建 npm 检测器
func NewNpmDetector() *NpmDetector {
	return &NpmDetector{
		BaseDetector: NewBaseDetector("npm", true),
	}
}

// Detect 执行检测
func (d *NpmDetector) Detect(ctx context.Context) (*models.DetectionResult, error) {
	// 检查 npm 命令是否存在
	cmd := exec.CommandContext(ctx, "npm", "--version")
	output, err := cmd.Output()

	if err != nil {
		// npm 未安装
		return d.CreateFailedResult(
			"未检测到 npm 包管理器",
			d.getFixCommand(),
		), nil
	}

	// 获取版本号
	version := strings.TrimSpace(string(output))

	return d.CreateSuccessResult(
		version,
		fmt.Sprintf("npm 版本 %s 检测通过", version),
	), nil
}

// getFixCommand 获取安装命令
func (d *NpmDetector) getFixCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install npm"
	case "windows":
		return "winget install OpenJS.NodeJS.NPM"
	case "linux":
		return "sudo apt-get install npm"
	default:
		return "请访问 https://www.npmjs.com/ 下载安装"
	}
}
