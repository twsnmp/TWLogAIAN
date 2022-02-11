<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from 'svelte';

  export let indexInfo;
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let logFiles = [];

  const getProcessInfo = () => {
    window.go.main.App.GetProcessInfo().then((r) => {
      if (r) {
        logFiles = r.LogFiles
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
    <div class="Box-body">
      <div class="markdown-body">
        <table>
          <tbody>
            <tr>
              <th>総数</th>
              <td>{indexInfo.Total}</td>
            </tr>
            <tr>
              <th>処理時間</th>
              <td>{indexInfo.Duration}</td>
            </tr>
            <tr>
              <th>抽出項目</th>
              <td>
                {#each indexInfo.Fields as f,i}
                  {f}
                  {#if (i+1) % 10 == 0}
                    <br>
                  {:else}
                    ,
                  {/if}
                {/each}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="mt-2 markdown-body">
        <table>
          <thead>
            <tr>
              <th>有効率</th>
              <th>完了</th>
              <th>対象</th>
              <th>処理時間</th>
              <th>サイズ</th>
              <th>パス/URL</th>
            </tr>
          </thead>
          <tbody>
          {#each logFiles as f }
            <tr>
              <td class:color-fg-danger={(f.Read ? (100.0 * f.Send/f.Read) : 100) < 50.0}>{f.Read ? (100.0 * f.Send/f.Read).toFixed(2) : 0}%</td>
              <td>{f.Read}</td>
              <td>{f.Send}</td>
              <td>{f.Duration}</td>
              <td>{f.Size}</td>
              <td>{f.URL}</td>
            </tr>
          {/each}
          </tbody>
        </table>
      </div>
    </div>
    <div class="Box-footer text-right">
      <button class="btn" type="button" on:click={back}>
        <X16 />
        戻る
      </button>
    </div>
</div>
