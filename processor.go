package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gravwell/gravwell/v3/timegrinder"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProcessInfo struct {
	Done     bool
	ErrorMsg string
	LogFiles []*LogFile
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
	if len(b.process.LogFiles) < 1 {
		return "処理するファイルがありません"
	}
	b.wg = &sync.WaitGroup{}
	b.stopProcess = false
	if err := b.StartLogIndexer(); err != nil {
		return fmt.Sprintf("インデクサーを起動できません。err=%v", err)
	}
	b.saveSettingsToDB()
	b.wg.Add(1)
	go b.logReader()
	return ""
}

// Stop : インデクス作成を停止する
func (b *App) Stop() string {
	wails.LogDebug(b.ctx, "Stop")
	b.stopProcess = true
	b.wg.Wait()
	return ""
}

// GetProcessInfo : 処理状態を返す
func (b *App) GetProcessInfo() ProcessInfo {
	wails.LogDebug(b.ctx, "GetProcessInfo")
	if b.process.Done {
		b.wg.Wait()
	}
	return b.process
}

func (b *App) makeLogFileList() string {
	b.process.LogFiles = []*LogFile{}
	for _, s := range b.logSources {
		switch s.Type {
		case "folder":
			if e := b.addLogFolder(&s); e != "" {
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

func (b *App) addLogFolder(s *LogSource) string {
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
	b.process.LogFiles = append(b.process.LogFiles, &LogFile{
		Type:      t,
		URL:       u,
		Path:      p,
		TimeStamp: s.ModTime().Unix(),
		Size:      s.Size(),
		Done:      0,
	})
	return ""
}

func (b *App) logReader() {
	defer func() {
		b.wg.Done()
		close(b.indexer.logCh)
	}()
	wails.LogDebug(b.ctx, "start logReader")
	for _, lf := range b.process.LogFiles {
		if b.stopProcess {
			return
		}
		b.readOneLogFile(lf)
	}
	wails.LogDebug(b.ctx, "stop logReader")
}

func (b *App) readOneLogFile(lf *LogFile) {
	wails.LogDebug(b.ctx, "start readOneLogFile path="+lf.Path)
	cfg := timegrinder.Config{
		EnableLeftMostSeed: true,
	}
	tg, err := timegrinder.New(cfg)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create new timegrinder err=%v", err))
		b.process.ErrorMsg = err.Error()
		return
	}
	file, err := os.Open(lf.Path)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create open log file err=%v", err))
		b.process.ErrorMsg = err.Error()
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	ln := 0
	for scanner.Scan() {
		l := scanner.Text()
		lf.Done += int64(len(l))
		ln++
		ts, ok, err := tg.Extract([]byte(l))
		if err != nil {
			wails.LogError(b.ctx, fmt.Sprintf("failed to get time stamp err=%v:%s", err, l))
		} else if ok {
			b.indexer.logCh <- &LogEnt{
				ID:   fmt.Sprintf("%s:%06d", lf.Path, ln),
				Time: ts.UnixNano(),
				Raw:  l,
			}
		} else {
			wails.LogError(b.ctx, fmt.Sprintf("no time stamp: %s", l))
		}
	}
	if err := scanner.Err(); err != nil {
		b.process.ErrorMsg = err.Error()
	}
	wails.LogDebug(b.ctx, "stop readOneLogFile")
}
