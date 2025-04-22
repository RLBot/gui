<script lang="ts">
import toast from "svelte-5-french-toast";
import { App, RHostBot, RHostServer } from "../../bindings/gui/index.js";
import { MAPS_STANDARD } from "../arena-names";
import closeIcon from "../assets/close.svg";
import Plus from "../assets/plus.svg.svelte";
import LauncherSelector from "../components/LauncherSelector.svelte";
import { mapStore } from "../settings";

let waiting = $state(false);

let bots: RHostBot[] = $state([]);
let botFamilies = $derived.by(() => {
  let families: {
    [name: string]: string[];
  } = {};
  for (const bot of bots) {
    const fam = bot.family !== "" ? bot.family : bot.name;
    if (!Object.hasOwn(families, fam)) {
      families[fam] = [];
    }
    families[fam].push(bot.name);
  }
  return families;
});

function refreshRHostBots() {
  App.GetRHostBots()
    .then((result) => {
      bots = result;
    })
    .catch((error) => {
      toast.error(`Couldn't resolve Rocket Host bots\n${error}`, {
        position: "top-center",
        duration: 5000,
      });
    });
}
refreshRHostBots();

let serverAddr: string = $state(
  localStorage.getItem("RHOST_SERVER_ADDR") || "",
);
$effect(() => {
  localStorage.setItem("RHOST_SERVER_ADDR", serverAddr);
});

let servers: RHostServer[] = $state([]);
$effect(() => {
  if (servers.length > 0) {
    serverAddr =
      serverAddr === "" ? `${servers[0].ip}:${servers[0].port}` : serverAddr;
  }
});

function refreshRHostServers() {
  App.GetRHostServers()
    .then((result) => {
      servers = result;
    })
    .catch((error) => {
      toast.error(`Couldn't resolve Rocket Host server addresses\n${error}`, {
        position: "top-center",
        duration: 5000,
      });
    });
}
refreshRHostServers();

let blueBots: string[] = $state([]);
let orangeBots: string[] = $state([]);
let launcherOptionsVisible = $state(false);
</script>

