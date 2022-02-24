<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showWorldMap, resizeWorldMap,getWorldMapImage } from "./world";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let geoFields = [];
  let geoField = "";
  let numField = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "位置（緯度経度）",
      width: "20%",
    },
    {
      name: "国",
      width: "10%",
    },
    {
      name: "都市",
      width: "10%",
    },
    {
      name: "件数",
      width: "10%",
    },
    {
      name: "値",
      width: "10%",
    },
  ];

  let pagination = false;

  const updateWorldMap = async () => {
    if( geoField == "" ){
      return;
    }
    await tick();
    const m = showWorldMap("chart", logs, geoField,numField,dark);
    data = [];
    columns[4].name = getFieldName(numField);
    m.forEach((e,k) => {
      data.push([k,e.country,e.city,e.count,e.value]);
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
    const tmp =  getFields(fields,"latlong");
    geoFields = [];
    tmp.forEach((g)=>{
      g = g.replace("_latlong","");
      if (geoField == "") {
        geoField = g;
      }
      geoFields.push(g);
    });
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateWorldMap();
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
      Type: "位置情報分析",
      Title: "位置情報分析",
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getWorldMapImage();
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
    resizeWorldMap();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">位置情報分析</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="位置情報項目"
      bind:value={geoField}
      on:change="{updateWorldMap}"
    >
      <option value="">位置情報の項目を選択して下さい</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="数値項目"
      bind:value={numField}
      on:change="{updateWorldMap}"
    >
    <option value="">色分けの数値項目を選択して下さい</option>
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
    height: 500px;
    margin: 5px auto;
  }
</style>
