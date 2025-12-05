package server

import (
	"context"
	"net/http"
	"time"

	"serve/internal/config"
	"serve/internal/proxy"
	"serve/internal/static"

	"github.com/sirupsen/logrus"
)

// Server HTTP/HTTPS 服务器
type Server struct {
	config     *config.Config
	httpServer *http.Server
	logger     *logrus.Logger
}

// NewServer 创建新的服务器实例
func NewServer(cfg *config.Config, logger *logrus.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 创建路由处理器
	mux := http.NewServeMux()

	// 创建静态文件处理器
	staticHandler := static.NewHandler(s.config.StaticDir, s.logger)

	// 创建代理处理器
	proxyHandler := proxy.NewHandler(s.config, s.logger)

	// 注册路由处理函数
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 首先检查是否为代理路径
		if proxyHandler.IsProxyPath(r.URL.Path) {
			proxyHandler.ServeHTTP(w, r)
			return
		}

		// 否则使用静态文件服务
		staticHandler.ServeHTTP(w, r)
	})

	// 创建 HTTP 服务器
	s.httpServer = &http.Server{
		Addr:         s.config.Host,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 根据配置启动 HTTP 或 HTTPS 服务
	if s.config.IsHTTPS() {
		s.logger.Infof("Starting HTTPS server on %s", s.config.Host)
		s.logger.Infof("Certificate: %s, Key: %s", s.config.CertFile, s.config.KeyFile)
		return s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	}

	s.logger.Infof("Starting HTTP server on %s", s.config.Host)
	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// GetAddr 获取服务器监听地址
func (s *Server) GetAddr() string {
	return s.config.Host
}

