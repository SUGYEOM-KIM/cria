package main

import (
	"embed"

	"cria/internal/logging"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logging.Init()
	defer logging.Close()
	logging.Userf("process start; log file at %s", logging.Path())

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "cria",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		logging.Errorf("wails.Run returned error: %v", err)
		println("Error:", err.Error())
	}
	logging.Userf("process exit")
}
