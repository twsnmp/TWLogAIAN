<script>
  import { getFieldName,getFieldType } from "../../js/define";
  import { Plus16,Trash16,File16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  export let conf;
  export let fields = [];
  let hasStringField = false;
  let hasNumberField = false;
  const geoFields = [];

  let history = "";
  const dispatch = createEventDispatcher();

  const setHistory = () => {
    if (history) {
      dispatch("update", { query: history, add: false });
    }
  };

  const addKeyword = () => {
    if( !conf.keyword.field || !conf.keyword.key){
      return
    }
    const q = conf.keyword.field == "" ? " " + conf.keyword.mode + conf.keyword.key
     : " " + conf.keyword.mode + conf.keyword.field + ":" + conf.keyword.key;
    dispatch("update", { query: q, add: true });
  };

  const addNumber = () => {
    if( !conf.number.field || !conf.number.oper || !conf.number.value){
      return
    }
    const oper = (conf.number.oper == "=") ? "=" : ":" + conf.number.oper
    const q = " " + conf.number.field + oper + conf.number.value;
    dispatch("update", { query: q, add: true });
  };

  const addRange = () => {
    let q = "";
    if (conf.range.start) {
      q += ` +time:>="` +
      conf.range.start +
      `:00+09:00"`;
    }
    if (conf.range.end) {
      q += ` +time:<="` +
      conf.range.end +
      `:00+09:00"`;
    }
    if (!q) {
      return
    }
    dispatch("update", { query: q, add: true });
  };
  const addGeo = () => {
    if( !conf.geo.field || !conf.geo.lat || !conf.geo.long || !conf.geo.range ){
      return
    }
    const q =
      " geo:" +
      conf.geo.field +
      "," +
      conf.geo.lat +
      "," +
      conf.geo.long +
      "," +
      conf.geo.range + "km";
    dispatch("update", { query: q, add: true });
  };

  fields.forEach((f) => {
    if (f.includes("_geo_latlong")) {
      geoFields.push(f);
      return
    }
    if (f.startsWith("_")) {
      return
    }
    if (getFieldType(f) == "string") {
      hasStringField = true;
    }
    if (getFieldType(f) == "number") {
      hasNumberField = true;
      if(conf.number.field=="" ){
        conf.number.field = f;
      }
    }
  });

  const clear = () => {
    conf.history.length = 0;
  }

  const loadKeyword = () => {
    window.go.main.App.LoadKeyword().then((r) => {
      if (r) {
        r.forEach((k)=>{
          const q =  conf.keyword.field == "" ? " " + conf.keyword.mode + k
           : " "  + conf.keyword.mode + conf.keyword.field + ":" + k;
          dispatch("update", { query: q, add: true });
        });
      }
    });
  }

</script>

<form>
  {#if conf.history.length > 0}
    <div class="container-lg clearfix mb-1">
      <div class="col-2 float-left">検索履歴</div>
      <div class="col-8 float-left">
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          style="width: 80%;"
          class="form-select"
          aria-label="履歴"
          bind:value={history}
          on:change={setHistory}
        >
          <option value="">履歴を選択してください</option>
          {#each conf.history as h}
            <option value={h}>{h}</option>
          {/each}
        </select>
      </div>
      <div class="col-2 float-left" />
        <button class="btn btn-danger" type="button" on:click={clear}>
          <Trash16 />
        </button>
      </div>
  {/if}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">検索期間</div>
    <div class="col-8 float-left">
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
    </div>
    <div class="col-2 float-left">
      <button class="btn" type="button" on:click={addRange}>
        <Plus16 />
      </button>
    </div>
  </div>
  {#if hasStringField}
    <div class="container-lg clearfix mt-1">
      <div class="col-2 float-left">キーワード</div>
      <div class="col-8 float-left">
        <select
          class="form-select"
          aria-label="項目"
          bind:value={conf.keyword.field}
        >
        <option value="">全体</option>
        {#each fields as f}
            {#if !f.startsWith("_") && getFieldType(f) == "string"}
              <option value={f}>{getFieldName(f)}</option>
            {/if}
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
        <select class="form-select" bind:value={conf.keyword.mode}>
          <option value="">含まれる</option>
          <option value="+">必須</option>
          <option value="-">含まれない</option>
        </select>
      </div>
      <div class="col-2 float-left">
        <button class="btn" type="button" on:click={addKeyword}>
          <Plus16 />
        </button>
        <button class="btn" type="button" on:click={loadKeyword}>
          <File16 />
        </button>
      </div>
    </div>
  {/if}
  {#if hasNumberField}
    <div class="container-lg clearfix mt-1">
      <div class="col-2 float-left">数値判定</div>
      <div class="col-8 float-left">
        <select
          class="form-select"
          aria-label="項目"
          bind:value={conf.number.field}
        >
          {#each fields as f}
            {#if !f.startsWith("_") && getFieldType(f) == "number"}
              <option value={f}>{getFieldName(f)}</option>
            {/if}
          {/each}
        </select>
        <select class="form-select" bind:value={conf.number.oper}>
          <option value="<">&lt;</option>
          <option value=">">&gt;</option>
          <option value="=">=</option>
          <option value="<=">&lt;=</option>
          <option value=">=">&gt;=</option>
        </select>
        <input
          class="form-control input-sm"
          type="text"
          style="width: 100px;"
          placeholder="数値"
          aria-label="数値"
          bind:value={conf.number.value}
        />
      </div>
      <div class="col-2 float-left">
        <button class="btn" type="button" on:click={addNumber}>
          <Plus16 />
        </button>
      </div>
    </div>
  {/if}
  {#if geoFields.length > 0}
    <div class="container-lg clearfix mt-1">
      <div class="col-2 float-left">IP位置情報</div>
      <div class="col-8 float-left">
        <select
          class="form-select"
          aria-label="項目"
          bind:value={conf.geo.field}
        >
          {#each geoFields as f}
            <option value={f}>{getFieldName(f)}</option>
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
      </div>
      <div class="col-2 float-left">
        <button class="btn" type="button" on:click={addGeo}>
          <Plus16 />
        </button>
      </div>
    </div>
  {/if}
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">最大件数</div>
    <div class="col-10 float-left">
      <select class="form-select" aria-label="最大件数" bind:value={conf.limit}>
        <option value="1000">1000</option>
        <option value="2000">2000</option>
        <option value="5000">5000</option>
        <option value="10000">10000</option>
        <option value="20000">20000</option>
        <option value="50000">50000</option>
        <option value="100000">100000</option>
        <option value="200000">200000</option>
      </select>
    </div>
  </div>
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">異常ログ検知方法</div>
    <div class="col-3 float-left">
      <select class="form-select" bind:value={conf.anomaly}>
        <option value="">検知しない</option>
        <option value="iforest">Isolation Forest</option>
        <option value="lof">Local Outlier Factor</option>
        <option value="autoencoder">Auto Encoder</option>
      </select>
    </div>
    {#if conf.anomaly}
    <div class="col-2 float-left">特徴量の計算方法</div>
    <div class="col-3 float-left">
      <select class="form-select" bind:value={conf.vector}>
        <option value="">数値データ</option>
        <option value="time">数値データ+曜日と時間帯</option>
        <option value="all">文字列と数値データ</option>
        <option value="alltime">文字列と数値データ+曜日と時間帯</option>
        <option value="sql">SQLインジェクション</option>
        <option value="oscmd">OSコマンドインジェクション</option>
        <option value="dirt">ディレクトリトラバーサル</option>
        <option value="walu">アクセスログ(Waluの方法)</option>
      </select>
    </div>
    {/if}
  </div>
</form>

<style>
  select {
    min-width: 100px;
  }
</style>
