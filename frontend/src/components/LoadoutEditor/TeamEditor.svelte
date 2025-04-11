<script lang="ts">
import type { TeamLoadoutConfig } from "../../../bindings/gui";
import RandomIcon from "../../assets/random.svg";
import ColorPicker from "./ColorPicker.svelte";
import ItemField from "./ItemField.svelte";
import { COLORS, PAINTS } from "./colors";
import type { CsvItem } from "./items";
import { ITEM_TYPES } from "./itemtypes";

let {
  items,
  team,
  loadout = $bindable(),
  onchange,
}: {
  items: {
    [x: string]: CsvItem[];
  };
  team: string;
  loadout: TeamLoadoutConfig;
  onchange: () => void;
} = $props();

function filterItems(category: string) {
  if (category !== "Skin" || loadout.carId === 0) return items[category];

  const bodyName = items.Body.find((el) => el.id === loadout.carId)?.name;

  return items.Skin.filter((el) => {
    if (el.name.includes(":")) {
      const [body, _] = el.name.split(": ");
      return body === bodyName;
    }

    return true;
  });
}

function randomizeTeamLoadout() {
  loadout.teamColorId = Math.floor(Math.random() * COLORS[team].length);
  loadout.customColorId = Math.floor(Math.random() * COLORS.secondary.length);

  for (const itemType of ITEM_TYPES) {
    const items = filterItems(itemType.category);
    const randomItem = items[Math.floor(Math.random() * items.length)];

    // @ts-ignore
    loadout[itemType.itemKey] = randomItem.id;

    if (itemType.paintKey) {
      loadout.paint[itemType.paintKey] = Math.floor(
        Math.random() * PAINTS.length,
      );
    }
  }
}
</script>

{#if loadout}
<div id={team} class="team">
  <div id="{team}-header"></div>
  <div class="team-colors">
    <ColorPicker 
      bind:value={loadout.teamColorId}
      team={team}
      text="Primary Color"
      onchange={onchange}
    />

    <ColorPicker
      bind:value={loadout.customColorId}
      team={null}
      text="Secondary Color"
      onchange={onchange}
    />

    <button class="randomize" onclick={randomizeTeamLoadout} title="Randomize entire {team} team loadout">
      <img src={RandomIcon} alt="Randomize colors" />
    </button>
  </div>
  <div class="items">
    {#each ITEM_TYPES as itemType}
      <ItemField
        items={filterItems(itemType.category)}
        itemType={itemType}
        team={team}
        bind:value={loadout}
        onchange={onchange}
      />
    {/each}
  </div>
</div>
{/if}

<style>
  button.randomize {
    margin-left: auto;
  }
  button.randomize > img {
    filter: invert() brightness(90%);
    vertical-align: middle;
    width: 24px;
    height: 24px;
  }
  .team {
    width: 530px;
  }
  .team > div {
    display: inline-flex;
    justify-content: space-between;
    margin-bottom: 10px;
  }
  .team-colors {
    gap: 10px;
    width: 100%;
  }
  .items {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
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