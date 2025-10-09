<script lang="ts">
import toast from "svelte-5-french-toast";
import { App } from "../../bindings/gui";
import closeIcon from "../assets/close.svg";
import repairIcon from "../assets/repair.svg";
import Modal from "./Modal.svelte";
import Switch from "./Switch.svelte";
import ProgressBar from "./ProgressBar.svelte";
import AddBotpack from "./AddBotpack.svelte";

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
    isDependency: boolean;
  }[];
} = $props();

let addBotpackVisible = $state(false);

function removePath(index: number) {
  const [removedItem] = paths.splice(index, 1);

  const dependency = paths.findIndex(
    (item) => item.installPath === removedItem.installPath,
  );
  if (dependency !== -1) removePath(dependency);
}

let downloadModalTitle = $state("Downloading & extracting repo/owner");
let downloadModalVisible = $state(false);
let downloadProgress = $state(0);
let downloadCurrentStep = $state(0);
let downloadTotalSteps = $state(0);

async function repairBotpack(index: number) {
  const info = paths[index];
  if (!info.repo) {
    return; // can't update something that doesn't have a repo
  }

  downloadProgress = 0;
  downloadCurrentStep = 0;
  downloadTotalSteps = 0;

  for (const item of paths) {
    if (!item.repo || item.installPath !== info.installPath) continue;
    downloadTotalSteps += 2;
  }

  let clearInstallPath = true;
  for (const item of paths) {
    if (!item.repo || item.installPath !== info.installPath) continue;

    downloadModalTitle = `Redownloading & extracting ${item.repo}`;
    visible = false;
    downloadModalVisible = true;

    let tagName = await App.RepairBotpack(
      item.repo,
      item.installPath,
      clearInstallPath,
    ).catch((err) => {
      toast.error(`Failed to download botpack: ${err}`, {
        duration: 10000,
      });

      return null;
    });
    if (!tagName) break;

    clearInstallPath = false;
    downloadProgress = 0;
    downloadCurrentStep += 1;
  }

  downloadModalVisible = false;
  visible = true;
}

function openAddBotpackModal() {
  visible = false;
  addBotpackVisible = true;
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
    isDependency: false,
  });
}

async function addFolder() {
  const result = await App.PickFolder();

  if (!result) {
    return;
  }

  addInstallPath(result);
}

function addFile() {
  App.PickRLBotToml()
    .then((result) => {
      if (!result) {
        return;
      }

      addInstallPath(result);
    })
    .catch((_) => {
      toast.error(
        "Failed to add file: Only bot.toml or script.toml files are allowed",
      );
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
      {#if path.isDependency}
        <div class="path" style="margin-left: 50px;">
          <pre>{path.repo}</pre>
        </div>
      {:else}
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
      {/if}
    {/each}
  </div>
</Modal>

<AddBotpack bind:parentVisible={visible} bind:visible={addBotpackVisible} bind:paths />

<Modal title={downloadModalTitle} bind:visible={downloadModalVisible} closeable={false}>
  <ProgressBar bind:percentComplete={downloadProgress} bind:currentStep={downloadCurrentStep} totalSteps={downloadTotalSteps} />
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
  .repair {
    color: var(--foreground);
  }
</style>
