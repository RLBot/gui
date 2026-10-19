<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RHostBot, RHostServer } from "../../bindings/gui/index.js";
import { MAPS_STANDARD } from "../arena-names";
import closeIcon from "../assets/close.svg";
import heartIcon from "../assets/heart.svg";
import infoIcon from "../assets/info_icon.svg";
import starIcon from "../assets/star.svg";
import filledStarIcon from "../assets/starFilled.svg";
import LauncherSelector from "../components/LauncherSelector.svelte";
import { mapStore } from "../settings";
import Modal from "../components/Modal.svelte";
import RLLobby from "../components/RLLobby.svelte";
import { Browser } from "@wailsio/runtime";
import { draggable, droppable, type DragDropState } from "@thisux/sveltednd";
import { flip } from "svelte/animate";

let waiting = $state(false);

let bots: RHostBot[] = $state([]);
let selectedBot: RHostBot | null = $state(null);
let showBotInfo = $state(false);
let selectedTag: string | null = $state(null);
let favoritesOnly = $state(false);
let favorites: string[] = $state(loadFavorites());

function loadFavorites(): string[] {
  try {
    const saved = JSON.parse(localStorage.getItem("RHOST_FAVORITES") || "[]");
    return Array.isArray(saved)
      ? saved.filter((name) => typeof name === "string")
      : [];
  } catch {
    return [];
  }
}

$effect(() => {
  localStorage.setItem("RHOST_FAVORITES", JSON.stringify(favorites));
});

function toggleFavorite(name: string) {
  favorites = favorites.includes(name)
    ? favorites.filter((favorite) => favorite !== name)
    : [...favorites, name];
}
const filterTags = [
  "1v1",
  "2v2",
  "3v3",
  "teamplay",
  "goalie",
  "hoops",
  "dropshot",
  "snow-day",
  "rumble",
  "spike-rush",
  "heatseeker",
  "memebot",
];
let availableTags = $derived(
  filterTags.filter((tag) => bots.some((bot) => bot.tags?.includes(tag))),
);
let filteredBots = $derived(
  bots.filter(
    (bot) =>
      (selectedTag === null || bot.tags?.includes(selectedTag)) &&
      (!favoritesOnly || favorites.includes(bot.name)),
  ),
);

function botDisplayName(name: string) {
  return bots.find((bot) => bot.name === name)?.display_name || name;
}

function sourceURL(source: string): string | null {
  try {
    const url = new URL(source);
    return url.protocol === "https:" || url.protocol === "http:"
      ? url.href
      : null;
  } catch {
    return null;
  }
}

let serverAddr: string = $state(
  localStorage.getItem("RHOST_SERVER_ADDR") || "",
);
$effect(() => {
  localStorage.setItem("RHOST_SERVER_ADDR", serverAddr);
});

