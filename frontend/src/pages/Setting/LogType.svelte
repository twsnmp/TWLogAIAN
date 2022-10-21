<script>
  import {
    X16,
    Plus16,
    Download16,
    Upload16,
  } from "svelte-octicons";
  import { createEventDispatcher } from "svelte";
  import Grid from "gridjs-svelte";
  import { h, html } from "gridjs";
  import EditExtractorType from "./EditExtractorType.svelte";
  import EditFieldType from "./EditFieldType.svelte";
  import { onMount } from "svelte";
  import jaJP from "../../js/gridjsJaJP";

  const dispatch = createEventDispatcher();
  let page = "";
  let extractorTypeErrorMsg = "";
  let fieldTypeErrorMsg = "";
  let gridSearch = true;

  let extractorTypes = {};
  let extractorTypeList = [];
  let extractorTypePagination = {
    enable: true,
    limit: 10,
  };

  let fieldTypes = {};
  let fieldTypeList = [];
  let fieldTypePagination = {
    enable: true,
    limit: 10,
  };

  onMount(() => {
    getFieldTypes();
    getExtractorTypes();
  });

  const getExtractorTypes = () => {
    window.go.main.App.GetExtractorTypes().then((r) => {
      if (r) {
        extractorTypes = r;
        extractorTypeList = [];
        for(let k in extractorTypes) {
          const e = extractorTypes[k];
          extractorTypeList.push([e.Key,e.Name,"",""]);
        };
      }
    });
  };

  const getFieldTypes = () => {
    window.go.main.App.GetFieldTypes().then((r) =>{
      if (r) {
        fieldTypes = r;
        fieldTypeList = [];
        for(let k in fieldTypes) {
          const e = fieldTypes[k];
          fieldTypeList.push([e.Key,e.Name,e.Type,e.Unit,"",""]);
        };
      }
    });
  }

  const importLogTypes = () => {
    window.go.main.App.ImportLogTypes().then((r) => {
      extractorTypeErrorMsg = r;
      if (r == "") {
        getExtractorTypes();
        getFieldTypes();
      }
    });
  };

  const exportLogTypes = () => {
    window.go.main.App.ExportLogTypes().then((r) => {
      extractorTypeErrorMsg = r;
    });
  };

  const deleteExtractorType = (key) => {
    window.go.main.App.DeleteExtractorType(key).then((r) => {
      extractorTypeErrorMsg = r;
      if (r == "") {
        getExtractorTypes();
        getFieldTypes();
      }
    });
  };

  const deleteFieldType = (key) => {
    window.go.main.App.DeleteFieldType(key).then((r) => {
      fieldTypeErrorMsg = r;
      if (r == "") {
        getFieldTypes();
      }
    });
  };
 
  let extractorType = {};
  let add  = true;
  
  const editExtractorType = (key) => {
    const et = extractorTypes[key] ||
      {
        Key: "",
        Name: "New",
        Grok: "",
        View: "",
        TimeField: "",
        IPFields: "",
      	MACFields: "",
	      View: "",
        CanEdit: true,
      };
    extractorType = Object.assign({},et);
    if (!extractorType.CanEdit) {
        extractorType.Key += "_copy";
        extractorType.CanEdit = true;
        add = true;
    } else {
      add = extractorType.Key === "";
      const now = new Date();
      extractorType.Key = "e" + now.getTime();
    }
    page = "extractorType";
  }

  const editExtractorTypeButton = (_, row) => {
    const key = row.cells[0].data;
    const et = extractorTypes[key];
    if ( !et ) {
      return "";
    }
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => editExtractorType(key),
      },
      html(
        et.CanEdit ?
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg"  viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>`
        :
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>`
      )
    );
  };

  const deleteExtractorTypeButton = (_, row) => {
    const key = row.cells[0].data;
    if (!extractorTypes[key] || !extractorTypes[key].CanEdit ) {
      return "";
    }
    return h(
      "button",
      {
        className: "btn btn-sm btn-danger",
        onClick: () => deleteExtractorType(key),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.576l.66-6.6a.75.75 0 00-1.492-.149l-.66 6.6a.25.25 0 01-.249.225h-5.19a.25.25 0 01-.249-.225l-.66-6.6z"></path></svg>`
      )
    );
  };

  const extractorTypeColumns = [
    {
      name: "Key",
      width: "20%",
    },
    {
      name: "名前",
      width: "70%",
    },
    {
      name: "編集",
      sort: false,
      width: "5%",
      formatter: editExtractorTypeButton,
    },
    {
      name: "削除",
      sort: false,
      width: "5%",
      formatter: deleteExtractorTypeButton,
    },
  ];

  let fieldType = {};

  const editFieldType = (key) => {
    const ft = fieldTypes[key] || {
      Key: "",
      Name: "New",
      Unit: "",
      CanEdit: true,
    };
    fieldType = Object.assign({},ft);
    if (!fieldType.CanEdit) {
      fieldType.Key += "_copy";
      fieldType.CanEdit = true;
      add = true;
    } else {
      add = fieldType.Key == "";
    }
    page = "fieldType";
  }

  const editFieldTypeButton = (_, row) => {
    const key = row.cells[0].data;
    const ft = fieldTypes[key] 
    if (!ft) {
      return "";
    }
    return h(
      "button",
      {
        className: "btn btn-sm",
        onClick: () => editFieldType(key),
      },
      html(
        ft.CanEdit ?
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg"  viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M11.013 1.427a1.75 1.75 0 012.474 0l1.086 1.086a1.75 1.75 0 010 2.474l-8.61 8.61c-.21.21-.47.364-.756.445l-3.251.93a.75.75 0 01-.927-.928l.929-3.25a1.75 1.75 0 01.445-.758l8.61-8.61zm1.414 1.06a.25.25 0 00-.354 0L10.811 3.75l1.439 1.44 1.263-1.263a.25.25 0 000-.354l-1.086-1.086zM11.189 6.25L9.75 4.81l-6.286 6.287a.25.25 0 00-.064.108l-.558 1.953 1.953-.558a.249.249 0 00.108-.064l6.286-6.286z"></path></svg>`
        :
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"></path><path fill-rule="evenodd" d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"></path></svg>`
      )
    );
  };

  const deleteFieldTypeButton = (_, row) => {
    const key = row.cells[0].data;
    if (!fieldTypes[key] || !fieldTypes[key].CanEdit ) {
        return "";
    }
    return h(
      "button",
      {
        className: "btn btn-sm btn-danger",
        onClick: () => deleteFieldType(key),
      },
      html(
        `<svg class="octicon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"><path fill-rule="evenodd" d="M6.5 1.75a.25.25 0 01.25-.25h2.5a.25.25 0 01.25.25V3h-3V1.75zm4.5 0V3h2.25a.75.75 0 010 1.5H2.75a.75.75 0 010-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75zM4.496 6.675a.75.75 0 10-1.492.15l.66 6.6A1.75 1.75 0 005.405 15h5.19c.9 0 1.652-.681 1.741-1.576l.66-6.6a.75.75 0 00-1.492-.149l-.66 6.6a.25.25 0 01-.249.225h-5.19a.25.25 0 01-.249-.225l-.66-6.6z"></path></svg>`
      )
    );
  };

  const fieldTypeColumns = [
    {
      name: "Key",
      width: "20%",
    },
    {
      name: "名前",
      width: "50%",
    },
    {
      name: "種類",
      width: "10%",
    },
    {
      name: "単位",
      width: "10%",
    },
    {
      name: "編集",
      sort: false,
      width: "5%",
      formatter: editFieldTypeButton,
    },
    {
      name: "削除",
      sort: false,
      width: "5%",
      formatter: deleteFieldTypeButton,
    },
  ];

  const close = () => {
    dispatch("done", {});
  };

  const clearMsg = () => {
    fieldTypeErrorMsg = "";
    extractorTypeErrorMsg = "";
  };

  const handleEditExtractorDone = (e) => {
    if (e && e.detail && e.detail.save) {
      getExtractorTypes();
    }
    page = "";
  };

  const handleEditFieldTypeDone = (e) => {
    if (e && e.detail && e.detail.save) {
      getFieldTypes();
    }
    page = "";
  };

