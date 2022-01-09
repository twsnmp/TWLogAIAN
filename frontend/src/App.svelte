<script>
  let name = "";
  let greeting = "";

  function greet() {
    window.go.main.App.Greet(name).then((result) => {
      greeting = result;
    });
  }
  import Grid from "gridjs-svelte";
  const data = [
    { name: "John", email: "john@example.com" },
    { name: "Mark", email: "mark@gmail.com" },
  ];
</script>

<main>
  <div id="input" data-wails-no-drag>
    <input id="name" type="text" bind:value={name} />
    <button class="button" on:click={greet}>Greet</button>
  </div>
  {#if greeting}
    <div id="result">{greeting}</div>
  {/if}
  <Grid data={data} sort pagination search />  
</main>

<style>
  main {
    height: 100%;
    width: 100%;
    margin:40px auto;
  }

  #result {
    margin-top: 1rem;
    font-size: 1.5rem;
    color: azure;
  }

  button {
    -webkit-appearance: default-button;
    padding: 6px;
  }

  #name {
    border-radius: 3px;
    outline: none;
    -webkit-font-smoothing: antialiased;
  }
</style>
