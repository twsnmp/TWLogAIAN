<script>
  import highlightWords from "highlight-words";
  import { X16, Check16, StarFill16, Reply16, Plus16,EyeClosed16,Eye16,TriangleDown16,TriangleUp16 } from "svelte-octicons";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { createEventDispatcher } from "svelte";
  import {
    getFieldName,
    getFieldType,
    getFieldUnit,
    isFieldValid,
    loadFieldTypes,
  } from "../../js/define";
  import { onMount } from "svelte";
  import EditFieldType from "./EditFieldType.svelte";
  import { _,getLocale } from '../../i18n/i18n';
  import {SaveExtractorType,TestGrok,AutoGrok} from '../../../wailsjs/go/main/App';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let extractorType;
  export let add = true;
  export let testLog = "";
  
  let data = [];
  let fields = [];
  let columns = [];
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let selected = '';
  let showInput = false;
  let showReplace = false;

  onMount(() => {
    loadFieldTypes();
    document.addEventListener(`selectionchange`, () => {
      const s = document.getSelection().toString();
      if (s == "") {
        showReplace = false;
        return;
      }
      const si = extractorType.Grok.indexOf(s);
      const li = extractorType.Grok.lastIndexOf(s);
      showReplace = si > 0 && (si == li);
      if (showReplace) {
        selected = s;
      }
    });
  });

  const back = () => {
    dispatch("done", {});
  };

  const save = () => {
    SaveExtractorType(extractorType).then((r) => {
      errorMsg = r;
      if (r == "") {
        dispatch("done", { save: true });
      }
    });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const test = () => {
    if (extractorType.Grok == "") {
      errorMsg = $_('EditExtractorType.InputPatMsg');
      return;
    }
    if (testLog == "") {
      errorMsg = $_('EditExtractorType.InputTestDataMsg');
      return;
    }
    TestGrok(extractorType.Grok, testLog).then((r) => {
      errorMsg = r.ErrorMsg;
      data = r.Data;
      columns = [];
      fields = r.Fields;
      fields.forEach((e) => {
        columns.push(getFieldName(e));
      });
    });
  };

  const auto = () => {
    if (testLog == "") {
      errorMsg = $_('EditExtractorType.InputTestDataMsg');
      return;
    }
    AutoGrok(testLog).then((r) => {
      errorMsg = r.ErrorMsg;
      if (r.Grok) {
        if (extractorType.Grok) {
          oldGrok.push(extractorType.Grok);
          oldGrok = oldGrok;
        }
        extractorType.Grok = r.Grok;
      }
    });
  };

  let oldGrok = [];
  const resetGrok = () => {
    if (oldGrok.length > 0) {
      extractorType.Grok = oldGrok.pop();
      oldGrok = oldGrok;
    }
  };

  const getGrokPat = (s) => {
    if (s.match(/^\d{1,3}(\.\d{1,3}){3}$/)) {
      return "%{IPV4:ipv4}";
    }
    if (s.match(/^[0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){5}$/i)) {
      return "%{MAC:mac}";
    }
    if (s.match(/https?:\/\/[-_.!~*\'()a-zA-Z0-9;\/?:\@&=+\$,%#]+/)) {
      return "%{URI:url}";
    }
    if (
      s.match(/^[A-Za-z0-9]{1}[A-Za-z0-9_.-]*@{1}[A-Za-z0-9_.-]+.[A-Za-z0-9]+$/)
    ) {
      return "%{EMAILADDRESS:email}";
    }
    if (s.match(/^[+,-]?([1-9]\d*|0)$/)) {
      return "%{INT:int}";
    }
    if (s.match(/^[+,-]?([1-9]\d*|0)(\.\d+)?$/)) {
      return "%{NUMBER:number}";
    }
    if (s.match(/^\w+$/)) {
      return "%{WORD:word}";
    }
    if (s.match(/^\s+$/)) {
      return "\\s+";
    }
    return "%{GREEDYDATA:data}";
  };

  const replaceGrok = (w) => {
    if (selected == "") {
      return;
    }
    const r = w ? getGrokPat(selected) : '.+';
    if (extractorType.Grok) {
      oldGrok.push(extractorType.Grok);
      oldGrok = oldGrok;
    }
    extractorType.Grok = extractorType.Grok.replace(selected,r);
  };

  $: grokChunks = highlightWords({
    text: extractorType.Grok,
    query: /(%\{.+?\}|\.\+|\\s\+)/,
  });

  const getGrokClass = (c) => {
    if (!c.match) {
      return "";
    }
    if (c.text.includes("%{")) {
      return "color-fg-attention text-underline";
    }
    if (c.text.includes("s")) {
      return "color-fg-accent";
    }
    return "color-fg-danger";
  };

  let page = "";
  const addFT = true;
  let fieldType = {};
  const addFieldType = (key) => {
    fieldType = {
      Key: key,
      Name: getFieldName(key),
      Type: getFieldType(key),
      Unit: "",
      CanEdit: true,
    };
    page = "fieldType";
  };
  const handleEditFieldTypeDone = (e) => {
    if (e && e.detail && e.detail.save) {
      loadFieldTypes();
      test();
    }
    page = "";
  };
</script>

{#if page == "fieldType"}
  <EditFieldType {fieldType} add={addFT} on:done={handleEditFieldTypeDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">{$_('EditExtractorType.Title')}</h3>
    </div>
    {#if errorMsg}
      <div class="flash flash-error">
        {errorMsg}
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
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('EditExtractorType.Key')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            disabled={!add}
            placeholder="{$_('EditExtractorType.PatKey')}"
            bind:value={extractorType.Key}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('EditExtractorType.Name')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder="{$_('EditExtractorType.NameOfPat')}"
            bind:value={extractorType.Name}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('EditExtractorType.ExtractPat')}
          {#if showInput}
            <button type="button" class="btn btn-sm ml-2" on:click={()=> showInput = false}>
              <TriangleUp16 />
            </button>
          {:else}
            <button type="button" class="btn btn-sm ml-2" on:click={()=> showInput = true}>
              <TriangleDown16 />
            </button>
          {/if}
            {#if showReplace }
              <button type="button" class="btn btn-sm btn-danger ml-2" on:click={() => replaceGrok(true)}><Eye16 /></button>
              <button type="button" class="btn btn-sm btn-danger ml-2" on:click={() => replaceGrok(false)}><EyeClosed16 /></button>
            {/if}
            {#if oldGrok.length > 0 }
              <button type="button" class="btn btn-sm ml-2" on:click={resetGrok}><Reply16 /></button>
            {/if}
          </h5>
        </div>
      {#if showInput}
        <div class="form-group-body">
          <input
            class="form-control grok"
            type="text"
            placeholder="{$_('EditExtractorType.ExtractPat')}"
            bind:value={extractorType.Grok}
          />
        </div>
      {/if}
        <div class="form-group-body mt-1">
          {#each grokChunks as chunk}
            <span class={getGrokClass(chunk)}>{chunk.text}</span>
          {/each}
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">{$_('EditExtractorType.ExtractInfo')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            style="width: 15%;"
            placeholder="{$_('EditExtractorType.TimeStampItem')}"
            bind:value={extractorType.TimeField}
          />
          <input
            class="form-control"
            type="text"
            style="width: 25%;"
            placeholder="{$_('EditExtractorType.IPField')}"
            bind:value={extractorType.IPFields}
          />
          <input
            class="form-control"
            type="text"
            placeholder="{$_('EditExtractorType.MacField')}"
            style="width: 20%;"
            bind:value={extractorType.MACFields}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('EditExtractorType.TestData')}</h5>
        </div>
        <div class="form-group-body">
          <textarea class="form-control testdata" bind:value={testLog} />
        </div>
      </div>
    </div>
    {#if fields.length > 0}
      <div class="Box-row markdown-body log">
        <h5>{$_('EditExtractorType.ExtractFields')}</h5>
        <table class="fields" width="100%">
          <thead>
            <tr>
              <th width="20%">{$_('EditExtractorType.ValueName')}</th>
              <th width="40%">{$_('EditExtractorType.Name')}</th>
              <th width="20%">{$_('EditExtractorType.Type')}</th>
              <th width="10%">{$_('EditExtractorType.Unit')}</th>
              <th width="10%">{$_('EditExtractorType.Add')}</th>
            </tr>
          </thead>
          <tbody>
            {#each fields as f}
              <tr>
                <td>{f}</td>
                <td class:color-fg-danger={!isFieldValid(f)}
                  >{getFieldName(f)}</td
                >
                <td>{getFieldType(f)}</td>
                <td>{getFieldUnit(f)}</td>
                <td>
                  {#if !isFieldValid(f)}
                    <button
                      class="btn btn-sm"
                      type="button"
                      on:click={() => addFieldType(f)}
                    >
                      <Plus16 />
                    </button>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="Box-row markdown-body log">
        <h5>{$_('EditExtractorType.ExtractData')}</h5>
        <Grid {data} {columns} language={gridLang} />
      </div>
    {/if}
    <div class="Box-footer text-right">
      <button class="btn btn-danger mr-1" type="button" on:click={auto}>
        <StarFill16 />
        {$_('EditExtractorType.AutoPatGen')}
      </button>
      <button class="btn btn-primary" type="button" on:click={test}>
        <Check16 />
        {$_('EditExtractorType.Test')}
      </button>
      <button class="btn btn-secondary mr-1" type="button" on:click={back}>
        <X16 />
        {$_('EditExtractorType.CancelBtn')}
      </button>
      <button class="btn btn-primary mr-1" type="button" on:click={save}>
        <X16 />
        {$_('EditExtractorType.SaveBtn')}
      </button>
    </div>
  </div>
{/if}

<style>
  .form-group-body input.grok {
    width: 99%;
  }

  .form-group-body textarea.testdata {
    height: 100px;
    min-height: 100px;
  }

  table.fields th {
    font-size: 12px;
    padding: 3px 6px;
  }

  table.fields td {
    font-size: 10px;
    padding: 3px 6px;
  }
</style>
