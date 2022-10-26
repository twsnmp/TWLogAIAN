package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/viant/afs/storage"
	"github.com/vjeantet/grok"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.etcd.io/bbolt"
)

type Config struct {
	Filter     string
	Extractor  string
	Grok       string
	TimeField  string
	Recursive  bool
	GeoIP      bool
	GeoIPDB    string
	GeoFields  string
	HostName   bool
	HostFields string
	VendorName bool
	MACFields  string
	InMemory   bool
	SampleLog  string
	ForceUTC   bool
}

type LogSource struct {
	No              int
	Type            string // ログソースの種類
	Server          string
	Path            string
	Pattern         string
	InternalPattern string
	User            string
	Password        string
	SSHKey          string
	scpSvc          storage.Storager
	Start           string
	End             string
	Tag             string
	Host            string
	// for Windows
	Channel string
	Auth    string
}
type WindowSettings struct {
	X     int
	Y     int
	Width int
	Hight int
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
			OutLog("SelectFile err=%v", err)
		}
		return dir
	}
	file, err := wails.OpenFileDialog(b.ctx, wails.OpenDialogOptions{
		Title:           title,
		ShowHiddenFiles: sh,
	})
	if err != nil {
		OutLog("SelectFile err=%v", err)
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
		OutLog("SetWorkDir err=%v", err)
		return fmt.Sprintf("作業フォルダが正しくありません err=%v", err)
	}
	if !fs.IsDir() {
		OutLog("SetWorkDir not dir")
		return "指定した作業フォルダはディレクトリではありません"
	}
	b.logSources = []*LogSource{}
	b.config = Config{}
	makeDefalutLogTypes()
	b.clearProcessStat()
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
	buckets := []string{"settings", "result"}
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
		if v != nil {
			if err := json.Unmarshal(v, &b.logSources); err != nil {
				return err
			}
		}
		v = bkt.Get([]byte("extractorTypes"))
		if len(v) > 1 {
			var list []ExtractorType
			if err := json.Unmarshal(v, &list); err == nil {
				for _, iet := range list {
					if et, ok := extractorTypes[iet.Key]; !ok || et.CanEdit {
						iet.CanEdit = true
						extractorTypes[iet.Key] = iet
					}
				}
			}
		}
		v = bkt.Get([]byte("fieldTypes"))
		if len(v) > 1 {
			var list []FieldType
			if err := json.Unmarshal(v, &list); err == nil {
				for _, ift := range list {
					if ft, ok := fieldTypes[ift.Key]; !ok || ft.CanEdit {
						fieldTypes[ift.Key] = ft
					}
				}
			}
		}
		v = bkt.Get([]byte("window"))
		if v != nil && b.ctx != nil {
			var win WindowSettings
			if err := json.Unmarshal(v, &win); err == nil {
				if win.Hight > 0 && win.Width > 0 {
					wails.WindowSetPosition(b.ctx, win.X, win.Y)
					wails.WindowSetSize(b.ctx, win.Width, win.Hight)
				}
			}
		}
		return b.loadResult(tx)
	})
}

func (b *App) loadResult(tx *bbolt.Tx) error {
	bkt := tx.Bucket([]byte("result"))
	v := bkt.Get([]byte("processStat"))
	if v != nil {
		if err := json.Unmarshal(v, &b.processStat); err != nil {
			OutLog("load processStat err=%v", err)
		}
	}
	v = bkt.Get([]byte("memos"))
	if v != nil {
		if err := json.Unmarshal(v, &b.memos); err != nil {
			OutLog("load memos err=%v", err)
		}
	}
	v = bkt.Get([]byte("history"))
	b.history = []string{}
	if v != nil {
		if err := json.Unmarshal(v, &b.history); err != nil {
			OutLog("load history err=%v", err)
		}
	}
	return nil
}

// saveSettingsToDB : 設定をDBに書き込む
func (b *App) saveSettingsToDB() error {
	cj, err := json.Marshal(b.config)
	if err != nil {
		return err
	}
	lsj, err := json.Marshal(b.logSources)
	if err != nil {
		return err
	}
	etList := []ExtractorType{}
	for _, et := range extractorTypes {
		if et.CanEdit {
			etList = append(etList, et)
		}
	}
	ietj, err := json.Marshal(etList)
	if err != nil {
		return err
	}
	ftList := []FieldType{}
	for _, ft := range fieldTypes {
		if ft.CanEdit {
			ftList = append(ftList, ft)
		}
	}
	iftj, err := json.Marshal(ftList)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))
		if b == nil {
			return fmt.Errorf("bucket settings is nil")
		}
		if err := b.Put([]byte("config"), cj); err != nil {
			return err
		}
		if err := b.Put([]byte("extractorTypes"), ietj); err != nil {
			return err
		}
		if err := b.Put([]byte("fieldTypes"), iftj); err != nil {
			return err
		}
		return b.Put([]byte("logSources"), lsj)
	})
}

