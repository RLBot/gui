<script lang="ts">
/** @import * from '../../bindings/gui' */
import toast from "svelte-5-french-toast";
// @ts-ignore
import {
  App,
  BotInfo,
  ExtraOptions,
  type StartMatchOptions,
} from "../../bindings/gui/index.js";
import arenaImages from "../arena-images";
import { MAPS_STANDARD } from "../arena-names";
import reloadIcon from "../assets/reload.svg";
import { BASE_PLAYERS } from "../base-players";
import BotList from "../components/BotList.svelte";
import BotpackNotif from "../components/BotpackToast.svelte";
// @ts-ignore
import MatchSettings from "../components/MatchSettings/Main.svelte";
import PathsViewer from "../components/PathsViewer.svelte";
// @ts-ignore
import Teams from "../components/Teams/Main.svelte";
import {
  type DraggablePlayer,
  type ToggleableScript,
  draggablePlayerToPlayerJs,
  uuidv4,
} from "../index";
import { mapStore } from "../settings";

const backgroundImage =
  arenaImages[Math.floor(Math.random() * arenaImages.length)];

let paths: {
  tagName: string | null;
  repo: string | null;
  installPath: string;
  visible: boolean;
}[] = $state(
  JSON.parse(window.localStorage.getItem("BOT_SEARCH_PATHS") || "[]"),
);

let botpackNotifIds: { [repo: string]: string } = {};

function updateBotpack(repoName: string) {
  const notifId = botpackNotifIds[repoName];
  if (notifId) {
    toast.dismiss(notifId);
  }

  const details = paths.find((x) => x.repo === repoName);
  if (details?.repo && details.tagName) {
    const tId = toast.loading(`Updating ${repoName}...`, {
      position: "top-center",
    });

    App.UpdateBotpack(details.repo, details.installPath, details.tagName)
      .then((newTagName) => {
        details.tagName = newTagName;
        toast.success(`${repoName} updated successfully`, {
          id: tId,
          position: "top-center",
          duration: 3000,
        });
      })
      .catch((error) => {
        console.error(error);
        toast.error(`Failed to update ${repoName}: ${error}`, {
          id: tId,
          position: "top-center",
          duration: 10000,
        });
      });
  }
}

function CheckForBotpackUpdates() {
  for (const path of paths) {
    if (path.visible && path.repo && path.tagName) {
      const repoName = path.repo;

      App.CheckForNewRelease(repoName, path.tagName).then((release) => {
        if (release) {
          // @ts-ignore
          const tId = toast(BotpackNotif, {
            props: {
              repoName,
              updateBotpack,
            },
            style: "max-width: 500px",
            position: "top-center",
            duration: 10000,
          });

          botpackNotifIds[repoName] = tId;
        }
      });
    }
  }
}

CheckForBotpackUpdates();

let launcherOptionsVisible = $state(false);
let selectedTeam = $state(null);
let showPathsViewer = $state(false);

let latestBotUpdateTime = null;
let loadingPlayers = $state(false);

let players: DraggablePlayer[] = $state(BASE_PLAYERS.slice(1));
let bluePlayers: DraggablePlayer[] = $state([BASE_PLAYERS[0]]);
let orangePlayers: DraggablePlayer[] = $state([]);
let showHuman = $derived(
  !(
    bluePlayers.some((x) => x.tags.includes("human")) ||
    orangePlayers.some((x) => x.tags.includes("human"))
  ),
);

let latestScriptUpdateTime = null;
let loadingScripts = $state(false);
let scripts: ToggleableScript[] = $state([]);
let enabledScripts: { [key: string]: boolean } = $state({});

function distinguishDuplicates(pool: BotInfo[]): [BotInfo, string?][] {
  const uniqueNames = [
    ...new Set(
      pool.filter((bot) => bot.tomlPath).map((bot) => bot.config.settings.name),
    ),
  ];
  const splitPath = (bot: BotInfo) => bot.tomlPath.split(/[\\|\/]/).reverse();

  let uniquePathSegments: [BotInfo, string?][] = [];

  for (const name of uniqueNames) {
    const bots = pool.filter((bot) => bot.config.settings.name === name);
    if (bots.length === 1) {
      uniquePathSegments.push([bots[0]]);
      continue;
    }

    for (let i = 0; bots.length > 0 && i < 99; i++) {
      const pathSegments = bots.map((b) => splitPath(b)[i]);

      for (const bot of bots.slice()) {
        const path = splitPath(bot);
        const count = pathSegments.filter((s) => s === path[i]).length;
        if (count === 1) {
          uniquePathSegments.push([bot, path[i]]);
          bots.splice(bots.indexOf(bot), 1);
        }
      }
    }
  }

  return uniquePathSegments;
}

