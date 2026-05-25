<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RLLobbyListing } from "../../bindings/gui/index.js";
import reloadIcon from "../assets/reload.svg";
import Modal from "./Modal.svelte";

let {
  servers = $bindable(),
  launcherOptionsVisible = $bindable(),
  visible = $bindable(false),
} = $props();
let joiningLobbyId = $state<string | null>(null);
let lobbies = $state<RLLobbyListing[]>([]);
let lobbiesLoading = $state(false);
let lobbiesError = $state<string | null>(null);
let hasLoaded = $state(false);

async function loadLobbies() {
  if (lobbiesLoading) return;
  lobbiesLoading = true;
  lobbiesError = null;

  try {
    lobbies = await App.GetRLLobbies();
    hasLoaded = true;
  } catch (error) {
    lobbiesError = error?.message || String(error);
  } finally {
    lobbiesLoading = false;
  }
}

$effect(() => {
  if (!visible && hasLoaded) {
    hasLoaded = false;
  }
});

$effect(() => {
  if (visible && !hasLoaded) {
    loadLobbies();
  }
});

function getServerName(name: string, ipAddress: string) {
  if (servers) {
    const server = servers.find((s) => s.ip === ipAddress);
    if (server) {
      return server.location;
    }
  }

  return name.replace(/:\d+$/, "");
}

async function joinLobby(lobby: RLLobbyListing) {
  if (joiningLobbyId) return;

  let launcher = localStorage.getItem("MS_LAUNCHER");
  if (!launcher) {
    toast.error("Please select a launcher first", {
      position: "top-center",
      duration: 5000,
    });

    launcherOptionsVisible = true;
    return;
  }

  joiningLobbyId = lobby.id;
  const serverAddr = `${lobby.ipAddress}:${lobby.port}`;
  const lobbyName = getServerName(lobby.name, lobby.ipAddress);
  const toastId = toast.loading(`Joining ${lobbyName}...`, {
    position: "top-center",
  });

  try {
    await App.JoinLobby({
      server: serverAddr,
      map: lobby.map,
      blueBots: [],
      orangeBots: [],
      launcher,
      launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || "",
    });
    toast.success(`Joining ${lobbyName}`, {
      id: toastId,
      position: "top-center",
      duration: 5000,
    });
  } catch (error) {
    toast.error(`Failed to join lobby\n${error}`, {
      id: toastId,
      position: "top-center",
      duration: 8000,
    });
  } finally {
    joiningLobbyId = null;
  }
}
</script>

{#snippet lobbyTitle()}
  <div class="lobbyTitle">
    <span>RocketHost Lobbies</span>
    <button
      class="refresh"
      type="button"
      title="Refresh lobbies"
      disabled={lobbiesLoading}
      onclick={loadLobbies}
    >
      <img src={reloadIcon} alt="Refresh lobbies" />
    </button>
  </div>
{/snippet}

<Modal title={lobbyTitle} bind:visible>
  {#if !hasLoaded && lobbiesLoading}
    <p>Loading lobbies...</p>
  {:else if lobbies.length === 0 && !lobbiesError}
    <p>No public lobbies found.</p>
  {:else if lobbies.length > 0}
    <div class="lobbyList">
      {#each lobbies as lobby}
        <div class="lobbyEntry">
          <div class="lobbyHeader">
            <h3>{getServerName(lobby.name, lobby.ipAddress)}:{lobby.port}</h3>
            <div class="lobbyActions">
              <span>{lobby.playerCount} player(s)</span>
              <button
                class="join"
                disabled={joiningLobbyId === lobby.id}
                onclick={() => joinLobby(lobby)}
              >
                {joiningLobbyId === lobby.id ? "Joining..." : "Join"}
              </button>
            </div>
          </div>
          <div class="lobbyMeta">
            <span>{lobby.map}</span>
            <span>|</span>
            <span>{lobby.secondsSinceUpdate}s ago</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if lobbiesError}
    <p class="lobbyError">Failed to load lobbies: {lobbiesError}</p>
  {/if}
</Modal>

<style>
  .lobbyTitle {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
  }
  .lobbyTitle .refresh {
    padding: 2px 4px 2px 4px;
  }
  .lobbyTitle .refresh:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  .refresh img {
    filter: invert() brightness(var(--icon-brightness));
    width: 24px;
  }
  .lobbyList {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    min-width: 400px;
  }
  .lobbyEntry {
    background-color: var(--background-alt);
    border-radius: 0.4rem;
    padding: 0.6rem 0.8rem;
  }
  .lobbyHeader {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }
  .lobbyActions {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }
  button.join {
    background-color: #15680e;
  }
  button.join:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  .lobbyHeader h3 {
    margin: 0;
    font-size: 1.1rem;
  }
  .lobbyMeta {
    display: flex;
    gap: 0.6rem;
    font-size: 0.9rem;
    opacity: 0.85;
  }
  .lobbyError {
    color: var(--orange);
    font-weight: 600;
  }
</style>
