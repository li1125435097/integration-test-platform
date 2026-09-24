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
	"integration-test-platform/server/src/interpreters"
	"integration-test-platform/server/src/menu"
	"integration-test-platform/server/src/paths"
	"integration-test-platform/server/src/scripts"
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
	if err := menuFile.Validate(); err != nil {
		log.Fatalf("menu validate: %v", err)
	}

	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	handler.RegisterMenu(r, menuFile.Items)

	dataDir := paths.DataDir(configDir)
	scriptSvc := scripts.NewService(dataDir)
	if err := scriptSvc.EnsureDataDir(); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	handler.RegisterScripts(r, scriptSvc)

	interpSvc := interpreters.NewService(dataDir)
	if err := interpSvc.EnsureDataDir(); err != nil {
		log.Fatalf("interpreters data: %v", err)
	}
	handler.RegisterInterpreters(r, interpSvc)

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
	r.StaticFS("/assets", webAssets.SubFS("assets"))

	log.Printf("integration-test-platform %s listening on %s (config: %s)", version, cfg.Addr, configDir)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
