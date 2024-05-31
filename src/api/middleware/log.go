package middleware

import (
	"bytes"
	"log"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewStatusCodeGreaterThan(code int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		w := responseWriter{
			ResponseWriter: ctx.Writer,
			b:              bytes.NewBuffer([]byte{}),
		}
		ctx.Writer = w
		ctx.Next()

		if ctx.Writer.Status() >= code {
			log.Printf("[%s] Response %s\n", ctx.Request.URL.String(), w.b.String())
		}
	}
}
