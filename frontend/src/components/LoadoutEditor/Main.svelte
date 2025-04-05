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
import ArrowsIcon from "../../assets/arrows.svg";
import EyeIcon from "../../assets/eye.svg";
import Modal from "../Modal.svelte";
import Switch from "../Switch.svelte";
import TeamEditor from "./TeamEditor.svelte";
import type { CsvItem } from "./items";

let {
  visible = $bindable(false),
  loadout = $bindable(),
  basePath,
  loadoutFile,
  items,
  name,
  map,
}: {
  visible: boolean;
  loadout: LoadoutConfig;
  basePath: string;
  loadoutFile: string;
  items: {
    [x: string]: CsvItem[];
  };
  name: string;
  map: string;
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

const showcaseTypes = [
  { id: "static", name: "Static" },
  { id: "throttle", name: "Drive around center" },
  { id: "boost", name: "Boost around center" },
  { id: "back-center-kickoff", name: "Back center kickoff" },
  { id: "goal-explosion", name: "Goal explosion" },
];

let lastShowcaseType: string | null = null;
let selectedShowcaseType: string = $state(
  localStorage.getItem("LOADOUT_SHOWCASE_TYPE") || "static",
);
$effect(() => {
  localStorage.setItem("LOADOUT_SHOWCASE_TYPE", selectedShowcaseType);
});

const autoPreviewSetCooldownMS = 100;

let lastPreviewSetTime = 0;
let previewMatchTeam: "blue" | "orange" | null = $state(null);
let previewOnChange = $state(
  localStorage.getItem("LOADOUT_PREVIEW_ON_CHANGE") === "true",
);
$effect(() => {
  localStorage.setItem("LOADOUT_PREVIEW_ON_CHANGE", previewOnChange.toString());
});

function onLoadoutChange(team: "blue" | "orange") {
  if (previewMatchTeam !== team || !previewOnChange) {
    return;
  }

  if (Date.now() - lastPreviewSetTime < autoPreviewSetCooldownMS) {
    return;
  }

  PreviewLoadout(previewMatchTeam);
}

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
    map,
    loadout: team === "blue" ? blueLoadout : orangeLoadout,
    team: team === "blue" ? 0 : 1,
    launcher,
    launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
  };

  lastPreviewSetTime = Date.now();

  LaunchMatch(options, team).catch((e) => {
    previewMatchTeam = null;
    toast.error(`Preview failed: ${e}`);
  });
}

async function LaunchMatch(
  options: LoadoutPreviewOptions,
  team: "blue" | "orange",
) {
  if (!previewMatchTeam) {
    await App.LaunchPreviewLoadout(
      options,
      ExistingMatchBehavior.ExistingMatchBehaviorRestart,
    );

    toast.success(`Launching preview for ${team} car`);
  } else {
    if (
      lastShowcaseType !== selectedShowcaseType ||
      previewMatchTeam !== team
    ) {
      await App.LaunchPreviewLoadout(
        options,
        ExistingMatchBehavior.ExistingMatchBehaviorContinueAndSpawn,
      );
    } else {
      await App.SetLoadout(options);
    }

    toast.success(`Preview updated for ${team} car`);
  }

  previewMatchTeam = team;
  lastShowcaseType = selectedShowcaseType;
  App.SetShowcaseType(selectedShowcaseType, team === "blue" ? 0 : 1);
}
</script>

<Modal title={`Loadout of ${name}`} bind:visible>
  <div id="team-editors">
    <TeamEditor
      items={items}
      team="blue"
      bind:loadout={blueLoadout}
      onchange={() => onLoadoutChange("blue")}
    />

    <TeamEditor
      items={items}
      team="orange"
      bind:loadout={orangeLoadout}
      onchange={() => onLoadoutChange("orange")}
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
      <select
        bind:value={selectedShowcaseType}
        style="background-image: url({ArrowsIcon})"
      >
        {#each showcaseTypes as showcaseType}
          <option value={showcaseType.id}>{showcaseType.name}</option>
        {/each}
      </select>
    </div>
    <div class="right">
      <button type="submit" onclick={saveLoadout}>Save and close</button>
      <button type="reset" onclick={revertChanges}>Revert changes</button>
    </div>
  </div>
</Modal>

<style>
  select {
    display: inline-block;
    padding: .375rem 1.75rem .375rem .75rem;
    font-size: 1rem;
    font-weight: 400;
    line-height: 1.5;
    color: #495057;
    vertical-align: middle;
    border: 1px solid #ced4da;
    appearance: none;
    background-repeat: no-repeat;
    background-position: right .75rem center;
    background-size: 8px 10px;
  }
  .left, .right {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
    justify-content: center;
    margin: 0 auto;
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
  #team-editors, #footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
  }
  #team-editors {
    gap: 30px;
    justify-content: center;
  }
  #footer {
    justify-content: space-between;
    margin-top: 10px;
    gap: 10px;
    width: 100%;
  }
</style>
