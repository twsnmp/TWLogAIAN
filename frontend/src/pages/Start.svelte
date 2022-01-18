<script>
  import { X16, FileDirectory16, Check16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();
  let workdir = "";
  let setWorkDirErr = false;
  let lastWorkDirs = [];
  let selLast = "";

  window.go.main.App.GetLastWorkDirs().then((wds) => {
    lastWorkDirs = wds;
  });
  const setWorkDir = () => {
    if (!workdir) {
      setWorkDirErr = true;
      return;
    }
    window.go.main.App.SetWorkDir(workdir).then((r) => {
      if (r) {
        dispatch("done", { page: "config" });
      } else {
        setWorkDirErr = true;
      }
    });
  };
  const selectWorkDir = () => {
    window.go.main.App.GetWorkDir().then((d) => {
      workdir = d;
    });
  };
  const cancel = () => {
    dispatch("done", { page: "wellcome" });
  };
  const checkSelectWorkDir = () => {
    if (selLast != "") {
      workdir = selLast;
    }
  };
  const clearMsg = () => {
    setWorkDirErr = false;
  };
</script>

<div class="Box mx-auto" style="max-width: 500px;">
  <div class="Box-header">
    <h3 class="Box-title">作業フォルダーの選択</h3>
  </div>
  {#if setWorkDirErr}
    <div class="flash flash-error">
      選択した作業フォルダーは利用できません
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
      <div class="input-group mb-5">
        <input
          class="form-control errored"
          type="text"
          placeholder="作業フォルダー"
          aria-label="作業フォルダー"
          value={workdir}
        />
        <span class="input-group-button">
          <button class="btn" type="button" on:click={selectWorkDir}>
            <FileDirectory16 />
          </button>
        </span>
      </div>
      {#if lastWorkDirs.length > 0}
        <p>最近使ったフォルダー</p>
        <!-- svelte-ignore a11y-no-onchange -->
        <select
          class="form-select"
          aria-label="最近使ったフォルダー"
          bind:value={selLast}
          on:change={checkSelectWorkDir}
        >
          <option value="">選択してください。</option>
          {#each lastWorkDirs as d}
            <option value={d}>{d}</option>
          {/each}
        </select>
      {/if}
    </form>
  </div>
  <div class="Box-footer text-right">
    <button class="btn" type="button btn-secondary mr-1" on:click={cancel}>
      <X16 />
      戻る
    </button>
    <button class="btn btn-primary" type="button" on:click={setWorkDir}>
      <Check16 />
      次へ
    </button>
  </div>
</div>
