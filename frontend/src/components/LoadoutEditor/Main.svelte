<script lang="ts">
import { LoadoutConfig, type TeamLoadoutConfig } from "../../../bindings/gui";
import Modal from "../Modal.svelte";
//@ts-ignore
import ItemsCsv from "../../assets/items.csv";
import { ITEM_TYPES } from "./itemtypes";
import ColorPicker from "./ColorPicker.svelte";
import ItemField from "./ItemField.svelte";

let {
  visible = $bindable(false),
  loadoutBackup,
  blueLoadout,
  orangeLoadout,
  name,
  onsave,
}: {
  loadoutBackup: LoadoutConfig | null;
  blueLoadout: TeamLoadoutConfig | null;
  orangeLoadout: TeamLoadoutConfig | null;
  name: string;
  visible: boolean;
  onsave: (value: LoadoutConfig) => void;
} = $props();

const teams = ["blue", "orange"];

let items: {
  [x: string]: { id: number; name: string }[];
} = {};

async function getAndParseItems() {
  const resp = await fetch(ItemsCsv);
  const csv = await resp.text();
  const lines = csv.split(/\r?\n/);

  for (const key in ITEM_TYPES) {
    const category = ITEM_TYPES[key].category;
    items[category] = [];
  }

  for (const line of lines) {
    const columns = line.split(",");
    const category = columns[1];

    if (items[category])
      items[category].push({ id: +columns[0], name: columns[3] });
  }

  // rename duplicate item names (append them with (2), (3), ...)
  for (const category in items) {
    const nameCounts: { [x: string]: number } = {};

    for (let item of items[category]) {
      if (nameCounts[item.name]) {
        nameCounts[item.name]++;
        item.name = `${item.name} (${nameCounts[item.name]})`;
      } else {
        nameCounts[item.name] = 1;
      }
    }
  }
}

getAndParseItems();

function getLoadout(team: string) {
  return team === "blue" ? blueLoadout : orangeLoadout;
}

function filterItems(team: string, category: string) {
  if (category !== "Skin") return items[category];

  const carId = getLoadout(team)?.carId;
  const bodyName = items.Body.find((el) => el.id === carId)?.name;

  return items.Skin.filter((el) => {
    if (el.name.includes(":")) {
      const [body, _] = el.name.split(": ");
      return body === bodyName;
    }

    return true;
  });
}

function revertChanges() {
  if (!loadoutBackup) return;
  blueLoadout = structuredClone(loadoutBackup.blueLoadout);
  orangeLoadout = structuredClone(loadoutBackup.orangeLoadout);
}
</script>

<Modal title={`Loadout of ${name}`} bind:visible>
  {#if blueLoadout && orangeLoadout}
  <div id="body">
    {#each teams as team}
      <div id={team} class="team">
        <div id="{team}-header"></div>
        <div class="team-colors">
          <ColorPicker 
            value={getLoadout(team)!.teamColorId}
            primary={true}
            team={team}
            text="Primary Color"
            onpick={(value) => {
              let loadout = getLoadout(team)!;
              loadout.teamColorId = value;
            }}
          />
  
          <ColorPicker
            value={getLoadout(team)!.customColorId}
            primary={false}
            team={team}
            text="Secondary Color"
            onpick={(value) => {
              let loadout = getLoadout(team)!;
              loadout.customColorId = value;
            }}
          />
        </div>
        <div class="items">
          {#each ITEM_TYPES as itemType}
            <ItemField
              items={filterItems(team, itemType.category)}
              itemType={itemType}
              team={team}
              value={getLoadout(team)!}
              onPickItem={(value) => {
                let loadout = getLoadout(team)!;
                // @ts-ignore
                loadout[itemType.itemKey as keyof TeamLoadoutConfig] = value;
              }}
              onPickPaint={(value) => {
                let loadout = getLoadout(team)!;
                // @ts-ignore
                loadout.paint[itemType.paintKey as keyof TeamPaintConfig] = value;
              }}
            />
          {/each}
        </div>
      </div>
    {/each}
  </div>
  <div id="footer">
    <div class="left"></div>
    <div class="right">
      <button onclick={revertChanges}>Revert Changes</button>
    </div>
  </div>
  {/if}
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
  .team {
    width: 500px;
  }
  .team > div {
    display: inline-flex;
    justify-content: space-between;
    margin-bottom: 10px;
  }
  .team-colors {
    gap: 10px;
  }
  .items {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
  #blue {
    margin-right: 15px;
  }
  #orange {
    margin-left: 15px;
  }
  #blue-header {
    width: 100%;
    height: 7px;
    background-color: rgb(0, 153, 255);
  }
  #orange-header {
    width: 100%;
    height: 7px;
    background-color: orange;
  }
</style>
