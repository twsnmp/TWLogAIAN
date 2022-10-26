<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher,tick } from "svelte";
  import { showLogHeatmap, resizeLogHeatmap} from "./logheatmap";
  import AutoEncoder from "./AutoEncoder.svelte";
  import { onMount } from 'svelte';
  import numeral from 'numeral';
  import {
    getFieldName,
    getFieldType,
    getFieldUnit,
    isFieldValid,
  } from "../../js/define";
  import * as echarts from "echarts";

  export let indexInfo;
  export let aecdata;
  export let dark = false;
  let readLines = 0;
  let skipLines = 0;
  let startTime = "";
  let endTime = "";

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let logFiles = [];

  const getProcessInfo = () => {
    window.go.main.App.GetProcessInfo().then((r) => {
      if (r) {
        if (r.LogFiles) {
          logFiles = r.LogFiles
        }
        if (r.ErrorMsg) {
          errorMsg = r.ErrorMsg;
        }
        readLines = r.ReadLines;
        skipLines = r.SkipLines;
        startTime = echarts.time.format(new Date(r.StartTime / (1000 * 1000)),"{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}");
        endTime = echarts.time.format(new Date(r.EndTime / (1000 * 1000)),"{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}");
        window.go.main.App.GetDark().then((dark) => {
          showChart(r.TimeLine,dark);
        });
      }
    });
  };

  const showChart =  async (timeLine,dark)  => {
    await tick();
    showLogHeatmap("chart",timeLine,dark);
  } 

  onMount(() => {
    getProcessInfo();
  });

  const back = () => {
    dispatch("done", {});
  }

  const onResize = () => {
    resizeLogHeatmap();
  };

</script>

<svelte:window on:resize={onResize} />

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">ログ読み込み結果</h3>
    </div>
    {#if errorMsg != ""}
      <div class="flash flash-error">
        {errorMsg}
      </div>
    {/if}
    <div class="Box-row markdown-body log">
      <h5>概要</h5>
      <table>
        <tbody>
          <tr>
            <th>インデックス上の総数</th>
            <td>{numeral(indexInfo.Total).format('0,0')}</td>
          </tr>
          <tr>
            <th>インデックス作成時間</th>
            <td>{indexInfo.Duration}</td>
          </tr>
          <tr>
            <th>処理したライン数</th>
            <td>{numeral(readLines).format('0,0')}</td>
          </tr>
          <tr>
            <th>スキップしたライン数</th>
            <td>{numeral(skipLines).format('0,0')}</td>
          </tr>
          <tr>
            <th>最初のログの日時</th>
            <td>{startTime}</td>
          </tr>
          <tr>
            <th>最後のログの日時</th>
            <td>{endTime}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="Box-row">
      <div id="chart" />
    </div>
    <div class="Box-row markdown-body log">
      <h5>読み込んだファイル</h5>
      <table width="100%">
        <thead>
          <tr>
            <th width="8%">有効率</th>
            <th width="8%">完了</th>
            <th width="8%">対象</th>
            <th width="8%">処理時間</th>
            <th width="8%">サイズ</th>
            <th width="60%">パス</th>
          </tr>
        </thead>
        <tbody>
        {#each logFiles as f }
          <tr>
            <td class:color-fg-danger={(f.Read ? (100.0 * f.Send/f.Read) : 100) < 50.0}>{f.Read ? (100.0 * f.Send/f.Read).toFixed(2) : 0}%</td>
            <td>{numeral(f.Read).format('0.00b')}</td>
            <td>{numeral(f.Send).format('0.00b')}</td>
            <td>{f.Duration}</td>
            <td>{numeral(f.Size).format('0.00b')}</td>
            <td>{(f.LogSrc.Type == "scp" || f.LogSrc.Type == "ssh") ? f.LogSrc.Server + ":" + f.Path : f.Path}</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
    {#if indexInfo.Fields.length > 0}
      <div class="Box-row markdown-body log">
        <h5>抽出した項目</h5>
        <table width="90%">
          <thead>
            <tr>
              <th width="30%">変数名</th>
              <th width="40%">名前</th>
              <th width="20%">種別</th>
              <th width="10%">単位</th>
            </tr>
          </thead>
          <tbody>
            {#each indexInfo.Fields as f}
              <tr>
                <td>{f}</td>
                <td class:color-fg-danger={!isFieldValid(f)}>{getFieldName(f)}</td>
                <td>{getFieldType(f)}</td>
                <td>{getFieldUnit(f)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    {#if aecdata.length > 0}
      <div class="Box-row">
        <h5>Auto Encoderの学習状況</h5>
        <AutoEncoder {dark} chartData={aecdata} />
      </div>
    {/if}
    <div class="Box-footer text-right">
      <button class="btn" type="button" on:click={back}>
        <X16 />
        戻る
      </button>
    </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 200px;
    margin: 5px auto;
  }
</style>
