<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RHostBot, RHostServer } from "../../bindings/gui/index.js";
import { MAPS_STANDARD } from "../arena-names";
import closeIcon from "../assets/close.svg";
import heartIcon from "../assets/heart.svg";
import Plus from "../assets/plus.svg.svelte";
import LauncherSelector from "../components/LauncherSelector.svelte";
import { mapStore } from "../settings";
import Modal from "../components/Modal.svelte";
import RLLobby from "../components/RLLobby.svelte";
import { Browser } from "@wailsio/runtime";
import { draggable, droppable, type DragDropState } from "@thisux/sveltednd";
import { flip } from "svelte/animate";

let waiting = $state(false);

let bots: RHostBot[] = $state([]);
let botFamilies = $derived.by(() => {
  let families: {
    [name: string]: string[];
  } = {};
  for (const bot of bots) {
    const fam = bot.family !== "" ? bot.family : bot.name;
    if (!Object.hasOwn(families, fam)) {
      families[fam] = [];
    }
    families[fam].push(bot.name);
  }
  return families;
});

function refreshRHostBots() {
  App.GetRHostBots()
    .then((result) => {
      bots = result;
    })
    .catch((error) => {
      toast.error(`Couldn't resolve Rocket Host bots\n${error}`, {
        position: "top-center",
        duration: 5000,
      });
    });
}
refreshRHostBots();

let serverAddr: string = $state(
  localStorage.getItem("RHOST_SERVER_ADDR") || "",
);
$effect(() => {
  localStorage.setItem("RHOST_SERVER_ADDR", serverAddr);
});

let servers: RHostServer[] = $state([]);
$effect(() => {
  if (servers.length > 0) {
    serverAddr =
      serverAddr === "" ? `${servers[0].ip}:${servers[0].port}` : serverAddr;
  }
});

function refreshRHostServers() {
  App.GetRHostServers()
    .then((result) => {
      servers = result;
    })
    .catch((error) => {
      toast.error(`Couldn't resolve Rocket Host server addresses\n${error}`, {
        position: "top-center",
        duration: 5000,
      });
    });
}
refreshRHostServers();

let blueBots: string[] = $state([]);
let orangeBots: string[] = $state([]);
let launcherOptionsVisible = $state(false);

let activePage: "start" | "rllobby" = $state("start");
const RH_BOTLIST = "rh-botlist";
const RH_BLUE = "rh-blue";
const RH_ORANGE = "rh-orange";
const flipDurationMs = 200;

function onDragEnd(state: DragDropState) {
  const { targetContainer, sourceContainer } = state;
  if (!targetContainer) return;
  console.log("onDragEnd", targetContainer, sourceContainer);
}

async function onDrop(state: DragDropState<string>) {
  const {
    targetContainer,
    sourceContainer,
    dropPosition,
    draggedItem: draggedItemStr,
  } = state;
  if (!targetContainer) return;

  // biome-ignore format: preserve manual layout
  let sourceItems =
    sourceContainer.startsWith(RH_BOTLIST) ? Object.values(botFamilies).flat() :
    sourceContainer.startsWith(RH_BLUE) ? blueBots :
    sourceContainer.startsWith(RH_ORANGE) ? orangeBots :
    [];

  // biome-ignore format: preserve manual layout
  let targetItems =
    targetContainer.startsWith(RH_BLUE) ? blueBots :
    targetContainer.startsWith(RH_ORANGE) ? orangeBots :
    [];

  console.log(sourceItems, targetItems);

  let draggedItem: { bot: string; i: number | null } =
    JSON.parse(draggedItemStr);
  let dropIndex = +(targetContainer.split("_").at(-1) ?? -1);

  if (dropPosition === "after") dropIndex += 1;

  let item = draggedItem.bot;

  if (draggedItem.i !== null) {
    sourceItems.splice(draggedItem.i, 1);
    if (sourceItems === targetItems) {
      dropIndex = draggedItem.i < dropIndex ? dropIndex - 1 : dropIndex;
    }
  }

  targetItems.splice(dropIndex, 0, item);
}

