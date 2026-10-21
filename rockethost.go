package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	rlbot "github.com/RLBot/go-interface"
	"github.com/RLBot/go-interface/flat"
)

type RHostBot struct {
	Name        string   `json:"name"`
	Family      string   `json:"family"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Developer   string   `json:"developer"`
	Language    string   `json:"language"`
	FunFact     string   `json:"fun_fact"`
	SourceLink  string   `json:"source_link"`
	Tags        []string `json:"tags"`
	Aliases     []string `json:"aliases"`
}

func (a *App) GetRHostBots(server string) ([]RHostBot, error) {
	if server == "" {
		return nil, errors.New("server is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/status", server))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RocketHost status request failed: %s", resp.Status)
	}

	var data struct {
		Bots struct {
			List []struct {
				RHostBot
				Status string `json:"status"`
			} `json:"list"`
		} `json:"bots"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid RocketHost status response: %w", err)
	}
	if data.Bots.List == nil {
		return nil, errors.New("invalid RocketHost status response: bots.list is missing or null")
	}

	bots := []RHostBot{}
	for _, bot := range data.Bots.List {
		if bot.Name != "" && bot.Status == "ready" {
			bots = append(bots, bot.RHostBot)
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

type RLLobbyListing struct {
	Id                 string   `json:"id"`
	Name               string   `json:"name"`
	Map                string   `json:"map"`
	Description        string   `json:"description"`
	PlayerCount        int      `json:"playerCount"`
	Players            []string `json:"players"`
	HasPassword        bool     `json:"hasPassword"`
	IpAddress          string   `json:"ipAddress"`
	Port               int      `json:"port"`
	SecondsSinceUpdate int      `json:"secondsSinceUpdate"`
}

type RLLobbyResponse struct {
	Lobbies []RLLobbyListing `json:"lobbies"`
}

func (a *App) GetRHostServers() ([]RHostServer, error) {
	resp, err := http.Get("http://serverlist.jetfox.ovh/servers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (a *App) GetRLLobbies() ([]RLLobbyListing, error) {
	resp, err := http.Get("https://rllobby.martinn.no/rllobby/lobby")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		return nil, err
	}

	var data RLLobbyResponse
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	sort.Slice(data.Lobbies, func(i, j int) bool {
		return data.Lobbies[i].Name < data.Lobbies[j].Name
	})

	return data.Lobbies, nil
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
		if resp.StatusCode != 200 {
			respRHostChan <- Result{false, fmt.Sprintf("%d %s", resp.StatusCode, string(body))}
			return
		}
		respRHostChan <- Result{true, string(body)}
	}()

	// TODO: Save this in App struct
	conn, err := rlbot.Connect(a.rlbotAddress)
	if err != nil {
		return "", errors.New("Failed to connect to RLBotServer at " + a.rlbotAddress)
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
		PlayerConfigurations:  []*flat.PlayerConfigurationT{},
		ScriptConfigurations:  []*flat.ScriptConfigurationT{},
		GameMode:              flat.GameModeSoccar,
		Mutators:              &flat.MutatorSettingsT{},
		ExistingMatchBehavior: flat.ExistingMatchBehaviorRestart,
		GameMapUpk:            settings.Map,
		EnableStateSetting:    true,
		EnableRendering:       flat.DebugRenderingOnByDefault,
		Launcher:              launcher,
		LauncherArg:           settings.LauncherArg,
	})
	if err != nil {
		return "", errors.New("Couldn't send matchconfiguration packet")
	}

	err = conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "",
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
		_, ok := packet.Value.(*flat.FieldInfoT)
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

	conn.SendPacket(&flat.DisconnectSignalT{})

	return result.Message, nil
}

func (a *App) JoinLobby(settings RHostMatchSettings) error {
	if settings.Server == "" {
		return errors.New("server is required")
	}

	conn, err := rlbot.Connect(a.rlbotAddress)
	if err != nil {
		return errors.New("Failed to connect to RLBotServer at " + a.rlbotAddress)
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
		PlayerConfigurations:  []*flat.PlayerConfigurationT{},
		ScriptConfigurations:  []*flat.ScriptConfigurationT{},
		GameMode:              flat.GameModeSoccar,
		Mutators:              &flat.MutatorSettingsT{},
		ExistingMatchBehavior: flat.ExistingMatchBehaviorRestart,
		GameMapUpk:            settings.Map,
		EnableStateSetting:    true,
		EnableRendering:       flat.DebugRenderingOnByDefault,
		Launcher:              launcher,
		LauncherArg:           settings.LauncherArg,
	})
	if err != nil {
		return errors.New("Couldn't send matchconfiguration packet")
	}

	err = conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "",
		WantsBallPredictions: false,
		WantsComms:           false,
		CloseBetweenMatches:  false,
	})
	if err != nil {
		return errors.New("Couldn't send connectionsettings packet")
	}

	for {
		packet, err := conn.RecvPacket()
		if err != nil {
			return errors.New("Error reading packet from rlbotserver: " + err.Error())
		}
		_, ok := packet.Value.(*flat.FieldInfoT)
		if ok {
			break
		}
	}

	err = conn.SendPacket(&flat.DesiredGameStateT{
		ConsoleCommands: []*flat.ConsoleCommandT{
			{
				Command: fmt.Sprintf("open %s?Lan?Password=", settings.Server),
			},
		},
	})
	if err != nil {
		return errors.New("Couldn't send join message")
	}

	conn.SendPacket(&flat.DisconnectSignalT{})
	return nil
}
