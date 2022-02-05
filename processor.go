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

	"github.com/oschwald/geoip2-golang"
	"github.com/vjeantet/grok"

	"github.com/gravwell/gravwell/v3/timegrinder"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProcessInfo struct {
	Done        bool
	ErrorMsg    string
	LogFiles    []*LogFile
	TimeGrinder *timegrinder.TimeGrinder
	Filter      *regexp.Regexp
	Extractor   *grok.Grok
	TimeFeild   string
	GeoIP       *geoip2.Reader
	GeoFeilds   []string
	HostFeilds  []string
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

func (b *App) setupProcess() string {
	var err error
	b.process.TimeGrinder, err = b.getTimeGrinder()
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create new timegrinder err=%v", err))
		return err.Error()
	}
	b.process.Filter, err = b.getFilter()
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to get filter err=%v", err))
		return err.Error()
	}
	b.process.Extractor, b.process.TimeFeild, err = b.getExtractor()
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to get extractor err=%v", err))
		return err.Error()
	}
	b.process.GeoFeilds = []string{}
	for _, f := range strings.Split(b.config.GeoFeilds, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			b.process.GeoFeilds = append(b.process.GeoFeilds, f)
		}
	}
	b.process.GeoIP, err = b.getGeoIP()
	if err != nil {
		if err != nil {
			wails.LogError(b.ctx, fmt.Sprintf("failed to get extractor err=%v", err))
			return err.Error()
		}
	}
	return ""
}

func (b *App) cleanupProcess() {
	if b.process.GeoIP != nil {
		b.process.GeoIP.Close()
		b.process.GeoIP = nil
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
func (b *App) GetProcessInfo() ProcessInfo {
	wails.LogDebug(b.ctx, "GetProcessInfo")
	if b.process.Done {
		b.wg.Wait()
		b.cleanupProcess()
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
	file, err := os.Open(lf.Path)
	if err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("failed to create open log file err=%v", err))
		b.process.ErrorMsg = err.Error()
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	ln := 0
	skip := 0
	send := 0
	var lastTime int64
	for scanner.Scan() {
		l := scanner.Text()
		lf.Done += int64(len(l))
		ln++
		if b.process.Filter != nil && !b.process.Filter.MatchString(l) {
			continue
		}
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%06d", lf.Path, ln),
			KeyValue: make(map[string]interface{}),
			All:      l,
		}
		if b.process.Extractor != nil {
			values, err := b.process.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				skip++
				continue
			}
			for k, v := range values {
				if k == "TWLOGAIAN" {
					continue
				}
				// 数値に変換可能な場合は数値として保存
				if fv, err := strconv.ParseFloat(v, 64); err == nil {
					wails.LogDebug(b.ctx, fmt.Sprintf("%s=%s %f", k, v, fv))
					log.KeyValue[k] = fv
				} else {
					log.KeyValue[k] = v
				}
			}
			tfi, ok := log.KeyValue[b.process.TimeFeild]
			if !ok {
				skip++
				continue
			}
			tf, ok := tfi.(string)
			if !ok {
				skip++
				continue
			}
			ts, ok, err := b.process.TimeGrinder.Extract([]byte(tf))
			if err != nil || !ok {
				skip++
				continue
			}
			if b.process.GeoIP != nil && len(b.process.GeoFeilds) > 0 {
				for _, f := range b.process.GeoFeilds {
					if ip, ok := log.KeyValue[f]; ok {
						if e := b.findGeo(ip.(string)); e != nil {
							log.KeyValue[f] = e
						}
					}
				}
			}
			if len(b.process.HostFeilds) > 0 {
				for _, f := range b.process.HostFeilds {
					if ip, ok := log.KeyValue[f]; ok {
						if e := b.findHost(ip.(string)); e != "" {
							log.KeyValue[f+"_host"] = e
						}
					}
				}
			}
			lastTime = ts.UnixNano()
		} else {
			ts, ok, err := b.process.TimeGrinder.Extract([]byte(l))
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
		send++
		b.indexer.logCh <- &log
	}
	if err := scanner.Err(); err != nil {
		b.process.ErrorMsg = err.Error()
	}
	if send < 1 || skip > (ln/2) {
		b.process.ErrorMsg = fmt.Sprintf("%s 総数:%d/送信:%d/エラー:%d件", lf.Path, ln, send, skip)
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("end readOneLogFile ln=%d send=%d skip=%d", ln, send, skip))
}

func (b *App) getTimeGrinder() (*timegrinder.TimeGrinder, error) {
	return timegrinder.New(timegrinder.Config{
		EnableLeftMostSeed: true,
	})
}

func (b *App) getFilter() (*regexp.Regexp, error) {
	if b.config.Filter == "" {
		return nil, nil
	}
	return regexp.Compile(b.config.Filter)
}

func (b *App) getExtractor() (*grok.Grok, string, error) {
	if b.config.Extractor == "timeonly" {
		return nil, "", nil
	}
	timeFeild := ""
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	switch b.config.Extractor {
	case "syslog":
		timeFeild = "timestamp"
		config.Patterns["TWLOGAIAN"] = `%{SYSLOGBASE} %{GREEDYDATA:message}`
	}

	g, err := grok.NewWithConfig(&config)
	if err != nil {
		return nil, "", err
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("getExtractor tf=%s p=%s", timeFeild, config.Patterns["TWLOGAIAN"]))
	return g, timeFeild, nil
}

func (b *App) getGeoIP() (*geoip2.Reader, error) {
	if b.config.GeoIPDB == "" {
		return nil, nil
	}
	return geoip2.Open(b.config.GeoIPDB)
}

func (b *App) findGeo(sip string) *GeoEnt {
	if e, ok := b.geoMap[sip]; ok {
		return e
	}
	ip := net.ParseIP(sip)
	if r, err := b.process.GeoIP.City(ip); err == nil {
		return &GeoEnt{
			IP:     sip,
			Lat:    r.Location.Latitude,
			Long:   r.Location.Longitude,
			Contry: r.Country.IsoCode,
			City:   r.City.Names["en"],
		}
	}
	return nil
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
