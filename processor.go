package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/viant/afs/scp"
	"github.com/vjeantet/grok"

	"github.com/gravwell/gravwell/v3/timegrinder"
)

type ProcessStat struct {
	Done        bool
	ErrorMsg    string
	LogFiles    []*LogFile
	IntLogFiles []*LogFile
}

type ProcessConf struct {
	TimeGrinder *timegrinder.TimeGrinder
	Filter      *regexp.Regexp
	Extractor   *grok.Grok
	TimeField   string
	GeoIP       *geoip2.Reader
	GeoFields   []string
	HostFields  []string
	MACFields   []string
	OuiMap      map[string]string
}

type LogFile struct {
	Name     string
	Path     string
	Size     int64
	Read     int64
	Send     int64
	Duration string
	LogSrc   *LogSource
}

// Start : インデックス作成を開始する
func (b *App) Start(c Config, noRead bool) string {
	OutLog("Start")
	b.config = c
	if !noRead {
		if e := b.makeLogFileList(); e != "" {
			OutLog("make log file list err=%s", e)
			return e
		}
		if len(b.processStat.LogFiles) < 1 {
			OutLog("no log files")
			return "処理するファイルがありません"
		}
	}
	if e := b.setupProcess(); e != "" {
		OutLog("make log file list err=%s", e)
		return e
	}
	b.wg = &sync.WaitGroup{}
	b.stopProcess = false
	b.logCh = make(chan *LogEnt, 10000)
	if err := b.StartLogIndexer(); err != nil {
		OutLog("start log indexer err=%v", err)
		return fmt.Sprintf("インデクサーを起動できません。err=%v", err)
	}
	if noRead {
		close(b.logCh)
		b.wg.Wait()
		return ""
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
	return ret
}

func (b *App) setupProcess() string {
	b.processStat.ErrorMsg = ""
	b.processStat.Done = false
	if err := b.setTimeGrinder(); err != nil {
		OutLog("failed to create new timegrinder err=%v", err)
		return err.Error()
	}
	if err := b.setFilter(); err != nil {
		OutLog("failed to get filter err=%v", err)
		return err.Error()
	}
	if err := b.setExtractor(); err != nil {
		OutLog("failed to get extractor err=%v", err)
		return err.Error()
	}
	if err := b.setGeoIP(); err != nil {
		OutLog("failed to get extractor err=%v", err)
		return err.Error()
	}
	b.setHostFields()
	b.setMACFields()
	return ""
}

func (b *App) cleanupProcess() {
	OutLog("cleanupProcess")
	for _, s := range b.logSources {
		if s.scpSvc != nil {
			s.scpSvc.Close()
		}
	}
}

// Stop : インデクス作成を停止する
func (b *App) Stop() string {
	OutLog("Stop")
	b.stopProcess = true
	b.wg.Wait()
	b.cleanupProcess()
	return ""
}

// GetProcessInfo : 処理状態を返す
func (b *App) GetProcessInfo() ProcessStat {
	OutLog("GetProcessInfo")
	if b.processStat.Done {
		b.wg.Wait()
		b.cleanupProcess()
	}
	return b.processStat
}

func (b *App) makeLogFileList() string {
	b.processStat.LogFiles = []*LogFile{}
	b.processStat.IntLogFiles = []*LogFile{}
	for _, s := range b.logSources {
		switch s.Type {
		case "folder":
			if e := b.addLogFolder(s); e != "" {
				return e
			}
		case "file":
			if e := b.addLogFile(s, s.Path); e != "" {
				return e
			}
		case "scp":
			if e := b.addLogFileFromSCP(s); e != "" {
				return e
			}
		case "cmd", "ssh":
			n := s.Path
			if s.Server != "" {
				n = s.Server + ":" + s.Path
			}
			b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
				Name:   n,
				Path:   s.Path,
				Size:   0,
				Read:   0,
				Send:   0,
				LogSrc: s,
			})
		//twsnmp
		case "twsnmp":
			b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
				Name:   s.Server,
				Path:   fmt.Sprintf("%s/?start=%s&end=%s&host=%s&tag=%s&message=%s", s.Server, s.Start, s.End, s.Host, s.Tag, s.Pattern),
				Size:   0,
				Read:   0,
				Send:   0,
				LogSrc: s,
			})
		case "gravwell":
			b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
				Name:   s.Server,
				Path:   fmt.Sprintf("%s/?start=%s&end=%s&query=`%s`", s.Server, s.Start, s.End, s.Pattern),
				Size:   0,
				Read:   0,
				Send:   0,
				LogSrc: s,
			})
		case "windows":
			b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
				Name:   s.Server,
				Path:   fmt.Sprintf("%s/?start=%s&end=%s&channel=`%s`", s.Server, s.Start, s.End, s.Channel),
				Size:   0,
				Read:   0,
				Send:   0,
				LogSrc: s,
			})
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
	files, err := filepath.Glob(filepath.Join(s.Path, pat))
	if err != nil {
		return err.Error()
	}
	for _, f := range files {
		if e := b.addLogFile(s, f); e != "" {
			return e
		}
	}
	return ""
}

