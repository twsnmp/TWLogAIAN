package main

import (
	"context"
	"os"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/TWLogAIAN/pkg/ai/tensai"
	"github.com/twsnmp/TWLogAIAN/pkg/model"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

// AIHardwareStatus contains hardware acceleration details
type AIHardwareStatus struct {
	Acceleration string `json:"acceleration"`
	Detail       string `json:"detail"`
	ModelDir     string `json:"model_dir"`
	LibDir       string `json:"lib_dir"`
	WGPULibPath  string `json:"wgpu_lib_path"`
	HasGPULib    bool   `json:"has_gpu_lib"`
}

// DownloadProgressEvent is emitted to frontend during downloads
type DownloadProgressEvent struct {
	Downloaded      int64  `json:"downloaded"`
	Total           int64  `json:"total"`
	Percent         int    `json:"percent"`
	DownloadedHuman string `json:"downloaded_human"`
	TotalHuman      string `json:"total_human"`
}

var (
	modelCancelMu  sync.Mutex
	modelCancelCtx context.CancelFunc

	gpuCancelMu  sync.Mutex
	gpuCancelCtx context.CancelFunc
)

// GetAIHardwareStatus returns hardware acceleration info and library paths
func (b *App) GetAIHardwareStatus() AIHardwareStatus {
	tensai.InitWGPULibrary()
	accType, accDetail := tensai.DetectAcceleration()

	wgpuLib := os.Getenv("TENSAI_WGPU_LIB")
	hasLib := wgpuLib != ""

	return AIHardwareStatus{
		Acceleration: string(accType),
		Detail:       accDetail,
		ModelDir:     model.DefaultModelDir(),
		LibDir:       model.DefaultLibDir(),
		WGPULibPath:  wgpuLib,
		HasGPULib:    hasLib,
	}
}

// GetLocalModels returns list of available local models
func (b *App) GetLocalModels() []model.ModelInfo {
	models, err := model.ListModels("")
	if err != nil {
		OutLog("GetLocalModels err=%v", err)
		return []model.ModelInfo{}
	}
	return models
}

// GetModelPresets returns list of recommended preset models
func (b *App) GetModelPresets() []model.PresetModelInfo {
	return model.PresetModelMetadata
}

// DownloadModel downloads a model by preset name or URL
func (b *App) DownloadModel(target string) string {
	modelCancelMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	modelCancelCtx = cancel
	modelCancelMu.Unlock()

	defer func() {
		modelCancelMu.Lock()
		modelCancelCtx = nil
		modelCancelMu.Unlock()
	}()

	progress := func(downloaded, total int64) {
		pct := 0
		if total > 0 {
			pct = int(float64(downloaded) / float64(total) * 100)
		}
		if b.ctx != nil {
			wails.EventsEmit(b.ctx, "model_download_progress", DownloadProgressEvent{
				Downloaded:      downloaded,
				Total:           total,
				Percent:         pct,
				DownloadedHuman: humanize.Bytes(uint64(downloaded)),
				TotalHuman:      humanize.Bytes(uint64(total)),
			})
		}
	}

	path, err := model.DownloadModel(ctx, "", target, progress)
	if err != nil {
		OutLog("DownloadModel err=%v", err)
		return err.Error()
	}
	OutLog("DownloadModel success path=%s", path)
	return ""
}

// CancelModelDownload cancels the ongoing model download
func (b *App) CancelModelDownload() string {
	modelCancelMu.Lock()
	defer modelCancelMu.Unlock()
	if modelCancelCtx != nil {
		modelCancelCtx()
		modelCancelCtx = nil
		return ""
	}
	return "no download in progress"
}

// DeleteLocalModel deletes a local model
func (b *App) DeleteLocalModel(name, title, message string) string {
	if title != "" && b.ctx != nil {
		result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
			Type:          wails.QuestionDialog,
			Title:         title,
			Message:       message,
			Buttons:       []string{"Yes", "No"},
			DefaultButton: "No",
			CancelButton:  "No",
		})
		if err != nil || result == "No" {
			return "cancel"
		}
	}
	if err := model.RemoveModel("", name); err != nil {
		OutLog("DeleteLocalModel err=%v", err)
		return err.Error()
	}
	OutLog("DeleteLocalModel success name=%s", name)
	return ""
}

// DownloadGPULibrary downloads and installs the wgpu-native library
func (b *App) DownloadGPULibrary() string {
	gpuCancelMu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	gpuCancelCtx = cancel
	gpuCancelMu.Unlock()

	defer func() {
		gpuCancelMu.Lock()
		gpuCancelCtx = nil
		gpuCancelMu.Unlock()
	}()

	progress := func(downloaded, total int64) {
		pct := 0
		if total > 0 {
			pct = int(float64(downloaded) / float64(total) * 100)
		}
		if b.ctx != nil {
			wails.EventsEmit(b.ctx, "gpu_download_progress", DownloadProgressEvent{
				Downloaded:      downloaded,
				Total:           total,
				Percent:         pct,
				DownloadedHuman: humanize.Bytes(uint64(downloaded)),
				TotalHuman:      humanize.Bytes(uint64(total)),
			})
		}
	}

	path, err := model.DownloadWGPULibrary(ctx, "", progress)
	if err != nil {
		OutLog("DownloadGPULibrary err=%v", err)
		return err.Error()
	}

	_ = os.Setenv("TENSAI_WGPU_LIB", path)
	OutLog("DownloadGPULibrary installed to %s", path)
	return ""
}

// CancelGPUDownload cancels the ongoing GPU library download
func (b *App) CancelGPUDownload() string {
	gpuCancelMu.Lock()
	defer gpuCancelMu.Unlock()
	if gpuCancelCtx != nil {
		gpuCancelCtx()
		gpuCancelCtx = nil
		return ""
	}
	return "no download in progress"
}
