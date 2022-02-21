<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showTimeChart, resizeTimeChart } from "./time";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let numField1 = "";
  let numField2 = "";
  let chartType = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "日時",
      width: "60%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: "項目1",
      width: "20%",
    },
    {
      name: "項目2",
      width: "20%",
    }
  ];

  let pagination = false;

  const updateTime = async () => {
    if( numField1 == "" ){
      return;
    }
    await tick();
    showTimeChart("chart", logs,numField1,numField2,chartType,dark);
    data = [];
    columns[1].name = getFieldName(numField1) || numField1;
    columns[2].name = getFieldName(numField2) || numField2;
    logs.forEach((l)=>{
      data.push([l.Time,l.KeyValue[numField1] || "",l.KeyValue[numField2] || "" ])
    });
    if (data.length > 10) {
      pagination = {
        limit: 10,
        enable: true,
      };
    }
  };

  onMount(() => {
    numFields = getFields(fields,"number");
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

  const onResize = () => {
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
      aria-label="数値項目1"
      bind:value={numField1}
      on:change="{updateTime}"
    >
      <option value="">１つ目の数値項目を選択して下さい</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select"
      aria-label="Y軸の項目"
      bind:value={numField2}
      on:change="{updateTime}"
    >
      <option value="">２つめの数値項目を選択して下さい</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select ml-2"
      aria-label="チャートの種類"
      bind:value={chartType}
      on:change="{updateTime}"
    >
      <option value="">ログの値</option>
      <option value="forcast">線形予測</option>
    </select>
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={jaJP} />
  </div>
  <div class="Box-footer text-right">
    <button class="btn  btn-secondary" type="button" on:click={back}>
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
