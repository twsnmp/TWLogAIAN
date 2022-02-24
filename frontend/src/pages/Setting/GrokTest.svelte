<script>
  import { X16, Check16,StarFill16 } from "svelte-octicons";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { createEventDispatcher } from "svelte";
  import {
    getFieldName,
    getFieldType,
    getFieldUnit,
    isFieldValid,
  } from "../../js/define";
  export let grok;
  export let extractorTypes;
  let selected = "";
  let testData = "";
  let data = [];
  let fields = [];
  let columns = [];
  const dispatch = createEventDispatcher();
  let errorMsg = "";

  const back = () => {
    dispatch("done", { update: false });
  };

  const save = () => {
    dispatch("done", { update: false,grok: grok });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const test = () => {
    if (grok == "") {
      errorMsg = "パターンを指定してください"
      return;
    }
    if (testData == "") {
      errorMsg = "テストデータを指定してください"
      return;
    }
    window.go.main.App.TestGrok(grok, testData).then((r) => {
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
    if (testData == "") {
      errorMsg = "テストデータを指定してください"
      return;
    }
    window.go.main.App.AutoGrok(testData).then((r) => {
      errorMsg = r.ErrorMsg;
      if (r.Grok) {
        grok = r.Grok;
      } 
    });
  };

  const setGrok = () => {
    if (selected == "") {
      return;
    }
    grok = selected;
  };
</script>

<div class="Box-header">
  <h3 class="Box-title">抽出(Grok)パターンテスト</h3>
</div>
{#if errorMsg }
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
  <form>
    <div class="form-group">
      <div class="form-group-header">
        <h5>テンプレート</h5>
      </div>
      <div class="form-group-body">
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          class="form-select"
          aria-label="テンプレート"
          bind:value={selected}
          on:change={setGrok}
        >
          <option value="">テンプレート選択</option>
          {#each extractorTypes as { Grok, Name }}
            <option value={Grok}>{Name}</option>
          {/each}
        </select>
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>抽出パターン</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control grok"
          type="text"
          placeholder="抽出パターン"
          aria-label="抽出パターン"
          bind:value={grok}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>テストデータ</h5>
      </div>
      <div class="form-group-body">
        <textarea class="form-control testdata" bind:value={testData} />
      </div>
    </div>
  </form>
</div>
{#if fields.length > 0}
  <div class="Box-row markdown-body log">
    <h5>抽出した項目</h5>
    <table class="fields">
      <thead>
        <tr>
          <th>変数名</th>
          <th>名前</th>
          <th>種別</th>
          <th>単位</th>
        </tr>
      </thead>
      <tbody>
        {#each fields as f}
          <tr>
            <td>{f}</td>
            <td class:color-fg-danger={!isFieldValid(f)}>{getFieldName(f)}</td>
            <td>{getFieldType(f)}</td>
            <td>{getFieldUnit(f)}</td>
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
  <button class="btn btn-outline mr-2" type="button" on:click={auto}>
    <StarFill16 />
    自動抽出パターン生成
  </button>
  <button class="btn btn-secondary mr-2" type="button" on:click={back}>
    <X16 />
    終了
  </button>
  {#if grok}
    <button class="btn btn-primary mr-2" type="button" on:click={save}>
      <X16 />
      適用
    </button>
    <button class="btn btn-primary" type="button" on:click={test}>
      <Check16 />
      テスト
    </button>
  {/if}
</div>

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
