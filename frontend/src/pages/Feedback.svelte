<script>
  import { X16, PaperAirplane16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import {SendFeedBack} from '../../wailsjs/go/main/App';
  import { _ } from '../i18n/i18n';

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
    SendFeedBack(feedbackMsg).then((r) => {
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
    <h3 class="Box-title">{$_('Feedback.Title')}</h3>
  </div>
  {#if feedbackWait}
    <div class="flash">
      {$_('Feedback.SendMsg')}<span class="AnimatedEllipsis" />
    </div>
  {:else if feedbackErr}
    <div class="flash flash-error">
      {$_('Feedback.SendError')}
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
      {$_('Feedback.SentMsg')}
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
    <div class="form-group" class:errored={feedbackErr}>
      <div class="form-group-header">
        <label for="feedbackMsg">{$_('Feedback.Message')}</label>
      </div>
      <div class="form-group-body">
        <textarea
          class="form-control"
          id="feedbackMsg"
          bind:value={feedbackMsg}
        />
      </div>
    </div>
  </div>
  <div class="Box-footer text-right">
    {#if !feedbackWait}
      <button class="btn btn-secondary mr-1" type="button" on:click={close}>
        <X16 />
        {$_('Feedback.CancelBtn')}
      </button>
      {#if feedbackMsg != ""}
        <button class="btn btn-primary" type="button" on:click={sendFeedBack}>
          <PaperAirplane16 />
          {$_('Feedback.SendBtn')}
        </button>
      {/if}
    {/if}
  </div>
</div>
