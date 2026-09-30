package handler

import (
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/assets"
)

const (
	// Vite fingerprints /assets filenames, so they can be cached until the binary changes.
	cacheControlHashed = "public, max-age=31536000, immutable"
	// HTML must revalidate so a new package's hashed asset names are picked up.
	cacheControlHTML = "no-cache"
)

// RegisterWeb serves the SPA: hashed assets with long cache, index.html always revalidated.
func RegisterWeb(r *gin.Engine, web *assets.Web) {
	r.GET("/", gzipStatic(), serveIndex(web))

	g := r.Group("/assets")
	g.Use(gzipStatic(), cacheHashedAssets())
	g.StaticFS("/", web.SubFS("assets"))
}

func cacheHashedAssets() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", cacheControlHashed)
		c.Next()
	}
}

func serveIndex(web *assets.Web) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", cacheControlHTML)
		if disk := web.DiskRoot(); disk != "" {
			c.File(filepath.Join(disk, "index.html"))
			return
		}
		data, err := fs.ReadFile(web.IORoot(), "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
}
