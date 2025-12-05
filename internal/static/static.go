package static

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// Handler 静态文件服务处理器
type Handler struct {
	dir        string // 静态文件目录
	fileServer http.Handler
	logger     *logrus.Logger
}

// NewHandler 创建静态文件服务处理器
func NewHandler(staticDir string, logger *logrus.Logger) *Handler {
	return &Handler{
		dir:        staticDir,
		fileServer: http.FileServer(http.Dir(staticDir)),
		logger:     logger,
	}
}

// ServeHTTP 处理静态文件请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 清理路径，防止路径遍历攻击
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 移除路径中的 ".." 和 "." 等不安全字符
	path = filepath.Clean(path)
	if strings.Contains(path, "..") {
		h.logger.Warnf("Invalid path detected: %s", r.URL.Path)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// 记录访问日志
	h.logger.Debugf("Serving static file: %s", path)

	// 使用标准文件服务器处理请求
	h.fileServer.ServeHTTP(w, r)
}
