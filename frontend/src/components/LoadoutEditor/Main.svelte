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
import Switch from "../Switch.svelte";

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

let previewMatchTeam: "blue" | "orange" | null = $state(null);
let previewOnChange = $state(localStorage.getItem("LOADOUT_PREVIEW_ON_CHANGE") === "true");
$effect(() => {
  localStorage.setItem("LOADOUT_PREVIEW_ON_CHANGE", previewOnChange.toString());
});

$effect(() => {
  if (!previewMatchTeam || !previewOnChange) {
    return;
  }

  PreviewLoadout(previewMatchTeam)
});

function PreviewLoadout(team: "blue" | "orange") {
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
    loadout: team === "blue" ? blueLoadout : orangeLoadout,
    team: team === "blue" ? 0 : 1,
    launcher,
    launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
  };

  if (!previewMatchTeam) {
    App.LaunchPreviewLoadout(
      options,
      ExistingMatchBehavior.ExistingMatchBehaviorRestart,
    )
      .then(() => {
        previewMatchTeam = team;
        toast.success(`Launching preview for ${previewMatchTeam} car`);
      })
      .catch((e) => {
        previewMatchTeam = null;
        toast.error(`Failed to launch preview: ${e}`);
      });
  } else {
    App.SetLoadout(options)
      .then(() => {
        previewMatchTeam = team;
        toast.success(
          `Preview updated for ${team} car`,
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
      <h3>Preview in game:</h3>
      <button id="preview-blue" onclick={() => PreviewLoadout("blue")}>
        <img src={EyeIcon} alt="eye">
        Blue car
      </button>
      <button id="preview-orange" onclick={() => PreviewLoadout("orange")}>
        <img src={EyeIcon} alt="eye">
        Orange car
      </button>
      {#if previewMatchTeam}
      <button id="preview-on-change" onclick={() => previewOnChange = !previewOnChange}>
        Auto-preview
        <Switch
          checked={previewOnChange}
          onchange={() => previewOnChange = !previewOnChange}
          stopClickPropagation={true}
          height={24}
          width={40}
        />
      </button>
      {/if}
    </div>
    <div class="right">
      <button type="submit" onclick={saveLoadout}>Save and close</button>
      <button type="reset" onclick={revertChanges}>Revert changes</button>
    </div>
  </div>
</Modal>

<style>
  .left, .right {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
  }
  #footer h3 {
    white-space: nowrap;
  }
  button {
    white-space: nowrap;
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
  button#preview-on-change {
    border: var(--foreground) 1px solid;
    gap: 10px;
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
    display: flex;
    flex-wrap: wrap;
    margin-top: 10px;
    gap: 10px;
  }
</style>
