<script lang="ts">
import { draggable, droppable } from "@thisux/sveltednd";
import { Browser } from "@wailsio/runtime";
import SuperJSON from "superjson";
import { flip } from "svelte/animate";
import toast from "svelte-5-french-toast";
import { App, BotInfo, HumanInfo } from "../../bindings/gui";
import warningIcon from "../assets/exclamation-triangle-fill.svg";
import infoIcon from "../assets/info_icon.svg";
import defaultIcon from "../assets/rlbot_mono.png";
import starIcon from "../assets/star.svg";
import filledStarIcon from "../assets/starFilled.svg";
import { BASE_PLAYERS } from "../base-players";
import {
  type DraggablePlayer,
  parseJSON,
  type ToggleableScript,
} from "../index";
import { getAndParseItems } from "./LoadoutEditor/items";
import LoadoutEditor from "./LoadoutEditor/Main.svelte";
import Modal from "./Modal.svelte";
import Switch from "./Switch.svelte";

let {
  bots = [],
  scripts = [],
  showHuman = $bindable(),
  searchQuery = "",
  selectedTeam = null,
  enabledScripts = $bindable({}),
  bluePlayers = $bindable(),
  orangePlayers = $bindable(),
  map,
  duplicateAgentIds = new Set<string>(),
  getLatestBotInfo,
  getLatestScriptInfo,
}: {
  bots: DraggablePlayer[];
  scripts: ToggleableScript[];
  showHuman: boolean;
  searchQuery: string;
  selectedTeam: "blue" | "orange" | null;
  enabledScripts: { [key: string]: boolean };
  bluePlayers: DraggablePlayer[];
  orangePlayers: DraggablePlayer[];
  map: string;
  duplicateAgentIds: Set<string>;
  getLatestBotInfo?: (tomlPath: string) => Promise<BotInfo | null>;
  getLatestScriptInfo?: (tomlPath: string) => Promise<BotInfo | null>;
} = $props();
const flipDurationMs = 100;

let favorites: string[] = $state(
  parseJSON(localStorage.getItem("FAVORITES")) || [],
);
$effect(() => {
  localStorage.setItem("FAVORITES", JSON.stringify(favorites));
});

let selectedTags: (string | null)[] = $state([null, null]);

const Category = {
  All: "All",
  Standard: "Standard",
  ExtraModes: "Extra Modes",
  Special: "Special bots/scripts",
  Favorites: "Favorites",
};

const categories: string[][] = [
  [Category.All],
  [Category.Standard, Category.ExtraModes, Category.Special],
  [Category.Favorites],
];

const subCategories: { [x: string]: string[] } = {
  [Category.Standard]: ["Bots for 1v1", "Bots with teamplay", "Goalie bots"],
  [Category.ExtraModes]: [
    "Hoops",
    "Dropshot",
    "Snow Day",
    "Rumble",
    "Spike Rush",
    "Heatseeker",
  ],
};

const subCategoryTags: { [x: string]: string[] } = {
  [Category.Standard]: ["1v1", "teamplay", "goalie"],
  [Category.ExtraModes]: [
    "hoops",
    "dropshot",
    "snow-day",
    "rumble",
    "spike-rush",
    "heatseeker",
  ],
  [Category.Special]: ["memebot"],
};

let showInfoModal = $state(false);
let infoModalWasOpen = false;
let showLoadoutEditor = $state(false);
$effect(() => {
  if (!showLoadoutEditor && infoModalWasOpen) {
    showInfoModal = true;
    infoModalWasOpen = false;
  }
});

let everSelectedAgent = $state(false);
$effect(() => {
  if (selectedAgent) {
    everSelectedAgent = true;
  }
});

let selectedAgent: [BotInfo, string, string] | null = $state(null);

const filteredBots: DraggablePlayer[] = $derived.by(() =>
  filterBots(bots, selectedTags, showHuman, searchQuery),
);
const filteredScripts: ToggleableScript[] = $derived.by(() =>
  filterScripts(scripts, selectedTags, searchQuery),
);

