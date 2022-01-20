package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.etcd.io/bbolt"
)

type Config struct {
	GeoIPDBPath   string
	PreFilter     string
	ExtractorType string
	Grok          string
	InMemory      bool
}

type LogFile struct {
	Path      string
	Size      int64
	TimeStamp int64
	Done      int64
}

type LogSource struct {
	ID       string // 識別
	Type     string // ログソースの
	URL      string
	Pattern  string
	User     string
	Password string
	LogFiles map[string]LogFile
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
	b.workdir = ""
	err = b.openDB(wd)
	if err != nil {
		return fmt.Sprintf("データベースを開けません err=%v", err)
	}
	b.addWorkDirs(wd)
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

// GetLogSourceList : ログソースリストの取得
func (b *App) GetLogSourceList() []LogSource {
	wails.LogDebug(b.ctx, "GetLogSourceList")
	return b.logSources
}

// AddLogSource : ログソースの追加
func (b *App) AddLogSource(s LogSource) string {
	wails.LogDebug(b.ctx, "AddLogSource")
	return ""
}
