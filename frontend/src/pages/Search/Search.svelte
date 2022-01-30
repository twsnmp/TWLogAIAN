<script>
  import { X16, Search16 } from "svelte-octicons";
  import { createEventDispatcher,onMount } from "svelte";
  import * as echarts from "echarts";
  import Grid from "gridjs-svelte";
  const dispatch = createEventDispatcher();
  const data = [];
  let   query = '';
  let errorMsg = "";

  let pagination = false;
  const search = () => {
    data.length = 0; // 空にする
    window.go.main.App.SearchLog(query).then((logs) => {
      if (logs) {
        logs.forEach((l) => {
          console.log(l)
          data.push([l.Score, l.Time, l.Raw]);
        });
        if (logs.length > 20) {
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
    echarts.init(document.getElementById("chart")).setOption({
      title: false,
      backgroundColor: "#ccc",
      tooltip: {},
      xAxis: {},
      yAxis: {},
      series: [
        {
          type: "bar",
          smooth: true,
          data: [
            [12, 5],
            [24, 20],
            [36, 36],
            [48, 10],
            [60, 10],
            [72, 20],
          ],
        },
      ],
    });
  });


  const columns = [
    {
      name: "スコア",
      sort: true,
      width: "10%",
    },
    {
      name: "日時",
      sort: true,
      width: "20%",
    },
    {
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
    <div class="Box-header">
      <h3 class="Box-title">ログ分析</h3>
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
        <div class="col-10 float-left">
          <input 
            class="form-control input-block"
            type="text"
            placeholder="検索文"
            aria-label="検索文"
            bind:value={query}
          />
        </div>
        <div class="col-2 float-left text-right">
          <button class="btn  btn-primary" type="button" on:click={search}>
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
    margin: 5% auto;
  }
</style>
