<script>
  import {
    X16,
    Plus16,
    Check16,
    File16,
    Checklist16,
    Trash16,
    Search16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import LogSource from "./LogSource.svelte";
  import LogType from "./LogType.svelte";
  import { onMount } from "svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { loadFieldTypes } from "../../js/define";
  import { _,getLocale } from '../../i18n/i18n';
  import {
    GetConfig,
    HasIndex,
    GetLogSources,
    GetExtractorTypes,
    DeleteLogSource,
    Start,
    ClearIndex,
    SelectFile,
    CloseWorkDir,
  } from '../../../wailsjs/go/main/App';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  const dispatch = createEventDispatcher();
  const data = [];
  let config = {
    Filter: "",
    Extractor: "auto",
    Grok: "",
    TimeField: "",
    GeoIP: false,
    GeoIPDB: "",
    GeoFields: "",
    HostName: false,
    HostFields: "",
    VendorName: false,
    MACFields: "",
    Recursive: false,
    InMemory: false,
    SampleLog: "",
    ForceUTC: false,
    Strict: false,
  };
  let logSource = {
    No: 0,
    Type: "folder",
    Path: "",
    Pattern: "",
    InternalPattern: "",
    User: "",
    Password: "",
    Server: "",
    KeyPath: "",
    Start: "",
    End: "",
    Tag: "",
    Host: "",
    Channel: "",
    Auth: "",
  };
  let logSources = [];
  let errorMsg = "";
  let infoMsg = "";
  let page = "";
  let orgConfig;
  let hasIndex = false;

  const getConfig = () => {
    GetConfig().then((c) => {
      if (!c.Extractor) {
        c.Extractor = "auto"
      }
      config = c;
      orgConfig = c;
    });
  };

  const getHasIndex = () => {
    HasIndex().then((r) => {
      hasIndex = r;
    });
  };

  const getLSPath = (e) => {
    switch (e.Type) {
      case "ssh":
      case "scp":
        return e.Server + ":" + e.Path;
      case "twsnmp":
        return (
          e.Server +
          "/?start=" +
          e.Start +
          "&end=" +
          e.End +
          "&host=" +
          e.Host +
          "&tag=" +
          e.Tag +
          "&message=" +
          e.Pattern
        );
      case "windows":
        return (
          e.Server +
          "/?start=" +
          e.Start +
          "&end=" +
          e.End +
          "&channel=`" +
          e.Channel +
          "`"
        );
    }
    return e.Path;
  };

  let pagination = false;
  const getLogSources = () => {
    GetLogSources().then((ds) => {
      data.length = 0;
      if (ds) {
        logSources = ds;
        logSources.forEach((e) => {
          const path = getLSPath(e);
          data.push([e.No, e.Type, path, ""]);
        });
        if (ds.length > 5) {
          pagination = {
            limit: 5,
            enable: true,
          };
        } else {
          pagination = false;
        }
      } else {
        logSources = [];
      }
    });
  };

  let extractorTypes = {};
  let extractorTypeList = [];

  const getExtractorTypes = () => {
    GetExtractorTypes().then((r) => {
      if (r) {
        extractorTypes = r;
        extractorTypeList = [];
        for (let k in extractorTypes) {
          extractorTypeList.push(extractorTypes[k]);
        }
        extractorTypeList.sort((a, b) => a.Name > b.Name);
      }
    });
  };

  onMount(() => {
    loadFieldTypes();
    getConfig();
    getLogSources();
    getExtractorTypes();
    getHasIndex();
  });

  const editLogSource = (sno) => {
    const no = sno * 1;
    if (sno == "" || no < 0 || no > logSources.length) {
      logSource = {
        No: 0,
        Type: "folder",
        Path: "",
        Pattern: "",
        Server: "",
        User: "",
        Password: "",
        KeyPath: "",
        Start: "",
        End: "",
        Host: "",
        Tag: "",
        Channel: "",
        Auth: "",
      };
    } else {
      logSource = logSources[no - 1];
    }
    page = "logSource";
  };

  const deleteLogSource = (sno) => {
    const no = sno * 1;
    DeleteLogSource(no,$_('Setting.DeleteLogSourceTitle'),$_('Setting.DeleteMsg')).then((e) => {
      if (e == "No") {
        return;
      }
      errorMsg = e;
      if (e == "") {
        getLogSources();
      }
    });
  };

  const formatLogSourceType = (t) => {
    switch (t) {
      case "folder":
        return $_('Setting.Folder');
      case "file":
        return $_('Setting.OneFile');
      case "scp":
        return $_('Setting.SCP');
      case "windows":
        return "Windows";
      case "cmd":
        return $_('Setting.Command');
      case "ssh":
        return $_('Setting.SSH');
      case "twsnmp":
        return $_('Setting.TSNMP');
    }
    return "";
  };

  const editLogSourceButtons = (_, row) => {
    const no = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => editLogSource(no),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>`
      )
    );
  };

  const deleteLogSourceButtons = (_, row) => {
    const no = row.cells[0].data;
    return h(
      "button",
      {
        className: "btn btn-sm btn-danger",
        onClick: () => deleteLogSource(no),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.576l.66-6.6a.75.75 0 00-1.492-.149l-.66 6.6a.25.25 0 01-.249.225h-5.19a.25.25 0 01-.249-.225l-.66-6.6z"></path></svg>`
      )
    );
  };

  const columns = [
    {
      name: "No",
      sort: true,
      width: "10%",
    },
    {
      name: $_('Setting.Type'),
      sort: true,
      width: "20%",
      formatter: formatLogSourceType,
    },
    {
      name: $_('Setting.Path'),
      sort: true,
      width: "50%",
    },
    {
      name: $_('Setting.Edit'),
      sort: false,
      width: "10%",
      formatter: editLogSourceButtons,
    },
    {
      name: $_('Setting.Delete'),
      sort: false,
      width: "10%",
      formatter: deleteLogSourceButtons,
    },
  ];
  let busy = false;
  const start = () => {
    busy = true;
    Start(config, false).then((e) => {
      busy = false;
      if (e && e != "") {
        errorMsg = e;
      } else {
        dispatch("done", { page: "processing" });
      }
    });
  };

  const clear = () => {
    ClearIndex($_('Setting.ClearIndexTitle'),$_('Setting.DeleteMsg')).then((e) => {
      if (e && e != "") {
        if (e == "No") {
          return;
        }
        errorMsg = e;
      } else {
        hasIndex = false;
      }
    });
  };

  const search = () => {
    busy = true;
    Start(config, true).then((e) => {
      busy = false;
      if (e && e != "") {
        errorMsg = e;
      } else {
        dispatch("done", { page: "logview" });
      }
    });
  };

  const selectGeoIPDB = () => {
    SelectFile("geoip",$_('Setting.IPGeoDB')).then((f) => {
      config.GeoIPDB = f;
    });
  };

  const cancel = () => {
    CloseWorkDir($_('Setting.StopTitle'),$_('Setting.CloseMsg')).then((r) => {
      if (r == "") {
        dispatch("done", { page: "wellcome" });
      }
    });
  };

  const clearMsg = () => {
    errorMsg = "";
    infoMsg = "";
  };

  const changeExtractor = () => {
    const e = extractorTypes[config.Extractor];
    if (e) {
      config.Grok = e.Grok;
      config.TimeField = e.TimeField;
      config.GeoFields = e.IPFields;
      config.HostFields = e.IPFields;
      config.MACFields = e.MACFields;
    } else if (orgConfig) {
      config.Grok = orgConfig.Grok;
      config.TimeField = orgConfig.TimeField;
      config.GeoFields = orgConfig.IPFields;
      config.HostFields = orgConfig.IPFields;
      config.MACFields = orgConfig.MACFields;
    }
  };

  const handleDone = (e) => {
    if (e && e.detail && e.detail.update) {
      getLogSources();
    }
    page = "";
  };

  const showLogTypePage = () => {
    page = "logType";
  };
