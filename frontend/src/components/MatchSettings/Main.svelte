<script lang="ts">
import { MAPS_NON_STANDARD, MAPS_STANDARD } from "../../arena-names";
import LauncherSelector from "../LauncherSelector.svelte";
import Modal from "../Modal.svelte";
import NiceSelect from "../NiceSelect.svelte";
import Select from "../NiceSelect.svelte";
import { type Gamemode, gamemodes } from "./rlmodes";
import { mutators as mutatorOptions } from "./rlmutators";

let {
  map = $bindable(),
  mode = $bindable(),
  extraOptions = $bindable(),
  mutators = $bindable(),
  launcherOptionsVisible = $bindable(),
  onStart = (randomizeMap: boolean) => {},
  onStop = () => {},
} = $props();
let showExtraOptions = $state(false);
let showMutators = $state(false);
let randomizeMap = $state(localStorage.getItem("MS_RANDOMIZE_MAP") === "true");
$effect(() => {
  localStorage.setItem("MS_RANDOMIZE_MAP", randomizeMap.toString());
});

const existingMatchBehaviors: { [n: string]: number } = {
  Restart: 0,
  "Continue and spawn": 1,
  "Restart if different": 2,
};

function cleanCase(toClean: string) {
  const halfClean = toClean.replaceAll("_", " ");
  return halfClean.charAt(0).toUpperCase() + halfClean.slice(1);
}

function resetMutators() {
  for (const key of Object.keys(mutators)) {
    mutators[key] = 0;
  }
  selectedPreset = "";
}

// the reason for default being "" and not null is that NiceSelect considers that the default
let selectedPreset: Gamemode | "" = $state("");
$effect(() => {
  if (selectedPreset !== "") {
    setPreset(selectedPreset);
  }
});

function setPreset(presetData: Gamemode) {
  if (presetData.match.game_mode !== undefined) {
    mode = presetData.match.game_mode;
  }

  if (presetData.match.game_map_upk !== undefined) {
    map = presetData.match.game_map_upk;
    randomizeMap = false;
  } else {
    randomizeMap = true;
  }

  for (const key of filteredMutatorOptions) {
    if (presetData.mutators[key] !== undefined) {
      mutators[key] = mutatorOptions[key].indexOf(presetData.mutators[key]);
    } else {
      mutators[key] = 0;
    }
  }
}

const filteredMutatorOptions = filterMutatorOptions();
function filterMutatorOptions() {
  let filtered = Object.keys(mutatorOptions).filter(
    (key) => key !== "game_mode",
  );
  filtered.sort();

  return filtered;
}

function getMaps() {
  const standardMaps = Object.entries(MAPS_STANDARD).sort(([a], [b]) =>
    a.localeCompare(b),
  );
  const nonStandardMaps = Object.entries(MAPS_NON_STANDARD).sort(([a], [b]) =>
    a.localeCompare(b),
  );

  return Object.fromEntries([...standardMaps, ...nonStandardMaps]);
}

const ALL_MAPS = getMaps();
</script>

<div class="matchSettings">
  <h1>Match Settings</h1>
  <div class="content">
    <div class="settings">
      <div class="left-controls">
        <Select
          options={ALL_MAPS}
          bind:value={map}
          placeholder="Select map"
        />
        <Select
          options={Object.fromEntries(mutatorOptions.game_mode.map((x) => [x, x]))}
          bind:value={mode}
          placeholder="Select mode"
        />
      </div>
      <div class="right-controls">
        <LauncherSelector bind:visible={launcherOptionsVisible} />
      </div>
    </div>
    <div class="controls">
      <div class="left-controls">
        <button
          onclick={() => {
            showMutators = true;
          }}>Mutators</button
        >
        <button
          onclick={() => {
            showExtraOptions = true;
          }}>Extra</button
        >
        <input
          type="checkbox"
          id="randomizeMap"
          bind:checked={randomizeMap}
        />
        <label for="randomizeMap">Randomize Map</label>
      </div>
      <div class="right-controls">
          <button class="start" onclick={()=>{onStart(randomizeMap)}}>Start Match</button>
          <button class="stop" onclick={()=>{onStop()}}>Stop</button>
      </div>
    </div>
  </div>
