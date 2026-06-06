<script lang="ts">
import { Browser, Events } from "@wailsio/runtime";
import { App } from "../../bindings/gui/index.js";
import arenaDiagramUrl from "../assets/arena_diagram.png";
import { PerformanceMonitor } from "../../bindings/github.com/RLBot/go-interface/flat/models.js";
import NiceSelect from "../components/NiceSelect.svelte";

const PIXEL_HEIGHT = 580;
const CANVAS_WIDTH = 410;

// Loaded arena background image
let arenaBgImage: HTMLImageElement | null = null;
{
  const img = new Image();
  img.onload = () => {
    arenaBgImage = img;
    if (canvasCtx) render();
  };
  img.src = arenaDiagramUrl;
}
const HISTORY_SECONDS = 5;
const HISTORY_INCREMENT_SECONDS = 0.1;

interface SandboxPhysics {
  location: { x: number; y: number; z: number };
  rotation: { pitch: number; yaw: number; roll: number };
  velocity: { x: number; y: number; z: number };
  angular_velocity: { x: number; y: number; z: number };
}

interface SandboxGamePacket {
  ball: { physics: SandboxPhysics };
  cars: {
    index: number;
    physics: SandboxPhysics;
    team: number;
    boost: number;
    is_bot: boolean;
    name: string;
  }[];
  seconds_elapsed: number;
}

interface CarState {
  index: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  rotation: number;
  team: number;
  boost: number;
  name: string;
  isBot: boolean;
}

let { visible = true }: { visible?: boolean } = $props();

let connectionState = $state<"disconnected" | "connecting" | "connected">(
  "disconnected",
);
let errorMessage = $state("");

// Canvas rendering state
let ball = $state({ x: 100, y: 100, vx: 0, vy: 0 });
let cars = $state<CarState[]>([]);
let secondsElapsed = $state(0);
let previousSecondsElapsed = $state(0);

// Packet history for rewind
let packetHistory: SandboxGamePacket[] = $state([]);
let hasPacketHistory = $state(false);

// Saved state for save/load
let savedState = $state<SandboxGamePacket | null>(null);
let hasSavedState = $state(false);

// Controls
let watching = $state(false);
let frozen = $state(false);
let gravity = $state("normal");
let gamespeed = $state(1);
let command = $state("");

// Debug rendering state
interface RenderAgent {
  index: number;
  name: string;
  is_bot: boolean;
}

interface MatchConfig {
  enable_rendering: number;
  performance_monitor: number;
  agents: RenderAgent[];
}

const performanceMonitorOptions: { [n: string]: number } = {
  "Show when suboptimal": 0,
  "Always show": 1,
  "Never show": 2,
};

let matchConfig = $state<MatchConfig | null>(null);
let renderStatuses = $state<Map<string, boolean>>(new Map());
let renderingDisabled = $state(false);
let perfMonDisplayMode = $state(
  PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal,
);

// Drag state
let dragging = $state(false);
let dragTarget: "ball" | "car" | null = null;
let dragCarIndex = -1;
let dragStartX = 0;
let dragStartY = 0;

// Canvas
let canvas: HTMLCanvasElement | undefined = $state(undefined);
let canvasCtx: CanvasRenderingContext2D | null = null;

// Latest packet for rendering
let latestPacket: SandboxGamePacket | null = $state(null);
let lastUpdateTime = 0;
const UPDATE_INTERVAL_MS = 50; // ~20 fps throttle for game state updates

// Coordinate conversion
function toCanvasVec(
  x: number,
  y: number,
  _z: number,
): { x: number; y: number } {
  return {
    x: x / -20 + CANVAS_WIDTH / 2,
    y: y / -20 + PIXEL_HEIGHT / 2,
  };
}

function toPacketVec(
  cx: number,
  cy: number,
  z: number,
): { x: number; y: number; z: number } {
  return {
    x: (cx - CANVAS_WIDTH / 2) * -20,
    y: (cy - PIXEL_HEIGHT / 2) * -20,
    z: z,
  };
}

