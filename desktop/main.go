package main

import (
	"embed"
	"log"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	wapp := application.New(application.Options{
		Name:     "kirokit",
		Services: []application.Service{application.NewService(app)},
		Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
		// v3 keeps a macOS app running with no window; kirokit quits as before.
		Mac: application.MacOptions{ApplicationShouldTerminateAfterLastWindowClosed: true},
	})
	app.setEmitter(func(event string, data any) { wapp.Event.Emit(event, data) })

	wapp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "kirokit",
		Width:            1180,
		Height:           760,
		MinWidth:         720,
		MinHeight:        480,
		BackgroundColour: application.NewRGB(23, 19, 31),
		// Windows and Linux show the application menu at the top of the window.
		UseApplicationMenu: true,
	})
	wapp.Menu.Set(buildMenu(wapp, menuFor(runtime.GOOS), app))

	// A smoke start (CI) quits once the page has loaded and its first State
	// call has answered (the page then sends smoke:ready), or fails after a minute.
	if os.Getenv("KIROKIT_DESKTOP_SMOKE") == "1" {
		ready := smokeWatch(smokeTimeout, wapp.Quit, func() {
			log.Print("smoke start: the page did not answer within ", smokeTimeout)
			os.Exit(1)
		})
		wapp.Event.On("smoke:ready", func(*application.CustomEvent) { ready() })
	}
	if err := wapp.Run(); err != nil {
		log.Fatal(err)
	}
}
