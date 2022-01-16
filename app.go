package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App application struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (b *App) startup(ctx context.Context) {
	// Perform your setup here
	b.ctx = ctx
}

// domReady is called after the front-end dom has been loaded
func (b *App) domReady(ctx context.Context) {
	// Add your action here
}

// shutdown is called at application termination
func (b *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// GetVersion : バージョンの取得
func (b *App) GetVersion() string {
	return fmt.Sprintf("%s(%s)", version, commit)
}

// OpenURL : ブラウザーで指定してURLを取得する
func (b *App) OpenURL(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("open url=%s err=%v", url, err))
	}
}

// GetWorkDir : 作業フォルダを選択する
func (b *App) GetWorkDir() string {
	dir, err := wails.OpenDirectoryDialog(b.ctx, wails.OpenDialogOptions{
		Title:                "作業フォルダの選択",
		CanCreateDirectories: true,
	})
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("GetWorkDir err=%v", err))
	}
	return dir
}

// GetLastWorkDirs : 作業フォルダを選択する
func (b *App) GetLastWorkDirs() []string {
	return []string{"test1", "test2", "/jsjsjsjjs/shhshs"}
}

// SetWorkDir : 作業フォルダを設定する
func (b *App) SetWorkDir(wd string) bool {
	return true
}

// SendFeedBack :  フィードバックを送信する
func (b *App) SendFeedBack(msg string) bool {
	msg += "\n\n----\nFrom TWLogAIAN"
	values := url.Values{}
	values.Set("msg", msg)
	values.Add("hash", calcHash(msg))

	req, err := http.NewRequest(
		"POST",
		"https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfb.cgi",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("send feedback err=%v", err))
		return false
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("send feedback err=%v", err))
		return false
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("send feedback err=%v", err))
		return false
	}
	if string(r) != "OK" {
		wails.LogError(b.ctx, fmt.Sprintf("send feedback resp=%s", r))
		return false
	}
	return true
}

func calcHash(msg string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(msg + time.Now().Format("2006/01/02T15"))); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
