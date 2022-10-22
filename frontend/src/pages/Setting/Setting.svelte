<script>
  import {
    X16,
    Plus16,
    Check16,
    File16,
    Checklist16,
    Trash16,
    Search16,
    Download16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import LogSource from "./LogSource.svelte";
  import LogType from "./LogType.svelte";
  import { onMount } from "svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { loadFieldTypes } from "../../js/define";

  const dispatch = createEventDispatcher();
  const data = [];
  let config = {
    Filter: "",
    Extractor: "timeonly",
    Grok: "",
    TimeField: "",
    GeoIP: false,
    GeoIPDB: "",
    GeoFields: "",
    HostName: false,
    HostFields: "",
    VendorName: false,
    MACFields: "",
    Recursive: false,
    InMemory: false,
    SampleLog: "",
    ForceUTC: false,
  };
  let logSource = {
    No: 0,
    Type: "folder",
    Path: "",
    Pattern: "",
    InternalPattern: "",
    User: "",
    Password: "",
    Server: "",
    KeyPath: "",
    Start: "",
    End: "",
    Tag: "",
    Host: "",
    Channel: "",
    Auth: "",
  };
  let logSources = [];
  let errorMsg = "";
  let infoMsg = "";
  let page = "";
  let orgConfig;
  let hasIndex = false;

  const getConfig = () => {
    window.go.main.App.GetConfig().then((c) => {
      config = c;
      orgConfig = c;
    });
  };

  const getHasIndex = () => {
    window.go.main.App.HasIndex().then((r) => {
      hasIndex = r;
    });
  };

  const getLSPath = (e) => {
    switch (e.Type) {
      case "ssh":
      case "scp":
        return e.Server + ":" + e.Path;
      case "twsnmp":
        return (
          e.Server +
          "/?start=" +
          e.Start +
          "&end=" +
          e.End +
          "&host=" +
          e.Host +
          "&tag=" +
          e.Tag +
          "&message=" +
          e.Pattern
        );
      case "gravwell":
        return (
          e.Server +
          "/?start=" +
          e.Start +
          "&end=" +
          e.End +
          "&qeury=`" +
          e.Pattern +
          "`"
        );
      case "windows":
        return (
          e.Server +
          "/?start=" +
          e.Start +
          "&end=" +
          e.End +
          "&channel=`" +
          e.Channel +
          "`"
        );
    }
    return e.Path;
  };

  let pagination = false;
  const getLogSources = () => {
    window.go.main.App.GetLogSources().then((ds) => {
      data.length = 0; // 空にする
      if (ds) {
        logSources = ds;
        logSources.forEach((e) => {
          const path = getLSPath(e);
          data.push([e.No, e.Type, path, ""]);
        });
        if (ds.length > 5) {
          pagination = {
            limit: 5,
            enable: true,
          };
        } else {
          pagination = false;
        }
      } else {
        logSources = [];
      }
    });
  };

  let extractorTypes = {};
  let extractorTypeList = [];

  const getExtractorTypes = () => {
    window.go.main.App.GetExtractorTypes().then((r) => {
      if (r) {
        extractorTypes = r;
        extractorTypeList = [];
        for (let k in extractorTypes) {
          extractorTypeList.push(extractorTypes[k]);
        }
        extractorTypeList.sort((a, b) => a.Name > b.Name);
      }
    });
  };

  onMount(() => {
    loadFieldTypes();
    getConfig();
    getLogSources();
    getExtractorTypes();
    getHasIndex();
  });

  const editLogSource = (sno) => {
    const no = sno * 1;
    if (sno == "" || no < 0 || no > logSources.length) {
      // 新規
      logSource = {
        No: 0,
        Type: "folder",
        Path: "",
        Pattern: "",
        Server: "",
        User: "",
        Password: "",
        KeyPath: "",
        Start: "",
        End: "",
        Host: "",
        Tag: "",
        Channel: "",
        Auth: "",
      };
    } else {
      logSource = logSources[no - 1];
    }
    page = "logSource";
  };

  const deleteLogSource = (sno) => {
    const no = sno * 1;
    window.go.main.App.DeleteLogSource(no).then((e) => {
      errorMsg = e;
    });
  };

  const formatLogSourceType = (t) => {
    switch (t) {
      case "folder":
        return "フォルダー";
      case "file":
        return "単一ファイル";
      case "scp":
        return "SCP転送";
      case "windows":
        return "Windows";
      case "cmd":
        return "コマンド";
      case "ssh":
        return "SSHコマンド";
      case "twsnmp":
        return "TWSNMP連携";
      case "gravwell":
        return "Gravwell連携";
    }
    return "";
  };

  const editLogSourceButtons = (_, row) => {
    const no = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => editLogSource(no),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>`
      )
    );
  };

  const deleteLogSourceButtons = (_, row) => {
    const no = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm btn-danger",
        onClick: () => deleteLogSource(no),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.576l.66-6.6a.75.75 0 00-1.492-.149l-.66 6.6a.25.25 0 01-.249.225h-5.19a.25.25 0 01-.249-.225l-.66-6.6z"></path></svg>`
      )
    );
  };

  const columns = [
    {
      name: "No",
      sort: true,
      width: "10%",
    },
    {
      name: "タイプ",
      sort: true,
      width: "20%",
      formatter: formatLogSourceType,
    },
    {
      name: "パス",
      sort: true,
      width: "50%",
    },
    {
      name: "編集",
      sort: false,
      width: "10%",
      formatter: editLogSourceButtons,
    },
    {
      name: "削除",
      sort: false,
      width: "10%",
      formatter: deleteLogSourceButtons,
    },
  ];
  let busy = false;
  const start = () => {
    // Index作成を開始
    busy = true;
    window.go.main.App.Start(config, false).then((e) => {
      busy = false;
      if (e && e != "") {
        errorMsg = e;
      } else {
        dispatch("done", { page: "processing" });
      }
    });
  };

  const clear = () => {
    // Indexをクリア
    busy = true;
    window.go.main.App.ClearIndex().then((e) => {
      busy = false;
      if (e && e != "") {
        errorMsg = e;
      } else {
        hasIndex = false;
      }
    });
  };

  const search = () => {
    // 既存のIndexで検索開始
    busy = true;
    window.go.main.App.Start(config, true).then((e) => {
      busy = false;
      if (e && e != "") {
        errorMsg = e;
      } else {
        dispatch("done", { page: "logview" });
      }
    });
  };

  const selectGeoIPDB = () => {
    window.go.main.App.SelectFile("geoip").then((f) => {
      config.GeoIPDB = f;
    });
  };

  const cancel = () => {
    window.go.main.App.CloseWorkDir();
    dispatch("done", { page: "wellcome" });
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg = "";
  };

  const changeExtractor = () => {
    const e = extractorTypes[config.Extractor];
    if (e) {
      config.Grok = e.Grok;
      config.TimeField = e.TimeField;
      config.GeoFields = e.IPFields;
      config.HostFields = e.IPFields;
      config.MACFields = e.MACFields;
    } else if (orgConfig) {
      config.Grok = orgConfig.Grok;
      config.TimeField = orgConfig.TimeField;
      config.GeoFields = orgConfig.IPFields;
      config.HostFields = orgConfig.IPFields;
      config.MACFields = orgConfig.MACFields;
    }
  };

  const handleDone = (e) => {
    if (e && e.detail && e.detail.update) {
      getLogSources();
    }
    page = "";
  };

  const showLogTypePage = () => {
    page = "logType";
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  {#if busy}
    <div class="Box-header">
      <h3 class="Box-title">ログ分析起動</h3>
    </div>
    <div class="flash mt-2">
      ログの読み込みを準備しています。お待ち下さい<span
        class="AnimatedEllipsis"
      />
    </div>
  {:else if page == "logSource"}
    <LogSource {logSource} on:done={handleDone} />
  {:else if page == "logType"}
    <LogType on:done={handleDone} />
  {:else}
    <div class="Box-header">
      <h3 class="Box-title">ログ分析の設定</h3>
    </div>
    {#if errorMsg != ""}
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
    {#if infoMsg != ""}
      <div class="flash">
        {infoMsg}
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
          <h5 class="pb-1">
            ログを読み込む場所
            <button
              class="btn btn-sm float-right"
              type="button"
              on:click={() => editLogSource("")}
            >
              <Plus16 />
            </button>
          </h5>
        </div>
        <div class="form-group-body markdown-body mt-3">
          <Grid {data} {pagination} {columns} language={jaJP} />
          <label class="p-1">
            <input type="checkbox" bind:checked={config.Recursive} />
            tar.gzの再帰読み込み
          </label>
          <label class="p-1">
            <input type="checkbox" bind:checked={config.ForceUTC} />
            タイムゾーン不明はUTC
          </label>
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">フィルター</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            style="width: 100%;"
            placeholder="フィルター"
            aria-label="フィルター"
            bind:value={config.Filter}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">ログの種類</h5>
        </div>
        <div class="form-group-body">
          <!-- svelte-ignore a11y-no-onchange -->
          <select
            class="form-select"
            aria-label="抽出パターン"
            bind:value={config.Extractor}
            on:change={changeExtractor}
          >
            <option value="timeonly">タイムスタンプのみ</option>
            {#each extractorTypeList as { Key, Name }}
              <option value={Key}>{Name}</option>
            {/each}
            <option value="custom">カスタム</option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">アドレス情報</h5>
        </div>
        <div class="form-group-body">
          <label class="p-1">
            <input type="checkbox" bind:checked={config.HostName} />
            ホスト名を調べる
          </label>
          <label class="p-1">
            <input class="ml-2" type="checkbox" bind:checked={config.GeoIP} />
            位置情報を調べる
          </label>
          <label class="p-1">
            <input
              class="ml-2"
              type="checkbox"
              bind:checked={config.VendorName}
            />
            ベンダー名を調べる
          </label>
        </div>
      </div>
      {#if config.Extractor == "custom" || config.Extractor.startsWith("EXT")}
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">抽出パターン</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="GROKパターン"
              aria-label="GROKパターン"
              style="width: 100%;"
              bind:value={config.Grok}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">取得情報</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control"
              type="text"
              style="width: 15%;"
              placeholder="タイムスタンプ項目"
              bind:value={config.TimeField}
            />
            {#if config.HostName}
              <input
                class="form-control"
                type="text"
                style="width: 25%;"
                placeholder="ホスト名解決項目"
                bind:value={config.HostFields}
              />
            {/if}
            {#if config.GeoIP}
              <input
                class="form-control"
                type="text"
                placeholder="IP位置情報項目"
                style="width: 25%;"
                bind:value={config.GeoFields}
              />
            {/if}
            {#if config.VendorName}
              <input
                class="form-control"
                type="text"
                placeholder="MACアドレス項目"
                style="width: 20%;"
                bind:value={config.MACFields}
              />
            {/if}
          </div>
        </div>
      {/if}
      <div class="form-group">
        <div class="form-group-header">
          <h5 class="pb-1">インデクサー設定</h5>
        </div>
        <div class="form-group-body">
          <div class="form-checkbox">
            <label>
              <input
                type="checkbox"
                disabled={hasIndex}
                bind:checked={config.InMemory}
                aria-describedby="help-text-for-inmemory"
              />
              インデックスをメモリ上に作成
            </label>
            <p class="note" id="help-text-for-inmemory">
              メモリに余裕があればオンにすると多少高速化できます。
            </p>
          </div>
        </div>
      </div>
      {#if config.GeoIP}
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">IP位置情報データベース</h5>
          </div>
          <div class="form-group-body">
            <div class="input-group">
              <input
                class="form-control"
                type="text"
                placeholder="Geo IPデータベース"
                aria-label="Geo IPデータベース"
                bind:value={config.GeoIPDB}
              />
              <span class="input-group-button">
                <button class="btn" type="button" on:click={selectGeoIPDB}>
                  <File16 />
                </button>
              </span>
            </div>
          </div>
        </div>
      {/if}
    </div>
    <div class="Box-footer text-right">
      <button
        class="btn btn-outline mr-1"
        type="button"
        on:click={showLogTypePage}
      >
        <Checklist16 />
        ログ定義
      </button>
      <button class="btn btn-secondary mr-1" type="button" on:click={cancel}>
        <X16 />
        キャンセル
      </button>
      {#if hasIndex}
        <button class="btn btn-danger mr-1" type="button" on:click={clear}>
          <Trash16 />
          インデックス削除
        </button>
        <button class="btn btn-danger mr-1" type="button" on:click={start}>
          <Check16 />
          追加読み込み
        </button>
        <button class="btn btn-primary mr-1" type="button" on:click={search}>
          <Search16 />
          検索画面へ
        </button>
      {:else}
        <button class="btn btn-primary" type="button" on:click={start}>
          <Check16 />
          インデクス作成を開始
        </button>
      {/if}
    </div>
  {/if}
</div>