<div class="page">
  <!-- <header>
    <button onclick={onBack}>
      <img src={arrow_left} alt="<-" class="leftArrowImage" />
    </button>
  </header> -->

  <div class="availableBots">
    <h2>Available bots</h2>
    <div class="availableBotsList">
      {#each Object.keys(botFamilies) as family, i}
        <div class="botEntry">
          {#if botFamilies[family].length == 1}
            <p>{botFamilies[family][0]}</p>
            <div class="expandMe"></div>
            <button
              class="addToTeam blue"
              onclick={() => {blueBots.push(botFamilies[family][0].split("(")[0].trim())}}
            >
              <Plus />
            </button>
            <button
              class="addToTeam orange"
              onclick={() => {orangeBots.push(botFamilies[family][0].split("(")[0].trim())}}
            >
              <Plus />
            </button>
          {:else}
            <p>{family}</p>
            <select
              name={botFamilies[family] + "-select"}
              id={botFamilies[family] + "-select"}
            >
              {#each botFamilies[family] as version, i}
                <option value={version.split("(")[0].trim()}
                  >{version}</option
                >
              {/each}
            </select>
            <div class="expandMe"></div>
            <button class="addToTeam blue" onclick={() => {
              blueBots.push((
                document.getElementById(botFamilies[family] + "-select"
              ) as any).value);
            }}>
              <Plus />
            </button>
            <button class="addToTeam orange" onclick={() => {
              orangeBots.push((
                document.getElementById(botFamilies[family] + "-select"
              ) as any).value);
            }}>
              <Plus />
            </button>
          {/if}
        </div>
      {/each}
    </div>
  </div>

  <div class="teams">
    <div class="blue">
      <h2>Blue</h2>
      <div class="botList">
        {#each blueBots as bot, i}
          <div class="bot">
            <p>{bot}</p>
            <button class="close" onclick={()=>{blueBots.splice(i, 1)}}>
              <img src={closeIcon} alt="X" />
            </button>
          </div>
        {/each}
        {#if blueBots.length == 0} <p>No bots selected for this team</p> {/if}
      </div>
    </div>
    <div class="orange">
      <h2>Orange</h2>
      <div class="botList">
        {#each orangeBots as bot, i}
          <div class="bot">
            <p>{bot}</p>
            <button class="close" onclick={()=>{orangeBots.splice(i, 1)}}>
              <img src={closeIcon} alt="X" />
            </button>
          </div>
        {/each}
        {#if orangeBots.length == 0} <p>No bots selected for this team</p> {/if}
      </div>
    </div>
  </div>

  <footer>
    <div class="options">
      <div>
        <label for="serverselect">Server</label>
        <select name="serverselect" id="serverselect" bind:value={serverAddr}>
          {#each servers as value, i}
            <option value={`${value.ip}:${value.port}`}>{value.location}</option>
          {/each}
        </select>
      </div>
      <div>
        <label for="mapselect">Map</label>
        <select name="mapselect" id="mapselect" bind:value={$mapStore}>
          {#each Object.entries(MAPS_STANDARD) as map, i}
            <option value={map[1]}>{map[0]}</option>
          {/each}
        </select>
      </div>
      <div>
        <label for="mapselect">Launcher</label>
        <LauncherSelector bind:visible={launcherOptionsVisible} />
      </div>
    </div>

    <div class="buttons">
      <button class="start" disabled={waiting} onclick={()=>{
        let launcher = localStorage.getItem("MS_LAUNCHER");
        if (!launcher) {
          toast.error("Please select a launcher first", {
            position: "top-center",
            duration: 5000,
          });

          launcherOptionsVisible = true;
          return;
        }

        waiting = true;
        let id = toast.loading("Starting rocket host game...", {
          position: "top-center"
        })
        App.StartRHostMatch({
          server: serverAddr,
          map: $mapStore,
          blueBots,
          orangeBots,
          launcher,
          launcherArg: localStorage.getItem("MS_LAUNCHER_ARG") || ''
        }).then((addr)=>{
          waiting = false;
          toast.success(
            `Started game with address ${addr}`,
            {position: "top-center", duration: 10000, id}
          )
        }).catch((e)=>{
          waiting = false;
          toast.error(
            "Failed to start Rocket Host game\n" + e,
            {position: "top-center", duration: 8000, id}
          )
        })
      }}>
        Start
      </button>
    </div>
  </footer>
</div>

<style>
  .page {
    display: flex;
    width: fit-content;
    height: 100%;
    max-width: 60rem;
    margin: 0px 3rem;
    flex-direction: column;
    align-items: center;
    padding: 1rem;
    gap: 1.5rem;
  }
  h2 {
    margin-bottom: 0.5rem;
  }
  footer {
    display: flex;
    width: 100%;
    justify-content: space-between;
  }
  .options {
    display: flex;
    gap: 1rem;
  }
  .options > div {
    display: flex;
    flex-direction: column;
  }
  .options select {
    font-size: 1.0rem;
    padding: 0.25rem;
  }
  .buttons {
    height: 100%;
    display: flex;
    align-items: end;
    gap: 0.5rem;
  }
  .buttons button {
    font-size: 1.2rem;
    height: fit-content;
  }
  button.start {
    background-color: #15680e;
  }
  .availableBots {
    display: flex;
    flex-direction: column;
  }
  .availableBotsList {
    display: inline-grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.4rem;
  }
  .expandMe {
    flex-grow: 1;
  }
  .botEntry {
    display: flex;
    align-items: center;
    padding: 0.1rem;
    background-color: var(--background-alt);
    user-select: none;
    -webkit-user-select: none;
    border-radius: 0.3rem;
  }
  .botEntry p {
    margin: 0px;
    margin-right: 1rem;
    margin-left: 0.5rem;
    font-size: 1.1rem;
  }
  .botEntry select {
    margin: 0px;
    margin-right: .6rem;
  }
  button.addToTeam {
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0px;
    padding: 0.1rem;
    border-radius: 50%;
    margin: 0px 0.2rem;
    height: 2rem;
    width: 2rem;
  }
  button.addToTeam.blue {
    background-color: #026df9aa;
  }
  button.addToTeam.orange {
    background-color: #f95402cc;
  }
  .teams {
    display: flex;
    width: 100%;
    justify-content: center;
    gap: 1rem;
  }
  .teams > div {
    display: flex;
    flex-direction: column;
    flex: 1;
  }
  .botList {
    display: flex;
    flex-direction: column;
    border-top: solid 3px;
    padding: 0.5rem;
    gap: 0.3rem;
  }
  .blue > .botList {
    border-color: #0054a6;
  }
  .orange > .botList {
    border-color: #f26522;
  }
  .bot {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--background-alt);
    width: 100%;
    padding: 0.3rem 0.5rem;
    border-radius: 0.3rem;
    font-size: 1.1rem;
  }
  button.close {
    padding: 0px;
  }
  button.close > img {
    filter: invert();
    height: 28px;
    width: 28px;
  }
</style>
