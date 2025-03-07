<script lang="ts">
import Home from "./pages/Home.svelte";
import RocketHost from "./pages/RocketHost.svelte";
import logo from "./assets/rlbot_logo.svg";
import { Toaster } from "svelte-5-french-toast";

let activePage = $state("home");
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
      <button onclick={alert.bind(null, "TODO: not implemented yet")}
        >Events</button
      >
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
</style>
