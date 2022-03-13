<script>
  import {
    X16,
    Search16,
    TriangleDown16,
    TriangleUp16,
    Check16,
    Trash16,
    Reply16,
    Copy16,
  } from "svelte-octicons";
  import Query from "./Query.svelte";
  import Result from "./Result.svelte";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showLogChart, resizeLogChart, getLogChartImage } from "./logchart";
  import { getLogData, getLogColums, getSelectedLogs, clearSelectedLogs } from "./logview";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import Ranking from "../Report/Ranking.svelte";
  import Time3D from "../Report/Time3D.svelte";
  import Time from "../Report/Time.svelte";
  import Cluster from "../Report/Cluster.svelte";
  import Histogram from "../Report/Histogram.svelte";
  import FFT from "../Report/FFT.svelte";
  import World from "../Report/World.svelte";
  import Graph from "../Report/Graph.svelte";
  import Globe from "../Report/Globe.svelte";
  import Heatmap from "../Report/Heatmap.svelte";
  import { getTableLimit, loadFieldTypes } from "../../js/define";
  import numeral from "numeral";
  import CopyClipBoard from "../../CopyClipBoard.svelte";

  const dispatch = createEventDispatcher();
  let page = "";
  let showQuery = false;
  let busy = false;
  let dark = false;
  const conf = {
    query: "",
    limit: "10000",
    history: [],
    keyword: {
      field: "",
      mode: "+",
      key: "",
    },
    number: {
      field: "",
      oper: "<",
      value: "0.0",
    },
    range: {
      start: "",
      end: "",
    },
    geo: {
      lat: "",
      long: "",
      range: "",
    },
  };
  let data = [];
  let columns = [];
  let indexInfo = {
    Total: 0,
    Fields: [],
    Duration: "",
  };
  let result = {
    Logs: [],
    View: "",
    Hit: 0,
    Duration: 0.0,
    ErrorMsg: "",
  };
  window.go.main.App.GetIndexInfo().then((r) => {
    if (r) {
      indexInfo = r;
    }
  });
  let pagination = false;
  let filter = {
    st: false,
    et: false,
  };
  let logView  = "";
  let selectedLogs = "";
  const setLogTable = () => {
    selectedLogs = "";
    clearSelectedLogs();
    columns = getLogColums(logView, indexInfo.Fields);
    data = getLogData(result,logView, filter);
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  };
  const search = () => {
    data.length = 0; // 空にする
    const limit = conf.limit * 1 > 100 ? conf.limit * 1 : 1000;
    busy = true;
    window.go.main.App.SearchLog(conf.query, limit).then((r) => {
      busy = false;
      if (r) {
        result = r;
        logView = r.View;
        setLogTable();
        if (r.ErrorMsg == "") {
          conf.history.push(conf.query);
        }
        updateChart();
      }
    });
  };

  onMount(() => {
    loadFieldTypes();
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateChart();
    });
  });

  const end = () => {
    window.go.main.App.CloseWorkDir();
    dispatch("done", { page: "wellcome" });
  };

  const back = () => {
    dispatch("done", { page: "setting" });
  };

  const clearMsg = () => {
    result.ErrorMsg = "";
  };
  let report = "";

  const handleDone = (e) => {
    page = "";
    report = "";
    updateChart();
  };

  const showReport = () => {
    page = report;
  };

  const zoomCallback = (st, et) => {
    filter.st = st;
    filter.et = et;
    setLogTable();
  };

  const updateChart = async () => {
    await tick();
    showLogChart("chart", result.Logs, dark, zoomCallback);
  };

  let exportType = "";
  let saveBusy = false;
  const exportLogs = () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: "ログ",
      Title: "ログ分析",
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getLogChartImage();
    }
    if (exportType != "logtypes") {
      columns.forEach((e) => {
        exportData.Header.push(e.name);
      });
      data.forEach((l) => {
        const row = [];
        if (logView == "data") {
          columns.forEach((c) => {
            const v =
              c.convert && c.formatter
                ? c.formatter(l[c.id])
                : l[c.id];
            row.push(v)
          });
        }else {
          l.forEach((e, i) => {
            const v =
              columns[i] && columns[i].convert && columns[i].formatter
                ? columns[i].formatter(e)
                : e;
            row.push(v);
          });
        }
        exportData.Data.push(row);
      });
    }
    window.go.main.App.Export(exportType, exportData).then(() => {
      saveBusy = false;
      exportType = "";
    });
  };

  const onResize = () => {
    if (pagination) {
      pagination.limit = getTableLimit();
    }
    resizeLogChart();
  };

  const handleUpdateQuery = (e) => {
    if (e && e.detail && e.detail.query) {
      if (e.detail.add) {
        conf.query += e.detail.query;
      } else {
        conf.query = e.detail.query;
      }
    }
  };

  const clear = () => {
    conf.query = "";
  };

  const rowClick = (e) => {
    setTimeout(()=>{
      selectedLogs = getSelectedLogs();
    },10);
  };

  let showCopy = false;
  const copy = () => {
    showCopy = true;
    const app = new CopyClipBoard({
      target: document.getElementById("clipboard"),
      props: { selectedLogs },
    });
    app.$destroy();
    setTimeout(()=>{
      showCopy = false;
    },2000);
  };

  const chnageLogView = ()  => {
    setLogTable();
  };

</script>

