package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rlbot "github.com/RLBot/go-interface"
	"github.com/RLBot/go-interface/flat"
	"github.com/ncruces/zenity"
)

type RawReleaseInfo struct {
	repo    string
	content GhRelease
}

// App struct
type App struct {
	latestReleaseJson []RawReleaseInfo
	rlbotAddress      string
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

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home := os.Getenv("HOME")
		xdgDataHome = filepath.Join(home, ".local/share")
	}

	return filepath.Join(xdgDataHome, "rlbotgui")
}

func (a *App) DownloadBotpack(repo string, installPath string) (string, error) {
	latestRelease, err := a.GetLatestReleaseData(repo)
	if err != nil {
		return "", err
	}

	var fileName string
	if runtime.GOOS == "windows" {
		fileName = "botpack_x86_64-windows.tar.xz"
	} else {
		fileName = "botpack_x86_64-linux.tar.xz"
	}

	var downloadUrl string
	for _, asset := range latestRelease.Assets {
		if asset.Name == fileName {
			downloadUrl = asset.BrowserDownloadURL
			break
		}
	}

	err = DownloadExtractArchive(downloadUrl, installPath)
	if err != nil {
		return "", err
	}

	return latestRelease.TagName, nil
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
	Freeplay              bool                `json:"freeplay"`
	EnableRendering       flat.DebugRendering `json:"enableRendering"`
	EnableStateSetting    bool                `json:"enableStateSetting"`
	InstantStart          bool                `json:"instantStart"`
	SkipReplays           bool                `json:"skipReplays"`
	AutoSaveReplay        bool                `json:"autoSaveReplay"`
	ExistingMatchBehavior byte                `json:"existingMatchBehavior"`
	AutoStartAgents       bool                `json:"autoStartAgents"`
	WaitForAgents         bool                `json:"waitForAgents"`
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

func ReadAllMessages(conn *rlbot.RLBotConnection, packetChan chan any) {
	for {
		packet, err := conn.RecvPacket() // This is the blocking call
		if err != nil {
			packetChan <- fmt.Errorf("error receiving packet: %w", err) // Send the error to the channel
			return                                                      // Exit goroutine on error
		}
		packetChan <- packet.Value // Send the received packet to the channel

		switch packet.Type {
		case flat.CoreMessageDisconnectSignal:
			return // Exit goroutine on disconnect signal
		}
	}
}

func WaitForMatchReady(
	conn *rlbot.RLBotConnection,
	rlbotAddress string,
	matchLoadDur time.Duration,
	matchReadyDur time.Duration,
) error {
	packetChan := make(chan any)

	// Goroutine to continuously receive packets from the connection
	go ReadAllMessages(conn, packetChan)

	// First wait: for initial MatchConfigurationT and GamePacketT, or timeout (matchStartDur)
	timer1 := time.NewTimer(matchLoadDur)
	defer timer1.Stop() // Ensure timer is stopped when the function exits

	// Wait for the previous match to ended, then reconnect
	// We can then guarantee that the subsequent GamePackets are from our new match
	reconnected := false
	for !reconnected {
		select {
		case item := <-packetChan: // Receive either packet or error
			if err, ok := item.(error); ok {
				return err // Propagate the error from the goroutine
			}

			switch item.(type) {
			case *flat.DisconnectSignalT:
				conn2, err := rlbot.Connect(rlbotAddress)
				if err != nil {
					return fmt.Errorf("Failed to reconnect to RLBotServer at %s", rlbotAddress)
				}

				conn = &conn2

				// Close the connection if a new match is started in the middle of waiting for this one
				conn.SendPacket(&flat.ConnectionSettingsT{
					AgentId:              "",
					WantsBallPredictions: false,
					WantsComms:           false,
					CloseBetweenMatches:  true,
				})

				// Start reading messages from the new connection
				go ReadAllMessages(conn, packetChan)

				reconnected = true
			}
		case <-timer1.C:
			conn.SendPacket(&flat.DisconnectSignalT{})
			return fmt.Errorf("Timed out waiting for match load after %s", matchLoadDur)
		}
	}

	var gamePacket *flat.GamePacketT
	for gamePacket == nil {
		select {
		case item := <-packetChan: // Receive either packet or error
			if err, ok := item.(error); ok {
				return err // Propagate the error from the goroutine
			}

			switch packet := item.(type) {
			case *flat.FieldInfoT:
				conn.SendPacket(&flat.InitCompleteT{})
			case *flat.GamePacketT:
				gamePacket = packet
			case *flat.DisconnectSignalT:
				return fmt.Errorf("Match was ended while waiting for it to load")
			}
		case <-timer1.C:
			conn.SendPacket(&flat.DisconnectSignalT{})
			return fmt.Errorf("Timed out waiting for match load after %s", matchLoadDur)
		}
	}

	// Second wait: for GamePacketT to indicate an active match phase, or timeout (matchReadyDur)
	timer2 := time.NewTimer(matchReadyDur)
	defer timer2.Stop()

	for gamePacket.MatchInfo.MatchPhase == flat.MatchPhaseEnded ||
		gamePacket.MatchInfo.MatchPhase == flat.MatchPhaseInactive ||
		gamePacket.MatchInfo.MatchPhase == flat.MatchPhasePaused {
		select {
		case item := <-packetChan: // Receive either packet or error
			if err, ok := item.(error); ok {
				return err // Propagate the error from the goroutine
			}

			switch packet := item.(type) {
			case *flat.GamePacketT:
				gamePacket = packet
			case *flat.DisconnectSignalT:
				return fmt.Errorf("Match was ended while waiting for it to start")
			}
		case <-timer2.C:
			conn.SendPacket(&flat.DisconnectSignalT{})
			return fmt.Errorf(
				"Timed out waiting for match ready after %s",
				matchReadyDur,
			)
		}
	}

	conn.SendPacket(&flat.DisconnectSignalT{})

	return nil
}

func StartAndWaitForMatch(rlbotAddress string, match *flat.MatchConfigurationT) error {
	conn, err := rlbot.Connect(rlbotAddress)
	if err != nil {
		return fmt.Errorf("Failed to reconnect to RLBotServer at %s", rlbotAddress)
	}

	// Rely on RLBotServer closing this connection when the new match starts
	// to differentiate between the new MatchConfigurationT and the old one.
	conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "",
		WantsBallPredictions: false,
		WantsComms:           false,
		CloseBetweenMatches:  true,
	})
	conn.SendPacket(&flat.InitCompleteT{})

	conn.SendPacket(match)

	// Wait for the match to start, with timeouts
	err = WaitForMatchReady(
		&conn,
		rlbotAddress,
		120*time.Second,
		20*time.Second,
	)
	if err != nil {
		conn.SendPacket(&flat.DisconnectSignalT{})
		return err
	}

	return nil
}