func (b *App) saveResultToDB() error {
	jps, err := json.Marshal(b.processStat)
	if err != nil {
		return err
	}
	jmemos, err := json.Marshal(b.memos)
	if err != nil {
		return err
	}
	jh, err := json.Marshal(b.history)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		bkt := tx.Bucket([]byte("result"))
		if bkt == nil {
			return fmt.Errorf("bucket settings is nil")
		}
		if err := bkt.Put([]byte("processStat"), jps); err != nil {
			return err
		}
		bkt.Put([]byte("history"), jh)
		return bkt.Put([]byte("memos"), jmemos)
	})
}

// CloseWorkDir : 作業フォルダを閉じる
func (b *App) CloseWorkDir() string {
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         "分析の終了",
		Message:       "終了しますか?",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return "No"
	}
	b.CloseIndexor()
	b.saveSettingsToDB()
	b.saveResultToDB()
	if b.processConf.GeoIP != nil {
		b.processConf.GeoIP.Close()
		b.processConf.GeoIP = nil
	}
	b.workdir = ""
	if b.db != nil {
		if err := b.saveWindow(); err != nil {
			OutLog("saveWindow err=%v", err)
		}
		b.db.Close()
	}
	return ""
}

// saveWindow : Windoを位置、サイズを保存する
func (b *App) saveWindow() error {
	var win WindowSettings
	if b.ctx != nil {
		win.X, win.Y = wails.WindowGetPosition(b.ctx)
		win.Width, win.Hight = wails.WindowGetSize(b.ctx)
	}
	if win.Hight < 1 {
		return nil
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("settings"))
		OutLog("save window=%v", win)
		if j, err := json.Marshal(&win); err == nil {
			return b.Put([]byte("window"), j)
		}
		return nil
	})
}

// addWorkDirs : 作業ディレクトリ履歴に追加
func (b *App) addWorkDirs(wd string) {
	wds := []string{wd}
	for i, w := range b.appConfig.LastWorkDirs {
		if w != wd {
			wds = append(wds, w)
		}
		if i > 10 {
			break
		}
	}
	b.appConfig.LastWorkDirs = wds
	b.saveAppConfig()
}

// GetConfig : 設定の取得
func (b *App) GetConfig() Config {
	OutLog("GetConfig")
	return b.config
}

// SetConfig : 設定の変更
func (b *App) SetConfig(c Config) string {
	OutLog("SetConfig")
	b.config = c
	return ""
}

// GetLogSources : ログソースリストの取得
func (b *App) GetLogSources() []*LogSource {
	OutLog("GetLogSources")
	return b.logSources
}

// UpdateLogSource : ログソースの更新
func (b *App) UpdateLogSource(ls LogSource) string {
	OutLog("UpdateLogSource")
	if ls.No > 0 && ls.No <= len(b.logSources) {
		// 既存
		b.logSources[ls.No-1] = &ls
		return ""
	}
	// 新規
	ls.No = len(b.logSources) + 1
	b.logSources = append(b.logSources, &ls)
	return ""
}

// DeleteLogSource : ログソースの更新
func (b *App) DeleteLogSource(no int) string {
	OutLog("DeleteLogSource no=%d", no)
	if no <= 0 || no > len(b.logSources) {
		return "送信元がありません"
	}
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         "ログソースの削除",
		Message:       "削除しますか?",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return "No"
	}
	ls := b.logSources
	b.logSources = []*LogSource{}
	n := 1
	for i, e := range ls {
		if i == (no - 1) {
			continue
		}
		e.No = n
		n += 1
		b.logSources = append(b.logSources, e)
	}
	return ""
}

type TestGrokResp struct {
	ErrorMsg string
	Fields   []string
	Data     [][]string
}

