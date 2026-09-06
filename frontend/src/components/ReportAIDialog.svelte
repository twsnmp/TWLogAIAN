<script>
  import {
    X16,
    Copy16,
    Check16,
    Sync16,
    Download16,
    CommentDiscussion16,
    Alert16,
  } from "svelte-octicons";
  import { copyText } from "svelte-copy";
  import { _, getLocale } from "../i18n/i18n";
  import { renderMarkdown } from "../js/markdown";
  import { ExportMarkdown } from "../../wailsjs/go/main/App";
  import catImage from "../assets/images/ai_cat_search.png";

  export let show = false;
  export let title = "";
  export let exportFilename = "report_ai_explanation";
  export let analyzeFunc = undefined;
  export let lang = getLocale() === "en" ? "en" : "ja";

  let loading = false;
  let content = "";
  let error = "";
  let copied = false;
  let hasAnalyzed = false;

  // AI思考中に浮遊させるデータ分析シンボル
  const tokens = [
    { text: "∑", top: "16%", left: "8%", delay: "0s" },
    { text: "f(x)", top: "74%", left: "6%", delay: "0.8s" },
    { text: "Δt", top: "12%", left: "76%", delay: "1.6s" },
    { text: "σ²", top: "75%", left: "78%", delay: "2.4s" },
    { text: "P(A|B)", top: "44%", left: "84%", delay: "1.2s" },
    { text: "λ", top: "48%", left: "2%", delay: "2.0s" },
  ];

  const analyze = async (targetLang = lang) => {
    if (!analyzeFunc) return;
    loading = true;
    error = "";
    content = "";
    copied = false;
    try {
      const resp = await analyzeFunc(targetLang);
      if (resp && resp.Error) {
        error = resp.Error;
      } else if (resp && resp.Answer) {
        content = resp.Answer;
      } else {
        error = $_("ReportAI.NoContent");
      }
    } catch (e) {
      error = e.message || String(e);
    } finally {
      loading = false;
      hasAnalyzed = true;
    }
  };

  const cancelLoading = () => {
    loading = false;
    show = false;
  };

  const setLang = (newLang) => {
    if (lang === newLang) return;
    lang = newLang;
    if (show && !loading) {
      analyze(newLang);
    }
  };

  const copyReport = async () => {
    if (!content) return;
    await copyText(content);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  };

  const exportReport = async () => {
    if (!content) return;
    try {
      await ExportMarkdown(exportFilename, content);
    } catch (e) {
      error = e.message || String(e);
    }
  };

  const close = () => {
    show = false;
  };

  $: if (show && !hasAnalyzed && !loading) {
    analyze(lang);
  }

  $: if (!show) {
    hasAnalyzed = false;
    loading = false;
  }

  $: renderedContent = renderMarkdown(content || "");
</script>