func (b *App) addLogFile(src *LogSource, p string) string {
	s, err := os.Stat(p)
	if err != nil {
		return err.Error()
	}
	if s.IsDir() {
		return ""
	}
	n := filepath.Base(p)
	b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
		Name:   n,
		Path:   p,
		Size:   s.Size(),
		Read:   0,
		Send:   0,
		LogSrc: src,
	})
	return ""
}

// getFileNameFilter : ファイル名のパターンチェックする正規表現を返す
func getFileNameFilter(f string) (*regexp.Regexp, error) {
	if f != "" {
		pat := f
		pat = strings.ReplaceAll(pat, "*", ".*")
		pat = strings.ReplaceAll(pat, "?", ".")
		return regexp.Compile(pat)
	}
	return nil, nil
}

func (b *App) addLogFileFromSCP(src *LogSource) string {
	OutLog("start addLogFileFromSCP")
	kpath := src.SSHKey
	if kpath == "" {
		kpath = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}
	auth := scp.NewKeyAuth(kpath, src.User, src.Password)
	provider := scp.NewAuthProvider(auth, nil)
	config, err := provider.ClientConfig()
	if err != nil {
		return err.Error()
	}
	sv := src.Server
	if !strings.Contains(sv, ":") {
		sv += ":22"
	}
	service, err := scp.NewStorager(sv, time.Duration(time.Second)*3, config)
	if err != nil {
		return err.Error()
	}
	OutLog("start service.List")
	files, err := service.List(context.Background(), src.Path)
	if err != nil {
		return err.Error()
	}
	filter, err := getFileNameFilter(src.Pattern)
	if err != nil {
		return err.Error()
	}
	src.scpSvc = service
	for _, file := range files {
		path := file.Name()
		if filter != nil && !filter.MatchString(path) {
			continue
		}
		b.processStat.LogFiles = append(b.processStat.LogFiles, &LogFile{
			Name:   path,
			Path:   src.Path + path,
			Size:   file.Size(),
			Read:   0,
			Send:   0,
			LogSrc: src,
		})
	}
	OutLog("end addLogFileFromSCP")
	return ""
}

func (b *App) logReader() {
	defer func() {
		b.wg.Done()
		close(b.logCh)
	}()
	OutLog("start logReader")
	for _, lf := range b.processStat.LogFiles {
		if b.stopProcess {
			return
		}
		if lf.LogSrc.Type == "cmd" {
			b.readLogFromCommand(lf)
			continue
		}
		if lf.LogSrc.Type == "ssh" {
			b.readLogFromSSH(lf)
			continue
		}
		if lf.LogSrc.Type == "twsnmp" {
			b.readLogFromTWSNMP(lf)
			continue
		}
		if lf.LogSrc.Type == "gravwell" {
			b.readLogFromGravwell(lf)
			continue
		}
		if lf.LogSrc.Type == "windows" {
			b.readLogFromWinEventLog(lf)
			continue
		}
		ext := strings.ToLower(filepath.Ext(lf.Path))
		if ext == ".zip" {
			if err := b.readLogFromZIP(lf); err != nil {
				OutLog("failed to read zip log file err=%v", err)
			}
			continue
		} else if ext == ".evtx" {
			if err := b.readWindowsEvtx(lf); err != nil {
				OutLog("failed to read evtx file err=%v", err)
			}
			continue
		} else if (ext == ".gz" && strings.HasSuffix(lf.Path, "tar.gz")) ||
			ext == ".tgz" ||
			ext == ".bin" {
			if err := b.readLogFromTarGZ(lf); err != nil {
				OutLog("failed to read tar gz log file err=%v", err)
			}
			continue
		}
		file, err := b.openLogFile(lf)
		if err != nil {
			OutLog("failed to open log file err=%v", err)
			b.processStat.ErrorMsg = err.Error()
			continue
		}
		defer file.Close()
		b.readOneLogFile(lf, file)
	}
	b.processStat.LogFiles = append(b.processStat.LogFiles, b.processStat.IntLogFiles...)
	b.processStat.IntLogFiles = []*LogFile{}
	OutLog("stop logReader")
}

