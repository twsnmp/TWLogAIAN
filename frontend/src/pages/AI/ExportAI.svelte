<script>
  import { X16, Upload16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from "svelte";
  import { GetAIConfigs, ExportAILog } from "../../../wailsjs/go/main/App";
  import { _ } from "../../i18n/i18n";

  export let logs = [];

  let errorMsg = "";

  const dispatch = createEventDispatcher();

  let aiConfigID = "";
  let aiExport = {
    Category: "",
    Descr: "",
    Logs: [],
  };

  let aiConfigList = [];
  let busy = false;

  onMount(async () => {
    aiConfigList = await GetAIConfigs();
  });

  const exportLog = async () => {
    busy = true;
    aiExport.Logs = logs;
    const r = await ExportAILog(aiConfigID, aiExport);
    busy = false;
    if (r != "") {
      errorMsg = r;
      return;
    }
    close();
  };

  const close = () => {
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
  };
</script>

{#if busy}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">{$_('AI.Export')}</h3>
      <div class="flash mt-2">
        {$_('AI.Exporting')}
        <span class="AnimatedEllipsis" />
      </div>
    </div>
  </div>
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">{$_('AI.Export')}</h3>
    </div>
    {#if errorMsg != ""}
      <div class="flash flash-error">
        {errorMsg}
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
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('AI.ExportTraget')}</h5>
        </div>
        <div class="form-group-body">
          <select class="form-select ml-2" bind:value={aiConfigID}>
            {#each aiConfigList as aic}
              <option value={aic.ID}
                >{aic.WeaviateURL + "/" + aic.ClassName}</option
              >
            {/each}
          </select>
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('AI.LogCategory')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            style="width: 100%;"
            placeholder="URL"
            bind:value={aiExport.Category}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('AI.LogDesr')}</h5>
        </div>
        <div class="form-group-body">
          <textarea
            class="form-control"
            style="width: 100%;"
            placeholder={$_('AI.LogDescrMsg')}
            bind:value={aiExport.Descr}
          />
        </div>
      </div>
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-primary" type="button" on:click={exportLog}>
        <Upload16 />
        {$_('AI.ExportBtn')}
      </button>
      <button class="btn btn-secondary mr-1" type="button" on:click={close}>
        <X16 />
        {$_('AI.Calcel')}
      </button>
    </div>
  </div>
{/if}