function drawField(ctx: CanvasRenderingContext2D) {
  if (arenaBgImage) {
    ctx.drawImage(arenaBgImage, 0, 0, CANVAS_WIDTH, PIXEL_HEIGHT);
  } else {
    // Fallback until image loads
    ctx.fillStyle = "#1a1a2e";
    ctx.fillRect(0, 0, CANVAS_WIDTH, PIXEL_HEIGHT);
  }
}

function drawBall(ctx: CanvasRenderingContext2D) {
  ctx.beginPath();
  ctx.arc(ball.x, ball.y, 8, 0, Math.PI * 2);
  ctx.fillStyle = "gray";
  ctx.fill();
  ctx.strokeStyle = "black";
  ctx.lineWidth = 2;
  ctx.stroke();
}

function drawCar(ctx: CanvasRenderingContext2D, car: CarState) {
  const w = 16;
  const h = 10;

  ctx.save();
  ctx.translate(car.x, car.y);
  ctx.rotate(car.rotation);
  ctx.fillStyle = car.team === 0 ? "blue" : "orange";
  ctx.fillRect(-w / 2, -h / 2, w, h);
  ctx.strokeStyle = "black";
  ctx.lineWidth = 2;
  ctx.strokeRect(-w / 2, -h / 2, w, h);
  ctx.restore();
}

function drawVelocityArrow(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  vx: number,
  vy: number,
  color: string,
  maxRadius: number,
) {
  const length = Math.sqrt(vx * vx + vy * vy);
  if (length < 1) return;

  // Scale velocity to screen coordinates
  const screenVx = vx / -20;
  const screenVy = vy / -20;

  const arrowLen = Math.sqrt(screenVx * screenVx + screenVy * screenVy);
  if (arrowLen < 2) return;

  const scale = Math.min(arrowLen, maxRadius / 50) / arrowLen;
  const ex = screenVx * scale;
  const ey = screenVy * scale;

  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(x, y);
  ctx.lineTo(x + ex, y + ey);
  ctx.stroke();

  // Arrowhead
  const headLen = 8;
  const angle = Math.atan2(ey, ex);
  ctx.fillStyle = color;
  ctx.beginPath();
  ctx.moveTo(x + ex, y + ey);
  ctx.lineTo(
    x + ex - headLen * Math.cos(angle - 0.4),
    y + ey - headLen * Math.sin(angle - 0.4),
  );
  ctx.lineTo(
    x + ex - headLen * Math.cos(angle + 0.4),
    y + ey - headLen * Math.sin(angle + 0.4),
  );
  ctx.closePath();
  ctx.fill();
}

function render() {
  if (!canvas || !canvasCtx) return;

  const ctx = canvasCtx;
  ctx.clearRect(0, 0, CANVAS_WIDTH, PIXEL_HEIGHT);

  drawField(ctx);

  // Draw velocity arrows first (behind objects)
  if (!dragging) {
    drawVelocityArrow(ctx, ball.x, ball.y, ball.vx, ball.vy, "gray", 6000);
    for (const car of cars) {
      drawVelocityArrow(
        ctx,
        car.x,
        car.y,
        car.vx,
        car.vy,
        car.team === 0 ? "blue" : "orange",
        2300,
      );
    }
  }

  // Draw cars
  for (const car of cars) {
    drawCar(ctx, car);
  }

  // Draw ball on top
  drawBall(ctx);
}

function setupCanvas() {
  if (!canvas) return;
  canvasCtx = canvas.getContext("2d");
  if (!canvasCtx) return;
  render();
}

// Event handlers for drag and drop
function getCanvasPos(e: MouseEvent | Touch): { x: number; y: number } | null {
  if (!canvas) return null;
  const rect = canvas.getBoundingClientRect();
  return {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top,
  };
}

function hitTestBall(pos: { x: number; y: number }): boolean {
  const dx = pos.x - ball.x;
  const dy = pos.y - ball.y;
  return dx * dx + dy * dy < 12 * 12;
}

function hitTestCar(pos: { x: number; y: number }): number {
  for (let i = cars.length - 1; i >= 0; i--) {
    const c = cars[i];
    const dx = pos.x - c.x;
    const dy = pos.y - c.y;
    if (dx * dx + dy * dy < 14 * 14) {
      return i;
    }
  }
  return -1;
}

