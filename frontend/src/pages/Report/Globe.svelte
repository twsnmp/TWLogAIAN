<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showGlobe, resizeGlobe,getGlobeImage } from "./globe";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  export let logs = [];
  export let fields = [];

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;


  let dark = false;
  let geoFields = [];
  let numFields = [];
  let srcField = "";
  let dstField = "";
  let numField = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Globe.StartItem'),
      width: "40%",
    },
    {
      name: $_('Globe.EndItem'),
      width: "40%",
    },
    {
      name: $_('Globe.Count'),
      width: "10%",
    },
    {
      name: $_('Globe.Value'),
      width: "10%",
    },
  ];

  let pagination = false;

  const updateGlobe = async () => {
    if( srcField == "" ){
      return;
    }
    await tick();
    const m = showGlobe("chart",logs,srcField,dstField,numField,dark);
    data = [];
    columns[0].name = getFieldName(srcField);
    columns[1].name = getFieldName(dstField);
    columns[3].name = getFieldName(numField);
    m.forEach((e) => {
      data.push([e.src,e.dst,e.count,e.value]);
    });
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  }

  onMount(() => {
    geoFields =  getFields(fields,"latlong");
    if(srcField == "" && geoFields.length >0 ){
      srcField = geoFields[0];
    }
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateGlobe();
    });
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Globe.ExportType'),
      Title: $_('Globe.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getGlobeImage();
    }
    columns.forEach((e)=>{
      exportData.Header.push(e.name);
    });
    data.forEach((l)=>{
      const row = [];
      l.forEach((e,i)=>{
        const v = columns[i] && columns[i].formatter ? columns[i].formatter(e) : e;
        row.push(v);
      });
      exportData.Data.push(row);
    });
    window.go.main.App.Export(exportType,exportData).then(()=>{
      saveBusy = false;
      exportType = "";
    });
  }

  const onResize = () => {
    resizeGlobe();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Globe.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="{$_('Globe.StartItem')}"
      bind:value={srcField}
      on:change="{updateGlobe}"
    >
      <option value="">{$_('Globe.SelectStartItemMsg')}</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="{$_('Globe.EndItem')}"
      bind:value={dstField}
      on:change="{updateGlobe}"
    >
      <option value="">{$_('Globe.Tokyo')}</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="{$_('Globe.NumberItem')}"
      bind:value={numField}
      on:change="{updateGlobe}"
    >
    <option value="">{$_('Globe.SelectNumberItemMsg')}</option>
      {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={gridLang} />
  </div>
  <div class="Box-footer text-right">
    {#if data.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>{$_('Globe.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('Globe.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('Globe.BackBtn')}
    </button>
  </div>
</div>

<style>
  #chart {
    width: 100%;
    height: 500px;
    margin: 5px auto;
  }
</style>
