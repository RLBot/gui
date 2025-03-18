<script lang="ts">
import toast from "svelte-5-french-toast";
import { ExistingMatchBehavior } from "../../../bindings/github.com/swz-git/go-interface/flat/models";
import {
  App,
  LoadoutConfig,
  LoadoutPreviewOptions,
  TeamLoadoutConfig,
} from "../../../bindings/gui";
import { MAPS_STANDARD } from "../../arena-names";
import EyeIcon from "../../assets/eye.svg";
import Modal from "../Modal.svelte";
import TeamEditor from "./TeamEditor.svelte";
import type { CsvItem } from "./items";

let {
  visible = $bindable(false),
  loadout = $bindable(),
  basePath,
  loadoutFile,
  items,
  name,
}: {
  visible: boolean;
  loadout: LoadoutConfig;
  basePath: string;
  loadoutFile: string;
  items: {
    [x: string]: CsvItem[];
  };
  name: string;
} = $props();

let blueLoadout: TeamLoadoutConfig = $state(
  structuredClone(loadout.blueLoadout),
);
let orangeLoadout: TeamLoadoutConfig = $state(
  structuredClone(loadout.orangeLoadout),
);
$effect(() => {
  blueLoadout = structuredClone(loadout.blueLoadout);
  orangeLoadout = structuredClone(loadout.orangeLoadout);
});

function revertChanges() {
  blueLoadout = structuredClone(loadout.blueLoadout);
  orangeLoadout = structuredClone(loadout.orangeLoadout);
}

function saveLoadout() {
  // structuredClone doesn't work here, likely because of $state
  loadout.blueLoadout = JSON.parse(JSON.stringify(blueLoadout));
  loadout.orangeLoadout = JSON.parse(JSON.stringify(orangeLoadout));

  App.SaveLoadoutToFile(basePath, loadoutFile, loadout)
    .then(() => {
      visible = false;
      toast.success(`Saved the loadout of ${name}`);
    })
    .catch((e) => {
      toast.error(`Failed to save the loadout of ${name}: ${e}`);
    });
}

let previewMatchLaunched = false;

function PreviewLoadout(isBlueTeam: boolean) {
  const launcher = localStorage.getItem("MS_LAUNCHER");
  if (!launcher) {
    toast.error("Please select a launcher first", {
      position: "bottom-right",
      duration: 5000,
    });

    return;
  }

  const options: LoadoutPreviewOptions = {
    map: MAPS_STANDARD["DFH Stadium"],
    loadout: isBlueTeam ? blueLoadout : orangeLoadout,
    team: isBlueTeam ? 0 : 1,
    launcher,
    launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
  };

  if (!previewMatchLaunched) {
    App.LaunchPreviewLoadout(
      options,
      ExistingMatchBehavior.ExistingMatchBehaviorRestart,
    )
      .then(() => {
        previewMatchLaunched = true;
        toast.success(
          `Launching preview for ${isBlueTeam ? "blue" : "orange"} car`,
        );
      })
      .catch((e) => {
        toast.error(`Failed to launch preview: ${e}`);
      });
  } else {
    App.SetLoadout(options)
      .then(() => {
        toast.success(
          `Preview updated for ${isBlueTeam ? "blue" : "orange"} car`,
        );
      })
      .catch((e) => {
        toast.error(`Failed to update preview: ${e}`);
      });
  }
}
</script>

<Modal title={`Loadout of ${name}`} bind:visible>
  <div id="body">
    <TeamEditor
      items={items}
      team="blue"
      bind:loadout={blueLoadout}
    />

    <TeamEditor
      items={items}
      team="orange"
      bind:loadout={orangeLoadout}
    />
  </div>
  <div id="footer">
    <div class="left">
      <button id="preview-blue" onclick={() => PreviewLoadout(true)}>
        <img src={EyeIcon} alt="eye">
        Preview blue car in game
      </button>
      <button id="preview-orange" onclick={() => PreviewLoadout(false)}>
        <img src={EyeIcon} alt="eye">
        Preview orange car in game
      </button>
    </div>
    <div class="right">
      <button type="submit" onclick={saveLoadout}>Save and close</button>
      <button type="reset" onclick={revertChanges}>Revert changes</button>
    </div>
  </div>
</Modal>

<style>
  .left, .right {
    gap: 10px;
  }
  button {
    display: inline-flex;
    align-items: center;
  }
  button > img {
    width: 24px;
    height: 24px;
    margin-right: 5px;
  }
  button#preview-blue {
    color: #2196F3;
    border: #2196F3 1px solid;
  }
  button#preview-orange {
    color: #FF9800;
    border: #FF9800 1px solid;
  }
  /* filters calculated with https://codepen.io/sosuke/pen/Pjoqqp */
  button#preview-blue > img {
    filter: invert(52%) sepia(37%) saturate(5199%) hue-rotate(185deg) brightness(99%) contrast(93%);
  }
  button#preview-orange > img {
    filter: invert(79%) sepia(58%) saturate(5589%) hue-rotate(0deg) brightness(103%) contrast(104%);
  }
  #body, #footer {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }
  #body {
    gap: 30px;
    flex-wrap: wrap;
    overflow: hidden;
    align-items: center;
    justify-content: center;
  }
  #footer {
    margin-top: 10px;
  }
</style>