async function updateBots() {
  loadingPlayers = true;
  const internalUpdateTime = new Date();
  latestBotUpdateTime = internalUpdateTime;
  const result = await App.GetBots(
    paths.filter((x) => x.visible).map((x) => x.installPath),
  );
  if (latestBotUpdateTime !== internalUpdateTime) {
    return; // if newer "search" already started, dont write old data
  }
  players = distinguishDuplicates(result).map(([x, uniquePathSegment]) => {
    return {
      displayName: x.config.settings.name,
      icon: x.config.settings.logoFile,
      player: new BotInfo(x),
      id: uuidv4(),
      tags: x.config.details.tags,
      uniquePathSegment,
      modified: false,
    };
  });

  players = [...BASE_PLAYERS.slice(1), ...players];
  loadingPlayers = false;
}

async function updateScripts() {
  loadingScripts = true;
  let internalUpdateTime = new Date();
  latestScriptUpdateTime = internalUpdateTime;
  const result = await App.GetScripts(
    paths.filter((x) => x.visible).map((x) => x.installPath),
  );
  if (latestScriptUpdateTime !== internalUpdateTime) {
    return; // if newer "search" already started, dont write old data
  }
  scripts = distinguishDuplicates(result).map(([x, uniquePathSegment]) => {
    return {
      id: uuidv4(),
      displayName: x.config.settings.name,
      icon: x.config.settings.logoFile,
      config: x,
      tags: x.config.details.tags,
      uniquePathSegment,
    };
  });

  for (const script of scripts) {
    if (enabledScripts[script.id] === undefined) {
      enabledScripts[script.id] = false;
    }
  }

  for (const id in Object.keys(enabledScripts)) {
    if (!scripts.some((script) => script.id === id)) {
      delete enabledScripts[id];
    }
  }

  loadingScripts = false;
}

$effect(() => {
  localStorage.setItem("BOT_SEARCH_PATHS", JSON.stringify(paths));
  updateBots();
  updateScripts();
});

function loadPaths() {
  updateBots();
  updateScripts();
}

let mode = $state(localStorage.getItem("MS_MODE") || "Soccer");
$effect(() => {
  localStorage.setItem("MS_MODE", mode);
});

function loadExtraOptions(): ExtraOptions {
  let extraOptions = JSON.parse(
    localStorage.getItem("MS_EXTRAOPTIONS") || '{"existingMatchBehavior": 0}',
  );

  // old versions of the GUI will have MS_EXTRAOPTIONS but might not have these values,
  // and they should default to true
  const newDefaultTrue = [
    "autoStartAgents",
    "waitForAgents",
    "enableStateSetting",
  ];
  for (const item of newDefaultTrue) {
    if (extraOptions[item] === undefined) {
      extraOptions[item] = true;
    }
  }

  return extraOptions;
}

let extraOptions = $state(loadExtraOptions());
$effect(() => {
  localStorage.setItem("MS_EXTRAOPTIONS", JSON.stringify(extraOptions));
});
let mutatorSettings = $state(
  JSON.parse(localStorage.getItem("MS_MUTATORS") || "{}"),
);
$effect(() => {
  localStorage.setItem("MS_MUTATORS", JSON.stringify(mutatorSettings));
});

let startMatchToastId: string | null = null;

async function onMatchStart(randomizeMap: boolean) {
  const launcher = localStorage.getItem("MS_LAUNCHER");
  if (!launcher) {
    toast.error("Please select a launcher first", {
      position: "top-center",
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
    scripts: scripts.filter((x) => enabledScripts[x.id]).map((x) => x.config),
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

  // only show the toast from the newest start match attempt
  if (startMatchToastId) {
    toast.dismiss(startMatchToastId);
  }

  const toastId = toast.loading("Starting match...", {
    position: "top-center",
  });
  startMatchToastId = toastId;

  const response = await App.StartMatch(options);

  if (toastId != startMatchToastId) return;
  startMatchToastId = null;

  if (response.success) {
    toast.success("Match started", {
      id: toastId,
    });
  } else {
    toast.error(`Match start failed\n${response.message}`, {
      id: toastId,
      duration: 10000,
    });
  }
}

async function onMatchStop() {
  const id = startMatchToastId ?? undefined;
  const response = await App.StopMatch(false);

  if (response.success) {
    toast.success("Sent stop match command", {
      id,
    });
  } else {
    toast.error(`Match stop failed\n${response.message}`, {
      id,
      duration: 10000,
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
      {#if loadingPlayers || loadingScripts}
        <h3>Searching...</h3>
      {:else}
        <button class="reloadButton" onclick={loadPaths}
          ><img src={reloadIcon} alt="reload" /></button
        >
      {/if}
      <div style="flex:1"></div>
      <input type="text" class="botSearch" placeholder="Search..." oninput={handleSearch}/>
    </header>
    <BotList
      bind:enabledScripts
      bind:bluePlayers
      bind:orangePlayers
      bind:showHuman
      bots={players}
      scripts={scripts}
      searchQuery={searchQuery}
      selectedTeam={selectedTeam}
      map={$mapStore}
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

<PathsViewer bind:visible={showPathsViewer} bind:paths />

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
    min-height: 93px;
    overflow: auto;
    display: flex;
    flex-direction: column;
  }
</style>
