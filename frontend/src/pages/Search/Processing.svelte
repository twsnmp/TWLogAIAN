<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from 'svelte';
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let logFiles = [];
  let timer;
  const getProcessInfo = () => {
    window.go.main.App.GetProcessInfo().then((r) => {
      if (r) {
        logFiles = r.LogFiles
        if (r.ErrorMsg) {
          errorMsg = r.ErrorMsg;
        }
        if (r.Done) {
          dispatch("done", { page: "logview" });
          return
        }
        timer = setTimeout(getProcessInfo,1000);
      }
    });
  };

  onMount(() => {
    getProcessInfo();
  });

  const stop = () => {
    // Index作成を停止
    window.go.main.App.Stop().then((r) => {
      if (r === "") {
        clearTimeout(timer);
        dispatch("done", { page: "setting" });
      } else {
        errorMsg = r;
      }
    });
  };

</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">ログ読み込み処理中....</h3>
    </div>
    {#if errorMsg != ""}
      <div class="flash flash-error">
        {errorMsg}
      </div>
    {/if}
    <div class="Box-body markdown-body">
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
            <td>{f.Read}</td>
            <td>{f.Send}</td>
            <td>{f.Duration}</td>
            <td>{f.Size}</td>
            <td>{f.Path}</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-danger" type="button" on:click={stop}>
        <X16 />
        停止
      </button>
    </div>
</div>
