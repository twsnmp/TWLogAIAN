package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/vjeantet/grok"

	"github.com/gravwell/gravwell/v3/timegrinder"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProcessStat struct {
	Done     bool
	ErrorMsg string
	LogFiles []*LogFile
}

type ProcessConf struct {
	TimeGrinder *timegrinder.TimeGrinder
	Filter      *regexp.Regexp
	Extractor   *grok.Grok
	TimeField   string
	GeoIP       *geoip2.Reader
	GeoFields   []string
	HostFields  []string
}

type LogFile struct {
	Type     string
	URL      string
	Path     string
	Size     int64
	Read     int64
	Send     int64
	Duration string
}

// Start : インデックス作成を開始する
func (b *App) Start(c Config) string {
	wails.LogDebug(b.ctx, "Start")
	b.config = c
	if e := b.makeLogFileList(); e != "" {
		return e
	}
	if len(b.processStat.LogFiles) < 1 {
		return "処理するファイルがありません"
	}
	if e := b.setupProcess(); e != "" {
		return e
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

// TestSampleLog : サンプルのログをテストしてログの種類を判別する
func (b *App) TestSampleLog(c Config) *ExtractorType {
	max := 0
	var ret *ExtractorType
	for i, e := range extractorTypes {
		s := b.testGrok(c.SampleLog, e.Grok)
		if s > max {
			ret = &extractorTypes[i]
			max = s
		}
	}
	if c.Grok != "" && b.testGrok(c.SampleLog, c.Grok) > max {
		return &ExtractorType{
			Key:  "custom",
			Name: "カスタム設定",
			Grok: c.Grok,
		}
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("testSampleLog  ret=%v", ret))
	return ret
}

func (b *App) setupProcess() string {
	if err := b.setTimeGrinder(); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create new timegrinder err=%v", err))
		return err.Error()
	}
	if err := b.setFilter(); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to get filter err=%v", err))
		return err.Error()
	}
	if err := b.setExtractor(); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to get extractor err=%v", err))
		return err.Error()
	}
	if err := b.setGeoIP(); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to get extractor err=%v", err))
		return err.Error()
	}
	b.setHostFields()
	return ""
}

func (b *App) cleanupProcess() {
	if b.processConf.GeoIP != nil {
		b.processConf.GeoIP.Close()
		b.processConf.GeoIP = nil
	}
}

// Stop : インデクス作成を停止する
func (b *App) Stop() string {
	wails.LogDebug(b.ctx, "Stop")
	b.stopProcess = true
	b.wg.Wait()
	b.cleanupProcess()
	return ""
}

// GetProcessInfo : 処理状態を返す
func (b *App) GetProcessInfo() ProcessStat {
	wails.LogDebug(b.ctx, "GetProcessInfo")
	if b.processStat.Done {
		b.wg.Wait()
		b.cleanupProcess()
	}
	return b.processStat
}

func (b *App) makeLogFileList() string {
	b.processStat.LogFiles = []*LogFile{}
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
	b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
		Type: t,
		URL:  u,
		Path: p,
		Size: s.Size(),
		Read: 0,
		Send: 0,
	})
	return ""
}

func (b *App) logReader() {
	defer func() {
		b.wg.Done()
		close(b.indexer.logCh)
	}()
	wails.LogDebug(b.ctx, "start logReader")
	for _, lf := range b.processStat.LogFiles {
		if b.stopProcess {
			return
		}
		b.readOneLogFile(lf)
	}
	wails.LogDebug(b.ctx, "stop logReader")
}