{#if show}
  {#if loading}
    <!-- ローディング時: サイバー猫アニメーションモーダル -->
    <div class="report-ai-backdrop" role="dialog" aria-modal="true">
      <div class="report-ai-loading-card">
        <div class="animation-container ai-mode">
          <!-- 外側の回転サイバーリング -->
          <div class="cyber-ring ring-outer"></div>
          <div class="cyber-ring ring-inner"></div>

          <!-- 猫イラスト本体 -->
          <div class="cat-avatar-wrapper">
            <img src={catImage} alt="AI Analyzing Cat" class="cat-image" />
            <!-- 猫のホログラムパネル上を走るスキャンライン光 -->
            <div class="scan-beam"></div>
          </div>

          <!-- 浮遊パーティクル -->
          <div class="particles-layer">
            {#each tokens as token}
              <span
                class="regex-token"
                style="top: {token.top}; left: {token.left}; animation-delay: {token.delay};"
              >
                {token.text}
              </span>
            {/each}
          </div>
          <div class="ai-pulse-wave"></div>
        </div>

        <div class="status-content">
          <div class="status-badge badge-ai">
            ✨ AI REPORT EXPLANATION
          </div>

          <h3 class="status-title">
            {$_("ReportAI.Loading")}
            <span class="AnimatedEllipsis"></span>
          </h3>

          <p class="status-subtext">
            {title || $_("ReportAI.Title")}
          </p>
        </div>

        <div class="action-footer mt-3">
          <button
            class="btn btn-danger cancel-btn"
            type="button"
            on:click={cancelLoading}
          >
            <X16 />
            <span class="ml-1">{$_("ReportAI.Cancel")}</span>
          </button>
        </div>
      </div>
    </div>
  {:else}
    <!-- 分析完了時: レポート表示モーダル -->
    <div class="report-ai-backdrop" role="dialog" aria-modal="true">
      <div class="Box report-ai-result-card">
        <div class="Box-header d-flex flex-justify-between flex-items-center">
          <div class="d-flex flex-items-center">
            <span class="color-fg-accent mr-2">
              <CommentDiscussion16 />
            </span>
            <h3 class="Box-title">
              {title || $_("ReportAI.Title")}
            </h3>
          </div>

          <div class="d-flex flex-items-center" style="gap: 8px;">
            <!-- 回答言語切り替えボタン -->
            <div class="d-flex flex-items-center">
              <span class="text-small color-fg-muted font-weight-bold mr-1">
                {$_("ReportAI.Lang")}
              </span>
              <div class="BtnGroup">
                <button
                  class="btn btn-sm BtnGroup-item"
                  class:btn-primary={lang === "ja"}
                  type="button"
                  on:click={() => setLang("ja")}
                >
                  {$_("ReportAI.LangJa")}
                </button>
                <button
                  class="btn btn-sm BtnGroup-item"
                  class:btn-primary={lang === "en"}
                  type="button"
                  on:click={() => setLang("en")}
                >
                  {$_("ReportAI.LangEn")}
                </button>
              </div>
            </div>

            <button
              class="btn btn-sm btn-invisible"
              type="button"
              on:click={close}
            >
              <X16 />
            </button>
          </div>
        </div>

        {#if error}
          <div class="flash flash-error m-3">
            <div class="d-flex flex-items-center">
              <Alert16 class="mr-2" />
              <span>{error}</span>
            </div>
          </div>
        {/if}

        <div class="Box-body p-3 result-body">
          {#if content}
            <div class="markdown-body report-markdown">
              {@html renderedContent}
            </div>
          {:else if !error}
            <p class="color-fg-muted text-center py-4">{$_("ReportAI.NoContent")}</p>
          {/if}
        </div>

        <div class="Box-footer d-flex flex-justify-between flex-items-center">
          <div>
            <button
              class="btn btn-sm"
              type="button"
              on:click={() => analyze(lang)}
            >
              <Sync16 />
              <span class="ml-1">{$_("ReportAI.ReAnalyze")}</span>
            </button>
          </div>

          <div class="d-flex flex-items-center" style="gap: 8px;">
            {#if content}
              <button
                class="btn btn-sm"
                class:btn-primary={copied}
                type="button"
                on:click={copyReport}
              >
                {#if copied}
                  <Check16 />
                  <span class="ml-1">{$_("ReportAI.Copied")}</span>
                {:else}
                  <Copy16 />
                  <span class="ml-1">{$_("ReportAI.Copy")}</span>
                {/if}
              </button>

              <button
                class="btn btn-sm"
                type="button"
                on:click={exportReport}
              >
                <Download16 />
                <span class="ml-1">{$_("ReportAI.Export")}</span>
              </button>
            {/if}

            <button
              class="btn btn-sm btn-secondary"
              type="button"
              on:click={close}
            >
              <X16 />
              <span class="ml-1">{$_("ReportAI.Close")}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
{/if}

<style>
  .report-ai-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(10, 14, 23, 0.78);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    animation: fadeIn 0.25s ease-out;
  }

  .report-ai-loading-card {
    position: relative;
    width: 440px;
    max-width: 92vw;
    background: linear-gradient(165deg, #1b212c 0%, #0d1117 100%);
    border: 1px solid rgba(56, 139, 253, 0.35);
    border-radius: 16px;
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.7), 0 0 25px rgba(56, 139, 253, 0.2);
    padding: 28px 24px 22px;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    animation: scaleUp 0.25s cubic-bezier(0.16, 1, 0.3, 1);
    color: #e6edf3;
  }


  .report-ai-result-card {
    width: 900px;
    max-width: 95vw;
    max-height: 88vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
    border-radius: 12px;
    overflow: hidden;
    animation: scaleUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .result-body {
    overflow-y: auto;
    flex: 1;
    min-height: 250px;
    max-height: calc(88vh - 120px);
  }

  .report-markdown {
    font-size: 14px;
    line-height: 1.6;
  }

  /* アニメーションエリア */
  .animation-container {
    position: relative;
    width: 170px;
    height: 170px;
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .cyber-ring {
    position: absolute;
    border-radius: 50%;
    pointer-events: none;
  }

  .ring-outer {
    inset: -12px;
    border: 2px dashed rgba(0, 240, 255, 0.75);
    animation: spinRing 8s linear infinite;
  }

  .ring-inner {
    inset: -5px;
    border: 1.5px solid transparent;
    border-top-color: rgba(0, 255, 196, 0.95);
    border-bottom-color: rgba(0, 180, 255, 0.85);
    animation: spinRingReverse 9s linear infinite;
  }

  .cat-avatar-wrapper {
    position: relative;
    width: 154px;
    height: 154px;
    border-radius: 50%;
    overflow: hidden;
    box-shadow: 0 0 26px rgba(0, 240, 255, 0.65), 0 0 12px rgba(0, 255, 196, 0.45);
    background: #0d1117;
    z-index: 1;
  }

  .cat-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    transform: scale(1.04);
  }

  .scan-beam {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 6px;
    background: linear-gradient(
      180deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(0, 240, 255, 0.8) 50%,
      rgba(255, 255, 255, 0.9) 100%
    );
    filter: drop-shadow(0 0 8px #00f0ff);
    animation: scanMove 2.2s ease-in-out infinite;
    pointer-events: none;
  }

  .particles-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 3;
  }

  .regex-token {
    position: absolute;
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 13px;
    font-weight: 700;
    color: #58a6ff;
    text-shadow: 0 0 8px rgba(88, 166, 255, 0.8);
    opacity: 0;
    animation: floatToken 3.2s ease-in-out infinite;
  }

  .ai-pulse-wave {
    position: absolute;
    inset: -4px;
    border-radius: 50%;
    border: 2px solid rgba(0, 240, 255, 0.7);
    animation: pulseWave 2s cubic-bezier(0.25, 0.46, 0.45, 0.94) infinite;
    pointer-events: none;
  }

  .status-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
  }

  .status-badge {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    padding: 3px 10px;
    border-radius: 20px;
    margin-bottom: 10px;
    display: inline-flex;
    align-items: center;
    background: rgba(163, 113, 247, 0.15);
    color: #d2a8ff;
    border: 1px solid rgba(163, 113, 247, 0.4);
    box-shadow: 0 0 12px rgba(163, 113, 247, 0.2);
  }

  .status-title {
    font-size: 15px;
    font-weight: 600;
    margin: 0 0 6px 0;
    color: #f0f6fc;
    display: flex;
    align-items: center;
    justify-content: center;
    white-space: nowrap;
  }


  .status-subtext {
    font-size: 13px;
    color: #8b949e;
    margin: 0;
    max-width: 90%;
    line-height: 1.45;
  }

  .cancel-btn {
    border-radius: 20px;
    padding: 4px 16px;
    font-size: 12px;
    font-weight: 500;
  }

  @keyframes spinRing {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  @keyframes spinRingReverse {
    from {
      transform: rotate(360deg);
    }
    to {
      transform: rotate(0deg);
    }
  }

  @keyframes scanMove {
    0%,
    100% {
      top: 5%;
      opacity: 0.2;
    }
    50% {
      top: 90%;
      opacity: 0.9;
    }
  }

  @keyframes floatToken {
    0% {
      transform: translateY(6px) scale(0.85);
      opacity: 0;
    }
    40% {
      opacity: 0.9;
    }
    80% {
      transform: translateY(-8px) scale(1.05);
      opacity: 0.7;
    }
    100% {
      transform: translateY(-16px) scale(1.1);
      opacity: 0;
    }
  }

  @keyframes pulseWave {
    0% {
      transform: scale(0.95);
      opacity: 0.8;
    }
    100% {
      transform: scale(1.22);
      opacity: 0;
    }
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes scaleUp {
    from {
      transform: scale(0.94);
      opacity: 0;
    }
    to {
      transform: scale(1);
      opacity: 1;
    }
  }
</style>
