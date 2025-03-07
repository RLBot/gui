<script lang="ts">
import { Browser } from "@wailsio/runtime";
import toast from "svelte-5-french-toast";
import {
  SHADOW_ITEM_MARKER_PROPERTY_NAME,
  TRIGGERS,
  dndzone,
} from "svelte-dnd-action";
import { flip } from "svelte/animate";
import { App, BotInfo } from "../../bindings/gui";
import infoIcon from "../assets/info_icon.svg";
import defaultIcon from "../assets/rlbot_mono.png";
import type { DraggablePlayer } from "../index";
import Modal from "./Modal.svelte";

let {
  items = [],
  showHuman = $bindable(true),
  searchQuery = "",
  selectedTeam = null,
  bluePlayers = $bindable(),
  orangePlayers = $bindable(),
}: {
  items: DraggablePlayer[];
  showHuman: boolean;
  searchQuery: string;
  selectedTeam: "blue" | "orange" | null;
  bluePlayers: DraggablePlayer[];
  orangePlayers: DraggablePlayer[];
} = $props();
const flipDurationMs = 100;

let selectedTags: (string | null)[] = $state([null, null]);
const extraModeTags = [
  "hoops",
  "dropshot",
  "snow-day",
  "spike-rush",
  "heatseaker",
];
const categories = [
  ["All"],
  ["Standard", "Extra Modes", "Special bots/scripts"],
  ["Bots for 1v1", "Bots with teamplay", "Goalie bots"],
];

let filteredItems: DraggablePlayer[] = $state([]);
let showModal = $state(false);
let selectedBot: [BotInfo, string, string] | null = $state(null);

$effect(() => {
  filteredItems = filterBots(selectedTags, showHuman, searchQuery);
});

function filterBots(
  filterTags: (string | null)[],
  showHuman: boolean,
  searchQuery: string,
) {
  let filtered = items.slice(1);

  if (filterTags[0]) {
    filtered = filtered.filter((bot) => {
      switch (filterTags[0]) {
        case categories[1][0]:
          return !bot.tags.some((tag) =>
            [...extraModeTags, "memebot", "human"].includes(tag),
          );
        case categories[1][1]:
          return bot.tags.some((tag) => extraModeTags.includes(tag));
        case categories[1][2]:
          return bot.tags.some((tag) => tag === "memebot");
        default:
          return true;
      }
    });
  }

  if (filterTags[1]) {
    filtered = filtered.filter((bot) => {
      switch (filterTags[1]) {
        case categories[2][0]:
          return bot.tags.some((tag) => tag === "1v1");
        case categories[2][1]:
          return bot.tags.some((tag) => tag === "teamplay");
        case categories[2][2]:
          return bot.tags.some((tag) => tag === "goalie");
        default:
          return true;
      }
    });
  }

  if (showHuman) {
    filtered = [items[0], ...filtered];
  }

  if (searchQuery) {
    filtered = filtered.filter((bot) =>
      bot.displayName.toLowerCase().includes(searchQuery.toLowerCase()),
    );
  }

  return filtered;
}

function handleTagClick(tag: string, groupIndex: number) {
  if (groupIndex === 0) {
    selectedTags = [null, null];
  } else {
    selectedTags[groupIndex - 1] =
      selectedTags[groupIndex - 1] === tag ? null : tag;
  }
}

function handleDndConsider(e: any) {
  const { trigger, id } = e.detail.info;
  if (trigger === TRIGGERS.DRAG_STARTED) {
    const newId = `${id}-${Math.round(Math.random() * 100000)}`;
    const idx = filteredItems.findIndex((item) => item.id === id);
    e.detail.items = e.detail.items.filter(
      (item: any) => !item[SHADOW_ITEM_MARKER_PROPERTY_NAME],
    );
    e.detail.items.splice(idx, 0, { ...filteredItems[idx], id: newId });
    filteredItems = e.detail.items;
  }
}
function handleDndFinalize(e: any) {
  filteredItems = e.detail.items;
}

function handleBotClick(bot: DraggablePlayer) {
  const newId = `${bot.id}-${Math.round(Math.random() * 100000)}`;
  const idx = filteredItems.findIndex((item) => item.id === bot.id);
  //@ts-ignore
  filteredItems.splice(idx, 1, { ...filteredItems[idx], id: newId });

  if (selectedTeam === "blue") {
    bluePlayers = [bot, ...bluePlayers];
  } else if (selectedTeam === "orange") {
    orangePlayers = [bot, ...orangePlayers];
  }
}

function handleInfoClick(bot: DraggablePlayer) {
  if (bot.player instanceof BotInfo) {
    selectedBot = [bot.player, bot.displayName, bot.icon];
    showModal = true;
  }
}

