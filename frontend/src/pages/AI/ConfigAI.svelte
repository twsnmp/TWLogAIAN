<script>
  import { X16, Plus16, Check16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import { onMount } from "svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _, getLocale } from "../../i18n/i18n";
  import {
    GetAIConfigs,
    DeleteAIConfig,
    AddAIConfig,
    TestAIConfigByID,
    SyncAIConfig,
  } from "../../../wailsjs/go/main/App";

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  const dispatch = createEventDispatcher();

  let addAIConfig = false;

  let aiConfigList = [];
  let gridSearch = true;
  let pagination = {
    enable: true,
    limit: 10,
  };

  onMount(() => {
    updateAIConfigs();
  });

  const updateAIConfigs = async () => {
    const r = await GetAIConfigs();
    aiConfigList = [];
    if (r) {
      for (const ac of r) {
        aiConfigList.push([
          ac.ID,
          ac.WeaviateURL,
          ac.ClassName,
          ac.OllamaURL,
          ac.GenerativeModel,
          ac.Text2vecModel,
        ]);
      }
    }
  };

  let aiConfig = {
    ID: "",
    WeaviateURL: "",
    OllamaURL: "",
    Text2vecModel: "",
    GenerativeModel: "",
    ClassName: "",
  };

  let busy = false;
  let errorMsg = "";
  let infoMsg = "";

  const testAIConigButton = (_, row) => {
    const id = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => testAIConifg(id),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"></path></svg>`
      )
    );
  };

  const copyAIConfigButton = (_, row) => {
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => copyAIConfig(row),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>`
      )
    );
  };

  const testAIConifg = async (id) => {
    busy = true;
    const r = await TestAIConfigByID(id);
    busy = false;
    console.log(r);
    if (r != "") {
      errorMsg = r;
    } else {
      infoMsg = $_('AI.ConnectOK');
      setTimeout(()=>{infoMsg =""},3000);
    }
  };

  const deleteAIConfigButton = (_, row) => {
    const id = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm btn-danger",
        onClick: () => deleteAIConfig(id),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.576l.66-6.6a.75.75 0 00-1.492-.149l-.66 6.6a.25.25 0 01-.249.225h-5.19a.25.25 0 01-.249-.225l-.66-6.6z"></path></svg>`
      )
    );
  };

  const deleteAIConfig = async (id) => {
    busy = true;
    const r = await DeleteAIConfig(id,$_('AI.DeleteAIConfig'),$_('AI.DeleteAIConfigMsg'));
    busy = false;
    if (r != "") {
      if (r != "No") {
        errorMsg = r;
      }
    } else {
      infoMsg = $_('AI.DeleteAIConfigDone');
      updateAIConfigs();
      setTimeout(()=>{infoMsg =""},3000);
    }
  };

  const columns = [
    {
      name: "ID",
      width: "14%",
    },
    {
      name: "Weaviate URL",
      width: "20%",
    },
    {
      name: $_('AI.ClassName'),
      width: "10%",
    },
    {
      name: "Ollama URL",
      width: "20%",
    },
    {
      name: $_('AI.VectorModel'),
      width: "10%",
    },
    {
      name: $_('AI.GenModel'),
      width: "10%",
    },
    {
      name: $_('AI.Test'),
      sort: false,
      width: "6%",
      formatter: testAIConigButton,
    },
    {
      name: $_('AI.Copy'),
      sort: false,
      width: "6%",
      formatter: copyAIConfigButton,
    },
    {
      name: $_("LogType.Delete"),
      sort: false,
      width: "5%",
      formatter: deleteAIConfigButton,
    },
  ];

  const showAddAIConfig = () => {
    aiConfig = {
      ID: "",
      WeaviateURL: "http://localhost:8080",
      OllamaURL: "http://host.docker.internal:11434",
      Text2vecModel: "nomic-embed-text",
      GenerativeModel: "llama3.2",
      ClassName: "",
    };
    addAIConfig = true;
  };

  const copyAIConfig = (row) => {
    aiConfig = {
      ID: "",
      WeaviateURL: row.cells[1].data,
      ClassName: row.cells[2].data + "_copy",
      OllamaURL: row.cells[3].data,
      GenerativeModel: row.cells[4].data,
      Text2vecModel: row.cells[5].data,
    };
    addAIConfig = true;
    syncAIConfig = false;
  };

  let syncAIConfig = false;

  const add = async () => {
    busy = true;
    errorMsg = "";
    if (syncAIConfig) {
      errorMsg = (await SyncAIConfig(aiConfig.WeaviateURL)) || "";
    } else {
      errorMsg = (await AddAIConfig(aiConfig)) || "";
    }
    busy = false;
    if (errorMsg != "") {
      return;
    }
    addAIConfig = false;
    updateAIConfigs();
  };

  const close = () => {
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg = "";
  };
</script>

{#if addAIConfig}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">{$_('LogView.AIConfig')}</h3>
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
          <h5 class="pb-1">Weaviate URL</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            style="width: 100%;"
            placeholder="URL"
            bind:value={aiConfig.WeaviateURL}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('AI.Sync')}</h5>
        </div>
        <div class="form-group-body">
          <div class="form-checkbox">
            <label>
              <input
                type="checkbox"
                bind:checked={syncAIConfig}
                aria-describedby="help-text-for-sync"
              />
              {$_('AI.SyncToWeaviate')}
            </label>
            <p class="note" id="help-text-for-sync">
              {$_('AI.AddClassFromWeaviate')}
            </p>
          </div>
        </div>
      </div>
      {#if !syncAIConfig}
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">Ollama URL</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              style="width: 100%;"
              placeholder="URL"
              bind:value={aiConfig.OllamaURL}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('AI.ClassName')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              style="width: 100%;"
              placeholder={$_('AI.ClassName')}
              bind:value={aiConfig.ClassName}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('AI.VectorModelName')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              style="width: 100%;"
              placeholder={$_('AI.ModelName')}
              bind:value={aiConfig.Text2vecModel}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('AI.GenModelName')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              style="width: 100%;"
              placeholder={$_('AI.ModelName')}
              bind:value={aiConfig.Text2vecModel}
            />
          </div>
        </div>
      {/if}
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-primary" type="button" on:click={add}>
        <Check16 />
        {$_('AI.AddBtn')}
      </button>
      <button
        class="btn btn-secondary mr-1"
        type="button"
        on:click={() => {
          addAIConfig = false;
        }}
      >
        <X16 />
        {$_('AI.Calcel')}
      </button>
    </div>
  </div>
{:else if busy}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">{$_('LogView.AIConfig')}</h3>
      <div class="flash mt-2">
        {$_('AI.Processing')}
        <span class="AnimatedEllipsis" />
      </div>
    </div>
  </div>
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">
        {$_('LogView.AIConfig')}
        <button
          class="btn btn-sm float-right"
          type="button"
          on:click={showAddAIConfig}
        >
          <Plus16 />
        </button>
      </h3>
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
      <div class="markdown-body mt-3">
        <Grid
          data={aiConfigList}
          sort
          resizable
          search={gridSearch}
          {pagination}
          {columns}
          language={gridLang}
        />
      </div>
    </div>
    <div class="Box-footer text-right">
      <button class="btn btn-secondary mr-1" type="button" on:click={close}>
        <X16 />
        {$_("LogType.CloseBtn")}
      </button>
    </div>
  </div>
{/if}
