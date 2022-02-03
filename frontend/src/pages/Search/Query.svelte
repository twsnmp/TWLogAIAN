<script>
  import { X16, Check16, Plus16, Trash16 } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import { validate_each_argument } from "svelte/internal";
  export let conf;
  export let fields = [];
  const dispatch = createEventDispatcher();
  let errorMsg = "";
  let limit = conf.limit;
  let start = conf.start;
  let end = conf.end;
  let range = conf.range;
  let keywords = conf.keywords;
  let geo = conf.geo;

  const keyword = {
    field: "",
    mode: "",
    key: "",
  };
  const addKeyword = () => {
    const tmp = keywords;
    tmp.push({
      field: keyword.field,
      mode: keyword.mode,
      key: keyword.key,
    });
    keywords = tmp;
  };
  const deleteKeyword = (i) => {
    const tmp = [];
    for (let j = 0; j < keywords.length; j++) {
      if (i !== j) {
        tmp.push(keywords[j]);
      }
    }
    keywords = tmp;
  };
  const cancel = () => {
    dispatch("done", {});
  };

  const save = () => {
    conf.limit = limit;
    conf.query = "";
    conf.range = range;
    conf.start = start;
    conf.end = end;
    conf.geo = geo;
    conf.keywors = keywords;
    conf.keywors.forEach((e)=> {
      if (e.field){
        conf.query += e.mode + e.field  + ":" + e.key;
      } else {
        conf.query += e.mode +  e.key;
      }
      conf.query += " ";
    });
    if(conf.range) {
      conf.query += `time:>="` + conf.start + `:00Z09:00" time:<="` + conf.end + `:00Z09:00" `;
    }
    if (conf.geo.enable) {
      conf.query += "geo:" + conf.geo.field + "," + conf.geo.lat + "," + conf.geo.long + "," + conf.geo.range;
    }
    dispatch("done", {});
  };

  const clearMsg = () => {
    errorMsg = "";
  };
</script>

<div class="Box-header">
  <h3 class="Box-title">検索条件</h3>
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
<div class="Box-body">
  <form>
    <div class="form-group">
      <div class="form-group-header">
        <h5>最大件数</h5>
      </div>
      <div class="form-group-body">
        <select class="form-select" aria-label="最大件数" bind:value={limit}>
          <option value="100">100</option>
          <option value="500">500</option>
          <option value="1000">1000</option>
          <option value="2000">2000</option>
          <option value="5000">5000</option>
          <option value="10000">10000</option>
          <option value="20000">20000</option>
        </select>
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <label>
          <input type="checkbox" bind:checked={range} />
          検索期間を指定する
        </label>
      </div>
      <div class="form-group-body">
        {#if range}
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder="開始"
            aria-label="開始"
            bind:value={start}
          />
          -
          <input
            class="form-control input-sm"
            type="datetime-local"
            placeholder="終了"
            aria-label="終了"
            bind:value={end}
          />
        {/if}
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <h5>キーワード</h5>
      </div>
      <div class="form-group-body markdown-body">
          <table>
            <thead>
              <tr>
                <th>項目</th>
                <th>有無</th>
                <th>キーワード</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {#each keywords as { field, mode, key }, i}
                <tr>
                  <td>{field}</td>
                  <td>{mode}</td>
                  <td>{key}</td>
                  <td>
                    <button
                      class="btn"
                      type="button"
                      on:click={() => {
                        deleteKeyword(i);
                      }}
                    >
                      <Trash16 />
                    </button>
                  </td>
                </tr>
              {/each}
              <tr>
                <td>
                  <select
                    class="form-select"
                    aria-label="項目"
                    bind:value={keyword.field}
                  >
                    <option value="" />
                    {#each fields as f }
                      <option value="{f}">{f}</option>
                    {/each}
                  </select>
                </td>
                <td>
                  <select
                    class="form-select"
                    aria-label="有無"
                    bind:value={keyword.mode}
                  >
                    <option value="">含む</option>
                    <option value="+">必須</option>
                    <option value="-">含まない</option>
                  </select>
                </td>
                <td>
                  <input
                    class="form-control input-sm"
                    type="text"
                    placeholder="キーワード"
                    aria-label="キーワード"
                    bind:value={keyword.key}
                  />
                </td>
                <td>
                  <button class="btn" type="button" on:click={addKeyword}>
                    <Plus16 />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
      </div>
    </div>
    <div class="form-group">
      <div class="form-group-header">
        <label>
          <input type="checkbox" bind:checked={geo.enable} />
          位置範囲を指定する
        </label>
      </div>
      <div class="form-group-body container-lg clearfix">
        {#if geo.enable}
          <div class="col-3 float-left p-2">
            <select class="form-select" aria-label="項目" bind:value={geo.field}>
              {#each fields as f }
                <option value="{f}">{f}</option>
              {/each}
            </select>
          </div>
          <div class="col-3 float-left p-2">
            <input
            class="form-control input-sm"
            type="text"
            placeholder="緯度"
            aria-label="緯度"
            bind:value={geo.lat}
            />
          </div>
          <div class="col-3 float-left p-2">
            <input
            class="form-control input-sm"
            type="text"
            placeholder="経度"
            aria-label="経度"
            bind:value={geo.long}
            />
          </div>
          <div class="col-3 float-left p-2">
            <input
              class="form-control input-sm col-3"
              type="text"
              placeholder="範囲"
              aria-label="範囲"
              bind:value={geo.range}
            />
          </div>
        {/if}
      </div>
    </div>
  </form>
</div>
<div class="Box-footer text-right">
  <button class="btn btn-secondary" type="button" on:click={cancel}>
    <X16 />
    キャンセル
  </button>
  <button class="btn btn-primary ml-2" type="button" on:click={save}>
    <Check16 />
    適用
  </button>
</div>

<style>
  select {
    min-width: 100px;
  }
</style>