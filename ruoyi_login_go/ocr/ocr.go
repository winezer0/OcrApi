package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ruoyi_login_go/logging"
)

// 全局变量
var (
	debugDirInitOnce sync.Once
)

// InitDebug 初始化 debug 目录（程序启动时调用）
func InitDebug(enabled bool) {
	if enabled {
		debugDirInitOnce.Do(func() {
			debugDir := "captcha_debug"
			// 清理上一次的缓存图片
			os.RemoveAll(debugDir)
			os.MkdirAll(debugDir, 0755)
		})
	}
}

type OcrRequest struct {
	ImageBase64 string `json:"imageBase64"`
}

type OcrResponse struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	Result string `json:"result"`
	Raw    string `json:"raw"`
}

// CheckHealth 检查 OCR API 是否可用
// ocrClient: 用于调用 OCR API 的 HTTP 客户端（可配置代理、TLS 等）
// ocrAPI: 远程 OCR API 地址
func CheckHealth(ocrClient *http.Client, ocrAPI string) error {
	// 发送一个空的 POST 请求来检查 API 是否可用
	testPayload := `{"imageBase64":""}`
	req, err := http.NewRequest("POST", ocrAPI, bytes.NewBuffer([]byte(testPayload)))
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ocrClient.Do(req)
	if err != nil {
		return fmt.Errorf("OCR API 不可达: %w", err)
	}
	defer resp.Body.Close()

	// 2xx 表示服务正常，400 表示 API 可用（只是请求参数错误）
	// 4xx 中的 401/403/404 等表示服务存在但接口不可用，5xx 表示服务端异常
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}
	if resp.StatusCode == 400 {
		return nil
	}

	return fmt.Errorf("OCR API 返回异常状态码: %d", resp.StatusCode)
}

// GetCaptchaImageWithContext 获取验证码图片并返回 base64 编码（支持 context 控制）
// captchaClient: 用于获取验证码图片的 HTTP 客户端（保持 Cookie 会话）
// captchaURL: 验证码接口 URL
// headers: 请求头
func GetCaptchaImageWithContext(ctx context.Context, captchaClient *http.Client, captchaURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", captchaURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建验证码请求失败: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := captchaClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取验证码失败: %w", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取验证码图片失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(imgData), nil
}

// SolveCaptchaWithContext 调用远程 OCR API 识别验证码（支持 context 取消）
// ocrClient: 用于调用 OCR API 的 HTTP 客户端（可配置代理）
// ocrAPI: 远程 OCR API 地址
// imageBase64: 验证码图片的 base64 编码
// debugEnabled: 是否保存验证码图片到 debug 目录
func SolveCaptchaWithContext(ctx context.Context, ocrClient *http.Client, ocrAPI, imageBase64 string, debugEnabled bool) (string, error) {
	ocrReq := OcrRequest{ImageBase64: imageBase64}
	reqBody, err := json.Marshal(ocrReq)
	if err != nil {
		return "", fmt.Errorf("序列化 OCR 请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ocrAPI, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建 OCR 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	ocrResp, err := ocrClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("调用 OCR API 失败: %w", err)
	}
	defer ocrResp.Body.Close()

	ocrBody, err := io.ReadAll(ocrResp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 OCR 响应失败: %w", err)
	}

	var ocrResult OcrResponse
	if err := json.Unmarshal(ocrBody, &ocrResult); err != nil {
		return "", fmt.Errorf("解析 OCR 响应失败: %w", err)
	}

	if ocrResult.Code == 200 && ocrResult.Result != "" {
		// 保存验证码图片和识别结果到文件，方便调试（异步写入，不阻塞主流程）
		if debugEnabled {
			go saveCaptchaDebugImage(imageBase64, ocrResult.Result)
		}
		return ocrResult.Result, nil
	}

	return "", fmt.Errorf("OCR 识别失败")
}

// saveCaptchaDebugImage 保存验证码图片和识别结果到文件
// 文件名格式: <识别结果>.时间戳.png
func saveCaptchaDebugImage(imageBase64, result string) {
	debugDir := "captcha_debug"

	imgData, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		return
	}

	timestamp := time.Now().Format("20060102150405")
	cleanResult := result
	if cleanResult == "" {
		cleanResult = "unknown"
	}
	filename := fmt.Sprintf("%s.%s.png", cleanResult, timestamp)
	filePath := filepath.Join(debugDir, filename)

	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return
	}

	logging.Debugf("验证码图片已保存: %s", filePath)
}
