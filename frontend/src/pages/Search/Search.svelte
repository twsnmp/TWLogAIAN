<script>
  import { X16, Search16 } from "svelte-octicons";
  import { createEventDispatcher,onMount } from "svelte";
  import {makeLogCountChart,updateLogCountChart} from "../../js/logchart";
  import Grid from "gridjs-svelte";
  import * as echarts from 'echarts'
  const dispatch = createEventDispatcher();
  const data = [];
  let query = '';
  let errorMsg = "";
  let indexInfo = {
    Total: 0,
    Fileds: [],
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
    window.go.main.App.SearchLog(query).then((r) => {
      if (r) {
        result = r;
        r.Logs.forEach((l) => {
          data.push([l.Score, l.Time, l.Raw]);
        });
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
      sort: true,
      width: "10%",
      formatter: (cell) => Number.parseFloat(cell).toFixed(2),
    },{
      name: "日時",
      sort: true,
      width: "20%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },{
      name: "ログ",
      sort: true,
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

</script>

<div class="Box mx-auto" style="max-width: 1600px;">
    <div class="Box-header d-flex flex-items-center">
      <h3 class="Box-title overflow-hidden flex-auto">ログ分析</h3>
      <span class="f6">
        ログ総数:{indexInfo.Total}/項目数:{ indexInfo.Fileds.length}/処理時間:{indexInfo.Duration}
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
    <div class="Box-body">
      <div id="chart" />
      <Grid {data} {pagination} {columns} />
    </div>
    <div class="Box-footer">
      <div class="clearfix">
        <div class="col-9 float-left">
          <input 
            class="form-control input-block"
            type="text"
            placeholder="検索文"
            aria-label="検索文"
            bind:value={query}
          />
        </div>
        <div class="col-3 float-left">
          <button class="btn  btn-primary ml-2" type="button" on:click={search}>
            <Search16 />
            検索
          </button>
          <button class="btn  btn-secondary" type="button" on:click={cancel}>
            <X16 />
            終了
          </button>
        </div>
      </div>
    </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 150px;
    margin: 5px auto;
  }
</style>
