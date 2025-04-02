package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paranoiachains/metrics/internal/logger"
	"go.uber.org/zap"
)

func LoggerMiddleware() gin.HandlerFunc {
	logger.Initialize()
	defer logger.Log.Sync()
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		encoding := c.Request.Header.Get("Accept-Encoding")

		c.Next()

		duration := time.Since(start)

		logger.Log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Duration("duration", duration),
			zap.String("Accept-Encoding", encoding),
		)
		logger.Log.Info("HTTP Response",
			zap.Int("status", c.Writer.Status()),
			zap.String("Content-Type", c.Writer.Header().Get("Content-Type")),
		)
	}
}

func shouldCompress(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
}

func shouldDecompress(req *http.Request) bool {
	return strings.Contains(req.Header.Get("Content-Encoding"), "gzip")
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldDecompress(c.Request) {
			r, _ := gzip.NewReader(c.Request.Body)
			defer r.Close()

			body, _ := io.ReadAll(r)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Request.Header.Del("Content-Encoding")
			logger.Log.Info("gzip", zap.Bool("decompressed", true))
		}
		if !shouldCompress(c.Request) {
			return
		}
		gz := gzip.NewWriter(c.Writer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Accept-Encoding", "gzip")
		c.Writer = &gzipWriter{c.Writer, gz}
		defer func() {
			c.Header("Content-Length", "0")
			gz.Close()
		}()
		logger.Log.Info("gzip", zap.Bool("compressed", true), zap.Int("size", c.Writer.Size()))
		c.Next()
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}
