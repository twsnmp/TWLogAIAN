<script>
  import { onMount } from "svelte";

  export let copyText;

  let textarea;

  onMount(() => {
    textarea.select();
    document.execCommand("copy");
  });
</script>

<textarea bind:value={copyText} bind:this={textarea} />
