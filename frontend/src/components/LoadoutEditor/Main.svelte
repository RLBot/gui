<script lang="ts">
import type { LoadoutConfig, TeamLoadoutConfig } from "../../../bindings/gui";
import Modal from "../Modal.svelte";
//@ts-ignore
import ItemsCsv from "../../assets/items.csv";
import { ITEM_TYPES } from "./itemtypes";
import TeamEditor from "./TeamEditor.svelte";

let {
  visible = $bindable(false),
  loadout,
  name,
  onsave,
}: {
  loadout: LoadoutConfig;
  name: string;
  visible: boolean;
  onsave: (value: LoadoutConfig) => void;
} = $props();

let blueLoadout: TeamLoadoutConfig = $state(structuredClone(loadout.blueLoadout));
let orangeLoadout: TeamLoadoutConfig = $state(structuredClone(loadout.orangeLoadout));

let items: {
  [x: string]: { id: number; name: string }[];
} | null = $state(null);

async function getAndParseItems() {
  const resp = await fetch(ItemsCsv);
  const csv = await resp.text();
  const lines = csv.split(/\r?\n/);

  let newItems: {
    [x: string]: { id: number; name: string }[];
  } = {};

  for (const key in ITEM_TYPES) {
    const category = ITEM_TYPES[key].category;
    newItems[category] = [];
  }

  for (const line of lines) {
    const columns = line.split(",");
    const category = columns[1];

    if (newItems[category])
      newItems[category].push({ id: +columns[0], name: columns[3] });
  }

  // rename duplicate item names (append them with (2), (3), ...)
  for (const category in newItems) {
    const nameCounts: { [x: string]: number } = {};

    for (const item of newItems[category]) {
      if (nameCounts[item.name]) {
        nameCounts[item.name]++;
        item.name = `${item.name} (${nameCounts[item.name]})`;
      } else {
        nameCounts[item.name] = 1;
      }
    }
  }

  items = newItems;
}

getAndParseItems();

function revertChanges() {
  blueLoadout = structuredClone(loadout.blueLoadout);
  orangeLoadout = structuredClone(loadout.orangeLoadout);
}
</script>

<Modal title={`Loadout of ${name}`} bind:visible>
  <div id="body">
    {#if visible && items}
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
    {/if}
  </div>
  <div id="footer">
    <div class="left"></div>
    <div class="right">
      <button onclick={revertChanges}>Revert Changes</button>
    </div>
  </div>
</Modal>

<style>
  #body, #footer {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }
  #body {
    gap: 15px;
    flex-wrap: wrap;
    overflow: hidden;
    align-items: center;
    justify-content: center;
  }
  #footer {
    margin-top: 10px;
  }
</style>
