<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from 'svelte';
  import numeral from 'numeral';
  import {
    getFieldName,
    getFieldType,
    getFieldUnit,
    isFieldValid,
  } from "../../js/define";

  export let indexInfo;
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
      }
    });
  };

  onMount(() => {
    getProcessInfo();
  });

  const back = () => {
    dispatch("done", {});
  }

</script>

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
            <th>総数</th>
            <td>{numeral(indexInfo.Total).format('0,0')}</td>
          </tr>
          <tr>
            <th>処理時間</th>
            <td>{indexInfo.Duration}</td>
          </tr>
        </tbody>
      </table>
    </div>
    {#if indexInfo.Fields.length > 0}
      <div class="Box-row markdown-body log">
        <h5>抽出した項目</h5>
        <table class="fields">
          <thead>
            <tr>
              <th>変数名</th>
              <th>名前</th>
              <th>種別</th>
              <th>単位</th>
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
    <div class="Box-row markdown-body log">
      <h5>読み込んだファイル</h5>
      <table>
        <thead>
          <tr>
            <th>有効率</th>
            <th>完了</th>
            <th>対象</th>
            <th>処理時間</th>
            <th>サイズ</th>
            <th>パス</th>
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
    <div class="Box-footer text-right">
      <button class="btn" type="button" on:click={back}>
        <X16 />
        戻る
      </button>
    </div>
</div>
