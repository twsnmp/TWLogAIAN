package main

import (
	"os"
	"path/filepath"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProcessInfo struct {
	Done     bool
	ErrorMsg string
	LogFiles []LogFile
}

type LogFile struct {
	Type      string
	URL       string
	Path      string
	TimeStamp int64
	Size      int64
	Done      int64
}

// Start : インデックス作成を開始する
func (b *App) Start(c Config) string {
	wails.LogDebug(b.ctx, "Start")
	b.config = c
	if e := b.makeLogFileList(); e != "" {
		return e
	}
	return ""
}

// Stop : インデクス作成を停止する
func (b *App) Stop() string {
	wails.LogDebug(b.ctx, "Stop")
	return ""
}

// GetProcessInfo : 処理状態を返す
func (b *App) GetProcessInfo() ProcessInfo {
	wails.LogDebug(b.ctx, "GetProcessInfo")
	return b.process
}

func (b *App) makeLogFileList() string {
	b.process.LogFiles = []LogFile{}
	for _, s := range b.logSources {
		switch s.Type {
		case "folder":
			if e := b.addLogFolder(s); e != "" {
				return e
			}
		case "file":
			if e := b.addLogFile("file", s.URL, s.URL); e != "" {
				return e
			}
		default:
			return "まだサポートしていません！"
		}
	}
	return ""
}

func (b *App) addLogFolder(s LogSource) string {
	pat := "*"
	if s.Pattern != "" {
		pat = s.Pattern
	}
	files, err := filepath.Glob(filepath.Join(s.URL, pat))
	if err != nil {
		return err.Error()
	}
	for _, f := range files {
		if e := b.addLogFile("folder", f, f); e != "" {
			return e
		}
	}
	return ""
}

func (b *App) addLogFile(t, u, p string) string {
	s, err := os.Stat(p)
	if err != nil {
		return err.Error()
	}
	b.process.LogFiles = append(b.process.LogFiles, LogFile{
		Type:      t,
		URL:       u,
		Path:      p,
		TimeStamp: s.ModTime().Unix(),
		Size:      s.Size(),
		Done:      0,
	})
	return ""
}
