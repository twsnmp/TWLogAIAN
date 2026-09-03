<script>
  import { createEventDispatcher, onMount, onDestroy } from "svelte";
  import { X16, Trash16, Download16, Cpu16, Check16, Sync16 } from "svelte-octicons";
  import { _ } from "../../i18n/i18n";
  import {
    GetAIHardwareStatus,
    GetLocalModels,
    GetModelPresets,
    DownloadModel,
    CancelModelDownload,
    DeleteLocalModel,
    DownloadGPULibrary,
    CancelGPUDownload,
  } from "../../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";

  const dispatch = createEventDispatcher();

  let hardwareStatus = {
    acceleration: "",
    detail: "",
    model_dir: "",
    lib_dir: "",
    wgpu_lib_path: "",
    has_gpu_lib: false,
  };

  let localModels = [];
  let presets = [];

  let selectedPreset = "";
  let customTarget = "";

  let isDownloadingModel = false;
  let modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };

  let isSettingUpGPU = false;
  let gpuProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };

  let errorMsg = "";
  let infoMsg = "";

  const refreshData = async () => {
    try {
      hardwareStatus = await GetAIHardwareStatus();
      localModels = (await GetLocalModels()) || [];
      presets = (await GetModelPresets()) || [];
      if (presets.length > 0 && !selectedPreset) {
        selectedPreset = presets[0].name;
      }
    } catch (e) {
      errorMsg = e.message || String(e);
    }
  };

  onMount(async () => {
    await refreshData();

    EventsOn("model_download_progress", (data) => {
      modelProgress = data;
    });

    EventsOn("gpu_download_progress", (data) => {
      gpuProgress = data;
    });
  });

  onDestroy(() => {
    EventsOff("model_download_progress");
    EventsOff("gpu_download_progress");
  });

  const downloadPreset = async () => {
    if (!selectedPreset) return;
    errorMsg = "";
    infoMsg = "";
    isDownloadingModel = true;
    modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadModel(selectedPreset);
      isDownloadingModel = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = "Model downloaded successfully.";
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e) {
      isDownloadingModel = false;
      errorMsg = e.message || String(e);
    }
  };

  const downloadCustom = async () => {
    if (!customTarget.trim()) return;
    errorMsg = "";
    infoMsg = "";
    isDownloadingModel = true;
    modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadModel(customTarget.trim());
      isDownloadingModel = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = "Model downloaded successfully.";
        customTarget = "";
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e) {
      isDownloadingModel = false;
      errorMsg = e.message || String(e);
    }
  };

  const cancelModel = async () => {
    try {
      await CancelModelDownload();
      isDownloadingModel = false;
    } catch (e) {
      errorMsg = e.message || String(e);
    }
  };

  const deleteModel = async (name) => {
    errorMsg = "";
    infoMsg = "";
    try {
      const err = await DeleteLocalModel(
        name,
        $_("ModelManager.Delete"),
        $_("ModelManager.DeleteConfirm", { values: { name } })
      );
      if (err && err !== "cancel") {
        errorMsg = err;
      } else if (!err) {
        infoMsg = "Model deleted successfully.";
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e) {
      errorMsg = e.message || String(e);
    }
  };

  const setupGPU = async () => {
    errorMsg = "";
    infoMsg = "";
    isSettingUpGPU = true;
    gpuProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadGPULibrary();
      isSettingUpGPU = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = "GPU library installed successfully.";
        await refreshData();
      }
    } catch (e) {
      isSettingUpGPU = false;
      errorMsg = e.message || String(e);
    }
  };

  const cancelGPU = async () => {
    try {
      await CancelGPUDownload();
      isSettingUpGPU = false;
    } catch (e) {
      errorMsg = e.message || String(e);
    }
  };

  const close = () => {
    dispatch("close");
  };
</script>

