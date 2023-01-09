<script>
  import { X16, FileDirectory16, Check16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { _ } from '../../i18n/i18n';
  import {GetLastWorkDirs,SetWorkDir,SelectFile} from '../../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();
  let workdir = "";
  let ErrorMsg = "";
  let lastWorkDirs = [];
  let selLast = "";

  GetLastWorkDirs().then((wds) => {
    lastWorkDirs = wds;
    if (lastWorkDirs.length > 0 && workdir == "") {
      workdir = lastWorkDirs[0];
    }
  });
  const setWorkDir = () => {
    if (!workdir) {
      ErrorMsg = $_('WorkDir.SelectWorkDirMsg');
      return;
    }
    SetWorkDir(workdir).then((r) => {
      if (r === "") {
        dispatch("done", { page: "setting" });
      } else {
        ErrorMsg = r;
      }
    });
  };
  const selectWorkDir = () => {
    SelectFile("work",$_('WorkDir.WorkDir')).then((d) => {
      workdir = d;
    });
  };
  const cancel = () => {
    dispatch("done", { page: "wellcome" });
  };
  const checkSelectWorkDir = () => {
    if (selLast != "") {
      workdir = selLast;
    }
  };
  const clearMsg = () => {
    ErrorMsg = "";
  };
</script>

<div class="Box mx-auto" style="max-width: 800px;">
  <div class="Box-header">
    <h3 class="Box-title">{$_('WorkDir.Title')}</h3>
  </div>
  {#if ErrorMsg != ""}
    <div class="flash flash-error">
      {ErrorMsg}
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
  <div class="Box-body">
    <div class="input-group mb-5">
      <input
        class="form-control"
        type="text"
        placeholder="{$_('WorkDir.WorkDir')}"
        aria-label="{$_('WorkDir.WorkDir')}"
        bind:value={workdir}
      />
      <span class="input-group-button">
        <button class="btn" type="button" on:click={selectWorkDir}>
          <FileDirectory16 />
        </button>
      </span>
    </div>
    {#if lastWorkDirs.length > 0}
      <p>{$_('WorkDir.History')}</p>
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        class="form-select"
        aria-label="{$_('WorkDir.History')}"
        bind:value={selLast}
        on:change={checkSelectWorkDir}
      >
        <option value="">{$_('WorkDir.SelectMsg')}</option>
        {#each lastWorkDirs as d}
          <option value={d}>{d}</option>
        {/each}
      </select>
    {/if}
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-secondary mr-1" type="button" on:click={cancel}>
      <X16 />
      {$_('WorkDir.CancelBtn')}
    </button>
    <button class="btn btn-primary" type="button" on:click={setWorkDir}>
      <Check16 />
      {$_('WorkDir.SelectBtn')}
    </button>
  </div>
</div>