function handleMouseDown(e: MouseEvent) {
  const pos = getCanvasPos(e);
  if (!pos) return;

  if (hitTestBall(pos)) {
    dragging = true;
    dragTarget = "ball";
    dragStartX = ball.x - pos.x;
    dragStartY = ball.y - pos.y;
  } else {
    const ci = hitTestCar(pos);
    if (ci >= 0) {
      dragging = true;
      dragTarget = "car";
      dragCarIndex = ci;
      dragStartX = cars[ci].x - pos.x;
      dragStartY = cars[ci].y - pos.y;
    }
  }
}

function handleMouseMove(e: MouseEvent) {
  if (!dragging) return;
  const pos = getCanvasPos(e);
  if (!pos) return;

  if (dragTarget === "ball") {
    ball.x = pos.x + dragStartX;
    ball.y = pos.y + dragStartY;
  } else if (dragTarget === "car" && dragCarIndex >= 0 && cars[dragCarIndex]) {
    cars[dragCarIndex].x = pos.x + dragStartX;
    cars[dragCarIndex].y = pos.y + dragStartY;
  }
}

function handleMouseUp(_e: MouseEvent) {
  if (!dragging) return;
  dragging = false;

  if (dragTarget === "ball") {
    const loc = toPacketVec(ball.x, ball.y, 10);
    App.SandboxSetState({
      ball: { location: { x: loc.x, y: loc.y, z: loc.z } },
    });
  } else if (dragTarget === "car" && dragCarIndex >= 0 && cars[dragCarIndex]) {
    const c = cars[dragCarIndex];
    const loc = toPacketVec(c.x, c.y, 1);
    App.SandboxSetState({
      cars: [
        {
          index: c.index,
          location: { x: loc.x, y: loc.y, z: loc.z },
        },
      ],
    });
  }

  dragTarget = null;
  dragCarIndex = -1;
}

