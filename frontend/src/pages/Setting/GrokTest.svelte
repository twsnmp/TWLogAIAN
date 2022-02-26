<script>
  import { X16, Check16, StarFill16, Reply16 } from "svelte-octicons";
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
    dispatch("done", { update: false, grok: grok });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const test = () => {
    if (grok == "") {
      errorMsg = "パターンを指定してください";
      return;
    }
    if (testData == "") {
      errorMsg = "テストデータを指定してください";
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
      errorMsg = "テストデータを指定してください";
      return;
    }
    window.go.main.App.AutoGrok(testData).then((r) => {
      errorMsg = r.ErrorMsg;
      if (r.Grok) {
        oldGrok = grok;
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

  let oldGrok = "";
  const resetGrok = () => {
    grok = oldGrok;
    oldGrok = "";
  };
  let selectedGrok = "";
  const replaceGrok = (e) => {
    if ( e.key != 'Tab' || selectedGrok == "") {
      return;
    }
    if (!e.target || !e.target.selectionStart){
      return;
    }
    e.preventDefault();
    const { selectionStart, selectionEnd, value } = e.target;
    const sel = value.slice(selectionStart, selectionEnd);
    if(sel == "") {
      return;
    }
    oldGrok = grok;
    grok = (
      value.slice(0, selectionStart) +
      selectedGrok +
      value.slice(selectionEnd)
    );
  };
</script>

<div class="Box-header">
  <h3 class="Box-title">抽出(Grok)パターンテスト</h3>
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
        <h5>抽出パターン
          <select class="form-select select-sm" bind:value={selectedGrok}>
            <option value="">TABキーで変換する項目を選択</option>
            <option value="{`\s+`}">空白部分</option>
            <option value="{`.*`}">無視する部分</option>
            <option value="{`%{NUMBER:number}`}">数値</option>
            <option value="{`%{INT:int}`}">整数</option>
            <option value="{`%{IP:ip}`}">IPアドレス</option>
            <option value="{`%{IPV4:ipv4}`}">IPv4アドレス</option>
            <option value="{`%{IPV6:ipv6}`}">IPv6アドレス</option>
            <option value="{`%{MAC:mac}`}">MACアドレス</option>
            <option value="{`%{URI:uri}`}">URI</option>
            <option value="{`%{LOGLEVEL:loglevel}`}">ログレベル</option>
            <option value="{`%{EMAILADDRESS:email}`}">メールアドレス</option>
            <option value="{`%{USER:user}`}">ユーザーID</option>
            <option value="{`%{GREEDYDATA:data}`}">何か情報</option>
            <option value="{`%{PATH:path}`}">パス</option>
            <option value="{`%{HOSTNAME:host}`}">ホスト名</option>
            <option value="{`%{IPHOST:iphost}`}">ホスト名またはIPアドレス</option>
            <option value="{`\s+`}">ホスト名:ポート番号</option>
            <option value="{`%{UUID:uuid}`}">UUID</option>
          </select>
        {#if oldGrok}
          <button class="btn btn-sm ml-1" on:click="{resetGrok}"><Reply16/></button>
        {/if}
        </h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control grok"
          type="text"
          placeholder="抽出パターン"
          aria-label="抽出パターン"
          bind:value={grok}
          on:keydown={replaceGrok}
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
  <button class="btn btn-danger mr-2" type="button" on:click={auto}>
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
