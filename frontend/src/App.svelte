<script>
  import Wellcome from "./pages/Wellcome.svelte";
  import Feedback from "./pages/Feedback.svelte";
  import WorkDir from "./pages/WorkDir.svelte";
  // import Config from "./pages/Config.svelte";
  // import Search from "./pages/Search.svelte";
  // import Result from "./pages/Result.svelte";
  let page = "wellcome";
  const handleDone = (e) => {
    if (e && e.detail  && e.detail.page) {
      page = e.detail.page;
    }
  }
</script>

<main>
  <div id="page" data-wails-no-drag>
    {#if page == "wellcome" }
     <Wellcome on:done={handleDone}/>
    {:else if page == "feedback"}
     <Feedback on:done={handleDone}/>
    {:else if page == "workdir"}
     <WorkDir on:done={handleDone}/>
    {/if}
  </div>
</main>

<style>
  main {
    height: 100%;
    width: 100%;
  }
  #page {
    height: 100%;
    width: 95%;
    margin:10px auto;
  }
</style>