// Lifecycle: connect/disconnect sandbox when page visibility changes
$effect(() => {
  if (!visible) return;

  // Reset state for a fresh connection
  connectionState = "connecting";
  errorMessage = "";
  watching = false;
  frozen = false;
  gravity = "normal";
  gamespeed = 1;
  command = "";
  matchConfig = null;
  renderStatuses = new Map();
  renderingDisabled = false;
  perfMonDisplayMode = PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal;
  ball = { x: 100, y: 100, vx: 0, vy: 0 };
  cars = [];
  secondsElapsed = 0;
  previousSecondsElapsed = 0;
  packetHistory = [];
  hasPacketHistory = false;
  savedState = null;
  hasSavedState = false;
  latestPacket = null;
  lastUpdateTime = 0;

  // Listen for sandbox events
  const unsubConnecting = Events.On("sandbox:connecting", () => {
    connectionState = "connecting";
  });

  const unsubConnected = Events.On("sandbox:connected", () => {
    connectionState = "connected";
  });

  const unsubDisconnected = Events.On("sandbox:disconnected", () => {
    connectionState = "disconnected";
    watching = false;
    matchConfig = null;
    renderStatuses = new Map();
    renderingDisabled = false;
    perfMonDisplayMode =
      PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal;
  });

  const unsubError = Events.On(
    "sandbox:error",
    (event: { data: { message: string } }) => {
      errorMessage = event.data.message;
      connectionState = "disconnected";
      matchConfig = null;
      renderStatuses = new Map();
      renderingDisabled = false;
      perfMonDisplayMode =
        PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal;
    },
  );

  const unsubMatchConfig = Events.On(
    "sandbox:match-config",
    (event: { data: MatchConfig }) => {
      const mc = event.data;
      matchConfig = mc;

      perfMonDisplayMode = mc.performance_monitor;

      // Determine default checked state based on DebugRendering mode
      const defaultChecked = mc.enable_rendering === 1; // DebugRenderingOnByDefault
      renderingDisabled = mc.enable_rendering === 2; // DebugRenderingAlwaysOff

      const statuses = new Map<string, boolean>();
      for (const agent of mc.agents) {
        const key = `${agent.is_bot ? "bot" : "script"}-${agent.index}`;
        statuses.set(key, defaultChecked);
      }
      renderStatuses = statuses;
    },
  );

  const unsubGamePacket = Events.On(
    "sandbox:game-packet",
    (event: { data: SandboxGamePacket }) => {
      const result = event.data;
      latestPacket = result;

      // Detect match reset: seconds_elapsed dropped significantly (new match)
      if (result.seconds_elapsed < secondsElapsed - 1) {
        previousSecondsElapsed = 0;
        packetHistory = [];
        hasPacketHistory = false;
        savedState = null;
        hasSavedState = false;
        lastUpdateTime = 0;

        // Reset debug rendering to defaults for the current match config
        if (matchConfig) {
          perfMonDisplayMode = matchConfig.performance_monitor;

          const defaultChecked = matchConfig.enable_rendering === 1;
          renderingDisabled = matchConfig.enable_rendering === 2;
          const statuses = new Map<string, boolean>();
          for (const agent of matchConfig.agents) {
            const key = `${agent.is_bot ? "bot" : "script"}-${agent.index}`;
            statuses.set(key, defaultChecked);
          }
          renderStatuses = statuses;
        }
      }

      secondsElapsed = result.seconds_elapsed;

      // Throttle visual updates to ~20 fps to prevent lag
      const now = Date.now();
      if (now - lastUpdateTime >= UPDATE_INTERVAL_MS) {
        lastUpdateTime = now;

        if (watching && !dragging && secondsElapsed > previousSecondsElapsed) {
          previousSecondsElapsed = secondsElapsed;

          // Update ball
          const ballPhys = result.ball.physics;
          const ballLoc = toCanvasVec(
            ballPhys.location.x,
            ballPhys.location.y,
            ballPhys.location.z,
          );
          ball.x = ballLoc.x;
          ball.y = ballLoc.y;
          ball.vx = ballPhys.velocity.x;
          ball.vy = ballPhys.velocity.y;

          // Update cars
          cars = result.cars.map((c) => {
            const carLoc = toCanvasVec(
              c.physics.location.x,
              c.physics.location.y,
              c.physics.location.z,
            );
            return {
              index: c.index,
              x: carLoc.x,
              y: carLoc.y,
              vx: c.physics.velocity.x,
              vy: c.physics.velocity.y,
              rotation: c.physics.rotation.yaw,
              team: c.team,
              boost: c.boost,
              name: c.name,
              isBot: c.is_bot,
            };
          });
        }
      }

      // Update packet history (not throttled, needs to stay accurate)
      if (packetHistory.length > 0) {
        const tail = packetHistory[packetHistory.length - 1];
        if (
          result.seconds_elapsed - tail.seconds_elapsed >
          HISTORY_INCREMENT_SECONDS
        ) {
          packetHistory = [...packetHistory, result];
          if (
            packetHistory.length >
            HISTORY_SECONDS / HISTORY_INCREMENT_SECONDS
          ) {
            packetHistory = packetHistory.slice(1);
          }
        }
      } else {
        packetHistory = [result];
      }
      hasPacketHistory = true;
    },
  );

  // Open sandbox connection
  App.OpenSandbox().catch((err: unknown) => {
    errorMessage = `Failed to open sandbox: ${err}`;
    connectionState = "disconnected";
  });

  // Cleanup when navigating away from the sandbox page
  return () => {
    unsubConnecting();
    unsubConnected();
    unsubDisconnected();
    unsubError();
    unsubMatchConfig();
    unsubGamePacket();
    App.CloseSandbox();

    // Reset state for when user comes back
    connectionState = "disconnected";
    errorMessage = "";
    matchConfig = null;
    renderStatuses = new Map();
    renderingDisabled = false;
    perfMonDisplayMode =
      PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal;
    packetHistory = [];
    hasPacketHistory = false;
    savedState = null;
    hasSavedState = false;
    previousSecondsElapsed = 0;
    lastUpdateTime = 0;
  };
});

