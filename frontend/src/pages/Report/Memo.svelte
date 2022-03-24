<script>
  import { X16, Copy16,Pencil16,Trash16,XCircle16,Stop16,Info16,Circle16,Check16 } from "svelte-octicons";
  import CopyClipBoard from "../../CopyClipBoard.svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import * as echarts from "echarts";

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

  let showCopy = false;
  const copy = () => {
    let copyText = "";
    memos.forEach((e)=> {
      copyText += formatTime(e.Time) + " " + getTypeName(e.Type) + " " + e.Memo
         + " " + e.Diff + "\n" + e.Log + "\n\n";
    });
    showCopy = true;
    const app = new CopyClipBoard({
      target: document.getElementById("memoCopy"),
      props: { copyText },
    });
    app.$destroy();
    setTimeout(()=>{
      showCopy = false;
    },2000);
  };

  const back = () => {
    dispatch("done", {});
  };

  const getTypeName = (t) => {
    switch(t) {
    case "info":
      return "情報:";
    case "warn":
      return "注意:";
    case "error":
      return "エラー:";
    }
    return "";
  }

</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">メモ</h3>
  </div>
  {#if editMode}
    <div class="Box-row">
      <select
        class="form-select"
        bind:value={memo.Type}
      >
        <option value="">デフォルト</option>
        <option value="info">情報</option>
        <option value="warn">注意</option>
        <option value="error">エラー</option>
      </select>
      <input
        class="form-control"
        type="text"
        placeholder="メモ"
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
        コピー
      </button>
      {#if showCopy}
        <span class="branch-name">Copied</span>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      戻る
    </button>
  </div>
</div>
<div id="memoCopy" />
