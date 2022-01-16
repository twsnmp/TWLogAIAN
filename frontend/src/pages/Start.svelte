<script>
   import { X16,FileDirectory16,PaperAirplane16, Check16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  let version = "1.0.0(xxxxx)";
  let state = "wellcome";
  let workdir = "";
  let feedbackSent = false;
  let feedbackErr = false;
  let setWorkDirErr = false;
  let feedbackMsg = "";
  let feedbackWait = false;
  let lastWorkDirs = [];
  let selLast = "";

  window.go.main.App.GetVersion().then((v) => {
    version = v;
  });
  window.go.main.App.GetLastWorkDirs().then((wds) => {
    lastWorkDirs = wds;
  });
  const setWorkDir = () => {
    window.go.main.App.SetWorkDir(workdir).then((r) => {
      if(r){
        dispatch("done", {});
      } else {
        setWorkDirErr = true;
      }
    });
  };
  const openURL = (url) => {
    window.go.main.App.OpenURL(url);
  };
  const selectWorkDir = () => {
   window.go.main.App.GetWorkDir().then((d) => {
     workdir = d;
   })
  }
  const cancel = () => {
    state = "wellcome";
  }
  const showSelect = () => {
    state = "select";
  }
  const checkSelectWorkDir = () => {
    if(selLast != "") {
      workdir = selLast;
    }
  }
  const showFeedBack = () => {
    state = "feedback";
  }
  const sendFeedBack = () => {
    feedbackWait = true;
    window.go.main.App.SendFeedBack(feedbackMsg).then((r) => {
      feedbackWait = false;
      if(r){
        feedbackSent = true;
        feedbackMsg = "";
        state = "wellcome";
      } else {
        feedbackErr = true;
      }
    });
  }
  const clearMsg = () => {
    feedbackSent = false;
    feedbackErr = false;
    setWorkDirErr = false;
  }
</script>

{#if state == "wellcome"}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">ようこそ TWLogAIANへ</h3>
    </div>
    {#if feedbackSent}
    <div class="flash">
      フィードバックを送信しました
      <button class="flash-close js-flash-close" type="button" aria-label="Close" on:click={clearMsg}>
        <X16/>
      </button>
    </div>
    {/if}
    <div class="Box-body">
      <div class="mx-auto" style="max-width: 200px;">
        <img id="logo" alt="TWLogAIAN Logo" src="./assets/images/appicon.png" />
      </div>
      <hr />
      <p>
        TWLogAIANはAIアシスト付きログ分析ツールです。<br />
        使い方は
        <a
          href="##"
          on:click={() => {openURL("https://note.com/twsnmp/m/meed0d0ddab5e")}}
          >Noteのマガジン</a
        >に書いています。<br />
        ソースコードは
        <a
          href="##"
          on:click={() => {openURL("https://github.com/twsnmp/TWLogAIAN")}}
        >
          GitHUB
        </a>にあります。<br />
        バグや要望は＜フィードバック＞か
        <a
          href="##"
          on:click={() => {openURL("https://github.com/twsnmp/TWLogAIAN/issues")}}
        >
          GitHubのissue
        </a>からお知らせください。
      </p>
      <hr />
      <p>TWLogAIANを利用いただきありがとうございます。</p>
      <div class="f6">
        <em><small>TWLogAIAN {version}© 2021 Masayuki Yamai</small></em>
      </div>
    </div>
    <div class="Box-footer">
      <button class="btn btn-outline" type="button" on:click={showSelect}>分析をはじめる</button>
      <button class="btn btn-danger" type="button" on:click={showFeedBack}>フィードバック</button>
    </div>
  </div>
{:else if (state == "select")}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">作業フォルダーの選択</h3>
    </div>
    {#if setWorkDirErr}
    <div class="flash flash-error">
       選択した作業フォルダーは利用できません
      <button class="flash-close js-flash-close" type="button" aria-label="Close" on:click={clearMsg}>
        <X16/>
      </button>
    </div>
    {/if}
    <div class="Box-body">
      <form>
        <div class="input-group mb-5">
          <input class="form-control errored" type="text" placeholder="作業フォルダー" aria-label="作業フォルダー" value="{workdir}">
          <span class="input-group-button">
            <button class="btn" type="button" on:click={selectWorkDir}>
              <FileDirectory16 />
            </button>
          </span>
        </div>
        {#if lastWorkDirs.length > 0 }
        <p>最近使ったフォルダー</p>
        <!-- svelte-ignore a11y-no-onchange -->
        <select class="form-select" aria-label="最近使ったフォルダー" bind:value={selLast} on:change={checkSelectWorkDir}>
          <option value="">選択してください。</option>
          {#each lastWorkDirs as d }
           <option value={d}>{d}</option>
          {/each}
        </select>
        {/if}
      </form>
    </div>
    <div class="Box-footer">
      <button class="btn btn-outline" type="button" on:click={setWorkDir}>
        <Check16/>
        選択
      </button>
      <button class="btn" type="button" on:click={cancel}>
        <X16 />
        キャンセル
      </button>
    </div>
  </div>
{:else}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">フィードバック</h3>
    </div>
    {#if feedbackWait}
      <div class="flash">
        フィードバックを送信中
      </div>
    {/if}
    {#if feedbackErr}
    <div class="flash flash-error">
      フィードバックの送信に失敗しました
      <button class="flash-close js-flash-close" type="button" aria-label="Close" on:click={clearMsg}>
        <X16/>
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
            <textarea class="form-control" id="feedbackMsg" bind:value={feedbackMsg}></textarea>
          </div>
        </div>
      </form>
    </div>
    <div class="Box-footer">
      {#if !feedbackWait }
        {#if feedbackMsg != "" }
          <button class="btn btn-outline" type="button" on:click="{sendFeedBack}">
            <PaperAirplane16 />
            送信
          </button>
        {/if}
        <button class="btn" type="button" on:click={cancel}>
          <X16 />
          キャンセル
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  #logo {
    height: 200px;
    width: 200px;
    /* background-color: ghostwhite; */
    transform: rotateY(560deg);
    animation: turn 3.5s ease-out forwards 1s;
  }

  @keyframes turn {
    100% {
      transform: rotateY(0deg);
    }
  }
</style>
