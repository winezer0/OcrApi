package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed config_template.yaml
var configTemplate []byte

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DashScopeConfig struct {
	ApiKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	Timeout int    `yaml:"timeout"`
}

type DdddOcrConfig struct {
	ModelDir string `yaml:"model_dir"`
}

type AppConfig struct {
	Server    ServerConfig    `yaml:"server"`
	DashScope DashScopeConfig `yaml:"dashscope"`
	DdddOcr   DdddOcrConfig   `yaml:"ddddocr"`
}

var (
	Cfg        *AppConfig
	configOnce bool
	configPath string
)

// InitConfig 初始化配置文件路径，必须在 LoadConfig 之前调用
func InitConfig(path string) {
	configPath = path
}

// GenConfig 生成默认配置文件到指定路径
func GenConfig(path string) error {
	if path == "" {
		path = "config.yaml"
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取配置文件绝对路径失败: %w", err)
	}

	if _, err := os.Stat(absPath); err == nil {
		return fmt.Errorf("配置文件已存在 [%s]，跳过生成", absPath)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置文件目录失败: %w", err)
	}

	if err := os.WriteFile(absPath, configTemplate, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("[*] 配置文件已生成: %s\n", absPath)
	return nil
}

// LoadConfig 从 yaml 配置文件加载配置
func LoadConfig() *AppConfig {
	if configOnce {
		return Cfg
	}

	if configPath == "" {
		configPath = "config.yaml"
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		panic(fmt.Sprintf("获取配置文件绝对路径失败: %v", err))
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败 [%s]: %v", absPath, err))
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	applyDefaults(&cfg)

	Cfg = &cfg
	configOnce = true
	return Cfg
}

// applyDefaults 填充配置默认值
func applyDefaults(cfg *AppConfig) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 5001
	}
	if cfg.DashScope.Timeout == 0 {
		cfg.DashScope.Timeout = 30
	}
	if cfg.DashScope.BaseURL == "" {
		cfg.DashScope.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if cfg.DashScope.Model == "" {
		cfg.DashScope.Model = "qwen3.5-flash"
	}
	if cfg.DdddOcr.ModelDir == "" {
		cfg.DdddOcr.ModelDir = "models"
	}
}

// GetDashScopeAPIKey 获取 DashScope API Key
func GetDashScopeAPIKey() string {
	return LoadConfig().DashScope.ApiKey
}

// GetDashScopeTimeout 获取 DashScope API 调用超时时间
func GetDashScopeTimeout() time.Duration {
	return time.Duration(LoadConfig().DashScope.Timeout) * time.Second
}

// GetServerAddr 获取服务监听地址
func GetServerAddr() string {
	cfg := LoadConfig()
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

// GetConfigDir 获取配置文件所在目录
func GetConfigDir() string {
	if configPath == "" {
		configPath = "config.yaml"
	}
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "."
	}
	return filepath.Dir(absPath)
}

// GetDdddOcrModelDir 获取 ddddocr 模型目录的绝对路径
func GetDdddOcrModelDir() string {
	cfg := LoadConfig()
	if filepath.IsAbs(cfg.DdddOcr.ModelDir) {
		return cfg.DdddOcr.ModelDir
	}
	return filepath.Join(GetConfigDir(), cfg.DdddOcr.ModelDir)
}