func (b *App) readLogFromZIP(lf *LogFile) error {
	r, err := zip.OpenReader(lf.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	filter, err := getFileNameFilter(lf.LogSrc.InternalPattern)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		p := filepath.Base(f.Name)
		if filter != nil && !filter.MatchString(p) {
			continue
		}
		file, err := f.Open()
		if err != nil {
			continue
		}
		ilf := &LogFile{
			Name:   f.Name,
			Path:   lf.Name + "->" + f.Name,
			Size:   int64(f.UncompressedSize64),
			Read:   0,
			Send:   0,
			LogSrc: lf.LogSrc,
		}
		b.processStat.IntLogFiles = append(b.processStat.IntLogFiles, ilf)
		if strings.HasSuffix(p, ".gz") {
			gzr, err := gzip.NewReader(file)
			if err != nil {
				OutLog("read gz log file in ZIP err=%v", err)
				continue
			}
			b.readOneLogFile(ilf, gzr)
		} else if strings.HasSuffix(p, ".evtx") {
			// ReadSeekerが必要なため一度メモリに読み込むため1GBまでにする
			if f.UncompressedSize64 > 1024*1024*1024 {
				OutLog("read evtx file in ZIP size over 1GB")
				continue
			}
			buf, err := ioutil.ReadAll(file)
			if err != nil {
				OutLog("evtx in zip err =%v", err)
				continue
			}
			if err := b.readWindowsEvtxInt(ilf, bytes.NewReader(buf)); err != nil {
				OutLog("read evtx file in ZIP err=%v", err)
				continue
			}
		} else {
			b.readOneLogFile(ilf, file)
		}
		lf.Read += ilf.Read
		lf.Send += ilf.Send
	}
	return nil
}

