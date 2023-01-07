<script>
  import Wellcome from "./pages/Wellcome.svelte";
  import Feedback from "./pages/Feedback.svelte";
  import WorkDir from "./pages/Setting/WorkDir.svelte";
  import Setting from "./pages/Setting/Setting.svelte";
  import Processing from "./pages/Search/Processing.svelte";
  import LogView from "./pages/Search/LogView.svelte";
  import { initI18n } from './i18n/i18n';

  let page = "wellcome";
  initI18n();
  const handleDone = (e) => {
    if (e && e.detail  && e.detail.page) {
      page = e.detail.page;
    }
  }
</script>

<main>
  <div id="page" style="--wails-draggable:no-drag">
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
    {:else if page == "logview"}
     <LogView on:done={handleDone}/>
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
