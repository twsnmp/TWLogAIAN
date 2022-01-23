<script>
  import { X16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { onMount } from 'svelte';
  import { h, html } from "gridjs";
  import { typeName } from "../common.js";
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  const data = [];
  let pagination = false;
  const getProcessInfo = () => {
    window.go.main.App.GetProcessInfo().then((r) => {
      data.length = 0; // 空にする
      if (r) {
        if (r.LogFiles) {
          r.LogFiles.forEach((e) => {
            data.push([e.Type, e.URL, e.Size, e.Done]);
          });
          if (data.length > 10) {
            pagination = {
              limit: 10,
              enable: true,
            };
          } else {
            pagination = false;
          }
        }
        if (r.ErrorMsg) {
          errorMsg = r.ErrorMsg;
        }
        if (r.Done) {
          dispatch("done", { page: "search" });
        }
      }
    });
    setTimeout(getProcessInfo,1000 * 5);
  };

  onMount(() => {
    getProcessInfo();
  });

  const columns = [
    {
      name: "タイプ",
      sort: true,
      width: "20%",
      formatter: typeName,
    },
    {
      name: "パス/URL",
      sort: true,
      width: "60%",
    },
    {
      name: "サイズ",
      sort: false,
      width: "10%",
    },
    {
      name: "完了",
      sort: false,
      width: "10%",
    },
  ];

  const stop = () => {
    // Index作成を停止
    window.go.main.App.Stop().then((r) => {
      if (r === "") {
        dispatch("done", { page: "setting" });
      } else {
        errorMsg = r;
      }
    });
  };

</script>

<div class="Box mx-auto" style="max-width: 800px;">
    <div class="Box-header">
      <h3 class="Box-title">ログ分析中....</h3>
    </div>
    {#if errorMsg != ""}
      <div class="flash flash-error">
        {errorMsg}
      </div>
    {/if}
    <div class="Box-body">
        <Grid {data} {pagination} {columns} />
    </div>
    <div class="Box-footer text-right">
      <button class="btn  btn-danger" type="button" on:click={stop}>
        <X16 />
        停止
      </button>
    </div>
</div>
