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
  import { _ } from '../../i18n/i18n';
  import {IsWindows,SelectFile,UpdateLogSource,DeleteLogSource} from '../../../wailsjs/go/main/App';
  export let logSource;

  let windows = false;

  IsWindows().then((r) => {
    windows = r;
  });

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let editMode = logSource && logSource.No > 0;

  const selectLogFolder = () => {
    SelectFile("logdir",$_('LogSorce.LogFolder'),).then((f) => {
      logSource.Path = f;
    });
  };

  const selectLogFile = () => {
    SelectFile("logfile",$_('LogSource.LogFile')).then((f) => {
      logSource.Path = f;
    });
  };

  const selectSSHKey = () => {
    SelectFile("sshkey",$_('LogSource.SSHKeyFile')).then((f) => {
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
    UpdateLogSource(logSource).then((e) => {
      errorMsg = e;
      if (e == "") {
        dispatch("done", { update: true });
      }
    });
  };

  const del = () => {
    DeleteLogSource(logSource.No,$_('LogSource.DeleteLogSource'),$_('LogSource.DeleteMsg')).then((e) => {
      errorMsg = e;
      if (e == "") {
        dispatch("done", { update: true });
      }
    });
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header">
    <h3 class="Box-title">{$_('LogSource.Title')}</h3>
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
        <h5>{$_('LogSource.Type')}</h5>
      </div>
      <div class="form-group-body">
        <select
          class="form-select"
          bind:value={logSource.Type}
          disabled={editMode}
        >
          <option value="folder">{$_('LogSource.Folder')}</option>
          <option value="file">{$_('LogSource.OneFile')}</option>
          <option value="scp">{$_('LogSource.SCP')}</option>
          {#if windows || logSource.Type == "windows"}
            <option value="windows">Windows</option>
          {/if}
          <option value="cmd">{$_('LogSource.Cmd')}</option>
          <option value="ssh">{$_('LogSource.SSH')}</option>
          <option value="twsnmp">{$_('LogSource.TWSNMPFC')}</option>
          <option value="gravwell">{$_('LogSource.Gravwell')}</option>
        </select>
      </div>
    </div>
    {#if logSource.Type == "folder"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.Folder')}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="{$_('LogSource.Folder')}"
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
          <h5>{$_('LogSource.OneFile')}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="{$_('LogSource.OneFile')}"
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
          <h5>{$_('LogSource.Comand')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="{$_('LogSource.Comand')}"
            bind:value={logSource.Path}
          />
        </div>
        <p class="note error" id="scppath-input-validation">
          {$_('LogSource.InputCmdMsg')}
        </p>
      </div>
    {/if}
    {#if logSource.Type == "scp" || logSource.Type == "ssh"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.Server')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="{$_('LogSource.Server')}"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      {#if logSource.Type == "scp"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.Path')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="{$_('LogSource.Path')}"
              bind:value={logSource.Path}
            />
          </div>
        </div>
      {/if}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.AccessSetting')}</h5>
        </div>
        <div class="form-group-body">
          <p>
            <input
              class="form-control ml-2"
              type="text"
              placeholder="{$_('LogSource.UserID')}"
              bind:value={logSource.User}
            />
          </p>
          <p>
            <input
              class="form-control ml-2"
              type="password"
              placeholder="{$_('LogSource.Password')}"
              bind:value={logSource.Password}
            />
          </p>
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.SSHKeyFile')}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="{$_('LogSource.SSHKeyFile')}"
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
          <h5>{$_('LogSource.FileNamePat')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder="{$_('LogSource.FileNamePat')}"
            bind:value={logSource.Pattern}
          />
        </div>
      </div>
    {/if}
    {#if logSource.Type == "folder" || logSource.Type == "file" || logSource.Type == "scp"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.FileNamePatInArc')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder="{$_('LogSource.Pat')}"
            bind:value={logSource.InternalPattern}
          />
        </div>
      </div>
    {/if}
    {#if logSource.Type == "twsnmp" || logSource.Type == "gravwell" || logSource.Type == "windows"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.Server')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="{$_('LogSource.Server')}"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_('LogSource.AccessSetting')}</h5>
        </div>
        <div class="form-group-body">
          <p>
            <input
              class="form-control ml-2"
              type="text"
              placeholder="{$_('LogSource.UserID')}"
              bind:value={logSource.User}
            />
          </p>
          <p>
            <input
              class="form-control ml-2"
              type="password"
              placeholder="{$_('LogSource.Password')}"
              bind:value={logSource.Password}
            />
          </p>
          {#if logSource.Type == "windows"}
            <p>
              <select class="form-select" bind:value={logSource.Channel}>
                <option value="">{$_('LogSource.Default')}</option>
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
          <h5>{$_('LogSource.TimeRange')}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder="{$_('LogSource.Start')}"
            bind:value={logSource.Start}
          />
          -
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder="{$_('LogSource.End')}"
            bind:value={logSource.End}
          />
        </div>
      </div>
      {#if logSource.Type == "twsnmp"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.HostNameFilter')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="{$_('LogSource.HostName')}"
              bind:value={logSource.Host}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.TagFilter')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="{$_('LogSource.Tag')}"
              bind:value={logSource.Tag}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.MsgFilter')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="{$_('LogSource.Message')}"
              bind:value={logSource.Pattern}
            />
          </div>
        </div>
      {:else if logSource.Type == "windows"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.Channel')}</h5>
          </div>
          <div class="form-group-body">
            <select class="form-select" bind:value={logSource.Channel}>
              <option value="System">{$_('LogSource.System')}</option>
              <option value="Security">{$_('LogSource.Security')}</option>
              <option value="Application">{$_('LogSource.Application')}</option>
            </select>
          </div>
        </div>
      {:else}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_('LogSource.SearchText')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder="{$_('LogSource.SearchText')}"
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
      {$_('LogSource.CancelBtn')}
    </button>
    {#if editMode}
      <button class="btn btn-danger mr-1" type="button" on:click={del}>
        <Trash16 />
        {$_('LogSource.DeleteBtn')}
      </button>
    {/if}
    <button class="btn btn-primary" type="button" on:click={save}>
      <Check16 />
      {#if editMode}
        {$_('LogSource.UpateBtn')}
      {:else}
        {$_('LogSource.AddBtn')}
      {/if}
    </button>
  </div>
</div>

<style>
  .form-group .form-control.input-block {
    width: 100%;
  }
</style>
