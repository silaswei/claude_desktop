package detector

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"claude_desktop/backend/models"
)

// GitDetector Git 版本控制检测器（可选）
type GitDetector struct {
	*BaseDetector
}

// NewGitDetector 创建 Git 检测器
func NewGitDetector() *GitDetector {
	return &GitDetector{
		BaseDetector: NewBaseDetector("Git", false), // 非必需
	}
}

// Detect 执行检测
func (d *GitDetector) Detect(ctx context.Context) (*models.DetectionResult, error) {
	// 检查 git 命令是否存在
	cmd := exec.CommandContext(ctx, "git", "--version")
	output, err := cmd.Output()

	if err != nil {
		// Git 未安装
		return d.CreateFailedResult(
			"未检测到 Git（建议安装以获得更好的版本控制体验）",
			d.getFixCommand(),
		), nil
	}

	// 获取版本号
	versionStr := strings.TrimSpace(string(output))
	// 输出格式是 "git version X.Y.Z"，提取版本号
	parts := strings.Fields(versionStr)
	if len(parts) >= 3 {
		versionStr = parts[2]
	}

	return d.CreateSuccessResult(
		versionStr,
		fmt.Sprintf("Git 版本 %s 检测通过", versionStr),
	), nil
}

// getFixCommand 获取安装命令
func (d *GitDetector) getFixCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install git"
	case "windows":
		return "winget install Git.Git"
	case "linux":
		return "sudo apt-get install git"
	default:
		return "请访问 https://git-scm.com/ 下载安装"
	}
}
