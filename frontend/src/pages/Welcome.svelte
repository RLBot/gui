<script lang="ts">
import AddBotpack from "../components/AddBotpack.svelte";
import LauncherSelector from "../components/LauncherSelector.svelte";
import logo from "../assets/rlbot_logo.svg";

let {
  paths = $bindable([]),
  closeMe,
}: {
  paths?: {
    tagName: string | null;
    repo: string | null;
    installPath: string;
    visible: boolean;
    isDependency: boolean;
  }[];
  closeMe?: () => void;
} = $props();

let addBotpackVisible = $state(false);
function openAddBotpackModal() {
  addBotpackVisible = true;
}
</script>

<div class="welcomeContainer">
  <div class="titleBox">
    <img src={logo} alt="logo">
    <h1>RLBot Setup</h1>
  </div>

  <div class="blurred">
    <div class="options">
      <p>Select your platform</p>
      <LauncherSelector />
      <br />
      <div style="display:flex;flex-direction:column;">
        <label for="wdlbotpack" style="color:#ff4444"> Recommended: </label>
        <button id="wdlbotpack" onclick={openAddBotpackModal}>Download the botpack</button>
      </div>
    </div>

    <button class="continue" onclick={closeMe}>Continue</button>
  </div>
</div>


<AddBotpack bind:visible={addBotpackVisible} bind:paths />

<style>
  .welcomeContainer {
    display: flex;
    align-items: center;
    justify-content: center;
    height: fit-content;
    margin: auto;
    border-radius: 1rem;
    flex-direction: column;
  }

  .blurred {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: rgba(0, 0, 0, 0.8);
    -webkit-backdrop-filter: blur(10px);
    backdrop-filter: blur(10px);
    padding: 2rem 3rem;
    border-radius: 1rem;
  }

  .titleBox {
    display: flex;
    align-items: center;
    padding-bottom: 1rem;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
  }
  .titleBox img {
    padding-right: .6rem;
    height: 2.8rem;
  }
  .titleBox h1 {
    font-size: 2rem;
  }

  #wdlbotpack {
    background-color: royalblue;
  }
  .continue {
    margin-left: auto;
    background-color: green;
  }

  .options > * {
    margin-bottom: 0.5rem;
  }
  .options {
    margin-bottom: 1rem;
  }
</style>
