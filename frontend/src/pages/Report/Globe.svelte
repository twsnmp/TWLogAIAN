<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showGlobe, resizeGlobe } from "./globe";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let dark = false;
  let geoFields = [];
  let numFields = [];
  let srcField = "";
  let dstField = "";
  let numField = "";

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

  const updateGlobe = async () => {
    if( srcField == "" ){
      return;
    }
    await tick();
    const m = showGlobe("chart",logs,srcField,dstField,numField,dark);
    data = [];
    columns[0].name = getFieldName(srcField);
    columns[1].name = getFieldName(dstField);
    columns[3].name = getFieldName(numField);
    m.forEach((e) => {
      data.push([e.src,e.dst,e.count,e.value]);
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
    geoFields =  getFields(fields,"latlong");
    if(srcField == "" && geoFields.length >0 ){
      srcField = geoFields[0];
    }
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateGlobe();
    });
  });

  const onResize = () => {
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeGlobe();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">フロー分析（地球儀）</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="始点項目"
      bind:value={srcField}
      on:change="{updateGlobe}"
    >
      <option value="">始点項目を選択して下さい</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="終点項目"
      bind:value={dstField}
      on:change="{updateGlobe}"
    >
      <option value="">東京</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="数値項目"
      bind:value={numField}
      on:change="{updateGlobe}"
    >
    <option value="">数値項目を選択して下さい</option>
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