func (b *App) readLogFromTarGZ(lf *LogFile) error {
	filter, err := getFileNameFilter(lf.LogSrc.InternalPattern)
	if err != nil {
		return err
	}
	r, err := os.Open(lf.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	return b.readLogFromTarGZSub(lf, r, filter, lf.Name)
}

func (b *App) readLogFromTarGZSub(lf *LogFile, r io.Reader, filter *regexp.Regexp, p string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tgzr := tar.NewReader(gzr)
	for {
		f, err := tgzr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if (ext == ".gz" && strings.HasSuffix(lf.Path, "tar.gz")) ||
			ext == ".tgz" {
			if b.config.Recursive {
				// 再帰読み込み
				err := b.readLogFromTarGZSub(lf, tgzr, filter, p+"->"+f.Name)
				if err != nil {
					OutLog("read sub tar gz log file err=%v", err)
				}
			}
			continue
		}
		if filter != nil && !filter.MatchString(f.Name) {
			continue
		}
		ilf := &LogFile{
			Name:   f.Name,
			Path:   p + "->" + f.Name,
			Size:   f.Size,
			Read:   0,
			Send:   0,
			LogSrc: lf.LogSrc,
		}
		b.processStat.IntLogFiles = append(b.processStat.IntLogFiles, ilf)
		if strings.HasSuffix(f.Name, ".gz") {
			gzr, err := gzip.NewReader(tgzr)
			if err != nil {
				continue
			}
			b.readOneLogFile(ilf, gzr)
		} else {
			b.readOneLogFile(ilf, tgzr)
		}
		lf.Read += ilf.Read
		lf.Send += ilf.Send
	}
}

func (b *App) openLogFile(lf *LogFile) (io.ReadCloser, error) {
	var r io.ReadCloser
	var err error
	if lf.LogSrc.Type == "scp" {
		r, err = lf.LogSrc.scpSvc.Open(context.Background(), lf.Path)
	} else {
		r, err = os.Open(lf.Path)
	}
	if err != nil {
		return r, err
	}
	if strings.HasSuffix(lf.Path, ".gz") {
		return gzip.NewReader(r)
	}
	return r, nil
}

func (b *App) readOneLogFile(lf *LogFile, reader io.Reader) {
	st := time.Now()
	OutLog("start readOneLogFile path=%s", lf.Path)
	scanner := bufio.NewScanner(reader)
	ln := 0
	skipF := 0
	var lastTime int64
	for scanner.Scan() {
		if b.stopProcess {
			return
		}
		l := scanner.Text()
		lf.Read += int64(len(l))
		ln++
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			skipF++
			continue
		}
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%06d", lf.Path, ln),
			KeyValue: make(map[string]interface{}),
			All:      l,
		}
		delta := int64(0)
		if b.processConf.Extractor != nil {
			values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				OutLog("grok err=%v:%s", err, l)
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
			var ts time.Time
			if b.processConf.TimeField != "" {
				tfi, ok := log.KeyValue[b.processConf.TimeField]
				if !ok {
					OutLog("no time field '%s' %s", b.processConf.TimeField, l)
					continue
				}
				tf, ok := tfi.(string)
				if !ok {
					continue
				}
				ts, ok, err = b.processConf.TimeGrinder.Extract([]byte(tf))
				if err != nil || !ok {
					OutLog("time parse err=%v:%s", err, l)
					continue
				}
			} else {
				// 日時のフィールドがない場合は全体から取得する
				var ok bool
				ts, ok, err = b.processConf.TimeGrinder.Extract([]byte(l))
				if err != nil || !ok {
					continue
				}
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
			if b.config.VendorName {
				for _, f := range b.processConf.MACFields {
					if ip, ok := log.KeyValue[f]; ok {
						if e := b.findVendor(ip.(string)); e != "" {
							log.KeyValue[f+"_vendor"] = e
						}
					}
				}
			}
			if lastTime > 0 {
				delta = ts.UnixNano() - lastTime
			}
			lastTime = ts.UnixNano()
		} else {
			ts, ok, err := b.processConf.TimeGrinder.Extract([]byte(l))
			if err != nil {
				// 複数行は同じタイムスタンプにする
				if lastTime < 1 {
					OutLog("failed to get time stamp err=%v:%s", err, l)
					continue
				}
			} else if ok {
				if lastTime > 0 {
					delta = ts.UnixNano() - lastTime
				}
				lastTime = ts.UnixNano()
			} else {
				OutLog("no time stamp: %s", l)
				continue
			}
		}
		log.Time = lastTime
		log.KeyValue["delta"] = float64(delta) / (1000.0 * 1000.0 * 1000.0)
		b.logCh <- &log
		lf.Send += int64(len(l))
	}
	if err := scanner.Err(); err != nil {
		b.processStat.ErrorMsg = err.Error()
	}
	lf.Duration = time.Since(st).String()
	OutLog("end readOneLogFile ln=%d skip=%d", ln, skipF)
}

func (b *App) setTimeGrinder() error {
	var err error
	b.processConf.TimeGrinder, err = timegrinder.New(timegrinder.Config{
		EnableLeftMostSeed: true,
	})
	if err == nil && b.processConf.TimeGrinder != nil && !b.config.ForceUTC {
		b.processConf.TimeGrinder.SetLocalTime()
	}
	return err
}

func (b *App) setFilter() error {
	if b.config.Filter == "" {
		b.processConf.Filter = nil
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
	return len(values)
}

func (b *App) findExtractorType(extractor string) *ExtractorType {
	for _, e := range extractorTypes {
		if e.Key == extractor {
			return &e
		}
	}
	for _, e := range b.importedExtractorTypes {
		if e.Key == extractor {
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
	grstr := b.config.Grok
	if b.config.Extractor != "custom" {
		et := b.findExtractorType(b.config.Extractor)
		if et == nil {
			return fmt.Errorf("invalid extractor type %s", b.processConf.Extractor)
		}
		b.config.GeoFields = et.IPFields
		b.config.HostFields = et.IPFields
		b.config.MACFields = et.MACFields
		b.processConf.TimeField = et.TimeField
		grstr = et.Grok
	}
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TWLOGAIAN"] = grstr
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		OutLog("%#v err=%v", config, err)
		return err
	}
	b.processConf.Extractor = g
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
	if b.processConf.GeoIP != nil {
		b.processConf.GeoIP.Close()
		b.processConf.GeoIP = nil
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
	if ip == nil {
		b.geoMap[sip] = &GeoEnt{
			Lat:     0.0,
			Long:    0.0,
			Country: "",
			City:    "",
		}
		return b.geoMap[sip]
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()
	if names, err := net.DefaultResolver.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		b.hostMap[ip] = names[0]
	} else {
		b.hostMap[ip] = ""
	}
	return b.hostMap[ip]
}
