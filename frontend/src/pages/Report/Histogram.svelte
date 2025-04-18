<script>
  import * as echarts from 'echarts'
  import { getFields, getFieldName,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showHistogramChart, resizeHistogramChart,getHistogramChartImage } from "./histogram";
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
  let selected = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Histogram.DateTime'),
      width: "80%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: $_('Histogram.Item1'),
      width: "20%",
    },
  ];
  let pagination = false;

  const updateHistogram = async () => {
    if (selected == "" ) {
      return;
    }
    await tick();
    showHistogramChart("chart", logs, selected,dark);
    data = [];
    columns[1].name = getFieldName(selected) || selected;
    logs.forEach((l)=> {
      const v = l.KeyValue[selected];
      switch (typeof v) {
      case "string":
      case "number":
      case "boolean":
        break;
      default:
        return;
      }
      data.push([l.Time,l.KeyValue[selected] || "" ])
    });
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  };

  onMount(() => {
    numFields = getFields(fields,"number");
    if( numFields.length > 0 ){
      selected = numFields[0];
      updateHistogram();
    }
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Histogram.ExportType'),
      Title: $_('Histogram.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getHistogramChartImage();
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
    resizeHistogramChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Histogram.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      bind:value={selected}
      on:change={updateHistogram}
    >
    <option value="">{$_('Histogram.SelectItemMsg')}</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
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
        <span>{$_('Histogram.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('Histogram.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('Histogram.BackBtn')}
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 350px;
    margin: 5px auto;
  }
</style>
