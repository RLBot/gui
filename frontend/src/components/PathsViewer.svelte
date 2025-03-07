<script lang="ts">
import toast from "svelte-5-french-toast";
import { App } from "../../bindings/gui";
import closeIcon from "../assets/close.svg";
import Modal from "./Modal.svelte";
import Switch from "./Switch.svelte";

const OFFICIAL_BOTPACK_REPO = "VirxEC/botpack-test";

let {
  visible = $bindable(false),
  paths = $bindable([]),
}: {
  visible?: boolean;
  paths?: {
    tagName: string | null;
    repo: string | null;
    installPath: string;
    visible: boolean;
  }[];
} = $props();

let addBotpackVisible = $state(false);
let selectedBotpackType = $state("official");
let customRepo = $state("");
let installPath = $state("");

function setDefaultPath() {
  App.GetDefaultPath().then((result) => {
    installPath = `${result}/RLBotPack`;
  });
}

setDefaultPath();

function removePath(index: number) {
  paths.splice(index, 1);
}

function openAddBotpackModal() {
  visible = false;
  addBotpackVisible = true;
}

function closeAddBotpackModal() {
  addBotpackVisible = false;
  visible = true;
  selectedBotpackType = "official";
  customRepo = "";
  setDefaultPath();
}

function addFolder() {
  App.PickFolder().then((result) => {
    if (result) {
      if (paths.some((x) => x.installPath === result)) {
        toast.error("Path already added");
        return;
      }

      paths.push({
        installPath: result,
        repo: null,
        tagName: null,
        visible: true,
      });
    }
  });
}

function addFile() {
  App.PickTomlFile().then((result) => {
    if (result) {
      if (paths.some((x) => x.installPath === result)) {
        toast.error("Path already added");
        return;
      }

      paths.push({
        installPath: result,
        repo: null,
        tagName: null,
        visible: true,
      });
    }
  });
}

function confirmAddBotpack() {
  if (!installPath) {
    toast.error("Install path cannot be blank");
    return;
  }

  let repo = OFFICIAL_BOTPACK_REPO;
  if (selectedBotpackType === "custom") {
    if (!customRepo) {
      toast.error("URL cannot be blank");
      return;
    }

    if (!/^[\w-]+\/[\w-]+$/.test(customRepo)) {
      toast.error("Custom repository must be in the format 'owner/repo'");
      return;
    }

    repo = customRepo;
  }

  if (paths.some((x) => x.repo === repo)) {
    toast.error("Botpack already added");
    return;
  }

  if (paths.some((x) => x.installPath === installPath)) {
    toast.error("Install path already in use");
    return;
  }

  const id = toast.loading("Downloading botpack...");
  App.DownloadBotpack(repo, installPath)
    .then((tagName) => {
      toast.success("Botpack downloaded successfully!", { id });

      paths.push({ installPath, repo, tagName, visible: true });
      closeAddBotpackModal();
    })
    .catch((err) => {
      toast.error(`Failed to download botpack: ${err}`, {
        duration: 10000,
        id,
      });
    });
}
</script>

<Modal title="Manage Paths" bind:visible={visible}>
  <div class="paths">
    <div class="button-row">
      <button onclick={addFolder}>Add folder</button>
      <button onclick={addFile}>Add File</button>
      <button onclick={openAddBotpackModal}>Add Botpack</button>
    </div>

    {#each paths as path, i}
      <div class="path">
        <pre>{path.repo ? `${path.repo} @ ${path.installPath}` : path.installPath}</pre>
        <div><Switch bind:checked={path.visible} /></div>
        <button class="close" onclick={() => removePath(i)}>
          <img src={closeIcon} alt="X" />
        </button>
      </div>
    {/each}
  </div>
</Modal>

<Modal title="Add Botpack" bind:visible={addBotpackVisible}>
  <div class="add-botpack">
    <label for="path">Botpack install path</label>
    <input type="text" id="path" placeholder="Enter install path" bind:value={installPath} />
    <div class="radio-group">
      <label>
        <input type="radio" name="botpackType" value="official" bind:group={selectedBotpackType} />
        Official RLBotPack
      </label>
      <label>
        <input type="radio" name="botpackType" value="custom" bind:group={selectedBotpackType} />
        Custom
      </label>
    </div>
    {#if selectedBotpackType === "custom"}
      <input type="text" placeholder="owner/repo" bind:value={customRepo} />
    {/if}
    <div class="button-row">
      <button onclick={confirmAddBotpack}>Confirm</button>
      <button onclick={closeAddBotpackModal}>Cancel</button>
    </div>
  </div>
</Modal>

<style>
  .paths {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 70vw;
    min-height: 50vh;
  }
  .button-row {
    display: flex;
    gap: 1rem;
  }
  .button-row {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
  }
  .button-row button {
    flex: 1;
  }
  .path {
    display: flex;
    align-items: center;
    gap: 1rem;
    justify-content: space-between;
  }
  .path div {
    margin-left: auto;
  }
  .path pre {
    font-size: 1rem;
    margin: 0px;
  }
  .path button {
    padding: 0px;
  }
  .path button img {
    filter: invert();
  }
  .add-botpack {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-width: 50vw;
  }
  .radio-group {
    flex-direction: column;
    display: flex;
    gap: 1rem;
  }
  .radio-group label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
</style>
