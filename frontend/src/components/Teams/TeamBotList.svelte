<script lang="ts">
import { dndzone } from "svelte-dnd-action";
import { flip } from "svelte/animate";
import type { DraggablePlayer } from "../..";
import closeIcon from "../../assets/close.svg";
import defaultIcon from "../../assets/rlbot_mono.png";

const flipDurationMs = 100;

let { items = $bindable() }: { items: DraggablePlayer[] } = $props();

function handleSort(e: any) {
  items = e.detail.items;
}

function remove(id: number): any {
  items = items.filter((x) => x.id !== id);
}
</script>

<div class="teamBotList">
    <p
        class="placeholder"
        style={items.length == 0
            ? "margin-top:.6rem;"
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
            dropTargetClasses: ["dropTarget"],
        }}
        onconsider={handleSort}
        onfinalize={handleSort}
    >
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        {#each items as bot (bot.id)}
            <!-- TODO: maybe remove stopPropagation and instead require a click on the team header -->
            <div class="bot" animate:flip={{ duration: items.length < 16 ? flipDurationMs : 0 }} onclick={e => e.stopPropagation()}>
                <img src={bot?.icon || defaultIcon} alt="icon" />
                <p>{bot?.displayName}</p>
                <div style="flex: 1;"></div>
                <button class="close" onclick={remove.bind(null, bot.id)}>
                    <img src={closeIcon} alt="X" />
                </button>
            </div>
        {/each}
    </div>
</div>

<style>
    * {
        user-select: none;
        -webkit-user-select: none;
    }
    .teamBotList {
        padding: 0.6rem;
        overflow: auto;
        height: 100%;
        min-height: 4rem;
        display: flex;
        flex-direction: column;
        position: relative;
    }
    .placeholder {
        position: absolute;
        transition: 100ms;
    }
    .bots {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        min-height: 100%;
        overflow-y: auto;
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
    .close {
        background-color: transparent;
        height: 100%;
        padding: 0;
    }
    .close img {
        height: 100%;
        width: auto;
        filter: invert();
    }
</style>