function filterScripts(
  allScripts: ToggleableScript[],
  filterTags: (string | null)[],
  searchQuery: string,
) {
  if (filterTags[1]) {
    return [];
  }

  let filtered = allScripts;

  const mainTag = filterTags[0];
  if (mainTag) {
    filtered = filtered.filter((script) => {
      switch (mainTag) {
        case Category.Standard:
        case Category.ExtraModes:
          return script.tags.some((tag) =>
            subCategoryTags[mainTag].includes(tag),
          );
        case Category.Special:
          return true;
        case Category.Favorites:
          return favorites.includes(script.info.tomlPath);
        default:
          return true; // Unknown category selected, allow all
      }
    });
  }

  if (searchQuery) {
    filtered = filtered.filter((script) =>
      script.displayName.toLowerCase().includes(searchQuery.toLowerCase()),
    );
  }

  return filtered;
}

function filterBots(
  allBots: DraggablePlayer[],
  filterTags: (string | null)[],
  showHuman: boolean,
  searchQuery: string,
) {
  let filtered = allBots;

  const mainTag = filterTags[0];
  if (mainTag) {
    filtered = filtered.filter((bot) => {
      switch (mainTag) {
        case Category.Standard:
        case Category.ExtraModes:
        case Category.Special:
          return bot.tags.some((tag) => subCategoryTags[mainTag].includes(tag));
        case Category.Favorites:
          return bot.info instanceof BotInfo
            ? favorites.includes(bot.info.tomlPath)
            : false;
        default:
          return true; // Unknown category selected, allow all
      }
    });

    const subTag = filterTags[1];
    if (subTag) {
      filtered = filtered.filter((bot) => {
        return bot.tags.includes(
          subCategoryTags[mainTag][subCategories[mainTag].indexOf(subTag)],
        );
      });
    }
  }

  if (searchQuery) {
    filtered = filtered.filter((bot) =>
      bot.displayName.toLowerCase().includes(searchQuery.toLowerCase()),
    );
  } else if (showHuman) {
    filtered = [BASE_PLAYERS[0], ...filtered];
  }

  return filtered;
}

function scriptDuplicateAgentIdWarning(d: ToggleableScript): string | null {
  const agentId = d.info.config.settings.agentId;

  if (!agentId) {
    return "Empty agent id found. Please add an agent_id field with a unique value to the script's toml file.";
  }

  if (duplicateAgentIds.has(d.info.config.settings.agentId)) {
    return `Duplicate agent id "${agentId}" found in another active config`;
  }

  return null;
}

function handleTagClick(tag: string) {
  if (tag === Category.All || selectedTags[0] === tag) {
    selectedTags = [null, null];
  } else {
    selectedTags = [tag, null];
  }
}

function handleSubTagClick(tag: string) {
  selectedTags[1] = selectedTags[1] === tag ? null : tag;
}

function handleBotClick(bot: DraggablePlayer) {
  const newId = crypto.randomUUID();

  if (selectedTeam === "blue") {
    bluePlayers = [{ ...bot, id: newId }, ...bluePlayers];
  } else if (selectedTeam === "orange") {
    orangePlayers = [{ ...bot, id: newId }, ...orangePlayers];
  }
}

function toggleScript(id: string) {
  enabledScripts[id] = !enabledScripts[id];
}

async function handleBotInfoClick(bot: DraggablePlayer) {
  if (bot.info instanceof BotInfo) {
    try {
      if (getLatestBotInfo) {
        const refreshed = await getLatestBotInfo(bot.info.tomlPath);
        if (refreshed) {
          selectedAgent = [
            refreshed,
            refreshed.config?.settings?.name ?? bot.displayName,
            (refreshed as any).icon ?? bot.icon,
          ];
          showInfoModal = true;
          return;
        }
      }
    } catch (err) {
      toast.error(`Failed to get latest bot info: ${err}`, { duration: 10000 });
    }

    selectedAgent = [bot.info, bot.displayName, bot.icon];
    showInfoModal = true;
  }
}

