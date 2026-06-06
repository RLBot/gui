<script lang="ts">
import { fly } from "svelte/transition";
import type { ExtraOptions } from "../../../bindings/gui";
import { MAPS_NON_STANDARD, MAPS_STANDARD } from "../../arena-names";
import LauncherSelector from "../LauncherSelector.svelte";
import Modal from "../Modal.svelte";
import NiceSelect from "../NiceSelect.svelte";
import Select from "../NiceSelect.svelte";
import { type Gamemode, gamemodes } from "./rlmodes";
import { mutatorCategories, mutators as mutatorOptions } from "./rlmutators";

let {
  map = $bindable(),
  mode = $bindable(),
  extraOptions = $bindable(),
  mutators = $bindable(),
  launcherOptionsVisible = $bindable(),
  onStart,
  onStop,
}: {
  map: string;
  mode: string;
  extraOptions: ExtraOptions;
  mutators: { [k: string]: number };
  launcherOptionsVisible: boolean;
  onStart: (randomizeMap: boolean) => void;
  onStop: () => void;
} = $props();

let showExtraOptions = $state(false);
let showMutators = $state(false);
let mutatorSearchQuery = $state("");
let randomizeMap = $state(localStorage.getItem("MS_RANDOMIZE_MAP") === "true");
$effect(() => {
  localStorage.setItem("MS_RANDOMIZE_MAP", randomizeMap.toString());
});

const existingMatchBehaviors: { [n: string]: number } = {
  Restart: 0,
  "Continue and spawn": 1,
  "Restart if different": 2,
};

const renderingOptions: { [n: string]: number } = {
  "Off by default": 0,
  "On by default": 1,
  "Always off": 2,
};

const performanceMonitorOptions: { [n: string]: number } = {
  "Show when suboptimal": 0,
  "Always show": 1,
  "Never show": 2,
};

function cleanCase(toClean: string): string {
  const halfClean = toClean.replaceAll("_", " ");
  return halfClean.charAt(0).toUpperCase() + halfClean.slice(1);
}

/**
 * Build the category list while verifying that every mutator (except game_mode)
 * appears in exactly one category, so none silently drop out of the UI.
 */
function buildMutatorCategories(): { name: string; keys: string[] }[] {
  const allMutatorKeys = Object.keys(mutators);
  const seen = new Set<string>();
  const categories: { name: string; keys: string[] }[] = [];

  for (const [name, keys] of Object.entries(mutatorCategories)) {
    for (const key of keys) {
      if (seen.has(key)) {
        console.warn(`Mutator "${key}" appears in multiple categories`);
      }
      seen.add(key);
    }
    categories.push({ name, keys: [...keys] });
  }

  // Catch any mutator keys not assigned to a category
  const uncategorized = allMutatorKeys.filter(
    (key) => key !== "game_mode" && !seen.has(key),
  );
  if (uncategorized.length > 0) {
    console.warn(`Mutators not in any category: ${uncategorized.join(", ")}`);
    categories.push({ name: "other", keys: uncategorized });
  }

  // Sort once at build time
  for (const cat of categories) {
    cat.keys.sort();
  }

  return categories;
}

const mutatorCategoryOptions = buildMutatorCategories();

function filterMutatorCategories(
  categories: { name: string; keys: string[] }[],
  query: string,
): { name: string; keys: string[] }[] {
  if (!query) {
    return categories;
  }

  const lowerQuery = query.toLowerCase();
  return categories
    .map(({ name, keys }) => ({
      name,
      keys: keys.filter((key) =>
        cleanCase(key).toLowerCase().includes(lowerQuery),
      ),
    }))
    .filter((group) => group.keys.length > 0);
}

function resetMutators() {
  for (const key of Object.keys(mutators)) {
    mutators[key] = 0;
  }
  selectedPreset = "";
}

// the reason for default being "" and not null is that NiceSelect considers that the default
let selectedPreset: Gamemode | "" = $state.raw("");
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

  for (const key of allMutatorKeys) {
    if (presetData.mutators[key] !== undefined) {
      mutators[key] = mutatorOptions[key].indexOf(presetData.mutators[key]);
    } else {
      mutators[key] = 0;
    }
  }
}

const allMutatorKeys = mutatorCategoryOptions.flatMap((c) => c.keys);

let searchedMutatorOptions = $derived(
  filterMutatorCategories(mutatorCategoryOptions, mutatorSearchQuery),
);

function countModifiedMutators(): number {
  let count = 0;
  for (const key of Object.keys(mutators)) {
    if (mutators[key] !== 0) {
      count++;
    }
  }

  return count;
}
let numModifiedMutators = $derived(countModifiedMutators());

