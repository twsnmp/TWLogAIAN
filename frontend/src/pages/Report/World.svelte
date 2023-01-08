<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showWorldMap, resizeWorldMap,getWorldMapImage } from "./world";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  let dark = false;
  let numFields = [];
  let geoFields = [];
  let geoField = "";
  let numField = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('World.LatLong'),
      width: "20%",
    },
    {
      name: $_('World.Country'),
      width: "10%",
    },
    {
      name: $_('World.City'),
      width: "10%",
    },
    {
      name: $_('World.Count'),
      width: "10%",
    },
    {
      name: $_('World.Value'),
      width: "10%",
    },
  ];

  let pagination = false;

  const updateWorldMap = async () => {
    if( geoField == "" ){
      return;
    }
    await tick();
    const m = showWorldMap("chart", logs, geoField,numField,dark);
    data = [];
    columns[4].name = getFieldName(numField);
    m.forEach((e,k) => {
      data.push([k,e.country,e.city,e.count,e.value]);
    });
    if (data.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
  };

  onMount(() => {
    const tmp =  getFields(fields,"latlong");
    geoFields = [];
    tmp.forEach((g)=>{
      g = g.replace("_latlong","");
      if (geoField == "") {
        geoField = g;
      }
      geoFields.push(g);
    });
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
      updateWorldMap();
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
      Type: $_('World.ExportType'),
      Title: $_('World.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getWorldMapImage();
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
    resizeWorldMap();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('World.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="{$_('World.LocInfoItem')}"
      bind:value={geoField}
      on:change="{updateWorldMap}"
    >
      <option value="">{$_('World.SelectLocItem')}</option>
      {#each geoFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="{$_('World.NumberValue')}"
      bind:value={numField}
      on:change="{updateWorldMap}"
    >
    <option value="">{$_('World.SelectColorItemMsg')}</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
  </div>
  <div class="Box-row">
    <div id="chart" />
  </div>
  <div class="Box-row markdown-body log">
    <Grid {data} sort search {pagination} {columns} language={jaJP} />
  </div>
  <div class="Box-footer text-right">
    {#if data.length > 0}
      <!-- svelte-ignore a11y-no-onchange -->
      {#if saveBusy}
        <span>{$_('World.Saving')}</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">{$_('World.ExportBtn')}</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      {$_('World.BackBtn')}
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
