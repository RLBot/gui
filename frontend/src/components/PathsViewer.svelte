<script lang="ts">
import toast from "svelte-5-french-toast";
import { App } from "../../bindings/gui";
import closeIcon from "../assets/close.svg";
import repairIcon from "../assets/repair.svg";
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

async function setDefaultPath() {
  const defaultPath = await App.GetDefaultPath();
  installPath = `${defaultPath}/RLBotPack`;
}

setDefaultPath();

function removePath(index: number) {
  paths.splice(index, 1);
}

function repairBotpack(index: number) {
  const id = toast.loading("Re-downloading botpack...");

  // @ts-ignore
  App.RepairBotpack(paths[index].repo, paths[index].installPath)
    .then((tagName) => {
      paths[index].tagName = tagName;
      toast.success("Botpack download successfully!", { id });
    })
    .catch((err) => {
      toast.error(`Failed to download botpack: ${err}`, {
        duration: 10000,
        id,
      });
    });
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

function addInstallPath(installPath: string) {
  if (paths.some((x) => x.installPath === installPath)) {
    toast.error("Path already added");
    return;
  }

  paths.push({
    installPath,
    visible: true,
    tagName: null,
    repo: null,
  });
}

async function addFolder() {
  const result = await App.PickFolder();

  if (!result) {
    return;
  }

  addInstallPath(result);
}

async function addFile() {
  const result = await App.PickTomlFile();

  if (!result) {
    return;
  }

  addInstallPath(result);
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
        {#if path.repo}
          <button class="no-left-margin repair" onclick={() => repairBotpack(i)}>
            <img src={repairIcon} alt="repair" />
          </button>
          <div><Switch bind:checked={path.visible} /></div>
        {:else}
          <div class="no-left-margin"><Switch bind:checked={path.visible} /></div>
        {/if}
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
  .path .no-left-margin {
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
  .repair {
    color: var(--foreground);
  }
</style>
