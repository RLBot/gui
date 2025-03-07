<script lang="ts">
import Modal from "./Modal.svelte";
import NiceSelect from "./NiceSelect.svelte";

let { visible = $bindable() } = $props();

let localLauncher = $state(localStorage.getItem("MS_LAUNCHER") || "");
let localLauncherArg = $state(localStorage.getItem("MS_LAUNCHER_ARG") || "");

const launcherOptions = {
  Steam: "steam",
  Epic: "epic",
  Custom: "custom",
  Legendary: "legendary",
  "Don't launch": "nolaunch",
};

let launcher = $state(
  localStorage.getItem("MS_LAUNCHER") === "custom" &&
    localStorage.getItem("MS_LAUNCHER_ARG") === "legendary"
    ? "legendary"
    : localStorage.getItem("MS_LAUNCHER") || "",
);

$effect(() => {
  if (launcher === "legendary") {
    localLauncher = "custom";
    localLauncherArg = "legendary";
  } else {
    localLauncher = launcher;
    if (localLauncherArg === "legendary") {
      localLauncherArg = "";
    }
  }

  localStorage.setItem("MS_LAUNCHER", localLauncher);
  localStorage.setItem("MS_LAUNCHER_ARG", localLauncherArg);
});
</script>

<button onclick={() => { visible = true }}>Launcher Options</button>

<Modal title="Select a launcher" bind:visible>
  <div class="container">
    <NiceSelect bind:value={launcher} options={launcherOptions} placeholder="Select a launcher" />
    {#if launcher === "custom"}
    <div class="launcherArg">
      <label for="launcherArg">Additional argument:</label>
      <input type="text" id="launcherArg" bind:value={localLauncherArg} placeholder="(Leave blank for default)">
    </div>
    {/if}
  </div>
</Modal>

<style>
.container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.launcherArg {
  display: flex;
  flex-direction: column;
}
</style>
