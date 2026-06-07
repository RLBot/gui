<script lang="ts">
import { Toaster } from "svelte-5-french-toast";
import AlarmIcon from "./assets/alarm.svg";
import CalendarPlusIcon from "./assets/calendar-plus.svg";
import logo from "./assets/rlbot_logo.svg";
import Events from "./components/Events.svelte";
import GuiSettings from "./components/GuiSettings.svelte";
import Home from "./pages/Home.svelte";
import RocketHost from "./pages/RocketHost.svelte";
import StateSettingSandbox from "./pages/StateSettingSandbox.svelte";
import StoryMode from "./pages/StoryMode.svelte";
import Welcome from "./pages/Welcome.svelte";
import { parseJSON } from "./index";
import arenaImages from "./arena-images";
import { currentTheme, THEMES } from "./settings";

const backgroundImage =
  arenaImages[Math.floor(Math.random() * arenaImages.length)];

let activePage = $state(
  localStorage.getItem("SHOW_WELCOME") !== "false" ? "welcome" : "home",
);

$effect(() => {
  if (activePage === "welcome") {
    localStorage.setItem("SHOW_WELCOME", "true");
  } else {
    localStorage.setItem("SHOW_WELCOME", "false");
  }
});

let eventsNow = $state(0);
let eventsFuture = $state(0);
let eventsVisible = $state(false);

let showGuiSettings = $state(false);

let mainClassString = $derived(
  (
    THEMES[$currentTheme] ??
    (window.matchMedia("(prefers-color-scheme: dark)").matches
      ? THEMES["Dark blurred"]
      : THEMES["Light"])
  ).join(" "),
);

let paths: {
  tagName: string | null;
  repo: string | null;
  installPath: string;
  visible: boolean;
  isDependency: boolean;
}[] = $state(parseJSON(window.localStorage.getItem("BOT_SEARCH_PATHS")) || []);
</script>

<Toaster />

<main
  class={mainClassString}
  style={`background-image: url("${backgroundImage}")`}
>
  <div class={"navbar blurred" + (activePage == "welcome" ? " offset" : "")}>
    <div>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_missing_attribute -->
      <a
        onclick={() => {
          activePage = "home";
        }}
      >
        <img class="logo" src={logo} alt="logo" />
        <h1>RLBot</h1>
      </a>

      {#if activePage == "rhost"}
        <h3>&nbsp; / Rocket Host</h3>
      {/if}
      {#if activePage == "statesettingsandbox"}
        <h3>&nbsp; / State Setting Sandbox</h3>
      {/if}
      {#if activePage == "storymode"}
        <h3>&nbsp; / Story Mode</h3>
      {/if}
    </div>
    <div class="navbuttons">
      <button id={eventsNow > 0 || eventsFuture > 0 ? "events" : ""} onclick={() => eventsVisible = true}>
        Events

        {#if eventsNow > 0}
        <span>
          <img src={AlarmIcon} alt="alarm" />
          {eventsNow}
        </span>
        {:else if eventsFuture > 0}
        <span>
          <img src={CalendarPlusIcon} alt="calendar" />
          {eventsFuture}
        </span>
        {/if}
      </button>
      <button onclick={() => {
        activePage = "storymode";
      }}
        >Story Mode</button
      >
      <button
        onclick={() => {
          activePage = "rhost";
        }}>Rocket Host</button
      >
      <div class="spacer"></div>
      <div class="dropdown">
        <button>Menu</button>
        <div class="dropmenu right">
          <button
            onclick={() => {
              activePage = "statesettingsandbox"
            }}
            >State Setting Sandbox</button
          >
          <button
            onclick={()=>{showGuiSettings = true}}
            >GUI Settings</button
          >
          <button
            onclick={()=>{
              activePage = "welcome";
              (document.activeElement as HTMLElement)?.blur()}
            }
            >Re-open setup screen</button
          >
        </div>
      </div>
    </div>
  </div>

  <div
    class={"pageContainer" + (activePage == "home" ? "" : " hidden")}
  >
    <Home bind:paths />
  </div>

  <div
    class={"pageContainer" + (activePage == "rhost" ? "" : " hidden")}
  >
    <RocketHost />
  </div>

  <div
    class={"pageContainer" + (activePage == "statesettingsandbox" ? "" : " hidden")}
  >
    <StateSettingSandbox visible={activePage === "statesettingsandbox"} />
  </div>

  <div
    class={"pageContainer" + (activePage == "storymode" ? "" : " hidden")}
  >
    <StoryMode />
  </div>

  <div
    class={"pageContainer" + (activePage == "welcome" ? "" : " hidden")}
  >
    <Welcome bind:paths closeMe={()=>{activePage = "home"}} />
  </div>
</main>

<Events bind:visible={eventsVisible} bind:eventsNow bind:eventsFuture />
<GuiSettings bind:visible={showGuiSettings} />

<style>
  main {
    display: flex;
    height: 100%;
    width: 100%;
    flex-direction: column;

    background-size: cover;
    background-repeat: no-repeat;
    background-position: center;
    background-attachment: fixed;
  }
  .navbar {
    display: flex;
    height: 3rem;
    justify-content: space-between;
    padding: 0.1rem;
    background: var(--background);
    color: var(--foreground);
    transition: translate 0.2s ease-in-out;
  }

  .navbar:has(.dropdown:focus-within) {
    /* We cannot set z-index in the child because it is implicitly set to
        0 here unless specified due to the backdrop-filter (i think?) */
    z-index: 1;
  }

  .navbar.offset {
    translate: 0 -100%;
  }
  .navbar > div {
    display: flex;
    align-items: center;
  }
  .navbar * {
    user-select: none;
    -webkit-user-select: none;
  }
  h1 {
    margin: 0px;
    margin-bottom: 0.2rem;
  }
  .logo {
    height: 3rem;
    margin-right: 0.2rem;
    padding: 0.3rem;
  }
  .navbuttons > * {
    margin: 0px 0.25rem;
  }
  .navbuttons > button {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .navbar .dropmenu {
    padding: 0.2rem;
  }
  .navbar .dropmenu > * {
    margin: 0.2rem;
  }
  a {
    cursor: pointer;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .pageContainer {
    display: flex;
    justify-content: center;
    height: 100%;
    width: 100%;
    background: inherit;
    visibility: visible;
    overflow: auto;
  }
  .hidden {
    opacity: 0;
    z-index: -99999;
    visibility: hidden;
    display: none;
  }
  #events {
    padding: 0.3rem 0.5rem;
  }
  #events span {
    align-items: center;
    vertical-align: middle;
    margin-left: 0.5rem;
    background-color: red;
    color: white;
    padding: 0.1rem 0.3rem 0 0.3rem;
    border-radius: 0.2rem;
  }
  #events img {
    filter: invert() brightness(var(--icon-brightness));
    width: 20px;
    vertical-align: middle;
    margin-bottom: 4px;
  }
</style>
