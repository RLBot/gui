package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ncruces/zenity"
	rlbot "github.com/swz-git/go-interface"
	"github.com/swz-git/go-interface/flat"
)

type RawReleaseInfo struct {
	repo    string
	content GhRelease
}

// App struct
type App struct {
	latest_release_json []RawReleaseInfo
	rlbot_address       string
}

func (a *App) IgnoreMe(
	_ BotInfo,
	_ PsyonixBotInfo,
	_ HumanInfo,
) {
}

func (a *App) GetDefaultPath() string {
	if runtime.GOOS == "windows" {
		localappdata := os.Getenv("LOCALAPPDATA")
		return filepath.Join(localappdata, "RLBotGUI")
	}

	// assume linux

	xdg_data_home := os.Getenv("XDG_DATA_HOME")
	if xdg_data_home == "" {
		home := os.Getenv("HOME")
		xdg_data_home = filepath.Join(home, ".local/share")
	}

	return filepath.Join(xdg_data_home, "rlbotgui")
}

func (a *App) DownloadBotpack(repo string, installPath string) (string, error) {
	latest_release, err := a.GetLatestReleaseData(repo)
	if err != nil {
		return "", err
	}

	var file_name string
	if runtime.GOOS == "windows" {
		file_name = "botpack_x86_64-windows.tar.xz"
	} else {
		file_name = "botpack_x86_64-linux.tar.xz"
	}

	var download_url string
	for _, asset := range latest_release.Assets {
		if asset.Name == file_name {
			download_url = asset.BrowserDownloadURL
			break
		}
	}

	err = DownloadExtractArchive(download_url, installPath)
	if err != nil {
		return "", err
	}

	return latest_release.TagName, nil
}

func (a *App) RepairBotpack(repo string, installPath string) (string, error) {
	err := os.RemoveAll(installPath)
	if err != nil {
		return "", err
	}

	return a.DownloadBotpack(repo, installPath)
}

// NewApp creates a new App application struct
func NewApp() *App {
	ip := os.Getenv("RLBOT_SERVER_IP")
	if ip == "" {
		ip = "127.0.0.1"
	}

	port := os.Getenv("RLBOT_SERVER_PORT")
	if port == "" {
		port = "23234"
	}

	rlbot_address := ip + ":" + port

	var latest_release_json []RawReleaseInfo
	return &App{
		latest_release_json,
		rlbot_address,
	}
}

func recursiveTomlSearch(root, tomlType string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(info.Name()) != ".toml" {
			return nil
		}

		if info.Name() == tomlType+".toml" || strings.HasSuffix(info.Name(), "."+tomlType+".toml") {
			matches = append(matches, path)
		}

		return nil
	})
	return matches, err
}

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ExtraOptions struct {
	Freeplay              bool `json:"freeplay"`
	EnableRendering       bool `json:"enableRendering"`
	EnableStateSetting    bool `json:"enableStateSetting"`
	InstantStart          bool `json:"instantStart"`
	SkipReplays           bool `json:"skipReplays"`
	AutoSaveReplay        bool `json:"autoSaveReplay"`
	ExistingMatchBehavior byte `json:"existingMatchBehavior"`
	AutoStartAgents       bool `json:"autoStartAgents"`
	WaitForAgents         bool `json:"waitForAgents"`
}

type StartMatchOptions struct {
	Map             string                `json:"map"`
	GameMode        string                `json:"gameMode"`
	Scripts         []BotInfo             `json:"scripts"`
	BluePlayers     []PlayerJs            `json:"bluePlayers"`
	OrangePlayers   []PlayerJs            `json:"orangePlayers"`
	MutatorSettings flat.MutatorSettingsT `json:"mutatorSettings"`
	ExtraOptions    ExtraOptions          `json:"extraOptions"`
	Launcher        string                `json:"launcher"`
	LauncherArg     string                `json:"launcherArg"`
}

