<script>
let {
  checked = $bindable(),
  width = 48,
  height = 28,
  stopClickPropagation = false,
  onchange = () => {},
} = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<label class="switch" style="width: {width}px; height: {height}px;" onclick={e => stopClickPropagation && e.stopPropagation()}>
  <input type="checkbox" bind:checked onchange={() => onchange()}/>
  <span class="slider" style="border-radius: {height / 2}px;"></span>
</label>

<style>
  .switch {
    position: relative;
    display: inline-block;
  }
  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: #ccc;
    transition: 0.1s;
    border-radius: 34px;
  }
  .slider:before {
    position: absolute;
    content: "";
    height: calc(100% - 8px);
    width: calc(50% - 4px);
    left: 4px;
    bottom: 4px;
    background-color: white;
    transition: 0.1s;
    border-radius: 50%;
  }
  input:checked + .slider {
    background-color: #2196f3;
  }
  input:checked + .slider {
    box-shadow: 0 0 1px #2196f3;
  }
  input:checked + .slider:before {
    transform: translateX(100%);
  }
</style>
