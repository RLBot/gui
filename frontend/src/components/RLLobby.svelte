<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RLLobbyListing } from "../../bindings/gui/index.js";
import reloadIcon from "../assets/reload.svg";

let { servers = $bindable(), launcherOptionsVisible = $bindable() } = $props();
let joiningLobbyId = $state<string | null>(null);
let lobbies = $state<RLLobbyListing[]>([]);
let lobbiesLoading = $state(false);
let lobbiesError = $state<string | null>(null);
let selectedLobby = $state<RLLobbyListing>({
  ipAddress: "127.0.0.1",
  port: 7777,
  map: "Stadium_P",
});

async function loadLobbies() {
  if (lobbiesLoading) return;
  lobbiesLoading = true;
  lobbiesError = null;

  try {
    lobbies = await App.GetRLLobbies();
    lobbies.sort((a, b) => a.secondsSinceUpdate - b.secondsSinceUpdate);
  } catch (error) {
    lobbiesError = error?.message || String(error);
  } finally {
    lobbiesLoading = false;
  }
}
loadLobbies();

function getServerName(name: string, ipAddress: string) {
  if (servers) {
    const server = servers.find((s) => s.ip === ipAddress);
    if (server) {
      return server.location;
    }
  }

  if (name === undefined) {
    return ipAddress;
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

<div class="container">
  <div class="lobbyTitle">
    <h2>RocketHost Lobbies</h2>
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

  {#if !lobbiesError}
    <div class="lobbyList">
      {#if lobbies.length === 0}
        <p>No public lobbies found.</p>
      {:else if lobbies.length > 0}
          {#each lobbies as lobby}
            <div class="lobbyEntry blurred">
              <div class="lobbyHeader">
                <h3>{getServerName(lobby.name, lobby.ipAddress)}:{lobby.port}</h3>
                <div class="lobbyActions">
                  <button
                    class="select"
                    disabled={joiningLobbyId === lobby.id}
                    onclick={() => selectedLobby = lobby}
                  >
                    Select
                    <!-- {joiningLobbyId === lobby.id ? "Joining..." : "Join"} -->
                  </button>
                </div>
              </div>
              <div class="lobbyMeta">
                <span>{lobby.map}</span>
                <span>|</span>
                <span>{lobby.secondsSinceUpdate}s ago</span>
                <span>|</span>
                <span>{lobby.playerCount} player(s)</span>
              </div>
            </div>
          {/each}
      {/if}
    </div>
  {:else}
    <p class="lobbyError">Failed to load lobbies: {lobbiesError}</p>
  {/if}
  <div class="inputs">
    <div style="display:flex;flex-direction:column;">
      <label>Ip/Host</label>
      <input bind:value={selectedLobby.ipAddress} type="text" placeholder="Host/IP">
    </div>
    <div style="display:flex;flex-direction:column;">
      <label>Port</label>
      <input bind:value={selectedLobby.port} type="number" placeholder="Port">
    </div>
    <button onclick={joinLobby.bind(null, selectedLobby)}>Join</button>
  </div>
</div>

<style>
  .container {
    display: flex;
    flex-direction: column;
    align-self: stretch;
    margin: 2rem;
    margin-top: 0;
    gap: 1rem;
  }
  .lobbyTitle {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    width: 100%;
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
    width: 100%;
    display: grid;
    /* The Magic Line */
    grid-template-columns: repeat(auto-fit, minmax(20rem, 1fr));
    gap: 0.6rem;
  }
  .lobbyEntry {
    display: flex;
    flex-direction: column;
    gap: .5rem;
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
  button.select {
    background-color: var(--blue);
    padding: .2rem;
  }
  button.select:disabled {
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
  .inputs {
    width: 100%;
    display: flex;
    gap: .2rem;
    justify-content: end;
    align-items: end;
  }
  .inputs button {
    background: green;
    height: fit-content;
    padding-left:  1.5rem;
    padding-right: 1.5rem;
  }
</style>
