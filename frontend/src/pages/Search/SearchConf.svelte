<script>
  import { getFieldName } from "../../js/define";

  export let conf;
  export let fields = [];
  export let extractorTypeList = [];
  import { _ } from '../../i18n/i18n';

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
  <div class="col-2 float-left">{$_('SearchConf.TimeRange')}</div>
  <div class="col-6 float-left">
    <select class="form-select" bind:value={conf.range.mode}>
      <option value="">{$_('SearchConf.NoSelect')}</option>
      <option value="target">{$_('SearchConf.TagetTime')}</option>
      <option value="range">{$_('SearchConf.StartEndTime')}</option>
    </select>
  </div>
  <div class="col-2 float-left" />
</div>

{#if conf.range.mode == "target"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_('SearchConf.TargetRange')}</div>
    <div class="col-6 float-left">
      <input
        class="form-control input-sm"
        type="text"
        style="width: 98%;"
        placeholder="{$_('SearchConf.TargetDateTime')}"
        aria-label="{$_('SearchConf.TargetDateTime')}"
        bind:value={conf.range.target}
      />
    </div>
    <div class="col-2 float-left">
      <input class="form-control input-sm" type="number" bind:value={conf.range.range}>Sec
    </div>
    <div class="col-2 float-left">
    </div>
  </div>
{/if}
{#if conf.range.mode == "range"}
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_('SearchConf.TimeRangeTitle')}</div>
    <div class="col-8 float-left">
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder="{$_('SearchConf.Start')}"
        aria-label="{$_('SearchConf.Start')}"
        bind:value={conf.range.start}
      />
      -
      <input
        class="form-control input-sm"
        type="datetime-local"
        placeholder="{$_('SearchConf.End')}"
        aria-label="{$_('SearchConf.End')}"
        bind:value={conf.range.end}
      />
    </div>
    <div class="col-2 float-left">
    </div>
  </div>
{/if}
{#if geoFields.length > 0  }
  <div class="container-lg clearfix">
    <div class="col-2 float-left">{$_('SearchConf.IPLocSearch')}</div>
    <div class="col-6 float-left">
      <select class="form-select" bind:value={conf.geo.mode}>
        <option value="">{$_('SearchConf.NoSearch')}</option>
        <option value="centor">{$_('SearchConf.DistFromCentor')}</option>
      </select>
    </div>
    <div class="col-2 float-left" />
  </div>
{/if}
{#if geoFields.length > 0 && conf.geo.mode != "" }
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">{$_('SearchConf.IPLoc')}</div>
    <div class="col-8 float-left">
      <select class="form-select" aria-label="{$_('SearchConf.Item')}" bind:value={conf.geo.field}>
        {#each geoFields as f}
          <option value={f}>{getFieldName(f)}</option>
        {/each}
      </select>
      {$_('SearchConf.Is')}
      <input
        class="form-control input-sm"
        type="number"
        step="0.01"
        style="width: 80px;"
        placeholder="{$_('SearchConf.Lat')}"
        aria-label="{$_('SearchConf.Lat')}"
        bind:value={conf.geo.lat}
      />
      ,
      <input
        class="form-control input-sm"
        type="number"
        step="0.01"
        style="width: 80px;"
        placeholder="{$_('SearchConf.Long')}"
        aria-label="{$_('SearchConf.Long')}"
        bind:value={conf.geo.long}
      />
      {$_('SearchConf.From')}
      <input
        class="form-control input-sm"
        type="number"
        step="5"
        style="width: 80px;"
        placeholder="{$_('SearchConf.Dist')}"
        aria-label="{$_('SearchConf.Dist')}"
        bind:value={conf.geo.range}
      />
      {$_('SearchConf.KMRange')}
    </div>
    <div class="col-2 float-left">
    </div>
  </div>
{/if}
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">{$_('SearchConf.Limit')}</div>
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
  <div class="col-2 float-left">{$_('SearchConf.AnomaryMode')}</div>
  <div class="col-3 float-left">
    <select class="form-select" bind:value={conf.anomaly}>
      <option value="">{$_('SearchConf.NotDetect')}</option>
      <option value="iforest">Isolation Forest</option>
      <option value="lof">Local Outlier Factor</option>
      <option value="autoencoder">Auto Encoder</option>
    </select>
  </div>
  {#if conf.anomaly}
    <div class="col-2 float-left">{$_('SearchConf.CalcVectorMode')}</div>
    <div class="col-3 float-left">
      <select class="form-select" bind:value={conf.vector}>
        <option value="">{$_('SearchConf.NumData')}</option>
        <option value="time">{$_('SearchConf.NumDataWeekDayH')}</option>
        <option value="all">{$_('SearchConf.StringNumData')}</option>
        <option value="alltime">{$_('SearchConf.StringNumWeekDayH')}</option>
        <option value="sql">{$_('SearchConf.SQLInjection')}</option>
        <option value="oscmd">{$_('SearchConf.OSCmdInjection')}</option>
        <option value="dirt">{$_('SearchConf.DirTra')}</option>
        <option value="walu">{$_('SearchConf.AccessLogWalu')}</option>
      </select>
    </div>
  {/if}
</div>
<div class="container-lg clearfix mt-1">
  <div class="col-2 float-left">{$_('SearchConf.ExtractOnSeach')}</div>
  <div class="col-3 float-left">
    <select class="form-select" bind:value={conf.extractor}>
      <option value="">{$_('SearchConf.NotUse')}</option>
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
