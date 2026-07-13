<script lang="ts">
/** @import * from '../../bindings/gui' */

import SuperJSON from "superjson";
import toast from "svelte-5-french-toast";
import {
  DebugRendering,
  ExistingMatchBehavior,
  PerformanceMonitor,
} from "../../bindings/github.com/RLBot/go-interface/flat/models.js";
import {
  App,
  BotInfo,
  ExtraOptions,
  Result,
  type StartMatchOptions,
} from "../../bindings/gui/index.js";
import { MAPS_STANDARD } from "../arena-names";
import reloadIcon from "../assets/reload.svg";
import { BASE_PLAYERS } from "../base-players";
import BotList from "../components/BotList.svelte";
import BotpackNotif from "../components/BotpackToast.svelte";
import MatchSettings from "../components/MatchSettings/Main.svelte";
import PathsViewer from "../components/PathsViewer.svelte";
import Teams from "../components/Teams/Main.svelte";
import {
  type DraggablePlayer,
  draggablePlayerToPlayerJs,
  parseJSON,
  parseSuperJSON,
  type ToggleableScript,
} from "../index";
import { mapStore } from "../settings";

let {
  paths = $bindable([]),
}: {
  paths?: {
    tagName: string | null;
    repo: string | null;
    installPath: string;
    visible: boolean;
    isDependency: boolean;
  }[];
} = $props();

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
          // @ts-expect-error
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
let bluePlayers: DraggablePlayer[] = $state(
  parseSuperJSON(localStorage.getItem("BLUE_PLAYERS")) || [BASE_PLAYERS[0]],
);
let orangePlayers: DraggablePlayer[] = $state(
  parseSuperJSON(localStorage.getItem("ORANGE_PLAYERS")) || [],
);
let showHuman = $derived(
  !(
    bluePlayers.some((x) => x.tags.includes("human")) ||
    orangePlayers.some((x) => x.tags.includes("human"))
  ),
);

$effect(() => {
  localStorage.setItem("BLUE_PLAYERS", SuperJSON.stringify(bluePlayers));
});
$effect(() => {
  localStorage.setItem("ORANGE_PLAYERS", SuperJSON.stringify(orangePlayers));
});

let latestScriptUpdateTime = null;
let loadingScripts = $state(false);
let scripts: ToggleableScript[] = $state([]);
const enabledScriptsStorageKey = "ENABLED_SCRIPTS";
let enabledScripts: { [key: string]: boolean } = $state(
  parseJSON(localStorage.getItem(enabledScriptsStorageKey)) || {},
);

$effect(() => {
  localStorage.setItem(
    enabledScriptsStorageKey,
    JSON.stringify(enabledScripts),
  );
});

function distinguishDuplicates(pool: BotInfo[]): [BotInfo, string?][] {
  const uniqueNames = [
    ...new Set(
      pool.filter((bot) => bot.tomlPath).map((bot) => bot.config.settings.name),
    ),
  ];
  const splitPath = (bot: BotInfo) => bot.tomlPath.split(/[\\|/]/).reverse();

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

function collectDuplicateAgentIds(
  bots: DraggablePlayer[],
  scripts: ToggleableScript[],
): Set<string> {
  const agentIdTomlMap: { [id: string]: string } = {};
  const duplicateAgentIds = new Set<string>();

  for (const bot of bots) {
    if (bot.info instanceof BotInfo) {
      const agentId = bot.info.config.settings.agentId;
      const path = bot.info.tomlPath;

      if (agentId in agentIdTomlMap && agentIdTomlMap[agentId] !== path) {
        duplicateAgentIds.add(agentId);
      } else {
        agentIdTomlMap[agentId] = path;
      }
    }
  }

  for (const script of scripts) {
    const agentId = script.info.config.settings.agentId;
    const path = script.info.tomlPath;

    if (agentId in agentIdTomlMap && agentIdTomlMap[agentId] !== path) {
      duplicateAgentIds.add(agentId);
    } else {
      agentIdTomlMap[agentId] = path;
    }
  }

  return duplicateAgentIds;
}

const duplicateAgentIds = $derived.by(() =>
  collectDuplicateAgentIds(bluePlayers.concat(orangePlayers), scripts),
);

function updateTeam(team: DraggablePlayer[]) {
  let newTeam: DraggablePlayer[] = [];
  for (let player of team) {
    const botInfo = player.info;
    if (!(botInfo instanceof BotInfo)) {
      newTeam.push(player);
      continue;
    }

    const found = players.find(
      (p) => p.info instanceof BotInfo && p.info.tomlPath === botInfo.tomlPath,
    );
    if (!found) {
      // bot was removed
      continue;
    }

    let newPlayer: DraggablePlayer = {
      ...found,
      id: player.id,
      overrides: player.overrides,
    };

    newTeam.push(newPlayer);
  }

  return newTeam;
}

async function updateBots() {
  loadingPlayers = true;
  const internalUpdateTime = new Date();
  latestBotUpdateTime = internalUpdateTime;

  players = [...BASE_PLAYERS.slice(1)];

  const result = await App.GetBots(
    paths.filter((x) => x.visible).map((x) => x.installPath),
  );

  if (latestBotUpdateTime !== internalUpdateTime) {
    return; // if newer "search" already started, dont write old data
  }

  players = players.concat(
    distinguishDuplicates(result).map(([x, uniquePathSegment]) => {
      return {
        displayName: x.config.settings.name,
        icon: x.icon,
        info: new BotInfo(x),
        id: crypto.randomUUID(),
        tags: x.config.details.tags,
        uniquePathSegment,
        overrides: {
          name: null,
          loadout: null,
          autoStart: true,
        },
      };
    }),
  );

  bluePlayers = updateTeam(bluePlayers);
  orangePlayers = updateTeam(orangePlayers);

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
      id: crypto.randomUUID(),
      displayName: x.config.settings.name,
      icon: x.config.settings.logoFile,
      info: x,
      tags: x.config.details.tags,
      uniquePathSegment,
    };
  });

  for (const script of scripts) {
    const agentId = script.info.config.settings.agentId;

    if (enabledScripts[agentId] === undefined) {
      enabledScripts[agentId] = false;
    }
  }

  loadingScripts = false;
}

