package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/executions"
)

// ExecutionRecords exposes execution history APIs.
type ExecutionRecords struct {
	Svc *executions.Service
}

// RegisterExecutionRecords mounts execution record routes on r.
func RegisterExecutionRecords(r *gin.Engine, svc *executions.Service) {
	h := ExecutionRecords{Svc: svc}
	g := r.Group("/api/execution-records")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}

func (h ExecutionRecords) List(c *gin.Context) {
	items, err := h.Svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []executions.Record{}
	}
	c.JSON(http.StatusOK, gin.H{"records": items})
}

func (h ExecutionRecords) Get(c *gin.Context) {
	id := c.Param("id")
	rec, err := h.Svc.Get(id)
	if err != nil {
		writeExecutionError(c, err)
		return
	}
	c.JSON(http.StatusOK, rec)
}

func writeExecutionError(c *gin.Context, err error) {
	if errors.Is(err, executions.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution record not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
