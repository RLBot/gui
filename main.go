package main

import (
	"embed"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	// "github.com/wailsapp/wails/v3/pkg/options"
	// "github.com/wailsapp/wails/v3/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func check_nvidia() bool {
	cmd := exec.Command("nvidia-smi")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// nvidia-smi likely not installed or not in PATH
		return false
	}

	// Check if the output contains signs of NVIDIA GPU presence.
	// This is a basic check; more robust parsing of the output is possible.
	return strings.Contains(string(output), "NVIDIA")
}

func main() {
	// see https://github.com/tauri-apps/tauri/issues/9394
	if check_nvidia() {
		os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}

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
		MinWidth:  650,
		MinHeight: 400,
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
