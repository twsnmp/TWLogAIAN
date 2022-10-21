<script>
  import {
    X16,
    File16,
    FileDirectory16,
    Check16,
    Trash16,
    FileBadge16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  export let logSource;
  let windows = false;

  window.go.main.App.IsWindows().then((r) => {
    windows = r;
  });

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let editMode = logSource && logSource.No > 0;

  const selectLogFolder = () => {
    window.go.main.App.SelectFile("logdir").then((f) => {
      logSource.Path = f;
    });
  };

  const selectLogFile = () => {
    window.go.main.App.SelectFile("logfile").then((f) => {
      logSource.Path = f;
    });
  };

  const selectSSHKey = () => {
    window.go.main.App.SelectFile("sshkey").then((f) => {
      logSource.SSHKey = f;
    });
  };

  const cancel = () => {
    dispatch("done", { update: false });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const save = () => {
    window.go.main.App.UpdateLogSource(logSource).then((e) => {
      errorMsg = e;
      if (e == "") {
        dispatch("done", { update: true });
      }
    });
  };

  const del = () => {
    window.go.main.App.DeleteLogSource(logSource.No).then((e) => {
      errorMsg = e;
      if (e == "") {
        dispatch("done", { update: true });
      }
    });
  };
</script>

<div class="Box-header">
  <h3 class="Box-title">ログソース編集</h3>
</div>
{#if errorMsg != ""}
  <div class="flash flash-error">
    {errorMsg}
    <button
      class="flash-close js-flash-close"
      type="button"
      on:click={clearMsg}
    >
      <X16 />
    </button>
  </div>
{/if}
<div class="Box-body">
  <div class="form-group">
    <div class="form-group-header">
      <h5>タイプ</h5>
    </div>
    <div class="form-group-body">
      <select
        class="form-select"
        bind:value={logSource.Type}
        disabled={editMode}
      >
        <option value="folder">フォルダー</option>
        <option value="file">単一ファイル</option>
        <option value="scp">SCP転送</option>
        {#if windows || logSource.Type == "windows"}
          <option value="windows">Windows</option>
        {/if}
        <option value="cmd">コマンド実行</option>
        <option value="ssh">SSHコマンド実行</option>
        <option value="twsnmp">TWSNMP FC連携</option>
        <option value="gravwell">Gravwell連携</option>
      </select>
    </div>
  </div>
  {#if logSource.Type == "folder"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>フォルダー</h5>
      </div>
      <div class="form-group-body">
        <div class="input-group">
          <input
            class="form-control"
            type="text"
            placeholder="フォルダー"
            bind:value={logSource.Path}
          />
          <span class="input-group-button">
            <button class="btn" type="button" on:click={selectLogFolder}>
              <FileDirectory16 />
            </button>
          </span>
        </div>
      </div>
    </div>
  {/if}
  {#if logSource.Type == "file"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>単一ファイル</h5>
      </div>
      <div class="form-group-body">
        <div class="input-group">
          <input
            class="form-control"
            type="text"
            placeholder="ファイル"
            bind:value={logSource.Path}
          />
          <span class="input-group-button">
            <button class="btn" type="button" on:click={selectLogFile}>
              <File16 />
            </button>
          </span>
        </div>
      </div>
    </div>
  {/if}
  {#if logSource.Type == "cmd" || logSource.Type == "ssh"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>コマンド</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control input-block"
          type="text"
          placeholder="コマンド"
          bind:value={logSource.Path}
        />
      </div>
      <p class="note error" id="scppath-input-validation">
        コマンドを指定してください
      </p>
    </div>
  {/if}
  {#if logSource.Type == "scp" || logSource.Type == "ssh"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>サーバー</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control input-block"
          type="text"
          placeholder="サーバー"
          bind:value={logSource.Server}
        />
      </div>
    </div>
    {#if logSource.Type == "scp"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>パス</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="パス"
            bind:value={logSource.Path}
          />
        </div>
      </div>
    {/if}
    <div class="form-group">
      <div class="form-group-header">
        <h5>アクセス設定</h5>
      </div>
      <div class="form-group-body">
        <p>
          <input
            class="form-control ml-2"
            type="text"
            placeholder="ユーザーID"
            bind:value={logSource.User}
          />
        </p>
        <p>
          <input
            class="form-control ml-2"
            type="password"
            placeholder="パスワード"
            bind:value={logSource.Password}
          />
        </p>
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>SSHキーファイル</h5>
      </div>
      <div class="form-group-body">
        <div class="input-group">
          <input
            class="form-control"
            type="text"
            placeholder="SSHキーファイル"
            bind:value={logSource.SSHKey}
          />
          <span class="input-group-button">
            <button class="btn" type="button" on:click={selectSSHKey}>
              <FileBadge16 />
            </button>
          </span>
        </div>
      </div>
    </div>
  {/if}
  {#if logSource.Type == "folder" || logSource.Type == "scp"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>ファイル名パターン</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="パターン"
          bind:value={logSource.Pattern}
        />
      </div>
    </div>
  {/if}
  {#if logSource.Type == "folder" || logSource.Type == "file" || logSource.Type == "scp"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>アーカイブ内ファイル名パターン</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="パターン"
          bind:value={logSource.InternalPattern}
        />
      </div>
    </div>
  {/if}
  {#if logSource.Type == "twsnmp" || logSource.Type == "gravwell" || logSource.Type == "windows"}
    <div class="form-group">
      <div class="form-group-header">
        <h5>サーバー</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control input-block"
          type="text"
          placeholder="サーバー"
          bind:value={logSource.Server}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>アクセス設定</h5>
      </div>
      <div class="form-group-body">
        <p>
          <input
            class="form-control ml-2"
            type="text"
            placeholder="ユーザーID"
            bind:value={logSource.User}
          />
        </p>
        <p>
          <input
            class="form-control ml-2"
            type="password"
            placeholder="パスワード"
            bind:value={logSource.Password}
          />
        </p>
        {#if logSource.Type == "windows"}
          <p>
            <select class="form-select" bind:value={logSource.Channel}>
              <option value="">デフォルト</option>
              <option value="Negotiate">Negotiate</option>
              <option value="NTLM">NTLM</option>
              <option value="Kerberos">Kerberos</option>
            </select>
          </p>
        {/if}
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>検索期間</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control input-sm"
          type="datetime-local"
          placeholder="開始"
          bind:value={logSource.Start}
        />
        -
        <input
          class="form-control input-sm"
          type="datetime-local"
          placeholder="終了"
          bind:value={logSource.End}
        />
      </div>
    </div>
    {#if logSource.Type == "twsnmp"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>ホスト名フィルター</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="ホスト名"
            bind:value={logSource.Host}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>タグフィルター</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="タグ"
            bind:value={logSource.Tag}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>メッセージフィルター</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="メッセージ"
            bind:value={logSource.Pattern}
          />
        </div>
      </div>
    {:else if logSource.Type == "windows"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>チャネル</h5>
        </div>
        <div class="form-group-body">
          <select class="form-select" bind:value={logSource.Channel}>
            <option value="System">システム</option>
            <option value="Security">セキュリティー</option>
            <option value="Application">アプリケーション</option>
          </select>
        </div>
      </div>
    {:else}
      <div class="form-group">
        <div class="form-group-header">
          <h5>検索文</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="検索文"
            bind:value={logSource.Pattern}
          />
        </div>
      </div>
    {/if}
  {/if}
</div>
<div class="Box-footer text-right">
  <button class="btn btn-secondary mr-1" type="button" on:click={cancel}>
    <X16 />
    キャンセル
  </button>
  {#if editMode}
    <button class="btn btn-danger mr-1" type="button" on:click={del}>
      <Trash16 />
      削除
    </button>
  {/if}
  <button class="btn btn-primary" type="button" on:click={save}>
    <Check16 />
    {#if editMode}
      更新
    {:else}
      追加
    {/if}
  </button>
</div>

<style>
  .form-group .form-control.input-block {
    width: 100%;
  }
</style>
