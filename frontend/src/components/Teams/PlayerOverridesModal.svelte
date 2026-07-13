<script lang="ts">
import { App, BotInfo, PsyonixBotInfo } from "../../../bindings/gui";
import type { DraggablePlayer } from "../../index";
import Modal from "../Modal.svelte";

let {
  player = $bindable(undefined),
  visible = $bindable(false),
}: {
  player?: DraggablePlayer | null;
  visible?: boolean;
} = $props();

function clearOverrides() {
  if (player) {
    player.overrides.name = null;
    player.overrides.loadout = null;
    player.overrides.autoStart = true;
  }
}

function hasNameOverride(): boolean {
  if (!player) return false;
  return player.overrides.name !== null;
}

async function pickLoadoutOverride() {
  if (!player) return;
  let path = await App.PickToml();
  if (!path) return;
  player.overrides.loadout = await App.GetLoadout(path);
}
</script>

{#if player}
<Modal title={`Edit ${player.displayName}`} bind:visible >
    <p style={hasNameOverride() ? "color: orange;" : ""}>In-game name:</p>
    <input
            type="text"
            placeholder=""
            id={`edit-name-${player.id}`}
            bind:value={
                () => player?.overrides.name ?? player.displayName,
                (v) => player.overrides.name = (v === player.displayName) ? null : v
            }
    >
    <br />
    <br />
    {#if player.info instanceof BotInfo || player.info instanceof PsyonixBotInfo}
        <p style={player.overrides.loadout ? "color: orange;" : ""}>Loadout: {player.overrides.loadout ? "Customized" : ""}</p>
        <button onclick={pickLoadoutOverride}>Pick TOML</button>
        <br />
        <br />
    {/if}
    {#if player.info instanceof BotInfo}
        <input
                type="checkbox"
                id={`edit-auto-start-${player.id}`}
                bind:checked={player.overrides.autoStart}
        >
        <label for={`edit-auto-start-${player.id}`} style={player.overrides.autoStart ? "" : "color: orange;"}>Auto start</label>
    {/if}
    <hr>
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