// TestGrok : 抽出パターンのテストを行う
func (b *App) TestGrok(p, testData string) TestGrokResp {
	ret := TestGrokResp{
		Fields: []string{},
		Data:   [][]string{},
	}
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TWLOGAIAN"] = p
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		OutLog("TestGrok err=%v", err)
		ret.ErrorMsg = err.Error()
		return ret
	}
	skip := 0
	total := 0
	ln := 0
	for _, l := range strings.Split(testData, "\n") {
		ln++
		if strings.TrimSpace(l) == "" {
			continue
		}
		total++
		values, err := g.Parse("%{TWLOGAIAN}", l)
		if err != nil {
			ret.ErrorMsg = fmt.Sprintf("%d行目のエラー:%v", ln, err)
			break
		} else if len(values) > 0 {
			if len(ret.Fields) < 1 {
				for k := range values {
					ret.Fields = append(ret.Fields, k)
				}
				sort.Strings(ret.Fields)
			}
			ent := []string{}
			for _, k := range ret.Fields {
				ent = append(ent, values[k])
			}
			ret.Data = append(ret.Data, ent)
		} else {
			OutLog("skip=%s", l)
			skip++
		}
	}
	if skip > 0 {
		ret.ErrorMsg = fmt.Sprintf("全%d行中%d行が対象外でした", total, skip)
	}
	return ret
}

type AutoGrokResp struct {
	ErrorMsg string
	Grok     string
}

var grokTestMap = map[string][]string{
	"timestamp": {
		"%{TIMESTAMP_ISO8601:timestamp}",
		"%{HTTPDERROR_DATE:timestamp}",
		"%{HTTPDATE:timestamp}",
		"%{DATESTAMP_EVENTLOG:timestamp}",
		"%{DATESTAMP_RFC2822:timestamp}",
		"%{SYSLOGTIMESTAMP:timestamp}",
		"%{DATESTAMP_OTHER:timestamp}",
		"%{DATESTAMP_RFC822:timestamp}",
	}, // Time
	"ipv4": {
		"%{IPV4:ipv4}",
	}, // IPv4
	"ipv6": {
		"%{IPV6:ipv6}",
	}, // IPv4
	"mac": {
		"%{MAC:mac}",
	},
	"email": {
		"%{EMAILADDRESS:email}",
	},
	"uri": {
		"%{URI:uri}",
	},
}

// AutoGrok : 抽出パターンを自動生成する
func (b *App) AutoGrok(testData string) AutoGrokResp {
	for _, l := range strings.Split(testData, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			testData = l
			break
		}
	}
	ret := AutoGrokResp{}
	replaceMap := make(map[string]string)
	for f, ps := range grokTestMap {
		findGrok(f, testData, ps, replaceMap)
	}
	findSplunkPat(testData, replaceMap)
	if len(replaceMap) < 1 {
		ret.ErrorMsg = "フィールドを検知できません"
		return ret
	}
	ret.Grok = b.makeGrok(testData, replaceMap)
	return ret
}

func findGrok(field, td string, groks []string, rmap map[string]string) {
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	for _, p := range groks {
		config.Patterns["TWLOGAIAN"] = p
		g, err := grok.NewWithConfig(&config)
		if err != nil {
			OutLog("find Grok err=%v", err)
			continue
		}
		values, err := g.Parse("%{TWLOGAIAN}", td)
		if err != nil {
			OutLog("find Grok err=%v", err)
			break
		} else if len(values) > 0 {
			for k, v := range values {
				if k == field && v != "" {
					rmap[v] = p
				}
			}
		}
	}
}

func findSplunkPat(td string, rmap map[string]string) {
	reg := regexp.MustCompile(`([a-zA-Z0-9]+)=(\w+)`)
	for _, m := range reg.FindAllStringSubmatch(td, -1) {
		rmap[m[0]] = fmt.Sprintf("%s=%%{WORD:%s}", m[1], m[1])
		OutLog("rmap %s -> %s", m[0], rmap[m[0]])
	}
}

func (b *App) makeGrok(td string, rmap map[string]string) string {
	r := regexp.QuoteMeta(td)
	for s, d := range rmap {
		r = strings.ReplaceAll(r, regexp.QuoteMeta(s), d)
	}
	return r
}

func (b *App) LoadKeyword() []string {
	ret := []string{}
	file, err := wails.OpenFileDialog(b.ctx, wails.OpenDialogOptions{
		Title: "キーワード",
		Filters: []wails.FileFilter{{
			DisplayName: "キーワードファイル",
			Pattern:     "*.txt",
		}},
	})
	if err != nil {
		OutLog("LoadKeyword err=%v", err)
		return ret
	}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		OutLog("LoadKeyword err=%v", err)
		return ret
	}
	for _, k := range strings.Split(string(buf), "\n") {
		k := strings.TrimSpace(k)
		if k != "" {
			ret = append(ret, k)
		}
	}
	return ret
}

// GetHistory : 履歴の取得
func (b *App) GetHistory() []string {
	OutLog("GetHisttory")
	return b.history
}

// SaveHistory : 履歴の保存
func (b *App) SaveHistory(h []string) {
	OutLog("SaveHistory")
	b.history = h
}
