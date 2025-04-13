<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { _ } from '../../i18n/i18n';
  import {SaveFieldType} from '../../../wailsjs/go/main/App';

  export let fieldType;
  export let add = true;

  const dispatch = createEventDispatcher();
  let errorMsg = "";

  const back = () => {
    dispatch("done", {});
  };

  const save = async () => {
    const r = await SaveFieldType(fieldType);
    errorMsg = r || "";
    if (r == "") {
      dispatch("done", {save:true});
    }
  };

  const clearMsg = () => {
    errorMsg = "";
  };

</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header">
    <h3 class="Box-title">{$_('EditFieldType.Title')}</h3>
  </div>
  {#if errorMsg}
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
        <h5>{$_('EditFieldType.Key')}</h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          disabled={!add}
          placeholder="{$_('EditFieldType.KeyInput')}"
          bind:value={fieldType.Key}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>
          {$_('EditFieldType.Name')}
        </h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="{$_('EditFieldType.NameOfFieldType')}"
          bind:value={fieldType.Name}
        />
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>
          {$_('EditFieldType.TypeOfData')}
        </h5>
      </div>
      <div class="form-group-body">
        <select
        class="form-select"
        bind:value={fieldType.Type}
      >
        <option value="string">{$_('EditFieldType.String')}</option>
        <option value="number">{$_('EditFieldType.Number')}</option>
      </select>
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>
          {$_('EditFieldType.Unit')}
        </h5>
      </div>
      <div class="form-group-body">
        <input
          class="form-control"
          type="text"
          placeholder="{$_('EditFieldType.UnitOfFieldType')}"
          bind:value={fieldType.Unit}
        />
      </div>
    </div>
  </div>
  <div class="Box-footer text-right">
    <button class="btn btn-secondary mr-1" type="button" on:click={back}>
      <X16 />
      {$_('EditFieldType.CancelBtn')}
    </button>
    <button class="btn btn-primary mr-1" type="button" on:click={save}>
      <X16 />
      {$_('EditFieldType.SaveBtn')}
    </button>
  </div>
</div>
