<script>
  import Wellcome from "./pages/Wellcome.svelte";
  import Feedback from "./pages/Feedback.svelte";
  import WorkDir from "./pages/Setting/WorkDir.svelte";
  import Setting from "./pages/Setting/Setting.svelte";
  import Processing from "./pages/Search/Processing.svelte";
  import TimeOnly from "./pages/Search/TimeOnly.svelte";
  import Syslog   from "./pages/Search/Syslog.svelte";
  import AccessLog   from "./pages/Search/AccessLog.svelte";
  let page = "wellcome";
  let back = "";
  const handleDone = (e) => {
    if (e && e.detail  && e.detail.page) {
      back = page;
      page = e.detail.page;
    }
  }
  const handleBack = (e) => {
    page = back;
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
    {:else if page == "setting"}
     <Setting on:done={handleDone}/>
    {:else if page == "processing"}
     <Processing on:done={handleDone}/>
    {:else if page == "timeonly"}
     <TimeOnly on:done={handleDone}/>
    {:else if page == "syslog"}
     <Syslog on:done={handleDone}/>
    {:else if page == "access"}
     <AccessLog on:done={handleDone}/>
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
