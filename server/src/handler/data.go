package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/dataio"
)

// Data exposes data-directory export and import.
type Data struct {
	Svc *dataio.Service
}

// RegisterData mounts data package routes on r.
func RegisterData(r *gin.Engine, svc *dataio.Service) {
	h := Data{Svc: svc}
	g := r.Group("/api/data")
	g.GET("/export", h.Export)
	g.POST("/import", h.Import)
}

func (h Data) Export(c *gin.Context) {
	name := "itp-data-" + time.Now().Format("20060102-150405") + ".zip"
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	if err := h.Svc.Export(c.Writer); err != nil && !c.Writer.Written() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h Data) Import(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, dataio.MaxImportBytes+(1<<20))
	file, err := c.FormFile("file")
	if err != nil {
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": dataio.ErrTooLarge.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "需要上传 zip 文件"})
		return
	}
	if file.Size > dataio.MaxImportBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": dataio.ErrTooLarge.Error()})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取上传文件"})
		return
	}
	defer opened.Close()

	result, err := h.Svc.Import(opened)
	if err != nil {
		switch {
		case errors.Is(err, dataio.ErrTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, dataio.ErrInvalidArchive):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}
