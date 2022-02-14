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
  let pathError = false;

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
    if (logSource.Path == "") {
      pathError = true;
      return
    }
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
      aria-label="Close"
      on:click={clearMsg}
    >
      <X16 />
    </button>
  </div>
{/if}
<div class="Box-body">
  <form>
    <div class="form-group">
      <div class="form-group-header">
        <h5>タイプ</h5>
      </div>
      <div class="form-group-body">
        <select
          class="form-select"
          aria-label="ログソース"
          bind:value={logSource.Type}
          disabled={editMode}
        >
          <option value="folder">フォルダー</option>
          <option value="file">単一ファイル</option>
          <option value="scp">SCP</option>
        </select>
      </div>
    </div>
    {#if logSource.Type == "folder"}
      <div class="form-group" class:errored={pathError}>
        <div class="form-group-header">
          <h5>フォルダー</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="フォルダー"
              aria-label="フォルダー"
              bind:value={logSource.Path}
              aria-describedby="path-input-validation"
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectLogFolder}>
                <FileDirectory16 />
              </button>
            </span>
          </div>
          <p class="note error" id="path-input-validation">
            フォルダを選択してください
          </p>
        </div>
      </div>
    {/if}
    {#if logSource.Type == "file"}
      <div class="form-group class:errored={pathError}">
        <div class="form-group-header">
          <h5>単一ファイル</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="ファイル"
              aria-label="ファイル"
              bind:value={logSource.Path}
              aria-describedby="file-input-validation"
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectLogFile}>
                <File16 />
              </button>
            </span>
          </div>
          <p class="note error" id="file-input-validation">
            ファイルを選択してください。
          </p>
        </div>
      </div>
    {/if}
    {#if logSource.Type == "scp" }
      <div class="form-group" class:errored={pathError}>
        <div class="form-group-header">
          <h5>サーバー</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="サーバー"
            aria-label="サーバー"
            bind:value={logSource.Server}
            aria-describedby="server-input-validation"
          />
        </div>
        <p class="note error" id="server-input-validation">
           サーバーが空欄か形式が正しくありません
        </p>
      </div>
      <div class="form-group" class:errored={pathError}>
        <div class="form-group-header">
          <h5>パス</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="パス"
            aria-label="パス"
            bind:value={logSource.Path}
            aria-describedby="scppath-input-validation"
          />
        </div>
        <p class="note error" id="scppath-input-validation">
           パスを指定してください
        </p>
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
            aria-label="ユーザーID"
            bind:value={logSource.User}
            />
          </p>
          <p>
            <input
            class="form-control ml-2"
            type="password"
            placeholder="パスワード"
            aria-label="パスワード"
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
              aria-label="SSHキーファイル"
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
            aria-label="パターン"
            bind:value={logSource.Pattern}
          />
        </div>
      </div>
    {/if}
    <div class="form-group">
      <div class="form-group-header">
        <h5>アーカイブ内ファイル名パターン</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="パターン"
          aria-label="パターン"
          bind:value={logSource.InternalPattern}
        />
      </div>
    </div>
</form>
</div>
<div class="Box-footer text-right">
  <button class="btn btn-secondary" type="button" on:click={cancel}>
    <X16 />
    キャンセル
  </button>
  {#if editMode}
    <button class="btn btn-danger ml-2" type="button" on:click={del}>
      <Trash16 />
      削除
    </button>
  {/if}
  <button class="btn btn-primary ml-2" type="button" on:click={save}>
    <Check16 />
    {#if editMode }
      更新
    {:else}
      追加
    {/if}
  </button>
</div>

<style>
  .form-group .form-control.input-block {
    width: 100%
  }
</style>