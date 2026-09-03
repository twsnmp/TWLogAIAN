<script>
  import { X16, Check16, ArrowLeft16, Search16 } from "svelte-octicons";
  import { createEventDispatcher, onMount } from "svelte";
  import numeral from "numeral";
  import { _ } from "../../i18n/i18n";
  import { GetProcessInfo, Stop } from "../../../wailsjs/go/main/App";

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let warnMsg = "";
  let logFiles = [];
  let timer;
  let isDone = false;
  let readLines = 0;

  const getProcessInfo = async () => {
    const r = await GetProcessInfo();
    if (r) {
      logFiles = r.LogFiles || [];
      readLines = r.ReadLines || 0;
      if (r.ErrorMsg) {
        errorMsg = r.ErrorMsg;
      }
      if (r.IntLogFiles) {
        r.IntLogFiles.forEach((lf) => {
          logFiles.push(lf);
        });
      }
      if (r.Done) {
        isDone = true;
        if (r.ErrorMsg) {
          // エラーが発生している場合は自動遷移せず停止して確認できるようにする
          return;
        }
        if (readLines === 0 && logFiles.length > 0) {
          // ログが0件の場合も警告を表示して確認できるようにする
          warnMsg = $_("Processing.NoLogsRead");
          return;
        }
        dispatch("done", { page: "logview" });
        return;
      }
      timer = setTimeout(getProcessInfo, 1000);
    }
  };

  onMount(() => {
    getProcessInfo();
    return () => {
      if (timer) {
        clearTimeout(timer);
      }
    };
  });

  const stop = async () => {
    if (timer) {
      clearTimeout(timer);
    }
    const r = await Stop();
    if (r === "") {
      dispatch("done", { page: "setting" });
    } else {
      errorMsg = r;
    }
  };

  const backToSetting = () => {
    if (timer) {
      clearTimeout(timer);
    }
    dispatch("done", { page: "setting" });
  };

  const goToLogView = () => {
    if (timer) {
      clearTimeout(timer);
    }
    dispatch("done", { page: "logview" });
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header">
    <h3 class="Box-title">
      {$_("Processing.Title")}
      {#if !isDone}
        <span class="AnimatedEllipsis"></span>
      {/if}
    </h3>
  </div>
  {#if errorMsg != ""}
    <div class="flash flash-error">
      <strong>{$_("Processing.CompletedWithErrors")}: </strong>
      {errorMsg}
    </div>
  {/if}
  {#if warnMsg != ""}
    <div class="flash flash-warn">
      {warnMsg}
    </div>
  {/if}
  <div class="Box-body markdown-body">
    <table width="100%">
      <thead>
        <tr>
          <th width="8%">{$_("Processing.Rate")}</th>
          <th width="8%">{$_("Processing.Done")}</th>
          <th width="8%">{$_("Processing.Target")}</th>
          <th width="8%">{$_("Processing.Time")}</th>
          <th width="8%">{$_("Processing.Size")}</th>
          <th width="15%">{$_("Processing.GrokPat")}</th>
          <th width="45%">{$_("Processing.Path")}</th>
        </tr>
      </thead>
      <tbody>
        {#each logFiles as f}
          <tr>
            <td
              class:color-fg-danger={(f.Read ? (100.0 * f.Send) / f.Read : 100) <
                50.0}
            >
              {f.Read ? ((100.0 * f.Send) / f.Read).toFixed(2) : 0}%
            </td>
            <td>{numeral(f.Read).format("0.00b")}</td>
            <td>{numeral(f.Send).format("0.00b")}</td>
            <td>{f.Duration || ""}</td>
            <td>{numeral(f.Size).format("0.00b")}</td>
            <td>{f.ETName || ""}</td>
            <td>
              {f.LogSrc && (f.LogSrc.Type == "scp" || f.LogSrc.Type == "ssh" || f.LogSrc.Type == "ftp")
                ? f.LogSrc.Server + ":" + f.Path
                : f.Path}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <div class="Box-footer d-flex flex-justify-between">
    <div>
      {#if !isDone}
        <button class="btn btn-danger" type="button" on:click={stop}>
          <X16 />
          {$_("Processing.StopBtn")}
        </button>
      {:else}
        <button class="btn btn-secondary" type="button" on:click={backToSetting}>
          <ArrowLeft16 />
          {$_("Processing.BackToSetting")}
        </button>
      {/if}
    </div>
    {#if isDone}
      <div>
        <button class="btn btn-primary" type="button" on:click={goToLogView}>
          <Search16 />
          {$_("Processing.GoToLogView")}
        </button>
      </div>
    {/if}
  </div>
</div>
