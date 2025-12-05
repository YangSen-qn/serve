package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"serve/internal/config"
	"serve/internal/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// 命令行参数
	host      string
	certFile  string
	keyFile   string
	logLevel  string
	staticDir string

	// 代理配置（格式：domain:use_https:insecure，多个用逗号分隔）
	proxyConfigs string
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "serve",
	Short: "HTTP/HTTPS 服务器，支持静态文件服务和反向代理",
	Long: `serve 是一个基于 Go 的 HTTP/HTTPS 服务器，集成了静态文件服务和反向代理功能。

功能特性：
- 支持 HTTP 和 HTTPS 协议
- 支持静态文件服务
- 支持反向代理，通过路径前缀匹配目标域名
- 支持配置日志等级
- 支持配置 SSL 证书

使用示例：
  # 启动 HTTP 服务器
  serve --host :8080 --static-dir ./static

  # 启动 HTTPS 服务器
  serve --host :8443 --cert-file cert.pem --key-file key.pem --static-dir ./static

  # 配置代理
  serve --host :8080 --static-dir ./static --proxy www.example.com:true:false
`,
	Run: runServer,
}

func init() {
	// 绑定命令行参数
	rootCmd.Flags().StringVar(&host, "host", ":8080", "监听地址（如 :8080）")
	rootCmd.Flags().StringVar(&certFile, "ssl-cert-file", "", "SSL 证书文件路径（启用 HTTPS）")
	rootCmd.Flags().StringVar(&keyFile, "ssl-key-file", "", "SSL 私钥文件路径（启用 HTTPS）")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "日志等级（debug, info, warn, error）")
	rootCmd.Flags().StringVar(&staticDir, "static-dir", "./static", "静态文件目录路径")
	rootCmd.Flags().StringVar(&proxyConfigs, "proxy", "",
		`反向代理配置，格式：domain:use_https:insecure，多个配置用逗号分隔
		
格式说明：
  - domain: 目标域名（如 www.example.com），将作为请求路径的第一段进行匹配
  - use_https: 是否使用 HTTPS 协议，可选值：true（使用 HTTPS）或 false（使用 HTTP）
  - insecure: 是否跳过 SSL 证书验证，可选值：true（跳过验证）或 false（验证证书）
            仅在 use_https 为 true 时生效

工作原理：
  请求路径格式：/{domain}/{path}?{query}
  例如：http://localhost:8080/www.example.com/api/users?id=1
  会被代理到：http://www.example.com/api/users?id=1（use_https=false）
  或：https://www.example.com/api/users?id=1（use_https=true）

使用示例：
  --proxy www.example.com:true:false
    代理到 https://www.example.com，验证 SSL 证书
  
  --proxy api.example.com:true:true
    代理到 https://api.example.com，跳过 SSL 证书验证
  
  --proxy test.example.com:false:false
    代理到 http://test.example.com
  
  --proxy "www.example.com:true:false,api.example.com:false:false"
    配置多个代理，用逗号分隔`)
}

// runServer 运行服务器
func runServer(cmd *cobra.Command, args []string) {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// 创建配置
	cfg := config.LoadConfig()
	cfg.Host = host
	cfg.CertFile = certFile
	cfg.KeyFile = keyFile
	cfg.LogLevel = logLevel
	cfg.StaticDir = staticDir

	// 解析代理配置
	if err := parseProxyConfigs(cfg, proxyConfigs); err != nil {
		logger.Fatalf("Failed to parse proxy configs: %v", err)
	}

	// 设置日志等级
	logger.SetLevel(cfg.GetLogLevel())

	// 验证配置
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("Invalid configuration: %v", err)
	}

	// 创建服务器
	srv := server.NewServer(cfg, logger)

	// 启动服务器（在 goroutine 中运行）
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Infof("Server started successfully on %s", srv.GetAddr())
	if cfg.IsHTTPS() {
		logger.Info("HTTPS mode enabled")
	} else {
		logger.Info("HTTP mode enabled")
	}
	logger.Infof("Static directory: %s", cfg.StaticDir)
	logger.Infof("Proxy configurations: %d", len(cfg.ProxyConfigs))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Received shutdown signal")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	} else {
		logger.Info("Server exited gracefully")
	}
}

// parseProxyConfigs 解析代理配置字符串
// 格式：domain:use_https:insecure，多个配置用逗号分隔
// 示例：www.example.com:true:false,api.example.com:false:false
func parseProxyConfigs(cfg *config.Config, proxyStr string) error {
	if proxyStr == "" {
		return nil
	}

	// 分割多个配置
	configs := strings.Split(proxyStr, ",")
	for _, configStr := range configs {
		configStr = strings.TrimSpace(configStr)
		if configStr == "" {
			continue
		}

		// 分割配置项
		parts := strings.Split(configStr, ":")
		if len(parts) != 3 {
			return fmt.Errorf("invalid proxy config format: %s (expected: domain:use_https:insecure)", configStr)
		}

		domain := strings.TrimSpace(parts[0])
		useHTTPSStr := strings.TrimSpace(parts[1])
		insecureStr := strings.TrimSpace(parts[2])

		// 解析 use_https
		useHTTPS := false
		if useHTTPSStr == "true" {
			useHTTPS = true
		} else if useHTTPSStr != "false" {
			return fmt.Errorf("invalid use_https value: %s (must be true or false)", useHTTPSStr)
		}

		// 解析 insecure
		insecure := false
		if insecureStr == "true" {
			insecure = true
		} else if insecureStr != "false" {
			return fmt.Errorf("invalid insecure value: %s (must be true or false)", insecureStr)
		}

		// 添加代理配置
		cfg.AddProxyConfig(domain, useHTTPS, insecure)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
