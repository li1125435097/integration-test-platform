package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/fingerprints"
	"integration-test-platform/server/src/kerneltest"
	"integration-test-platform/server/src/scripts"
)

// RegisterKernelTests mounts POST /api/kernel-tests/run.
func RegisterKernelTests(r *gin.Engine, svc *kerneltest.Service) {
	r.POST("/api/kernel-tests/run", func(c *gin.Context) {
		runKernelTest(c, svc)
	})
}

func runKernelTest(c *gin.Context, svc *kerneltest.Service) {
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kernel test is not configured"})
		return
	}
	var body struct {
		ScriptID    string   `json:"scriptId"`
		Concurrency int      `json:"concurrency"`
		Kernels     []string `json:"kernels"`
		Args        []string `json:"args"`
		Fingerprint struct {
			SavedName     string `json:"savedName"`
			ContentBase64 string `json:"contentBase64"`
			Plaintext     string `json:"plaintext"`
			BaseURL       string `json:"baseUrl"`
			Bearer        string `json:"bearer"`
		} `json:"fingerprint"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	out, err := svc.Run(kerneltest.Input{
		ScriptID:    body.ScriptID,
		Concurrency: body.Concurrency,
		KernelPaths: body.Kernels,
		Args:        body.Args,
		Fingerprint: kerneltest.FingerprintInput{
			SavedName:     body.Fingerprint.SavedName,
			ContentBase64: body.Fingerprint.ContentBase64,
			Plaintext:     body.Fingerprint.Plaintext,
			BaseURL:       body.Fingerprint.BaseURL,
			Bearer:        body.Fingerprint.Bearer,
		},
	})
	if err != nil {
		writeKernelTestErr(c, err)
		return
	}
	if out.Results == nil {
		out.Results = []kerneltest.Result{}
	}
	c.JSON(http.StatusOK, out)
}

func writeKernelTestErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, scripts.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
	case errors.Is(err, fingerprints.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, kerneltest.ErrNoScript),
		errors.Is(err, kerneltest.ErrNoKernel),
		errors.Is(err, kerneltest.ErrScriptPrefix),
		errors.Is(err, kerneltest.ErrUnknownKernel),
		errors.Is(err, kerneltest.ErrFingerprint),
		errors.Is(err, fingerprints.ErrInvalidName),
		errors.Is(err, fingerprints.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
