<script>
  import { X16, Question16, Copy16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { onMount } from "svelte";
  import { GetAIConfigs, AskAIAboutLog } from "../../../wailsjs/go/main/App";
  import { copyText } from "svelte-copy";
  import { _ } from '../../i18n/i18n';

  export let log = "";

  let errorMsg = "";
  let infoMsg = "";
  let answer = "";

  const dispatch = createEventDispatcher();

  let aiConfigID = "";
  let prompt = "";
  let limit = 2;

  let aiConfigList = [];
  let busy = false;

  onMount(async () => {
    aiConfigList = await GetAIConfigs();
  });

  const askAI = async () => {
    busy = true;
    const r = await AskAIAboutLog(aiConfigID, prompt, log, limit * 1);
    busy = false;
    if (r) {
      errorMsg = r.Error;
      answer = r.Answer;
      return;
    }
    errorMsg = $_('AI.NoAnswer');
  };

  const close = () => {
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg = "";
  };

  const copy = async () => {
    await copyText(answer);
    infoMsg = $_("LogView.Copied");
    setTimeout(() => {
      infoMsg = "";
    }, 2000);
  };
</script>

{#if busy}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">{$_('AI.AskAI')}</h3>
      <div class="flash mt-2">
        {$_('AI.AIThinking')}
        <span class="AnimatedEllipsis" />
      </div>
    </div>
  </div>
{:else if answer}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">{$_('AI.Answer')}</h3>
    </div>
    {#if infoMsg != ""}
      <div class="flash">
        {infoMsg}
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
      <pre style="text-wrap: wrap;">{answer}</pre>
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-secondary mr-1" type="button" on:click={copy}>
        <Copy16 />
        {$_('AI.Copy')}
      </button>
      <button class="btn btn-secondary mr-1" type="button" on:click={close}>
        <X16 />
        {$_('AI.Close')}
      </button>
    </div>
  </div>
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">{$_('AI.AskAI')}</h3>
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
          <h5 class="pb-1">{$_('AI.Class')}</h5>
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
          <h5 class="pb-1">{$_('AI.Quattion')}</h5>
        </div>
        <div class="form-group-body">
          <textarea
            class="form-control"
            style="width: 100%;"
            placeholder={$_('AI.QContent')}
            bind:value={prompt}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('AI.Log')}</h5>
        </div>
        <div class="form-group-body">
          <textarea
            class="form-control"
            style="width: 100%;"
            placeholder={$_('AI.Log')}
            bind:value={log}
          />
        </div>
      </div>
      <div class="Box-footer text-right">
        <button class="btn btn-primary" type="button" on:click={askAI}>
          <Question16 />
          {$_('AI.Ask')}
        </button>
        <button class="btn btn-secondary mr-1" type="button" on:click={close}>
          <X16 />
          {$_('AI.Close')}
        </button>
      </div>
    </div>
  </div>
{/if}
