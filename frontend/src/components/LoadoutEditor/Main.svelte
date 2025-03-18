<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, LoadoutConfig, TeamLoadoutConfig } from "../../../bindings/gui";
import Modal from "../Modal.svelte";
//@ts-ignore
import { type CsvItem } from "./itemtypes";
import TeamEditor from "./TeamEditor.svelte";

let {
  visible = $bindable(false),
  loadout = $bindable(),
  basePath,
  loadoutFile,
  items,
  name,
}: {
  visible: boolean;
  loadout: LoadoutConfig;
  basePath: string;
  loadoutFile: string;
  items: {
    [x: string]: CsvItem[];
  };
  name: string;
} = $props();

let blueLoadout: TeamLoadoutConfig = $state(
  structuredClone(loadout.blueLoadout),
);
let orangeLoadout: TeamLoadoutConfig = $state(
  structuredClone(loadout.orangeLoadout),
);
$effect(() => {
  blueLoadout = structuredClone(loadout.blueLoadout);
  orangeLoadout = structuredClone(loadout.orangeLoadout);
});

function revertChanges() {
  blueLoadout = structuredClone(loadout.blueLoadout);
  orangeLoadout = structuredClone(loadout.orangeLoadout);
}

function saveLoadout() {
  // structuredClone doesn't work here, likely because of $state
  loadout.blueLoadout = JSON.parse(JSON.stringify(blueLoadout));
  loadout.orangeLoadout = JSON.parse(JSON.stringify(orangeLoadout));

  App.SaveLoadoutToFile(basePath, loadoutFile, loadout)
    .then(() => {
      visible = false;
      toast.success(`Saved the loadout of ${name}`);
    })
    .catch((e) => {
      toast.error(`Failed to save the loadout of ${name}: ${e}`);
    });
}
</script>

<Modal title={`Loadout of ${name}`} bind:visible>
  <div id="body">
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
  </div>
  <div id="footer">
    <div class="left"></div>
    <div class="right">
      <button type="submit" onclick={saveLoadout}>Save and close</button>
      <button type="reset" onclick={revertChanges}>Revert changes</button>
    </div>
  </div>
</Modal>

<style>
  .right {
    gap: 10px;
  }
  #body, #footer {
    width: 100%;
    display: flex;
    justify-content: space-between;
  }
  #body {
    gap: 30px;
    flex-wrap: wrap;
    overflow: hidden;
    align-items: center;
    justify-content: center;
  }
  #footer {
    margin-top: 10px;
  }
</style>
