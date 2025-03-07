package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	rlbot "github.com/swz-git/go-interface"
	"github.com/swz-git/go-interface/flat"
)

type RHostBot struct {
	Name   string `json:"name"`
	Family string `json:"family"`
}

func (a *App) GetRHostBots() ([]RHostBot, error) {
	resp, err := http.Get("https://rocketleaguemaps.us/api/botList2.json")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	jsonBots, ok := data.([]interface{})
	if !ok {
		return nil, errors.New("invalid json shape: root isn't array")
	}

	bots := []RHostBot{}

	for _, jsonBotOrFamily := range jsonBots {
		if botString, ok := jsonBotOrFamily.(string); ok {
			bots = append(bots, RHostBot{
				Name:   botString,
				Family: "",
			})
		} else if botFamily, ok := jsonBotOrFamily.(map[string]interface{}); ok {
			familyName, ok := botFamily["name"].(string)
			if !ok {
				return nil, errors.New("invalid json shape: family object doesn't have \"name\"")
			}

			familyBots, ok := botFamily["versions"].([]interface{})
			if !ok {
				return nil, errors.New("invalid json shape: family object doesn't have \"versions\"")
			}
			for _, familyBot := range familyBots {
				familyBotStr, ok := familyBot.(string)
				if !ok {
					return nil, errors.New("invalid json shape: \"versions\" isn't an array of strings")
				}
				bots = append(bots, RHostBot{
					Name:   familyBotStr,
					Family: familyName,
				})
			}
		}
	}

	return bots, nil
}

type RHostServer struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Location string `json:"location"`
	Domain   string `json:"domain"`
}

func (a *App) GetRHostServers() ([]RHostServer, error) {
	resp, err := http.Get("http://serverlist.jetfox.ovh/servers")
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, err
	}
	var data []RHostServer
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	return data, nil
}

// /?mapName=ARC_Darc_P&gameMode=TAGame.GameInfo_Soccar_TA&
// mutators=BotsNone,PlayerCount8&
// networkOptions=?NumPublicConnections=10?NumOpenPublicConnections=10?Lan?Listen
// &rlbot=yes&blueBots=FridgeV5&orangeBots=

type RHostMatchSettings struct {
	Server      string   `json:"server"`
	Map         string   `json:"map"`
	BlueBots    []string `json:"blueBots"`
	OrangeBots  []string `json:"orangeBots"`
	Launcher    string   `json:"launcher"`
	LauncherArg string   `json:"launcherArg"`
}

func (a *App) StartRHostMatch(settings RHostMatchSettings) (string, error) {
	respRHostChan := make(chan Result)

	// Request rockethost server
	go func() {
		blueBotsStr := strings.Join(settings.BlueBots, ",")
		orangeBotsStr := strings.Join(settings.OrangeBots, ",")
		resp, err :=
			http.Get(fmt.Sprintf(
				"http://%s/?mapName=%s&gameMode=TAGame.GameInfo_Soccar_TA&mutators=BotsNone,PlayerCount8&networkOptions=?NumPublicConnections=10?NumOpenPublicConnections=10?Lan?Listen&rlbot=yes&blueBots=%s&orangeBots=%s",
				settings.Server, settings.Map, blueBotsStr, orangeBotsStr,
			))
		if err != nil {
			respRHostChan <- Result{false, err.Error()}
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			respRHostChan <- Result{false, err.Error()}
			return
		}
		respRHostChan <- Result{resp.StatusCode == 200, string(body)}
	}()

	// TODO: Save this in App struct
	conn, err := rlbot.Connect(a.rlbot_address)
	if err != nil {
		return "", errors.New("Failed to connect to RLBotServer at " + a.rlbot_address)
	}

	var launcher flat.Launcher
	switch settings.Launcher {
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

	err = conn.SendPacket(&flat.MatchConfigurationT{
		PlayerConfigurations: []*flat.PlayerConfigurationT{},
		ScriptConfigurations: []*flat.ScriptConfigurationT{
			{
				Name:       "GUIv5",
				RootDir:    "",
				RunCommand: "",
				SpawnId:    0,
				AgentId:    "rlbot/gui",
			},
		},
		GameMode:              flat.GameModeSoccer,
		Mutators:              &flat.MutatorSettingsT{},
		ExistingMatchBehavior: flat.ExistingMatchBehaviorRestart,
		GameMapUpk:            settings.Map,
		EnableStateSetting:    true,
		EnableRendering:       true,
		Launcher:              launcher,
		LauncherArg:           settings.LauncherArg,
	})
	if err != nil {
		return "", errors.New("Couldn't send matchconfiguration packet")
	}

	err = conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "rlbot/gui",
		WantsBallPredictions: false,
		WantsComms:           false,
		CloseBetweenMatches:  false,
	})
	if err != nil {
		return "", errors.New("Couldn't send connectionsettings packet")
	}

	println("Waiting for FieldInfo...")
	for {
		packet, err := conn.RecvPacket()
		if err != nil {
			return "", errors.New("Error reading packet from rlbotserver: " + err.Error())
		}
		_, ok := packet.(*flat.FieldInfoT)
		if ok {
			break
		}
	}

	println("Waiting for RocketHost server...")
	var result Result
outer:
	for {
		select {
		case result = <-respRHostChan:
			break outer
		default:
			err = conn.SendPacket(&flat.RenderGroupT{
				RenderMessages: []*flat.RenderMessageT{
					{
						Variety: &flat.RenderTypeT{
							Type: flat.RenderTypeString2D,
							Value: &flat.String2DT{
								Text:  "Loading RocketHost game...",
								X:     0.5,
								Y:     0.5,
								Scale: 1.5,
								Foreground: &flat.ColorT{
									R: 255,
									G: 255,
									B: 255,
									A: 255,
								},
								Background: &flat.ColorT{
									R: 0,
									G: 0,
									B: 0,
									A: 255,
								},
								HAlign: flat.TextHAlignCenter,
								VAlign: flat.TextVAlignCenter,
							},
						},
					},
				},
				Id: 0,
			})
			if err != nil {
				println("WARN: couldn't render loading message")
			}
			time.Sleep(1 * time.Second)
		}
	}

	if !result.Success {
		time.Sleep(time.Second * 1)
		err = conn.SendPacket(&flat.StopCommandT{
			ShutdownServer: false,
		})
		if err != nil {
			return "", errors.New("Couldn't send stopcommand packet")
		}

		return "", errors.New(result.Message)
	}

	err = conn.SendPacket(&flat.DesiredGameStateT{
		ConsoleCommands: []*flat.ConsoleCommandT{
			{
				Command: fmt.Sprintf("start %s/?Lan?Password=", result.Message),
			},
		},
	})
	if err != nil {
		return "", errors.New("Couldn't send join message")
	}

	conn.SendPacket(nil) // Tell core that we want to disconnect

	return result.Message, nil
}
