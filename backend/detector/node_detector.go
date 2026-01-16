package detector

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"claude_desktop/backend/models"
)

// NodeDetector Node.js 环境检测器
type NodeDetector struct {
	*BaseDetector
	minVersion string
}

// NewNodeDetector 创建 Node.js 检测器
func NewNodeDetector(minVersion string) *NodeDetector {
	return &NodeDetector{
		BaseDetector: NewBaseDetector("Node.js", true),
		minVersion:   minVersion,
	}
}

// Detect 执行检测
func (d *NodeDetector) Detect(ctx context.Context) (*models.DetectionResult, error) {
	// 检查 node 命令是否存在
	cmd := exec.CommandContext(ctx, "node", "--version")
	output, err := cmd.Output()

	if err != nil {
		// Node.js 未安装
		return d.CreateFailedResult(
			"未检测到 Node.js",
			d.getFixCommand(),
		), nil
	}

	// 获取版本号
	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "v")

	// 验证版本是否满足要求
	if !d.checkVersion(version, d.minVersion) {
		return d.CreateFailedResult(
			fmt.Sprintf("Node.js 版本过低: 当前 %s，要求 %s 或更高", version, d.minVersion),
			d.getUpgradeCommand(version),
		), nil
	}

	return d.CreateSuccessResult(
		version,
		fmt.Sprintf("Node.js 版本 %s 检测通过", version),
	), nil
}

// checkVersion 检查版本是否满足要求（简单比较）
func (d *NodeDetector) checkVersion(current, min string) bool {
	currentParts := strings.Split(current, ".")
	minParts := strings.Split(min, ".")

	for i := 0; i < len(minParts); i++ {
		if i >= len(currentParts) {
			return false
		}

		var curr, minVal int
		fmt.Sscanf(currentParts[i], "%d", &curr)
		fmt.Sscanf(minParts[i], "%d", &minVal)

		if curr < minVal {
			return false
		}
		if curr > minVal {
			return true
		}
	}

	return true
}

// getFixCommand 获取安装命令
func (d *NodeDetector) getFixCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew install node"
	case "windows":
		return "winget install OpenJS.NodeJS.LTS"
	case "linux":
		return "curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"
	default:
		return "请访问 https://nodejs.org/ 下载安装"
	}
}

// getUpgradeCommand 获取升级命令
func (d *NodeDetector) getUpgradeCommand(currentVersion string) string {
	switch runtime.GOOS {
	case "darwin":
		return "brew upgrade node"
	case "windows":
		return "winget upgrade OpenJS.NodeJS.LTS"
	case "linux":
		return "sudo apt-get install --only-upgrade nodejs"
	default:
		return "请访问 https://nodejs.org/ 下载最新版本"
	}
}