function startRHostMatch() {
  let launcher = localStorage.getItem("MS_LAUNCHER");
  if (!launcher) {
    toast.error("Please select a launcher first", {
      position: "top-center",
      duration: 5000,
    });

    launcherOptionsVisible = true;
    return;
  }

  waiting = true;
  let id = toast.loading("Starting rocket host game...", {
    position: "top-center",
  });
  App.StartRHostMatch({
    server: serverAddr,
    map: $mapStore,
    blueBots,
    orangeBots,
    launcher,
    launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
  })
    .then((addr) => {
      waiting = false;
      toast.success(`Started game with address ${addr}`, {
        position: "top-center",
        duration: 10000,
        id,
      });
    })
    .catch((e) => {
      waiting = false;
      toast.error("Failed to start Rocket Host game\n" + e, {
        position: "top-center",
        duration: 8000,
        id,
      });
    });
}
</script>

<div class="page blurred">
  <div class="internalNav">
    <button
      class={activePage === "start" ? "active" : ""}
      onclick={() => { activePage = "start"; }}
    >Rocket Host</button>
    <button
      class={activePage === "rllobby" ? "active" : ""}
      onclick={() => { activePage = "rllobby"; }}
    >Lobbies</button>
  </div>
  <div class={"internalPage " + (activePage !== "start" ? "hidden" : "")}>
    <div class="availableBots">
      <h2>Available bots</h2>
      <div class="availableBotsList">
        {#each Object.keys(botFamilies) as family, _ }
          {#each botFamilies[family] as version, _ (family + version)}
            <div
              class="botEntry blurred"
              use:draggable={{
                container: RH_BOTLIST,
                dragData: JSON.stringify({ bot: version.split("(")[0].trim(), i: null })
              }}
              animate:flip={{ duration: flipDurationMs }}
            >
              <p>{version}</p>
            </div>
          {/each}
        {/each}
      </div>
    </div>

    <div class="teams">
      {#each [
        {team: "blue", bots: blueBots, containerName: RH_BLUE},
        {team: "orange", bots: orangeBots, containerName: RH_ORANGE},
      ] as {team, bots, containerName}}
      <div class={team}>
        <h2>{team.toUpperCase()}</h2>
        <div class="botList" use:droppable={{ container: `${containerName}_-1`, callbacks: { onDrop }, disabled: bots.length != 0 }}>
          {#each bots as bot, i}
            <div
              class="bot blurred"
              use:draggable={{ container: `${containerName}_${i}`, dragData: JSON.stringify({ bot, i }), callbacks: { onDragEnd } }}
              use:droppable={{ container: `${containerName}_${i}`, callbacks: { onDrop } }}
            >
              <p>{bot}</p>
              <button class="close" onclick={()=>{bots.splice(i, 1)}}>
                <img src={closeIcon} alt="X" />
              </button>
            </div>
          {/each}
          {#if bots.length == 0}
            <p>No bots selected for this team</p>
          {:else}
            <div
              style="flex:1;width:100%"
              use:droppable={{ container: `${containerName}_${bots.length}`, callbacks: { onDrop }}}
            ></div>
          {/if}
        </div>
      </div>
      {/each}
    </div>

    <div class="options">
      {#if servers.length > 0}
        <div>
          <label for="serverselect">Server</label>
            <select name="serverselect" id="serverselect" bind:value={serverAddr}>
              {#each servers as value, i}
                <option value={`${value.ip}:${value.port}`}>{value.location}</option>
              {/each}
            </select>
        </div>
        <div>
          <label for="mapselect">Map</label>
          <select name="mapselect" id="mapselect" bind:value={$mapStore}>
            {#each Object.entries(MAPS_STANDARD) as map, i}
              <option value={map[1]}>{map[0]}</option>
            {/each}
          </select>
        </div>
        <div>
          <button onclick={() => { launcherOptionsVisible = true }}>Launcher Options</button>
        </div>
        <button class="start" disabled={waiting} onclick={startRHostMatch}>
          Start
        </button>
      {:else}
        <div class="maintenance blurred">RocketHost is currently down for maintenance. Please try again later.</div>
      {/if}
    </div>

    <div class="donateBar" style="--icon-url: url({heartIcon});">
      <div>
        <h3>Please consider donating</h3>
        <p>RocketHost relies on donations; any amount will help expand server capacity</p>
      </div>
      <button
        onclick={() => Browser.OpenURL("https://www.patreon.com/WcW")}
      >Donate</button>
    </div>
  </div>
  <div class={"internalPage " + (activePage !== "rllobby" ? "hidden" : "")}>
    <RLLobby bind:servers bind:launcherOptionsVisible/>
  </div>
</div>

<Modal title="Select a launcher" bind:visible={launcherOptionsVisible}>
  <LauncherSelector />
</Modal>

<style>
  .page {
    display: flex;
    height: fit-content;
    background-color: var(--background);
    border-radius: 1rem;
    margin: 3rem;
    margin-bottom: 0;
    flex-direction: column;
    align-items: center;
    overflow: hidden;
    width: 100%;
  }
  .internalPage {
    display: flex;
    width: 100%;
    flex-direction: column;
    gap: 1.5rem;
  }
  .internalPage > * {
    padding: 0 2rem;
  }
  .internalNav {
    margin: 0;
    padding: 0;
    display: flex;
    gap: 2px;
    width: 100%;
    margin-bottom: 1rem;
  }
  .internalNav button {
    flex: 1;
    font-size: 1.2rem;
    border-bottom: 2px solid transparent;
    background: transparent;
  }
  .internalNav button.active {
    border-bottom: 2px solid var(--foreground);
  }
  h2 {
    margin-bottom: 0.5rem;
  }
  .options {
    display: flex;
    gap: 1rem;
    width: 100%
  }
  .options > div {
    display: flex;
    justify-content: end;
    flex-direction: column;
  }
  .options select {
    font-size: 1.0rem;
    padding: 0.25rem;
  }
  button.start {
    background-color: #15680e;
    font-size: 1.2rem;
    height: fit-content;
    margin-top: auto;
    margin-left: auto;
  }
  .availableBots {
    display: flex;
    flex-direction: column;
  }
  .availableBotsList {
    display: flex;
    flex-wrap: wrap;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.4rem;
  }
  .expandMe {
    flex-grow: 1;
  }
  .botEntry {
    display: flex;
    align-items: center;
    padding: 0.1rem;
    background-color: var(--background-alt);
    user-select: none;
    -webkit-user-select: none;
    border-radius: 0.3rem;
  }
  .botEntry p {
    margin: 0px;
    margin-right: 1rem;
    margin-left: 0.5rem;
    padding: .2rem;
    font-size: 1.1rem;
  }
  .botEntry select {
    margin: 0px;
    margin-right: .6rem;
  }
  button.addToTeam {
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0px;
    padding: 0.1rem;
    border-radius: 50%;
    margin: 0px 0.2rem;
    height: 2rem;
    width: 2rem;
  }
  button.addToTeam.blue {
    background-color: #026df9aa;
  }
  button.addToTeam.orange {
    background-color: #f95402cc;
  }
  .teams {
    display: flex;
    width: 100%;
    gap: 1rem;
    align-items: stretch;
  }
  .teams > div {
    display: flex;
    flex-direction: column;
    flex: 1;
  }
  .botList {
    display: flex;
    flex-direction: column;
    border-top: solid 3px;
    padding: 0.5rem;
    gap: 0.3rem;
    flex: 1;
  }
  .blue > .botList {
    border-color: #0054a6;
  }
  .orange > .botList {
    border-color: #f26522;
  }
  .bot {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--background-alt);
    width: 100%;
    padding: 0.3rem 0.5rem;
    border-radius: 0.3rem;
    font-size: 1.1rem;
  }
  .bot * {
    user-select: none;
  }
  button.close {
    padding: 0px;
  }
  button.close > img {
    filter: invert() brightness(var(--icon-brightness));;
    height: 28px;
    width: 28px;
  }
  .maintenance {
    width: 100%;
    text-align: center;
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    border: 1px solid var(--orange);
    color: var(--orange);
    font-weight: 600;
  }
  .donateBar {
    display: flex;
    width: 100%;
    padding: 1rem 2rem;
    border-radius: 0px 0px 1rem 1rem;
    justify-content: space-between;
    border-bottom: solid 0.2rem #e8a0ac;
  }
  .donateBar button {
    display: flex;
    align-items: center;
    background-color: #e8a0ac;
    color: #1c0004;
    font-weight: 600;
  }
  .donateBar button::after {
    content: var(--icon-url);
    height: 1.5rem;
    width: 1.5rem;
    margin-left: 0.5rem;
  }
  .hidden, .hidden * {
    display: none;
  }
</style>