func (a *App) StartMatch(options StartMatchOptions) Result {
	var gameMode flat.GameMode
	switch options.GameMode {
	case "Soccar":
		gameMode = flat.GameModeSoccar
	case "Hoops":
		gameMode = flat.GameModeHoops
	case "Dropshot":
		gameMode = flat.GameModeDropshot
	case "Snowday":
		gameMode = flat.GameModeSnowday
	case "Rumble":
		gameMode = flat.GameModeRumble
	case "Heatseeker":
		gameMode = flat.GameModeHeatseeker
	case "Gridiron":
		gameMode = flat.GameModeGridiron
	case "Knockout":
		gameMode = flat.GameModeKnockout
	default:
		println("No mode chosen, defaulting to soccar")
		gameMode = flat.GameModeSoccar
	}

	var launcher flat.Launcher
	switch options.Launcher {
	case "steam":
		launcher = flat.LauncherSteam
	case "epic":
		launcher = flat.LauncherEpic
	case "legendary", "heroic":
		launcher = flat.LauncherCustom
		options.LauncherArg = options.Launcher
	case "custom":
		launcher = flat.LauncherCustom
	case "nolaunch":
		launcher = flat.LauncherNoLaunch
	default:
		println("No launcher chosen, defaulting to NoLaunch")
		launcher = flat.LauncherNoLaunch
	}

	playerConfigs :=
		make([]*flat.PlayerConfigurationT, len(options.BluePlayers)+len(options.OrangePlayers))

	for i, playerInfo := range options.BluePlayers {
		playerConfigs[i] = playerInfo.ToPlayer().ToPlayerConfig(0)
	}

	for i, playerInfo := range options.OrangePlayers {
		playerConfigs[i+len(options.BluePlayers)] = playerInfo.ToPlayer().ToPlayerConfig(1)
	}

	scriptConfigs :=
		make([]*flat.ScriptConfigurationT, len(options.Scripts))
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

	err := StartAndWaitForMatch(a.rlbotAddress, &match)
	if err != nil {
		return Result{false, err.Error()}
	}

	return Result{true, ""}
}

func (a *App) StopMatch(shutdownServer bool) Result {
	conn, err := rlbot.Connect(a.rlbotAddress)
	if err != nil {
		return Result{false, "Failed to connect to rlbot"}
	}

	conn.SendPacket(&flat.StopCommandT{
		ShutdownServer: shutdownServer,
	})
	conn.SendPacket(&flat.DisconnectSignalT{})

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
	if filename == "bot.toml" || filename == "script.toml" ||
		strings.HasSuffix(filename, ".bot.toml") ||
		strings.HasSuffix(filename, ".script.toml") {
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
