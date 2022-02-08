<script>
  import { Plus16, Trash16 } from "svelte-octicons";
  import { createEventDispatcher,onMount, tick } from "svelte";
  export let conf;
  export let fields = [];
  const geoFields = [];
  fields.forEach((f) =>{
    if (f.includes("_geo_latlong")) {
      geoFields.push(f);
    }
  });
  let history = "";
  const dispatch = createEventDispatcher();

  const setHistory = () => {
    if (history){
      dispatch("update",{query: history ,add:false})
    }
  }

  const addKeyword = () => {
    const  q =  " " + conf.keyword.mode + conf.keyword.field + ":" + conf.keyword.key;
    dispatch("update",{query: q ,add:true})
  }

  const addRange  = () => {
    const q = ` time:>="` + conf.range.start + `:00Z09:00" time:<="` + conf.range.end + `:00Z09:00" `;
    dispatch("update",{query: q,add:true})
  }
  const addGeo = () =>{
    const q = " geo:" + conf.geo.field + "," + conf.geo.lat + "," + conf.geo.long + "," + conf.geo.range;
    dispatch("update",{query: q,add:true})
  }

</script>

<form>
  {#if conf.history.length > 0}
    <div class="mb-2">
      <label>
        検索履歴:
        <!-- svelte-ignore a11y-no-onchange -->
        <select style="width: 80%;" class="form-select" aria-label="履歴" bind:value={history} on:change="{setHistory}">
          <option value="">履歴を選択してください</option>
          {#each conf.history as h }
            <option value="{h}">{h}</option>
          {/each}
        </select>
      </label>
    </div>
  {/if}
  <div>
    <label>
      検索期間:
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder="開始"
        aria-label="開始"
        bind:value={conf.range.start}
      />
      -
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder="終了"
        aria-label="終了"
        bind:value={conf.range.end}
      />
    </label>
    <button class="btn" type="button" on:click={addRange}>
      <Plus16 />
    </button>
  </div>
  <div class="mt-2">
    <label>
      キーワード:
      <select
        class="form-select"
        aria-label="項目"
        bind:value={conf.keyword.field}
      >
        <option value="" >全体</option>
        {#each fields as f }
          <option value="{f}">{f}</option>
        {/each}
      </select>
      に
      <input
        class="form-control input-sm"
        type="text"
        style="width: 100px;"
        placeholder="キーワード"
        aria-label="キーワード"
        bind:value={conf.keyword.key}
      />
      が
      <select
        class="form-select"
        aria-label="有無"
        bind:value={conf.keyword.mode}
      >
        <option value="">含まれる</option>
        <option value="+">必須</option>
        <option value="-">含まれない</option>
      </select>
    </label>
    <button class="btn" type="button" on:click={addKeyword}>
      <Plus16 />
    </button>
  </div>
  {#if geoFields.length > 0 }
    <div class="mt-2">
      <label>
        IP位置情報:
        <select class="form-select" aria-label="項目" bind:value={conf.geo.field}>
          {#each geoFields as f }
            <option value="{f}">{f}</option>
          {/each}
        </select>
        が
        <input
          class="form-control input-sm"
          type="number"
          step="0.01"
          style="width: 80px;"
          placeholder="緯度"
          aria-label="緯度"
          bind:value={conf.geo.lat}
        />
        ,
        <input
          class="form-control input-sm"
          type="number"
          step="0.01"
          style="width: 80px;"
          placeholder="経度"
          aria-label="経度"
          bind:value={conf.geo.long}
        />
        から
        <input
          class="form-control input-sm"
          type="number"
          step="5"
          style="width: 80px;"
          placeholder="範囲"
          aria-label="範囲"
          bind:value={conf.geo.range}
        />
        Kmの範囲
      </label>
      <button class="btn" type="button" on:click={addGeo}>
        <Plus16 />
      </button>
    </div>
  {/if}
  <div class="mt-2">
    <label>
      最大件数:
      <select class="form-select" aria-label="最大件数" bind:value={conf.limit}>
        <option value="100">100</option>
        <option value="500">500</option>
        <option value="1000">1000</option>
        <option value="2000">2000</option>
        <option value="5000">5000</option>
        <option value="10000">10000</option>
        <option value="20000">20000</option>
      </select>
    </label>
  </div>
</form>

<style>
  select {
    min-width: 100px;
  }
</style>