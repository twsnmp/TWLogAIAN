<script>
  import * as echarts from 'echarts'
  import { getFieldName,getFields } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showClusterChart, resizeClusterChart } from "../../js/AnCluster";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let field1 = "";
  let field2 = "";
  let cluster = 2;

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

  const updateCluster = async () => {
    if( field1 == "" || field2 == "" ){
      return;
    }
    await tick();
    showClusterChart("chart", logs,field1,field2,cluster * 1 || 2 ,dark);
    data = [];
    columns[1].name = getFieldName(field1) || "項目1";
    columns[2].name = getFieldName(field2) || "項目2";
    logs.forEach((l)=>{
      data.push([l.Time,l.KeyValue[field1] || "",l.KeyValue[field2] || "" ])
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
    <h3 class="Box-title overflow-hidden flex-auto">クラスター分析</h3>
    <select
      class="form-select"
      aria-label="分析項目1"
      bind:value={field1}
      on:change="{updateCluster}"
    >
    <option value="">項目1を選択して下さい</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select"
      aria-label="分析項目2"
      bind:value={field2}
      on:change="{updateCluster}"
    >
      <option value="">項目2を選択して下さい</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <select
      class="form-select ml-2"
      aria-label="クラスター数"
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
