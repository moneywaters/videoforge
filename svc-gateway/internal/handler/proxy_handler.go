package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// ProxyHandler handles proxying requests to internal services
type ProxyHandler struct {
	serviceURLs map[string]string
	timeout    time.Duration
}

// NewProxyHandler creates a new ProxyHandler
func NewProxyHandler(serviceURLs map[string]string) *ProxyHandler {
	return &ProxyHandler{
		serviceURLs: serviceURLs,
		timeout:    30 * time.Second,
	}
}

// SetTimeout sets the proxy timeout
func (p *ProxyHandler) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

// ProxyTo proxies a request to a specific target URL
// TODO: Replace with NATS request-reply for production
func (p *ProxyHandler) ProxyTo(targetURL string, w http.ResponseWriter, r *http.Request) {
	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Customize the director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve headers from original request
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", r.Proto)

		// Copy user info from context if present
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}
		if role := r.Header.Get("X-User-Role"); role != "" {
			req.Header.Set("X-User-Role", role)
		}
	}

	// Set timeout
	httpClient := &http.Client{
		Timeout: p.timeout,
	}
	proxy.Client = httpClient

	log.Printf("Proxying %s %s to %s", r.Method, r.URL.Path, targetURL)

	proxy.ServeHTTP(w, r)
}

// ProxyToService proxies a request to a specific service
func (p *ProxyHandler) ProxyToService(serviceName string, w http.ResponseWriter, r *http.Request) {
	baseURL, ok := p.serviceURLs[serviceName]
	if !ok {
		http.NotFound(w, r)
		log.Printf("Service not found: %s", serviceName)
		return
	}

	targetURL := baseURL + r.URL.Path
	p.ProxyTo(targetURL, w, r)
}

// ProxyAPI handles API proxy requests
// Routes based on path prefix to appropriate service
func (p *ProxyHandler) ProxyAPI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract service name from path
	// /api/v1/service_name/... -> service_name
	parts := strings.SplitN(path, "/", 4)
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	serviceName := parts[2] // "users", "briefs", etc.

	baseURL, ok := p.serviceURLs[serviceName]
	if !ok {
		http.NotFound(w, r)
		log.Printf("Service not found: %s", serviceName)
		return
	}

	targetURL := baseURL + path
	p.ProxyTo(targetURL, w, r)
}

// ServeHTTP implements http.Handler for direct routing
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ProxyAPI(w, r)
}

// ProxyHandlerFunc returns an http.HandlerFunc for proxying
func (p *ProxyHandler) ProxyHandlerFunc(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p.ProxyToService(serviceName, w, r)
	}
}

// CopyResponse copies the response from the proxy to the original writer
func CopyResponse(proxyResp *http.Response, w http.ResponseWriter) error {
	// Copy headers
	for k, v := range proxyResp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	// Set status
	w.WriteHeader(proxyResp.StatusCode)

	// Copy body
	_, err := io.Copy(w, proxyResp.Body)
	return err
}