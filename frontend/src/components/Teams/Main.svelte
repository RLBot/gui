<script lang="ts">
import type { DraggablePlayer } from "../..";
import TeamBotList from "./TeamBotList.svelte";

let {
  bluePlayers = $bindable(),
  orangePlayers = $bindable(),
  selectedTeam = $bindable(null),
  globalAutoStart = $bindable(true),
  duplicateAgentIds = new Set<string>(),
}: {
  bluePlayers: DraggablePlayer[];
  orangePlayers: DraggablePlayer[];
  selectedTeam: "blue" | "orange" | null;
  globalAutoStart: boolean;
  duplicateAgentIds: Set<string>;
} = $props();

function toggleTeam(team: "blue" | "orange") {
  if (selectedTeam === team) {
    selectedTeam = null;
  } else {
    selectedTeam = team;
  }
}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="teams">
    <div class="team box blurred blue" onclick={() => toggleTeam('blue')} class:selected={selectedTeam === 'blue'}>
        <header class="blue">
            <h3>Blue team</h3>
            <div style="flex: 1;"></div>
            <h3 class="dimmed">{bluePlayers?.length || 0} bots</h3>
        </header>
        <TeamBotList bind:items={bluePlayers} bind:globalAutoStart {duplicateAgentIds} />
    </div>
    <div class="team box blurred orange" onclick={() => toggleTeam('orange')} class:selected={selectedTeam === 'orange'}>
        <header class="orange">
            <h3>Orange team</h3>
            <div style="flex: 1;"></div>
            <h3 class="dimmed">{orangePlayers?.length || 0} bots</h3>
        </header>
        <TeamBotList bind:items={orangePlayers} bind:globalAutoStart {duplicateAgentIds} />
    </div>
</div>

<style>
    .box {
        border-radius: 0.4rem;
        background-color: var(--background);
        padding: 0px 0.6rem;
    }
    .teams {
        display: flex;
        gap: 1rem;
        height: 100%;
    }
    .teams > .team {
        width: 50%;
        padding: 0px 0;
        display: flex;
        flex-direction: column;
        border: 2px solid transparent;
        border-radius: 0.6rem;
        max-height: 100%; /* Ensures the team does not grow beyond its container */
    }
    .teams > .team > header {
        border-radius: 0.4rem 0.4rem 0px 0px;
        padding: 0px 0.6rem;
        text-transform: uppercase;
        display: flex;
        color: white;
    }
    header {
        border: 2px solid;
        color: white;
    }
    header h3 {
        color: white !important;
    }
    header.blue {
        border-color: #0054a6;
        background-color: #0054a6;
    }
    header.orange {
        border-color: #f26522;
        background-color: #f26522;
    }
    .dimmed {
        color: #ffffffcc;
    }
    .team.selected {
        border: 2px solid;
    }
    .team.selected.blue {
        border-color: #0054a6;
    }
    .team.selected.orange {
        border-color: #f26522;
    }
</style>
