package config

import (
	"fmt"
	"net/url"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	LoginURL       string `long:"login-url" short:"u" default:"https://demo.ruoyi.vip/login" description:"登录接口 URL"`
	CaptchaURL     string `long:"captcha-url" short:"c" default:"https://demo.ruoyi.vip/captcha/captchaImage?type=math" description:"验证码接口 URL"`
	Proxy          string `long:"proxy" short:"x" default:"http://127.0.0.1:8080" description:"HTTP/HTTPS 代理地址"`
	OcrProxy       string `long:"ocr-proxy" default:"" description:"OCR API 代理地址 (默认为空，不使用代理)"`
	Success        string `long:"success" short:"s" default:"操作成功" description:"登录成功关键字 (多个请用逗号分隔, 为空不主动停止)"`
	Failure        string `long:"failure" short:"f" default:"用户不存在/密码错误" description:"登录失败关键字"`
	CaptchaErr     string `long:"captcha-err" short:"e" default:"验证码错误" description:"验证码错误关键字"`
	Output         string `long:"output" short:"o" default:"scan_results.csv" description:"结果保存的 CSV 路径"`
	OcrAPI         string `long:"ocr-api" short:"a" default:"http://127.0.0.1:5001/ruoyi/base64" description:"本地 OCR API 地址"`
	LoginWorkers   int    `long:"login-workers" short:"L" default:"10" description:"并发登录线程数"`
	FillerWorkers  int    `long:"filler-workers" short:"F" default:"15" description:"并发验证码填充线程数"`
	PoolSize       int    `long:"pool-size" short:"S" default:"20" description:"验证码池容量"`
	MaxConnections int    `long:"max-connections" short:"C" default:"100" description:"全局最大 HTTP 连接数"`
	MaxKeepalive   int    `long:"max-keepalive" short:"K" default:"20" description:"最大 Keep-Alive 连接数"`
	UserFile       string `long:"user-file" short:"U" default:"username.txt" description:"用户名字典文件路径"`
	PassFile       string `long:"pass-file" short:"P" default:"password.txt" description:"密码字典文件路径"`
	DebugCaptcha   bool   `long:"debug-captcha" description:"是否保存验证码图片到 debug 目录用于调试"`
	VerifySSL      bool   `long:"verify-ssl" description:"是否验证 SSL 证书 (默认关闭)"`
	LogFile        string `long:"lf" description:"日志文件路径 (默认: 空)"`
	LogLevel       string `long:"ll" description:"日志级别 (debug/info/warn/error)" default:"info"`
	ConsoleFormat  string `long:"cf" description:"控制台日志格式 (T L C M F 组合或 off|null 禁用)" default:"TLCM"`
	NoCache        bool   `long:"no-cache" description:"启动时忽略缓存过滤 (仍会写入缓存)"`
}

type Config struct {
	LoginURL            string
	CaptchaURL          string
	Proxy               string
	OcrProxy            string
	Success             string
	Failure             string
	CaptchaErr          string
	Output              string
	OcrAPI              string
	LoginWorkers        int
	FillerWorkers       int
	PoolSize            int
	MaxConnections      int
	MaxKeepalive        int
	UserFile            string
	PassFile            string
	DebugCaptcha        bool
	LogFile             string
	LogLevel            string
	ConsoleFormat       string
	NoCache             bool
	SuccessKeywords     []string
	FailureKeywords     []string
	CaptchaKeywords     []string
	ProxyURL            string
	OcrProxyURL         string
	VerifySSL           bool
	CacheFile           string
	SuccessFile         string
	DebugCaptchaEnabled bool
}

func ParseArgs() (*Config, error) {
	var opts Options
	_, err := flags.Parse(&opts)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		LoginURL:       opts.LoginURL,
		CaptchaURL:     opts.CaptchaURL,
		Proxy:          opts.Proxy,
		OcrProxy:       opts.OcrProxy,
		Success:        opts.Success,
		Failure:        opts.Failure,
		CaptchaErr:     opts.CaptchaErr,
		Output:         opts.Output,
		OcrAPI:         opts.OcrAPI,
		LoginWorkers:   opts.LoginWorkers,
		FillerWorkers:  opts.FillerWorkers,
		PoolSize:       opts.PoolSize,
		MaxConnections: opts.MaxConnections,
		MaxKeepalive:   opts.MaxKeepalive,
		UserFile:       opts.UserFile,
		PassFile:       opts.PassFile,
		DebugCaptcha:   opts.DebugCaptcha,
		VerifySSL:      opts.VerifySSL,
		LogFile:        opts.LogFile,
		LogLevel:       opts.LogLevel,
		ConsoleFormat:  opts.ConsoleFormat,
		NoCache:        opts.NoCache,
	}

	if opts.Success != "" {
		cfg.SuccessKeywords = splitKeywords(opts.Success)
	}
	cfg.FailureKeywords = splitKeywords(opts.Failure)
	cfg.CaptchaKeywords = splitKeywords(opts.CaptchaErr)

	if opts.Proxy != "" {
		if _, err := url.Parse(opts.Proxy); err != nil {
			return nil, fmt.Errorf("代理地址格式无效: %w", err)
		}
		cfg.ProxyURL = opts.Proxy
	}

	if opts.OcrProxy != "" {
		if _, err := url.Parse(opts.OcrProxy); err != nil {
			return nil, fmt.Errorf("OCR 代理地址格式无效: %w", err)
		}
		cfg.OcrProxyURL = opts.OcrProxy
	}

	host := extractHost(opts.LoginURL)
	cfg.CacheFile = fmt.Sprintf("%s.error.cache", host)
	cfg.SuccessFile = fmt.Sprintf("%s.success.cache", host)

	return cfg, nil
}

func splitKeywords(kw string) []string {
	parts := strings.Split(kw, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}
	return u.Host
}
