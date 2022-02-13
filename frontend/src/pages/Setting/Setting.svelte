<script>
  import { X16, Plus16, Check16, File16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import LogSource from "./LogSource.svelte";
  import { onMount } from "svelte";
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
    InMemory: false,
    SampleLog: "",
  };
  let logSource = {
    No: 0,
    Type: "folder",
    Path: "",
    Pattern: "",
    User: "",
    Password: "",
    Server: "",
    KeyPath: "",
  };
  let logSources = [];
  let errorMsg = "";
  let infoMsg = "";
  let edit = false;

  const getConfig = () => {
    window.go.main.App.GetConfig().then((c) => {
      config = c;
    });
  };
  let pagination = false;
  const getLogSources = () => {
    window.go.main.App.GetLogSources().then((ds) => {
      data.length = 0; // 空にする
      if (ds) {
        logSources = ds;
        logSources.forEach((e) => {
          const path = e.Type == "scp" ? e.Server + ":" +e.Path : e.Path;
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
  let extractorTypes = [];
  const hasIPMap = {};
  const getExtractorTypes = () => {
    window.go.main.App.GetExtractorTypes().then((r) => {
      extractorTypes = r;
      extractorTypes.forEach((e) => {
        hasIPMap[e.Key] = e.IP;
      });
      hasIPMap["timeonly"] = false;
      hasIPMap["custom"] = true;
    });
  };
  onMount(() => {
    getConfig();
    getLogSources();
    getExtractorTypes();
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
      };
    } else {
      logSource = logSources[no - 1];
    }
    edit = true;
  };

  const formatLogSourceType = (t) => {
    switch (t) {
      case "folder":
        return "フォルダー";
      case "file":
        return "単一ファイル";
      case "scp":
        return "SCP";
    }
    return "";
  }

  const actionButtons = (_, row) => {
    const no = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => editLogSource(no),
      },
      html(
        `<svg xmlns="http://www.w3.org/2000/svg" class="octicon" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>`
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
      width: "60%",
    },
    {
      name: "編集",
      sort: false,
      width: "10%",
      formatter: actionButtons,
    },
  ];

  const start = () => {
    // Index作成を開始
    window.go.main.App.Start(config).then((e) => {
      if (e && e != "") {
        errorMsg = e;
      } else {
        dispatch("done", { page: "processing" });
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

  const testSampleLog = () => {
    clearMsg();
    if (config.SampleLog == "") {
      errorMsg = "サンプルログが空欄では私にはログの種類を判断できません";
      return;
    }
    window.go.main.App.TestSampleLog(config).then((et) => {
      if (!et) {
        errorMsg = "私にはログの種類を判断できませんでした";
      } else {
        infoMsg = "ログの種類を" + et.Name + "に設定しました。";
        config.Extractor = et.Key;
      }
    });
  };

  const handleDone = (e) => {
    if (e && e.detail && e.detail.update) {
      getLogSources();
    }
    edit = false;
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 800px;">
  {#if edit}
    <LogSource {logSource} on:done={handleDone} />
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
      <form>
        <div class="form-group">
          <div class="form-group-header">
            <h5>
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
            <Grid {data} {pagination} {columns} />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>フィルター</h5>
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
            <h5>ログの種類</h5>
          </div>
          <div class="form-group-body">
            <select
              class="form-select"
              aria-label="抽出パターン"
              bind:value={config.Extractor}
            >
              <option value="timeonly">タイムスタンプのみ</option>
              {#each extractorTypes as { Key, Name }}
                <option value={Key}>{Name}</option>
              {/each}
              <option value="custom">カスタム</option>
            </select>
            <div class="input-group mt-2">
              <input
                class="form-control"
                type="text"
                placeholder="ログのサンプル"
                aria-label="ログのサンプル"
                bind:value={config.SampleLog}
              />
              <span class="input-group-button">
                <button class="btn" type="button" on:click={testSampleLog}>
                  <Check16 />
                </button>
              </span>
            </div>
          </div>
        </div>
        {#if hasIPMap[config.Extractor]}
          <div class="form-group">
            <div class="form-group-header">
              <h5>IPアドレス情報</h5>
            </div>
            <div class="form-group-body">
              <label>
                <input
                  type="checkbox"
                  bind:checked={config.HostName}
                />
                ホスト名を調べる
              </label>
              <label>
                <input
                  class="ml-2"
                  type="checkbox"
                  bind:checked={config.GeoIP}
                />
                位置情報を検索
              </label>
            </div>
          </div>
        {/if}
        {#if config.Extractor == "custom"}
          <div class="form-group">
            <div class="form-group-header">
              <h5>カスタム抽出パターン</h5>
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
              <h5>タイムスタンプ項目</h5>
            </div>
            <div class="form-group-body">
              <input
                class="form-control mt-2"
                type="text"
                placeholder="項目"
                aria-label="タイムスタンプ項目"
                bind:value={config.TimeField}
              />
            </div>
          </div>
          {#if config.HostName}
            <div class="form-group">
              <div class="form-group-header">
                <h5>ホスト名解決項目</h5>
              </div>
              <div class="form-group-body">
                <input
                  class="form-control"
                  type="text"
                  placeholder="ホスト名解決項目"
                  aria-label="ホスト名解決項目"
                  bind:value={config.HostFields}
                />
              </div>
            </div>
          {/if}
          {#if config.GeoIP}
            <div class="form-group">
              <div class="form-group-header">
                <h5>IP位置情報項目</h5>
              </div>
              <div class="form-group-body">
                <input
                  class="form-control"
                  type="text"
                  placeholder="IP位置情報項目"
                  aria-label="IP位置情報項目"
                  bind:value={config.GeoFields}
                />
              </div>
            </div>
          {/if}
        {/if}
        <div class="form-group">
          <div class="form-group-header">
            <h5>インデクサー設定</h5>
          </div>
          <div class="form-group-body">
            <div class="form-checkbox">
              <label>
                <input
                  type="checkbox"
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
              <h5>IP位置情報データベース</h5>
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
      </form>
    </div>
    <div class="Box-footer text-right">
      <button class="btn  btn-secondary" type="button" on:click={cancel}>
        <X16 />
        キャンセル
      </button>
      <button class="btn btn-primary ml-2" type="button" on:click={start}>
        <Check16 />
        開始
      </button>
    </div>
  {/if}
</div>
