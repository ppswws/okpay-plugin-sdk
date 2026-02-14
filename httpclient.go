package plugin

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClientConfig 描述插件侧 HTTP 客户端配置。
type HTTPClientConfig struct {
	Timeout     time.Duration
	Retry       int
	ConnTimeout time.Duration
	ReqTimeout  time.Duration
	ProxyURL    string
}

// HTTPClient 提供可复用连接的请求客户端（带重试与耗时统计）。
type HTTPClient struct {
	httpClient *http.Client
	retry      int
}

const (
	defaultRetry       = 3
	defaultConnTimeout = 2 * time.Second
	defaultReqTimeout  = 10 * time.Second
)

// NewHTTPClient 创建插件侧 HTTP 客户端。
func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	retry := cfg.Retry
	if retry <= 0 {
		retry = defaultRetry
	}

	connTimeout := cfg.ConnTimeout
	if connTimeout <= 0 {
		connTimeout = defaultConnTimeout
	}
	reqTimeout := cfg.ReqTimeout
	if reqTimeout <= 0 {
		reqTimeout = defaultReqTimeout
	}

	var proxyFunc func(*http.Request) (*url.URL, error)
	if strings.TrimSpace(cfg.ProxyURL) != "" {
		if parsed, err := url.Parse(cfg.ProxyURL); err == nil {
			proxyFunc = http.ProxyURL(parsed)
		}
	}
	transport := &http.Transport{
		Proxy: proxyFunc,
		DialContext: (&net.Dialer{
			Timeout:   connTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: reqTimeout,
		TLSHandshakeTimeout:   2 * time.Second,
		IdleConnTimeout:       60 * time.Second,
		MaxIdleConns:          100,
	}

	clientTimeout := connTimeout + reqTimeout
	if cfg.Timeout > 0 {
		clientTimeout = cfg.Timeout
	}

	return &HTTPClient{
		httpClient: &http.Client{
			Timeout:   clientTimeout,
			Transport: transport,
		},
		retry: retry,
	}
}

// Do 执行请求，返回响应内容与请求统计。
// method: GET/POST/PUT/DELETE...
// reqBody: 原始请求体（用于记录），contentType 为空则不设置。
func (c *HTTPClient) Do(ctx context.Context, method, endpoint, reqBody, contentType string) (string, int16, int32, error) {
	start := time.Now()
	body, count, err := c.do(ctx, func() (*http.Request, error) {
		var reader io.Reader
		if reqBody != "" {
			reader = strings.NewReader(reqBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		return req, nil
	})
	reqMs := int32(time.Since(start).Milliseconds())
	return body, count, reqMs, err
}

func (c *HTTPClient) do(ctx context.Context, build func() (*http.Request, error)) (string, int16, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if build == nil {
		return "", 0, http.ErrBodyNotAllowed
	}
	attempts := c.retry
	if attempts <= 0 {
		attempts = defaultRetry
	}
	var count int16
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return "", count, ctx.Err()
		}
		req, err := build()
		if err != nil {
			return "", count, err
		}
		count++
		resp, err := c.httpClient.Do(req)
		if err == nil && resp != nil && resp.StatusCode < http.StatusInternalServerError {
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", count, err
			}
			return string(b), count, nil
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if i == attempts-1 {
			if err != nil {
				return "", count, err
			}
			if resp != nil {
				return "", count, nil
			}
		}
		if ctx.Err() != nil {
			return "", count, ctx.Err()
		}
	}
	return "", count, context.DeadlineExceeded
}
