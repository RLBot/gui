package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/RLBot/go-interface/flat"
	"github.com/wailsapp/mimetype"
)

func (botInfo BotInfo) ToScriptConfig() *flat.ScriptConfigurationT {
	var runCommand string
	if runtime.GOOS == "windows" {
		runCommand = botInfo.Config.Settings.RunCommand
	} else if runtime.GOOS == "linux" {
		runCommand = botInfo.Config.Settings.RunCommandLinux
	}

	return &flat.ScriptConfigurationT{
		Name:       botInfo.Config.Settings.Name,
		AgentId:    botInfo.Config.Settings.AgentId,
		RootDir:    botInfo.Config.Settings.RootDir,
		RunCommand: runCommand,
		ScriptId:   0, // let core do this
	}
}

func (a *App) GetScripts(paths []string) []BotInfo {
	potentialConfigs := []string{}

	for _, path := range paths {
		new, err := recursiveTomlSearch(path, "script")
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

		infos = append(infos, BotInfo{
			Config:   conf,
			TomlPath: potentialConfigPath,
		})
	}

	// sort infos by bot name
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Config.Settings.Name < infos[j].Config.Settings.Name
	})

	return infos
}
