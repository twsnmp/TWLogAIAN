<script>
  import { getFields, getFieldName,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { getRanking, showRankingChart, resizeRankingChart } from "./ranking";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";

  export let logs = [];
  export let fields = [];
  let list = [];
  let catFields = [];
  let selected = "";
  let dark = false;

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: "順位",
      width: "10%",
    },
    {
      name: "項目",
      width: "80%",
    },
    {
      name: "件数",
      width: "10%",
    },
  ];
  let pagination = false;

  const updateRanking = async () => {
    if (selected == "" ) {
      return;
    }
    await tick();
    list = getRanking(logs, selected);
    showRankingChart("chart", list, 50,dark);
    if (list.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
    data = [];
    for (let i = 0; i < list.length; i++) {
      data.push([i + 1, list[i].Name, list[i].Total]);
    }
  };

  onMount(() => {
    catFields = getFields(fields,"string");
    if( catFields.length > 0 ){
      selected = catFields[0];
      window.go.main.App.GetDark().then((v) => {
        dark = v;
        updateRanking();
      });
    }
  });

  const onResize = () => {
    if(pagination) {
      pagination.limit = getTableLimit();
    }
    resizeRankingChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">ランキング分析</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="項目"
      bind:value={selected}
      on:change={updateRanking}
    >
    <option value="">集計する項目を選択してください</option>
      {#each catFields as f}
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
