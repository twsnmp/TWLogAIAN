<script>
  import * as echarts from "echarts";
  import { getFieldName, getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { getFFTMap, showFFTChart, resizeFFTChart } from "./fft";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let catFields = [];
  let keyList = [];
  let selected = "";
  let key = "";
  let fftType = "hz";
  let fftMap;

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "日時",
      width: "80%",
      formatter: (cell) =>
        echarts.time.format(
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
        ),
    },
    {
      name: "項目1",
      width: "20%",
    },
  ];

  let pagination = false;

  const updateFFT = () => {
    if (selected == "") {
      return;
    }
    fftMap = getFFTMap(logs, selected);
    keyList = [];
    for (let key of fftMap.keys()) {
      keyList.push(key);
    }
    data = [];
    columns[1].name = getFieldName(selected) || selected;
    logs.forEach((l) => {
      const v = l.KeyValue[selected];
      switch (typeof v) {
        case "string":
        case "number":
        case "boolean":
          break;
        default:
          return;
      }
      data.push([l.Time, l.KeyValue[selected] || ""]);
    });
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
    updateFFTChart();
  };

  const updateFFTChart = async () => {
    if (!fftMap) {
      return;
    }
    await tick();
    showFFTChart("chart", selected, fftMap, key, fftType, dark);
  };

  onMount(() => {
    catFields = getFields(fields, "string");
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

  const onResize = () => {
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeFFTChart();
  };

  const back = () => {
    dispatch("done", {});
  };
</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">FFT分析</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="集計項目"
      bind:value={selected}
      on:change={updateFFT}
    >
      <option value="">集計項目を選択して下さい</option>
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    {#if keyList.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        aria-label="表示項目"
        bind:value={key}
        on:change={updateFFTChart}
      >
        <option value="">3Dチャート</option>
        {#each keyList as k}
          <option value={k}>{k}</option>
        {/each}
      </select>
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        aria-label="グラフの表示モード"
        bind:value={fftType}
        on:change={updateFFTChart}
      >
        <option value="hz">周波数</option>
        <option value="time">周期</option>
      </select>
    {/if}
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
