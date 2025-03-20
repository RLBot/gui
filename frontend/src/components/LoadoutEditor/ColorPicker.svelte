<script lang="ts">
import { COLORS } from "./colors";

let {
  value,
  primary,
  team,
  text,
  onpick,
}: {
  value: number;
  primary: boolean;
  team: string;
  text: string;
  onpick: (value: number) => void;
} = $props();

const ROWS = 7;
const COLUMNS = primary ? 10 : 15;

function getColorStyle(colorID: number) {
  let colors = primary ? COLORS[team] : COLORS.secondary;
  let rgb = colors[colorID];
  return rgb ? rgb.toString() : "";
}

function getColorID(row: number, column: number) {
  return row * COLUMNS + column;
}

function pickedItemClass(colorID: number) {
  return value === colorID ? "selected-color" : "";
}

function pickColor(colorID: number) {
  value = colorID;
  onpick(colorID);
}
</script>

<div class="dropdown">
  <button>
    <span class="color-indicator" style="background-color: rgb({getColorStyle(value)})"></span>
    {text}
  </button>
  <div class="dropmenu {primary ? 'left' : 'center'}">
    <table>
      <tbody>
      {#each Array(ROWS) as _, i}
        <tr>
          {#each Array(COLUMNS) as _, j}
            <td>
              <!-- svelte-ignore a11y_consider_explicit_label -->
              <button
                style="background-color: rgb({getColorStyle(getColorID(i, j))})"
                onclick={() => pickColor(getColorID(i, j))}
              >
                <div class="colorpicker-color {pickedItemClass(getColorID(i, j))}"></div>
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