</div>

<Modal title="Rocket League Mutators" bind:visible={showMutators}>
  <div class="mutators">
    {#each filteredMutatorOptions as mutatorKey}
      <div class="mutator">
        <label
          style={mutators[mutatorKey] == 0 ? "color:lightgrey" : ""}
          for={mutatorKey}>{cleanCase(mutatorKey)}</label
        >

        <select
          name={mutatorKey}
          id={mutatorKey}
          bind:value={mutators[mutatorKey]}
          onchange={() => {selectedPreset = ""}}
        >
          {#each mutatorOptions[mutatorKey] as value, i}
              <option value={i}>{value}</option>
          {/each}
        </select>
      </div>
    {/each}
  </div>
  <div class="bottomButtons">
    <p>Settings are saved automatically</p>
    <NiceSelect bind:value={selectedPreset} options={gamemodes} placeholder="Select a preset" />
    <button
      class="mutatorResetButton"
      onclick={resetMutators}>Reset</button
    >
  </div>
</Modal>

<Modal title="RLBot Extra Options" bind:visible={showExtraOptions}>
  <div class="extraoptions">
    <p>Existing match behaviour</p>
    <NiceSelect bind:value={extraOptions.existingMatchBehavior} options={existingMatchBehaviors} placeholder="Existing Match Behavior" />
    <br />
    <br />
    <input
      type="checkbox"
      id="enableRendering"
      bind:checked={extraOptions.enableRendering}
    />
    <label for="enableRendering">
      Enable Rendering (bots can draw on screen)
    </label>
    <br />
    <input
      type="checkbox"
      id="enableStateSetting"
      bind:checked={extraOptions.enableStateSetting}
    />
    <label for="enableStateSetting">
      Enable State Setting (bots can teleport)
    </label>
    <br />
    <input
      type="checkbox"
      id="autoSaveReplay"
      bind:checked={extraOptions.autoSaveReplay}
    />
    <label for="autoSaveReplay"> Auto Save Replay </label>
    <br />
    <input
      type="checkbox"
      id="skipReplays"
      bind:checked={extraOptions.skipReplays}
    />
    <label for="skipReplays"> Skip Replays </label>
    <br />
    <input
      type="checkbox"
      id="instantStart"
      bind:checked={extraOptions.instantStart}
    />
    <label for="instantStart"> Instant Start </label>
    <br />
    <input
      type="checkbox"
      id="freeplay"
      bind:checked={extraOptions.freeplay}
    />
    <label for="freeplay"> Freeplay </label>
    <br />
  </div>
  <div class="bottomButtons">
    <p>Settings are saved automatically</p>
  </div>
</Modal>

<style>
  h1 {
    margin-bottom: 0.6rem;
  }
  .settings,
  .controls {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
  }
  #randomizeMap {
    transform: scale(1.2);
  }
  .left-controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .right-controls {
    display: flex;
    gap: 0.5rem;
  }
  .content {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .mutators {
    display: grid;
    grid-template-columns: auto auto auto;
    gap: 1rem;
  }
  .mutator {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .bottomButtons {
    display: flex;
    margin-top: 1rem;
    gap: 0.5rem;
    justify-content: space-between;
    align-items: center;
  }
  .bottomButtons :first-child {
    flex-grow: 1;
    margin-right: .5rem;
  }
  .mutatorResetButton {
    background-color: red;
  }

  .extraoptions > * {
    margin-bottom: 0.5rem;
  }

  button.start {
    background-color: #15680e;
  }
  button.stop {
    background-color: #cc1414;
  }
</style>
