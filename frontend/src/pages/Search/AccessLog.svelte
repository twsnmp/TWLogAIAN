<script>
  import { X16, Search16,TriangleDown16,TriangleUp16,Check16,Trash16 } from "svelte-octicons";
  import Query from "./Query.svelte"
  import Result from "./Result.svelte"
  import { createEventDispatcher,onMount, tick } from "svelte";
  import {makeLogCountChart,updateLogCountChart} from "../../js/logchart";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import * as echarts from 'echarts';
  import { html } from "gridjs";

  const dispatch = createEventDispatcher();
  let page = "";
  let showQuery = false;
  const conf = {
    query: '',
    limit: 1000,
    history: [],
    keyword: {
      field: "",
      mode: "",
      key: "",
    },
    number: {
      field: "",
      oper: "<",
      value: "0.0",
    },
    range:{
      start: "",
      end: "",
    },
    geo:{
      lat: "",
      long: "",
      range: "",
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
          console.log(l);
          let cl = l.KeyValue.clientip;
          if (l.KeyValue.clientip_host) {
            cl += "(" + l.KeyValue.clientip_host +")"
          }
          let country = l.KeyValue.clientip_geo ? l.KeyValue.clientip_geo.Country :"";
          d.push([
            l.KeyValue.response,
            l.Time,
            l.KeyValue.verb,
            l.KeyValue.bytes,
            cl,
            country,
            l.KeyValue.request,
          ]);
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
        conf.history.push(conf.query);
      }
    });
  };

  onMount(() => {
    makeLogCountChart("chart");
  });

  const formatCode = (code) => {
    if (code < 300) {
      return html(`<div class="color-fg-default">${code}</div>`);
    } else if (code < 400) {
      return html(`<div class="color-fg-attention">${code}</div>`);
    } else if (code < 500) {
      return html(`<div class="color-fg-danger">${code}</div>`);
    }
    return html(`<div class="color-fg-danger-emphasis">${code}</div>`);
  }

  const columns = [
    {
      name: "応答",
      width: "8%",
      formatter: (cell) => formatCode(cell),
    },{
      name: "日時",
      width: "15%",
      formatter: (cell) => echarts.time.format(new Date(cell/(1000*1000)), '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'),
    },{
      name: "リクエスト",
      width: "10%",
    },{
      name: "サイズ",
      width: "8%",
    },{
      name: "アクセス元",
      width: "25%",
    },{
      name: "国",
      width: "8%",
    },{
      name: "パス",
      width: "26%",
    },
  ];

  const cancel = () => {
    window.go.main.App.CloseWorkDir();
    dispatch("done", { page: "wellcome" });
  };


  const clearMsg = () => {
    errorMsg = "";
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

  const handleUpdateQuery = (e) => {
    if (e && e.detail && e.detail.query) {
      if (e.detail.add) {
        conf.query += e.detail.query;
      } else {
        conf.query = e.detail.query;
      }
    }
  }

  const clear = () => {
    conf.query = "";
  }

</script>

<svelte:window on:resize={onResize} />
{#if page == "result"}
  <Result {indexInfo} on:done={handleDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
      <div class="Box-header d-flex flex-items-center">
        <h3 class="Box-title overflow-hidden flex-auto">アクセスログ分析</h3>
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
              <button class="btn  btn-secondary" type="button" on:click={() => { showQuery= true}}>
                <TriangleDown16 />
              </button>
            {:else}
              <button class="btn  btn-secondary" type="button" on:click={() => { showQuery= false}}>
                <TriangleUp16 />
              </button>
            {/if}
            <button class="btn  btn-primary ml-2" type="button" on:click={search}>
              <Search16 />
              検索
            </button>
          </div>
        </div>
      </div>
      {#if showQuery}
      <div class="Box-row">
        <Query {conf} fields={indexInfo.Fields} on:update={handleUpdateQuery}/>
      </div>
      {/if}
      <div class="Box-row">
        <div id="chart" />
      </div>
      <div class="Box-row markdown-body log">
        <Grid {data} sort search {pagination} {columns} language={jaJP} />
      </div>
      <div class="Box-footer text-right">
        <button class="btn  btn-secondary" type="button" on:click={()=> { page = "result"}}>
          <Check16 />
          処理結果
        </button>
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
