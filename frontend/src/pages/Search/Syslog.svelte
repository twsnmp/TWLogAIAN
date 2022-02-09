<script>
  import { X16, Search16,Gear16 } from "svelte-octicons";
  import Query from "./Query.svelte"
  import { createEventDispatcher,onMount, tick } from "svelte";
  import {makeLogCountChart,updateLogCountChart} from "../../js/logchart";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import * as echarts from 'echarts'
  const dispatch = createEventDispatcher();
  let page = "";
  const conf = {
    query: '',
    limit: 1000,
    range: false,
    start: "",
    end: "",
    keywords: [],
    geo: {
      eable: false,
      field: '',
      lat: '',
      long: '',
      range: '',
    },
  }
  let data = [];
  let errorMsg = "";
  let indexInfo = {
    Total: 0,
    Fields: [],
    Duration: "",
  };
  let result = {
    Logs:[],
    Hit: 0,
    Duration: 0.0,
  };
  window.go.main.App.GetIndexInfo().then((r) => {
    if (r) {
      indexInfo = r;
    }
  });
  let pagination = false;
  const search = () => {
    data.length = 0; // 空にする
    window.go.main.App.SearchLog(conf.query,conf.limit).then((r) => {
      if (r) {
        result = r;
        const d = [];
        r.Logs.forEach((l) => {
          d.push([l.Score, l.Time, l.All]);
        });
        data = d
        updateLogCountChart(r.Logs);
        if (r.Logs.length > 20) {
          pagination = {
            limit: 10,
            enable: true,
          };
        } else {
          pagination = false;
        }
      }
    });
  };

  onMount(() => {
    makeLogCountChart("chart");
  });

  const columns = [
    {
      name: "スコア",
      width: "10%",
      formatter: (cell) => Number.parseFloat(cell).toFixed(2),
    },{
      name: "日時",
      width: "20%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'),
    },{
      name: "ログ",
      width: "70%",
    },
  ];

  const cancel = () => {
    window.go.main.App.CloseWorkDir();
    dispatch("done", { page: "wellcome" });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const setting = () => {
    page = "query";
  };

  const handleDone = (e) => {
    page = "";
    updateChart();
  };

  const updateChart = async () => {
    await tick();
    makeLogCountChart("chart");
    updateLogCountChart(result.Logs);
  };

  const onResize = () => {
    updateLogCountChart(result.Logs);
  };

</script>

<svelte:window on:resize={onResize} />
{#if page == "query"}
  <Query {conf} fields={indexInfo.Fields} on:done={handleDone}  />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 95%;">
      <div class="Box-header d-flex flex-items-center">
        <h3 class="Box-title overflow-hidden flex-auto">Syslog分析</h3>
        <span class="f6">
          ログ総数:{indexInfo.Total}/項目数:{ indexInfo.Fields.length}/処理時間:{indexInfo.Duration}
          {#if result.Hit > 0 }
            /ヒット数:{result.Hit}/検索時間:{result.Duration}
          {/if}
        </span>
      </div>
      {#if errorMsg != ""}
        <div class="flash flash-error">
          {errorMsg}
          <button
            class="flash-close js-flash-close"
            type="button"
            aria-label="Close"
            on:click={clearMsg}
          >
            <X16 />
          </button>
        </div>
      {/if}
      <div class="Box-body log">
        <div class="clearfix">
          <div class="col-9 float-left">
            <input 
              class="form-control input-block"
              type="text"
              placeholder="検索文"
              aria-label="検索文"
              bind:value={conf.query}
            />
          </div>
          <div class="col-3 float-left">
            <button class="btn  btn-primary ml-2" type="button" on:click={search}>
              <Search16 />
              検索
            </button>
            <button class="btn  btn-secondary" type="button" on:click={setting}>
              <Gear16 />
              条件
            </button>
          </div>
        </div>
        <div id="chart" />
        <Grid {data} sort search {pagination} {columns} language={jaJP} />
      </div>
      <div class="Box-footer text-right">
        <button class="btn  btn-secondary" type="button" on:click={cancel}>
          <X16 />
          終了
        </button>
      </div>
  </div>
{/if}

<style>
  #chart {
    width: 100%;
    height: 150px;
    margin: 5px auto;
  }
</style>
