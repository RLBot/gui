<script lang="ts">
/** @import * from '../../bindings/gui' */
import toast from "svelte-5-french-toast";
// @ts-ignore
import {
  App,
  BotInfo,
  type StartMatchOptions,
} from "../../bindings/gui/index.js";
import arenaImages from "../arena-images";
import { MAPS_STANDARD } from "../arena-names";
import reloadIcon from "../assets/reload.svg";
import { BASE_PLAYERS } from "../base-players";
import BotList from "../components/BotList.svelte";
// @ts-ignore
import MatchSettings from "../components/MatchSettings/Main.svelte";
import PathsViewer from "../components/PathsViewer.svelte";
// @ts-ignore
import Teams from "../components/Teams/Main.svelte";
import { type DraggablePlayer, draggablePlayerToPlayerJs } from "../index";
import { mapStore } from "../settings";

const backgroundImage =
  arenaImages[Math.floor(Math.random() * arenaImages.length)];

let paths: {
  tagName: string | null;
  repo: string | null;
  installPath: string;
}[] = $state(
  JSON.parse(window.localStorage.getItem("BOT_SEARCH_PATHS") || "[]"),
);

let launcherOptionsVisible = $state(false);
let players: DraggablePlayer[] = $state([...BASE_PLAYERS]);
let selectedTeam = $state(null);

let loadingPlayers = $state(false);
let latestBotUpdateTime = null;
let showPathsViewer = $state(false);

async function updateBots() {
  loadingPlayers = true;
  let internalUpdateTime = new Date();
  latestBotUpdateTime = internalUpdateTime;
  const result = await App.GetBots(paths.map((x) => x.installPath));
  if (latestBotUpdateTime !== internalUpdateTime) {
    return; // if newer "search" already started, dont write old data
  }
  players = result.map((x: BotInfo) => {
    // @ts-ignore
    const n: typeof DraggablePlayer = {
      displayName: x.config.settings.name,
      icon: x.config.settings.logoFile,
      player: new BotInfo(x),
      id: Math.random(),
      tags: x.config.details.tags,
    };
    return n;
  });
  players = [...BASE_PLAYERS, ...players];
  loadingPlayers = false;
  console.log("Loaded bots:", result);
}

$effect(() => {
  window.localStorage.setItem("BOT_SEARCH_PATHS", JSON.stringify(paths));
  updateBots();
});

let bluePlayers: DraggablePlayer[] = $state([BASE_PLAYERS[0]]);
let orangePlayers: DraggablePlayer[] = $state([]);
let showHuman = $state(true);
$effect(() => {
  showHuman = !(
    bluePlayers.some((x) => x.tags.includes("human")) ||
    orangePlayers.some((x) => x.tags.includes("human"))
  );
});

let mode = $state(localStorage.getItem("MS_MODE") || "Soccer");
$effect(() => {
  localStorage.setItem("MS_MODE", mode);
});
let extraOptions = $state(
  JSON.parse(
    localStorage.getItem("MS_EXTRAOPTIONS") || '{"enableStateSetting": true}',
  ),
);
$effect(() => {
  localStorage.setItem("MS_EXTRAOPTIONS", JSON.stringify(extraOptions));
});
let mutatorSettings = $state(
  JSON.parse(localStorage.getItem("MS_MUTATORS") || "{}"),
);
$effect(() => {
  localStorage.setItem("MS_MUTATORS", JSON.stringify(mutatorSettings));
});

async function onMatchStart(randomizeMap: boolean) {
  const launcher = localStorage.getItem("MS_LAUNCHER");
  if (!launcher) {
    toast.error("Please select a launcher first", {
      position: "bottom-right",
      duration: 5000,
    });

    launcherOptionsVisible = true;
    return;
  }

  if (randomizeMap) {
    $mapStore =
      Object.values(MAPS_STANDARD)[
        Math.floor(Math.random() * Object.keys(MAPS_STANDARD).length)
      ];
  }

  const options: StartMatchOptions = {
    map: $mapStore,
    gameMode: mode,
    bluePlayers: bluePlayers.map((x: DraggablePlayer) => {
      // @ts-ignore
      return draggablePlayerToPlayerJs(x);
    }),
    orangePlayers: orangePlayers.map((x: DraggablePlayer) => {
      // @ts-ignore
      return draggablePlayerToPlayerJs(x);
    }),
    launcher,
    launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
    mutatorSettings,
    extraOptions,
  };

  toast("Starting match...", {
    position: "bottom-right",
  });

  const response = await App.StartMatch(options);

  if (response.success) {
    toast.success("Sent start match command", {
      position: "bottom-right",
      duration: 5000,
    });
  } else {
    toast.error(`Match start failed\n${response.message}`, {
      position: "bottom-right",
      duration: 5000,
    });
  }
}

async function onMatchStop() {
  toast("Stopping match...", {
    position: "bottom-right",
  });
  const response = await App.StopMatch(false);

  if (response.success) {
    toast.success("Sent stop match command", {
      position: "bottom-right",
      duration: 5000,
    });
  } else {
    toast.error(`Match stop failed\n${response.message}`, {
      position: "bottom-right",
      duration: 5000,
    });
  }
}

let searchQuery = $state("");

function handleSearch(event: Event) {
  searchQuery = (event.target as HTMLInputElement).value;
}
</script>

<div class="page" style={`background-image: url("${backgroundImage}")`}>
    <div class="availableBots box">
        <header>
            <h1>Bots</h1>
            <div class="dropdown">
                <button onclick={() => { showPathsViewer = true }}>Add/Remove</button>
            </div>
            {#if loadingPlayers}
                <h3>Searching...</h3>
            {:else}
                <button class="reloadButton" onclick={updateBots}
                    ><img src={reloadIcon} alt="reload" /></button
                >
            {/if}
            <div style="flex:1"></div>
            <input type="text" class="botSearch" placeholder="Search..." oninput={handleSearch}/>
        </header>
        <BotList
            bind:bluePlayers
            bind:orangePlayers
            bind:showHuman
            items={players}
            searchQuery={searchQuery}
            selectedTeam={selectedTeam}
        />
    </div>

    <div class="teams"><Teams bind:bluePlayers bind:orangePlayers bind:selectedTeam /></div>

    <div class="box">
        <MatchSettings
            onStart={onMatchStart}
            onStop={onMatchStop}
            bind:map={$mapStore}
            bind:mode
            bind:mutators={mutatorSettings}
            bind:extraOptions
            bind:launcherOptionsVisible
        />
    </div>
</div>

<PathsViewer bind:visible={showPathsViewer} bind:paths={paths} />

<style>
    .page {
        padding: 1rem;
        height: 100%;
        width: 100%;
        display: flex;
        flex-direction: column;
        overflow: auto;
        background-size: cover;
        background-repeat: no-repeat;
        background-position: center;
        background-attachment: fixed;
    }
    .page * {
        user-select: none;
        -webkit-user-select: none;
    }
    .box {
        border-radius: 0.4rem;
        background-color: var(--background);
        padding: 0.6rem;
    }
    .page > div:not(:first-child) {
        margin-top: 1rem;
    }
    .availableBots {
        padding-bottom: 0.6rem;
        display: flex;
        flex-direction: column;
    }
    .availableBots header {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin-bottom: 0.6rem;
    }
    .reloadButton {
        padding: 0px;
    }
    .reloadButton img {
        filter: invert();
    }
    .teams {
        overflow: auto;
        display: flex;
        flex-direction: column;
    }
</style>
