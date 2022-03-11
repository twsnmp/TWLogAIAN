<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showHeatmap, resizeHeatmap,getHeatmapImage } from "./heatmap";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let field = "";
  let sumUnit = "day";
  let calcMode = "sum";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "曜日",
    },
    {
      name: "時間帯",
    },
    {
      name: "値",
    },
  ];

  let pagination = false;

  const updateHeatmap = async () => {
    await tick();
    data = showHeatmap("chart",logs,field,sumUnit,field ? calcMode : "sum",dark);
    columns[0].name = sumUnit == "day" ? "日付" : "曜日";
    columns[2].name = field ? getFieldName(field) : "件数";
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
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateHeatmap();
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
      Type: "時間帯別ヒートマップ",
      Title: "時間帯別ヒートマップ",
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
    window.go.main.App.Export(exportType,exportData).then(()=>{
      saveBusy = false;
      exportType = "";
    });
  }

  const onResize = () => {
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeHeatmap();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">時間帯別ヒートマップ</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="集計項目"
      bind:value={field}
      on:change="{updateHeatmap}"
    >
      <option value="">回数</option>
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
      <option value="week">曜日単位</option>
      <option value="day">日単位</option>
    </select>
    {#if field != "" }
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select ml-2"
        bind:value={calcMode}
        on:change="{updateHeatmap}"
      >
        <option value="sum">合計</option>
        <option value="mean">平均</option>
        <option value="median">中央値</option>
        <option value="variance">分散</option>
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
    height: 450px;
    margin: 5px auto;
  }
</style>
