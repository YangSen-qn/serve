package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	Host     string `json:"host"`      // 监听地址，如 :8080
	CertFile string `json:"cert_file"` // SSL 证书文件路径
	KeyFile  string `json:"key_file"`  // SSL 私钥文件路径

	// 日志配置
	LogLevel string `json:"log_level"` // 日志等级：debug, info, warn, error

	// 静态文件服务配置
	StaticDir string `json:"static_dir"` // 静态文件目录路径

	// 代理配置
	ProxyConfigs map[string]*ProxyConfig `json:"proxy_configs"` // 代理配置映射，key 为目标域名
}

// ProxyConfig 代理配置结构
type ProxyConfig struct {
	TargetDomain string `json:"target_domain"` // 目标域名，如果为空则使用路径第一段作为目标域名
	UseHTTPS     bool   `json:"use_https"`     // 是否使用 HTTPS 协议
	Insecure     bool   `json:"insecure"`      // 是否跳过 SSL 证书验证（仅在 use_https 为 true 时生效）
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	return &Config{
		Host:         ":8080",
		LogLevel:     "info",
		StaticDir:    "./static",
		ProxyConfigs: make(map[string]*ProxyConfig),
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证日志等级
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error", c.LogLevel)
	}

	// 如果配置了证书文件，验证文件是否存在
	if c.CertFile != "" || c.KeyFile != "" {
		if c.CertFile == "" || c.KeyFile == "" {
			return fmt.Errorf("both cert_file and key_file must be provided for HTTPS")
		}

		// 检查证书文件是否存在
		if _, err := os.Stat(c.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("certificate file not found: %s", c.CertFile)
		}
		if _, err := os.Stat(c.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("key file not found: %s", c.KeyFile)
		}

		// 验证证书和私钥是否匹配
		if _, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile); err != nil {
			return fmt.Errorf("failed to load certificate pair: %v", err)
		}
	}

	// 验证静态文件目录
	if c.StaticDir != "" {
		absPath, err := filepath.Abs(c.StaticDir)
		if err != nil {
			return fmt.Errorf("failed to resolve static directory path: %v", err)
		}
		c.StaticDir = absPath

		// 检查目录是否存在
		if info, err := os.Stat(c.StaticDir); os.IsNotExist(err) {
			return fmt.Errorf("static directory not found: %s", c.StaticDir)
		} else if !info.IsDir() {
			return fmt.Errorf("static path is not a directory: %s", c.StaticDir)
		}
	}

	return nil
}

// IsHTTPS 判断是否启用 HTTPS
func (c *Config) IsHTTPS() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// GetLogLevel 获取日志等级
func (c *Config) GetLogLevel() logrus.Level {
	switch c.LogLevel {
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// AddProxyConfig 添加代理配置
// pathPrefix: 路径前缀（用于匹配请求路径第一段）
// targetDomain: 目标域名，如果为空则使用路径第一段作为目标域名
// useHTTPS: 是否使用 HTTPS 协议
// insecure: 是否跳过 SSL 证书验证
func (c *Config) AddProxyConfig(pathPrefix, targetDomain string, useHTTPS, insecure bool) {
	c.ProxyConfigs[pathPrefix] = &ProxyConfig{
		TargetDomain: targetDomain,
		UseHTTPS:     useHTTPS,
		Insecure:     insecure,
	}
}

// GetProxyConfig 获取指定路径前缀的代理配置
func (c *Config) GetProxyConfig(pathPrefix string) (*ProxyConfig, bool) {
	config, exists := c.ProxyConfigs[pathPrefix]
	return config, exists
}
