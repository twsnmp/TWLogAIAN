<script>
  import { 
    GetIndexInfo,
    SearchLog,
    GetDark,
    GetHistory,
    GetExtractorTypes,
    SaveHistory,
    CloseWorkDir,
    Export,
    AddMemo,
  } from "../../../wailsjs/go/main/App.js";
  import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime.js";
  import {
    X16,
    Search16,
    TriangleDown16,
    TriangleUp16,
    Check16,
    Trash16,
    Reply16,
    Checklist16,
    Question16,
  } from "svelte-octicons";
  import SerachConf from "./SearchConf.svelte";
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
  import LogType from "../Setting/LogType.svelte";
  import EditExtractorType from "../Setting/EditExtractorType.svelte";
  import { getTableLimit, loadFieldTypes, getFieldType } from "../../js/define";
  import numeral from "numeral";
  import AutoEncoder from "./AutoEncoder.svelte";
  import * as echarts from "echarts";
  import { _,getLocale } from '../../i18n/i18n';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  const dispatch = createEventDispatcher();
  const excludeColMap = { copy: true, memo: true, extractor: true };
  let page = "";
  let showConf = false;
  let busy = false;
  let dark = false;
  const conf = {
    mode: "simple",
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
      start: echarts.time.format(Date.now() - 3600 * 1000,"{yyyy}-{MM}-{dd}T{HH}:{mm}"),
      end: echarts.time.format(Date.now(), "{yyyy}-{MM}-{dd}T{HH}:{mm}"),
      range: "-60",
      target: "",
      mode: "",
    },
    geo: {
      mode: "",
      lat: "35.689487",
      long: "139.691711",
      range: "100",
    },
    extractor: "",
    highlightMode: "",
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
    StartTime: 0,
    EndTime: 0,
  };
  let result = {
    Logs: [],
    View: "",
    Hit: 0,
    Duration: 0.0,
    ErrorMsg: "",
    Fields: [],
    ExFields: [],
    LastTime: 0,
  };
  let pagination = false;
  let logView = "";
  let gridSearch = true;
  let filter = {
    st: false,
    et: false,
  };
  const setLogTable = () => {
    columns = getLogColums(
      logView,
      logView == "ex_data" ? result.ExFields : result.Fields,
      conf
    );
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

  const getTZStr = () => {
    const o = -(new Date().getTimezoneOffset());
    const ao = Math.abs(o);
    return (o < 0 ? '-' : '+') + ( '00' +   Math.trunc(ao/60) ).slice(-2) + ":" +   ('00' + ao % 60).slice(-2);
  };

  const getTargetRange = () => {
    let t = Date.parse(conf.range.target);
    if (!t || t == NaN) {
      const nt = d < 0 ? indexInfo.EndTime : indexInfo.StartTime;
      t = new Date(nt / (1000 * 1000)).getTime();
    }
    let d = parseInt(conf.range.range);
    d = d == 0 ?  3600 : d;
    const sd = d < 0 ? -d : 0;
    const ed = d < 0 ?  0 : d;
    const tz = getTZStr();
    return ` +time:>="` +
          echarts.time.format(
            new Date(t - (sd * 1000)),
            "{yyyy}-{MM}-{dd}T{HH}:{mm}:{ss}" +tz
          ) +
          `"` +
          ` +time:<="` +
          echarts.time.format(
            new Date(t + (ed * 1000)),
            "{yyyy}-{MM}-{dd}T{HH}:{mm}:{ss}"+ tz
          ) +
          `"`;
  };

  const getTimeFilter = () => {
    let ret = "";
    const tz = getTZStr();
    switch(conf.range.mode) {
    case "target":
      return getTargetRange();
      break;
    case "range":
      if (conf.range.start) {
        ret += ` +time:>="` + echarts.time.format(Date.parse(conf.range.start), "{yyyy}-{MM}-{dd}T{HH}:{mm}:{ss}") + `${tz}"`;
      }
      if (conf.range.end) {
        ret += ` +time:<="` + echarts.time.format(Date.parse(conf.range.end), "{yyyy}-{MM}-{dd}T{HH}:{mm}:{ss}") + `${tz}"`;
      }
      break;
    }
    return ret;
  }

  const getGeoFilter = () => {
    if (conf.geo.mode == "" || !conf.geo.field) {
      return "";
    }
    const lat = conf.geo.lat || 0;
    const long = conf.geo.long || 0;
    const range = conf.geo.range || 100;
    return conf.geo.field +
      "," +
      lat +
      "," +
      long +
      "," +
      range +
      "km";
  };

  let searchInfo = "";

  const setSearchInfo = () => {
    let r =  $_("LogView.Total") + ":" +
        numeral(indexInfo.Total).format("0,0") + "/" +
        $_("LogView.IndexTime") + ":" +
        indexInfo.Duration + "/";
      r += $_("LogView.Field") + ":" + 
        indexInfo.Fields.length;
    if (result.Hit > 0) {
      r += "/" + $_("LogView.Hit") + ":" + numeral(result.Hit).format("0,0");
      r += "/" + $_("LogView.SearchDur") + ":" + result.Duration;
    }
    if (anomalyTime) {
      r += "/" + $_("LogView.AnomalyTime") + ":" + anomalyTime
    }
    searchInfo = r;
  }
  GetIndexInfo().then((r) => {
    if (r) {
      indexInfo = r;
      setSearchInfo();
    }
  });

  const search = () => {
    data.length = 0;
    aecdata = [];
    anomalyTime = "";
    showConf = false;
    filter.st = false;
    filter.et = false;
    const request = {
      Mode: conf.mode,
      Query: conf.query,
      TimeFilter: getTimeFilter(),
      GeoFilter: getGeoFilter(),
      Anomaly: conf.anomaly,
      Vector: conf.vector,
      Extractor: conf.extractor,
      Limit: conf.limit * 1 > 100 ? conf.limit * 1 : 1000,
    }
    busy = true;
    SearchLog(request).then((r) => {
      busy = false;
      if (r) {
        result = r;
        if (logView == "") {
          logView = r.View == "auto" ? "timeonly" : r.View;
        }
        if (lastExtractor != conf.extractor && r.ExFields.length > 0) {
          logView = "ex_data";
        }
        lastExtractor = conf.extractor;
        if (logView == "ex_data" && r.ExFields.length < 1) {
          logView = r.View;
        }
        if (r.ErrorMsg == "" && r.Logs.length > 0 && conf.query != "") {
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
        setSearchInfo();
      }
    });
  };

  let extractorTypes = {};
  let extractorTypeList = [];

  onMount(() => {
    loadFieldTypes();
    getExtractorTypes();
    GetDark().then((v) => {
      dark = v;
      updateChart();
    });
    GetHistory().then((r) => {
      if (r) {
        conf.history = r;
      }
    });
  });

  const getExtractorTypes = () => {
    GetExtractorTypes().then((r) => {
      if (r) {
        extractorTypes = r;
        extractorTypeList = [];
        for (let k in extractorTypes) {
          extractorTypeList.push(extractorTypes[k]);
        }
        extractorTypeList.sort((a, b) => a.Name > b.Name);
        extractorTypeList.unshift(
          {
            Key: "",
            Name: $_("SearchConf.NotUse"),
          }
        )
      }
    });
  };

  const end = () => {
    SaveHistory(conf.history);
    CloseWorkDir($_('Setting.StopTitle'),$_('Setting.CloseMsg')).then((r) => {
      if (r == "") {
        dispatch("done", { page: "wellcome" });
      }
    });
  };

  const back = () => {
    SaveHistory(conf.history);
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
            const v =
              c.convert && c.formatter ? c.formatter(l[c.id]) : l[c.id] || "";
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
    Export(exportType, exportData).then(() => {
      saveBusy = false;
      exportType = "";
    });
  };

  const onResize = () => {
    resizeLogChart();
    if (pagination) {
      pagination.limit = getTableLimit();
    }
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
      copyLog(row._cells, me.metaKey);
      return;
    } else if (col.id == "memo") {
      memoLog(row._cells);
      return;
    } else if (col.id == "extractor") {
      showEditExtractorType(cell.data, me.metaKey);
      return;
    }
    // Command + Click  to Filter
    if (!me.metaKey || !cell || !cell.data) {
      return;
    }
    // Exclude Alt Key
    setFilter(col.id, cell.data, me.altKey);
  };

  const exFieldList = [
    "level",
    "score",
    "anomalyScore",
    "copy",
    "memo",
    "extract",
    "all",
  ];

  const setFilter = (id, data, exclude) => {
    if (!data) {
      return;
    }
    if (id === "_timestamp") {
      conf.range.target = formatTime(data * 1, true);
      conf.range.mode = "target";
      conf.range.range = exclude ? 300 : -300;
      showConf = true;
      return;
    }
    if (conf.mode != "full") {
      return;
    }
    if (exFieldList.includes(id)) {
      return;
    }
    let op = "";
    switch (getFieldType(id)) {
      case "number":
        op = exclude ? "<" : "=";
        conf.query += ` ${id}:${op}${data}`;
        return;
      case "string":
        if (id == "clientip") {
          const a = data.split("(");
          if (a.length > 1) {
            data = a[0];
          }
        }
        op = exclude ? "-" : "+";
        conf.query += ` ${op}${id}:${data}`;
        return;
    }
    return;
  };

  const copyLog = (cells, tm) => {
    const list = [];
    let timeStamp = "";
    if (cells.length < columns.length) {
      return;
    }
    for (let i = 0; i < cells.length; i++) {
      if (!excludeColMap[columns[i].id]) {
        const v =
          columns[i] && columns[i].convert && columns[i].formatter
            ? columns[i].formatter(cells[i].data)
            : cells[i].data;
        list.push(v);
        if (columns[i].id == "_timestamp") {
          timeStamp = formatTime(cells[i].data * 1, true);
        }
      }
    }
    copy(tm ? timeStamp : list.join("\t"));
  };

  const copy = (text) => {
    if (!navigator.clipboard || !navigator.clipboard) {
      errorMsg = $_('LogView.CantCopy');
    }
    navigator.clipboard.writeText(text).then(
      () => {
        infoMsg = $_('LogView.Copied');
        setTimeout(() => {
          infoMsg = "";
        }, 2000);
      },
      () => {
        errorMsg = $_('LogView.CopyError');
      }
    );
  };

  const memoLog = (cells) => {
    const list = [];
    let timeStamp = 0;
    if (cells.length < columns.length) {
      return;
    }
    for (let i = 0; i < cells.length; i++) {
      if (!excludeColMap[columns[i].id]) {
        const v =
          columns[i] && columns[i].convert && columns[i].formatter
            ? columns[i].formatter(cells[i].data)
            : cells[i].data;
        list.push(v);
        if (columns[i].id == "_timestamp") {
          timeStamp = cells[i].data * 1;
        }
      }
    }
    if (list.length < 1 || timeStamp == 0) {
      return;
    }
    memo(timeStamp, list.join("\t"));
  };

  const memo = (time, text) => {
    infoMsg = $_('LogView.MemoMsg');
    AddMemo({
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
  };

  let testLog = "";
  const add = true;
  const showEditExtractorType = (log, cmd) => {
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
    };
    page = "extractorType";
  };

  const handleEditExtractorDone = (e) => {
    if (e && e.detail && e.detail.save) {
      loadFieldTypes();
      getExtractorTypes();
    }
    page = "";
    updateChart();
  };

  const showLogTypePage = () => {
    page = "logType";
  };

</script>

<svelte:window on:resize={onResize} />

{#if page == "result"}
  <Result {indexInfo} {dark} {aecdata} on:done={handleDone} />
{:else if page == "memo"}
  <Memo on:done={handleDone} dark/>
{:else if page == "ranking"}
  <Ranking fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "time"}
  <Time fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "time3d"}
  <Time3D fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "cluster"}
  <Cluster fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "histogram"}
  <Histogram fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "fft"}
  <FFT fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "world"}
  <World fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "graph"}
  <Graph fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "globe"}
  <Globe fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "heatmap"}
  <Heatmap fields={result.Fields} logs={result.Logs} on:done={handleDone} dark />
{:else if page == "extractorType"}
  <EditExtractorType
    {extractorType}
    {add}
    {testLog}
    on:done={handleEditExtractorDone}
  />
{:else if page == "logType"}
  <LogType on:done={handleDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header d-flex flex-items-center">
      <h3 class="Box-title overflow-hidden flex-auto">{$_('LogView.Title')}</h3>
      <span class="f6">{searchInfo}</span>
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
      <div id="infoMsg" class="flash">
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
            placeholder="{$_('LogView.SearchText')}"
            aria-label="{$_('LogView.SearchText')}"
            bind:value={conf.query}
          />
        </div>
        <div class="col-3 float-left">
          {#if conf.query != ""}
            <button class="btn btn-danger" type="button" on:click={clear}>
              <Trash16 />
            </button>
          {/if}
          <button
            class="btn  btn-secondary"
            type="button"
            on:click={() => {
              BrowserOpenURL("https://note.com/twsnmp/n/n1e8665af0ce2");
            }}
          >
            <Question16 />
          </button>
          {#if !showConf}
            <button
              class="btn  btn-secondary"
              type="button"
              on:click={() => {
                showConf = true;
              }}
            >
              <TriangleDown16 />
            </button>
          {:else}
            <button
              class="btn  btn-secondary"
              type="button"
              on:click={() => {
                showConf = false;
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
              {$_('LogView.SearchBtn')}
            </button>
          {:else}
            <button class="btn btn-primary ml-2" aria-disabled="true">
              <Search16 />
              <span>{$_('LogView.Searching')}</span><span class="AnimatedEllipsis" />
            </button>
          {/if}
        </div>
      </div>
    </div>
    {#if showConf}
      <div class="Box-row">
        <SerachConf
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
        language={gridLang}
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
            <option value="timeonly">{$_('LogView.TimeOnly')}</option>
            {#if result.View == "syslog" || result.View == "auto"}
              <option value="syslog">syslog</option>
            {/if}
            {#if result.View == "access" || result.View == "auto"}
              <option value="access">{$_('LogView.AccessLog')}</option>
            {/if}
            {#if result.View == "windows" || result.View == "auto"}
              <option value="windows">Windows</option>
            {/if}
            {#if result.Fields.length > 0}
              <option value="data">{$_('LogView.ExtractData')}</option>
            {/if}
            {#if result.ExFields.length > 0}
              <option value="ex_data">{$_('LogView.ExtractDataOnSearch')}</option>
            {/if}
            {#if conf.anomaly != ""}
              <option value="anomary">{$_('LogView.LogAnomaryScore')}</option>
            {/if}
          </select>
        {/if}
        <!-- svelte-ignore a11y-no-onchange -->
        {#if saveBusy}
          <span>{$_('LogView.Saving')}</span><span class="AnimatedEllipsis" />
        {:else}
          <select
            class="form-select mr-1"
            bind:value={exportType}
            on:change={exportLogs}
          >
            <option value="">{$_('LogView.ExportMenu')}</option>
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
            <option value="">{$_('LogView.ReportMenu')}</option>
            <option value="memo">{$_('LogView.MemoMenu')}</option>
            <option value="ranking">{$_('LogView.RankingMenu')}</option>
            <option value="time">{$_('LogView.TimeMenu')}</option>
            <option value="time3d">{$_('LogView.Time3DMenu')}</option>
            <option value="cluster">{$_('LogView.ClusterMenu')}</option>
            <option value="histogram">{$_('LogView.HistogramMenu')}</option>
            <option value="fft">{$_('LogView.FFTMenu')}</option>
            <option value="world">{$_('LogView.LocMenu')}</option>
            <option value="graph">{$_('LogView.GraphMenu')}</option>
            <option value="globe">{$_('LogView.GlobeMenu')}</option>
            <option value="heatmap">{$_('LogView.HeatmapMenu')}</option>
          </select>
        {/if}
        <button
          class="btn btn-outline mr-1"
          type="button"
          on:click={showLogTypePage}
        >
          <Checklist16 />
          {$_('LogView.LogDefBtn')}
        </button>
        <button
          class="btn  btn-outline mr-1"
          type="button"
          on:click={() => {
            page = "result";
          }}
        >
          <Check16 />
          {$_('LogView.ResultBtn')}
        </button>
        <button class="btn  btn-secondary mr-1" type="button" on:click={back}>
          <Reply16 />
          {$_('LogView.BackBtn')}
        </button>
        <button class="btn  btn-secondary" type="button" on:click={end}>
          <X16 />
          {$_('LogView.EndBtn')}
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
  #infoMsg {
    position: absolute;
    z-index: 100;
    bottom: 20px;
    right: 20px;
  }
</style>
