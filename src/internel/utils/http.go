package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var defaultHTTPTransporter = NewHTTPTransporter(30)

// NewHTTPTransporter is used to create HTTP transporter
func NewHTTPTransporter(callTimeout int) *HTTPTransporter {
	return &HTTPTransporter{
		client: &http.Client{
			Timeout: time.Duration(callTimeout) * time.Second,
		},
	}
}

// HTTPTransporter define
type HTTPTransporter struct {
	client *http.Client
}

// Request is used to send HTTP request
func (s HTTPTransporter) Request(method, url string, params, headers map[string]interface{}, data interface{}) ([]byte, error) {
	// Set request data
	var req *http.Request
	var err error
	// Initialize request client
	// Accept "GET" and "POST" request method
	switch method {
	case "GET":
		req, err = http.NewRequest(method, url, nil)
	case "POST", "PUT":
		var v []byte
		v, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("requset `%s` error: `%s`", url, err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(v))
		if err != nil {
			return nil, fmt.Errorf("requset `%s` error: `%s`", url, err)
		}
		// Request auto supplement header "Content-Type:application/json"
		req.Header.Set("Content-Type", "application/json")
	default:
		err = fmt.Errorf("no support `%s` method", method)
	}
	if err != nil || req == nil {
		return nil, fmt.Errorf("requset `%s` error: `%s`", url, err)
	}
	// Add request params
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, fmt.Sprint(val))
	}
	req.URL.RawQuery = q.Encode()
	// Add request header
	for key, val := range headers {
		req.Header.Set(key, fmt.Sprint(val))
	}
	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requset `%s` error: `%s`", url, err)
	}
	// Get request data
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("requset `%s` error: `%s`", url, err)
	}
	defer resp.Body.Close()
	// Return request result
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("requset `%s` status `%d`, context `%s`", url, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// HTTPRequest is used to send HTTP request
func HTTPRequest(method, url string, params, headers map[string]interface{}, data interface{}) ([]byte, error) {
	return defaultHTTPTransporter.Request(method, url, params, headers, data)
}

func HTTPRequestWithJSON(method, url string, params, headers map[string]interface{}, data, dst interface{}) error {
	output, err := HTTPRequest(method, url, params, headers, data)
	if err != nil {
		return err
	}
	return json.Unmarshal(output, dst)
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Response   *http.Response
}

func HTTPRequestFullWithTimeout(method, url string, params, headers map[string]interface{}, data interface{}, timeout int) (*HTTPResponse, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			// 控制连接建立超时
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: time.Duration(timeout) * time.Second,
			}).DialContext,
			// 关键：TLS 握手超时不能太短！
			TLSHandshakeTimeout: 15 * time.Second,
			// 其他可选优化
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
			TLSClientConfig: &tls.Config{
				// 可选：仅调试时开启
				// InsecureSkipVerify: true, // ⚠️ 不要在线上跳过验证
			},
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// 构造 URL 参数
	reqURL, err := urlWithParams(url, params)
	if err != nil {
		return nil, fmt.Errorf("构造 URL 参数失败: %w", err)
	}

	var reqBody io.Reader
	if method == "POST" || method == "PUT" {
		bodyBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 headers
	for key, val := range headers {
		req.Header.Set(key, fmt.Sprint(val))
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// 读取响应内容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Response:   resp,
	}, nil
}

// HTTPRequestFull 支持任意 method、params、headers 和 body，返回完整响应信息
func HTTPRequestFull(method, url string, params, headers map[string]interface{}, data interface{}) (*HTTPResponse, error) {
	return HTTPRequestFullWithTimeout(method, url, params, headers, data, 30)
}

// urlWithParams 将 map 参数拼接到 url
func urlWithParams(rawURL string, params map[string]interface{}) (string, error) {
	if params == nil || len(params) == 0 {
		return rawURL, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for k, v := range params {
		q.Add(k, fmt.Sprint(v))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
