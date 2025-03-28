<script lang="ts">
import UpdateIcon from "../assets/update.svg";

let {
  repoName = "",
  updateBotpack = () => {},
}: {
  repoName: string;
  updateBotpack: (repoName: string) => void;
} = $props();
</script>

<span>{repoName} has an update!</span>
<button id="update" onclick={() => updateBotpack(repoName)}><img src={UpdateIcon} alt="Update" /> Update</button>

<style>
  img {
    filter: invert();
    width: 24px;
    height: 24px;
    margin-right: 0.25em;
  }
  button#update {
    display: inline-flex;
    background-color: #4caf50;
    color: white;
    padding: 0.5em;
    border: none;
    border-radius: 0.25em;
    align-items: center;
    margin-right: 0.25em;
  }
  button#update:hover {
    background-color: #388e3c;
  }
  button#update:active {
    background-color: #2e7d32;
  }
</style>