let servers: RHostServer[] = $state([]);
$effect(() => {
  if (
    servers.length > 0 &&
    !servers.some((server) => `${server.ip}:${server.port}` === serverAddr)
  ) {
    serverAddr = `${servers[0].ip}:${servers[0].port}`;
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
let botsLoading = $state(false);
let botsLoaded = $state(false);

$effect(() => {
  const server = serverAddr;
  let cancelled = false;
  bots = [];
  selectedTag = null;
  selectedBot = null;
  showBotInfo = false;
  blueBots = [];
  orangeBots = [];
  botsLoaded = false;
  botsLoading = server !== "";

  if (server !== "") {
    App.GetRHostBots(server)
      .then((result) => {
        if (cancelled) return;
        bots = result;
        botsLoaded = true;
      })
      .catch((error) => {
        if (cancelled) return;
        toast.error(`Could not load RocketHost bots\n${error}`, {
          position: "top-center",
          duration: 5000,
        });
      })
      .finally(() => {
        if (!cancelled) botsLoading = false;
      });
  }

  return () => {
    cancelled = true;
  };
});

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
    sourceContainer.startsWith(RH_BOTLIST) ? bots.map((bot) => bot.name) :
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
      {#if botsLoading}
        <p>Loading bots...</p>
      {:else if serverAddr === ""}
        <p>Select a server to load bots.</p>
      {:else if !botsLoaded}
        <p>Could not load bots. Select a server to try again.</p>
      {:else if bots.length === 0}
        <p>No bots are ready on this server.</p>
      {/if}
      {#if botsLoaded && bots.length > 0}
        <div class="tag-buttons" role="group" aria-label="Filter bots">
          <div class="tag-group">
            <button
              class:selected={selectedTag === null && !favoritesOnly}
              aria-pressed={selectedTag === null && !favoritesOnly}
              onclick={() => { selectedTag = null; favoritesOnly = false; }}
            >All</button>
          </div>
          <div class="tag-group">
            {#each availableTags as tag (tag)}
              <button
                class:selected={selectedTag === tag}
                aria-pressed={selectedTag === tag}
                onclick={() => { selectedTag = selectedTag === tag ? null : tag; }}
              >{tag}</button>
            {/each}
          </div>
          <div class="tag-group">
            <button
              class:selected={favoritesOnly}
              aria-pressed={favoritesOnly}
              onclick={() => { favoritesOnly = !favoritesOnly; }}
            >Favorites</button>
          </div>
        </div>
      {/if}
      {#if botsLoaded && bots.length > 0 && filteredBots.length === 0}
        <p>No bots match these filters.</p>
      {/if}
      <div class="availableBotsList">
        {#each filteredBots as bot (bot.name)}
            <div
              class="botEntry blurred"
              use:draggable={{
                container: RH_BOTLIST,
                dragData: JSON.stringify({ bot: bot.name, i: null }),
                interactive: [".info-button"]
              }}
              animate:flip={{ duration: flipDurationMs }}
            >
              <p>{bot.display_name || bot.name}</p>
              <button
                class="info-button"
                aria-label={`Show details for ${bot.display_name || bot.name}`}
                onclick={() => { selectedBot = bot; showBotInfo = true; }}
              >
                <img src={infoIcon} alt="" />
              </button>
            </div>
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
              <p>{botDisplayName(bot)}</p>
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
        <button class="start" disabled={waiting || !botsLoaded} onclick={startRHostMatch}>
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

{#snippet botInfoTitle()}
  <span class="bot-info-title">
    {selectedBot?.display_name || selectedBot?.name || "Bot details"}
    {#if selectedBot}
      {@const isFavorite = favorites.includes(selectedBot.name)}
      <button
        class="favorite-button"
        aria-label={isFavorite ? "Remove from favorites" : "Add to favorites"}
        aria-pressed={isFavorite}
        onclick={() => { if (selectedBot) toggleFavorite(selectedBot.name); }}
      >
        <img src={isFavorite ? filledStarIcon : starIcon} alt="" />
      </button>
    {/if}
  </span>
{/snippet}

<Modal title={botInfoTitle} bind:visible={showBotInfo}>
  {#if selectedBot}
    {@const url = sourceURL(selectedBot.source_link)}
    <div class="bot-info">
      <p><b>Developers:</b> {selectedBot.developer}</p>
      <p><b>Description:</b> {selectedBot.description}</p>
      <p><b>Fun fact:</b> {selectedBot.fun_fact}</p>
      <p><b>Source code:</b>
        {#if url}
          <a href={url} onclick={(event) => { event.preventDefault(); Browser.OpenURL(url); }} target="_blank" rel="noopener noreferrer">{selectedBot.source_link}</a>
        {:else}
          {selectedBot.source_link}
        {/if}
      </p>
      <p><b>Language:</b> {selectedBot.language}</p>
      {#if selectedBot.tags?.length}
        <div class="tags">
          <b>Tags:</b>
          {#each selectedBot.tags as tag}
            <span class="tag">{tag}</span>
          {/each}
        </div>
      {/if}

      {#if selectedBot.aliases?.length}
        <p><b>Aliases:</b> {selectedBot.aliases.join(", ")}</p>
      {/if}
    </div>
  {/if}
</Modal>

<Modal title="Select a launcher" bind:visible={launcherOptionsVisible}>
  <LauncherSelector />
</Modal>

<style>
  .bot-info-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .favorite-button {
    display: flex;
    padding: 0;
    background: transparent;
  }
  .favorite-button img {
    width: 24px;
    height: 24px;
    filter: invert() brightness(var(--icon-brightness));
  }
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
  .tag-buttons {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem;
    margin-bottom: 0.6rem;
  }
  .tag-group {
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    border: solid 1px gray;
    border-radius: 0.25rem;
    overflow: hidden;
  }
  .tag-buttons button {
    padding: 0.5rem 1rem;
    border-radius: 0;
    cursor: pointer;
  }
  .tag-buttons button.selected {
    background-color: var(--foreground);
    color: var(--background);
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
  .info-button {
    display: flex;
    align-items: center;
    padding: 0.2rem;
    background: transparent;
  }
  .info-button img {
    width: 1.2rem;
    height: 1.2rem;
    filter: invert() brightness(var(--icon-brightness));
  }
  .bot-info {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 42rem;
    overflow-wrap: anywhere;
  }
  .bot-info p {
    white-space: pre-wrap;
  }
  .tags {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
  }
  .tag {
    padding: 0.2rem 0.5rem;
    border-radius: 0.3rem;
    background: var(--background-alt);
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
