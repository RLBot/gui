<script lang="ts">
import { type DragDropState, draggable, droppable } from "@thisux/sveltednd";
import SuperJSON from "superjson";
import { untrack } from "svelte";
import { flip } from "svelte/animate";
import { fade } from "svelte/transition";
import type { DraggablePlayer } from "../..";
import {BotInfo, PsyonixBotInfo} from "../../../bindings/gui";
import closeIcon from "../../assets/close.svg";
import duplicateIcon from "../../assets/duplicate.svg";
import editIcon from "../../assets/edit.svg";
import cannotAutoRunIcon from "../../assets/cannot_play.svg";
import defaultIcon from "../../assets/rlbot_mono.png";
import PlayerOverridesModal from "./PlayerOverridesModal.svelte";

let {
  items = $bindable(),
  globalAutoStart = $bindable(true)
}: {
  items: DraggablePlayer[],
  globalAutoStart: boolean,
} = $props();

function remove(id: string): any {
  items = items.filter((x) => x.id !== id);
}

function dupe(id: string): any {
  const index = items.findIndex((x) => x.id === id);
  if (index !== -1) {
    // Create a copy with a new ID
    const newItem = {
      ...items[index],
      overrides: {...items[index].overrides},
      id: crypto.randomUUID() // Generate a new unique ID
    };
    // Insert the new item after the original
    items = [...items.slice(0, index + 1), newItem, ...items.slice(index + 1)];
  }
}

let editedPlayer: DraggablePlayer | undefined = $state(undefined)
let showEditPrompt = $state(false)

async function showEditModal(d: DraggablePlayer) {
  if (d) {
    editedPlayer = d
    showEditPrompt = true;
  }
}

function reasonsForManualStart(d: DraggablePlayer): string[] {
  let reasons: string[] = []
  if (d.player instanceof BotInfo) {
    if (!globalAutoStart) reasons.push("Auto-start disabled in Extra");
    if (!d.player.config.settings.runCommand) reasons.push("No run_command declared");
    if (!d.overrides.autoStart) reasons.push("Auto-start disabled for bot");
  }
  return reasons
}

function canAutoStart(d: DraggablePlayer): boolean {
  return reasonsForManualStart(d).length == 0;
}

function hasOverrides(d: DraggablePlayer): boolean {
  let expectedName = (d.player instanceof BotInfo) ? d.displayName : "";
  return d.overrides.name !== expectedName || !d.overrides.autoStart || d.overrides.loadout != null
}

// :fire: :fire: :fire:
let resolveDeleted = () => {};
let _waitingForDelete: Promise<void>;
function resetWaitingForDelete() {
  _waitingForDelete = new Promise<void>((r, _) => {
    resolveDeleted = r;
  });
}
resetWaitingForDelete();
async function waitingForDelete() {
  await _waitingForDelete;
  resetWaitingForDelete();
}

function checkSameTeam(c1: string, c2: string) {
  return (
    c1.split("_").slice(0, 2).join("") === c2.split("_").slice(0, 2).join("")
  );
}

function onDragEnd(state: DragDropState) {
  const { targetContainer, sourceContainer } = state;
  if (!targetContainer) return;
  let itemIndex =
    // @ts-ignore
    +sourceContainer.split("_").at(-1);
  items.splice(itemIndex, 1);
  items = [...items];
  if (checkSameTeam(sourceContainer, targetContainer)) resolveDeleted();
}

async function onDrop(state: DragDropState<string>) {
  const { targetContainer, sourceContainer } = state;
  if (!targetContainer) return;
  let dropIndex =
    // @ts-ignore
    +targetContainer.split("_").at(-1);

  let val: DraggablePlayer = SuperJSON.parse(state.draggedItem);
  val.id = crypto.randomUUID();

  if (checkSameTeam(sourceContainer, targetContainer)) {
    // we need to make sure that we delete the item we picked up, before we insert it again.
    // this only applies if the source and target container belong to the same teambotlist
    await waitingForDelete();
  }
  items.splice(dropIndex, 0, val);
  items = [...items];
}

const dnd_container_namespace = `team_${crypto.randomUUID()}`;
</script>

