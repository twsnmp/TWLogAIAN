<script>
  import {
    X16,
    PaperAirplane16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  let feedbackSent = false;
  let feedbackErr = false;
  let feedbackMsg = "";
  let feedbackWait = false;

  const close = () => {
    dispatch("done", { page: "wellcome" });
  };
  const sendFeedBack = () => {
    feedbackWait = true;
    window.go.main.App.SendFeedBack(feedbackMsg).then((r) => {
      feedbackWait = false;
      if (r) {
        feedbackSent = true;
        feedbackMsg = "";
      } else {
        feedbackErr = true;
      }
    });
  };
  const clearMsg = () => {
    feedbackSent = false;
    feedbackErr = false;
  };
</script>

<div class="Box mx-auto" style="max-width: 800px;">
  <div class="Box-header">
    <h3 class="Box-title">フィードバック</h3>
  </div>
  {#if feedbackWait}
    <div class="flash">フィードバックを送信中<span class="AnimatedEllipsis"></span></div>
  {:else if feedbackErr}
    <div class="flash flash-error">
      フィードバックの送信に失敗しました
      <button
        class="flash-close js-flash-close"
        type="button"
        aria-label="Close"
        on:click={clearMsg}
      >
        <X16 />
      </button>
    </div>
  {:else if feedbackSent}
    <div class="flash">
      フィードバックを送信しました
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
      <div class="form-group" class:errored={feedbackErr}>
        <div class="form-group-header">
          <label for="feedbackMsg">メッセージ</label>
        </div>
        <div class="form-group-body">
          <textarea
            class="form-control"
            id="feedbackMsg"
            bind:value={feedbackMsg}
          />
        </div>
      </div>
    </form>
  </div>
  <div class="Box-footer text-right">
    {#if !feedbackWait}
      <button class="btn btn-secondary mr-2" type="button" on:click={close}>
        <X16 />
        キャンセル
      </button>
      {#if feedbackMsg != ""}
        <button class="btn btn-primary" type="button" on:click={sendFeedBack}>
          <PaperAirplane16 />
          送信
        </button>
      {/if}
    {/if}
  </div>
</div>
