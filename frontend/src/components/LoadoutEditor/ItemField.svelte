<script lang="ts">
import type { TeamLoadoutConfig } from "../../../bindings/gui";
import { PAINTS } from "./colors";
import type { ItemType } from "./itemtypes";
import ArrowsIcon from "../../assets/arrows.svg";
import RandomIcon from "../../assets/random.svg";

let {
  value = $bindable(),
  items,
  itemType,
  team,
}: {
  value: TeamLoadoutConfig;
  items: { id: number; name: string }[];
  itemType: ItemType;
  team: string;
} = $props();

let itemSelection = $state(loadItemSelection());
function loadItemSelection() {
  const loadout = value[itemType.itemKey];
  if (loadout === 0) return;

  const item = items.find((el) => el.id === loadout);
  if (!item) {
    console.warn(`Item with id ${loadout} not found`);
    return;
  }

  return item.name;
}

$effect(() => {
  itemSelection = loadItemSelection();
})

$effect(() => {
  const item = items.find((el) => el.name === itemSelection);
  if (!item) {
    return;
  }

  // @ts-ignore
  value[itemType.itemKey] = item.id;
});

let selectedPaint = $state(loadPaintSelection());
function loadPaintSelection() {
  if (!itemType.paintKey) {
    return;
  }

  const paint = value.paint[itemType.paintKey];
  return paint;
}

$effect(() => {
  selectedPaint = loadPaintSelection();
});

$effect(() => {
  if (!selectedPaint || !itemType.paintKey) {
    return;
  }

  value.paint[itemType.paintKey] = selectedPaint;
});

function selectedPaintColorClass() {
  const color = PAINTS.find((el) => el.id === selectedPaint);
  return color ? color.class : "";
}

function randomizeItem() {
  const randomIndex = Math.floor(Math.random() * items.length);
  itemSelection = items[randomIndex].name;
}
</script>

<div id="row">
  <div class="expandable-input">
    <label for="item-selection">{itemType.name}</label>
  </div>
  <div id="text-input" class="input-group shinkable-input">
    <button class="randomize" onclick={randomizeItem}>
      <img src={RandomIcon} alt="Randomize colors" />
    </button>
    <input
      type="text"
      id="item-selection"
      list="list{itemType.name}{team}"
      autocomplete="off"
      bind:value={itemSelection}
      onmousedown={() => itemSelection = ""}
    >
    <datalist id="list{itemType.name}{team}">
      {#each items as item}
        <option value={item.name}></option>
      {/each}
    </datalist>
  </div>
  <div class="input-group paint-group {itemType.paintKey ? "" : "hidden"}">
    <select
      id="paintSelection"
      bind:value={selectedPaint}
      class="paint-color {selectedPaintColorClass()}"
      style="background-image: url({ArrowsIcon})"
    >
      {#each PAINTS as color}
        <option value={color.id} class="paint-color {color.class}">{color.name}</option>
      {/each}
    </select>
  </div>
</div>

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
  button.randomize {
    padding: 0;
    border-radius: 0;
    height: 100%;
    padding: 0 0.5rem;
  }
  button.randomize > img {
    filter: invert() brightness(90%);
    vertical-align: middle;
    width: 20px;
    height: 20px;
  }
  #item-selection {
    border: none;
    border-radius: 0;
  }
  #text-input {
    border: solid 1px var(--foreground-alt);
    background-color: var(--background-but);
    border-radius: 0.25rem;
  }
  #row {
    display: flex;
    justify-content: space-between;
    width: 100%;
    gap: 10px;
  }
  .hidden {
    visibility: hidden;
  }
  .expandable-input {
    display: flex;
    align-items: center;
    flex-grow: 1;
  }
  .input-group {
    display: flex;
    align-items: center;
  }
  .paint-group {
    gap: 10px;
    width: max-content;
  }
  label {
    white-space: nowrap;
  }
  .paint-color {
    width: 100%;
  }
  .paint-color.black {
    background-color: #111;
    color: #dddddd;
    border-color: #dddddd;
  }
  .paint-color.burntsienna {
    background-color: #ffc2b1;
    color: #882104;
    border-color: #882104;
  }
  .paint-color.cobalt {
    background-color: #ccd3ff;
    color: #3f51b5;
    border-color: #3f51b5;
  }
  .paint-color.crimson {
    background-color: #ffcece;
    color: #d50000;
    border-color: #d50000;
  }
  .paint-color.forestgreen {
    background-color: #aae7ac;
    color: #199e1e;
    border-color: #199e1e;
  }
  .paint-color.grey {
    background-color: #cacaca;
    color: #3d3d3d;
    border-color: #3d3d3d;
  }
  .paint-color.lime {
    background-color: #f3ffd2;
    color: #5ebd00;
    border-color: #5ebd00;
  }
  .paint-color.orange {
    background-color: #fff3d3;
    color: #ff9d00;
    border-color: #ff9d00;
  }
  .paint-color.pink {
    background-color: #ffcdde;
    color: #ff4081;
    border-color: #ff4081;
  }
  .paint-color.purple {
    background-color: #e2b4eb;
    color: #9c27b0;
    border-color: #9c27b0;
  }
  .paint-color.saffron {
    background-color: #fffce2;
    color: #ffd000;
    border-color: #ffd000;
  }
  .paint-color.skyblue {
    background-color: #c4ecff;
    color: #03a9f4;
    border-color: #03a9f4;
  }
  .paint-color.titaniumwhite {
    background-color: #fff;
    color: #929292;
    border-color: #929292;
  }
  .paint-color.gold {
    background-color: #daa520;
    color: #3d3d3d;
    border-color: #3d3d3d;
  }
  .paint-color.rosegold {
    background-color: #b76e79;
    color: #3d3d3d;
    border-color: #3d3d3d;
  }
  .paint-color.whitegold {
    background-color: #f0f0f0;
    color: #3d3d3d;
    border-color: #3d3d3d;
  }
  .paint-color.onyx {
    background-color: #000000;
    color: #ffffff;
    border-color: #ffffff;
  }
  .paint-color.platinum {
    background-color: #e5e5e5;
    color: #3d3d3d;
    border-color: #3d3d3d;
  }
</style>