<div class="Box mx-auto Box--condensed" style="max-width: 99%; display: flex; flex-direction: column; max-height: calc(100vh - 40px);">
  <div class="Box-header d-flex flex-items-center flex-justify-between">
    <h3 class="Box-title">
      <Cpu16 class="mr-1" />
      {$_("ModelManager.Title")}
    </h3>
    <button class="btn-octicon" type="button" on:click={close}>
      <X16 />
    </button>
  </div>

  {#if errorMsg}
    <div class="flash flash-error m-2">
      {errorMsg}
      <button class="flash-close js-flash-close" type="button" on:click={() => (errorMsg = "")}>
        <X16 />
      </button>
    </div>
  {/if}

  {#if infoMsg}
    <div class="flash flash-success m-2">
      {infoMsg}
      <button class="flash-close js-flash-close" type="button" on:click={() => (infoMsg = "")}>
        <X16 />
      </button>
    </div>
  {/if}

  <div class="Box-body" style="overflow-y: auto; flex: 1;">
    <!-- Hardware Acceleration Status -->
    <div class="Box p-3 mb-3 bg-gray-light">
      <h4 class="mb-2 d-flex flex-items-center">
        <Cpu16 class="mr-2" />
        {$_("ModelManager.GPUAcceleration")}
      </h4>
      <div class="d-flex flex-wrap flex-items-center mb-2">
        <span class="mr-2 text-bold">{$_("ModelManager.ActiveBackend")}:</span>
        <span class="Label Label--large {hardwareStatus.acceleration === 'GPU' ? 'Label--success' : hardwareStatus.acceleration.includes('SIMD') ? 'Label--accent' : 'Label--secondary'}">
          {hardwareStatus.acceleration || "Detecting..."}
        </span>
        <span class="ml-2 color-fg-muted">{hardwareStatus.detail}</span>
      </div>
      <div class="d-flex flex-wrap flex-items-center">
        <span class="mr-2 text-bold">{$_("ModelManager.WGPULibrary")}:</span>
        {#if hardwareStatus.has_gpu_lib}
          <span class="color-fg-success d-flex flex-items-center mr-2">
            <Check16 class="mr-1" />
            {$_("ModelManager.GPULibInstalled")}
          </span>
          <span class="color-fg-muted text-small">({hardwareStatus.wgpu_lib_path})</span>
        {:else}
          <span class="color-fg-attention mr-3">
            {$_("ModelManager.GPULibNotInstalled")}
          </span>
          {#if isSettingUpGPU}
            <div class="d-flex flex-items-center" style="flex: 1;">
              <span class="mr-2">{$_("ModelManager.SettingUpGPU")} {gpuProgress.percent}% ({gpuProgress.downloaded_human} / {gpuProgress.total_human})</span>
              <button class="btn btn-sm btn-danger ml-2" type="button" on:click={cancelGPU}>
                {$_("ModelManager.Cancel")}
              </button>
            </div>
          {:else}
            <button class="btn btn-sm btn-primary" type="button" on:click={setupGPU}>
              <Download16 class="mr-1" />
              {$_("ModelManager.SetupGPU")}
            </button>
          {/if}
        {/if}
      </div>
      <p class="note mt-1">{$_("ModelManager.GPUHelp")}</p>
    </div>

    <!-- Downloaded Models List -->
    <div class="Box mb-3">
      <div class="Box-header d-flex flex-items-center flex-justify-between">
        <h4 class="Box-title">{$_("ModelManager.LocalModels")}</h4>
        <span class="Counter">{localModels.length}</span>
      </div>
      {#if localModels.length === 0}
        <div class="Box-body color-fg-muted text-center py-4">
          {$_("ModelManager.NoModels")}
        </div>
      {:else}
        <div class="Box-body p-0">
          <table class="width-full">
            <thead>
              <tr class="border-bottom text-left">
                <th class="p-2">{$_("ModelManager.Name")}</th>
                <th class="p-2">{$_("ModelManager.Size")}</th>
                <th class="p-2">{$_("ModelManager.ModTime")}</th>
                <th class="p-2 text-right">{$_("ModelManager.Action")}</th>
              </tr>
            </thead>
            <tbody>
              {#each localModels as m}
                <tr class="border-bottom">
                  <td class="p-2 text-bold">{m.name}</td>
                  <td class="p-2">{m.size_human}</td>
                  <td class="p-2 color-fg-muted">{new Date(m.mod_time).toLocaleString()}</td>
                  <td class="p-2 text-right">
                    <button class="btn btn-sm btn-danger" type="button" on:click={() => deleteModel(m.name)}>
                      <Trash16 />
                      {$_("ModelManager.Delete")}
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>

    <!-- Model Download Progress -->
    {#if isDownloadingModel}
      <div class="Box p-3 mb-3 bg-blue-light border-blue">
        <div class="d-flex flex-items-center flex-justify-between mb-2">
          <span class="text-bold">
            <Sync16 class="anim-rotate mr-1" />
            {$_("ModelManager.Downloading")} {modelProgress.percent}% ({modelProgress.downloaded_human} / {modelProgress.total_human})
          </span>
          <button class="btn btn-sm btn-danger" type="button" on:click={cancelModel}>
            {$_("ModelManager.Cancel")}
          </button>
        </div>
        <div class="Progress" style="height: 12px;">
          <span class="Progress-item bg-blue" style="width: {modelProgress.percent}%;"></span>
        </div>
      </div>
    {/if}

    <!-- Download Preset Model -->
    <div class="Box p-3 mb-3">
      <h4 class="mb-2">{$_("ModelManager.DownloadPreset")}</h4>
      <p class="note mb-2">{$_("ModelManager.PresetHelp")}</p>
      <div class="d-flex flex-items-center">
        <select class="form-select flex-1 mr-2" bind:value={selectedPreset} disabled={isDownloadingModel}>
          {#each presets as p}
            <option value={p.name}>
              {p.name} ({p.size}, {p.description})
            </option>
          {/each}
        </select>
        <button class="btn btn-primary" type="button" on:click={downloadPreset} disabled={isDownloadingModel || !selectedPreset}>
          <Download16 class="mr-1" />
          {$_("ModelManager.Download")}
        </button>
      </div>
    </div>

    <!-- Download Custom Model -->
    <div class="Box p-3">
      <h4 class="mb-2">{$_("ModelManager.DownloadCustom")}</h4>
      <div class="d-flex flex-items-center">
        <input
          type="text"
          class="form-control flex-1 mr-2"
          placeholder={$_("ModelManager.CustomPlaceholder")}
          bind:value={customTarget}
          disabled={isDownloadingModel}
        />
        <button class="btn btn-secondary" type="button" on:click={downloadCustom} disabled={isDownloadingModel || !customTarget.trim()}>
          <Download16 class="mr-1" />
          {$_("ModelManager.Download")}
        </button>
      </div>
    </div>
  </div>

  <div class="Box-footer text-right">
    <button class="btn btn-secondary" type="button" on:click={close}>
      {$_("ModelManager.Close")}
    </button>
  </div>
</div>
