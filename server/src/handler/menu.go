package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/menu"
)

// Menu exposes loaded menu items as JSON.
type Menu struct {
	Items []menu.Item
}

// RegisterMenu mounts GET /api/menu.
func RegisterMenu(r *gin.Engine, items []menu.Item) {
	h := Menu{Items: items}
	r.GET("/api/menu", h.Get)
}

func (h Menu) Get(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.Items})
}
