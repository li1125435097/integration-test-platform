package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/fingerprints"
)

// RegisterFingerprints mounts fingerprint encrypt, decrypt, and file routes.
func RegisterFingerprints(r *gin.Engine, svc *fingerprints.Service) {
	g := r.Group("/api/fingerprints")
	g.POST("/encrypt", func(c *gin.Context) { encryptFingerprint(c, svc) })
	g.POST("/decrypt", func(c *gin.Context) { decryptFingerprint(c, svc) })
	g.GET("", func(c *gin.Context) { listFingerprints(c, svc) })
	g.POST("", func(c *gin.Context) { saveFingerprint(c, svc) })
	g.GET("/:name", func(c *gin.Context) { readFingerprint(c, svc) })
}

func encryptFingerprint(c *gin.Context, svc *fingerprints.Service) {
	var body struct {
		Plaintext string `json:"plaintext"`
		BaseURL   string `json:"baseUrl"`
		Bearer    string `json:"bearer"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	encoded, err := svc.Encrypt(body.Plaintext, body.BaseURL, body.Bearer)
	if err != nil {
		writeFingerprintErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contentBase64": encoded})
}

func decryptFingerprint(c *gin.Context, svc *fingerprints.Service) {
	var body struct {
		ContentBase64 string `json:"contentBase64"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	text, err := svc.Decrypt(body.ContentBase64)
	if err != nil {
		writeFingerprintErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plaintext": text})
}

func listFingerprints(c *gin.Context, svc *fingerprints.Service) {
	items, err := svc.List()
	if err != nil {
		writeFingerprintErr(c, err)
		return
	}
	if items == nil {
		items = []fingerprints.FileInfo{}
	}
	c.JSON(http.StatusOK, gin.H{"files": items})
}

func saveFingerprint(c *gin.Context, svc *fingerprints.Service) {
	var body struct {
		Name          string `json:"name"`
		ContentBase64 string `json:"contentBase64"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body.ContentBase64))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密文不是合法 Base64"})
		return
	}
	if err := svc.Save(body.Name, raw); err != nil {
		writeFingerprintErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": body.Name})
}

func readFingerprint(c *gin.Context, svc *fingerprints.Service) {
	name := c.Param("name")
	data, err := svc.Read(name)
	if err != nil {
		writeFingerprintErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"name":          name,
		"contentBase64": base64.StdEncoding.EncodeToString(data),
	})
}

func writeFingerprintErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, fingerprints.ErrScriptNotFound), errors.Is(err, fingerprints.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, fingerprints.ErrInvalidName), errors.Is(err, fingerprints.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
