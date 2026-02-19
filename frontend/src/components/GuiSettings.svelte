<script lang="ts">
import { currentTheme, showWindowControls, THEMES } from "../settings";
import Modal from "./Modal.svelte";
import NiceSelect from "./NiceSelect.svelte";
import Switch from "./Switch.svelte";

// TODO: Save settings, svelte store + localstorage?
//       Perhaps change all localstorage state to svelte stores?

let { visible = $bindable(false) } = $props();
</script>

<Modal bind:visible title="GUI Settings">
  <div class="inner">

    <div>
      <label>Theme</label>
      <NiceSelect
        bind:value={$currentTheme}
        options={Object.fromEntries(
          Object.keys(THEMES).map(k => [k, k])
        )}
        placeholder="Select a theme"
      />
    </div>

    <div style="display:flex; align-items:center; gap:1rem;">
      <Switch bind:checked={$showWindowControls} />
      <p>Custom window controls</p>
    </div>
    <!-- TODO: Match start timeouts -->
    <!-- TODO: Refresh bots behavior (remove on refresh, remove not found agents, etc.) -->
    <!-- TODO: Telemetry settings if added -->
    <!-- TODO: Auto update botpack -->
  </div>
</Modal>

<style>
  .inner {
    display: flex;
    justify-content: center;
    flex-direction: column;
    height: 100%;
    gap: 1rem;
  }
  h3 {
    padding: 5rem;
  }
</style>
