<script>
  import Quill from "quill";
  import "quill/dist/quill.bubble.css";

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
  import {SaveExtractorType,TestGrok,AutoGrok,SaveFieldType} from '../../../wailsjs/go/main/App';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let extractorType;
  export let add = true;
  export let testLog = "";
  
  let data = [];
  let fields = [];
  let types = {};
  let columns = [];
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let quill;

  const back = () => {
    dispatch("done", {});
  };

  const save = async () => {
    const r = await SaveExtractorType(extractorType);
    errorMsg = r || "";
    if (r == "") {
      if (fields) {
        fields.forEach((key) => {
          if (!isFieldValid(key)) {
            SaveFieldType(
              {
                Key: key,
                Name: _getFieldName(key),
                Type: _getFieldType(key),
                Unit: "",
                CanEdit: true,
              }
            );
          }
        });
      }
      dispatch("done", { save: true });
    }
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const _getFieldName = (k) => {
    const n = getFieldName(k);
    return  n.includes("(unkn") ?  k : n;
  }

  const _getFieldType = (k) => {
    const n = getFieldName(k);
    if (n.includes("(unkn")) {
      return types[k] || "string";
    }
    return   getFieldType(k);
  }

  const test = async () => {
    if (extractorType.Grok == "") {
      errorMsg = $_('EditExtractorType.InputPatMsg');
      return;
    }
    if (testLog == "") {
      errorMsg = $_('EditExtractorType.InputTestDataMsg');
      return;
    }
    const r = await TestGrok(extractorType.Grok, testLog);
    if (r) {
      errorMsg = r.ErrorMsg;
      data = r.Data;
      columns = [];
      fields = r.Fields;
      types  = r.Types;
      fields.forEach((e) => {
        columns.push(_getFieldName(e));
      });
    }
  };

  const auto = async () => {
    if (testLog == "") {
      errorMsg = $_('EditExtractorType.InputTestDataMsg');
      return;
    }
    const r = await AutoGrok(testLog);
    if (r) {
      errorMsg = r.ErrorMsg;
      if (r.Grok) {
        extractorType.Grok = r.Grok;
        if (quill) {
          quill.setText(r.Grok,"user");
        }
      }
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
    if(!quill) {
      return;
    }
    const range = quill.getSelection();
    if (!range || range.length === 0) {
      return;
    }
    const text = quill.getText(range.index, range.length);
    if (text == "") {
      return;
    }
    const r = w ? getGrokPat(text) : '.+';
    extractorType.Grok = extractorType.Grok.substring(0,range.index) + r + extractorType.Grok.substring(range.index + range.length);
    quill.setText(extractorType.Grok,"user");
  };

  const getGrokColor = (c) => {
    if (!c.match) {
      return "";
    }
    if (c.text.includes("%{IP")) {
      return "#1a7f37";
    }
    if (c.text.includes("%{MAC")) {
      return "#8250df";
    }
    if (c.text.includes("%{INT") || c.text.includes("%{NUM") ) {
      return "#6fdd8b";
    }
    if (c.text.includes("%{URL")) {
      return "#bc4c00";
    }
    if (c.text.includes("%{")) {
      return "#bf8722";
    }
    if (c.text.includes("s")) {
      return "#0969da";
    }
    return "#cf222e";
  };

  let page = "";
  const addFT = true;
  let fieldType = {};
  const addFieldType = (key) => {
    fieldType = {
      Key: key,
      Name: _getFieldName(key),
      Type: _getFieldType(key),
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

  onMount(() => {
    loadFieldTypes();
    quill = new Quill('#quill', {
      theme: 'bubble',
      modules: {
       toolbar: ['underline','strike'],
        history: {
          delay: 2000,
          maxStack: 500,
          userOnly: true,
        },
      }
    });
    if (!quill) {
      return;
    }
    const toolbar = quill.getModule('toolbar');
    if(toolbar) {
      toolbar.addHandler('underline', (v) => { replaceGrok(true) });
      toolbar.addHandler('strike', (v) => { replaceGrok(false) });
    }
    quill.on('text-change', (delta, oldDelta, source) => {
      if(source == "api") {
        return;
      }
      extractorType.Grok = quill.getText().trim();
      if (extractorType.Grok.includes("\n")) {
        extractorType.Grok = extractorType.Grok.replaceAll("\n","");
        quill.setText(extractorType.Grok,"user");
        return;
      }
      const chunk =  highlightWords({
          text: extractorType.Grok,
          query: /(%\{.+?\}|\.\+|\\s\+)/,
      });
      let i = 0;
      quill.removeFormat(0,extractorType.Grok.length);
      for (const c of chunk) {
        const col = getGrokColor(c)
        if (col != "") {
          quill.formatText(i,c.text.length,"color",col)
          if (c.text.startsWith("%{")) {
            quill.formatText(i,c.text.length,"underline",true)
          } else {
            quill.formatText(i,c.text.length,"underline",false)
          }
        }
        i += c.text.length;
      }
    });
    quill.keyboard.addBinding({ key: 'S', metaKey: true }, (r,c) => {replaceGrok(false) });
    quill.keyboard.addBinding({ key: 'W', metaKey: true }, (r,c) => {replaceGrok(true) });
    quill.setText(extractorType.Grok,"user");
  });
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
          </h5>
        </div>
        <div id="quill" class="form-group-body mt-1">
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
                  >{_getFieldName(f)}</td
                >
                <td>{_getFieldType(f)}</td>
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

  #quill {
    font-size: 14px;
    font-family: monospace, serif;
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