func (a *App) StartMatch(options StartMatchOptions) Result {
	// TODO: Save this in App struct
	conn, err := rlbot.Connect(a.rlbot_address)
	if err != nil {
		return Result{false, "Failed to connect to RLBotServer at " + a.rlbot_address}
	}

	var gameMode flat.GameMode
	switch options.GameMode {
	case "Soccer":
		gameMode = flat.GameModeSoccer
	case "Hoops":
		gameMode = flat.GameModeHoops
	case "Dropshot":
		gameMode = flat.GameModeDropshot
	case "Hockey":
		gameMode = flat.GameModeHockey
	case "Rumble":
		gameMode = flat.GameModeRumble
	case "Heatseeker":
		gameMode = flat.GameModeHeatseeker
	case "Gridiron":
		gameMode = flat.GameModeGridiron
	case "Knockout":
		gameMode = flat.GameModeKnockout
	default:
		println("No mode chosen, defaulting to soccer")
		gameMode = flat.GameModeSoccer
	}

	var launcher flat.Launcher
	switch options.Launcher {
	case "steam":
		launcher = flat.LauncherSteam
	case "epic":
		launcher = flat.LauncherEpic
	case "custom":
		launcher = flat.LauncherCustom
	case "nolaunch":
		launcher = flat.LauncherNoLaunch
	default:
		println("No launcher chosen, defaulting to NoLaunch")
		launcher = flat.LauncherNoLaunch
	}

	playerConfigs := make([]*flat.PlayerConfigurationT, len(options.BluePlayers)+len(options.OrangePlayers))

	for i, playerInfo := range options.BluePlayers {
		playerConfigs[i] = playerInfo.ToPlayer().ToPlayerConfig(0)
	}

	for i, playerInfo := range options.OrangePlayers {
		playerConfigs[i+len(options.BluePlayers)] = playerInfo.ToPlayer().ToPlayerConfig(1)
	}

	scriptConfigs := make([]*flat.ScriptConfigurationT, len(options.Scripts))
	for i, info := range options.Scripts {
		scriptConfigs[i] = info.ToScriptConfig()
	}

	match := flat.MatchConfigurationT{
		AutoStartAgents:       options.ExtraOptions.AutoStartAgents,
		WaitForAgents:         options.ExtraOptions.WaitForAgents,
		GameMapUpk:            options.Map,
		PlayerConfigurations:  playerConfigs,
		ScriptConfigurations:  scriptConfigs,
		GameMode:              gameMode,
		Mutators:              &options.MutatorSettings,
		Freeplay:              options.ExtraOptions.Freeplay,
		EnableRendering:       options.ExtraOptions.EnableRendering,
		EnableStateSetting:    options.ExtraOptions.EnableStateSetting,
		InstantStart:          options.ExtraOptions.InstantStart,
		SkipReplays:           options.ExtraOptions.SkipReplays,
		AutoSaveReplay:        options.ExtraOptions.AutoSaveReplay,
		Launcher:              launcher,
		LauncherArg:           options.LauncherArg,
		ExistingMatchBehavior: flat.ExistingMatchBehavior(options.ExtraOptions.ExistingMatchBehavior),
	}

	conn.SendPacket(&match)
	conn.SendPacket(nil) // Tell core that we want to disconnect

	return Result{true, ""}
}

func (a *App) StopMatch(shutdownServer bool) Result {
	// TODO: Save this in App struct
	// TODO: Make dynamic, pull from env var?
	conn, err := rlbot.Connect(a.rlbot_address)
	if err != nil {
		return Result{false, "Failed to connect to rlbot"}
	}

	conn.SendPacket(&flat.StopCommandT{
		ShutdownServer: shutdownServer,
	})
	conn.SendPacket(nil) // Tell core that we want to disconnect

	return Result{true, ""}
}

func (a *App) PickFolder() string {
	path, err := zenity.SelectFile(zenity.Directory())
	if err != nil {
		println("ERR: File picker failed")
	}
	return path
}

func (a *App) PickRLBotToml() (string, error) {
	path, err := zenity.SelectFile(zenity.FileFilter{
		Name:     ".toml files",
		Patterns: []string{"*.toml"},
	})
	if err != nil {
		return "", nil
	}

	filename := filepath.Base(path)
	if filename == "bot.toml" || filename == "script.toml" || strings.HasSuffix(filename, ".bot.toml") || strings.HasSuffix(filename, ".script.toml") {
		return path, nil
	}

	return "", fmt.Errorf("invalid file name")
}

func (a *App) ShowPathInExplorer(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	// if is dir
	var folder string
	if fileInfo.IsDir() {
		folder = path
	} else {
		folder = filepath.Dir(path)
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("explorer.exe", folder)
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		cmd := exec.Command("xdg-open", folder)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
