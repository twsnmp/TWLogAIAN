<script>
  import {
    X16,
    Check16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  export let conf;
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let limit = conf.limit;
  const cancel = () => {
    dispatch("done", {});
  };

  const save = () => {
    conf.limit = limit;
    conf.query = "";
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
  };

</script>

<div class="Box-header">
  <h3 class="Box-title">検索条件</h3>
</div>
{#if errorMsg != ""}
  <div class="flash flash-error">
    {errorMsg}
    <button
      class="flash-close js-flash-close"
      type="button"
      aria-label="Close"
      on:click={clearMsg}
    >
      <X16 />
    </button>
  </div>
{/if}
<div class="Box-body">
  <form>
    <div class="form-group">
      <div class="form-group-header">
        <h5>最大件数</h5>
      </div>
      <div class="form-group-body">
        <select
          class="form-select"
          aria-label="最大件数"
          bind:value={limit}
        >
          <option value="100">100</option>
          <option value="500">500</option>
          <option value="1000">1000</option>
          <option value="2000">2000</option>
          <option value="5000">5000</option>
          <option value="10000">10000</option>
          <option value="20000">20000</option>
        </select>
      </div>
    </div>
  </form>
</div>
<div class="Box-footer text-right">
  <button class="btn btn-secondary" type="button" on:click={cancel}>
    <X16 />
    キャンセル
  </button>
  <button class="btn btn-primary ml-2" type="button" on:click={save}>
    <Check16 />
      適用
  </button>
</div>