</script>

{#if page == "logSource"}
  <LogSource {logSource} on:done={handleDone} />
{:else if page == "logType"}
  <LogType on:done={handleDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    {#if busy}
      <div class="Box-header">
        <h3 class="Box-title">{$_('Setting.StartingTitle')}</h3>
      </div>
      <div class="flash mt-2">
        {$_('Setting.Starting')}<span
          class="AnimatedEllipsis"
        />
      </div>
    {:else}
      <div class="Box-header">
        <h3 class="Box-title">{$_('Setting.Title')}</h3>
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
      {#if infoMsg != ""}
        <div class="flash">
          {infoMsg}
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
            <h5 class="pb-1">
              {$_('Setting.LogFrom')}
              <button
                class="btn btn-sm float-right"
                type="button"
                on:click={() => editLogSource("")}
              >
                <Plus16 />
              </button>
            </h5>
          </div>
          <div class="form-group-body markdown-body mt-3">
            <Grid {data} {pagination} {columns} language={gridLang} />
            <label class="p-1">
              <input type="checkbox" bind:checked={config.Recursive} />
              {$_('Setting.RecTGZ')}
            </label>
            <label class="p-1">
              <input type="checkbox" bind:checked={config.ForceUTC} />
              {$_('Setting.ForceUTC')}
            </label>
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('Setting.Filter')}</h5>
          </div>
          <div class="form-group-body">
            <input
              class="form-control input-block"
              type="text"
              style="width: 100%;"
              placeholder="{$_('Setting.Filter')}"
              aria-label="{$_('Setting.Filter')}"
              bind:value={config.Filter}
            />
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('Setting.LogType')}</h5>
          </div>
          <div class="form-group-body">
            <!-- svelte-ignore a11y-no-onchange -->
            <select
              class="form-select"
              aria-label="{$_('Setting.ExtractPat')}"
              bind:value={config.Extractor}
              on:change={changeExtractor}
            >
              <option value="timeonly">{$_('Setting.TimeOnly')}</option>
              <option value="auto">{$_('Setting.AutoTypeDetect')}</option>
              {#each extractorTypeList as { Key, Name }}
                <option value={Key}>{Name}</option>
              {/each}
              <option value="custom">{$_('Setting.CustomLogType')}</option>
            </select>
            <label class="p-1">
              <input type="checkbox" bind:checked={config.Strict} />
                {$_('Setting.StrictPatCheck')}
              </label>
          </div>
        </div>
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('Setting.TimeExtract')}</h5>
          </div>
          <div class="form-group-body">
            <!-- svelte-ignore a11y-no-onchange -->
            <select
              class="form-select"
              aria-label="{$_('Setting.TimeGrinderOverride')}"
              bind:value={config.TimeGrinderOverride}
            >
              <option value="">{$_('Setting.Auto')}</option>
              <option value="AnsiC">AnsiC(Jan _2 15:04:05 2006)</option>
              <option value="Unix">Unix(Jan _2 15:04:05 MST 2006)</option>
              <option value="Ruby">Ruby(Jan _2 15:04:05 -0700 2006)</option>
              <option value="RFC822">RFC822(02 Jan 06 15:04 MST)</option>
              <option value="RFC822Z">RFC822Z(02 Jan 06 15:04 -0700)</option>
              <option value="RFC850">RFC850(02-Jan-06 15:04:05 MST)</option>
              <option value="RFC1123">RFC1123(02 Jan 2006 15:04:05 MST)</option>
              <option value="RFC1123Z">RFC1123Z(02 Jan 2006 15:04:05 -0700)</option>
              <option value="RFC3339">RFC3339(2006-01-02T15:04:05Z07:00)</option>
              <option value="RFC3339Nano">RFC3339Nano(2006-01-02T15:04:05.999999999Z07:00)</option>
              <option value="ZonelessRFC3339">ZonelessRFC3339(2006-01-02T15:04:05.999999999)</option>
              <option value="Apache">Apache(_2/Jan/2006:15:04:05 -0700)</option>
              <option value="ApacheNoTz">ApacheNoTz(_2/Jan/2006:15:04:05)</option>
              <option value="NGINX">NGINX(2006/01/02 15:04:05)</option>
              <option value="Syslog">Syslog(Jan _2 15:04:05)</option>
              <option value="SyslogFile">SyslogFile(2006-01-02T15:04:05.999999999-07:00)</option>
              <option value="SyslogFileTZ">SyslogFileTZ(2006-01-02T15:04:05.999999999-0700)</option>
              <option value="DPKG">DPKG(2006-01-02 15:04:05)</option>
              <option value="SyslogVariant">SyslogVariant(Jan 02 2006 15:04:05)</option>
              <option value="UnpaddedDateTime">UnpaddedDateTime(2006-1-2 15:04:05)</option>
              <option value="UnpaddedMilliDateTime">UnpaddedMilliDateTime(2006-1-2 15:04:05.999999999)</option>
              <option value="UK">UK(02/01/2006 15:04:05,99999)</option>
              <option value="Gravwell">Gravwell(1-2-2006 15:04:05.99999)</option>
              <option value="Bind">Bind(02-Jan-2006 15:04:05.999)</option>
              <option value="DirectAdmin">DirectAdmin(2006:01:02-15:04:05)</option>
              <option value="custom01">Jan _2 15:04:05 2006</option>
              <option value="custom02">2006/1/2 3:04:05</option>
              <option value="custom00">{$_('Setting.CustomLogType')}</option>
            </select>
          </div>
          {#if config.TimeGrinderOverride == "custom00"}
            <div class="form-group-body">
              <input
                class="form-control"
                type="text"
                placeholder="{$_('Setting.TimeGrinderRegExp')}"
                aria-label="{$_('Setting.TimeGrinderRegExp')}"
                style="width: 45%;"
                bind:value={config.TimeGrinderRegExp}
              />
              <input
                class="form-control"
                type="text"
                placeholder="{$_('Setting.TimeGrinderFormat')}"
                aria-label="{$_('Setting.TimeGrinderFormat')}"
                style="width: 45%;"
                bind:value={config.TimeGrinderFormat}
              />
            </div>
          {/if}
        </div>

        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('Setting.AddressInfo')}</h5>
          </div>
          <div class="form-group-body">
            <label class="p-1">
              <input type="checkbox" bind:checked={config.HostName} />
              {$_('Setting.CheckHostName')}
            </label>
            <label class="p-1">
              <input class="ml-2" type="checkbox" bind:checked={config.GeoIP} />
              {$_('Setting.CheckIPLoc')}
            </label>
            <label class="p-1">
              <input
                class="ml-2"
                type="checkbox"
                bind:checked={config.VendorName}
              />
              {$_('Setting.CheckVendorName')}
            </label>
          </div>
        </div>
        {#if config.Extractor == "custom" || config.Extractor == "auto" || config.Extractor.startsWith("EXT")}
          <div class="form-group">
            <div class="form-group-header">
              <h5 class="pb-1">{$_('Setting.ExtractPat')}</h5>
            </div>
            <div class="form-group-body">
              <input
                class="form-control input-block"
                type="text"
                placeholder="{$_('Setting.GrokPat')}"
                aria-label="{$_('Setting.GrokPat')}"
                style="width: 100%;"
                bind:value={config.Grok}
              />
            </div>
          </div>
          <div class="form-group">
            <div class="form-group-header">
              <h5 class="pb-1">{$_('Setting.ExtractInfo')}</h5>
            </div>
            <div class="form-group-body">
              <input
                class="form-control"
                type="text"
                style="width: 15%;"
                placeholder="{$_('Setting.TimeField')}"
                bind:value={config.TimeField}
              />
              {#if config.HostName}
                <input
                  class="form-control"
                  type="text"
                  style="width: 25%;"
                  placeholder="{$_('Setting.HostNameField')}"
                  bind:value={config.HostFields}
                />
              {/if}
              {#if config.GeoIP}
                <input
                  class="form-control"
                  type="text"
                  placeholder="{$_('Setting.IPLocFiled')}"
                  style="width: 25%;"
                  bind:value={config.GeoFields}
                />
              {/if}
              {#if config.VendorName}
                <input
                  class="form-control"
                  type="text"
                  placeholder="{$_('Setting.MacFiled')}"
                  style="width: 20%;"
                  bind:value={config.MACFields}
                />
              {/if}
            </div>
          </div>
        {/if}
        <div class="form-group">
          <div class="form-group-header">
            <h5 class="pb-1">{$_('Setting.IndexerSetting')}</h5>
          </div>
          <div class="form-group-body">
            <div class="form-checkbox">
              <label>
                <input
                  type="checkbox"
                  disabled={hasIndex}
                  bind:checked={config.InMemory}
                  aria-describedby="help-text-for-inmemory"
                />
                {$_('Setting.IndexInMemory')}
              </label>
              <p class="note" id="help-text-for-inmemory">
                {$_('Setting.InMemoryMsg')}
              </p>
            </div>
          </div>
        </div>
        {#if config.GeoIP}
          <div class="form-group">
            <div class="form-group-header">
              <h5 class="pb-1">{$_('Setting.GeoDB')}</h5>
            </div>
            <div class="form-group-body">
              <div class="input-group">
                <input
                  class="form-control"
                  type="text"
                  placeholder="{$_('Setting.GeoIPDB')}"
                  aria-label="{$_('Setting.GeoIPDB')}"
                  bind:value={config.GeoIPDB}
                />
                <span class="input-group-button">
                  <button class="btn" type="button" on:click={selectGeoIPDB}>
                    <File16 />
                  </button>
                </span>
              </div>
            </div>
          </div>
        {/if}
      </div>
      <div class="Box-footer text-right">
        <button
          class="btn btn-outline mr-1"
          type="button"
          on:click={showLogTypePage}
        >
          <Checklist16 />
          {$_('Setting.LogDefBtn')}
        </button>
        <button class="btn btn-secondary mr-1" type="button" on:click={cancel}>
          <X16 />
          {$_('Setting.EndBtn')}
        </button>
        {#if hasIndex}
          <button class="btn btn-danger mr-1" type="button" on:click={clear}>
            <Trash16 />
            {$_('Setting.DelteIndexBtn')}
          </button>
          <button class="btn btn-danger mr-1" type="button" on:click={start}>
            <Check16 />
            {$_('Setting.LoadMoreBtn')}
          </button>
          <button class="btn btn-primary mr-1" type="button" on:click={search}>
            <Search16 />
            {$_('Setting.ToSearch')}
          </button>
        {:else}
          <button class="btn btn-primary" type="button" on:click={start}>
            <Check16 />
            {$_('Setting.StartBtn')}
          </button>
        {/if}
      </div>
    {/if}
  </div>
{/if}
