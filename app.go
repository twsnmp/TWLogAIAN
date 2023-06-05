package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.etcd.io/bbolt"
)

// Debug Mode
var debug = false

// App application struct
type App struct {
	ctx       context.Context
	appConfig struct {
		DarkMode     bool
		LastWorkDirs []string
	}
	workdir     string
	db          *bbolt.DB
	config      Config
	logSources  []*LogSource
	processStat ProcessStat
	processConf ProcessConf
	indexer     LogIndexer
	logCh       chan *LogEnt
	stopProcess bool
	wg          *sync.WaitGroup
	hostMap     map[string]string
	geoMap      map[string]*GeoEnt
	memos       []Memo
	history     []string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		hostMap: make(map[string]string),
		geoMap:  make(map[string]*GeoEnt),
	}
}

// startup is called at application startup
func (b *App) startup(ctx context.Context) {
	// Perform your setup here
	env := wails.Environment(ctx)
	debug = env.BuildType != "production" || debug
	b.ctx = ctx
	b.loadAppConfig()
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

// loadAppConfig : アプリ設定を読み込み
func (b *App) loadAppConfig() {
	b.appConfig.DarkMode = false
	b.appConfig.LastWorkDirs = []string{}
	conf := b.getConfigName()
	if conf == "" {
		return
	}
	j, err := os.ReadFile(conf)
	if err != nil {
		OutLog("loadAppConfig err=%v", err)
		return
	}
	json.Unmarshal(j, &b.appConfig)
}

// saveAppConfig : アプリ設定の保存
func (b *App) saveAppConfig() {
	conf := b.getConfigName()
	if conf == "" {
		return
	}
	j, err := json.Marshal(&b.appConfig)
	if err != nil {
		OutLog("saveAppConfig err=%v", err)
		return
	}
	os.WriteFile(conf, j, 0600)
}

// getConfigName : 設定ファイル名の取得
func (b *App) getConfigName() string {
	c, err := os.UserConfigDir()
	if err != nil {
		OutLog("getConfigName err=%v", err)
		return ""
	}
	return path.Join(c, "twlogaian.conf")
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
		OutLog("open url=%s err=%v", url, err)
	}
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
		OutLog("send feedback err=%v", err)
		return false
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		OutLog("send feedback err=%v", err)
		return false
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		OutLog("send feedback err=%v", err)
		return false
	}
	if string(r) != "OK" {
		OutLog("send feedback resp=%s", r)
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

// GetDark :  ダークモードの取得
func (b *App) GetDark() bool {
	return b.appConfig.DarkMode
}

// SetDark :  ダークモードの切り替え
func (b *App) SetDark(dark bool) {
	if b.appConfig.DarkMode == dark {
		return
	}
	b.appConfig.DarkMode = dark
	b.saveAppConfig()
}

func OutLog(format string, v ...interface{}) {
	if debug {
		log.Printf(format, v...)
	}
}
