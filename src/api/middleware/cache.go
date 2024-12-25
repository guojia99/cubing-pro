package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheEntry 存储缓存的响应内容及其创建时间
type CacheEntry struct {
	ResponseBody []byte
	CreatedAt    time.Time
}

// ResponseRecorder 捕获 gin 响应的自定义 writer
type ResponseRecorder struct {
	gin.ResponseWriter
	Writer io.Writer
}

// Write 捕获响应数据
func (w *ResponseRecorder) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// 使用 sync.Map 存储缓存数据
var cacheStore sync.Map

// GenerateCacheKey 根据请求生成一个唯一的缓存键
func GenerateCacheKey(c *gin.Context) (string, error) {
	var buf bytes.Buffer

	// 1. 路由路径
	buf.WriteString(c.FullPath())

	// 2. 查询参数
	queryParams, _ := json.Marshal(c.Request.URL.Query())
	buf.Write(queryParams)

	// 3. 请求体
	if c.Request.Body != nil {
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return "", err
		}
		buf.Write(bodyBytes)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // 重置请求体以便后续中间件或处理函数读取
	}

	// 使用 SHA256 哈希生成唯一的缓存键
	hash := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(hash[:]), nil
}

// CacheMiddleware 缓存中间件
func CacheMiddleware(cacheDuration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey, err := GenerateCacheKey(c)
		if err != nil {
			c.Next()
			return
		}

		// 检查缓存中是否存在该请求的响应
		if entry, found := cacheStore.Load(cacheKey); found {
			cacheEntry := entry.(CacheEntry)
			// 检查缓存是否过期
			if time.Since(cacheEntry.CreatedAt) < cacheDuration {
				// 使用缓存响应数据
				c.Data(http.StatusOK, "application/json", cacheEntry.ResponseBody)
				c.Abort()
				return
			}
			// 缓存过期，从缓存中删除
			cacheStore.Delete(cacheKey)
		}

		// 使用 gin 提供的 ResponseRecorder 捕获响应
		responseBody := new(bytes.Buffer)
		writer := io.MultiWriter(c.Writer, responseBody)
		c.Writer = &ResponseRecorder{c.Writer, writer}

		// 继续处理请求
		c.Next()

		// 缓存响应数据
		cacheStore.Store(cacheKey, CacheEntry{
			ResponseBody: responseBody.Bytes(),
			CreatedAt:    time.Now(),
		})
	}
}
