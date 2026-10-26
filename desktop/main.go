package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	opts := &options.App{
		Title:     "kirokit",
		Width:     1180,
		Height:    760,
		MinWidth:  720,
		MinHeight: 480,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 23, G: 19, B: 31, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []any{app},
	}
	// A smoke start (CI) quits once the page has loaded, or fails after a minute.
	if os.Getenv("KIROKIT_DESKTOP_SMOKE") == "1" {
		opts.OnDomReady = smokeWatch(smokeTimeout, runtime.Quit, func() {
			log.Print("smoke start: the page did not load within ", smokeTimeout)
			os.Exit(1)
		})
	}
	if err := wails.Run(opts); err != nil {
		log.Fatal(err)
	}
}
