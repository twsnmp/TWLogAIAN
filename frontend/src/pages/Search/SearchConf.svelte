<script>
  import { getFieldName } from "../../js/define";

  export let conf;
  export let fields = [];
  export let extractorTypeList = [];

  const geoFields = [];

  fields.forEach((f) => {
    if (f.includes("_geo_latlong")) {
      geoFields.push(f);
      return;
    }
    if (f.startsWith("_")) {
      return;
    }
  });

</script>

<div class="container-lg clearfix">
  <div class="col-2 float-left">検索期間</div>
  <div class="col-6 float-left">
    <select class="form-select" bind:value={conf.range.mode}>
      <option value="">指定しない</option>
      <option value="target">対象指定</option>
      <option value="range">範囲指定</option>
    </select>
  </div>
  <div class="col-2 float-left" />
</div>

{#if conf.range.mode == "target"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">期間(対象)</div>
    <div class="col-6 float-left">
      <input
        class="form-control input-sm"
        type="text"
        style="width: 98%;"
        placeholder="対象の日時"
        aria-label="対象の日時"
        bind:value={conf.range.target}
      />
    </div>
    <div class="col-2 float-left">
      <select class="form-select" bind:value={conf.range.range}>
        <option value="-5">まで5秒間</option>
        <option value="-10">まで10秒間</option>
        <option value="-30">まで30秒間</option>
        <option value="-60">まで1分間</option>
        <option value="-1800">まで30分間</option>
        <option value="-3600">まで1時間</option>
        <option value="5">から5秒間</option>
        <option value="10">から10秒間</option>
        <option value="30">から30秒間</option>
        <option value="60">から1分間</option>
        <option value="1800">から30分間</option>
        <option value="3600">から1時間</option>
      </select>
    </div>
    <div class="col-2 float-left">
    </div>
  </div>
{/if}
{#if conf.range.mode == "range"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">期間(範囲)</div>
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
    </div>
  </div>
{/if}
{#if geoFields.length > 0  }
  <div class="container-lg clearfix">
    <div class="col-2 float-left">IP位置情報検索</div>
    <div class="col-6 float-left">
      <select class="form-select" bind:value={conf.geo.mode}>
        <option value="">検索しない</option>
        <option value="centor">中心からの範囲</option>
      </select>
    </div>
    <div class="col-2 float-left" />
  </div>
{/if}
{#if geoFields.length > 0 && conf.geo.mode != "" }
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">IP位置情報</div>
    <div class="col-8 float-left">
      <select class="form-select" aria-label="項目" bind:value={conf.geo.field}>
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
    </div>
  </div>
{/if}
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">最大件数</div>
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
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">検索時データ抽出</div>
  <div class="col-3 float-left">
    <select class="form-select" bind:value={conf.extractor}>
      <option value="">しない</option>
      {#each extractorTypeList as { Key, Name }}
        <option value={Key}>{Name}</option>
      {/each}
    </select>
  </div>
</div>

<style>
  select {
    min-width: 100px;
  }
</style>
