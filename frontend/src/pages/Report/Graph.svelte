<script>
  import { getFieldName,getFields,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { showGraph, resizeGraph,getGraphImage } from "./graph";
  import Grid from "gridjs-svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  let dark = false;
  let catFields = [];
  let numFields = [];
  let srcField = "";
  let dstField = "";
  let numField = "";
  let graphType = "gl";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Graph.StartItem'),
      width: "40%",
    },
    {
      name: $_('Graph.EndItem'),
      width: "40%",
    },
    {
      name: $_('Graph.Count'),
      width: "10%",
    },
    {
      name: $_('Graph.Value'),
      width: "10%",
    },
  ];

  let pagination = false;

  const updateGraph = async () => {
    if( srcField == "" || dstField == "" ){
      return;
    }
    await tick();
    const m = showGraph("chart", logs, srcField,dstField,numField, graphType,dark);
    data = [];
    columns[0].name = getFieldName(srcField);
    columns[1].name = getFieldName(dstField);
    columns[3].name = getFieldName(numField);
    m.forEach((e) => {
      data.push([e.source,e.target,e.count,e.value]);
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
    catFields =  getFields(fields,"string");
    numFields = getFields(fields,"number");
    if(numField == "" && numFields.length >0 ){
      numField = numFields[0];
    }
    window.go.main.App.GetDark().then((v) => {
      dark = v;
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
      Type: $_('Graph.ExportType'),
      Title: $_('Graph.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getGraphImage();
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
    resizeGraph();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%;">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">$_('Graph.Title')</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      aria-label="$_('Graph.StartItem')"
      bind:value={srcField}
      on:change="{updateGraph}"
    >
      <option value="">$_('Graph.SelectStartItemMsg')</option>
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="$_('Graph.EndItem')"
      bind:value={dstField}
      on:change="{updateGraph}"
    >
    <option value="">$_('Graph.SelectEndItemMsg')</option>
    {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="$_('Graph.NumberItem')"
      bind:value={numField}
      on:change="{updateGraph}"
    >
    <option value="">$_('Graph.SelectNumberItem')</option>
    {#each numFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select ml-2"
      aria-label="$_('Graph.DisplayMode')"
      bind:value={graphType}
      on:change="{updateGraph}"
    >
    <option value="gl">3D</option>
    <option value="force">$_('Graph.ForceModel')</option>
    <option value="circular">$_('Graph.Circle')</option>
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
        <span>$_('Graph.Saving')</span><span class="AnimatedEllipsis"></span>
      {:else}
        <select class="form-select" bind:value={exportType} on:change="{exportReport}">
          <option value="">$_('Graph.ExportBtn')</option>
          <option value="csv">CSV</option>
          <option value="excel">Excel</option>
        </select>
      {/if}
    {/if}
    <button class="btn btn-secondary" type="button" on:click={back}>
      <X16 />
      $_('Graph.BackBtn')
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