// Bind canvas after mount
$effect(() => {
  if (canvas) {
    setupCanvas();
  }
});

// Reactive render: redraw whenever displayed state changes
$effect(() => {
  if (!canvas || !canvasCtx) return;

  // Track reactive state to trigger re-render
  const _ball = ball;
  const _cars = cars;
  const _dragging = dragging;

  render();
});

// Control handlers
function startWatching() {
  previousSecondsElapsed = secondsElapsed;
}

function rewind() {
  if (packetHistory.length > 0) {
    const firstPacket = packetHistory[0];
    const lastPacket = packetHistory[packetHistory.length - 1];

    // Doctor the time
    firstPacket.seconds_elapsed = lastPacket.seconds_elapsed;

    // Truncate history
    packetHistory = [firstPacket];

    const carsSetting = firstPacket.cars.map((c) => ({
      index: c.index,
      location: {
        x: c.physics.location.x,
        y: c.physics.location.y,
        z: c.physics.location.z,
      },
      velocity: {
        x: c.physics.velocity.x,
        y: c.physics.velocity.y,
        z: c.physics.velocity.z,
      },
      rotation: {
        pitch: c.physics.rotation.pitch,
        yaw: c.physics.rotation.yaw,
        roll: c.physics.rotation.roll,
      },
      boost: c.boost,
    }));

    const ballPhys = firstPacket.ball.physics;
    App.SandboxSetState({
      cars: carsSetting,
      ball: {
        location: {
          x: ballPhys.location.x,
          y: ballPhys.location.y,
          z: ballPhys.location.z,
        },
        velocity: {
          x: ballPhys.velocity.x,
          y: ballPhys.velocity.y,
          z: ballPhys.velocity.z,
        },
        rotation: {
          pitch: ballPhys.rotation.pitch,
          yaw: ballPhys.rotation.yaw,
          roll: ballPhys.rotation.roll,
        },
      },
    });
  }
  hasPacketHistory = false;
}

function saveState() {
  if (latestPacket) {
    savedState = latestPacket;
    hasSavedState = true;
  }
}

function loadSavedState() {
  if (!savedState) return;

  const carsSetting = savedState.cars.map((c) => ({
    index: c.index,
    location: {
      x: c.physics.location.x,
      y: c.physics.location.y,
      z: c.physics.location.z,
    },
    velocity: {
      x: c.physics.velocity.x,
      y: c.physics.velocity.y,
      z: c.physics.velocity.z,
    },
    rotation: {
      pitch: c.physics.rotation.pitch,
      yaw: c.physics.rotation.yaw,
      roll: c.physics.rotation.roll,
    },
    boost: c.boost,
  }));

  const ballPhys = savedState.ball.physics;
  App.SandboxSetState({
    cars: carsSetting,
    ball: {
      location: {
        x: ballPhys.location.x,
        y: ballPhys.location.y,
        z: ballPhys.location.z,
      },
      velocity: {
        x: ballPhys.velocity.x,
        y: ballPhys.velocity.y,
        z: ballPhys.velocity.z,
      },
      rotation: {
        pitch: ballPhys.rotation.pitch,
        yaw: ballPhys.rotation.yaw,
        roll: ballPhys.rotation.roll,
      },
    },
  });
}

function handleWatchingChange() {
  if (watching) {
    startWatching();
  }
}

function handleFrozenChange() {
  App.SandboxSetState({
    console_commands: ["pause"],
  });
}

function handleGravityChange() {
  App.SandboxSetState({
    game_info: { world_gravity_z: gravity === "zero" ? -0.000001 : -650 },
  });
}

function setGamespeed() {
  App.SandboxSetState({
    game_info: { game_speed: gamespeed },
  });
}

function executeCommand() {
  if (command.trim()) {
    App.SandboxSetState({
      console_commands: [command],
    });
    command = "";
  }
}
</script>

