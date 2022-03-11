<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showTimeChart, resizeTimeChart,getTimeChartImage } from "./time";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let field = "";
  let chartType = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [];
  let pagination = false;

  const updateTime = async () => {
    if(field == "" ){
      return;
    }
    await tick();
    data = showTimeChart("chart", logs, field,chartType,dark);
    if(chartType == "1h" || chartType == "1m" ){
      columns = [
        {
          name: "日時",
          formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: "平均値",
        },{
          name: "最大値",
        },{
          name: "最小値",
        },{
          name: "中央値",
        },{
          name: "分散",
        }
      ];
    } else if (chartType != ""){
      columns = [
        {
          name: "日時",
          formatter: (cell) => echarts.time.format(new Date(cell), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: "回帰分析(" + chartType + ")",
        }
      ];
    } else {
      columns = [
        {
          name: "日時",
          formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
        },{
          name: getFieldName(field),
        }
      ];

    }
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
    if(field == "" && numFields.length > 0 ){
      field = numFields[0];
    }
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
      Type: "時系列分析",
      Title: "時系列分析",
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getTimeChartImage();
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
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeTimeChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <!-- svelte-ignore a11y-no-onchange -->
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">時系列分析</h3>
    <select
      class="form-select"
      bind:value={field}
      on:change="{updateTime}"
    >
      <option value="">項目を選択して下さい</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select ml-2"
      bind:value={chartType}
      on:change="{updateTime}"
    >
      <option value="">実データ</option>
      <option value="1m">分単位の集計</option>
      <option value="1h">時間単位の集計</option>
      <option value="linear">回帰分析(linear)</option>
      <option value="exponential">回帰分析(exponential)</option>
      <option value="logarithmic">回帰分析(logarithmic)</option>
      <option value="polynomial">回帰分析(polynomial)</option>
    </select>
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={jaJP} />
  </div>
  <div class="Box-footer text-right">
    {#if data.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>保存中</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">エクスポート</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      戻る
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 400px;
    margin: 5px auto;
  }
</style>
