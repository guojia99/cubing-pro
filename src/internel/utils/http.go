package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

// HTTPRequestFull 支持任意 method、params、headers 和 body，返回完整响应信息
func HTTPRequestFull(method, url string, params, headers map[string]interface{}, data interface{}) (*HTTPResponse, error) {
	client := &http.Client{}
	client.Timeout = time.Duration(30) * time.Second

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
	defer resp.Body.Close()

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
