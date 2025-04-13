package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/swz-git/go-interface/flat"
	"github.com/wailsapp/mimetype"
)

type PlayerJs struct {
	Sort   string          `json:"sort"`
	Player json.RawMessage `json:"player"`
}

func (playerJs PlayerJs) ToPlayer() Player {
	switch playerJs.Sort {
	case "rlbot":
		var correct BotInfo
		if err := json.Unmarshal([]byte(playerJs.Player), &correct); err != nil {
			log.Fatal("unable to unmarshal PlayerJs")
		}
		return correct
	case "psyonix":
		var correct PsyonixBotInfo
		if err := json.Unmarshal([]byte(playerJs.Player), &correct); err != nil {
			log.Fatal("unable to unmarshal PlayerJs")
		}
		return correct
	case "human":
		var correct HumanInfo
		if err := json.Unmarshal([]byte(playerJs.Player), &correct); err != nil {
			log.Fatal("unable to unmarshal PlayerJs")
		}
		return correct
	}
	log.Println("ERROR: invalid sort field in PlayerJs")
	return PsyonixBotInfo{}
}

type Player interface {
	ToPlayerConfig(team uint32) *flat.PlayerConfigurationT
}

type PsyonixBotInfo struct {
	// Beginner: 0, Rookie: 1, Pro: 2, AllStar: 3
	Skill byte `json:"skill"`
}

func (info PsyonixBotInfo) ToPlayerConfig(team uint32) *flat.PlayerConfigurationT {
	return &flat.PlayerConfigurationT{
		Variety: &flat.PlayerClassT{
			Type: flat.PlayerClassPsyonix,
			Value: &flat.PsyonixT{
				BotSkill: flat.PsyonixSkill(info.Skill),
			},
		},
		Name:       "",
		Team:       team,
		RootDir:    "",
		RunCommand: "",
		Loadout:    &flat.PlayerLoadoutT{},
		SpawnId:    0,
		Hivemind:   false,
	}
}

type HumanInfo struct{}

func (info HumanInfo) ToPlayerConfig(team uint32) *flat.PlayerConfigurationT {
	return &flat.PlayerConfigurationT{
		Variety: &flat.PlayerClassT{
			Type:  flat.PlayerClassHuman,
			Value: &flat.HumanT{},
		},
		Name:       "",
		Team:       team,
		RootDir:    "",
		RunCommand: "",
		Loadout:    &flat.PlayerLoadoutT{},
		SpawnId:    0,
		Hivemind:   false,
	}
}

type TeamPaintConfig struct {
	CarPaintId           uint32 `toml:"car_paint_id" json:"carPaintId"`
	DecalPaintId         uint32 `toml:"decal_paint_id" json:"decalPaintId"`
	WheelsPaintId        uint32 `toml:"wheels_paint_id" json:"wheelsPaintId"`
	BoostPaintId         uint32 `toml:"boost_paint_id" json:"boostPaintId"`
	AntennaPaintId       uint32 `toml:"antenna_paint_id" json:"antennaPaintId"`
	HatPaintId           uint32 `toml:"hat_paint_id" json:"hatPaintId"`
	TrailsPaintId        uint32 `toml:"trails_paint_id" json:"trailsPaintId"`
	GoalExplosionPaintId uint32 `toml:"goal_explosion_paint_id" json:"goalExplosionPaintId"`
}

type TeamLoadoutConfig struct {
	TeamColorId     uint32          `toml:"team_color_id" json:"teamColorId"`
	CustomColorId   uint32          `toml:"custom_color_id" json:"customColorId"`
	CarId           uint32          `toml:"car_id" json:"carId"`
	DecalId         uint32          `toml:"decal_id" json:"decalId"`
	WheelsId        uint32          `toml:"wheels_id" json:"wheelsId"`
	BoostId         uint32          `toml:"boost_id" json:"boostId"`
	AntennaId       uint32          `toml:"antenna_id" json:"antennaId"`
	HatId           uint32          `toml:"hat_id" json:"hatId"`
	PaintFinishId   uint32          `toml:"paint_finish_id" json:"paintFinishId"`
	CustomFinishId  uint32          `toml:"custom_finish_id" json:"customFinishId"`
	EngineAudioId   uint32          `toml:"engine_audio_id" json:"engineAudioId"`
	TrailsId        uint32          `toml:"trails_id" json:"trailsId"`
	GoalExplosionId uint32          `toml:"goal_explosion_id" json:"goalExplosionId"`
	Paint           TeamPaintConfig `toml:"paint" json:"paint"`
}

