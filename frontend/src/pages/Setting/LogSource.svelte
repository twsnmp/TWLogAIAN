<script>
  import {
    X16,
    File16,
    FileDirectory16,
    Check16,
    Trash16,
    FileBadge16,
  } from "svelte-octicons";
  import { createEventDispatcher, onMount } from "svelte";
  import { _ } from "../../i18n/i18n";
  import {
    IsWindows,
    SelectFile,
    UpdateLogSource,
    DeleteLogSource,
  } from "../../../wailsjs/go/main/App";
  export let logSource;

  let windows = false;

  onMount(async () => {
    windows = await IsWindows();
    if (!logSource.Type) {
      logSource.Type = "folder";
    }
  });

  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let editMode = logSource && logSource.No > 0;

  const selectLogFolder = async () => {
    logSource.Path = (await SelectFile("logdir", $_("LogSource.LogFolder"))) || "";
  };

  const selectLogFile = async () => {
    logSource.Path = (await SelectFile("logfile", $_("LogSource.LogFile"))) || "";
  };

  const selectSSHKey = async () => {
    logSource.SSHKey = await SelectFile("sshkey", $_("LogSource.SSHKeyFile"));
  };

  const selectCACert = async () => {
    logSource.CACert = await SelectFile("cacert", $_("LogSource.CACert"));
  };

  const selectClientCert = async () => {
    logSource.ClientCert = await SelectFile("cert", $_("LogSource.ClientCert"));
  };

  const selectClientKey = async () => {
    logSource.ClientKey = await SelectFile("key", $_("LogSource.ClientKey"));
  };

  const cancel = () => {
    dispatch("done", { update: false });
  };

  const clearMsg = () => {
    errorMsg = "";
  };

  const save = async () => {
    const e = await UpdateLogSource(logSource);
    errorMsg = e || "";
    if (e == "") {
      dispatch("done", { update: true });
    }
  };

  const del = async () => {
    const e = await DeleteLogSource(
      logSource.No,
      $_("LogSource.DeleteLogSource"),
      $_("LogSource.DeleteMsg")
    );
    errorMsg = e || "";
    if (e == "") {
      dispatch("done", { update: true });
    }
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header">
    <h3 class="Box-title">{$_("LogSource.Title")}</h3>
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
        <h5>{$_("LogSource.Type")}</h5>
      </div>
      <div class="form-group-body">
        <select
          class="form-select"
          bind:value={logSource.Type}
          disabled={editMode}
        >
          <option value="folder">{$_("LogSource.Folder")}</option>
          <option value="file">{$_("LogSource.OneFile")}</option>
          <option value="parquet">{$_("LogSource.Parquet")}</option>
          <option value="scp">{$_("LogSource.SCP")}</option>
          <option value="ftp">{$_("LogSource.FTP")}</option>
          <option value="loki">{$_("LogSource.Loki")}</option>
          <option value="es">{$_("LogSource.ES")}</option>
          <option value="imap">{$_("LogSource.IMAP")}</option>
          <option value="pop3">{$_("LogSource.POP3")}</option>
          <option value="twlogeye">{$_("LogSource.TWLogEye")}</option>
          <option value="twsnmp">{$_("LogSource.TWSNMPFC")}</option>
          {#if windows || logSource.Type == "windows"}
            <option value="windows">Windows Event Log</option>
          {/if}
          <option value="cmd">{$_("LogSource.Cmd")}</option>
          <option value="ssh">{$_("LogSource.SSH")}</option>
        </select>
      </div>
    </div>

    <!-- Folder -->
    {#if logSource.Type == "folder"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Folder")}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder={$_("LogSource.Folder")}
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
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.FileNamePat")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.FileNamePat")}
            bind:value={logSource.Pattern}
          />
        </div>
      </div>
    {/if}

    <!-- Single File -->
    {#if logSource.Type == "file"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.OneFile")}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder={$_("LogSource.OneFile")}
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

    <!-- Parquet -->
    {#if logSource.Type == "parquet"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Parquet")}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder={$_("LogSource.Parquet")}
              bind:value={logSource.Path}
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectLogFile}>
                <File16 />
              </button>
            </span>
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectLogFolder}>
                <FileDirectory16 />
              </button>
            </span>
          </div>
        </div>
      </div>
    {/if}

    <!-- Command / SSH Command -->
    {#if logSource.Type == "cmd" || logSource.Type == "ssh"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Comand")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder={$_("LogSource.Comand")}
            bind:value={logSource.Path}
          />
        </div>
        <p class="note error" id="scppath-input-validation">
          {$_("LogSource.InputCmdMsg")}
        </p>
      </div>
    {/if}

    <!-- SCP / SSH / FTP Server & Auth -->
    {#if logSource.Type == "scp" || logSource.Type == "ssh" || logSource.Type == "ftp"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="host:port (e.g. 192.168.1.1:21)"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Path")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="Remote directory or file path"
            bind:value={logSource.Path}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.AccessSetting")}</h5>
        </div>
        <div class="form-group-body d-flex" style="gap: 8px;">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.UserID")}
            bind:value={logSource.User}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Password")}
            bind:value={logSource.Password}
            style="flex: 1;"
          />
        </div>
      </div>
      {#if logSource.Type == "scp" || logSource.Type == "ssh"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.SSHKeyFile")}</h5>
          </div>
          <div class="form-group-body">
            <div class="input-group">
              <input
                class="form-control"
                type="text"
                placeholder={$_("LogSource.SSHKeyFile")}
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
      {#if logSource.Type == "ftp"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.FileNamePat")}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control"
              type="text"
              placeholder="*.log, *.gz"
              bind:value={logSource.Pattern}
            />
          </div>
        </div>
        <div class="form-group d-flex" style="gap: 16px;">
          <label style="cursor: pointer;">
            <input type="checkbox" bind:checked={logSource.TLS} />
            {$_("LogSource.TLS")} (Explicit TLS / FTPS)
          </label>
          <label style="cursor: pointer;">
            <input type="checkbox" bind:checked={logSource.InsecureSkip} />
            {$_("LogSource.InsecureSkip")}
          </label>
        </div>
      {/if}
    {/if}

    <!-- Grafana Loki -->
    {#if logSource.Type == "loki"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="http://localhost:3100"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Query")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder='LogQL query, e.g. &#123;job="syslog"&#125;'
            bind:value={logSource.Query}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.AccessSetting")}</h5>
        </div>
        <div class="form-group-body d-flex" style="gap: 8px;">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.UserID")}
            bind:value={logSource.User}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Password")}
            bind:value={logSource.Password}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Token")}
            bind:value={logSource.Token}
            style="flex: 1;"
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.OrgID")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control"
            type="text"
            placeholder="Tenant ID (X-Scope-OrgID)"
            bind:value={logSource.OrgID}
          />
        </div>
      </div>
      <div class="form-group">
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.InsecureSkip} />
          {$_("LogSource.InsecureSkip")}
        </label>
      </div>
    {/if}

    <!-- Elasticsearch / OpenSearch -->
    {#if logSource.Type == "es"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="http://localhost:9200"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group d-flex" style="gap: 8px;">
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("LogSource.Index")}</h5>
          </div>
          <input
            class="form-control width-full"
            type="text"
            placeholder="e.g. logs-*"
            bind:value={logSource.Index}
          />
        </div>
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("Setting.TimeField")}</h5>
          </div>
          <input
            class="form-control width-full"
            type="text"
            placeholder="@timestamp"
            bind:value={logSource.TimeField}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Query")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="Lucene query or JSON DSL"
            bind:value={logSource.Query}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.AccessSetting")}</h5>
        </div>
        <div class="form-group-body d-flex" style="gap: 8px;">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.UserID")}
            bind:value={logSource.User}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Password")}
            bind:value={logSource.Password}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.APIKey")}
            bind:value={logSource.APIKey}
            style="flex: 1;"
          />
        </div>
      </div>
      <div class="form-group">
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.InsecureSkip} />
          {$_("LogSource.InsecureSkip")}
        </label>
      </div>
    {/if}

    <!-- Email IMAP / POP3 -->
    {#if logSource.Type == "imap" || logSource.Type == "pop3"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder={logSource.Type == "imap"
              ? "imap.example.com:993"
              : "pop.example.com:995"}
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.AccessSetting")}</h5>
        </div>
        <div class="form-group-body d-flex" style="gap: 8px;">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.UserID")}
            bind:value={logSource.User}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Password")}
            bind:value={logSource.Password}
            style="flex: 1;"
          />
        </div>
      </div>
      {#if logSource.Type == "imap"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.Folder")}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control"
              type="text"
              placeholder="INBOX"
              bind:value={logSource.Folder}
            />
          </div>
        </div>
      {/if}
      <div class="form-group d-flex" style="gap: 16px;">
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.TLS} />
          {$_("LogSource.TLS")}
        </label>
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.InsecureSkip} />
          {$_("LogSource.InsecureSkip")}
        </label>
      </div>
    {/if}

    <!-- TWLogEye -->
    {#if logSource.Type == "twlogeye"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder="localhost:8081"
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group d-flex" style="gap: 8px;">
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("LogSource.Target")}</h5>
          </div>
          <select class="form-select width-full" bind:value={logSource.Target}>
            <option value="notify">notify (通知)</option>
            <option value="logs">logs (ログ)</option>
            <option value="report">report (レポート)</option>
          </select>
        </div>
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("LogSource.SubTarget")}</h5>
          </div>
          <select
            class="form-select width-full"
            bind:value={logSource.SubTarget}
          >
            {#if logSource.Target == "logs"}
              <option value="syslog">syslog</option>
              <option value="trap">trap</option>
              <option value="netflow">netflow</option>
              <option value="ipfix">ipfix</option>
              <option value="sflow">sflow</option>
              <option value="winevent">winevent</option>
              <option value="otel">otel</option>
              <option value="mqtt">mqtt</option>
              <option value="event">event</option>
            {:else if logSource.Target == "report"}
              <option value="syslog">syslog</option>
              <option value="trap">trap</option>
              <option value="netflow">netflow</option>
              <option value="winevent">winevent</option>
              <option value="otel">otel</option>
              <option value="mqtt">mqtt</option>
              <option value="monitor">monitor</option>
              <option value="anomaly">anomaly</option>
            {:else}
              <option value="">(All)</option>
            {/if}
          </select>
        </div>
      </div>
      {#if logSource.Target == "notify"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.Level")}</h5>
          </div>
          <select class="form-select" bind:value={logSource.Level}>
            <option value="">All</option>
            <option value="info">Info</option>
            <option value="warn">Warn</option>
            <option value="error">Error</option>
          </select>
        </div>
      {/if}
      {#if logSource.Target == "report" && logSource.SubTarget == "anomaly"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.ReportType")}</h5>
          </div>
          <select class="form-select" bind:value={logSource.ReportType}>
            <option value="monitor">monitor</option>
            <option value="syslog">syslog</option>
            <option value="trap">trap</option>
            <option value="netflow">netflow</option>
            <option value="winevent">winevent</option>
          </select>
        </div>
      {/if}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.CACert")}</h5>
        </div>
        <div class="form-group-body">
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="ca.crt"
              bind:value={logSource.CACert}
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectCACert}>
                <File16 />
              </button>
            </span>
          </div>
        </div>
      </div>
      <div class="form-group d-flex" style="gap: 8px;">
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("LogSource.ClientCert")}</h5>
          </div>
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="client.crt"
              bind:value={logSource.ClientCert}
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectClientCert}>
                <File16 />
              </button>
            </span>
          </div>
        </div>
        <div style="flex: 1;">
          <div class="form-group-header">
            <h5>{$_("LogSource.ClientKey")}</h5>
          </div>
          <div class="input-group">
            <input
              class="form-control"
              type="text"
              placeholder="client.key"
              bind:value={logSource.ClientKey}
            />
            <span class="input-group-button">
              <button class="btn" type="button" on:click={selectClientKey}>
                <File16 />
              </button>
            </span>
          </div>
        </div>
      </div>
      <div class="form-group d-flex" style="gap: 16px;">
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.TLS} />
          {$_("LogSource.TLS")}
        </label>
        <label style="cursor: pointer;">
          <input type="checkbox" bind:checked={logSource.InsecureSkip} />
          {$_("LogSource.InsecureSkip")}
        </label>
      </div>
    {/if}

    <!-- TWSNMP FC & Windows Settings -->
    {#if logSource.Type == "twsnmp" || logSource.Type == "windows"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.Server")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-block"
            type="text"
            placeholder={$_("LogSource.Server")}
            bind:value={logSource.Server}
          />
        </div>
      </div>
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.AccessSetting")}</h5>
        </div>
        <div class="form-group-body d-flex" style="gap: 8px;">
          <input
            class="form-control"
            type="text"
            placeholder={$_("LogSource.UserID")}
            bind:value={logSource.User}
            style="flex: 1;"
          />
          <input
            class="form-control"
            type="password"
            placeholder={$_("LogSource.Password")}
            bind:value={logSource.Password}
            style="flex: 1;"
          />
          {#if logSource.Type == "windows"}
            <select class="form-select" bind:value={logSource.Auth}>
              <option value="">{$_("LogSource.Default")}</option>
              <option value="Negotiate">Negotiate</option>
              <option value="NTLM">NTLM</option>
              <option value="Kerberos">Kerberos</option>
            </select>
          {/if}
        </div>
      </div>
      {#if logSource.Type == "twsnmp"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.SubTarget")}</h5>
          </div>
          <select
            class="form-select"
            bind:value={logSource.SubTarget}
          >
            <option value="syslog">Syslog</option>
            <option value="eventlog">EventLog</option>
            <option value="trap">SNMP Trap</option>
            <option value="netflow">NetFlow</option>
            <option value="ipfix">IPFIX</option>
            <option value="sflow">sFlow</option>
            <option value="arp">ARP Log</option>
          </select>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.HostNameFilter")}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder={$_("LogSource.HostName")}
              bind:value={logSource.Host}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.TagFilter")}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder={$_("LogSource.Tag")}
              bind:value={logSource.Tag}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.MsgFilter")}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              placeholder={$_("LogSource.Message")}
              bind:value={logSource.Pattern}
            />
          </div>
        </div>
      {:else if logSource.Type == "windows"}
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.Channel")}</h5>
          </div>
          <div class="form-group-body">
            <select class="form-select" bind:value={logSource.Channel}>
              <option value="System">{$_("LogSource.System")}</option>
              <option value="Security">{$_("LogSource.Security")}</option>
              <option value="Application">{$_("LogSource.Application")}</option>
            </select>
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5>{$_("LogSource.ShiftJISToUTF8")}</h5>
          </div>
          <div class="form-group-body">
            <label class="p-1">
              <input type="checkbox" bind:checked={logSource.ShiftJIS} />
              {$_("LogSource.Convert")}
            </label>
          </div>
        </div>
      {/if}
    {/if}

    <!-- Time Range (For remote queries: Loki, ES, TWSNMP, TWLogEye, Windows) -->
    {#if logSource.Type == "twsnmp" || logSource.Type == "windows" || logSource.Type == "loki" || logSource.Type == "es" || logSource.Type == "twlogeye"}
      <div class="form-group">
        <div class="form-group-header">
          <h5>{$_("LogSource.TimeRange")}</h5>
        </div>
        <div class="form-group-body">
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder={$_("LogSource.Start")}
            bind:value={logSource.Start}
          />
          -
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder={$_("LogSource.End")}
            bind:value={logSource.End}
          />
        </div>
      </div>
    {/if}
  </div>
  <div class="Box-footer d-flex flex-justify-between">
    <div>
      <button class="btn btn-secondary mr-1" type="button" on:click={cancel}>
        <X16 />
        {$_("LogSource.CancelBtn")}
      </button>
      {#if editMode}
        <button class="btn btn-danger" type="button" on:click={del}>
          <Trash16 />
          {$_("LogSource.DeleteBtn")}
        </button>
      {/if}
    </div>
    <div>
      <button class="btn btn-primary" type="button" on:click={save}>
        <Check16 />
        {#if editMode}
          {$_("LogSource.UpateBtn")}
        {:else}
          {$_("LogSource.AddBtn")}
        {/if}
      </button>
    </div>
  </div>
</div>

<style>
  .form-group .form-control.input-block {
    width: 100%;
  }
</style>