func (b *App) readOneLogFile(lf *LogFile) {
	st := time.Now()
	wails.LogDebug(b.ctx, "start readOneLogFile path="+lf.Path)
	file, err := os.Open(lf.Path)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create open log file err=%v", err))
		b.processStat.ErrorMsg = err.Error()
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	ln := 0
	var lastTime int64
	for scanner.Scan() {
		l := scanner.Text()
		lf.Read += int64(len(l))
		ln++
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			continue
		}
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%06d", lf.Path, ln),
			KeyValue: make(map[string]interface{}),
			All:      l,
		}
		if b.processConf.Extractor != nil {
			values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				continue
			}
			for k, v := range values {
				if k == "TWLOGAIAN" {
					continue
				}
				// 数値に変換可能な場合は数値として保存
				if fv, err := strconv.ParseFloat(v, 64); err == nil {
					log.KeyValue[k] = fv
				} else {
					log.KeyValue[k] = v
				}
			}
			tfi, ok := log.KeyValue[b.processConf.TimeField]
			if !ok {
				continue
			}
			tf, ok := tfi.(string)
			if !ok {
				continue
			}
			ts, ok, err := b.processConf.TimeGrinder.Extract([]byte(tf))
			if err != nil || !ok {
				continue
			}
			if b.config.GeoIP {
				for _, f := range b.processConf.GeoFields {
					if ip, ok := log.KeyValue[f]; ok {
						if e := b.findGeo(ip.(string)); e != nil {
							log.KeyValue[f+"_geo"] = e
						}
					}
				}
			}
			if b.config.HostName {
				for _, f := range b.processConf.HostFields {
					if ip, ok := log.KeyValue[f]; ok {
						if e := b.findHost(ip.(string)); e != "" {
							log.KeyValue[f+"_host"] = e
						}
					}
				}
			}
			lastTime = ts.UnixNano()
		} else {
			ts, ok, err := b.processConf.TimeGrinder.Extract([]byte(l))
			if err != nil {
				// 複数行は同じタイムスタンプにする
				if lastTime < 1 {
					wails.LogError(b.ctx, fmt.Sprintf("failed to get time stamp err=%v:%s", err, l))
					continue
				}
			} else if ok {
				lastTime = ts.UnixNano()
			} else {
				wails.LogError(b.ctx, fmt.Sprintf("no time stamp: %s", l))
				continue
			}
		}
		log.Time = lastTime
		b.indexer.logCh <- &log
		lf.Send += int64(len(l))
	}
	if err := scanner.Err(); err != nil {
		b.processStat.ErrorMsg = err.Error()
	}
	lf.Duration = time.Since(st).String()
	wails.LogDebug(b.ctx, fmt.Sprintf("end readOneLogFile ln=%d", ln))
}

func (b *App) setTimeGrinder() error {
	var err error
	b.processConf.TimeGrinder, err = timegrinder.New(timegrinder.Config{
		EnableLeftMostSeed: true,
	})
	return err
}

func (b *App) setFilter() error {
	if b.config.Filter == "" {
		return nil
	}
	var err error
	b.processConf.Filter, err = regexp.Compile(b.config.Filter)
	return err
}

func (b *App) testGrok(l, p string) int {
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TWLOGAIAN"] = p
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		return -1
	}
	values, err := g.Parse("%{TWLOGAIAN}", l)
	if err != nil {
		return -1
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("test %s=%d", p, len(values)))
	return len(values)
}

func (b *App) findExtractorType() *ExtractorType {
	for _, e := range extractorTypes {
		if e.Key == b.config.Extractor {
			return &e
		}
	}
	return nil
}

func (b *App) setExtractor() error {
	if b.config.Extractor == "timeonly" || b.config.Extractor == "" {
		b.processConf.Extractor = nil
		return nil
	}
	et := b.findExtractorType()
	if et == nil {
		return fmt.Errorf("invalid extractor type %s", b.processConf.Extractor)
	}
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TWLOGAIAN"] = et.Grok
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		return err
	}
	b.config.GeoFields = et.IPFields
	b.config.HostFields = et.IPFields
	b.processConf.Extractor = g
	b.processConf.TimeField = et.TimeField
	wails.LogDebug(b.ctx, fmt.Sprintf("getExtractor %s=%#v", b.config.Extractor, et))
	return nil
}

func (b *App) setGeoIP() error {
	b.processConf.GeoFields = []string{}
	if !b.config.GeoIP {
		return nil
	}
	for _, f := range strings.Split(b.config.GeoFields, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			b.processConf.GeoFields = append(b.processConf.GeoFields, f)
		}
	}
	if len(b.processConf.GeoFields) < 1 {
		b.config.GeoIP = false
		return nil
	}
	var err error
	b.processConf.GeoIP, err = geoip2.Open(b.config.GeoIPDB)
	return err
}

func (b *App) setHostFields() {
	b.processConf.HostFields = []string{}
	if !b.config.HostName {
		return
	}
	for _, f := range strings.Split(b.config.HostFields, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			b.processConf.HostFields = append(b.processConf.HostFields, f)
		}
	}
}

func (b *App) findGeo(sip string) *GeoEnt {
	if e, ok := b.geoMap[sip]; ok {
		return e
	}
	ip := net.ParseIP(sip)
	if r, err := b.processConf.GeoIP.City(ip); err == nil {
		b.geoMap[sip] = &GeoEnt{
			Lat:     r.Location.Latitude,
			Long:    r.Location.Longitude,
			Country: r.Country.IsoCode,
			City:    r.City.Names["en"],
		}
	} else {
		b.geoMap[sip] = &GeoEnt{
			Lat:     0.0,
			Long:    0.0,
			Country: "",
			City:    "",
		}
	}
	return b.geoMap[sip]
}

func (b *App) findHost(ip string) string {
	if h, ok := b.hostMap[ip]; ok {
		return h
	}
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		b.hostMap[ip] = names[0]
	} else {
		b.hostMap[ip] = ""
	}
	return b.hostMap[ip]
}
