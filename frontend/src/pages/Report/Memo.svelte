<script>
  import { X16, Copy16,Pencil16,Trash16,XCircle16,Stop16,Info16,Circle16,Check16 } from "svelte-octicons";
  import { createEventDispatcher, onMount } from "svelte";
  import * as echarts from "echarts";
  import { _ } from '../../i18n/i18n';

  let memos = [];
  let editMode = false;
  let memo = {
    Time: 0,
    Type: "",
    Memo: "",
    Log: "",
  }
  const dispatch = createEventDispatcher();

  const updateMemo = () => {
    window.go.main.App.GetMemos().then((r)=>{
      memos = r || [];
    });
  }

  const editMemo = (i) => {
    if (i >= 0 && i < memos.length) {
      memo = memos[i];
      editMode = true;
    }
  }

  const saveMemo = () => {
    window.go.main.App.SetMemo(memo).then(()=>{
      updateMemo();
      editMode = false;
    })
  }

  const formatTime = (t) => {
    return echarts.time.format(new Date(t / (1000*1000)), "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}")
  }

  const deleteMemo = (i) => {
    if (i >= 0 && i < memos.length) {
      window.go.main.App.DeleteMemo(memos[i]).then(()=>{
        updateMemo()
      })
    }
  }

  onMount(() => {
    updateMemo()
  });
  let infoMsg = "";
  let errorMsg = "";
  const copy = () => {
    let copyText = "";
    memos.forEach((e)=> {
      copyText += formatTime(e.Time) + " " + getTypeName(e.Type) + " " + e.Memo
         + " " + e.Diff + "\n" + e.Log + "\n\n";
    });
    if (!navigator.clipboard || !navigator.clipboard) {
      errorMsg = $_('Memo.CantCopy');
      return;
    } 
    navigator.clipboard.writeText(copyText).then(() => {
      infoMsg = $_('Memo.Copied')
      setTimeout(() => {
        infoMsg = "";
      }, 2000);
    }, () => {
      errorMsg = $_('Memo.CopyError');
    });
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg =  "";
  }

  const back = () => {
    dispatch("done", {});
  };

  const getTypeName = (t) => {
    switch(t) {
    case "info":
      return $_('Memo.TypeNameInfo');
    case "warn":
      return $_('Memo.TypeNameWarn');
    case "error":
      return $_('Memo.TypeNameError');
    }
    return "";
  }

</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">$_('Memo.Title')</h3>
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
{#if editMode}
    <div class="Box-row">
      <select
        class="form-select"
        bind:value={memo.Type}
      >
        <option value="">$_('Memo.Deffault')</option>
        <option value="info">$_('Memo.Info')</option>
        <option value="warn">$_('Memo.Warn')</option>
        <option value="error">$_('Memo.Error')</option>
      </select>
      <input
        class="form-control"
        type="text"
        placeholder="$_('Memo.Memo')"
        bind:value={memo.Memo}
      />
      <button class="btn btn-secondary" type="button" on:click={() => {editMode = false}}>
        <X16 />
      </button>
      <button class="btn btn-primary" type="button" on:click={saveMemo}>
        <Check16 />
      </button>
    </div>
  {/if}
  <div class="Box-body">
    {#each memos as { Time, Type, Memo, Log, Diff },i}
      <div class="TimelineItem">
        {#if Type == "error"}
          <div class="TimelineItem-badge color-bg-danger-emphasis color-fg-on-emphasis">
            <XCircle16 />
          </div>
        {:else if Type =="warn"}
          <div class="TimelineItem-badge color-bg-attention-emphasis color-fg-on-emphasis">
            <Stop16 />
          </div>
        {:else if Type =="info"}
          <div class="TimelineItem-badge color-bg-accent-emphasis color-fg-on-emphasis">
            <Info16 />
          </div>
        {:else}
          <div class="TimelineItem-badge">
            <Circle16 />
          </div>
         {/if}
        <div class="TimelineItem-body">
          <p class="h4">
            {formatTime(Time)} {getTypeName(Type)}{Memo} | {Diff}
          </p>
          <p class="f6">
            {Log}
          </p>
          <div>
             <button class="btn btn-danger" type="button" on:click={() => deleteMemo(i)}>
              <Trash16 />
            </button>
            <button class="btn btn-primary" type="button" on:click={() => editMemo(i) }>
              <Pencil16 />
            </button>
          </div>
        </div>
      </div>
    {/each}
  </div>
  <div class="Box-footer text-right">
    {#if memos.length > 0}
      <button class="btn btn-secondary" type="button" on:click={copy}>
        <Copy16 />
        $_('Memo.CopyBtn')
      </button>
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      $_('Memo.BackBtn')
    </button>
  </div>
</div>
<div id="memoCopy" />
