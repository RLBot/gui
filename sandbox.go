package main

import (
	"sync"
	"time"

	rlbot "github.com/RLBot/go-interface"
	"github.com/RLBot/go-interface/flat"
)

// SandboxState holds the current sandbox connection and manages concurrent access.
type SandboxState struct {
	conn   *rlbot.RLBotConnection
	mu     sync.Mutex
	active bool
}

// SandboxBallPacket is sent to the frontend for each game tick.
type SandboxGamePacket struct {
	Ball           SandboxBall  `json:"ball"`
	Cars           []SandboxCar `json:"cars"`
	SecondsElapsed float32      `json:"seconds_elapsed"`
}

type SandboxBall struct {
	Physics SandboxPhysics `json:"physics"`
}

type SandboxCar struct {
	Index   int32          `json:"index"`
	Physics SandboxPhysics `json:"physics"`
	Team    uint32         `json:"team"`
	Boost   float32        `json:"boost"`
	IsBot   bool           `json:"is_bot"`
	Name    string         `json:"name"`
}

type SandboxPhysics struct {
	Location        SandboxVec3 `json:"location"`
	Rotation        SandboxRot3 `json:"rotation"`
	Velocity        SandboxVec3 `json:"velocity"`
	AngularVelocity SandboxVec3 `json:"angular_velocity"`
}

type SandboxVec3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type SandboxRot3 struct {
	Pitch float32 `json:"pitch"`
	Yaw   float32 `json:"yaw"`
	Roll  float32 `json:"roll"`
}

// SandboxStateSetting is sent from the frontend to set game state.
type SandboxStateSetting struct {
	Ball *SandboxBallSetting `json:"ball,omitempty"`
	Cars []SandboxCarSetting `json:"cars,omitempty"`
	Game *SandboxGameSetting `json:"game_info,omitempty"`
	Cmds []string            `json:"console_commands,omitempty"`
}

type SandboxBallSetting struct {
	Location *SandboxVec3 `json:"location,omitempty"`
	Velocity *SandboxVec3 `json:"velocity,omitempty"`
	Rotation *SandboxRot3 `json:"rotation,omitempty"`
}

type SandboxCarSetting struct {
	Index    int32        `json:"index"`
	Location *SandboxVec3 `json:"location,omitempty"`
	Velocity *SandboxVec3 `json:"velocity,omitempty"`
	Rotation *SandboxRot3 `json:"rotation,omitempty"`
	Boost    *float32     `json:"boost,omitempty"`
}

type SandboxGameSetting struct {
	GravityZ  *float32 `json:"world_gravity_z,omitempty"`
	GameSpeed *float32 `json:"game_speed,omitempty"`
}

// SandboxMatchConfig is sent to the frontend with match configuration info for debug rendering.
type SandboxMatchConfig struct {
	EnableRendering    int                  `json:"enable_rendering"`
	PerformanceMonitor int                  `json:"performance_monitor"`
	Agents             []SandboxRenderAgent `json:"agents"`
}

type SandboxRenderAgent struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	IsBot bool   `json:"is_bot"`
}

type SandboxRenderingStatus struct {
	Index  uint32 `json:"index"`
	IsBot  bool   `json:"is_bot"`
	Status bool   `json:"status"`
}

// OpenSandbox connects to RLBotServer and starts reading game packets.
// The frontend should call this when entering the sandbox page.
func (a *App) OpenSandbox() error {
	a.sandboxMu.Lock()
	defer a.sandboxMu.Unlock()

	if a.sandbox != nil && a.sandbox.active {
		return nil // already connected
	}

	a.app.Event.Emit("sandbox:connecting", nil)

	conn, err := rlbot.Connect(a.rlbotAddress)
	if err != nil {
		a.app.Event.Emit("sandbox:error", map[string]any{
			"message": "Failed to connect to RLBotServer: " + err.Error(),
		})
		return err
	}

	// Send initial connection settings
	err = conn.SendPacket(&flat.ConnectionSettingsT{
		AgentId:              "",
		WantsBallPredictions: false,
		WantsComms:           false,
		CloseBetweenMatches:  false,
	})
	if err != nil {
		conn.SendPacket(&flat.DisconnectSignalT{})
		a.app.Event.Emit("sandbox:error", map[string]any{
			"message": "Failed to send connection settings: " + err.Error(),
		})
		return err
	}

	state := &SandboxState{
		conn:   &conn,
		active: true,
	}
	a.sandbox = state

	// Start the reader goroutine
	go a.sandboxReader(state, conn)

	return nil
}

