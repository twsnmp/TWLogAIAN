<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showTime3DChart, resizeTime3DChart } from "./time3d";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let catFields = [];
  let xType = "";
  let zType = "";
  let colorType = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "項目(X軸)",
      width: "50%",
    },
    {
      name: "日時",
      width: "30%",
      formatter: (cell) => echarts.time.format(cell, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: "値(Z軸）",
      width: "10%",
    },
    {
      name: "値(カラー）",
      width: "10%",
    },
  ];

  let pagination = false;

  const updateTime3DChart = async () => {
    if( xType == "" || zType == "" || colorType == "" ){
      return;
    }
    await tick();
    data = showTime3DChart("chart", logs, xType,zType,colorType,dark);
    if (data.length > 10) {
      pagination = {
        limit: 10,
        enable: true,
      };
    }
  };

  onMount(() => {
    catFields = getFields(fields,"string");
    numFields = getFields(fields,"number");
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

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
    <h3 class="Box-title overflow-hidden flex-auto">3D時系列分析</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="X軸"
      bind:value={xType}
      on:change="{updateTime3DChart}"
    >
      <option value="">X軸の項目を選択して下さい</option>
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="Z軸"
      bind:value={zType}
      on:change="{updateTime3DChart}"
    >
    <option value="">Z軸の項目を選択して下さい</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="カラー"
      bind:value={colorType}
      on:change="{updateTime3DChart}"
    >
      <option value="">色分けの項目を選択して下さい</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
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
    height: 500px;
    margin: 5px auto;
  }
</style>