function OpenSelectedBotSource() {
  if (selectedBot) {
    Browser.OpenURL(selectedBot[0].config.details.sourceLink);
  }
}

function EditSelectedBotAppearance() {
  if (selectedBot) {
    alert.bind(null, "Not implemented yet")();
  }
}

function ShowSelectedBotFiles() {
  if (selectedBot) {
    App.ShowPathInExplorer(selectedBot[0].tomlPath).catch((err) =>
      toast.error(err, { duration: 10000 }),
    );
  }
}
</script>

<div class="tag-buttons">
  {#each categories as tagGroup, groupIndex}
    <div class="tag-group">
      {#each tagGroup as tag}
        <button
          onclick={() => handleTagClick(tag, groupIndex)}
          class:selected={tag === categories[0][0] ? selectedTags.every(t => t == null) : selectedTags[groupIndex-1] === tag}
        >
          {tag}
        </button>
      {/each}
    </div>
  {/each}
</div>

<div
  class="bots"
  use:dndzone={{
    items: filteredItems,
    flipDurationMs,
    centreDraggedOnCursor: true,
    dropTargetStyle: {},
    dropTargetClasses: ["dropTarget"],
  }}
  onconsider={handleDndConsider}
  onfinalize={handleDndFinalize}
>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  {#each filteredItems as bot (bot.id)}
    <div class="bot" animate:flip={{ duration: flipDurationMs }} onclick={() => handleBotClick(bot)}>
      <img src={bot.icon || defaultIcon} alt="icon" />
      <p>{bot.displayName}</p>
      {#if bot.uniquePathSegment}
        <span class="unique-bot-identifier">({bot.uniquePathSegment})</span>
      {/if}
      {#if bot.player && bot.player instanceof BotInfo}
        <button class="info-button" onclick={(e) => {e.stopPropagation();handleInfoClick(bot)}}>
          <img src={infoIcon} alt="i">
        </button>
      {:else}
        <!-- Empty div to keep gap consistent with the above case -->
        <div></div>
      {/if}
    </div>
  {/each}
</div>

{#if filteredItems.length === 0}
  <span>No bots available for this category.</span>
{/if}

<Modal title={selectedBot ? selectedBot[1] : ""} bind:visible={showModal}>
{#if selectedBot}
  <div class="modal-content">
    <div class="bot-left-column">
      <p>Developers: {selectedBot[0].config.details.developer}</p>
      <p>Description: {selectedBot[0].config.details.description}</p>
      <p>Fun fact: {selectedBot[0].config.details.funFact}</p>
      <p>Source code:
        <!-- svelte-ignore a11y_invalid_attribute -->
        <a href="#" onclick={OpenSelectedBotSource} target="_blank">
          {selectedBot[0].config.details.sourceLink}
        </a>
      </p>
      <p>Language: {selectedBot[0].config.details.language}</p>
      {#if selectedBot[0].config.details.tags.length > 0}
      <div class="tags">
        Tags:
        {#each selectedBot[0].config.details.tags as tag}
          <span class="tag">{tag}</span>
        {/each}
      </div>
      {/if}
      <p id="toml-path">{selectedBot[0].tomlPath}</p>
      <div id="button-group">
        <button onclick={EditSelectedBotAppearance}>Edit Appearance</button>
        <button onclick={ShowSelectedBotFiles}>Show Files</button>
      </div>
    </div>
    {#if selectedBot[2]}
    <div class="bot-right-column">
      <img src={selectedBot[2]} alt="icon" />
    </div>
    {/if}
  </div>
{/if}
</Modal>

<style>
  span {
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
    background-color: var(--background-alt);
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
    filter: invert() brightness(90%);
    height: 100%;
    width: auto;
  }
  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 1rem;
  }
  .tag {
    background-color: grey;
    color: white;
    padding: 0.2rem 0.5rem;
    border-radius: 0.25rem;
  }
  .modal-content {
    display: flex;
    flex-direction: row;
    gap: 1rem;
  }
  .bot-left-column {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 60vw;
  }
  .bot-right-column {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .bot-right-column img {
    max-height: 250px;
    max-width: 250px;
    width: auto;
  }
  .unique-bot-identifier {
    color: #868686;
  }
  #button-group {
    display: flex;
    gap: 1rem;
  }
  #button-group button {
    background: var(--background-alt);
    color: var(--foreground);
    cursor: pointer;
    font-size: 1rem;
    align-self: flex-start;
  }
  #toml-path {
    font-size: 0.8rem;
    color: grey;
  }
</style>