func (teamLoadout TeamLoadoutConfig) ToPlayerLoadout() *flat.PlayerLoadoutT {
	return &flat.PlayerLoadoutT{
		TeamColorId:     teamLoadout.TeamColorId,
		CustomColorId:   teamLoadout.CustomColorId,
		CarId:           teamLoadout.CarId,
		DecalId:         teamLoadout.DecalId,
		WheelsId:        teamLoadout.WheelsId,
		BoostId:         teamLoadout.BoostId,
		AntennaId:       teamLoadout.AntennaId,
		HatId:           teamLoadout.HatId,
		PaintFinishId:   teamLoadout.PaintFinishId,
		CustomFinishId:  teamLoadout.CustomFinishId,
		EngineAudioId:   teamLoadout.EngineAudioId,
		TrailsId:        teamLoadout.TrailsId,
		GoalExplosionId: teamLoadout.GoalExplosionId,
		LoadoutPaint: &flat.LoadoutPaintT{
			CarPaintId:           teamLoadout.Paint.CarPaintId,
			DecalPaintId:         teamLoadout.Paint.DecalPaintId,
			WheelsPaintId:        teamLoadout.Paint.WheelsPaintId,
			BoostPaintId:         teamLoadout.Paint.BoostPaintId,
			AntennaPaintId:       teamLoadout.Paint.AntennaPaintId,
			HatPaintId:           teamLoadout.Paint.HatPaintId,
			TrailsPaintId:        teamLoadout.Paint.TrailsPaintId,
			GoalExplosionPaintId: teamLoadout.Paint.GoalExplosionPaintId,
		},
		PrimaryColorLookup:   &flat.ColorT{},
		SecondaryColorLookup: &flat.ColorT{},
	}
}

type LoadoutConfig struct {
	Blue   TeamLoadoutConfig `toml:"blue_loadout" json:"blueLoadout"`
	Orange TeamLoadoutConfig `toml:"orange_loadout" json:"orangeLoadout"`
}

type BotInfo struct {
	Config   BotConfig      `json:"config"`
	Loadout  *LoadoutConfig `json:"loadout,omitempty"`
	TomlPath string         `json:"tomlPath"`
}

func (botInfo BotInfo) ToPlayerConfig(team uint32) *flat.PlayerConfigurationT {
	var runCommand string
	if runtime.GOOS == "windows" {
		runCommand = botInfo.Config.Settings.RunCommand
	} else if runtime.GOOS == "linux" {
		runCommand = botInfo.Config.Settings.RunCommandLinux
	}

	var loadout *flat.PlayerLoadoutT = nil
	if botInfo.Loadout != nil {
		var teamLoadout *TeamLoadoutConfig
		if team == 0 {
			teamLoadout = &botInfo.Loadout.Blue
		} else {
			teamLoadout = &botInfo.Loadout.Orange
		}

		loadout = teamLoadout.ToPlayerLoadout()
	}

	return &flat.PlayerConfigurationT{
		Variety: &flat.PlayerClassT{
			Type:  flat.PlayerClassCustomBot,
			Value: &flat.CustomBotT{},
		},
		Name:       botInfo.Config.Settings.Name,
		AgentId:    botInfo.Config.Settings.AgentId,
		Team:       team,
		RootDir:    botInfo.Config.Settings.RootDir,
		RunCommand: runCommand,
		Loadout:    loadout,
		SpawnId:    0, // let core do this
		Hivemind:   botInfo.Config.Settings.Hivemind,
	}
}

type BotConfig struct {
	Settings BotSettings `toml:"settings" json:"settings"`
	Details  BotDetails  `toml:"details" json:"details"`
}

