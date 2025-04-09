<script lang="ts">
import { Toaster } from "svelte-5-french-toast";
import AlarmIcon from "./assets/alarm.svg";
import CalendarPlusIcon from "./assets/calendar-plus.svg";
import logo from "./assets/rlbot_logo.svg";
import Events from "./components/Events.svelte";
import Home from "./pages/Home.svelte";
import RocketHost from "./pages/RocketHost.svelte";

let activePage = $state("home");

let eventsNow = $state(0);
let eventsFuture = $state(0);
let eventsVisible = $state(false);
</script>

<Toaster />

<main>
  <div class="navbar">
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
      <button onclick={alert.bind(null, "TODO: not implemented yet")}
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
            onclick={alert.bind(null, "TODO: not implemented yet")}
            >State Setting Sandbox</button
          >
        </div>
      </div>
    </div>
  </div>

  <div
    class={activePage == "home" ? "pageContainer" : "pageContainer hidden"}
  >
    <Home />
  </div>

  <div
    class={activePage == "rhost" ? "pageContainer" : "pageContainer hidden"}
  >
    <RocketHost />
  </div>
</main>

<Events bind:visible={eventsVisible} bind:eventsNow bind:eventsFuture />

<style>
  main {
    display: flex;
    height: 100%;
    width: 100%;
    flex-direction: column;
  }
  .navbar {
    display: flex;
    height: 3rem;
    justify-content: space-between;
    padding: 0.1rem;
    /* background: var(--background-alt);
    color: var(--foreground-alt); */
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
    overflow: scroll;
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
    filter: invert() brightness(90%);
    width: 20px;
    vertical-align: middle;
    margin-bottom: 4px;
  }
</style>
