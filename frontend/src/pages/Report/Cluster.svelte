<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showClusterChart, resizeClusterChart,getClusterChartImage } from "./cluster";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let xField = "";
  let yField = "";
  let cluster = 2;
  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Cluster.DateTime'),
      width: "60%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: $_('Cluster.Item1'),
      width: "20%",
    },
    {
      name: $_('Cluster.Item2'),
      width: "20%",
    }
  ];

  let pagination = false;

  const updateCluster = async () => {
    if( xField == "" || yField == "" ){
      return;
    }
    await tick();
    showClusterChart("chart", logs,xField,yField,cluster * 1 || 2 ,dark);
    data = [];
    columns[1].name = getFieldName(xField) || xField;
    columns[2].name = getFieldName(yField) || yField;
    logs.forEach((l)=>{
      data.push([l.Time,l.KeyValue[xField] || "",l.KeyValue[yField] || "" ])
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
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Cluster.ExportType'),
      Title: $_('Cluster.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getClusterChartImage();
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
    window.go.main.App.Export(exportType,exportData).then(()=>{
      saveBusy = false;
      exportType = "";
    });
  }

  const onResize = () => {
    resizeClusterChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <!-- svelte-ignore a11y-no-onchange -->
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Cluster.Title')}</h3>
    <select
      class="form-select"
      aria-label="{$_('Cluster.XIterm')}"
      bind:value={xField}
      on:change="{updateCluster}"
    >
    <option value="">{$_('Cluster.SelectXItemMsg')}</option>
    {#each numFields as f}
      <option value={f}>{getFieldName(f)}</option>
    {/each}
    </select>
    <select
      class="form-select"
      aria-label="{$_('Cluster.YItem')}"
      bind:value={yField}
      on:change="{updateCluster}"
    >
      <option value="">{$_('Cluster.SelectYItemMsg')}</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select ml-2"
      aria-label="{$_('Cluster.NumberOfCluster')}"
      bind:value={cluster}
      on:change="{updateCluster}"
    >
      <option value="2">2</option>
      <option value="3">3</option>
      <option value="4">4</option>
      <option value="5">5</option>
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
        <span>{$_('Cluster.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('Cluster.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('Cluster.BackBtn')}
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 500px;
    margin: 5px auto;
  }
</style>
