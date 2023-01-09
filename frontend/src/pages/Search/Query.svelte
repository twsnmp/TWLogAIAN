<script>
  import { getFieldName, getFieldType } from "../../js/define";
  import { Plus16, Trash16, File16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { _ } from '../../i18n/i18n';
  import {LoadKeyword} from '../../../wailsjs/go/main/App';


  export let conf;
  export let fields = [];

  let hasStringField = false;
  let hasNumberField = false;

  let history = "";
  const dispatch = createEventDispatcher();

  const setHistory = () => {
    if (history) {
      dispatch("update", { query: history, add: false });
    }
  };

  const addKeyword = () => {
    if (!conf.keyword.key) {
      return;
    }
    const q =
      conf.keyword.field == ""
        ? " " + conf.keyword.mode + conf.keyword.key
        : " " + conf.keyword.mode + conf.keyword.field + ":" + conf.keyword.key;
    dispatch("update", { query: q, add: true });
  };

  const addNumber = () => {
    if (!conf.number.field || !conf.number.oper || !conf.number.value) {
      return;
    }
    const oper = conf.number.oper == "=" ? "=" : ":" + conf.number.oper;
    const q = " " + conf.number.field + oper + conf.number.value;
    dispatch("update", { query: q, add: true });
  };


  fields.forEach((f) => {
    if (f.startsWith("_")) {
      return;
    }
    if (getFieldType(f) == "string") {
      hasStringField = true;
    }
    if (getFieldType(f) == "number") {
      hasNumberField = true;
      if (conf.number.field == "") {
        conf.number.field = f;
      }
    }
  });

  const clear = () => {
    conf.history.length = 0;
  };

  const loadKeyword = () => {
    LoadKeyword().then((r) => {
      if (r) {
        r.forEach((k) => {
          const q =
            conf.keyword.field == ""
              ? " " + conf.keyword.mode + k
              : " " + conf.keyword.mode + conf.keyword.field + ":" + k;
          dispatch("update", { query: q, add: true });
        });
      }
    });
  };
</script>

{#if conf.history.length > 0}
  <div class="container-lg clearfix mb-1">
    <div class="col-2 float-left">{$_('Query.History')}</div>
    <div class="col-8 float-left">
      <!-- svelte-ignore a11y-no-onchange -->
      <select
        style="width: 80%;"
        class="form-select"
        aria-label="{$_('Query.History')}"
        bind:value={history}
        on:change={setHistory}
      >
        <option value="">{$_('Query.SelectHistoryMsg')}</option>
        {#each conf.history as h}
          <option value={h}>{h}</option>
        {/each}
      </select>
    </div>
    <div class="col-2 float-left" />
    <button class="btn btn-danger" type="button" on:click={clear}>
      <Trash16 />
    </button>
  </div>
{/if}

{#if hasStringField}
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">{$_('Query.Keyword')}</div>
    <div class="col-8 float-left">
      <select
        class="form-select"
        aria-label="{$_('Query.Item')}"
        bind:value={conf.keyword.field}
      >
        <option value="">{$_('Query.All')}</option>
        {#each fields as f}
          {#if !f.startsWith("_") && getFieldType(f) == "string"}
            <option value={f}>{getFieldName(f)}</option>
          {/if}
        {/each}
      </select>
      {$_('Query.Include')}
      <input
        class="form-control input-sm"
        type="text"
        style="width: 150px;"
        placeholder="{$_('Query.Keyword')}"
        aria-label="{$_('Query.Keyword')}"
        bind:value={conf.keyword.key}
      />
      {$_('Query.Is')}
      <select class="form-select" bind:value={conf.keyword.mode}>
        <option value="+">{$_('Query.Must')}</option>
        <option value="-">{$_('Query.Exclude')}</option>
      </select>
    </div>
    <div class="col-2 float-left">
      <button class="btn" type="button" on:click={addKeyword}>
        <Plus16 />
      </button>
      <button class="btn" type="button" on:click={loadKeyword}>
        <File16 />
      </button>
    </div>
  </div>
{/if}
{#if hasNumberField}
  <div class="container-lg clearfix mt-1">
    <div class="col-2 float-left">{$_('Query.NumberCond')}</div>
    <div class="col-8 float-left">
      <select
        class="form-select"
        aria-label="{$_('Query.Item')}"
        bind:value={conf.number.field}
      >
        {#each fields as f}
          {#if !f.startsWith("_") && getFieldType(f) == "number"}
            <option value={f}>{getFieldName(f)}</option>
          {/if}
        {/each}
      </select>
      <select class="form-select" bind:value={conf.number.oper}>
        <option value="<">&lt;</option>
        <option value=">">&gt;</option>
        <option value="=">=</option>
        <option value="<=">&lt;=</option>
        <option value=">=">&gt;=</option>
      </select>
      <input
        class="form-control input-sm"
        type="text"
        style="width: 100px;"
        placeholder="$_('Query.Number')"
        aria-label="{$_('Query.Number')}"
        bind:value={conf.number.value}
      />
    </div>
    <div class="col-2 float-left">
      <button class="btn" type="button" on:click={addNumber}>
        <Plus16 />
      </button>
    </div>
  </div>
{/if}

<style>
  select {
    min-width: 100px;
  }
</style>
