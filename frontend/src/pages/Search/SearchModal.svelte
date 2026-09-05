<script>
  import { createEventDispatcher } from "svelte";
  import { X16 } from "svelte-octicons";
  import { _ } from "../../i18n/i18n";
  import catImage from "../../assets/images/ai_cat_search.png";

  export let show = false;
  // status: "ai" (AI正規表現生成中) | "search" (ログ検索中)
  export let status = "search";
  export let detailMessage = "";

  const dispatch = createEventDispatcher();

  const handleCancel = () => {
    dispatch("cancel");
  };

  // AI思考中に浮遊させる正規表現シンボル
  const regexTokens = [
    { text: ".*", top: "18%", left: "10%", delay: "0s" },
    { text: "\\d+", top: "72%", left: "8%", delay: "0.8s" },
    { text: "[a-z]", top: "12%", left: "75%", delay: "1.6s" },
    { text: "^...$", top: "75%", left: "78%", delay: "2.4s" },
    { text: "\\w+", top: "45%", left: "85%", delay: "1.2s" },
    { text: "(?:)", top: "48%", left: "3%", delay: "2.0s" },
  ];
</script>

{#if show}
  <div class="search-modal-backdrop" role="dialog" aria-modal="true">
    <div class="search-modal-card">
      <!-- アニメーション表示エリア -->
      <div class="animation-container" class:ai-mode={status === "ai"}>
        <!-- 外側の回転サイバーリング -->
        <div class="cyber-ring ring-outer"></div>
        <div class="cyber-ring ring-inner"></div>

        <!-- レーダースキャン（検索中モード時） -->
        {#if status === "search"}
          <div class="radar-scan"></div>
        {/if}

        <!-- 猫イラスト本体 -->
        <div class="cat-avatar-wrapper">
          <img src={catImage} alt="Searching Cat" class="cat-image" />
          <!-- 猫のホログラムパネル上を走るスキャンライン光 -->
          <div class="scan-beam"></div>
        </div>

        <!-- AI思考中モードの正規表現浮遊パーティクル -->
        {#if status === "ai"}
          <div class="particles-layer">
            {#each regexTokens as token}
              <span
                class="regex-token"
                style="top: {token.top}; left: {token.left}; animation-delay: {token.delay};"
              >
                {token.text}
              </span>
            {/each}
          </div>
          <div class="ai-pulse-wave"></div>
        {/if}
      </div>

      <!-- ステータス・テキスト表示エリア -->
      <div class="status-content">
        <div class="status-badge" class:badge-ai={status === "ai"}>
          {#if status === "ai"}
            ✨ AI ASSISTANT
          {:else}
            🔍 LOG SEARCH
          {/if}
        </div>

        <h3 class="status-title">
          {#if status === "ai"}
            {$_("SearchModal.AIThinking") || "AIが正規表現を考案中..."}
          {:else}
            {$_("SearchModal.Searching") || "ログを検索中..."}
          {/if}
          <span class="AnimatedEllipsis"></span>
        </h3>

        <p class="status-subtext">
          {#if detailMessage}
            {detailMessage}
          {:else if status === "ai"}
            {$_("SearchModal.AISubText") || "入力された条件から最適なパターンを組み立てています"}
          {:else}
            {$_("SearchModal.SearchSubText") || "条件に合致するログを探索しています"}
          {/if}
        </p>
      </div>

      <!-- 中止（キャンセル）ボタン -->
      <div class="action-footer">
        <button
          class="btn btn-danger cancel-btn"
          type="button"
          on:click={handleCancel}
        >
          <X16 />
          <span class="ml-1">{$_("SearchModal.Cancel") || "中止"}</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .search-modal-backdrop {
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

  .search-modal-card {
    position: relative;
    width: 380px;
    max-width: 90vw;
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

  /* サイバーリング演出 */
  .cyber-ring {
    position: absolute;
    border-radius: 50%;
    pointer-events: none;
  }

  .ring-outer {
    inset: -12px;
    border: 2px dashed rgba(56, 139, 253, 0.6);
    animation: spinRing 14s linear infinite;
  }

  .ring-inner {
    inset: -5px;
    border: 1.5px solid transparent;
    border-top-color: rgba(63, 185, 80, 0.8);
    border-bottom-color: rgba(56, 139, 253, 0.8);
    animation: spinRingReverse 9s linear infinite;
  }

  .ai-mode .ring-outer {
    border-color: rgba(0, 240, 255, 0.75);
    animation-duration: 8s;
  }

  .ai-mode .ring-inner {
    border-top-color: rgba(0, 255, 196, 0.95);
    border-bottom-color: rgba(0, 180, 255, 0.85);
  }

  /* レーダースキャン（検索中モード） */
  .radar-scan {
    position: absolute;
    inset: -6px;
    border-radius: 50%;
    background: conic-gradient(
      from 0deg,
      rgba(56, 139, 253, 0.35) 0deg,
      rgba(56, 139, 253, 0) 70deg
    );
    animation: radarSweep 2.5s linear infinite;
    pointer-events: none;
    z-index: 2;
  }

  /* 猫画像コンテナ */
  .cat-avatar-wrapper {
    position: relative;
    width: 154px;
    height: 154px;
    border-radius: 50%;
    overflow: hidden;
    box-shadow: 0 0 20px rgba(56, 139, 253, 0.4);
    background: #0d1117;
    z-index: 1;
  }

  .ai-mode .cat-avatar-wrapper {
    box-shadow: 0 0 26px rgba(0, 240, 255, 0.65), 0 0 12px rgba(0, 255, 196, 0.45);
  }

  .cat-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  /* 猫の手元のホログラムパネル上を走るスキャンビーム */
  .scan-beam {
    position: absolute;
    top: -100%;
    left: 0;
    right: 0;
    height: 60px;
    background: linear-gradient(
      to bottom,
      transparent 0%,
      rgba(56, 139, 253, 0.15) 50%,
      rgba(63, 185, 80, 0.4) 100%
    );
    animation: scanMotion 2.2s ease-in-out infinite;
    pointer-events: none;
  }

  .ai-mode .scan-beam {
    background: linear-gradient(
      to bottom,
      transparent 0%,
      rgba(0, 240, 255, 0.25) 50%,
      rgba(0, 255, 196, 0.5) 100%
    );
    animation-duration: 1.8s;
  }

  /* AIモード波紋パルス */
  .ai-pulse-wave {
    position: absolute;
    inset: -8px;
    border-radius: 50%;
    border: 2px solid rgba(0, 240, 255, 0.65);
    animation: pulseRing 2s cubic-bezier(0.25, 0.8, 0.25, 1) infinite;
    pointer-events: none;
  }

  /* 正規表現浮遊パーティクル */
  .particles-layer {
    position: absolute;
    inset: -30px;
    pointer-events: none;
    z-index: 3;
  }

  .regex-token {
    position: absolute;
    font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    font-size: 13px;
    font-weight: 700;
    color: #00f0ff;
    text-shadow: 0 0 8px rgba(0, 240, 255, 0.9), 0 0 14px rgba(0, 255, 196, 0.5);
    opacity: 0;
    animation: floatToken 3s ease-in-out infinite;
  }

  /* ステータス情報エリア */
  .status-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    margin-bottom: 22px;
  }

  .status-badge {
    display: inline-block;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    padding: 3px 10px;
    border-radius: 20px;
    background: rgba(56, 139, 253, 0.15);
    border: 1px solid rgba(56, 139, 253, 0.4);
    color: #58a6ff;
  }

  .status-badge.badge-ai {
    background: rgba(0, 240, 255, 0.12);
    border-color: rgba(0, 240, 255, 0.5);
    color: #00f0ff;
    text-shadow: 0 0 6px rgba(0, 240, 255, 0.4);
  }

  .status-title {
    font-size: 17px;
    font-weight: 600;
    color: #f0f6fc;
    margin: 4px 0 0 0;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 26px;
  }

  .status-subtext {
    font-size: 12px;
    color: #8b949e;
    margin: 0;
    max-width: 320px;
    line-height: 1.4;
  }

  /* フッター / 中止ボタン */
  .action-footer {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .cancel-btn {
    min-width: 130px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
    padding: 6px 16px;
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(218, 54, 51, 0.3);
    transition: all 0.15s ease;
  }

  .cancel-btn:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 16px rgba(218, 54, 51, 0.4);
  }

  /* キーフレームアニメーション */
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
      opacity: 0;
      transform: scale(0.92);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
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

  @keyframes radarSweep {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  @keyframes scanMotion {
    0% {
      top: -60px;
      opacity: 0;
    }
    20% {
      opacity: 0.8;
    }
    80% {
      opacity: 0.8;
    }
    100% {
      top: 160px;
      opacity: 0;
    }
  }

  @keyframes pulseRing {
    0% {
      transform: scale(0.96);
      opacity: 0.9;
    }
    50% {
      opacity: 0.4;
    }
    100% {
      transform: scale(1.15);
      opacity: 0;
    }
  }

  @keyframes floatToken {
    0% {
      transform: translateY(12px) scale(0.85);
      opacity: 0;
    }
    30% {
      opacity: 0.9;
    }
    70% {
      opacity: 0.9;
      transform: translateY(-8px) scale(1.05);
    }
    100% {
      transform: translateY(-22px) scale(0.9);
      opacity: 0;
    }
  }
</style>