type BotSettings struct {
	// In-game name of the bot
	Name string `toml:"name" json:"name"`
	// A unique string identifying this type of bot, typically on the form "<developer>/<botname>"
	AgentId string `toml:"agent_id" json:"agentId"`
	// Path to loadout.toml, describing the bots "loadout"
	LoadoutFile string `toml:"loadout_file" json:"loadoutFile"`
	// Optional working dir of the bot
	RootDir string `toml:"root_dir" json:"rootDir"`
	// Path to the logo of the bot, if ignored, RLBot will look for logo.png
	LogoFile string `toml:"logo_file" json:"logoFile"`
	// The command RLBot will call to start your bot on Windows
	RunCommand string `toml:"run_command" json:"runCommand"`
	// The command RLBot will call to start your bot on Linux
	// If not defined, RLBot may try to run your bot under wine
	RunCommandLinux string `toml:"run_command_linux" json:"runCommandLinux"`
	// If bot can handle multiple agents with one client
	Hivemind bool `toml:"hivemind" json:"hivemind"`
}

type BotDetails struct {
	// Short description of thebot
	Description string `toml:"description" json:"description"`
	// A fun fact about the bot
	FunFact string `toml:"fun_fact" json:"funFact"`
	// Link to the source code of the bot (if its available)
	SourceLink string `toml:"source_link" json:"sourceLink"` // TODO: Rename this field to repo?
	// Name(s) of the bot developer(s)
	Developer string `toml:"developer" json:"developer"`
	// Programming language the bot is written in.
	// (RLGym for example is also valid even though it is written in Python)
	Language string `toml:"language" json:"language"`
	// ALL POSSIBLE TAGS: 1v1, teamplay, goalie, hoops, dropshot, snow-day, spike-rush, heatseeker, memebot
	// NOTE: Only add the goalie tag if your bot only plays as a goalie; this directly contrasts with the teamplay tag!
	// NOTE: Only add a tag for a special game mode if you bot properly supports it
	Tags []string `toml:"tags" json:"tags"`
}

func (a *App) GetBots(paths []string) []BotInfo {
	potentialConfigs := []string{}

	for _, path := range paths {
		new, err := recursiveTomlSearch(path, "bot")
		if err != nil {
			println("WARN: failed to search path: " + path)
			continue
		}
		potentialConfigs = append(potentialConfigs, new...)
	}

	infos := []BotInfo{}

	for _, potentialConfigPath := range potentialConfigs {
		data, err := os.ReadFile(potentialConfigPath)
		if err != nil {
			println("WARN: skipping config, couldn't read config at " + potentialConfigPath)
			continue
		}
		var conf BotConfig
		toml.Decode(string(data), &conf)

		// make location path relative to parent of bot.toml
		conf.Settings.RootDir = filepath.Join(filepath.Dir(potentialConfigPath), conf.Settings.RootDir)

		var logo_file string
		if conf.Settings.LogoFile == "" {
			logo_file = filepath.Join(conf.Settings.RootDir, "logo.png")
		} else {
			logo_file = filepath.Join(conf.Settings.RootDir, conf.Settings.LogoFile)
		}

		// Read logo file and convert it to data url so the frontend can use it
		logo_data, err := os.ReadFile(logo_file)
		if err != nil {
			// only warn if the logo file was explicitly set
			if conf.Settings.LogoFile != "" {
				println("WARN: failed to read logo file at " + conf.Settings.LogoFile)
			}
		} else {
			mtype := mimetype.Detect(logo_data)
			b64data := base64.StdEncoding.EncodeToString(logo_data)
			conf.Settings.LogoFile = "data:" + mtype.String() + ";base64," + b64data
		}

		var loadout *LoadoutConfig = nil
		if conf.Settings.LoadoutFile != "" {
			loadoutPath := filepath.Join(filepath.Dir(potentialConfigPath), conf.Settings.LoadoutFile)
			loadoutData, err := os.ReadFile(loadoutPath)
			if err != nil {
				println("WARN: failed to read loadout file at " + conf.Settings.LoadoutFile)
			} else {
				toml.Decode(string(loadoutData), &loadout)
			}
		}

		infos = append(infos, BotInfo{
			Config:   conf,
			Loadout:  loadout,
			TomlPath: potentialConfigPath,
		})
	}

	// sort infos by bot name
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Config.Settings.Name < infos[j].Config.Settings.Name
	})

	return infos
}
