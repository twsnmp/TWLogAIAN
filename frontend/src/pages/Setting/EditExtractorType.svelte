<script>
  import highlightWords from "highlight-words";
  import { X16, Check16, StarFill16, Reply16, Plus16 } from "svelte-octicons";
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

  export let extractorType;
  export let add = true;
  export let testLog = "";
  
  let data = [];
  let fields = [];
  let columns = [];
  const dispatch = createEventDispatcher();
  let errorMsg = "";

  onMount(() => {
    loadFieldTypes();
  });

  const back = () => {
    dispatch("done", {});
  };

  const save = () => {
    window.go.main.App.SaveExtractorType(extractorType).then((r) => {
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
      errorMsg = "パターンを指定してください";
      return;
    }
    if (testLog == "") {
      errorMsg = "テストデータを指定してください";
      return;
    }
    window.go.main.App.TestGrok(extractorType.Grok, testLog).then((r) => {
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
      errorMsg = "テストデータを指定してください";
      return;
    }
    window.go.main.App.AutoGrok(testLog).then((r) => {
      errorMsg = r.ErrorMsg;
      if (r.Grok) {
        oldGrok.push(extractorType.Grok);
        extractorType.Grok = r.Grok;
      }
    });
  };

  const oldGrok = [];
  const resetGrok = () => {
    if (oldGrok.length > 0) {
      extractorType.Grok = oldGrok.pop();
    } else {
      extractorType.Grok = "";
    }
  };

  const hasOldGrok = () => {
    return  oldGrok.length > 0;
  }

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

  const replaceGrok = (e) => {
    if (e.key != "Tab" && e.key != "Escape") {
      return;
    }
    if (!e.target || !e.target.selectionStart) {
      return;
    }
    e.preventDefault();
    const { selectionStart, selectionEnd, value } = e.target;
    const sel = value.slice(selectionStart, selectionEnd);
    if (sel == "") {
      return;
    }
    const newPat = e.key == "Tab" ? getGrokPat(sel) : ".+";
    oldGrok.push(extractorType.Grok);
    extractorType.Grok =
      value.slice(0, selectionStart) + newPat + value.slice(selectionEnd);
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
    }
    page = "";
  };
</script>

{#if page == "fieldType"}
  <EditFieldType {fieldType} add={addFT} on:done={handleEditFieldTypeDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title">抽出(Grok)パターン編集</h3>
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
          <h5>キー</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            disabled={!add}
            placeholder="抽出パターンのキー"
            bind:value={extractorType.Key}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>名前</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder="抽出パターンの名前"
            bind:value={extractorType.Name}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>
            抽出パターン
            {#if hasOldGrok()}
              <button type="button" class="btn btn-sm ml-2" on:click={resetGrok}
                ><Reply16 /></button
              >
            {/if}
          </h5>
        </div>
        <div class="form-group-body">
          <p class="f6 color-fg-accent">
            選択後にTabキーで変数に変換、ESCキーで無視する部分に変換
          </p>
          <input
            class="form-control grok"
            type="text"
            placeholder="抽出パターン"
            bind:value={extractorType.Grok}
            on:keydown={replaceGrok}
          />
        </div>
      </div>
      <div class="mt-1">
        {#each grokChunks as chunk}
          <span class={getGrokClass(chunk)}>{chunk.text}</span>
        {/each}
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>テストデータ</h5>
        </div>
        <div class="form-group-body">
          <textarea class="form-control testdata" bind:value={testLog} />
        </div>
      </div>
    </div>
    {#if fields.length > 0}
      <div class="Box-row markdown-body log">
        <h5>抽出した項目</h5>
        <table class="fields" width="100%">
          <thead>
            <tr>
              <th width="20%">変数名</th>
              <th width="40%">名前</th>
              <th width="20%">種別</th>
              <th width="10%">単位</th>
              <th width="10%">追加</th>
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
        <h5>抽出したデータ</h5>
        <Grid {data} {columns} language={jaJP} />
      </div>
    {/if}
    <div class="Box-footer text-right">
      <button class="btn btn-danger mr-1" type="button" on:click={auto}>
        <StarFill16 />
        自動抽出パターン生成
      </button>
      <button class="btn btn-primary" type="button" on:click={test}>
        <Check16 />
        テスト
      </button>
      <button class="btn btn-secondary mr-1" type="button" on:click={back}>
        <X16 />
        キャンセル
      </button>
      <button class="btn btn-primary mr-1" type="button" on:click={save}>
        <X16 />
        保存
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