</script>

{#if page == "extractorType"}
  <EditExtractorType {extractorType} {add} on:done={handleEditExtractorDone} />
{:else if page == "fieldType"}
  <EditFieldType {fieldType} {add} on:done={handleEditFieldTypeDone} />
{:else}
  <div class="Box mx-auto Box--condensed" style="max-width: 99%;">
    <div class="Box-header">
      <h3 class="Box-title pb-2">
        抽出パターン定義
        <button
        class="btn btn-sm float-right"
        type="button"
        on:click={() => editExtractorType("")}
      >
        <Plus16 />
      </button>
      </h3>
    </div>
    {#if extractorTypeErrorMsg != ""}
      <div class="flash flash-error">
        {extractorTypeErrorMsg}
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
      <div class="markdown-body mt-3">
        <Grid 
          data={extractorTypeList} 
          sort
          resizable
          search={gridSearch}
          pagination={extractorTypePagination} 
          columns={extractorTypeColumns} 
          language={jaJP} 
        />
      </div>
    </div>
    <div class="Box-header">
      <h3 class="Box-title pb-2">
        フィールド定義
        <button
          class="btn btn-sm float-right"
          type="button"
          on:click={() => editFieldType("")}>
        <Plus16 />
        </button>
      </h3>
    </div>
    {#if fieldTypeErrorMsg != ""}
      <div class="flash flash-error">
        {fieldTypeErrorMsg}
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
      <div class="markdown-body mt-3">
        <Grid 
          data={fieldTypeList} 
          sort
          resizable
          search={gridSearch}
          pagination={fieldTypePagination} 
          columns={fieldTypeColumns} 
          language={jaJP} 
        />
      </div>
    </div>
    <div class="Box-footer text-right">
      <button
      class="btn btn-outline mr-1"
      type="button"
      on:click={exportLogTypes}
    >
      <Download16 />
      エクスポート
      </button>
      <button
      class="btn btn-outline mr-1"
      type="button"
      on:click={importLogTypes}
    >
      <Upload16 />
      インポート
    </button>
    <button class="btn btn-secondary mr-1" type="button" on:click={close}>
        <X16 />
        閉じる
      </button>
    </div>
  </div>
{/if}
