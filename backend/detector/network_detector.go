package detector

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"claude_desktop/backend/models"
)

// NetworkDetector 网络连通性检测器
type NetworkDetector struct {
	*BaseDetector
	timeout    int    // 超时时间（秒）
	retryCount int    // 重试次数
	apiURL     string // API 端点
}

// NewNetworkDetector 创建网络检测器
func NewNetworkDetector(timeout, retryCount int) *NetworkDetector {
	return &NetworkDetector{
		BaseDetector: NewBaseDetector("网络连通性", true),
		timeout:      timeout,
		retryCount:   retryCount,
		apiURL:       "https://api.anthropic.com",
	}
}

// Detect 执行检测
func (d *NetworkDetector) Detect(ctx context.Context) (*models.DetectionResult, error) {
	client := &http.Client{
		Timeout: time.Duration(d.timeout) * time.Second,
	}

	var lastErr error
	for i := 0; i < d.retryCount; i++ {
		// 创建请求
		req, err := http.NewRequestWithContext(ctx, "HEAD", d.apiURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// 等待一秒后重试
			time.Sleep(1 * time.Second)
			continue
		}

		// 关闭响应体
		resp.Body.Close()

		// 检查状态码
		if resp.StatusCode >= 200 && resp.StatusCode < 500 {
			// 网络连通
			return d.CreateSuccessResult(
				"",
				fmt.Sprintf("网络连接正常，API 端点可访问（状态码: %d）", resp.StatusCode),
			), nil
		}

		lastErr = fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	// 所有重试都失败
	return d.CreateFailedResult(
		fmt.Sprintf("无法连接到 Claude API: %v", lastErr),
		d.getFixSuggestion(),
	), nil
}

// getFixSuggestion 获取修复建议
func (d *NetworkDetector) getFixSuggestion() string {
	return `请检查：
1. 网络连接是否正常
2. 是否需要配置代理
3. 防火墙设置
4. DNS 解析是否正常

如需使用代理，请在环境变量中设置：
export HTTP_PROXY=http://proxy.example.com:port
export HTTPS_PROXY=http://proxy.example.com:port`
}
