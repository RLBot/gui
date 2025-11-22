<script lang="ts">
import Modal from "../Modal.svelte";
import type {DraggablePlayer} from "../../index";
import {BotInfo} from "../../../bindings/gui";

let {
    player = $bindable(undefined),
    visible = $bindable(false)
}: {
    player?: DraggablePlayer | null
    visible?: boolean
} = $props();

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
</Modal>
{/if}