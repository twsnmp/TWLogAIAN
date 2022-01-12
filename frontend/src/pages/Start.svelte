<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  let version = "1.0.0(xxxxx)";
  let state = "wellcome";
  let workdir = "";

  const setWorkDir = () => {
    dispatch("done", {});
  };
  const openURL = (url) => {
    console.log(url);
  };
  const cancel = () => {
    state = "wellcome";
  }
  const showSelect = () => {
    state = "select";
  }
  const showFeedBack = () => {
    state = "feedback";
  }
  const sendFeedBack = () => {
    state = "wellcome";
  }
</script>

{#if state == "wellcome"}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">ようこそ TWLogAIANへ</h3>
    </div>
    <div class="Box-body">
      <div class="mx-auto" style="max-width: 200px;">
        <img id="logo" alt="TWLogAIAN Logo" src="./assets/images/appicon.png" />
      </div>
      <hr />
      <p>
        TWLogAIANはAIアシスト付きログ分析ツールです。<br />
        使い方は
        <a
          href="https://note.com/twsnmp/m/meed0d0ddab5e"
          on:click={openURL("https://note.com/twsnmp/m/meed0d0ddab5e")}
          >Noteのマガジン</a
        >に書いています。<br />
        ソースコードは
        <a
          href="https://github.com/twsnmp/TWLogAIAN"
          on:click={openURL("https://github.com/twsnmp/TWLogAIAN")}
        >
          GitHUB
        </a>にあります。<br />
        バグや要望は＜フィードバック＞か
        <a
          href="https://github.com/twsnmp/TWLogAIAN/issues"
          on:click={openURL("https://github.com/twsnmp/TWLogAIAN/issues")}
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
{:else if (state = "select")}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">作業ディレクトリの選択</h3>
    </div>
    <div class="Box-body">
      <form>
        <div class="input-group">
          <input class="form-control" type="text" placeholder="作業ディレクトリ" aria-label="作業ディレクトリ" value="{workdir}">
          <span class="input-group-button">
            <button class="btn" type="button">
              <!-- <%= octicon "clippy" %> -->
              <svg class="octicon octicon-clippy" viewBox="0 0 14 16" version="1.1" width="14" height="16" aria-hidden="true"><path fill-rule="evenodd" d="M2 13h4v1H2v-1zm5-6H2v1h5V7zm2 3V8l-3 3 3 3v-2h5v-2H9zM4.5 9H2v1h2.5V9zM2 12h2.5v-1H2v1zm9 1h1v2c-.02.28-.11.52-.3.7-.19.18-.42.28-.7.3H1c-.55 0-1-.45-1-1V4c0-.55.45-1 1-1h3c0-1.11.89-2 2-2 1.11 0 2 .89 2 2h3c.55 0 1 .45 1 1v5h-1V6H1v9h10v-2zM2 5h8c0-.55-.45-1-1-1H8c-.55 0-1-.45-1-1s-.45-1-1-1-1 .45-1 1-.45 1-1 1H3c-.55 0-1 .45-1 1z"></path></svg>
            </button>
          </span>
        </div>
        <p>最近使ったディレクトリ</p>
        <select class="form-select" aria-label="Important decision">
          <option>新規</option>
          <option value="option 2">テスト</option>
        </select>        
      </form>
    </div>
    <div class="Box-footer">
      <button class="btn btn-outline" type="button" on:click={setWorkDir}>選択</button>
      <button class="btn" type="button" on:click={cancel}>キャンセル</button>
    </div>
  </div>
{:else}
  <div class="Box mx-auto" style="max-width: 500px;">
    <div class="Box-header">
      <h3 class="Box-title">フィードバック</h3>
    </div>
    <div class="Box-body">
    </div>
    <div class="Box-footer">
      <button class="btn btn-outline" type="button" on:click="{sendFeedBack}">送信</button>
      <button class="btn" type="button" on:click={cancel}>キャンセル</button>
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
