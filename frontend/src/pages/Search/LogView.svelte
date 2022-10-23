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
    Pencil16,
  } from "svelte-octicons";
  import Query from "./Query.svelte";
  import Result from "./Result.svelte";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showLogChart, resizeLogChart, getLogChartImage } from "./logchart";
  import { getLogData, getLogColums, getGridSearch } from "./logview";
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
  import Memo from "../Report/Memo.svelte";
  import EditExtractorType from "../Setting/EditExtractorType.svelte";
  import { getTableLimit, loadFieldTypes, getFieldType } from "../../js/define";
  import numeral from "numeral";
  import AutoEncoder from "./AutoEncoder.svelte";
  import * as echarts from "echarts";

  const dispatch = createEventDispatcher();
  const excludeColMap = {"copy": true, "memo": true, "extractor": true};
  let page = "";
  let showQuery = false;
  let busy = false;
  let dark = false;
  const conf = {
    query: "",
    limit: "10000",
    anomaly: "",
    vector: "",
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
    extractor: "",
  };
  let data = [];
  let columns = [];
  let aecdata = [];
  let aeStart;
  let infoMsg = "";
  let errorMsg = "";
  let anomalyTime = "";
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
    Fields: [],
    ExFields: [],
  };
  window.go.main.App.GetIndexInfo().then((r) => {
    if (r) {
      indexInfo = r;
    }
  });
  let pagination = false;
  let logView = "";
  let gridSearch = true;
  let filter = {
    st: false,
    et: false,
  };
  const setLogTable = () => {
    columns = getLogColums(logView, logView == "ex_data" ? result.ExFields :  result.Fields);
    data = getLogData(result, logView, filter);
    gridSearch = getGridSearch(logView);
    if (data.length > 10) {
      pagination = {
        page: 0,
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  };

  let lastExtractor = "";

  const search = () => {
    data.length = 0; // 空にする
    aecdata = [];
    anomalyTime = "";
    showQuery= false; // 検索する時は詳細設定を表示しない
    filter.st = false; // 時間フィルターをリセットする
    filter.et = false;
    const limit = conf.limit * 1 > 100 ? conf.limit * 1 : 1000;
    busy = true;
    window.go.main.App.SearchLog(
      conf.query,
      conf.anomaly,
      conf.vector,
      conf.extractor,
      limit
    ).then((r) => {
      busy = false;
      if (r) {
        result = r;
        if (logView == "") {
          logView = r.View;
        } 
        if (lastExtractor != conf.extractor && r.ExFields.length > 0) {
          logView = "ex_data";
        }
        lastExtractor = conf.extractor;
        if (logView == "ex_data" && r.ExFields.length < 1 ) {
          logView = r.View;
        }
        if (r.ErrorMsg == "" && r.Logs.length > 0 && conf.query != "" ) {
          conf.history = conf.history.filter((h) => h != conf.query);
          conf.history.push(conf.query);
        }
        if (conf.anomaly == "autoencoder") {
          busy = true;
          showAutoencoder = true;
          aeStart = new Date();
          return;
        } else if (conf.anomaly != "") {
          anomalyTime = (r.AnomalyDur / 1000.0).toFixed(3) + "s";
        }
        errorMsg = r.ErrorMsg;
        setLogTable();
        updateChart();
      }
    });
  };

  let extractorTypes = {};
  let extractorTypeList = [];

  onMount(() => {
    loadFieldTypes();
    getExtractorTypes();
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateChart();
    });
    window.go.main.App.GetHistory().then((r) => {
      if (r) {
        conf.history = r;
      }
    });
  });

  const getExtractorTypes = () => {
    window.go.main.App.GetExtractorTypes().then((r) => {
      if (r) {
        extractorTypes = r;
        extractorTypeList = [];
        for (let k in extractorTypes) {
          extractorTypeList.push(extractorTypes[k]);
        }
        extractorTypeList.sort((a, b) => a.Name > b.Name);
      }
    });
  };
 

  const end = () => {
    window.go.main.App.SaveHistory(conf.history);
    window.go.main.App.CloseWorkDir();
    dispatch("done", { page: "wellcome" });
  };

  const back = () => {
    window.go.main.App.SaveHistory(conf.history);
    dispatch("done", { page: "setting" });
  };

  const clearMsg = () => {
    infoMsg = "";
    errorMsg = "";
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
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getLogChartImage();
    }
    columns.forEach((e) => {
      if (!excludeColMap[e.id]) {
        exportData.Header.push(e.name);
      }
    });
    data.forEach((l) => {
      const row = [];
      if (logView == "data" || logView == "ex_data") {
        columns.forEach((c) => {
          if (!excludeColMap[c.id]) {
            const v = c.convert && c.formatter ? c.formatter(l[c.id]) : l[c.id] || "";
            row.push(v);
          }
        });
      } else {
        l.forEach((e, i) => {
          if (!excludeColMap[columns[i].id]) {
            const v =
              columns[i] && columns[i].convert && columns[i].formatter
                ? columns[i].formatter(e)
                : e;
            row.push(v);
          }
        });
      }
      exportData.Data.push(row);
    });
    window.go.main.App.Export(exportType, exportData).then(() => {
      saveBusy = false;
      exportType = "";
    });
  };

  const onResize = () => {
    resizeLogChart();
    if(pagination) {
      pagination.limit = getTableLimit();
    }
  };

  let dY = 0;
  const onWheel = (e) => {
    if (!pagination || logView == "data") {
      return;
    }
    dY += e.deltaY;
    if (dY > 120) {
      dY = 0;
      if (data.length > (pagination.page + 1) * pagination.limit){
        pagination.page++;
      }
    }
    if (dY < -120 ){
      dY = 0;
      if (pagination.page > 0) {
        pagination.page--;
      }
    }
    e.preventDefault();
  }

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

  const formatTime = (cell, query) => {
    return echarts.time.format(
      new Date(cell / (1000 * 1000)),
      query
        ? "{yyyy}-{MM}-{dd}T{HH}:{mm}:{ss}+09:00"
        : "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}"
    );
  };

  const cellClick = (e) => {
    if (!e || !e.detail || e.detail.length < 4) {
      return;
    }
    const me = e.detail[0];
    const cell = e.detail[1];
    const col = e.detail[2];
    const row = e.detail[3];

    if (col.id == "copy") {
      copyLog(row._cells,me.metaKey);
      return;
    } else if (col.id == "memo") {
      memoLog(row._cells);
      return;
    } else if (col.id == "extractor") {
      showEditExtractorType(cell.data,me.metaKey);
      return;
    }
    // それ以外はmetaキー(Command?)を押していたらフィルターに入力
    if (!me.metaKey || !cell || !cell.data) {
      return;
    }
    // altキーを押した場合は、除外
    conf.query += getFilter(col.id, cell.data, me.altKey);
  };

  const exFieldList = ["level","score","anomalyScore","copy","memo","extract","all"];

  const getFilter = (id, data, exclude) => {
    if (!data) {
      return "";
    }
    if (exFieldList.includes(id)) {
      return "";
    }
    let op = "";
    if (id == "timestamp") {
      op = exclude ? "<=" : ">=";
      const time = formatTime(data * 1, true);
      return ` +time:${op}"${time}"`;
    }
    switch (getFieldType(id)) {
      case "number":
        op = exclude ? "<" : "=";
        return ` ${id}:${op}${data}`;
      case "string":
        if (id == "clientip") {
          const a = data.split("(");
          if (a.length > 1) {
            data = a[0];
          }
        }
        op = exclude ? "-" : "+";
        return ` ${op}${id}:${data}`;
    }
    return "";
  };

  const copyLog = (cells,tm) => {
    const list = [];
    let timeStamp = "";
    if (cells.length < columns.length) {
      return;
    }
    for (let i = 0; i < cells.length ; i++) {
      if (!excludeColMap[columns[i].id]) {
        const v = columns[i] && columns[i].convert && columns[i].formatter
                ? columns[i].formatter(cells[i].data)
                : cells[i].data;
        list.push(v);
        if (columns[i].id == "_timestamp") {
          timeStamp = formatTime(cells[i].data * 1,true);
        }
      }
    }
    copy(tm ? timeStamp : list.join("\t"));
  };

  const copy = (text) => {
    if (!navigator.clipboard || !navigator.clipboard) {
      errorMsg = "コピーできません。";
    }
    navigator.clipboard.writeText(text).then(
      () => {
        infoMsg = "コピーしました。";
        setTimeout(() => {
          infoMsg = "";
        }, 2000);
      },
      () => {
        errorMsg = "コピーエラーです。";
      }
    );
  };

  const memoLog = (cells) => {
    const list = [];
    let timeStamp = 0;
    if (cells.length < columns.length) {
      return;
    }
    for (let i = 0; i < cells.length ; i++) {
      if (!excludeColMap[columns[i].id]) {
        const v = columns[i] && columns[i].convert && columns[i].formatter
                ? columns[i].formatter(cells[i].data)
                : cells[i].data;
        list.push(v);
        if (columns[i].id == "_timestamp") {
          timeStamp = cells[i].data * 1;
        }
      }
    }
    if (list.length < 1 || timeStamp == 0 ) {
      return;
    }
    memo(timeStamp, list.join("\t"));
  };

  const memo = (time, text) => {
    infoMsg = "メモしました。";
    window.go.main.App.AddMemo({
      Time: time,
      Log: text,
    });
    setTimeout(() => {
      infoMsg = "";
    }, 2000);
  };

  const chnageLogView = () => {
    setLogTable();
  };

  let showAutoencoder = false;
  const handleAutoencoder = () => {
    showAutoencoder = false;
    const aeEnd = new Date();
    anomalyTime =
      (
        (result.AnomalyDur + aeEnd.getTime() - aeStart.getTime()) /
        1000.0
      ).toFixed(3) + "s";
    busy = false;
    setLogTable();
    updateChart();
  };
  
  let extractorType = {
    Key: "",
    Name: "",
    Grok: "",
    CanEdit: true,
  }

  let testLog = "";
  const add = true;
  const showEditExtractorType = (log,cmd) => {
    if (cmd) {
      if (testLog != "") {
        testLog += "\n";
      }
      testLog += log;
    } else {
      testLog = log;
    }
    const now = new Date();
    extractorType = {
      Key: "e" + now.getTime(),
      Name: "New",
      Grok: "",
      CanEdit: true,
    }    
    page = "extractorType";
  }

  const handleEditExtractorDone = (e) => {
    if (e && e.detail && e.detail.save) {
      loadFieldTypes();
      getExtractorTypes();
    }
    page = "";
    updateChart();
  };


