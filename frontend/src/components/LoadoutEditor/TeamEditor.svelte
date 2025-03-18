<script lang="ts">
import ColorPicker from "./ColorPicker.svelte";
import ItemField from "./ItemField.svelte";
import RandomIcon from "../../assets/random.svg";
import type { TeamLoadoutConfig } from "../../../bindings/gui";
import { COLORS } from "./colors";
import { ITEM_TYPES } from "./itemtypes";

let {
  items,
  team,
  loadout = $bindable(),
}: {
  items: {
    [x: string]: { id: number; name: string }[];
  };
  team: string;
  loadout: TeamLoadoutConfig;
} = $props();

function filterItems(category: string) {
  if (category !== "Skin") return items[category];

  const bodyName = items.Body.find((el) => el.id === loadout.carId)?.name;

  return items.Skin.filter((el) => {
    if (el.name.includes(":")) {
      const [body, _] = el.name.split(": ");
      return body === bodyName;
    }

    return true;
  });
}

function randomizeColors() {
  loadout.teamColorId = Math.floor(Math.random() * COLORS[team].length);
  loadout.teamColorId = Math.floor(Math.random() * COLORS.secondary.length);
}
</script>

{#if loadout}
<div id={team} class="team">
  <div id="{team}-header"></div>
  <div class="team-colors">
    <ColorPicker 
      bind:value={loadout.teamColorId}
      primary={true}
      team={team}
      text="Primary Color"
    />

    <ColorPicker
      bind:value={loadout.teamColorId}
      primary={false}
      team={team}
      text="Secondary Color"
    />

    <button class="randomize" onclick={randomizeColors}>
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