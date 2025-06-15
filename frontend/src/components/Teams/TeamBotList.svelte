<script lang="ts">
import { untrack } from "svelte";
import { dndzone } from "svelte-dnd-action";
import { flip } from "svelte/animate";
import type { DraggablePlayer } from "../..";
import { BotInfo, type PsyonixBotInfo } from "../../../bindings/gui";
import closeIcon from "../../assets/close.svg";
import duplicateIcon from "../../assets/duplicate.svg";
import editIcon from "../../assets/edit.svg";
import defaultIcon from "../../assets/rlbot_mono.png";
import ModalPrompt from "../ModalPrompt.svelte";

const flipDurationMs = 100;

let { items = $bindable() }: { items: DraggablePlayer[] } = $props();

function handleSort(e: any) {
  items = e.detail.items;
}

function remove(id: string): any {
  items = items.filter((x) => x.id !== id);
}

function dupe(id: string): any {
  const index = items.findIndex((x) => x.id === id);
  if (index !== -1) {
    // Create a copy with a new ID
    const newItem = {
      ...items[index],
      id: crypto.randomUUID(), // Generate a new unique ID
    };
    // Insert the new item after the original
    items = [...items.slice(0, index + 1), newItem, ...items.slice(index + 1)];
  }
}

let editPrompts: { [key: string]: ModalPrompt } = {};
let editDataNames: { [key: string]: string } = $state({});

// Update editDataNames once items updates, but try to keep as much state as possible
// Once editing more stuff is supported, this approach should probably change
$effect(() => {
  let localEditDataNames = untrack(() => {
    return editDataNames;
  });
  let updated = Object.fromEntries(
    items
      .filter((x) => {
        return x.player instanceof BotInfo;
      })
      .map((bot) => {
        return [
          bot.id,
          // use ! + ?: instead of ?? so that "" is treated as null
          // this is due to the input binding value to "" before this is ran
          !localEditDataNames[bot.id]
            ? (<BotInfo>bot.player).config.settings.name
            : localEditDataNames[bot.id],
        ];
      }),
  );
  untrack(() => {
    editDataNames = updated;
  });
});

async function edit_custom_bot(id: string): Promise<void> {
  let modal = editPrompts[id];

  let modified = await modal.prompt();
  if (modified) {
    const index = items.findIndex((x) => x.id === id);
    let copy = {
      ...items[index],
      // We need to deepclone here to make sure we don't modify the player globally.
      // We also need to do BotInfo.createFrom to make sure that instanceof BotInfo == true still.
      player: BotInfo.createFrom(structuredClone(items[index].player)),
    };
    copy.modified = true;
    if (copy.player instanceof BotInfo) {
      copy.player.config.settings.name = editDataNames[id];
    }
    items[index] = copy;
  }
}
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
    use:dndzone={{
      items,
      flipDurationMs,
      dropTargetStyle: {},
      dropTargetClasses: ["dropTarget"]
    }}
    onconsider={handleSort}
    onfinalize={handleSort}
  >
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    {#each items as bot (bot.id)}
      <!-- TODO: maybe remove stopPropagation and instead require a click on the team header -->
      <div class="bot" animate:flip={{ duration: items.length < 16 ? flipDurationMs : 0 }} onclick={e => e.stopPropagation()}>
        <img src={bot.icon || defaultIcon} alt="icon" />
        <p
          style={bot.modified ? "color: orange" : ""}
          title={bot.modified ? "(modified)" : undefined}
        >{bot.displayName}</p>
        {#if bot.uniquePathSegment}
          <span class="unique-bot-identifier">({bot.uniquePathSegment})</span>
        {/if}
        <div style="flex: 1;"></div>
        {#if bot.player instanceof BotInfo}
          <button class="edit" title="edit" onclick={edit_custom_bot.bind(null, bot.id)}>
            <img src={editIcon} alt="Dupe">
          </button>
        {/if}
        <!-- TODO: support editing psyonix bots too, skill level (and name?) -->
        {#if !bot.tags.includes("human")}
          <button class="duplicate" title="Duplicate" onclick={dupe.bind(null, bot.id)}>
            <img src={duplicateIcon} alt="Dupe">
          </button>
        {/if}
        <button class="close" onclick={remove.bind(null, bot.id)}>
          <img src={closeIcon} alt="X" />
        </button>
      </div>
    {/each}
  </div>
</div>

{#each items as bot (bot.id)}
  <ModalPrompt title={"Edit " + bot.displayName} bind:this={editPrompts[bot.id]}>
    <div style="display: flex; flex-direction: column;">
      <label for={`edit-name-${bot.id}`}>Bot name (note: only in-game)</label>
      <input
        type="text"
        placeholder="Bot name"
        id={`edit-name-${bot.id}`}
        bind:value={editDataNames[bot.id]}
      >
      <!-- TODO: Add a bunch of more stuff to edit, loadout etc -->
      <!-- TODO: Perhaps add a way to save mods to bots as new tomls? -->
    </div>
  </ModalPrompt>
{/each}

<style>
  * {
    user-select: none;
    -webkit-user-select: none;
  }
  .teamBotList {
    padding: 0.6rem;
    overflow: auto;
    height: 100%;
    min-height: 8rem;
    max-height: 12rem;
    display: flex;
    flex-direction: column;
    position: relative;
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
    gap: 0.5rem;
    min-height: 100%;
    padding-bottom: 1rem;
  }
  .bot {
    display: flex;
    align-items: center;
    background-color: var(--background-alt);
    height: 2.25rem;
    padding: 0.2rem;
    gap: 0.5rem;
    border-radius: 0.2rem;
  }
  .bot img {
    height: 2rem;
    width: auto;
  }
  .bot:not(:hover) :is(.duplicate, .edit) {
    visibility: hidden;
  }
  .close, .duplicate, .edit {
    background-color: transparent;
    height: 100%;
    padding: 0;
  }
  .duplicate {
    padding: 0 .2rem;
  }
  :is(.close, .duplicate, .edit) img {
    height: 100%;
    width: auto;
    filter: invert();
  }
  .unique-bot-identifier {
    color: #868686;
  }
</style>
