package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/kernels"
)

// RegisterKernels mounts GET /api/kernels.
func RegisterKernels(r *gin.Engine) {
	r.GET("/api/kernels", listKernels)
}

func listKernels(c *gin.Context) {
	items, err := kernels.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []kernels.Kernel{}
	}
	c.JSON(http.StatusOK, gin.H{"kernels": items})
}