async function handleScriptInfoClick(script: ToggleableScript) {
  try {
    if (getLatestScriptInfo) {
      const refreshed = await getLatestScriptInfo(script.info.tomlPath);
      if (refreshed) {
        selectedAgent = [
          refreshed,
          refreshed.config?.settings?.name ?? script.displayName,
          (refreshed as any).icon ?? script.icon,
        ];
        showInfoModal = true;
        return;
      }
    }
  } catch (err) {
    toast.error(`Failed to get latest script info: ${err}`, {
      duration: 10000,
    });
  }

  selectedAgent = [script.info, script.displayName, script.icon];
  showInfoModal = true;
}

function OpenSelectedBotSource() {
  if (selectedAgent) {
    Browser.OpenURL(selectedAgent[0].config.details.sourceLink);
  }
}

function EditSelectedBotLoadout() {
  if (selectedAgent) {
    infoModalWasOpen = showInfoModal;
    showInfoModal = false;
    showLoadoutEditor = true;
  }
}

function ShowSelectedBotFiles() {
  if (selectedAgent) {
    App.ShowPathInExplorer(selectedAgent[0].tomlPath).catch((err) => {
      toast.error(`${err}`, { duration: 10000 });
    });
  }
}

function SelectedToggleFavorite() {
  if (!selectedAgent) return;

  const path = selectedAgent[0].tomlPath;
  const index = favorites.indexOf(path);
  if (index !== -1) {
    favorites.splice(index, 1);
  } else {
    favorites.push(path);
  }
}
</script>

