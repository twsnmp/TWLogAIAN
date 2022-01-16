<script>
  import { Router, Link, Route } from "svelte-routing";
  import Start from "./pages/Start.svelte";
  import DataSource from "./pages/DataSource.svelte";
  import Result from "./pages/Result.svelte";
  export let url = "/datasrc";
  let isOpenDB = false;

  const handleDone = () => {
    isOpenDB = true;
  }
  const close = () => {
    isOpenDB = false;
  }
</script>

<main>
  <div id="page" data-wails-no-drag>
    {#if !isOpenDB }
     <Start on:done={handleDone}/>
    {:else}
    <Router url="{url}">
      <nav class="UnderlineNav">
        <div class="UnderlineNav-body" role="tablist">
          <Link to="/datasrc" class="UnderlineNav-item">データソース</Link>
          <Link to="/datasrc1" class="UnderlineNav-item">インデックス</Link>
          <Link to="/datasrc2" class="UnderlineNav-item">分析設定</Link>
          <Link to="/result" class="UnderlineNav-item">結果</Link>
        </div>
        <div class="UnderlineNav-actions">
          <a href="##" class="btn btn-sm" on:click={close}>終了</a>
        </div>
        </nav>
      <div>
        <Route path="/datasrc"><DataSource /></Route>
        <Route path="/datasrc1"><DataSource /></Route>
        <Route path="/datasrc2"><DataSource /></Route>
        <Route path="/result"><Result /></Route>
      </div>
    </Router>
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
