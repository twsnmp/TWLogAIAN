<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showHeatmap, resizeHeatmap,getHeatmapImage } from "./heatmap";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';
  import {Export} from '../../../wailsjs/go/main/App';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  export let dark = false;
  let numFields = [];
  let field = "";
  let sumUnit = "day";
  let calcMode = "sum";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Heatmap.WeekDay'),
    },
    {
      name: $_('Heatmap.TimeRange'),
    },
    {
      name: $_('Heatmap.Value'),
    },
  ];

  let pagination = false;

  const updateHeatmap = async () => {
    await tick();
    data = showHeatmap("chart",logs,field,sumUnit,field ? calcMode : "sum",dark);
    columns[0].name = sumUnit == "day" ? $_('Heatmap.Date') : $_('Heatmap.WeekDay');
    columns[2].name = field ? getFieldName(field) : $_('Heatmap.Count');
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  }

  onMount(() => {
    numFields = getFields(fields,"number");
    updateHeatmap();
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Heatmap.ExportType'),
      Title: $_('Heatmap.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getHeatmapImage();
    }
    columns.forEach((e)=>{
      exportData.Header.push(e.name);
    });
    data.forEach((l)=>{
      const row = [];
      l.forEach((e,i)=>{
        const v = columns[i] && columns[i].formatter ? columns[i].formatter(e) : e;
        row.push(v);
      });
      exportData.Data.push(row);
    });
    await Export(exportType,exportData);
    saveBusy = false;
    exportType = "";
  }

  const onResize = () => {
    resizeHeatmap();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Heatmap.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      bind:value={field}
      on:change="{updateHeatmap}"
    >
      <option value="">{$_('Heatmap.CountItem')}</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      bind:value={sumUnit}
      on:change="{updateHeatmap}"
    >
      <option value="week">{$_('Heatmap.Weekly')}</option>
      <option value="day">{$_('Heatmap.Daily')}</option>
    </select>
    {#if field != "" }
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        bind:value={calcMode}
        on:change="{updateHeatmap}"
      >
        <option value="sum">{$_('Heatmap.Sum')}</option>
        <option value="mean">{$_('Heatmap.Mean')}</option>
        <option value="median">{$_('Heatmap.Median')}</option>
        <option value="variance">{$_('Heatmap.Variance')}</option>
      </select>
    {/if}
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={gridLang} />
  </div>
  <div class="Box-footer text-right">
    {#if data.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>{$_('Heatmap.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('Heatmap.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('Heatmap.BackBtn')}
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 450px;
    margin: 5px auto;
  }
</style>
