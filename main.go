package main

import (
	"embed"

	"claude_desktop/backend/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	_ "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the newApp structure
	newApp := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "claude_desktop",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        newApp.Startup,
		OnDomReady:       newApp.DomReady,
		OnShutdown:       newApp.Shutdown,
		Bind: []interface{}{
			newApp,
		},
		LogLevel: logger.DEBUG,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