function getMaps(): { [k: string]: string } {
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
        <button onclick={() => { launcherOptionsVisible = true }}>Launcher Options</button>
      </div>
    </div>
    <div class="controls">
      <div class="left-controls">
        <button onclick={() => { showMutators = true; }}>
          Mutators
          {#if numModifiedMutators != 0}
          <span>{numModifiedMutators}</span>
          {/if}
        </button>
        <button onclick={() => { showExtraOptions = true; }}>
          Extra
        </button>
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

<Modal title="Select a launcher" bind:visible={launcherOptionsVisible}>
  <LauncherSelector />
</Modal>

<Modal title="Rocket League Mutators" bind:visible={showMutators}>
  {#snippet children()}
    <div class="mutator-search">
      <input
        type="search"
        placeholder="Search mutators…"
        bind:value={mutatorSearchQuery}
      />
    </div>
    <div class="mutators">
      {#each searchedMutatorOptions as { name, keys } (name)}
        <div class="category-header" in:fly={{ duration: 500, y: 8 }}>{cleanCase(name)}</div>
        {#each keys as mutatorKey (mutatorKey)}
          <div class="mutator" in:fly={{ duration: 500, y: 8 }}>
            <label
              class={mutators[mutatorKey] == 0 ? "" : "mutatorChanged"}
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
      {/each}
    </div>
  {/snippet}
  {#snippet footer()}
    <footer class="bottomButtons">
      <p>Settings are saved automatically</p>
      <NiceSelect bind:value={selectedPreset} options={gamemodes} placeholder="Select a preset" />
      <button
        class="mutatorResetButton"
        onclick={resetMutators}>Reset</button
      >
    </footer>
  {/snippet}
</Modal>

<Modal title="RLBot Extra Options" bind:visible={showExtraOptions}>
  {#snippet children()}
    <div class="extraoptions">
      <p>Existing match behaviour</p>
      <NiceSelect bind:value={extraOptions.existingMatchBehavior} options={existingMatchBehaviors} placeholder="Existing Match Behavior" />
      <br />
      <br />
      <p>Rendering (bots can draw on screen)</p>
      <NiceSelect bind:value={extraOptions.enableRendering} options={renderingOptions} placeholder="Rendering" />
      <br />
      <br />
      <p>Performance Monitor</p>
      <NiceSelect bind:value={extraOptions.performanceMonitor} options={performanceMonitorOptions} placeholder="Performance Monitor" />
      <br />
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
        id="autoStartAgents"
        bind:checked={extraOptions.autoStartAgents}
      />
      <label for="autoStartAgents">
        Auto-start agents
      </label>
      <br />
      <input
        type="checkbox"
        id="waitForAgents"
        bind:checked={extraOptions.waitForAgents}
      />
      <label for="waitForAgents">
        Wait for agents to connect
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
  {/snippet}
  {#snippet footer()}
    <footer class="bottomButtons">
      <p>Settings are saved automatically</p>
    </footer>
  {/snippet}
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
  .controls button span {
    margin-left: 0.5rem;
    background-color: var(--orange);
    color: var(--foreground-opp);
    padding: 0.1rem 0.3rem;
    border-radius: 0.2rem;
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

  .mutator-search {
    flex-shrink: 0;
    background: var(--background);
    padding: 0.5rem 0 0.75rem 0;
  }
  .mutator-search input {
    width: 100%;
    box-sizing: border-box;
    padding: 0.4rem 0.6rem;
    border-radius: 0.3rem;
    border: 1px solid var(--border, #555);
    background: var(--background, #222);
    color: var(--foreground, #eee);
    font-size: 0.9rem;
  }
  .mutator-search input:focus {
    outline: none;
    border-color: var(--accent, #4a9eff);
  }

  :global(.modalBody) {
    display: flex;
    flex-direction: column;
  }

  .mutators {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: grid;
    grid-template-columns: auto auto auto auto auto;
    gap: 1rem;
    align-content: start;
  }
  @media (max-width: 980px) {
    .mutators {
      grid-template-columns: auto auto auto auto;
    }
  }
  @media (max-width: 840px) {
    .mutators {
      grid-template-columns: auto auto auto;
    }
  }
  @media (max-width: 650px) {
    .mutators {
      grid-template-columns: auto auto;
    }
  }
  .mutator {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }
  .mutator label {
    color: var(--foreground);
  }
  label.mutatorChanged {
    color: var(--orange);
  }
  .category-header {
    grid-column: 1 / -1;
    font-weight: 700;
    font-size: 1.1rem;
    margin-top: 0.5rem;
    padding-bottom: 0.25rem;
    border-bottom: 1px solid var(--border, #555);
    text-transform: capitalize;
  }
  .bottomButtons {
    display: flex;
    gap: 0.5rem;
    justify-content: space-between;
    align-items: center;
    position: sticky;
    bottom: 0;
    background: var(--background);
    z-index: 1;
  }
  .bottomButtons :first-child {
    flex-grow: 1;
    margin-right: .5rem;
  }
  .mutatorResetButton {
    background-color: red;
    color: white
  }

  .extraoptions > * {
    margin-bottom: 0.5rem;
  }

  button.start, button.stop {
    color: white
  }
  button.start {
    background-color: #15680e;
  }
  button.stop {
    background-color: #cc1414;
  }
</style>
