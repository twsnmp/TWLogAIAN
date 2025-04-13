<script>
  import { getFieldName, getFieldType } from "../../js/define";
  import { Plus16, Trash16, File16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { _ } from "../../i18n/i18n";
  import { LoadKeyword } from "../../../wailsjs/go/main/App";
  import AutoComplete from "simple-svelte-autocomplete";

  export let conf;
  export let fields = [];
  export let extractorTypeList = [];
  const geoFields = [];

  const setSelectedExtractor =  () => {
    for (const e of extractorTypeList) {
      if (e.Key == conf.extractor) {
      return e;
      }
    }
    return extractorTypeList[0];
  }

  let hasStringField = false;
  let hasNumberField = false;
  let selectedExtractor = setSelectedExtractor();


  let history = "";
  const dispatch = createEventDispatcher();

  const setHistory = () => {
    if (history) {
      dispatch("update", { query: history, add: false });
    }
  };

  const addKeyword = () => {
    if (!conf.keyword.key) {
      return;
    }
    const q =
      conf.keyword.field == ""
        ? " " + conf.keyword.mode + conf.keyword.key
        : " " + conf.keyword.mode + conf.keyword.field + ":" + conf.keyword.key;
    dispatch("update", { query: q, add: true });
  };

  const addNumber = () => {
    if (!conf.number.field || !conf.number.oper || !conf.number.value) {
      return;
    }
    const oper = conf.number.oper == "=" ? ":" : ":" + conf.number.oper;
    const q = " " + conf.number.field + oper + conf.number.value;
    dispatch("update", { query: q, add: true });
  };

  fields.forEach((f) => {
    if (f.includes("_geo_latlong")) {
      geoFields.push(f);
      return;
    }
    if (f.startsWith("_")) {
      return;
    }
    if (getFieldType(f) == "string") {
      hasStringField = true;
    }
    if (getFieldType(f) == "number") {
      hasNumberField = true;
      if (conf.number.field == "") {
        conf.number.field = f;
      }
    }
  });

  const clear = () => {
    conf.history.length = 0;
  };

  const loadKeyword = async () => {
    const r = await LoadKeyword();
    if (r) {
      r.forEach((k) => {
        const q =
          conf.keyword.field == ""
            ? " " + conf.keyword.mode + k
            : " " + conf.keyword.mode + conf.keyword.field + ":" + k;
        dispatch("update", { query: q, add: true });
      });
    }
  };

</script>

{#if conf.history.length > 0}
  <div class="container-lg clearfix mb-1">
    <div class="col-2 float-left">{$_("SearchConf.History")}</div>
    <div class="col-8 float-left">
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        style="width: 80%;"
        class="form-select"
        bind:value={history}
        on:change={setHistory}
      >
        <option value="">{$_("SearchConf.SelectHistoryMsg")}</option>
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
  <div class="col-2 float-left">{$_("SearchConf.QueryMode")}</div>
  <div class="col-6 float-left">
    <select class="form-select" bind:value={conf.mode}>
      <option value="simple">{$_("SearchConf.Simple")}</option>
      <option value="regexp">{$_("SearchConf.Regexp")}</option>
      <option value="full">{$_("SearchConf.FullTextSearch")}</option>
    </select>
  </div>
  <div class="col-2 float-left" />
</div>

<div class="container-lg clearfix">
  <div class="col-2 float-left">{$_("SearchConf.TimeRange")}</div>
  <div class="col-6 float-left">
    <select class="form-select" bind:value={conf.range.mode}>
      <option value="">{$_("SearchConf.NoSelect")}</option>
      <option value="target">{$_("SearchConf.TagetTime")}</option>
      <option value="range">{$_("SearchConf.StartEndTime")}</option>
    </select>
  </div>
  <div class="col-2 float-left" />
</div>

{#if conf.range.mode == "target"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_("SearchConf.TargetRange")}</div>
    <div class="col-6 float-left">
      <input
        class="form-control input-sm"
        type="text"
        style="width: 98%;"
        placeholder={$_("SearchConf.TargetDateTime")}
        bind:value={conf.range.target}
      />
    </div>
    <div class="col-2 float-left">
      <input
        class="form-control input-sm"
        type="number"
        bind:value={conf.range.range}
      />
    </div>
    <div class="col-2 float-left">
      {$_("SearchConf.Sec")}
    </div>
  </div>
{/if}
{#if conf.range.mode == "range"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_("SearchConf.TimeRangeTitle")}</div>
    <div class="col-8 float-left">
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder={$_("SearchConf.Start")}
        bind:value={conf.range.start}
      />
      -
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder={$_("SearchConf.End")}
        bind:value={conf.range.end}
      />
    </div>
    <div class="col-2 float-left" />
  </div>
{/if}
{#if geoFields.length > 0}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_("SearchConf.IPLocSearch")}</div>
    <div class="col-6 float-left">
      <select class="form-select" bind:value={conf.geo.mode}>
        <option value="">{$_("SearchConf.NoSearch")}</option>
        <option value="centor">{$_("SearchConf.DistFromCentor")}</option>
      </select>
    </div>
    <div class="col-2 float-left" />
  </div>
{/if}
{#if geoFields.length > 0 && conf.geo.mode != ""}
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">{$_("SearchConf.IPLoc")}</div>
    <div class="col-8 float-left">
      <select
        class="form-select"
        bind:value={conf.geo.field}
      >
        {#each geoFields as f}
          <option value={f}>{getFieldName(f)}</option>
        {/each}
      </select>
      {$_("SearchConf.Is")}
      <input
        class="form-control input-sm"
        type="number"
        step="0.01"
        style="width: 80px;"
        placeholder={$_("SearchConf.Lat")}
        bind:value={conf.geo.lat}
      />
      ,
      <input
        class="form-control input-sm"
        type="number"
        step="0.01"
        style="width: 80px;"
        placeholder={$_("SearchConf.Long")}
        bind:value={conf.geo.long}
      />
      {$_("SearchConf.From")}
      <input
        class="form-control input-sm"
        type="number"
        step="5"
        style="width: 80px;"
        placeholder={$_("SearchConf.Dist")}
        bind:value={conf.geo.range}
      />
      {$_("SearchConf.KMRange")}
    </div>
    <div class="col-2 float-left" />
  </div>
{/if}
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">{$_("SearchConf.Limit")}</div>
  <div class="col-10 float-left">
    <select class="form-select" bind:value={conf.limit}>
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
  <div class="col-2 float-left">{$_("SearchConf.AnomaryMode")}</div>
  <div class="col-3 float-left">
    <select class="form-select" bind:value={conf.anomaly}>
      <option value="">{$_("SearchConf.NotDetect")}</option>
      <option value="iforest">Isolation Forest</option>
{#if conf.limit <= 10000}
      <option value="lof">Local Outlier Factor</option>
      <option value="autoencoder">Auto Encoder</option>
{/if}
      <option value="sum">Sum</option>
    </select>
  </div>
  {#if conf.anomaly}
    <div class="col-2 float-left">{$_("SearchConf.CalcVectorMode")}</div>
    <div class="col-3 float-left">
      <select class="form-select" bind:value={conf.vector}>
        <option value="">{$_("SearchConf.NumData")}</option>
        <option value="time">{$_("SearchConf.NumDataWeekDayH")}</option>
        <option value="all">{$_("SearchConf.StringNumData")}</option>
        <option value="alltime">{$_("SearchConf.StringNumWeekDayH")}</option>
        <option value="sql">{$_("SearchConf.SQLInjection")}</option>
        <option value="oscmd">{$_("SearchConf.OSCmdInjection")}</option>
        <option value="dirt">{$_("SearchConf.DirTra")}</option>
        <option value="walu">{$_("SearchConf.AccessLogWalu")}</option>
        <option value="tfidf">TF-IDF</option>
      </select>
    </div>
  {/if}
</div>
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">{$_("SearchConf.ExtractOnSeach")}</div>
  <div class="col-3 float-left">
    <AutoComplete
      items="{extractorTypeList}"
      bind:value ="{conf.extractor}"
      labelFieldName="Name"
      valueFieldName="Key"
      inputClassName="form-control"
      bind:selectedItem="{selectedExtractor}"
      />
  </div>
</div>
<div class="container-lg clearfix">
  <div class="col-2 float-left">{$_("SearchConf.HighlighMode")}</div>
  <div class="col-6 float-left">
    <select class="form-select" bind:value={conf.highlightMode}>
      <option value="">{$_("SearchConf.None")}</option>
      <option value="keyword">{$_("SearchConf.HighlighKeyword")}</option>
      <option value="code">{$_("SearchConf.HighlighCode")}</option>
    </select>
  </div>
  <div class="col-2 float-left" />
</div>
{#if conf.mode == "full"}
  {#if hasStringField}
    <div class="container-lg clearfix mt-1">
      <div class="col-2 float-left">{$_("SearchConf.Keyword")}</div>
      <div class="col-8 float-left">
        <select
          class="form-select"
          bind:value={conf.keyword.field}
        >
          <option value="">{$_("SearchConf.All")}</option>
          {#each fields as f}
            {#if !f.startsWith("_") && getFieldType(f) == "string"}
              <option value={f}>{getFieldName(f)}</option>
            {/if}
          {/each}
        </select>
        {$_("SearchConf.Include")}
        <input
          class="form-control input-sm"
          type="text"
          style="width: 150px;"
          placeholder={$_("SearchConf.Keyword")}
          bind:value={conf.keyword.key}
        />
        {$_("SearchConf.Is")}
        <select class="form-select" bind:value={conf.keyword.mode}>
          <option value="+">{$_("SearchConf.Must")}</option>
          <option value="-">{$_("SearchConf.Exclude")}</option>
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
      <div class="col-2 float-left">{$_("SearchConf.NumberCond")}</div>
      <div class="col-8 float-left">
        <select
          class="form-select"
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
          placeholder="{$_('SearchConf.Number')}"
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
{/if}

<style>
  select {
    min-width: 100px;
  }
</style>
