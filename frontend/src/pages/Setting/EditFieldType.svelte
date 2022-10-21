<script>
  import { X16, Check16, StarFill16, Reply16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  export let fieldType;
  export let add = true;

  const dispatch = createEventDispatcher();
  let errorMsg = "";

  const back = () => {
    dispatch("done", {});
  };

  const save = () => {
    window.go.main.App.SaveFieldType(fieldType).then((r) => {
      errorMsg = r;
      if (r == "") {
        dispatch("done", {save:true});
      }
    });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header">
    <h3 class="Box-title">フィールドタイプ編集</h3>
  </div>
  {#if errorMsg}
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
    <div class="form-group">
      <div class="form-group-header">
        <h5>キー（変数名）</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          disabled={!add}
          placeholder="フィールドタイプのキー"
          bind:value={fieldType.Key}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>
          名前
        </h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="フィールドタイプの名前"
          bind:value={fieldType.Name}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>
          単位
        </h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="フィールドタイプの単位"
          bind:value={fieldType.Unit}
        />
      </div>
    </div>
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-secondary mr-1" type="button" on:click={back}>
      <X16 />
      キャンセル
    </button>
    <button class="btn btn-primary mr-1" type="button" on:click={save}>
      <X16 />
      保存
    </button>
  </div>
</div>
