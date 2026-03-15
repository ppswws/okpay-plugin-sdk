package sdk

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ppswws/okpay-plugin-sdk/contract"
)

// HTTPClientConfig 描述插件侧 HTTP 客户端配置。
type HTTPClientConfig struct {
	ProxyURL             string
	InsecureSkipVerify   bool
	MaxResponseBodyBytes int64
}

// HTTPClient 提供可复用连接的请求客户端（带重试与耗时统计）。
type HTTPClient struct {
	attempts             int
	proxyFn              func(*http.Request) (*url.URL, error)
	insecureSkipVerify   bool
	maxResponseBodyBytes int64
	clients              sync.Map // map[time.Duration]*http.Client
}

const (
	defaultAttempts            = 2
	defaultTimeoutMS           = int64(10000)
	maxTimeoutMS               = int64(20000)
	minTimeoutMS               = int64(5000)
	defaultMaxIdleConns        = 200
	defaultMaxIdleConnsPerHost = 100
	defaultMaxConnsPerHost     = 200
	defaultMaxResponseBodySize = int64(2 * 1024 * 1024) // 2MiB
)

// NewHTTPClient 创建插件侧 HTTP 客户端。
func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	var proxyFunc func(*http.Request) (*url.URL, error)
	if strings.TrimSpace(cfg.ProxyURL) != "" {
		if parsed, err := url.Parse(cfg.ProxyURL); err == nil {
			proxyFunc = http.ProxyURL(parsed)
		}
	}
	maxBodyBytes := cfg.MaxResponseBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxResponseBodySize
	}
	return &HTTPClient{
		attempts:             defaultAttempts,
		proxyFn:              proxyFunc,
		insecureSkipVerify:   cfg.InsecureSkipVerify,
		maxResponseBodyBytes: maxBodyBytes,
	}
}

// Do 执行请求，返回响应内容与请求统计。
// method: GET/POST/PUT/DELETE...
// reqBody: 原始请求体（用于记录），contentType 为空则不设置。
func (c *HTTPClient) Do(
	ctx context.Context,
	method, endpoint, reqBody, contentType string,
	headers ...map[string]string,
) (string, int16, int32, error) {
	if c == nil {
		c = NewHTTPClient(HTTPClientConfig{})
	}
	start := time.Now()
	timeout := resolveHTTPTimeout(ctx)
	client := c.clientForTimeout(timeout)
	body, count, err := c.do(ctx, client, func(callCtx context.Context) (*http.Request, error) {
		var reader io.Reader
		if reqBody != "" {
			reader = strings.NewReader(reqBody)
		}
		req, err := http.NewRequestWithContext(callCtx, method, endpoint, reader)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		if len(headers) > 0 {
			for k, v := range headers[0] {
				if strings.TrimSpace(k) == "" {
					continue
				}
				req.Header.Set(k, v)
			}
		}
		return req, nil
	})
	reqMs := int32(time.Since(start).Milliseconds())
	return body, count, reqMs, err
}

func (c *HTTPClient) do(ctx context.Context, client *http.Client, build func(context.Context) (*http.Request, error)) (string, int16, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		if c != nil {
			client = c.clientForTimeout(time.Duration(defaultTimeoutMS) * time.Millisecond)
		}
		if client == nil {
			client = &http.Client{Timeout: time.Duration(defaultTimeoutMS) * time.Millisecond}
		}
	}
	if build == nil {
		return "", 0, http.ErrBodyNotAllowed
	}
	maxResponseBodyBytes := defaultMaxResponseBodySize
	if c != nil && c.maxResponseBodyBytes > 0 {
		maxResponseBodyBytes = c.maxResponseBodyBytes
	}
	attempts := defaultAttempts
	if c != nil && c.attempts > 0 {
		attempts = c.attempts
	}
	var count int16
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return "", count, ctx.Err()
		}
		req, err := build(ctx)
		if err != nil {
			return "", count, err
		}
		count++
		resp, err := client.Do(req)
		if resp != nil {
			body, readErr := readBodyWithLimit(resp.Body, maxResponseBodyBytes)
			if readErr != nil {
				return "", count, readErr
			}
			// Any response means upstream replied; do not retry in this case.
			if err != nil {
				return body, count, err
			}
			if isHTTPSuccess(resp.StatusCode) {
				return body, count, nil
			}
			return body, count, fmt.Errorf("http status %d", resp.StatusCode)
		}
		if err != nil && (i == attempts-1 || !shouldRetryOnTransportError(err)) {
			return "", count, err
		}
		if ctx.Err() != nil {
			return "", count, ctx.Err()
		}
	}
	return "", count, context.DeadlineExceeded
}

func resolveHTTPTimeout(ctx context.Context) time.Duration {
	timeout, ok := contract.HTTPTimeoutFromContext(ctx)
	if !ok || timeout <= 0 {
		return time.Duration(defaultTimeoutMS) * time.Millisecond
	}
	ms := timeout.Milliseconds()
	if ms <= 0 {
		return time.Duration(defaultTimeoutMS) * time.Millisecond
	}
	if ms < minTimeoutMS {
		ms = minTimeoutMS
	}
	if ms > maxTimeoutMS {
		ms = maxTimeoutMS
	}
	return time.Duration(ms) * time.Millisecond
}

func (c *HTTPClient) clientForTimeout(timeout time.Duration) *http.Client {
	if c == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = time.Duration(defaultTimeoutMS) * time.Millisecond
	}
	if v, ok := c.clients.Load(timeout); ok {
		if hc, ok := v.(*http.Client); ok && hc != nil {
			return hc
		}
	}
	connTimeout := timeout / 2
	transport := &http.Transport{
		Proxy: c.proxyFn,
		DialContext: (&net.Dialer{
			Timeout:   connTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		IdleConnTimeout:     60 * time.Second,
		MaxIdleConns:        defaultMaxIdleConns,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		MaxConnsPerHost:     defaultMaxConnsPerHost,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: c.insecureSkipVerify},
	}
	hc := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	actual, _ := c.clients.LoadOrStore(timeout, hc)
	if out, ok := actual.(*http.Client); ok && out != nil {
		return out
	}
	return hc
}

func isHTTPSuccess(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}

func shouldRetryOnTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	return true
}

func readBodyWithLimit(body io.ReadCloser, maxBytes int64) (string, error) {
	if body == nil {
		return "", nil
	}
	defer body.Close()
	if maxBytes <= 0 {
		maxBytes = defaultMaxResponseBodySize
	}
	limited := io.LimitReader(body, maxBytes+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if int64(len(b)) > maxBytes {
		_, _ = io.Copy(io.Discard, body)
		return "", fmt.Errorf("response body too large: limit=%d", maxBytes)
	}
	return string(b), nil
}