<svelte:window on:resize={onResize} />
{#if page == "result"}
  <Result {indexInfo} on:done={handleDone} />
{:else if page == "ranking"}
  <Ranking fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "time"}
  <Time fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "time3d"}
  <Time3D fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "cluster"}
  <Cluster fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "histogram"}
  <Histogram
    fields={indexInfo.Fields}
    logs={result.Logs}
    on:done={handleDone}
  />
{:else if page == "fft"}
  <FFT fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "world"}
  <World fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "graph"}
  <Graph fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "globe"}
  <Globe fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "heatmap"}
  <Heatmap fields={indexInfo.Fields} logs={result.Logs} on:done={handleDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header d-flex flex-items-center">
      <h3 class="Box-title overflow-hidden flex-auto">ログ分析</h3>
      <span class="f6">
        ログ総数:{numeral(indexInfo.Total).format("0,0")}/項目数:{indexInfo
          .Fields.length}/処理時間:{indexInfo.Duration}
        {#if result.Hit > 0}
          /ヒット数:{numeral(result.Hit).format(
            "0,0"
          )}/検索時間:{result.Duration}
        {/if}
      </span>
    </div>
    {#if result.ErrorMsg != ""}
      <div class="flash flash-error">
        {result.ErrorMsg}
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
    <div class="Box-row">
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
          {#if conf.query != ""}
            <button class="btn btn-danger" type="button" on:click={clear}>
              <Trash16 />
            </button>
          {/if}
          {#if !showQuery}
            <button
              class="btn  btn-secondary"
              type="button"
              on:click={() => {
                showQuery = true;
              }}
            >
              <TriangleDown16 />
            </button>
          {:else}
            <button
              class="btn  btn-secondary"
              type="button"
              on:click={() => {
                showQuery = false;
              }}
            >
              <TriangleUp16 />
            </button>
          {/if}
          {#if !busy}
            <button
              class="btn btn-primary ml-2"
              type="button"
              on:click={search}
            >
              <Search16 />
              検索
            </button>
          {:else}
            <button class="btn btn-primary ml-2" aria-disabled="true">
              <Search16 />
              <span>検索中</span><span class="AnimatedEllipsis" />
            </button>
          {/if}
        </div>
      </div>
    </div>
    {#if showQuery}
      <div class="Box-row">
        <Query {conf} fields={indexInfo.Fields} on:update={handleUpdateQuery} />
      </div>
    {/if}
    <div class="Box-row">
      <div id="chart" />
    </div>
    <div class="Box-row markdown-body log">
      <Grid
        {data}
        sort
        resizable
        search
        {pagination}
        {columns}
        language={jaJP}
        on:rowClick={rowClick}
      />
    </div>
    <div class="Box-footer text-right">
      {#if result && result.Hit > 0 }
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          class="form-select mr-2"
          bind:value={logView}
          on:change={chnageLogView}
        >
          <option value="">タイムオンリー</option>
          {#if result.View == "syslog"}
            <option value="syslog">syslog</option>
          {/if}
          {#if result.View == "access"}
            <option value="access">アクセスログ</option>
          {/if}
          {#if result.View == "windows"}
            <option value="windows">Windows</option>
          {/if}
          {#if indexInfo.Fields.length > 0}
            <option value="data">抽出データ</option>
          {/if}
        </select>
      {/if}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>保存中</span><span class="AnimatedEllipsis" />
      {:else}
        <select
          class="form-select mr-2"
          bind:value={exportType}
          on:change={exportLogs}
        >
          <option value="">エクスポート</option>
          {#if result && result.Hit > 0 && indexInfo.Fields.length > 0}
            <option value="csv">CSV</option>
            <option value="excel">Excel</option>
          {/if}
          <option value="logtypes">ログ種別定義</option>
        </select>
      {/if}
      {#if result && result.Hit > 0 && indexInfo.Fields.length > 0}
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          class="form-select mr-2"
          bind:value={report}
          on:change={showReport}
        >
          <option value="">レポート</option>
          <option value="ranking">ランキング分析</option>
          <option value="time">時系列分析</option>
          <option value="time3d">時系列3D分析</option>
          <option value="cluster">クラスター分析</option>
          <option value="histogram">ヒストグラム分析</option>
          <option value="fft">FFT分析</option>
          <option value="world">位置情報分析</option>
          <option value="graph">グラフ（フロー）分析</option>
          <option value="globe">フロー分析（地球儀)</option>
          <option value="heatmap">ヒートマップ</option>
        </select>
      {/if}
      {#if selectedLogs != ""}
        <button class="btn btn-outline mr-2" type="button" on:click={copy}>
          <Copy16 />
          コピー
        </button>
        {#if showCopy}
          <span class="branch-name">Copied</span>
        {/if}
      {/if}
      <button
        class="btn  btn-outline mr-2"
        type="button"
        on:click={() => {
          page = "result";
        }}
      >
        <Check16 />
        処理結果
      </button>
      <button class="btn  btn-secondary mr-2" type="button" on:click={back}>
        <Reply16 />
        戻る
      </button>
      <button class="btn  btn-secondary" type="button" on:click={end}>
        <X16 />
        終了
      </button>
    </div>
  </div>
{/if}
<div id="clipboard" />

<style>
  #chart {
    width: 100%;
    height: 220px;
    margin: 5px auto;
  }
</style>
