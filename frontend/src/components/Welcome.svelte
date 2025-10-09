<script lang="ts">
import AddBotpack from "./AddBotpack.svelte";
import LauncherSelector from "./LauncherSelector.svelte";
import Modal from "./Modal.svelte";

let {
  paths = $bindable([]),
}: {
  paths?: {
    tagName: string | null;
    repo: string | null;
    installPath: string;
    visible: boolean;
    isDependency: boolean;
  }[];
} = $props();

let visible = $state(!localStorage.getItem("SHOW_WELCOME"));
$effect(() => {
  if (!visible) localStorage.setItem("SHOW_WELCOME", "false");
});

let addBotpackVisible = $state(false);
function openAddBotpackModal() {
  visible = false;
  addBotpackVisible = true;
}
</script>

<Modal bind:visible title="Setup RLBot">
  <div id="options">
    <p>Select your platform</p>
    <LauncherSelector />
    <br />
    <button onclick={openAddBotpackModal}>Download the botpack</button>
  </div>
</Modal>

<AddBotpack bind:parentVisible={visible} bind:visible={addBotpackVisible} bind:paths />

<style>
  #options > * {
    margin-bottom: 0.5rem;
  }
</style>