<div class="tag-buttons">
  {#each categories as tagGroup}
    <div class="tag-group">
      {#each tagGroup as tag}
        <button
          onclick={() => handleTagClick(tag)}
          class:selected={tag === Category.All ? selectedTags.every(t => t == null) : selectedTags[0] === tag}
        >
          {tag}
        </button>
      {/each}
    </div>
  {/each}
</div>

{#if selectedTags[0] && subCategories[selectedTags[0]]}
<div class="tag-buttons">
  <div class="tag-group">
    {#each subCategories[selectedTags[0]] as tag}
      <button
        onclick={() => handleSubTagClick(tag)}
        class:selected={selectedTags[1] === tag}
      >
        {tag}
      </button>
    {/each}
  </div>
</div>
{/if}

<div
  class="bots"
  use:droppable={{container: "botlist"}}
>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  {#each filteredBots as bot (bot.id)}
    <div
      class="bot blurred"
      use:draggable={{
        container: "botlist",
        dragData: SuperJSON.stringify(bot),
        interactive: [".info-button"]
      }}
      animate:flip={{ duration: flipDurationMs }}
      onclick={() => handleBotClick(bot)}
    >
      <img
        src={bot.icon || defaultIcon}
        alt="icon"
        style={ /* Fix light and dark theme for default icon */
          (!bot.icon || bot.info instanceof HumanInfo)
          ? "filter: brightness(var(--icon-brightness))" : ""
        }
      />
      <p>{bot.displayName}</p>
      {#if bot.uniquePathSegment}
        <span class="unique-bot-identifier">({bot.uniquePathSegment})</span>
      {/if}
      {#if bot.info && bot.info instanceof BotInfo}
        <button class="info-button" onclick={(e) => {e.stopPropagation();handleBotInfoClick(bot)}}>
          <img src={infoIcon} alt="i">
        </button>
      {:else}
        <!-- Empty div to keep gap consistent with the above case -->
        <div></div>
      {/if}
    </div>
  {/each}
</div>

{#if filteredBots.length === 0}
  <span>No bots available for this category.</span>
{/if}

{#if filteredScripts.length !== 0}
  <h3 id="scripts-header">Scripts</h3>
{/if}

<div class="bots">
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  {#each filteredScripts as script (script.id)}
    {@const warningText = scriptDuplicateAgentIdWarning(script)}

    <div class="bot blurred" animate:flip={{ duration: flipDurationMs }} onclick={() => toggleScript(script.info.config.settings.agentId)}>
      <Switch
        checked={enabledScripts[script.info.config.settings.agentId]}
        width={36}
        height={22}
        onchange={() => toggleScript(script.info.config.settings.agentId)}
      />
      <img
        src={script.icon || defaultIcon}
        alt="icon"
        style={ /* Fix light and dark theme for default icon */
          (!script.icon)
          ? "filter: brightness(var(--icon-brightness))" : ""
        }
      />
      <p>{script.displayName}</p>
      {#if script.uniquePathSegment}
        <span class="unique-bot-identifier">({script.uniquePathSegment})</span>
      {/if}
      {#if warningText && enabledScripts[script.info.config.settings.agentId]}
        <a href="javascript:void(0)" onclick={() => {Browser.OpenURL("https://wiki.rlbot.org/v5/botmaking/config-files/#bot-script-config-files")}} target="_blank">
          <img src={warningIcon} class="duplicate-agent-icon" alt={warningText} title={warningText} />
        </a>
      {/if}
      <button
        class="info-button"
        onclick={(e) => {
          e.stopPropagation();
          handleScriptInfoClick(script)
        }}
      >
        <img src={infoIcon} alt="i">
      </button>
    </div>
  {/each}
</div>

{#snippet botInfoTitle()}
{#if selectedAgent}
  {@const isFavorite = favorites.includes(selectedAgent[0].tomlPath)}
  <span id="bot-info-title">
    {selectedAgent[1]}
    <input
      type="checkbox"
      checked={isFavorite}
      onchange={SelectedToggleFavorite}
      id="favorite-checkbox"
    >
    <label for="favorite-checkbox">
      {#if isFavorite}
        <img src={filledStarIcon} alt="unmark as favorite">
      {:else}
        <img src={starIcon} alt="mark as favorite">
      {/if}
    </label>
  </span>
{/if}
{/snippet}

<Modal title={botInfoTitle} bind:visible={showInfoModal}>
{#if selectedAgent}
  <div class="info-layout">
    <div class="info-main">
      <p><b>Developers:</b> {selectedAgent[0].config.details.developer}</p>
      <p><b>Description:</b> {selectedAgent[0].config.details.description}</p>
      <p><b>Fun fact:</b> {selectedAgent[0].config.details.funFact}</p>
      <p><b>Source code:</b>
        <!-- svelte-ignore a11y_invalid_attribute -->
        <a href="javascript:void(0)" onclick={OpenSelectedBotSource} target="_blank">
          {selectedAgent[0].config.details.sourceLink}
        </a>
      </p>
      <p><b>Language:</b> {selectedAgent[0].config.details.language}</p>
      {#if selectedAgent[0].config.details.tags.length > 0}
      <div class="tags">
        <b>Tags:</b>
        {#each selectedAgent[0].config.details.tags as tag}
          <span class="tag">{tag}</span>
        {/each}
      </div>
      {/if}
    </div>

    {#if selectedAgent[2]}
    <div class="info-logo">
      <img src={selectedAgent[2]} alt="icon"/>
    </div>
    {/if}

    <div class="info-extra">
      <p class="toml-path">{selectedAgent[0].tomlPath}</p>
      <div class="button-group">
        {#if selectedAgent[0].loadout}
        <button onclick={EditSelectedBotLoadout}>Edit Loadout</button>
        {/if}
        <button onclick={ShowSelectedBotFiles}>Show Files</button>
      </div>
    </div>
  </div>
{/if}
</Modal>

<!-- prevent loading the items if unneeded,
 but also prevent loading the items more than once -->
{#if everSelectedAgent}
  <!-- svelte-ignore block_empty -->
  {#await getAndParseItems() then items}
    {#if selectedAgent && selectedAgent[0].loadout}
      <LoadoutEditor
        bind:visible={showLoadoutEditor}
        basePath={selectedAgent[0].tomlPath}
        loadoutFile={selectedAgent[0].config.settings.loadoutFile}
        loadout={selectedAgent[0].loadout}
        {items}
        name={selectedAgent[1]}
        {map}
      />
    {/if}
  {/await}
{/if}

<style>
  .bots span {
    color: gray;
  }
  .tag-buttons {
    display: flex;
    flex-direction: row;
    justify-content: center;

    gap: 0.5rem;
    margin-bottom: .6rem;
  }
  .tag-group {
    display: flex;
    gap: 2px;
    border: solid 1px gray;
    --border-radius: 0.25rem;
    border-radius: var(--border-radius);
  }
  .tag-buttons button {
    padding: 0.5rem 1rem;
    border-radius: 0;
    cursor: pointer;
    /*background-color: var(--background-alt);*/
  }
  .tag-buttons button:first-child {
    border-radius: var(--border-radius) 0 0 var(--border-radius);
  }
  .tag-buttons button:last-child {
    border-radius: 0 var(--border-radius) var(--border-radius) 0;
  }
  .tag-buttons button:only-child {
    border-radius: var(--border-radius);
  }
  .tag-buttons button.selected {
    background-color: var(--foreground);
    color: var(--background);
    /* font-size: 1.1rem; */
  }
  .bots {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .bot {
    display: flex;
    align-items: center;
    background-color: var(--background-alt);
    color: var(--foreground);
    height: 2.25rem;
    padding: 0.2rem;
    /* padding-right: 0.6rem; */
    gap: 0.5rem;
    border-radius: 0.2rem;
    cursor: pointer;
  }
  .bot img {
    height: 2rem;
    width: auto;
    /*user-drag: none;*/
    -webkit-user-drag: none;
  }
  .info-button {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    height: 100%;
    padding: 0;
    color: var(--foreground);
    cursor: pointer;
    font-size: 1rem;
  }
  .info-button img {
    filter: invert() brightness(var(--icon-brightness));
    height: 100%;
    width: auto;
  }
  .tags {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    margin-top: auto;
  }
  .tag {
    background-color: grey;
    color: white;
    padding: 0.2rem 0.5rem;
    border-radius: 0.25rem;
    margin: 0 0.3rem;
  }
  .info-layout {
    display: grid;
    gap: 1rem;
    max-width: 800px;
  }
  .info-layout {
    grid-template-columns: 1fr;
    grid-template-areas:
    "main"
    "logo"
    "extra";
  }
  @media (min-width: 850px) {
    .info-layout {
    grid-template-columns: 1fr auto;
    grid-template-areas:
        "main logo"
        "extra logo";
    }
  }
  .info-main {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    grid-area: main;
  }
  .info-main a {
    overflow-wrap: anywhere;
    word-break: break-word;
  }
  .info-extra {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    grid-area: extra;
  }
  .info-logo {
    display: flex;
    align-items: center;
    justify-content: center;
    grid-area: logo;
  }
  .info-logo img {
    max-width: 250px;
    max-height: 250px;
    width: auto;
    height: auto;
  }
  .unique-bot-identifier {
    color: #868686;
  }
  .button-group {
    display: flex;
    gap: 1rem;
  }
  .button-group button {
    background: var(--background-alt);
    color: var(--foreground);
    cursor: pointer;
    font-size: 1rem;
    align-self: flex-start;
  }
  .toml-path {
    font-size: 0.8rem;
    color: grey;
  }
  .duplicate-agent-icon {
    max-height: 1.5rem;
    /* filter calculated with https://codepen.io/sosuke/pen/Pjoqqp */
    filter: invert(83%) sepia(55%) saturate(4490%) hue-rotate(5deg) brightness(103%) contrast(102%);
  }
  #scripts-header {
    margin-top: 12px;
    margin-bottom: 5px;
    font-weight: bold;
  }
  #bot-info-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  #bot-info-title label {
    display: flex;
    align-items: center;
    cursor: pointer;
  }
  #bot-info-title input {
    display: none;
  }
  #bot-info-title img {
    filter: invert();
    height: 24px;
  }
</style>
