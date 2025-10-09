<script lang="ts">
import toast from "svelte-5-french-toast";
import { App } from "../../bindings/gui";
import Modal from "./Modal.svelte";
import ProgressBar from "./ProgressBar.svelte";

const OFFICIAL_BOTPACK_REPOS = [
  "VirxEC/botpack-test",
  "VirxEC/pytorch-archive",
];

let {
  parentVisible = $bindable(false),
  visible = $bindable(false),
  paths = $bindable([]),
}: {
  parentVisible?: boolean;
  visible?: boolean;
  paths?: {
    tagName: string | null;
    repo: string | null;
    installPath: string;
    visible: boolean;
    isDependency: boolean;
  }[];
} = $props();

let selectedBotpackType = $state("official");
let customRepo = $state("");
let installPath = $state("");

async function setDefaultPath() {
  const defaultPath = await App.GetDefaultPath();
  installPath = `${defaultPath}/RLBotPack`;
}

setDefaultPath();

let downloadModalTitle = $state("Downloading & extracting repo/owner");
let downloadModalVisible = $state(false);
let downloadProgress = $state(0);
let downloadCurrentStep = $state(0);
let downloadTotalSteps = $state(0);

function closeAddBotpackModal() {
  downloadModalVisible = false;
  visible = false;
  parentVisible = true;
  selectedBotpackType = "official";
  customRepo = "";
  setDefaultPath();
}

async function confirmAddBotpack() {
  if (!installPath) {
    toast.error("Install path cannot be blank");
    return;
  }

  let repo: string, dep: string | null;

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
    dep = null;
  } else {
    [repo, dep] = OFFICIAL_BOTPACK_REPOS;
  }

  if (paths.some((x) => x.installPath === installPath)) {
    toast.error(`Install path "${installPath}" already in use for ${repo}`);
    return;
  }

  if (paths.some((x) => x.repo === repo)) {
    toast.error(`Botpack ${repo} already added`);
    return;
  }

  downloadModalTitle = `Downloading & extracting ${repo}`;
  downloadProgress = 0;
  visible = false;
  downloadModalVisible = true;
  downloadCurrentStep = 0;
  downloadTotalSteps = dep ? 4 : 2;

  const tagName = await App.DownloadBotpack(repo, installPath).catch((err) => {
    toast.error(`Failed to download botpack: ${err}`, {
      duration: 10000,
    });

    return null;
  });
  if (!tagName) {
    downloadModalVisible = false;
    visible = true;
    return;
  }

  if (dep) {
    downloadModalTitle = `Downloading & extracting ${dep}`;
    downloadProgress = 0;
    downloadCurrentStep += 1;

    // Download possible dependency of the given botpack
    const tagName = await App.DownloadBotpack(dep, installPath).catch((err) => {
      toast.error(`Failed to download botpack dependency: ${err}`, {
        duration: 10000,
      });

      return null;
    });
    if (!tagName) {
      downloadModalVisible = false;
      visible = true;
      return;
    }
  }

  paths.push({
    installPath,
    repo,
    tagName,
    visible: true,
    isDependency: false,
  });
  if (dep) {
    paths.push({
      installPath,
      repo: dep,
      tagName,
      visible: false,
      isDependency: true,
    });
  }

  toast.success("Botpack downloaded successfully!");
  closeAddBotpackModal();
}
</script>

<Modal title="Add Botpack" bind:visible={visible}>
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

<Modal title={downloadModalTitle} bind:visible={downloadModalVisible} closeable={false}>
  <ProgressBar bind:percentComplete={downloadProgress} bind:currentStep={downloadCurrentStep} totalSteps={downloadTotalSteps} />
</Modal>

<style>
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
