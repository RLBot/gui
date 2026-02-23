<script lang="ts">
import { flip } from "svelte/animate";
import { cubicInOut } from "svelte/easing";
import { Toaster } from "svelte-5-french-toast";
import AlarmIcon from "./assets/alarm.svg";
import CalendarPlusIcon from "./assets/calendar-plus.svg";
import logo from "./assets/rlbot_logo.svg";
import controlsCloseIcon from "./assets/window_controls/close.svg";
import controlsMaximizeIcon from "./assets/window_controls/maximize.svg";
import controlsMinimizeIcon from "./assets/window_controls/minimize.svg";
import Events from "./components/Events.svelte";
import GuiSettings from "./components/GuiSettings.svelte";
import Home from "./pages/Home.svelte";
import RocketHost from "./pages/RocketHost.svelte";
import StoryMode from "./pages/StoryMode.svelte";
import Welcome from "./pages/Welcome.svelte";
import { parseJSON } from "./index";
import arenaImages from "./arena-images";
import { currentTheme, showWindowControls, THEMES } from "./settings";
import { Window } from "@wailsio/runtime";

const backgroundImage =
  arenaImages[Math.floor(Math.random() * arenaImages.length)];

const pageMap = {
  home: { name: "Home", component: Home, hidden: false },
  welcome: { name: "Welcome", component: Welcome, hidden: true },
  rhost: { name: "Rocket Host", component: RocketHost, hidden: false },
  storymode: { name: "Story Mode", component: StoryMode, hidden: false },
} as const;

type page = keyof typeof pageMap;

let activePage: page = $state(
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

let roundedCorners = $state(false);
async function handleResize() {
  roundedCorners =
    $showWindowControls &&
    !(await Window.IsFullscreen()) &&
    !(await Window.IsMaximised());
}
$effect(() => {
  $showWindowControls;
  handleResize();
});
handleResize();

let borderStyle = $derived(roundedCorners ? "border-radius: .6rem;" : "");

$effect(() => {
  document.body.style = borderStyle;
});
</script>

<Toaster />

<svelte:window onresize={handleResize} onload={()=>{
  document.body.classList.add("loaded")
}} />

<main
  class={mainClassString}
  style={`background-image: url("${backgroundImage}");${borderStyle}`}
>
  <div class={"navbar blurred" + (activePage == "welcome" ? " offset" : "")}>
    <div>
      <img class="logo" src={logo} alt="logo" />
      <h1>RLBot</h1>
      <h3 style="margin: .5rem; opacity: 0.8;">/</h3>
      <div class="pageNav">
        {#each
          (Object.keys(pageMap) as page[])
            .filter(p => !pageMap[p].hidden)
            .sort((a, b) => b == activePage ? 1 : a ==activePage ? -1 : 0)
          as page (page)}
          <a
            href="#"
            onclick={() => activePage = page}
            animate:flip={{ duration: 200, easing: cubicInOut }}
          >
            <h3 class={activePage === page ? "active" : ""}>
              {pageMap[page].name}
            </h3>
          </a>
        {/each}
      </div>
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
      <div class="dropdown">
        <button>Menu</button>
        <div class="dropmenu right">
          <button
            onclick={alert.bind(null, "TODO: not implemented yet")}
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
      <div class={"controls" + ($showWindowControls ? "" : " hidden")}>
        <button onclick={Window.Minimise} class="minimise">
          <img src={controlsMinimizeIcon} alt="-">
        </button>
        <button onclick={Window.ToggleMaximise} class="maximise">
          <img src={controlsMaximizeIcon} alt="□">
        </button>
        <button onclick={Window.Close} class="close">
          <img src={controlsCloseIcon} alt="X">
        </button>
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
    class={"pageContainer" + (activePage == "storymode" ? "" : " hidden")}
  >
    <StoryMode />
  </div>

  <div
    class={"pageContainer" + (activePage == "welcome" ? " welcome" : " hidden")}
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
    --header-height: 2.5rem;

    background-size: cover;
    background-repeat: no-repeat;
    background-position: center;
    background-attachment: fixed;
  }
  .pageNav {
    display: flex;
    align-items: center;
    gap: 0.8rem;
  }
  .pageNav a {
    color: var(--foreground);
    text-decoration: none;
  }
  .pageNav h3 {
    cursor: pointer;
    transition: font-size .2s ease-in-out;
  }
  .pageNav h3:not(.active) {
    opacity: 0.8;
    font-size: 1rem;
  }
  .pageNav h3.active {
    font-size: 1.3rem;
  }
  .navbar {
    --wails-draggable: drag;
    display: flex;
    height: var(--header-height);
    justify-content: space-between;
    background: var(--background);
    --blur-radius: 4rem;
    --blur-alpha: 0.9;
    color: var(--foreground);
    transition: translate 0.2s ease-in-out;
  }
  .navbar > * > * {
    --wails-draggable: no-drag;
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
    margin-bottom: 0.1rem;
    font-size: calc(var(--header-height) * 2/3);
  }
  .logo {
    height: 100%;
    max-width: var(--header-height);
    margin-right: 0.2rem;
    padding: 0.3rem;
  }
  .navbuttons > * {
    margin: 0px 0.25rem;
  }
  .navbuttons > button,
  .navbuttons > .dropdown > button {
    display: flex;
    align-items: center;
    justify-content: center;
    height: calc(var(--header-height) * 0.8);
  }
  .navbar .dropmenu {
    padding: 0.2rem;
  }
  .navbar .dropmenu > * {
    margin: 0.2rem;
  }
  .navbar .controls {
    display: flex;
    align-items: start;
    height: 100%;
    margin: 0rem;
    margin-left: 0.5rem;
  }
  .navbar .controls button {
    display: flex;
    padding: calc(var(--header-height) * .3);
    height: 100%;
    width: var(--header-height);
    border-radius: 50%;
    background-color: transparent;
    backdrop-filter: none;
  }
  .navbar .controls button img {
    filter: brightness(var(--icon-brightness));
    width: 100%;
    height: auto;
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
  .pageContainer.welcome {
    --wails-draggable: drag;
  }
  .hidden, .hidden * {
    opacity: 0;
    z-index: -99999;
    visibility: hidden;
    display: none !important;
  }
  #events {
    padding: 0.3rem 0.5rem;
  }
  #events span {
    display: flex;
    margin-left: 0.5rem;
    background-color: red;
    padding: 0.15rem;
    padding-right: .3rem;
    gap: 0.2rem;
    align-items: center;
    border-radius: 0.2rem;
  }
  #events img {
    filter: invert() brightness(var(--icon-brightness));
    height: 100%;
    width: auto;
  }
</style>