</script>

<svelte:window on:resize={onResize}  on:wheel={onWheel} />
{#if page == "result"}
  <Result {indexInfo} {dark} {aecdata} on:done={handleDone} />
{:else if page == "memo"}
  <Memo on:done={handleDone} />
{:else if page == "ranking"}
  <Ranking fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "time"}
  <Time fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "time3d"}
  <Time3D fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "cluster"}
  <Cluster fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "histogram"}
  <Histogram fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "fft"}
  <FFT fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "world"}
  <World fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "graph"}
  <Graph fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "globe"}
  <Globe fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "heatmap"}
  <Heatmap fields={result.Fields} logs={result.Logs} on:done={handleDone} />
{:else if page == "extractorType"}
  <EditExtractorType {extractorType} {add} {testLog} on:done={handleEditExtractorDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header d-flex flex-items-center">
      <h3 class="Box-title overflow-hidden flex-auto">ログ分析</h3>
      <span class="f6">
        ログ総数:{numeral(indexInfo.Total).format("0,0")}
        /処理時間:{indexInfo.Duration}
        {#if result.Hit < 1}
          /項目数:{indexInfo.Fields.length}
        {/if}
        {#if result.Hit > 0}
          /項目数:{result.Fields.length}
          /ヒット数:{numeral(result.Hit).format(
            "0,0"
          )}/検索時間:{result.Duration}
          {#if anomalyTime}
            /異常検知:{anomalyTime}
          {/if}
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
    {#if infoMsg != ""}
      <div class="flash">
        {infoMsg}
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
        <Query
          {conf}
          fields={indexInfo.Fields}
          {extractorTypeList}
          on:update={handleUpdateQuery}
        />
      </div>
    {/if}
    {#if showAutoencoder}
      <div class="Box-row">
        <AutoEncoder
          {dark}
          chartData={aecdata}
          logs={result.Logs}
          on:done={handleAutoencoder}
        />
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
        search={gridSearch}
        {pagination}
        {columns}
        language={jaJP}
        on:cellClick={cellClick}
      />
    </div>
    {#if !busy}
      <div class="Box-footer text-right">
        {#if result && result.Hit > 0}
          <!-- svelte-ignore a11y-no-onchange -->
          <select
            class="form-select mr-1"
            bind:value={logView}
            on:change={chnageLogView}
          >
            <option value="timeonly">タイムオンリー</option>
            {#if result.View == "syslog"}
              <option value="syslog">syslog</option>
            {/if}
            {#if result.View == "access"}
              <option value="access">アクセスログ</option>
            {/if}
            {#if result.View == "windows"}
              <option value="windows">Windows</option>
            {/if}
            {#if result.Fields.length > 0}
              <option value="data">抽出データ</option>
            {/if}
            {#if result.ExFields.length > 0}
              <option value="ex_data">検索時抽出データ</option>
            {/if}
            {#if conf.anomaly != ""}
              <option value="anomary">異常ログスコア</option>
            {/if}
          </select>
        {/if}
        <!-- svelte-ignore a11y-no-onchange -->
        {#if saveBusy}
          <span>保存中</span><span class="AnimatedEllipsis" />
        {:else}
          <select
            class="form-select mr-1"
            bind:value={exportType}
            on:change={exportLogs}
          >
            <option value="">エクスポート</option>
            {#if result && result.Hit > 0 && result.Fields.length > 0}
              <option value="csv">CSV</option>
              <option value="excel">Excel</option>
            {/if}
          </select>
        {/if}
        {#if result && result.Hit > 0 && result.Fields.length > 0}
          <!-- svelte-ignore a11y-no-onchange -->
          <select
            class="form-select mr-1"
            bind:value={report}
            on:change={showReport}
          >
            <option value="">レポート</option>
            <option value="memo">メモ</option>
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
        <button
          class="btn  btn-outline mr-1"
          type="button"
          on:click={() => {
            page = "result";
          }}
        >
          <Check16 />
          処理結果
        </button>
        <button class="btn  btn-secondary mr-1" type="button" on:click={back}>
          <Reply16 />
          戻る
        </button>
        <button class="btn  btn-secondary" type="button" on:click={end}>
          <X16 />
          終了
        </button>
      </div>
    {/if}
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
