package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	rlbot "github.com/swz-git/go-interface"
	"github.com/swz-git/go-interface/flat"
)

func (a *App) SaveLoadoutToFile(basePath string, loadoutFile string, loadout LoadoutConfig) error {
	baseDir := filepath.Dir(basePath)
	fullPath := filepath.Join(baseDir, loadoutFile)

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileContents, err := toml.Marshal(loadout)
	if err != nil {
		return err
	}

	_, err = file.Write(fileContents)
	return err
}

type LoadoutPreviewOptions struct {
	Map         string            `json:"map"`
	Loadout     TeamLoadoutConfig `json:"loadout"`
	Team        uint32            `json:"team"`
	Launcher    string            `json:"launcher"`
	LauncherArg string            `json:"launcherArg"`
}

func (options LoadoutPreviewOptions) GetPreviewMatch(existingMatchBehavior flat.ExistingMatchBehavior) (*flat.MatchConfigurationT, error) {
	loadout := options.Loadout.ToPlayerLoadout()

	playerConfigs := []*flat.PlayerConfigurationT{
		{
			Variety: &flat.PlayerClassT{
				Type:  flat.PlayerClassCustomBot,
				Value: &flat.CustomBotT{},
			},
			Name:       "Showcase",
			AgentId:    "gui/loadout-preview",
			Team:       options.Team,
			RootDir:    "",
			RunCommand: "",
			Loadout:    loadout,
			SpawnId:    0,
			Hivemind:   false,
		},
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
		return nil, errors.New("no launcher specified")
	}

	return &flat.MatchConfigurationT{
		AutoStartBots:        false,
		GameMapUpk:           options.Map,
		PlayerConfigurations: playerConfigs,
		ScriptConfigurations: []*flat.ScriptConfigurationT{},
		GameMode:             flat.GameModeSoccer,
		Mutators: &flat.MutatorSettingsT{
			MatchLength: flat.MatchLengthMutatorUnlimited,
		},
		Freeplay:              false,
		EnableRendering:       false,
		EnableStateSetting:    true,
		InstantStart:          true,
		SkipReplays:           true,
		AutoSaveReplay:        false,
		Launcher:              launcher,
		LauncherArg:           options.LauncherArg,
		ExistingMatchBehavior: existingMatchBehavior,
	}, nil
}

func (a *App) LaunchPreviewLoadout(options LoadoutPreviewOptions, existingMatchBehavior flat.ExistingMatchBehavior) error {
	conn, err := rlbot.Connect(a.rlbot_address)
	if err != nil {
		return err
	}

	match, err := options.GetPreviewMatch(existingMatchBehavior)
	if err != nil {
		return err
	}

	conn.SendPacket(match)
	conn.SendPacket(nil)

	return nil
}

func (a *App) SetLoadout(options LoadoutPreviewOptions) error {
	conn, err := rlbot.Connect(a.rlbot_address)
	if err != nil {
		return err
	}

	conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "",
		WantsBallPredictions: false,
		WantsComms:           false,
		CloseBetweenMatches:  false,
	})

	conn.SendPacket(&flat.InitCompleteT{})

	// wait for GamePacket
	var gamePacket *flat.GamePacketT
	for gamePacket == nil {
		packet, err := conn.RecvPacket()
		if err != nil {
			return err
		}

		switch packet := packet.(type) {
		case *flat.GamePacketT:
			gamePacket = packet
		}
	}

	// if the match is over, launch a new match
	isMatchOver := gamePacket.MatchInfo.MatchPhase == flat.MatchPhaseEnded || gamePacket.MatchInfo.MatchPhase == flat.MatchPhaseInactive
	// ensure the player is a custom bot
	isPlayerBot := len(gamePacket.Players) == 1 && gamePacket.Players[0].IsBot
	// ensure unlimited time
	isUnlimitedTime := gamePacket.MatchInfo.IsUnlimitedTime

	if isMatchOver || !isPlayerBot || !isUnlimitedTime {
		match, err := options.GetPreviewMatch(flat.ExistingMatchBehaviorRestart)
		if err != nil {
			return err
		}

		conn.SendPacket(match)
		conn.SendPacket(nil)
		return nil;
	}

	// ensure the player is on the correct team
	if gamePacket.Players[0].Team != options.Team {
		match, err := options.GetPreviewMatch(flat.ExistingMatchBehaviorContinueAndSpawn)
		if err != nil {
			return err
		}

		conn.SendPacket(match)
		conn.SendPacket(nil)
		return nil;
	}

	conn.SendPacket(&flat.SetLoadoutT{
		Index:   0,
		Loadout: options.Loadout.ToPlayerLoadout(),
	})

	conn.SendPacket(nil)

	return nil
}
