<script lang="ts">
import Modal from "../Modal.svelte";
import type {DraggablePlayer} from "../../index";
import {BotInfo, PsyonixBotInfo} from "../../../bindings/gui";
import {Dialogs} from "@wailsio/runtime";

let {
    player = $bindable(undefined),
    visible = $bindable(false)
}: {
    player?: DraggablePlayer | null
    visible?: boolean
} = $props();

async function clearOverrides() {
    if (player) {
        player.overrides.name = player.player instanceof PsyonixBotInfo ? "" : player.displayName;
        player.overrides.loadout = null;
        player.overrides.auto_start = true;
    }
}

</script>

{#if player}
<Modal title={`Edit ${player.displayName}`} bind:visible >
    <p>In-game name</p>
    <input
            type="text"
            placeholder="Bot name"
            id={`edit-name-${player.id}`}
            bind:value={player.overrides.name}
    >
    <br />
    <br />
    {#if player.player instanceof BotInfo}
    <input
            type="checkbox"
            id={`edit-auto-start-${player.id}`}
            bind:checked={player.overrides.auto_start}
    >
    <label for={`edit-auto-start-${player.id}`}>Auto start</label>
    {/if}
    <!-- TODO: Add loadout file picker for non-humans -->
    <div class="buttons">
        <button onclick={clearOverrides}>Clear Overrides</button>
    </div>
</Modal>
{/if}

<style>
    .buttons {
        display: flex;
        width: 100%;
        justify-content: left;
        gap: .5rem;
        padding-top: 1rem;
    }
</style>