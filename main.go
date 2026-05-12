package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	initLog()

	app := application.New(application.Options{
		Name:        "Mise Manager",
		Description: "A visual management tool for mise",
		Services: []application.Service{
			application.NewService(&MiseService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Mise Manager",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(30, 30, 30),
		URL:              "/",
		Width:            900,
		Height:           600,
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
