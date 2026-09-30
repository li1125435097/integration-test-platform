package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/kernelplans"
)

// KernelPlans exposes saved kernel-test form plans.
type KernelPlans struct {
	Svc *kernelplans.Service
}

// RegisterKernelPlans mounts /api/kernel-test-plans.
func RegisterKernelPlans(r *gin.Engine, svc *kernelplans.Service) {
	h := KernelPlans{Svc: svc}
	g := r.Group("/api/kernel-test-plans")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

func (h KernelPlans) List(c *gin.Context) {
	items, err := h.Svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []kernelplans.Plan{}
	}
	c.JSON(http.StatusOK, gin.H{"plans": items})
}

func (h KernelPlans) Get(c *gin.Context) {
	plan, err := h.Svc.Get(c.Param("id"))
	if err != nil {
		writeKernelPlanError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

type kernelPlanBody struct {
	Name string           `json:"name"`
	Form kernelplans.Form `json:"form"`
}

func (h KernelPlans) Create(c *gin.Context) {
	var body kernelPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	plan, err := h.Svc.Create(body.Name, body.Form)
	if err != nil {
		writeKernelPlanError(c, err)
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h KernelPlans) Update(c *gin.Context) {
	var body kernelPlanBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	plan, err := h.Svc.Update(c.Param("id"), body.Name, body.Form)
	if err != nil {
		writeKernelPlanError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h KernelPlans) Delete(c *gin.Context) {
	if err := h.Svc.Delete(c.Param("id")); err != nil {
		writeKernelPlanError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func writeKernelPlanError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, kernelplans.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, kernelplans.ErrNameExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, kernelplans.ErrEmptyName):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
