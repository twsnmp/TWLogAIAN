<script>
  import * as echarts from 'echarts'
  import { getFields, getFieldName } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showHistogramChart, resizeHistogramChart } from "./histogram";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let numFields = [];
  let selected = "";
  let dark = false;

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "日時",
      width: "80%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },
    {
      name: "項目1",
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
        limit: 10,
        enable: true,
      };
    }
  };

  onMount(() => {
    numFields = getFields(fields,"number");
    if( numFields.length > 0 ){
      selected = numFields[0];
      window.go.main.App.GetDark().then((v) => {
        dark = v;
        updateHistogram();
      });
    }
  });

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
    <h3 class="Box-title overflow-hidden flex-auto">ヒストグラム分析</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="項目"
      bind:value={selected}
      on:change={updateHistogram}
    >
    <option value="">集計する項目を選択してください</option>
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
    height: 350px;
    margin: 5px auto;
  }
</style>
