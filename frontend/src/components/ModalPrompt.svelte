<script lang="ts">
import Modal from "./Modal.svelte";

let {
  title = "Modal",
  children,
}: {
  title?: string | (() => any);
  children?: () => any;
} = $props();

function createOneshot<T>(): {
  sender: (value: T) => void;
  recver: Promise<T>;
} {
  let resolve: (value: T) => void;
  const promise = new Promise((res, rej) => {
    resolve = res;
  });
  // @ts-ignore
  return { sender: resolve, recver: promise };
}

let { sender: cur_sender, recver: cur_recver } = createOneshot<boolean>();
let visible = $state(false);

function sender(val: boolean) {
  cur_sender(val);
  let new_oneshot = createOneshot<boolean>();
  cur_sender = new_oneshot.sender;
  cur_recver = new_oneshot.recver;
}

export async function prompt(): Promise<boolean> {
  visible = true;
  return cur_recver;
}
</script>

<Modal {title} bind:visible closeable={false}>
  {@render children?.()}
  <div class="buttons">
    <button class="submit" onclick={()=>{visible = false; sender(true)}}>Submit</button>
    <button onclick={()=>{visible = false; sender(false)}}>Close</button>
  </div>
</Modal>

<style>
  .buttons {
    display: flex;
    width: 100%;
    justify-content: right;
    gap: .5rem;
    padding-top: 1rem;
  }
  .submit {background-color: darkgreen;}
</style>