// CloseSandbox disconnects from RLBotServer.
// The frontend should call this when leaving the sandbox page.
func (a *App) CloseSandbox() {
	a.sandboxMu.Lock()
	state := a.sandbox
	a.sandbox = nil
	a.sandboxMu.Unlock()

	if state == nil || !state.active {
		return
	}

	state.mu.Lock()
	state.active = false
	if state.conn != nil {
		state.conn.SendPacket(&flat.DisconnectSignalT{})
	}
	state.mu.Unlock()
}

// SandboxSetState sets the game state (ball, cars, game info, console commands).
func (a *App) SandboxSetState(setting SandboxStateSetting) error {
	a.sandboxMu.Lock()
	state := a.sandbox
	a.sandboxMu.Unlock()

	if state == nil || !state.active {
		return nil // silently ignore if not connected
	}

	desired := &flat.DesiredGameStateT{}

	if setting.Ball != nil {
		ballState := &flat.DesiredBallStateT{}
		if setting.Ball.Location != nil {
			ballState.Physics = &flat.DesiredPhysicsT{
				Location: &flat.Vector3PartialT{
					X: &flat.FloatT{Val: setting.Ball.Location.X},
					Y: &flat.FloatT{Val: setting.Ball.Location.Y},
					Z: &flat.FloatT{Val: setting.Ball.Location.Z},
				},
			}
		}
		if setting.Ball.Velocity != nil {
			if ballState.Physics == nil {
				ballState.Physics = &flat.DesiredPhysicsT{}
			}
			ballState.Physics.Velocity = &flat.Vector3PartialT{
				X: &flat.FloatT{Val: setting.Ball.Velocity.X},
				Y: &flat.FloatT{Val: setting.Ball.Velocity.Y},
				Z: &flat.FloatT{Val: setting.Ball.Velocity.Z},
			}
		}
		desired.BallStates = []*flat.DesiredBallStateT{ballState}
	}

	if len(setting.Cars) > 0 {
		carStates := make([]*flat.DesiredCarStateT, 0)
		for _, cs := range setting.Cars {
			carState := &flat.DesiredCarStateT{}
			if cs.Location != nil {
				carState.Physics = &flat.DesiredPhysicsT{
					Location: &flat.Vector3PartialT{
						X: &flat.FloatT{Val: cs.Location.X},
						Y: &flat.FloatT{Val: cs.Location.Y},
						Z: &flat.FloatT{Val: cs.Location.Z},
					},
				}
			}
			if cs.Velocity != nil {
				if carState.Physics == nil {
					carState.Physics = &flat.DesiredPhysicsT{}
				}
				carState.Physics.Velocity = &flat.Vector3PartialT{
					X: &flat.FloatT{Val: cs.Velocity.X},
					Y: &flat.FloatT{Val: cs.Velocity.Y},
					Z: &flat.FloatT{Val: cs.Velocity.Z},
				}
			}
			if cs.Rotation != nil {
				if carState.Physics == nil {
					carState.Physics = &flat.DesiredPhysicsT{}
				}
				carState.Physics.Rotation = &flat.RotatorPartialT{
					Pitch: &flat.FloatT{Val: cs.Rotation.Pitch},
					Yaw:   &flat.FloatT{Val: cs.Rotation.Yaw},
					Roll:  &flat.FloatT{Val: cs.Rotation.Roll},
				}
			}
			if cs.Boost != nil {
				carState.BoostAmount = &flat.FloatT{Val: *cs.Boost}
			}
			// Ensure slice is large enough to hold this car index
			needed := int(cs.Index + 1)
			if len(carStates) < needed {
				padded := make([]*flat.DesiredCarStateT, needed)
				copy(padded, carStates)
				carStates = padded
			}
			carStates[cs.Index] = carState
		}
		desired.CarStates = carStates
	}

	if setting.Game != nil {
		desired.MatchInfo = &flat.DesiredMatchInfoT{}
		if setting.Game.GravityZ != nil {
			desired.MatchInfo.WorldGravityZ = &flat.FloatT{Val: *setting.Game.GravityZ}
		}
		if setting.Game.GameSpeed != nil {
			desired.MatchInfo.GameSpeed = &flat.FloatT{Val: *setting.Game.GameSpeed}
		}
	}

	if len(setting.Cmds) > 0 {
		cmds := make([]*flat.ConsoleCommandT, len(setting.Cmds))
		for i, cmd := range setting.Cmds {
			cmds[i] = &flat.ConsoleCommandT{Command: cmd}
		}
		desired.ConsoleCommands = cmds
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.conn != nil {
		return state.conn.SendPacket(desired)
	}
	return nil
}

// sandboxReader runs in a goroutine and reads packets from the RLBot connection.
func (a *App) sandboxReader(state *SandboxState, conn rlbot.RLBotConnection) {
	defer func() {
		a.sandboxMu.Lock()
		if a.sandbox == state {
			a.sandbox = nil
		}
		a.sandboxMu.Unlock()

		a.app.Event.Emit("sandbox:disconnected", nil)
	}()

	// Wait for MatchConfigurationT and FieldInfoT
	hasFieldInfo := false
	waitStart := time.Now()
	timeout := 30 * time.Second

	for !hasFieldInfo {
		if time.Since(waitStart) > timeout {
			a.app.Event.Emit("sandbox:error", map[string]any{
				"message": "Timed out waiting for FieldInfo",
			})
			return
		}

		packet, err := conn.RecvPacket()
		if err != nil {
			a.app.Event.Emit("sandbox:error", map[string]any{
				"message": "Connection error: " + err.Error(),
			})
			return
		}

		switch p := packet.Value.(type) {
		case *flat.FieldInfoT:
			// Send InitCompleteT to signal we're ready for GamePackets
			conn.SendPacket(&flat.InitCompleteT{})
			hasFieldInfo = true
		case *flat.MatchConfigurationT:
			// Extract match config info for debug rendering UI
			matchConfig := simplifyMatchConfig(p)
			a.app.Event.Emit("sandbox:match-config", matchConfig)
		case *flat.DisconnectSignalT:
			a.app.Event.Emit("sandbox:error", map[string]any{
				"message": "Received disconnect signal while connecting",
			})
			return
		}
	}

	// Signal that we're connected and ready
	a.app.Event.Emit("sandbox:connected", nil)

	// Now read GamePackets in a loop
	for {
		state.mu.Lock()
		isActive := state.active
		state.mu.Unlock()

		if !isActive {
			return
		}

		packet, err := conn.RecvPacket()
		if err != nil {
			// Connection closed
			return
		}

		switch p := packet.Value.(type) {
		case *flat.GamePacketT:
			sandboxPacket := simplifyGamePacket(p)
			a.app.Event.Emit("sandbox:game-packet", sandboxPacket)
		case *flat.DisconnectSignalT:
			return
		}
	}
}

// SandboxSetRendering sends per-agent rendering status updates to RLBot.
func (a *App) SandboxSetRendering(settings []SandboxRenderingStatus) error {
	a.sandboxMu.Lock()
	state := a.sandbox
	a.sandboxMu.Unlock()

	if state == nil || !state.active {
		return nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.conn == nil {
		return nil
	}

	for _, s := range settings {
		err := state.conn.SendPacket(&flat.RenderingStatusT{
			Index:  s.Index,
			IsBot:  s.IsBot,
			Status: s.Status,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) SandboxSetPerfMonDisplayMode(mode int) error {
	a.sandboxMu.Lock()
	state := a.sandbox
	a.sandboxMu.Unlock()

	if state == nil || !state.active {
		return nil
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.conn == nil {
		return nil
	}

	var displayMode flat.PerformanceMonitor
	switch mode {
	case 0:
		displayMode = flat.PerformanceMonitorShowWhenSuboptimal
	case 1:
		displayMode = flat.PerformanceMonitorAlwaysShow
	case 2:
		displayMode = flat.PerformanceMonitorNeverShow
	default:
		return nil // invalid mode, ignore
	}

	return state.conn.SendPacket(&flat.UpdatePerformanceMonitorT{
		Show: displayMode,
	})
}

// simplifyMatchConfig extracts rendering-relevant info from MatchConfigurationT.
func simplifyMatchConfig(mc *flat.MatchConfigurationT) SandboxMatchConfig {
	cfg := SandboxMatchConfig{
		EnableRendering:    int(mc.EnableRendering),
		PerformanceMonitor: int(mc.PerformanceMonitor),
	}

	// Collect bots (players that are not humans)
	// The loop index corresponds to the player's position in the GamePacket's Players array.
	for i, player := range mc.PlayerConfigurations {
		if player.Variety != nil && player.Variety.Type != flat.PlayerClassHuman {
			name := ""
			switch v := player.Variety.Value.(type) {
			case *flat.CustomBotT:
				name = v.Name
			case *flat.PsyonixBotT:
				name = v.Name
			}
			cfg.Agents = append(cfg.Agents, SandboxRenderAgent{
				Index: i,
				Name:  name,
				IsBot: true,
			})
		}
	}

	// Collect scripts
	for i, script := range mc.ScriptConfigurations {
		cfg.Agents = append(cfg.Agents, SandboxRenderAgent{
			Index: i,
			Name:  script.Name,
			IsBot: false,
		})
	}

	return cfg
}

// simplifyGamePacket converts a flat.GamePacketT to a lightweight SandboxGamePacket for the frontend.
func simplifyGamePacket(gp *flat.GamePacketT) SandboxGamePacket {
	pkt := SandboxGamePacket{}

	if gp.MatchInfo != nil {
		pkt.SecondsElapsed = gp.MatchInfo.SecondsElapsed
	}

	if len(gp.Balls) > 0 && gp.Balls[0] != nil && gp.Balls[0].Physics != nil {
		phys := gp.Balls[0].Physics
		pkt.Ball.Physics = SandboxPhysics{
			Location: SandboxVec3{
				X: phys.Location.X,
				Y: phys.Location.Y,
				Z: phys.Location.Z,
			},
			Rotation: SandboxRot3{
				Pitch: phys.Rotation.Pitch,
				Yaw:   phys.Rotation.Yaw,
				Roll:  phys.Rotation.Roll,
			},
			Velocity: SandboxVec3{
				X: phys.Velocity.X,
				Y: phys.Velocity.Y,
				Z: phys.Velocity.Z,
			},
			AngularVelocity: SandboxVec3{
				X: phys.AngularVelocity.X,
				Y: phys.AngularVelocity.Y,
				Z: phys.AngularVelocity.Z,
			},
		}
	}

	pkt.Cars = make([]SandboxCar, len(gp.Players))
	for i, player := range gp.Players {
		if player == nil {
			continue
		}
		car := SandboxCar{
			Index: int32(i),
			Team:  player.Team,
			Boost: player.Boost,
			IsBot: player.IsBot,
			Name:  player.Name,
		}
		if player.Physics != nil {
			car.Physics = SandboxPhysics{
				Location: SandboxVec3{
					X: player.Physics.Location.X,
					Y: player.Physics.Location.Y,
					Z: player.Physics.Location.Z,
				},
				Rotation: SandboxRot3{
					Pitch: player.Physics.Rotation.Pitch,
					Yaw:   player.Physics.Rotation.Yaw,
					Roll:  player.Physics.Rotation.Roll,
				},
				Velocity: SandboxVec3{
					X: player.Physics.Velocity.X,
					Y: player.Physics.Velocity.Y,
					Z: player.Physics.Velocity.Z,
				},
				AngularVelocity: SandboxVec3{
					X: player.Physics.AngularVelocity.X,
					Y: player.Physics.AngularVelocity.Y,
					Z: player.Physics.AngularVelocity.Z,
				},
			}
		}
		pkt.Cars[i] = car
	}

	return pkt
}
