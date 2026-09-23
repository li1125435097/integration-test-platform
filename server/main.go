package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"integration-test-platform/server/src/assets"
	"integration-test-platform/server/src/config"
	"integration-test-platform/server/src/handler"
	"integration-test-platform/server/src/menu"
	"integration-test-platform/server/src/paths"
)

var version = "0.1.0-dev"

func main() {
	cfg := config.Parse(version)
	if cfg.ShowVer {
		fmt.Println(version)
		return
	}

	configDir, err := paths.ResolveConfigDir(cfg.ConfigDir)
	if err != nil {
		log.Fatalf("config dir: %v", err)
	}
	menuPath := filepath.Join(configDir, "menu.json")

	webAssets, err := assets.New()
	if err != nil {
		log.Fatalf("web assets: %v", err)
	}

	menuFile, err := menu.Load(menuPath)
	if err != nil {
		log.Fatalf("menu: %v", err)
	}
	if err := menuFile.Validate(webAssets.PageExists); err != nil {
		log.Fatalf("menu validate: %v", err)
	}

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	handler.RegisterMenu(r, menuFile.Items)

	r.GET("/", func(c *gin.Context) {
		if disk := webAssets.DiskRoot(); disk != "" {
			c.File(filepath.Join(disk, "index.html"))
			return
		}
		data, err := fs.ReadFile(webAssets.IORoot(), "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "index not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
	r.StaticFS("/static", webAssets.SubFS("static"))
	r.StaticFS("/pages", webAssets.SubFS("pages"))

	log.Printf("integration-test-platform %s listening on %s (config: %s)", version, cfg.Addr, configDir)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
