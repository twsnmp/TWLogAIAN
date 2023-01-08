<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showTime3DChart, resizeTime3DChart,getTime3DChartImage } from "./time3d";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let catFields = [];
  let xField = "";
  let zField = "";
  let colorField = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Time3D.ItemX'),
      width: "50%",
    },
    {
      name: $_('Time3D.DateTime'),
      width: "30%",
      formatter: (cell) => echarts.time.format(cell, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: $_('Time3D.ItemZ'),
      width: "10%",
    },
    {
      name: $_('Time3D.ColorValue'),
      width: "10%",
    },
  ];

  let pagination = false;

  const updateTime3DChart = async () => {
    if( xField == "" || zField == "" || colorField == "" ){
      return;
    }
    await tick();
    data = showTime3DChart("chart", logs, xField,zField,colorField,dark);
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
    catFields = getFields(fields,"string");
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
      Type: $_('Time3D.ExportType'),
      Title: $_('Time3D.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getTime3DChartImage();
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
    resizeTime3DChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Time3D.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="{$_('Time3D.X')}"
      bind:value={xField}
      on:change="{updateTime3DChart}"
    >
      <option value="">{$_('Time3D.SelectXMsg')}</option>
      {#if catFields.length < 1 }
        <option value="_None">{$_('Time3D.NoItem')}</option>
      {/if}
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="{$_('Time3D.Z')}"
      bind:value={zField}
      on:change="{updateTime3DChart}"
    >
    <option value="">{$_('Time3D.SelectZMsg')}</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="{$_('Time3D.Color')}"
      bind:value={colorField}
      on:change="{updateTime3DChart}"
    >
      <option value="">{$_('Time3D.SelectColorMsg')}</option>
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
        <span>{$_('Time3D.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('Time3D.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('Time3D.BackBtn')}
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
