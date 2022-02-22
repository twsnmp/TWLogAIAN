<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showGraph, resizeGraph } from "./graph";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let catFields = [];
  let numFields = [];
  let srcField = "";
  let dstField = "";
  let numField = "";
  let graphType = "gl";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "始点項目",
      width: "40%",
    },
    {
      name: "終点項目",
      width: "40%",
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

  const updateGraph = async () => {
    if( srcField == "" || dstField == "" ){
      return;
    }
    await tick();
    const m = showGraph("chart", logs, srcField,dstField,numField, graphType,dark);
    data = [];
    columns[0].name = getFieldName(srcField);
    columns[1].name = getFieldName(dstField);
    columns[3].name = getFieldName(numField);
    m.forEach((e) => {
      data.push([e.source,e.target,e.count,e.value]);
    });
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
    catFields =  getFields(fields,"string");
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
    });
  });

  const onResize = () => {
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeGraph();
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
      aria-label="始点項目"
      bind:value={srcField}
      on:change="{updateGraph}"
    >
      <option value="">始点項目を選択して下さい</option>
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="終点項目"
      bind:value={dstField}
      on:change="{updateGraph}"
    >
    <option value="">終点項目を選択して下さい</option>
    {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="数値項目"
      bind:value={numField}
      on:change="{updateGraph}"
    >
    <option value="">数値項目を選択して下さい</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="表示形式"
      bind:value={graphType}
      on:change="{updateGraph}"
    >
    <option value="gl">3D</option>
    <option value="force">力学モデル</option>
    <option value="circular">円形</option>
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
