<script>
  import { getFields, getFieldName,getTableLimit } from "../../js/define";
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { getRanking, showRankingChart, resizeRankingChart,getRankingChartImage } from "./ranking";
  import Grid from "../../components/Grid.svelte";
  import jaJP from "../../js/gridjsJaJP";
  import { _,getLocale } from '../../i18n/i18n';
  import {Export, LLMExplainReport} from '../../../wailsjs/go/main/App';
  import ReportAIDialog from "../../components/ReportAIDialog.svelte";

  let locale = getLocale();
  let gridLang = locale == "ja" ? jaJP : undefined;

  export let logs = [];
  export let fields = [];
  export let dark = false;
  export let hasLLM = false;

  let showAIDialog = false;
  let aiLang = locale === "en" ? "en" : "ja";
  let selectAILang = "";

  const onSelectAILang = () => {
    if (!selectAILang) return;
    aiLang = selectAILang;
    showAIDialog = true;
    selectAILang = "";
  };

  let list = [];
  let catFields = [];
  let selected = "";

  const dispatch = createEventDispatcher();
  let data = [];
  let columns = [
    {
      name: $_('Ranking.Rank'),
      width: "10%",
    },
    {
      name: $_('Ranking.Item'),
      width: "80%",
    },
    {
      name: $_('Ranking.Count'),
      width: "10%",
    },
  ];
  let pagination = false;

  const updateRanking = async () => {
    if (selected == "" ) {
      return;
    }
    await tick();
    list = getRanking(logs, selected);
    showRankingChart("chart", list, 50,dark);
    if (list.length > 10) {
      pagination = {
        limit: getTableLimit(),
        enable: true,
      };
    } else {
      pagination = false;
    }
    data = [];
    for (let i = 0; i < list.length; i++) {
      data.push([i + 1, list[i].Name, list[i].Total]);
    }
  };

  onMount(() => {
    catFields = getFields(fields,"string,number");
    if( catFields.length > 0 ){
      selected = catFields[0];
      updateRanking();
    }
  });

  let exportType = '';
  let saveBusy = false;
  const exportReport = async () => {
    if (exportType == "") {
      return;
    }
    saveBusy = true;
    const exportData = {
      Type: $_('Ranking.ExportType'),
      Title: $_('Ranking.ExportTitle'),
      Header: [],
      Data: [],
      Image: "",
    };
    if (exportType == "excel") {
      exportData.Image = getRankingChartImage();
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
    await Export(exportType,exportData);
    saveBusy = false;
    exportType = "";
  }

  const onResize = () => {
    resizeRankingChart();
  };

  const back = () => {
    dispatch("done", {});
  };

</script>

<svelte:window on:resize={onResize} />
<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{$_('Ranking.Title')}</h3>
    <!-- svelte-ignore a11y-no-onchange -->
    <select
      class="form-select"
      bind:value={selected}
      on:change={updateRanking}
    >
    <option value="">{$_('Ranking.SelectSumItemMsg')}</option>
      {#each catFields as f}
        <option value={f}>{getFieldName(f)}</option>
      {/each}
    </select>
  </div>
  <div style="overflow-y: auto; flex: 1;">
    <div class="Box-row">
      <div id="chart"></div>
    </div>
    <div class="Box-row markdown-body log">
      <Grid {data} sort search {pagination} {columns} language={gridLang} />
    </div>
  </div>
  <div class="Box-footer d-flex flex-justify-between">
    <div class="d-flex flex-items-center">
      <button class="btn btn-secondary" type="button" on:click={back}>
        <X16 />
        {$_('Ranking.BackBtn')}
      </button>
      {#if hasLLM}
        <!-- svelte-ignore a11y-no-onchange -->
        <select class="form-select ml-2" bind:value={selectAILang} on:change={onSelectAILang}>
          <option value="">{$_('ReportAI.AIExplain')}</option>
          <option value="ja">{$_('ReportAI.LangJa')}</option>
          <option value="en">{$_('ReportAI.LangEn')}</option>
        </select>
      {/if}
    </div>
    <div>
      {#if data.length > 0}
        <!-- svelte-ignore a11y-no-onchange -->
        {#if saveBusy}
          <span>{$_('Ranking.Saving')}</span><span class="AnimatedEllipsis"></span>
        {:else}
          <select class="form-select" bind:value={exportType} on:change="{exportReport}">
            <option value="">{$_('Ranking.ExportBtn')}</option>
            <option value="csv">CSV</option>
            <option value="excel">Excel</option>
          </select>
        {/if}
      {/if}
    </div>
  </div>
</div>

<ReportAIDialog
  bind:show={showAIDialog}
  bind:lang={aiLang}
  title={$_('ReportAI.Title') + ' - ' + $_('Ranking.Title')}
  exportFilename="ranking_ai_explanation"
  analyzeFunc={(targetLang) => LLMExplainReport({
    ReportType: "ranking",
    Title: $_('Ranking.Title'),
    Field: getFieldName(selected) || selected,
    Columns: columns.map(c => c.name),
    Data: data.slice(0, 60),
    Lang: targetLang
  })}
/>



<style>
  #chart {
    width: 100%;
    height: 250px;
    margin: 5px auto;
  }
</style>