async function getLatestBotInfo(tomlPath: string): Promise<BotInfo | null> {
  try {
    const result = await App.GetBots([tomlPath]);

    const found = result.find((b) => b.tomlPath === tomlPath);
    if (!found) return null; // No bot found

    const botInfo = new BotInfo(found);

    const index = players.findIndex(
      (p) => p.info instanceof BotInfo && p.info.tomlPath === tomlPath,
    );

    if (index !== -1) {
      players[index].info = botInfo;
      players[index].displayName = found.config.settings.name;
      players[index].icon = found.icon;
      players[index].tags = found.config.details.tags;

      bluePlayers = updateTeam(bluePlayers);
      orangePlayers = updateTeam(orangePlayers);
    }

    return botInfo;
  } catch (err) {
    console.error("Failed to get latest bot info: ", err);
    return null;
  }
}

async function getLatestScriptInfo(tomlPath: string): Promise<BotInfo | null> {
  try {
    const result = await App.GetScripts([tomlPath]);

    const found = result.find((s) => s.tomlPath === tomlPath);
    if (!found) return null; // No script found

    const index = scripts.findIndex((s) => s.info.tomlPath === tomlPath);

    if (index !== -1) {
      const oldAgentId = scripts[index].info.config.settings.agentId;
      const newAgentId = found.config.settings.agentId;

      scripts[index].displayName = found.config.settings.name;
      scripts[index].icon = found.config.settings.logoFile;
      scripts[index].info = found;
      scripts[index].tags = found.config.details.tags;

      enabledScripts[newAgentId] = enabledScripts[oldAgentId];
      if (oldAgentId !== newAgentId) {
        delete enabledScripts[oldAgentId];
      }
    }

    return found;
  } catch (err) {
    console.error("Failed to get latest script info: ", err);
    return null;
  }
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

let mode = $state(localStorage.getItem("MS_MODE") || "Soccar");
$effect(() => {
  localStorage.setItem("MS_MODE", mode);
});

let extraOptions: ExtraOptions = $state({
  existingMatchBehavior: ExistingMatchBehavior.ExistingMatchBehaviorRestart,
  enableRendering: DebugRendering.DebugRenderingOffByDefault,
  performanceMonitor: PerformanceMonitor.PerformanceMonitorShowWhenSuboptimal,
  enableStateSetting: true,
  autoStartAgents: true,
  waitForAgents: true,
  // rest are fine with being nullish
  ...(parseJSON(localStorage.getItem("MS_EXTRAOPTIONS")) || {}),
});
$effect(() => {
  localStorage.setItem("MS_EXTRAOPTIONS", JSON.stringify(extraOptions));
});
let mutatorSettings = $state(
  parseJSON(localStorage.getItem("MS_MUTATORS")) || {},
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

  // Update bots and scripts
  const botsInMatch = [
    ...new Set(
      [...bluePlayers, ...orangePlayers]
        .filter((p) => p.info instanceof BotInfo)
        .map((p) => (p.info as BotInfo).tomlPath),
    ),
  ];

  const scriptsInMatch = [
    ...new Set(
      scripts // We could reuse this for StartMatchOptions, and not map twice
        .filter((x) => enabledScripts[x.info.config.settings.agentId])
        .map((x) => x.info.tomlPath),
    ),
  ];

  await Promise.all([
    ...botsInMatch.map((tomlPath) => getLatestBotInfo(tomlPath)),
    ...scriptsInMatch.map((tomlPath) => getLatestScriptInfo(tomlPath)),
  ]);

  const options: StartMatchOptions = {
    map: $mapStore,
    gameMode: mode,
    scripts: scripts
      .filter((x) => enabledScripts[x.info.config.settings.agentId])
      .map((x) => x.info),
    bluePlayers: bluePlayers.map(draggablePlayerToPlayerJs),
    orangePlayers: orangePlayers.map(draggablePlayerToPlayerJs),
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

  let response: Result;
  try {
    response = await App.StartMatch(options);
  } catch (e) {
    toast.error(`Match start failed\n${e}`, {
      id: toastId,
      duration: 10000,
    });
    return;
  }

  if (toastId !== startMatchToastId) return;
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

<div class="page">
  <div class="availableBots box blurred">
    <header>
      <h1>Bots</h1>
      <div class="dropdown">
        <button onclick={() => { showPathsViewer = true }}>Add/Remove</button>
      </div>
      <button
        class="reloadButton"
        title="(note: you need to re-add a bot to a team to apply changes)"
        onclick={loadPaths}
        disabled={loadingPlayers || loadingScripts}
      ><img src={reloadIcon} alt="reload" /></button>
      {#if loadingPlayers || loadingScripts}
        <h3>Searching...</h3>
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
      {duplicateAgentIds}
      getLatestBotInfo={getLatestBotInfo}
      getLatestScriptInfo={getLatestScriptInfo}
    />
  </div>

  <div class="teams">
    <Teams
      bind:bluePlayers
      bind:orangePlayers
      bind:selectedTeam
      bind:globalAutoStart={extraOptions.autoStartAgents}
      {duplicateAgentIds}
    />
  </div>

  <div class="box blurred">
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
    filter: invert() brightness(var(--icon-brightness));
  }
  .teams {
    display: flex;
    flex-direction: column;
  }
</style>
