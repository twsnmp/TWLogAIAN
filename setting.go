package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.etcd.io/bbolt"
)

type Config struct {
	GeoIPDB   string
	Filter    string
	Extractor string
	Grok      string
	InMemory  bool
}

type LogSource struct {
	No       int
	ID       int64
	Type     string // ログソースの種類
	URL      string
	Pattern  string
	AuthType string
	User     string
	Password string
}

// SelectFile : ファイル/フォルダを選択する
func (b *App) SelectFile(t string) string {
	title := ""
	dir := false
	sh := false
	switch t {
	case "work":
		dir = true
		title = "作業フォルダ"
	case "geoip":
		dir = false
		title = "GeoIPデータベース"
	case "sshkey":
		dir = false
		title = "SSHキー"
		sh = true
	case "logdir":
		dir = true
		title = "ログフォルダ"
	case "logfile":
		dir = false
		title = "ログファイル"
	}
	if dir {
		dir, err := wails.OpenDirectoryDialog(b.ctx, wails.OpenDialogOptions{
			Title:                title,
			CanCreateDirectories: true,
		})
		if err != nil {
			wails.LogError(b.ctx, fmt.Sprintf("SelectFile err=%v", err))
		}
		return dir
	}
	file, err := wails.OpenFileDialog(b.ctx, wails.OpenDialogOptions{
		Title:           title,
		ShowHiddenFiles: sh,
	})
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("SelectFile err=%v", err))
	}
	return file
}

// GetLastWorkDirs : 作業フォルダを選択する
func (b *App) GetLastWorkDirs() []string {
	return b.appConfig.LastWorkDirs
}

// SetWorkDir : 作業フォルダを設定する
func (b *App) SetWorkDir(wd string) string {
	fs, err := os.Stat(wd)
	if err != nil {
		return fmt.Sprintf("作業フォルダが正しくありません err=%v", err)
	}
	if !fs.IsDir() {
		return "指定した作業フォルダはディレクトリではありません"
	}
	err = b.openDB(wd)
	if err != nil {
		return fmt.Sprintf("データベースを開けません err=%v", err)
	}
	b.addWorkDirs(wd)
	b.workdir = wd
	return ""
}

// openDB : データベースをオープンして設定を読み込む
func (b *App) openDB(wd string) error {
	var err error
	path := filepath.Join(wd, "twlogaian.db")
	b.db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	if err = b.initDB(); err != nil {
		b.db.Close()
		b.db = nil
		return err
	}
	if err = b.loadSettingsFromDB(); err != nil {
		b.db.Close()
		b.db = nil
		return err
	}
	return nil
}

// initDB : DBにバケットを作る
func (b *App) initDB() error {
	buckets := []string{"settings", "results", "reports"}
	return b.db.Update(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// loadSettingsFromDB : 設定をDBから読み込む
func (b *App) loadSettingsFromDB() error {
	return b.db.View(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte("settings"))
		v := bkt.Get([]byte("config"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &b.config); err != nil {
			return err
		}
		v = bkt.Get([]byte("logSources"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &b.logSources); err != nil {
			return err
		}
		return nil
	})
}

// CloseWorkDir : 作業フォルダを設定する
func (b *App) CloseWorkDir() string {
	b.workdir = ""
	if b.db != nil {
		b.db.Close()
	}
	return ""
}

// addWorkDirs : 作業ディレクトリ履歴に追加
func (b *App) addWorkDirs(wd string) {
	wds := []string{wd}
	for _, w := range b.appConfig.LastWorkDirs {
		if w != wd {
			wds = append(wds, w)
		}
	}
	b.appConfig.LastWorkDirs = wds
	b.saveAppConfig()
}

// GetConfig : 設定の取得
func (b *App) GetConfig() Config {
	wails.LogDebug(b.ctx, "GetConfig")
	return b.config
}

// SetConfig : 設定の変更
func (b *App) SetConfig(c Config) string {
	wails.LogDebug(b.ctx, "SetConfig")
	b.config = c
	return ""
}

// GetLogSources : ログソースリストの取得
func (b *App) GetLogSources() []LogSource {
	wails.LogDebug(b.ctx, "GetLogSources")
	return b.logSources
}

// UpdateLogSource : ログソースの更新
func (b *App) UpdateLogSource(ls LogSource) string {
	wails.LogDebug(b.ctx, "UpdateLogSource")
	if ls.No > 0 && ls.No <= len(b.logSources) {
		// 既存
		b.logSources[ls.No-1] = ls
		return ""
	}
	// 新規
	ls.ID = time.Now().UnixNano()
	ls.No = len(b.logSources) + 1
	b.logSources = append(b.logSources, ls)
	return ""
}

// DeleteLogSource : ログソースの更新
func (b *App) DeleteLogSource(no int) string {
	wails.LogDebug(b.ctx, fmt.Sprintf("DeleteLogSource no=%d", no))
	if no <= 0 || no > len(b.logSources) {
		return "送信元がありません"
	}
	ls := b.logSources
	b.logSources = []LogSource{}
	n := 1
	for i, e := range ls {
		if i == (no - 1) {
			// TODO:ここで関連するログファイルを削除する
			continue
		}
		e.No = n
		n += 1
		b.logSources = append(b.logSources, e)
	}
	return ""
}
