<script>
  import { getFieldName, getFieldType } from "../../js/define";
  import { Plus16, Trash16, File16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { _, getLocale } from "../../i18n/i18n";
  import { LoadKeyword, GenerateRegex, CancelAIReport } from "../../../wailsjs/go/main/App";
  import AutoComplete from "simple-svelte-autocomplete";

  export let conf;
  export let fields = [];
  export let extractorTypeList = [];
  export let storageEngine = "bluge";
  export let hasLLM = false;

  const geoFields = [];

  const setSelectedExtractor = () => {
    for (const e of extractorTypeList) {
      if (e.Key == conf.extractor) {
        return e;
      }
    }
    return extractorTypeList[0];
  };

  let hasStringField = false;
  let hasNumberField = false;
  let selectedExtractor = setSelectedExtractor();

  let history = "";
  const dispatch = createEventDispatcher();

  // Mode validation based on available features
  $: if (storageEngine !== "bluge" && conf.mode === "full") {
    conf.mode = "simple";
  }
  $: if (!hasLLM && conf.mode === "ai") {
    conf.mode = "simple";
  }

  // AI regex generation states
  let aiPrompt = "";
  let aiBusy = false;
  let aiRegex = "";
  let aiExplain = "";
  let aiError = "";
  let isComposing = false;

  const generateRegex = async () => {
    if (!aiPrompt.trim()) return;
    aiBusy = true;
    aiError = "";
    aiRegex = "";
    aiExplain = "";
    try {
      const loc = getLocale() === "ja" ? "ja" : "en";
      const res = await GenerateRegex(aiPrompt, loc);
      if (res && res.Regex) {
        aiRegex = res.Regex;
        aiExplain = res.Explain || "";
      } else if (res && res.ErrorMsg) {
        aiError = res.ErrorMsg;
      } else {
        aiError = "Failed to generate regex pattern";
      }
    } catch (err) {
      aiError = err.toString();
    } finally {
      aiBusy = false;
    }
  };

  const cancelAIRegex = () => {
    CancelAIReport();
    aiBusy = false;
    aiError = $_("AI.Canceled") || "処理を中止しました";
  };

  const applyAIRegex = () => {
    if (!aiRegex) return;
    conf.query = aiRegex;
    conf.mode = "regexp";
    dispatch("update", { query: aiRegex, add: false });
  };

  const setHistory = () => {
    if (history) {
      dispatch("update", { query: history, add: false });
    }
  };

  const addKeyword = () => {
    if (!conf.keyword.key) {
      return;
    }
    const q =
      conf.keyword.field == ""
        ? " " + conf.keyword.mode + conf.keyword.key
        : " " + conf.keyword.mode + conf.keyword.field + ":" + conf.keyword.key;
    dispatch("update", { query: q, add: true });
  };

  const addNumber = () => {
    if (!conf.number.field || !conf.number.oper || !conf.number.value) {
      return;
    }
    const oper = conf.number.oper == "=" ? ":" : ":" + conf.number.oper;
    const q = " " + conf.number.field + oper + conf.number.value;
    dispatch("update", { query: q, add: true });
  };

  fields.forEach((f) => {
    if (f.includes("_geo_latlong")) {
      geoFields.push(f);
      return;
    }
    if (f.startsWith("_")) {
      return;
    }
    if (getFieldType(f) == "string") {
      hasStringField = true;
    }
    if (getFieldType(f) == "number") {
      hasNumberField = true;
      if (conf.number.field == "") {
        conf.number.field = f;
      }
    }
  });

  const clear = () => {
    conf.history.length = 0;
  };

  const loadKeyword = async () => {
    const r = await LoadKeyword();
    if (r) {
      r.forEach((k) => {
        const q =
          conf.keyword.field == ""
            ? " " + conf.keyword.mode + k
            : " " + conf.keyword.mode + conf.keyword.field + ":" + k;
        dispatch("update", { query: q, add: true });
      });
    }
  };
</script>

<div class="search-conf-container">
  <!-- 検索履歴行 (ある場合のみ上部にコンパクトに表示) -->
  {#if conf.history.length > 0}
    <div class="conf-history-row">
      <span class="conf-label">{$_("SearchConf.History")}</span>
      <div class="d-flex flex-items-center flex-1 gap-2">
        <select
          class="form-select select-sm flex-1"
          bind:value={history}
          on:change={setHistory}
        >
          <option value="">{$_("SearchConf.SelectHistoryMsg")}</option>
          {#each conf.history as h}
            <option value={h}>{h}</option>
          {/each}
        </select>
        <button class="btn btn-danger btn-sm" type="button" on:click={clear}>
          <Trash16 />
        </button>
      </div>
    </div>
  {/if}

  <!-- メイン設定: 左右2列グリッド -->
  <div class="conf-main-grid">
    <!-- 左列: 基本検索条件 -->
    <div class="conf-col">
      <!-- 検索文のモード -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.QueryMode")}</span>
        <select class="form-select select-sm conf-control" bind:value={conf.mode}>
          <option value="simple">{$_("SearchConf.Simple")}</option>
          <option value="regexp">{$_("SearchConf.Regexp")}</option>
          {#if storageEngine === "bluge"}
            <option value="full">{$_("SearchConf.FullTextSearch")}</option>
          {/if}
          {#if hasLLM}
            <option value="ai">{$_("SearchConf.AI")}</option>
          {/if}
        </select>
      </div>

      <!-- 最大件数 -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.Limit")}</span>
        <select class="form-select select-sm conf-control" bind:value={conf.limit}>
          <option value="1000">1000</option>
          <option value="2000">2000</option>
          <option value="5000">5000</option>
          <option value="10000">10000</option>
          <option value="20000">20000</option>
          <option value="50000">50000</option>
          <option value="100000">100000</option>
          <option value="200000">200000</option>
        </select>
      </div>

      <!-- 検索期間モード -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.TimeRange")}</span>
        <select class="form-select select-sm conf-control" bind:value={conf.range.mode}>
          <option value="">{$_("SearchConf.NoSelect")}</option>
          <option value="target">{$_("SearchConf.TagetTime")}</option>
          <option value="range">{$_("SearchConf.StartEndTime")}</option>
        </select>
      </div>
    </div>

    <!-- 右列: 分析・表示オプション -->
    <div class="conf-col">
      <!-- 異常ログ検知方法 (+ 特徴量計算方法) -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.AnomaryMode")}</span>
        <div class="d-flex flex-items-center flex-1 gap-1">
          <select class="form-select select-sm conf-control" bind:value={conf.anomaly}>
            <option value="">{$_("SearchConf.NotDetect")}</option>
            <option value="iforest">Isolation Forest</option>
            <option value="zscore">{$_("SearchConf.ZScore")}</option>
            {#if conf.limit <= 10000}
              <option value="lof">Local Outlier Factor</option>
              <option value="knn">{$_("SearchConf.KNN")}</option>
              <option value="mahalanobis">{$_("SearchConf.Mahalanobis")}</option>
              <option value="autoencoder">Auto Encoder</option>
              <option value="lstm">{$_("SearchConf.LSTM")}</option>
            {/if}
            <option value="sum">Sum</option>
          </select>
          {#if conf.anomaly}
            <select
              class="form-select select-sm conf-control"
              bind:value={conf.vector}
              title={$_("SearchConf.CalcVectorMode")}
            >
              <option value="">{$_("SearchConf.NumData")}</option>
              <option value="time">{$_("SearchConf.NumDataWeekDayH")}</option>
              <option value="all">{$_("SearchConf.StringNumData")}</option>
              <option value="alltime">{$_("SearchConf.StringNumWeekDayH")}</option>
              <option value="sql">{$_("SearchConf.SQLInjection")}</option>
              <option value="oscmd">{$_("SearchConf.OSCmdInjection")}</option>
              <option value="dirt">{$_("SearchConf.DirTra")}</option>
              <option value="walu">{$_("SearchConf.AccessLogWalu")}</option>
              <option value="tfidf">TF-IDF</option>
            </select>
          {/if}
        </div>
      </div>

      <!-- 検索時データ抽出 -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.ExtractOnSeach")}</span>
        <div class="conf-control autocomplete-container">
          <AutoComplete
            items="{extractorTypeList}"
            bind:value="{conf.extractor}"
            labelFieldName="Name"
            valueFieldName="Key"
            inputClassName="form-control input-sm width-full"
            bind:selectedItem="{selectedExtractor}"
          />
        </div>
      </div>

      <!-- ハイライト表示 -->
      <div class="conf-item">
        <span class="conf-label">{$_("SearchConf.HighlighMode")}</span>
        <select class="form-select select-sm conf-control" bind:value={conf.highlightMode}>
          <option value="">{$_("SearchConf.None")}</option>
          <option value="keyword">{$_("SearchConf.HighlighKeyword")}</option>
          <option value="code">{$_("SearchConf.HighlighCode")}</option>
        </select>
      </div>
    </div>
  </div>

  <!-- 下部展開エリア 1: 検索期間の詳細指定 -->
  {#if conf.range.mode == "target"}
    <div class="conf-sub-row">
      <span class="conf-label">{$_("SearchConf.TargetRange")}</span>
      <div class="d-flex flex-items-center flex-1 gap-2">
        <input
          class="form-control input-sm flex-1"
          type="text"
          placeholder={$_("SearchConf.TargetDateTime")}
          bind:value={conf.range.target}
        />
        <input
          class="form-control input-sm"
          style="width: 85px;"
          type="number"
          bind:value={conf.range.range}
        />
        <span class="text-small">{$_("SearchConf.Sec")}</span>
      </div>
    </div>
  {/if}
  {#if conf.range.mode == "range"}
    <div class="conf-sub-row">
      <span class="conf-label">{$_("SearchConf.TimeRangeTitle")}</span>
      <div class="d-flex flex-items-center flex-1 gap-2">
        <input
          class="form-control input-sm flex-1"
          type="datetime-local"
          placeholder={$_("SearchConf.Start")}
          bind:value={conf.range.start}
        />
        <span>-</span>
        <input
          class="form-control input-sm flex-1"
          type="datetime-local"
          placeholder={$_("SearchConf.End")}
          bind:value={conf.range.end}
        />
      </div>
    </div>
  {/if}

  <!-- 下部展開エリア 2: IP位置情報検索（geoFieldsがある場合のみ） -->
  {#if geoFields.length > 0}
    <div class="conf-sub-row">
      <span class="conf-label">{$_("SearchConf.IPLocSearch")}</span>
      <div class="d-flex flex-items-center flex-wrap flex-1 gap-2">
        <select class="form-select select-sm" bind:value={conf.geo.mode}>
          <option value="">{$_("SearchConf.NoSearch")}</option>
          <option value="centor">{$_("SearchConf.DistFromCentor")}</option>
        </select>
        {#if conf.geo.mode != ""}
          <select class="form-select select-sm" bind:value={conf.geo.field}>
            {#each geoFields as f}
              <option value={f}>{getFieldName(f)}</option>
            {/each}
          </select>
          <span class="text-small">{$_("SearchConf.Is")}</span>
          <input
            class="form-control input-sm"
            type="number"
            step="0.01"
            style="width: 80px;"
            placeholder={$_("SearchConf.Lat")}
            bind:value={conf.geo.lat}
          />
          <span>,</span>
          <input
            class="form-control input-sm"
            type="number"
            step="0.01"
            style="width: 80px;"
            placeholder={$_("SearchConf.Long")}
            bind:value={conf.geo.long}
          />
          <span class="text-small">{$_("SearchConf.From")}</span>
          <input
            class="form-control input-sm"
            type="number"
            step="5"
            style="width: 80px;"
            placeholder={$_("SearchConf.Dist")}
            bind:value={conf.geo.range}
          />
          <span class="text-small">{$_("SearchConf.KMRange")}</span>
        {/if}
      </div>
    </div>
  {/if}

  <!-- 下部展開エリア 3: 自然言語(AI) 正規表現生成アシスト（hasLLMかつconf.mode === 'ai'の時） -->
  {#if hasLLM && conf.mode === "ai"}
    <div class="conf-sub-row ai-helper-box">
      <span class="conf-label ai-label">✨ {$_("SearchConf.AI")}</span>
      <div class="d-flex flex-column flex-1 gap-1">
        <div class="d-flex flex-items-center gap-2">
          <input
            class="form-control input-sm flex-1"
            type="text"
            placeholder={$_("SearchConf.AIPromptPlaceholder")}
            bind:value={aiPrompt}
            on:compositionstart={() => {
              isComposing = true;
            }}
            on:compositionend={() => {
              setTimeout(() => {
                isComposing = false;
              }, 50);
            }}
            on:keydown={(e) => {
              if (e.key === "Enter") {
                if (isComposing || e.isComposing || e.keyCode === 229) {
                  return;
                }
                generateRegex();
              }
            }}
          />
          {#if !aiBusy}
            <button class="btn btn-sm btn-primary" type="button" on:click={generateRegex}>
              {$_("SearchConf.AIGenerateBtn")}
            </button>
          {:else}
            <button class="btn btn-sm btn-primary" type="button" disabled>
              {$_("SearchConf.AIGenerating")}
            </button>
            <button class="btn btn-sm btn-danger ml-1" type="button" on:click={cancelAIRegex}>
              {$_("AI.Cancel") || "Cancel"}
            </button>
          {/if}
        </div>
        {#if aiRegex}
          <div class="d-flex flex-items-center gap-2 mt-1">
            <span class="text-bold text-small">{$_("SearchConf.AIRegexLabel")}:</span>
            <code class="ai-regex-preview flex-1">{aiRegex}</code>
            {#if aiExplain}
              <span class="text-small color-fg-muted">({aiExplain})</span>
            {/if}
            <button class="btn btn-sm btn-outline" type="button" on:click={applyAIRegex}>
              {$_("SearchConf.AIApplyBtn")}
            </button>
          </div>
        {/if}
        {#if aiError}
          <div class="color-fg-danger text-small mt-1">{aiError}</div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- 下部展開エリア 4: 全文検索クエリ作成ヘルパー (Bluge かつ conf.mode === 'full' の時のみ) -->
  {#if storageEngine === "bluge" && conf.mode === "full"}
    {#if hasStringField}
      <div class="conf-sub-row">
        <span class="conf-label">{$_("SearchConf.Keyword")}</span>
        <div class="d-flex flex-items-center flex-wrap flex-1 gap-2">
          <select class="form-select select-sm" bind:value={conf.keyword.field}>
            <option value="">{$_("SearchConf.All")}</option>
            {#each fields as f}
              {#if !f.startsWith("_") && getFieldType(f) == "string"}
                <option value={f}>{getFieldName(f)}</option>
              {/if}
            {/each}
          </select>
          <span class="text-small">{$_("SearchConf.Include")}</span>
          <input
            class="form-control input-sm"
            type="text"
            style="width: 140px;"
            placeholder={$_("SearchConf.Keyword")}
            bind:value={conf.keyword.key}
          />
          <span class="text-small">{$_("SearchConf.Is")}</span>
          <select class="form-select select-sm" bind:value={conf.keyword.mode}>
            <option value="+">{$_("SearchConf.Must")}</option>
            <option value="-">{$_("SearchConf.Exclude")}</option>
          </select>
          <button class="btn btn-sm" type="button" on:click={addKeyword}>
            <Plus16 />
          </button>
          <button class="btn btn-sm" type="button" on:click={loadKeyword}>
            <File16 />
          </button>
        </div>
      </div>
    {/if}
    {#if hasNumberField}
      <div class="conf-sub-row">
        <span class="conf-label">{$_("SearchConf.NumberCond")}</span>
        <div class="d-flex flex-items-center flex-wrap flex-1 gap-2">
          <select class="form-select select-sm" bind:value={conf.number.field}>
            {#each fields as f}
              {#if !f.startsWith("_") && getFieldType(f) == "number"}
                <option value={f}>{getFieldName(f)}</option>
              {/if}
            {/each}
          </select>
          <select class="form-select select-sm" bind:value={conf.number.oper}>
            <option value="<">&lt;</option>
            <option value=">">&gt;</option>
            <option value="=">=</option>
            <option value="<=">&lt;=</option>
            <option value=">=">&gt;=</option>
          </select>
          <input
            class="form-control input-sm"
            type="text"
            style="width: 100px;"
            placeholder={$_("SearchConf.Number")}
            bind:value={conf.number.value}
          />
          <button class="btn btn-sm" type="button" on:click={addNumber}>
            <Plus16 />
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .search-conf-container {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 2px 4px;
    font-size: 13px;
  }

  .conf-history-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding-bottom: 6px;
    border-bottom: 1px solid var(--color-border-muted, #30363d);
  }

  .conf-main-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px 24px;
  }

  @media (max-width: 768px) {
    .conf-main-grid {
      grid-template-columns: 1fr;
    }
  }

  .conf-col {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .conf-item {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .conf-label {
    width: 120px;
    flex-shrink: 0;
    font-weight: 500;
    color: var(--color-fg-muted, #8b949e);
  }

  .conf-control {
    flex: 1;
    min-width: 0;
  }

  .autocomplete-container {
    flex: 1;
    min-width: 0;
  }

  .conf-sub-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding-top: 6px;
    border-top: 1px dashed var(--color-border-muted, #30363d);
  }

  .ai-helper-box {
    background: rgba(56, 139, 253, 0.08);
    border: 1px solid rgba(56, 139, 253, 0.25);
    border-radius: 6px;
    padding: 8px 12px;
  }

  .ai-label {
    color: #58a6ff;
    font-weight: 600;
  }

  .ai-regex-preview {
    background: var(--color-canvas-subtle, #161b22);
    border: 1px solid var(--color-border-default, #30363d);
    padding: 3px 8px;
    border-radius: 4px;
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 12px;
    word-break: break-all;
  }

  .gap-1 {
    gap: 4px;
  }

  .gap-2 {
    gap: 8px;
  }
</style>