<div class="teamBotList">
  <p
      class="placeholder"
      style={items.length == 0
          ? ""
          : "opacity: 0;z-index: -999"}
  >
      Drop bots here...
  </p>
  <div
    class="bots"
  >
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    {#each items as bot, i (bot.id)}
      <div
        class="botContainer"
        use:droppable={{
          container: `${dnd_container_namespace}_${i}`,
          callbacks: { onDrop },
        }}
        animate:flip={{ duration: 100 }}
        in:fade={{ duration: 100 }}
        out:fade={{ duration: 100 }}
      >
        <div
          class="bot"
          use:draggable={{
            container: `${dnd_container_namespace}_${i}`,
            dragData: SuperJSON.stringify(bot),
            interactive: ["button"],
            callbacks: { onDragEnd }
          }}
          onclick={e => e.stopPropagation()}
        >
          <img src={bot.icon || defaultIcon} alt="icon" />
          <p>{bot.displayName === bot.overrides.name || (bot.player instanceof PsyonixBotInfo && bot.overrides.name === "") ? bot.displayName : `${bot.overrides.name}`}</p>
          {#if bot.uniquePathSegment}
            <span class="unique-bot-identifier">({bot.uniquePathSegment})</span>
          {/if}
          {#if !canAutoStart(bot)}
            <img class="bot-icon" src={cannotAutoRunIcon} alt="No auto-run" title={`Must be launched manually (${reasonsForManualStart(bot).join("; ")})`}>
          {/if}
          <div style="flex: 1;"></div>
          {#if !bot.tags.includes("human")}
            <button class="duplicate" title="Duplicate" onclick={dupe.bind(null, bot.id)}>
              <img src={duplicateIcon} alt="Dupe">
            </button>
          {/if}
          {#if bot.player instanceof BotInfo || bot.player instanceof PsyonixBotInfo}
            <button class={hasOverrides(bot) ? "edit has-overrides" : "edit"} title="Edit" onclick={() => showEditModal(bot)}>
              <img src={editIcon} alt="Edit">
            </button>
          {/if}
          <button class="close" onclick={remove.bind(null, bot.id)}>
            <img src={closeIcon} alt="X" />
          </button>
        </div>
      </div>
    {/each}
    <div class="emptyDroppable" use:droppable={{
      container: `${dnd_container_namespace}_${items.length}`,
      callbacks: { onDrop }
    }}></div>
  </div>
</div>

<PlayerOverridesModal bind:player={editedPlayer} bind:visible={showEditPrompt} />

<style>
  * {
    user-select: none;
    -webkit-user-select: none;
  }
  .teamBotList {
    overflow: auto;
    height: 100%;
    min-height: 8rem;
    max-height: 12rem;
    display: flex;
    flex-direction: column;
    position: relative;
  }
  :global(.dragging) {
    opacity: 0.75;
  }
  .placeholder {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 1.1rem;
    font-style: italic;
    opacity: 0.8;
    max-width: 90%;
    width: max-content;
    transition: 100ms;
  }
  .bots {
    display: flex;
    flex-direction: column;
    min-height: 100%;
    overflow-y: auto;
  }
  .emptyDroppable {
    flex: 1;
    width: 100%;
    min-height: 1rem;
    z-index: 2;
  }
  .botContainer {
    padding: 0.6rem;
    padding-bottom: 0;
  }
  .bot {
    display: flex;
    align-items: center;
    background-color: var(--background-alt);
    height: 2.25rem;
    padding: 0.2rem;
    gap: 0.5rem;
    border-radius: 0.2rem;
    cursor: grab;
  }
  .bot img {
    height: 2rem;
    width: auto;
    user-drag: none;
    -webkit-user-drag: none;
  }
  .bot:not(:hover) :is(.duplicate, .edit) {
    visibility: hidden;
  }
  .has-overrides img {
    visibility: visible;
  }
  .close, .duplicate, .edit {
    background-color: transparent;
    height: 100%;
    padding: 0;
  }
  .duplicate {
    padding: 0 .1rem;
  }
  :is(.close, .duplicate, .edit) img {
    height: 100%;
    width: auto;
    filter: invert();
  }
  :is(.has-overrides) img {
    filter: invert(68%) sepia(66%) saturate(2755%) hue-rotate(360deg) brightness(103%) contrast(104%);
  }
  .bot-icon {
    background-color: transparent;
    padding: 0 .1rem;
    margin: .4rem 0;
    height: 100%;
    width: auto;
    filter: invert(60%);
  }
  .unique-bot-identifier {
    color: #868686;
  }
</style>
