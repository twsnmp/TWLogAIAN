<script>
  import { MortarBoard16, PaperAirplane16, Sun16, Moon16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  let version = "1.0.0(xxxxx)";

  window.go.main.App.GetVersion().then((v) => {
    version = v;
  });
  const openURL = (url) => {
    window.go.main.App.OpenURL(url);
  };
  const feedback = () => {
    dispatch("done", { page: "feedback" });
  };
  const workdir = () => {
    dispatch("done", { page: "workdir" });
  };

  let dark = true;
  window.go.main.App.GetDark().then((v) => {
    if (!v) {
      toggleDark()
    }
  });
  const toggleDark = () => {
    dark = !dark;
    window.go.main.App.SetDark(dark)
    const e = document.querySelector("body");
    if (e) {
      e.dataset.colorMode = dark ? "dark" : "light";
      e.dataset.darkThme = dark ? "dark" : "light";
    }
  };
</script>

<div class="Box mx-auto" style="max-width: 500px;">
  <div class="Box-header">
    <h3 class="Box-title">ようこそ TWLogAIANへ
      <button class="btn float-right" type="button" on:click={toggleDark}>
        {#if dark}
          <Sun16 />
        {:else}
          <Moon16 />
       {/if}
      </button>
    </h3>
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
        href="##"
        on:click={() => {
          openURL("https://note.com/twsnmp/m/meed0d0ddab5e");
        }}>Noteのマガジン</a
      >に書いています。<br />
      ソースコードは
      <a
        href="##"
        on:click={() => {
          openURL("https://github.com/twsnmp/TWLogAIAN");
        }}
      >
        GitHUB
      </a>にあります。<br />
      バグや要望は＜フィードバック＞か
      <a
        href="##"
        on:click={() => {
          openURL("https://github.com/twsnmp/TWLogAIAN/issues");
        }}
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
  <div class="Box-footer text-right">
    <button class="btn btn-primary mr-1" type="button" on:click={workdir}>
      <MortarBoard16 />
      分析をはじめる
    </button>
    <button class="btn btn-danger" type="button" on:click={feedback}>
      <PaperAirplane16 />
      フィードバック
    </button>
  </div>
</div>

<style>
  #logo {
    height: 200px;
    width: 200px;
    transform: rotateY(560deg);
    animation: turn 3.5s ease-out forwards 1s;
  }

  @keyframes turn {
    100% {
      transform: rotateY(0deg);
    }
  }
</style>
