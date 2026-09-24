package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/interpreters"
)

// Interpreters exposes interpreter CRUD and discovery APIs.
type Interpreters struct {
	Svc *interpreters.Service
}

// RegisterInterpreters mounts interpreter routes on r.
func RegisterInterpreters(r *gin.Engine, svc *interpreters.Service) {
	h := Interpreters{Svc: svc}
	g := r.Group("/api/interpreters")
	g.GET("", h.List)
	g.GET("/discover", h.Discover)
	g.POST("/batch", h.BatchCreate)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.PUT("/:id/path", h.UpdatePath)
	g.PUT("/:id/default", h.SetDefault)
	g.DELETE("/:id", h.Delete)
}

func (h Interpreters) List(c *gin.Context) {
	items, err := h.Svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []interpreters.Interpreter{}
	}
	c.JSON(http.StatusOK, gin.H{"interpreters": items})
}

func (h Interpreters) Discover(c *gin.Context) {
	lang := c.Query("language")
	items, err := interpreters.Discover(lang)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []interpreters.DiscoverCandidate{}
	}
	c.JSON(http.StatusOK, gin.H{"candidates": items})
}

type interpreterBody struct {
	Language    string   `json:"language"`
	Path        string   `json:"path"`
	DefaultArgs []string `json:"defaultArgs"`
	Version     string   `json:"version"`
}

func (h Interpreters) Create(c *gin.Context) {
	var body interpreterBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	item, err := h.Svc.Create(interpreters.SaveInput{
		Language:    body.Language,
		Path:        body.Path,
		DefaultArgs: body.DefaultArgs,
	})
	if err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h Interpreters) Update(c *gin.Context) {
	id := c.Param("id")
	var body interpreterBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	item, err := h.Svc.Update(id, interpreters.SaveInput{
		Language:    body.Language,
		Path:        body.Path,
		DefaultArgs: body.DefaultArgs,
	})
	if err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

type pathBody struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type defaultBody struct {
	IsDefault bool `json:"isDefault"`
}

func (h Interpreters) SetDefault(c *gin.Context) {
	id := c.Param("id")
	var body defaultBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	item, err := h.Svc.SetDefault(id, body.IsDefault)
	if err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h Interpreters) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Svc.Delete(id); err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h Interpreters) UpdatePath(c *gin.Context) {
	id := c.Param("id")
	var body pathBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	item, err := h.Svc.UpdatePath(id, body.Path, body.Version)
	if err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

type batchBody struct {
	Items []interpreterBody `json:"items"`
}

func (h Interpreters) BatchCreate(c *gin.Context) {
	var body batchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	items := make([]interpreters.BatchItem, 0, len(body.Items))
	for _, it := range body.Items {
		items = append(items, interpreters.BatchItem{
			Language:    it.Language,
			Path:        it.Path,
			DefaultArgs: it.DefaultArgs,
			Version:     it.Version,
		})
	}
	result, err := h.Svc.BatchCreate(items)
	if err != nil {
		writeInterpreterError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func writeInterpreterError(c *gin.Context, err error) {
	if errors.Is(err, interpreters.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "interpreter not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
