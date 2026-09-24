package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/scriptexec"
	"integration-test-platform/server/src/scripts"
)

// Scripts exposes script CRUD and version APIs.
type Scripts struct {
	Svc    *scripts.Service
	Runner *scriptexec.Runner
}

// RegisterScripts mounts script routes on r.
func RegisterScripts(r *gin.Engine, svc *scripts.Service, runner *scriptexec.Runner) {
	h := Scripts{Svc: svc, Runner: runner}
	g := r.Group("/api/scripts")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.POST("/run-preview", h.RunPreview)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.GET("/:id/versions", h.ListVersions)
	g.POST("/:id/versions", h.AddVersion)
	g.PUT("/:id/versions/:versionId", h.UpdateVersionRemark)
	g.DELETE("/:id/versions/:versionId", h.DeleteVersion)
	g.POST("/:id/restore", h.Restore)
	g.POST("/:id/run", h.RunSaved)
}

type runPreviewBody struct {
	Language      string `json:"language"`
	InterpreterID string `json:"interpreterId"`
	Content       string `json:"content"`
}

func (h Scripts) RunPreview(c *gin.Context) {
	if h.Runner == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "runner not configured"})
		return
	}
	var body runPreviewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if body.Language == "" {
		body.Language = "javascript"
	}
	out := h.Runner.RunPreview(scriptexec.PreviewInput{
		Language:      body.Language,
		InterpreterID: body.InterpreterID,
		Content:       body.Content,
	})
	c.JSON(http.StatusOK, out)
}

func (h Scripts) RunSaved(c *gin.Context) {
	if h.Runner == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "runner not configured"})
		return
	}
	id := c.Param("id")
	out := h.Runner.RunSaved(id)
	if out.Error != "" {
		switch out.Error {
		case scripts.ErrNotFound.Error():
			c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
			return
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h Scripts) List(c *gin.Context) {
	items, err := h.Svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []scripts.ListItem{}
	}
	c.JSON(http.StatusOK, gin.H{"scripts": items})
}

type scriptBody struct {
	Name          string `json:"name"`
	Description     string `json:"description"`
	Language        string `json:"language"`
	Content         string `json:"content"`
	InterpreterID   string `json:"interpreterId"`
}

func (h Scripts) Create(c *gin.Context) {
	var body scriptBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if body.Language == "" {
		body.Language = "javascript"
	}
	detail, err := h.Svc.Create(scripts.CreateInput{
		Name:          body.Name,
		Description:     body.Description,
		Language:        body.Language,
		Content:         body.Content,
		InterpreterID:   body.InterpreterID,
	})
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusCreated, detail)
}

func (h Scripts) Get(c *gin.Context) {
	id := c.Param("id")
	detail, err := h.Svc.Get(id)
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h Scripts) Update(c *gin.Context) {
	id := c.Param("id")
	var body scriptBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if body.Language == "" {
		body.Language = "javascript"
	}
	detail, err := h.Svc.Update(id, scripts.UpdateInput{
		Name:          body.Name,
		Description:     body.Description,
		Language:        body.Language,
		Content:         body.Content,
		InterpreterID:   body.InterpreterID,
	})
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h Scripts) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Svc.Delete(id); err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h Scripts) ListVersions(c *gin.Context) {
	id := c.Param("id")
	vers, err := h.Svc.ListVersions(id)
	if err != nil {
		writeScriptError(c, err)
		return
	}
	if vers == nil {
		vers = []scripts.Version{}
	}
	c.JSON(http.StatusOK, gin.H{"versions": vers})
}

type versionBody struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

func (h Scripts) AddVersion(c *gin.Context) {
	id := c.Param("id")
	var body versionBody
	_ = c.ShouldBindJSON(&body)
	ver, err := h.Svc.AddVersion(id, body.Name, body.Remark)
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ver)
}

type versionRemarkBody struct {
	Remark string `json:"remark"`
}

func (h Scripts) UpdateVersionRemark(c *gin.Context) {
	id := c.Param("id")
	versionID := c.Param("versionId")
	var body versionRemarkBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	ver, err := h.Svc.UpdateVersionRemark(id, versionID, body.Remark)
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, ver)
}

func (h Scripts) DeleteVersion(c *gin.Context) {
	id := c.Param("id")
	versionID := c.Param("versionId")
	if err := h.Svc.DeleteVersion(id, versionID); err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type restoreBody struct {
	VersionID string `json:"versionId"`
}

func (h Scripts) Restore(c *gin.Context) {
	id := c.Param("id")
	var body restoreBody
	if err := c.ShouldBindJSON(&body); err != nil || body.VersionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "versionId is required"})
		return
	}
	item, err := h.Svc.Restore(id, body.VersionID)
	if err != nil {
		writeScriptError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func writeScriptError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, scripts.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
	case errors.Is(err, scripts.ErrVersionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