<div class="page">
  {#if connectionState === "connecting"}
    <div class="loading">
      <div class="spinner"></div>
      <h2>Waiting for match start...</h2>
      {#if errorMessage}
        <p class="error">{errorMessage}</p>
      {/if}
    </div>
  {:else if connectionState === "disconnected"}
    <div class="loading">
      <div class="disconnected-icon">!</div>
      <h2>Disconnected</h2>
      <p>
        {errorMessage || "Not connected to RLBotServer. Make sure a match is running."}
      </p>
      <button
        onclick={() => {
          errorMessage = "";
          connectionState = "connecting";
          App.OpenSandbox();
        }}
      >
        Retry Connection
      </button>
    </div>
  {:else}
    <div class="sandbox-layout">
      <div class="canvas-container">
        <canvas
          bind:this={canvas}
          width={CANVAS_WIDTH}
          height={PIXEL_HEIGHT}
          class="arena-canvas"
          onmousedown={handleMouseDown}
          onmousemove={handleMouseMove}
          onmouseup={handleMouseUp}
          onmouseleave={handleMouseUp}
        ></canvas>
      </div>

      <div class="controls">
        <div class="control-section">
          <h3>Commands</h3>

          <div class="checkbox-row">
            <label>
              <input type="checkbox" bind:checked={watching} onchange={handleWatchingChange} />
              Watch Game
            </label>
            <label>
              <input type="checkbox" bind:checked={frozen} onchange={handleFrozenChange} />
              Freeze Game
            </label>
          </div>

          <div class="radio-row">
            <label>
              <input
                type="radio"
                bind:group={gravity}
                value="normal"
                onchange={handleGravityChange}
              />
              Normal gravity
            </label>
            <label>
              <input
                type="radio"
                bind:group={gravity}
                value="zero"
                onchange={handleGravityChange}
              />
              Zero gravity
            </label>
          </div>

          <div class="state-btn-row">
            <button onclick={rewind} disabled={!hasPacketHistory}>
              Rewind 5s
            </button>
            <button onclick={saveState}>
              Save State
            </button>
            <button onclick={loadSavedState} disabled={!hasSavedState}>
              Load State
            </button>
          </div>
        </div>

        <div class="control-section">
          <div class="input-row">
            <label for="gamespeed">Game Speed</label>
            <input
              id="gamespeed"
              type="number"
              bind:value={gamespeed}
              step="0.1"
            />
            <button onclick={setGamespeed}>Set</button>
          </div>

          <div class="input-row">
            <label for="command">Console Command</label>
            <input
              id="command"
              type="text"
              bind:value={command}
              onkeydown={(e) => {
                if (e.key === "Enter") executeCommand();
              }}
              placeholder="e.g. QueSaveReplay"
            />
            <button onclick={executeCommand}>Execute</button>
          </div>
          <p class="cmd-hint">See <a href="#" onclick={() => { Browser.OpenURL('https://wiki.rlbot.org/v5/framework/console-commands/'); }}>available console commands</a></p>
        </div>

        <div class="control-section">
          <h3>Debug Rendering</h3>
          {#if matchConfig}
            {#if renderingDisabled}
              <p class="disabled-hint">Rendering is globally disabled</p>
            {/if}
            {#each matchConfig.agents as agent}
              {@const key = `${agent.is_bot ? "bot" : "script"}-${agent.index}`}
              <div class="checkbox-row">
                <label class:disabled-label={renderingDisabled}>
                  <input
                    type="checkbox"
                    checked={renderStatuses.get(key) ?? false}
                    disabled={renderingDisabled}
                    onchange={(e) => {
                      if (renderingDisabled) return;
                      const newStatuses = new Map(renderStatuses);
                      newStatuses.set(key, e.currentTarget.checked);
                      renderStatuses = newStatuses;
                      App.SandboxSetRendering([{
                        index: agent.index,
                        is_bot: agent.is_bot,
                        status: e.currentTarget.checked,
                      }]);
                    }}
                  />
                  {agent.name}
                </label>
              </div>
            {/each}
          {:else}
            <p class="disabled-hint">Waiting for match config...</p>
          {/if}
        </div>

        <div class="control-section">
          <h3>Performance monitor display mode</h3>
          <NiceSelect bind:value={perfMonDisplayMode} options={performanceMonitorOptions} placeholder="Performance Monitor"
            on_change={() => {
              App.SandboxSetPerfMonDisplayMode(perfMonDisplayMode);
            }}
            />
        </div>

        <p class="hint">
          You can drag and drop to move objects around in the game!
        </p>
      </div>
    </div>
  {/if}
</div>

<style>
  .page {
    display: flex;
    height: 100%;
    width: 100%;
    justify-content: center;
    align-items: flex-start;
    padding: 1.5rem;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    background: var(--background);
    border-radius: 1rem;
    padding: 3rem;
    margin-top: 4rem;
  }

  .loading h2 {
    margin: 0;
    color: var(--foreground);
  }

  .loading p {
    color: var(--foreground);
    opacity: 0.8;
    max-width: 400px;
    text-align: center;
  }

  .error {
    color: #ff6b6b;
  }

  .disconnected-icon {
    font-size: 3rem;
    font-weight: bold;
    color: #ff6b6b;
    width: 5rem;
    height: 5rem;
    border-radius: 50%;
    border: 3px solid #ff6b6b;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .spinner {
    width: 3rem;
    height: 3rem;
    border: 3px solid var(--foreground);
    border-top-color: transparent;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .sandbox-layout {
    container-type: inline-size;
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 1.5rem;
    background: var(--background);
    border-radius: 1rem;
    padding: 1.5rem;
    max-width: 100%;
    overflow: hidden;
  }

  @container (max-width: 884px) {
    .canvas-container {
      width: 580px;
      height: 410px;
      overflow: hidden;
    }
    .arena-canvas {
      transform: translate(580px, 0) rotate(90deg);
      transform-origin: 0 0;
    }
  }

  .canvas-container {
    flex-shrink: 0;
  }

  .arena-canvas {
    border-radius: 4px;
    cursor: grab;
    display: block;
  }

  .arena-canvas:active {
    cursor: grabbing;
  }

  .controls {
    flex: 1 1 250px;
    min-width: 450px;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .control-section {
    background: var(--background-alt);
    border-radius: 0.5rem;
    padding: 1rem;
  }

  .control-section h3 {
    margin: 0 0 0.75rem 0;
    font-size: 1.1rem;
    color: var(--foreground);
  }

  .checkbox-row {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 0.75rem;
  }

  .checkbox-row label,
  .radio-row label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    cursor: pointer;
    user-select: none;
    color: var(--foreground);
  }

  .radio-row {
    display: flex;
    gap: 1rem;
    margin-bottom: 0.75rem;
  }

  .state-btn-row {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.75rem;
  }

  .state-btn-row button {
    flex: 1;
  }

  .input-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .input-row label {
    min-width: 6rem;
    font-size: 0.9rem;
    color: var(--foreground);
    white-space: nowrap;
  }

  .input-row input[type="number"],
  .input-row input[type="text"] {
    flex: 1;
    padding: 0.3rem 0.5rem;
    border: 1px solid #555;
    border-radius: 0.3rem;
    background: var(--background);
    color: var(--foreground);
    font-size: 0.9rem;
  }

  .input-row button {
    flex-shrink: 0;
  }

  button {
    padding: 0.4rem 0.8rem;
    border: 1px solid #555;
    border-radius: 0.3rem;
    background: var(--background);
    color: var(--foreground);
    cursor: pointer;
    font-size: 0.9rem;
  }

  button:hover {
    background: var(--background-alt);
  }

  button:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .cmd-hint {
    font-size: 0.8rem;
    color: var(--foreground);
    opacity: 0.6;
    margin: 0.2rem 0 0 0;
  }
  .cmd-hint a {
    color: var(--link-color, #58a6ff);
    text-decoration: underline;
    cursor: pointer;
  }
  .cmd-hint a:hover {
    opacity: 0.8;
  }

  .disabled-label {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .disabled-hint {
    font-size: 0.85rem;
    color: var(--foreground);
    opacity: 0.5;
    margin: 0 0 0.5rem 0;
    font-style: italic;
  }

  .hint {
    font-size: 0.85rem;
    color: var(--foreground);
    opacity: 0.7;
    margin: 0;
    text-align: center;
  }
</style>
