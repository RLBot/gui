<script lang="ts">
import type { TeamLoadoutConfig, TeamPaintConfig } from "../../../bindings/gui";
import { PAINTS } from "./colors";
import type { ItemType } from "./itemtypes";
import ArrowsIcon from "../../assets/arrows.svg";

let {
  value,
  items,
  itemType,
  team,
  onPickItem,
  onPickPaint,
}: {
  value: TeamLoadoutConfig;
  items: { id: number; name: string }[];
  itemType: ItemType;
  team: string;
  onPickItem: (value: number) => void;
  onPickPaint: (value: number) => void;
} = $props();

let itemSelection = $state(loadItemSelection());
function loadItemSelection() {
  let loadout = value[itemType.itemKey as keyof TeamLoadoutConfig];
  let item = items.find((el) => el.id === loadout);
  if (!item) {
    return;
  }

  return item.name;
}

let selectedPaint = $state(loadPaintSelection());
function loadPaintSelection() {
  if (!itemType.paintKey) {
    return;
  }

  let paint = value.paint[itemType.paintKey as keyof TeamPaintConfig];
  return paint;
}

function selectedPaintColorClass() {
  let color = PAINTS.find((el) => el.id === selectedPaint);
  return color ? color.class : "";
}

$effect(() => {
  let item = items.find((el) => el.name === itemSelection);
  if (!item) {
    return;
  }

  onPickItem(item.id);
});

$effect(() => {
  if (!selectedPaint) {
    return;
  }

  onPickPaint(selectedPaint);
});
</script>

<div id="row">
  <div class="expandable-input">
    <label for="itemSelection">{itemType.name}</label>
  </div>
  <div class="input-group shinkable-input">
    <input
      type="text"
      id="itemSelection"
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
    gap: 10px;
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
