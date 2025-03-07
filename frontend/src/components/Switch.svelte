<script>
let { checked = $bindable() } = $props();
</script>

<label class="switch">
  <input type="checkbox" bind:checked />
  <span class="slider"></span>
</label>

<style>
  .switch {
    position: relative;
    display: inline-block;
    width: 48px;
    height: 28px;
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
    transition: 0.4s;
    border-radius: 34px;
  }
  
  .slider:before {
    position: absolute;
    content: "";
    height: 20px;
    width: 20px;
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
    transform: translateX(20px);
  }
</style>
