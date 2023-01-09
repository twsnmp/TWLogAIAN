<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from 'svelte';
  import numeral from 'numeral';
  import { _ } from '../../i18n/i18n';
  import {GetProcessInfo,Stop} from '../../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let logFiles = [];
  let timer;
  const getProcessInfo = () => {
    GetProcessInfo().then((r) => {
      if (r) {
        logFiles = r.LogFiles
        if (r.ErrorMsg) {
          errorMsg = r.ErrorMsg;
        }
        if (r.IntLogFiles) {
          r.IntLogFiles.forEach((lf) => {
            logFiles.push(lf);
          });
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
    Stop().then((r) => {
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
      <h3 class="Box-title">{$_('Processing.Title')}<span class="AnimatedEllipsis"></span></h3>
    </div>
    {#if errorMsg != ""}
      <div class="flash flash-error">
        {errorMsg}
      </div>
    {/if}
    <div class="Box-body markdown-body">
      <table width="100%">
        <thead>
          <tr>
            <th width="8%">{$_('Processing.Rate')}</th>
            <th width="8%">{$_('Processing.Done')}</th>
            <th width="8%">{$_('Processing.Target')}</th>
            <th width="8%">{$_('Processing.Time')}</th>
            <th width="8%">{$_('Processing.Size')}</th>
            <th width="15%">{$_('Processing.GrokPat')}</th>
            <th width="45%">{$_('Processing.Path')}</th>
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
            <td>{f.ETName}</td>
            <td>{(f.LogSrc.Type == "scp" || f.LogSrc.Type == "ssh") ? f.LogSrc.Server + ":" + f.Path : f.Path}</td>
          </tr>
        {/each}
        </tbody>
      </table>
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-danger" type="button" on:click={stop}>
        <X16 />
        {$_('Processing.StopBtn')}
      </button>
    </div>
</div>
