package main

import (
	"embed"
	"log"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
	// "github.com/wailsapp/wails/v3/pkg/options"
	// "github.com/wailsapp/wails/v3/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := application.New(application.Options{
		Name: "rlbotgui",
		Services: []application.Service{
			application.NewService(NewApp()),
		},
		LogLevel: slog.LevelWarn,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	// Create application with options
	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:     "RLBotGUI",
		Width:     1300,
		Height:    870,
		MinWidth:  725,
		MinHeight: 600,
		// AssetServer: &assetserver.Options{
		// 	Assets: assets,
		// },
		BackgroundColour: application.NewRGBA(27, 38, 54, 1),
		// Bind: []interface{}{
		// 	app,
		// 	&HumanInfo{},
		// 	&PsyonixBotInfo{},
		// 	&BotInfo{},
		// 	*Player,
		// },
	})

	// go func() {
	// 	for {
	// 		now := time.Now().Format(time.RFC1123)
	// 		app.EmitEvent("time", now)
	// 		time.Sleep(time.Second)
	// 	}
	// }()

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
