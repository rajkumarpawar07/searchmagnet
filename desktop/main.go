// Command desktop runs the SearchMagnet desktop application built with Wails.
//
// The app reuses the core SearchMagnet packages:
//   - internal/embedder — Gemini multimodal embedding generation
//   - internal/store    — local vector index (JSON backend)
//
// The frontend lives in frontend/dist and is embedded into the binary.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "SearchMagnet",
		Width:     1280,
		Height:    832,
		MinWidth:  1024,
		MinHeight: 640,

		AssetServer: &assetserver.Options{
			Assets: assets,
			Handler: app, // serves /preview?path=... for indexed files
		},

		BackgroundColour: &options.RGBA{R: 10, G: 8, B: 26, A: 1},

		OnStartup:  app.startup,
		OnShutdown: app.shutdown,

		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
		},

		Bind: []interface{}{app},

		Windows: &windows.Options{
			Theme: windows.Dark,
		},
	})
	if err != nil {
		panic(err)
	}
}
