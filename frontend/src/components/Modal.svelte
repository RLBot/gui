<script lang="ts">
import close from "../assets/close.svg";
let {
  title = "Modal",
  visible = $bindable(true),
  children,
}: {
  title?: string | (() => any);
  visible?: boolean;
  children?: () => any;
} = $props();

let wrap: EventTarget;

let mouseDownWasOutside = false;

function handleOuter(e: MouseEvent) {
  if (e.target === wrap && mouseDownWasOutside) visible = false;
}
function handleMouseDown(e: MouseEvent) {
  const res = e.target === wrap;
  mouseDownWasOutside = res;
  return res;
}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class={"modalContainer " + (visible ? "" : "hidden")} bind:this={wrap} onclick={handleOuter} onmousedown={handleMouseDown}>
  <div class="modal">
    <header>
      <h2>
        {#if typeof title === "string"}
          {title}
        {:else}
          {@render title()}
        {/if}
      </h2>
      <button
        onclick={() => {
          visible = false;
        }}><img src={close} alt="close" /></button
      >
    </header>
    <div class="modalBody">
      {@render children?.()}
    </div>
  </div>
</div>

<style>
  .modalContainer {
    display: flex;
    height: 100vh;
    width: 100vw;
    position: fixed;
    top: 0px;
    left: 0px;
    justify-content: center;
    align-items: center;
    z-index: 100;
    background-color: #00000060;

    opacity: 1;
    visibility: visible;
    transition: opacity 0.2s ease-in-out, visibility 0.2s ease-in-out;
  }

  .hidden, .hidden * {
    opacity: 0;
    visibility: hidden;
    transition: opacity 0.2s ease-in-out, visibility 0s linear 0.2s;
  }

  .modal {
    background-color: var(--background);
    padding: 0.2rem;
    border-radius: 0.6rem;
    min-width: 20vw;
    min-height: 20vh;
  }
  header {
    padding: 0.2rem;
    padding-left: 0.8rem;
    border-bottom: 1px solid var(--background-alt);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  header button {
    padding: 0px;
    background-color: transparent;
    filter: invert();
  }
  .modalBody {
    padding: 1rem;
    max-width: 90vw;
    max-height: 80vh;
    overflow-y: auto;
  }
</style>
