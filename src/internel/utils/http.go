package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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
	if err != nil {
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
