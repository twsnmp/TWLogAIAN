<script>
  import { X16, FileDirectory16, Check16 } from "svelte-octicons";
  import { createEventDispatcher,onMount } from "svelte";
  import { _ } from '../../i18n/i18n';
  import {GetLastWorkDirs,SetWorkDir,SelectFile} from '../../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();
  let workdir = "";
  let ErrorMsg = "";
  let lastWorkDirs = [];
  let selLast = "";
 
  onMount(async () => {
    const wds = await GetLastWorkDirs();
    lastWorkDirs = wds || [];
    if (lastWorkDirs.length > 0 && workdir == "") {
      workdir = lastWorkDirs[0];
    }
  });

  const setWorkDir = async () => {
    if (!workdir) {
      ErrorMsg = $_('WorkDir.SelectWorkDirMsg');
      return;
    }
    const r = await SetWorkDir(workdir);
    if (r === "") {
      dispatch("done", { page: "setting" });
    } else {
      ErrorMsg = r || "";
    }
  };

  const selectWorkDir = async () => {
    workdir = await SelectFile("work",$_('WorkDir.WorkDir'));
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
