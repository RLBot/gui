<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RLLobbyListing } from "../../bindings/gui/index.js";
import Modal from "./Modal.svelte";

let {
  servers = $bindable(),
  launcherOptionsVisible = $bindable(),
  visible = $bindable(false),
} = $props();
let joiningLobbyId = $state<string | null>(null);

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

<Modal title="RocketHost Lobbies" bind:visible>
  {#if visible}
    {#await App.GetRLLobbies()}
      <p>Loading lobbies...</p>
    {:then lobbies}
      {#if lobbies.length === 0}
        <p>No public lobbies found.</p>
      {:else}
        <div class="lobbyList">
          {#each lobbies as lobby}
            <div class="lobbyEntry">
              <div class="lobbyHeader">
                <h3>{getServerName(lobby.name, lobby.ipAddress)}</h3>
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
    {:catch error}
      <p class="lobbyError">Failed to load lobbies: {error?.message || error}</p>
    {/await}
  {/if}
</Modal>

<style>
  .lobbyList {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    min-width: 350px;
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
