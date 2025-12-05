package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"serve/internal/config"

	"github.com/sirupsen/logrus"
)

// Handler 反向代理处理器
type Handler struct {
	config *config.Config
	logger *logrus.Logger
}

// NewHandler 创建反向代理处理器
func NewHandler(cfg *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{
		config: cfg,
		logger: logger,
	}
}

// ServeHTTP 处理代理请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 解析请求路径
	path := r.URL.Path

	// 移除开头的 "/"
	path = strings.TrimPrefix(path, "/")

	// 分割路径，第一段作为目标域名
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		h.logger.Warnf("Invalid proxy path: %s", r.URL.Path)
		http.Error(w, "Invalid proxy path", http.StatusBadRequest)
		return
	}

	targetDomain := pathParts[0]

	// 检查是否存在该域名的代理配置
	proxyConfig, exists := h.config.GetProxyConfig(targetDomain)
	if !exists {
		h.logger.Debugf("No proxy config found for domain: %s", targetDomain)
		http.Error(w, fmt.Sprintf("No proxy configuration found for domain: %s", targetDomain), http.StatusNotFound)
		return
	}

	// 构建目标 URL
	// 后续路径段拼接
	var targetPath string
	if len(pathParts) > 1 {
		targetPath = "/" + strings.Join(pathParts[1:], "/")
	} else {
		targetPath = "/"
	}

	// 确定协议 scheme
	scheme := "http"
	if proxyConfig.UseHTTPS {
		scheme = "https"
	}

	// 构建完整的目标 URL
	targetURL := fmt.Sprintf("%s://%s%s", scheme, targetDomain, targetPath)

	// 添加查询参数
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	h.logger.Infof("Proxying request: %s %s -> %s", r.Method, r.URL.String(), targetURL)

	// 解析目标 URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		h.logger.Errorf("Failed to parse target URL: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// 配置传输层，处理 SSL 证书验证
	if proxyConfig.UseHTTPS && proxyConfig.Insecure {
		// 跳过 SSL 证书验证
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		h.logger.Debugf("SSL certificate verification disabled for: %s", targetDomain)
	}

	// 修改请求，设置正确的 Host 头
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetDomain
		req.URL.Host = targetDomain
		req.URL.Scheme = scheme
		req.URL.Path = targetPath

		// 保留原始查询参数
		if r.URL.RawQuery != "" {
			req.URL.RawQuery = r.URL.RawQuery
		}

		h.logger.Debugf("Proxy request details: Method=%s, URL=%s, Host=%s",
			req.Method, req.URL.String(), req.Host)
	}

	// 执行代理请求
	proxy.ServeHTTP(w, r)
}

// IsProxyPath 判断请求路径是否为代理路径
// 检查路径的第一段是否匹配已配置的代理域名
func (h *Handler) IsProxyPath(path string) bool {
	// 清理路径
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	// 分割路径
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		return false
	}

	targetDomain := pathParts[0]
	_, exists := h.config.GetProxyConfig(targetDomain)
	return exists
}
