<script>
  import { onMount } from "svelte";

  export let selectedLogs;

  let textarea;

  onMount(() => {
    textarea.select();
    document.execCommand("copy");
  });
</script>

<textarea bind:value={selectedLogs} bind:this={textarea} />
