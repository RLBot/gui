<script lang="ts">
import { COLORS } from "./colors";

let {
  value = $bindable(),
  team,
  text,
}: {
  value: number;
  team: string | null;
  text: string;
} = $props();

let pickedColor = $derived(getColorStyle(value));

const ROWS = 7;
const COLUMNS = team ? 10 : 15;

function getColorStyle(colorID: number) {
  const colors = team ? COLORS[team] : COLORS.secondary;
  const rgb = colors[colorID];
  return rgb ? rgb.toString() : "";
}

function pickedItemClass(colorID: number) {
  return value === colorID ? "selected-color" : "";
}

const COLOR_IDS: number[][] = [];
for (let i = 0; i < ROWS; i++) {
  COLOR_IDS.push([]);
  for (let j = 0; j < COLUMNS; j++) {
    COLOR_IDS[i].push(i * COLUMNS + j);
  }
}
</script>

<div class="dropdown">
  <button>
    <span class="color-indicator" style="background-color: rgb({pickedColor})"></span>
    {text}
  </button>
  <div class="dropmenu {team ? 'left' : 'center'}">
    <table>
      <tbody>
      {#each COLOR_IDS as idsRow}
        <tr>
          {#each idsRow as id}
            <td>
              <!-- svelte-ignore a11y_consider_explicit_label -->
              <button
                style="background-color: rgb({getColorStyle(id)})"
                onclick={() => value = id}
              >
                <div class="colorpicker-color {pickedItemClass(id)}"></div>
              </button>
            </td>
          {/each}
        </tr>
      {/each}
      </tbody>
    </table>
  </div>
</div>

<style>
  button {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  table {
    border-spacing: 0;
  }
  .dropmenu {
    padding: 0px;
    margin-left: 10px;
    margin-right: 10px;
    margin-bottom: -5px;
    border: 8px solid var(--background-but);
    border-radius: 5px;
  }
  .dropmenu td, .dropmenu button {
    border: 0;
    border-radius: 0;
    padding: 0;
  }
  .colorpicker-color {
    width: 25px;
    height: 25px;
    cursor: pointer;
  }
  .colorpicker-color:hover {
    border: 2px solid rgba(0, 0, 0, 0.74);
  }
  .selected-color {
    border: 2px dashed rgba(0, 0, 0, 0.897);
  }
  .color-indicator {
    border-radius: 12px;
    width: 24px;
    height: 24px;
    display: inline-block;
    vertical-align: middle;
  }
</style>
