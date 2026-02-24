package main

import (
	"embed"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	// "github.com/wailsapp/wails/v3/pkg/options"
	// "github.com/wailsapp/wails/v3/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func checkNvidia() bool {
	cmd := exec.Command("nvidia-smi")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// nvidia-smi likely not installed or not in PATH
		return false
	}

	return strings.Contains(string(output), "NVIDIA")
}

func main() {
	// see https://github.com/tauri-apps/tauri/issues/9394
	if checkNvidia() {
		os.Setenv("__NV_DISABLE_EXPLICIT_SYNC", "1")
	}

	// wails sets this to "1", but that breaks blur, and we already work around
	// this using the variable above.
	// TODO: file issue in the wails repo
	os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "0")

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
		// Not possible on linux unfortunately
		// TODO: Another reason to switch away from localstorage
		Windows: application.WindowsOptions{
			WebviewUserDataPath: filepath.Join(os.Getenv("LOCALAPPDATA"), "RLBot5/webviewdata"),
		},
	})

	// Create application with options
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "RLBotGUI",
		Width:     1100,
		Height:    800,
		MinWidth:  650,
		MinHeight: 400,
		Frameless: true,
		Windows: application.WindowsWindow{
	    DisableFramelessWindowDecorations: false,
			HiddenOnTaskbar: false,
		},
		BackgroundType: application.BackgroundTypeTransparent,
		// BackgroundColour: application.NewRGBA(27, 38, 54, 1),
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
