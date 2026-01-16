package detector

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"claude_desktop/backend/models"
)

// ClaudeDetector Claude Code CLI 检测器
type ClaudeDetector struct {
	*BaseDetector
	minVersion string
}

// NewClaudeDetector 创建 Claude CLI 检测器
func NewClaudeDetector(minVersion string) *ClaudeDetector {
	return &ClaudeDetector{
		BaseDetector: NewBaseDetector("Claude Code CLI", true),
		minVersion:   minVersion,
	}
}

// Detect 执行检测
func (d *ClaudeDetector) Detect(ctx context.Context) (*models.DetectionResult, error) {
	// 检查 claude 命令是否存在
	cmd := exec.CommandContext(ctx, "claude", "--version")
	output, err := cmd.Output()

	if err != nil {
		// Claude Code CLI 未安装
		return d.CreateFailedResult(
			"未检测到 Claude Code CLI",
			d.getFixCommand(),
		), nil
	}

	// 获取版本号
	rawOutput := strings.TrimSpace(string(output))
	fmt.Printf("DEBUG: Raw claude --version output: [%s]\n", rawOutput)

	// 解析版本号，支持多种格式：
	// "2.1.2 (Claude Code)"
	// "claude version 2.1.2"
	// "2.1.2"

	versionStr := rawOutput

	// 移除可能的括号内容（如 " (Claude Code)"）
	if idx := strings.Index(versionStr, " ("); idx != -1 {
		versionStr = versionStr[:idx]
		fmt.Printf("DEBUG: Removed bracket content, versionStr: [%s]\n", versionStr)
	}

	// 按空格分割，取第一部分
	parts := strings.Fields(versionStr)
	if len(parts) >= 1 {
		versionStr = strings.TrimSpace(parts[0])
		fmt.Printf("DEBUG: Extracted first part: [%s]\n", versionStr)
	}

	// 验证版本号格式（应该只包含数字和点）
	if !isValidVersionFormat(versionStr) {
		fmt.Printf("DEBUG: Invalid version format [%s], extracting from raw output\n", versionStr)
		// 如果格式不对，尝试从原输出中提取
		versionStr = extractVersionFromOutput(rawOutput)
		fmt.Printf("DEBUG: Final extracted version: [%s]\n", versionStr)
	}

	// 验证版本是否满足要求（如果有最低版本要求）
	if d.minVersion != "" && !d.checkVersion(versionStr, d.minVersion) {
		// 添加调试信息
		fmt.Printf("DEBUG: Claude version check failed - current: %s, min: %s\n", versionStr, d.minVersion)
		return d.CreateFailedResult(
			fmt.Sprintf("Claude Code CLI 版本过低2: 当前 %s，要求 %s 或更高", versionStr, d.minVersion),
			d.getUpgradeCommand(),
		), nil
	}

	return d.CreateSuccessResult(
		versionStr,
		fmt.Sprintf("Claude Code CLI 版本 %s 检测通过", versionStr),
	), nil
}

// checkVersion 检查版本是否满足要求
func (d *ClaudeDetector) checkVersion(current, min string) bool {
	currentParts := strings.Split(current, ".")
	minParts := strings.Split(min, ".")

	fmt.Printf("DEBUG: Version comparison - current: %v (%d parts), min: %v (%d parts)\n",
		currentParts, len(currentParts), minParts, len(minParts))

	for i := 0; i < len(minParts); i++ {
		if i >= len(currentParts) {
			fmt.Printf("DEBUG: Not enough version parts at index %d\n", i)
			return false
		}

		var curr, minVal int
		fmt.Sscanf(currentParts[i], "%d", &curr)
		fmt.Sscanf(minParts[i], "%d", &minVal)

		fmt.Printf("DEBUG: Comparing part %d: current=%d, min=%d\n", i, curr, minVal)

		if curr < minVal {
			fmt.Printf("DEBUG: Version check failed at part %d: %d < %d\n", i, curr, minVal)
			return false
		}
		if curr > minVal {
			fmt.Printf("DEBUG: Version check passed at part %d: %d > %d\n", i, curr, minVal)
			return true
		}
	}

	fmt.Printf("DEBUG: Version check passed (all parts equal or higher)\n")
	return true
}

// isValidVersionFormat 验证版本号格式是否有效
func isValidVersionFormat(version string) bool {
	if version == "" {
		return false
	}

	// 版本号应该只包含数字和点
	for _, char := range version {
		if char != '.' && (char < '0' || char > '9') {
			return false
		}
	}

	return true
}

// extractVersionFromOutput 从原始输出中提取版本号
func extractVersionFromOutput(output string) string {
	// 移除所有换行符
	output = strings.ReplaceAll(output, "\n", " ")
	output = strings.ReplaceAll(output, "\r", " ")

	// 按空格分割
	words := strings.Fields(output)

	// 查找符合版本号格式的词（如 2.1.2）
	for _, word := range words {
		if isValidVersionFormat(word) {
			return word
		}
	}

	// 如果找不到，返回第一个词
	if len(words) > 0 {
		return words[0]
	}

	return output
}

// getFixCommand 获取安装命令
func (d *ClaudeDetector) getFixCommand() string {
	// Claude Code CLI 通过 npm 安装
	return "npm install -g @anthropic-ai/claude-code"
}

// getUpgradeCommand 获取升级命令
func (d *ClaudeDetector) getUpgradeCommand() string {
	switch runtime.GOOS {
	case "darwin", "linux":
		return "sudo npm update -g @anthropic-ai/claude-code"
	case "windows":
		return "npm update -g @anthropic-ai/claude-code"
	default:
		return "npm update -g @anthropic-ai/claude-code"
	}
}
