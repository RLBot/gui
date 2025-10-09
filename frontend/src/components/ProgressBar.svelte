<script lang="ts">
import { Events } from "@wailsio/runtime";

let {
  currentStep = $bindable(),
  percentComplete = $bindable(),
  totalSteps,
}: {
  currentStep: number;
  totalSteps: number;
  percentComplete: number;
} = $props();

Events.On("monitor:download-progress", (event) => {
  const { status, done } = event.data.at(-1);

  if (done) {
    percentComplete = 0;
    currentStep += 1;
  } else {
    percentComplete = status;
  }
});

let finalPercents = $derived.by(() => {
  let items = [];
  for (let i = 0; i < currentStep; i++) {
    items.push(100);
  }

  if (currentStep !== totalSteps) items.push(percentComplete);

  for (let i = currentStep + 1; i < totalSteps; i++) {
    items.push(0);
  }

  return items;
});
</script>

{#each finalPercents as finalPercent}
  <div id="progress" style="--p:{finalPercent}%;" role="progressbar">
    <span class="label">{Math.floor(finalPercent)}%</span>
  </div>
{/each}

<style>
  #progress {
    --p: 0%;
    width: 100%;
    height: 30px;
    position: relative;
    margin-top: 5%;
    background: linear-gradient(to right, #228B22, #006400);
    border: 1px solid #333;
    border-radius: 15px;
    overflow: hidden;
  }

  /* Overlay hides the not-yet-filled portion */
  #progress::after {
    content: "";
    position: absolute;
    top: 0;
    bottom: 0;
    left: var(--p);
    right: 0;
    background: #1a1a1a;
    transition: left 0.1s linear;
    z-index: 0;
  }

  .label {
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    font-size: 1rem;
    color: #fff;
    text-shadow: 0 0 2px rgba(0,0,0,0.6);
    pointer-events: none;
    user-select: none;
    z-index: 1;
  }
</style>
