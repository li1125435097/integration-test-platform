package handler

import (
	"compress/gzip"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

var gzipExts = map[string]struct{}{
	"":      {},
	".html": {},
	".js":   {},
	".mjs":  {},
	".css":  {},
	".json": {},
	".svg":  {},
	".txt":  {},
	".xml":  {},
	".map":  {},
	".wasm": {},
}

func gzipStatic() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodHead || !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}
		ext := strings.ToLower(path.Ext(c.Request.URL.Path))
		if _, ok := gzipExts[ext]; !ok {
			c.Next()
			return
		}

		orig := c.Writer
		gz := gzip.NewWriter(orig)
		wrapped := &gzipResponseWriter{ResponseWriter: orig, gz: gz}
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = wrapped
		c.Next()
		if wrapped.started {
			_ = gz.Close()
			return
		}
		c.Writer.Header().Del("Content-Encoding")
	}
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	gz      *gzip.Writer
	started bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if code == http.StatusNoContent || code == http.StatusNotModified {
		w.Header().Del("Content-Encoding")
		w.ResponseWriter.WriteHeader(code)
		return
	}
	w.Header().Del("Content-Length")
	w.started = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.ResponseWriter.Written() {
		w.WriteHeader(http.StatusOK)
	}
	status := w.Status()
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return 0, nil
	}
	w.started = true
	return w.gz.Write(data)
}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
