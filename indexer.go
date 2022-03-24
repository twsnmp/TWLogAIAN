package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/blugelabs/bluge"
	querystr "github.com/blugelabs/query_string"
)

type LogIndexer struct {
	config    bluge.Config
	writer    *bluge.Writer
	logBuffer []*LogEnt
	logCh     chan *LogEnt
	logMap    map[string]*LogEnt
	duration  time.Duration
}

type LogEnt struct {
	ID       string
	Time     int64
	All      string
	KeyValue map[string]interface{}
}

type GeoEnt struct {
	Lat     float64
	Long    float64
	Country string
	City    string
}

func (b *App) StartLogIndexer() error {
	var err error
	if b.indexer.writer == nil {
		if b.config.InMemory {
			b.indexer.config = bluge.InMemoryOnlyConfig()
		} else {
			dir := filepath.Join(b.workdir, "bluge")
			if err := os.MkdirAll(dir, 0777); err != nil {
				return err
			}
			b.indexer.config = bluge.DefaultConfig(dir)
		}
		b.indexer.writer, err = bluge.OpenWriter(b.indexer.config)
		if err != nil {
			return err
		}
	}
	b.indexer.logMap = make(map[string]*LogEnt)
	b.indexer.logCh = make(chan *LogEnt, 10000)
	b.wg.Add(1)
	go b.logIndexer()
	return nil
}

// 作業ディレクトリにインデックスがあるか？
func (b *App) HasIndex() bool {
	if b.config.InMemory {
		return false
	}
	if st, err := os.Stat(filepath.Join(b.workdir, "bluge")); err == nil && st.IsDir() {
		return true
	}
	return false
}

func (b *App) CloseIndexor() error {
	if b.indexer.writer != nil {
		err := b.indexer.writer.Close()
		b.indexer.writer = nil
		return err
	}
	return nil
}

// 作業ディレクトリにインデックスがあるか？
func (b *App) ClearIndex() string {
	if b.config.InMemory {
		return ""
	}
	if st, err := os.Stat(filepath.Join(b.workdir, "bluge")); err != nil && !st.IsDir() {
		return ""
	}
	if err := os.RemoveAll(filepath.Join(b.workdir, "bluge")); err != nil {
		return err.Error()
	}
	return ""
}

func (b *App) logIndexer() {
	defer b.wg.Done()
	OutLog("start logindexer")
	st := time.Now()
	timer := time.NewTicker(time.Millisecond * 200)
	b.indexer.logBuffer = []*LogEnt{}
	bFirstLog := true
	skip := 0
	total := 0
	for {
		select {
		case l, ok := <-b.indexer.logCh:
			if !ok {
				timer.Stop()
				if bFirstLog && len(b.indexer.logBuffer) > 0 {
					bFirstLog = false
					b.setFieldTypes(b.indexer.logBuffer[0])
				}
				b.addLogToIndex()
				b.indexer.logBuffer = []*LogEnt{}
				b.processStat.Done = true
				b.indexer.duration = time.Since(st)
				OutLog("stop logindexer")
				return
			}
			b.indexer.logBuffer = append(b.indexer.logBuffer, l)
		case <-timer.C:
			if len(b.indexer.logBuffer) > 10000 {
				if bFirstLog {
					bFirstLog = false
					b.setFieldTypes(b.indexer.logBuffer[0])
				}
				// Index作成
				b.addLogToIndex()
				total += len(b.indexer.logBuffer)
				b.indexer.logBuffer = []*LogEnt{}
				OutLog("total=%d skip=%d", total, skip)
			} else {
				skip++
			}
		}
	}
}

func (b *App) addLogToIndex() {
	st := time.Now()
	if len(b.indexer.logBuffer) < 1 {
		return
	}
	batch_len := 0
	batch := bluge.NewBatch()
	for _, l := range b.indexer.logBuffer {
		doc := bluge.NewDocument(l.ID)
		if b.config.InMemory {
			doc.AddField(bluge.NewTextField("_all", l.All))
			doc.AddField(bluge.NewDateTimeField("time", time.Unix(0, l.Time)))
			b.indexer.logMap[l.ID] = l
		} else {
			doc.AddField(bluge.NewTextField("_all", l.All).StoreValue())
			doc.AddField(bluge.NewDateTimeField("time", time.Unix(0, l.Time)).StoreValue())
		}
		for k, i := range l.KeyValue {
			switch v := i.(type) {
			case string:
				doc.AddField(bluge.NewTextField(k, v))
			case float64:
				if k == "delta" {
					doc.AddField(bluge.NewNumericField(k, v).StoreValue())
				} else {
					doc.AddField(bluge.NewNumericField(k, v))
				}
			case *GeoEnt:
				l.KeyValue[k+"_country"] = v.Country
				l.KeyValue[k+"_city"] = v.City
				l.KeyValue[k+"_latlong"] = fmt.Sprintf("%0.3f,%0.3f", v.Lat, v.Long)
				doc.AddField(bluge.NewTextField(k+"_country", v.Country))
				doc.AddField(bluge.NewTextField(k+"_city", v.City))
				doc.AddField(bluge.NewGeoPointField(k+"_latlong", v.Long, v.Lat))
			default:
				// Unknown Type
				OutLog("unknown type %s=%v", k, v)
				continue
			}
		}
		batch.Insert(doc)
		batch_len++
	}
	OutLog("batch len=%d %s", batch_len, time.Since(st))
	if err := b.indexer.writer.Batch(batch); err != nil {
		OutLog("error executing batch: %v", err)
	}
	batch.Reset()
	OutLog("end batch %s", time.Since(st))
}

type IndexInfo struct {
	Total    uint64
	Fields   []string
	Duration string
}

func (b *App) GetIndexInfo() (IndexInfo, error) {
	OutLog("GetIndexInfo")
	ret := IndexInfo{}
	reader, err := b.indexer.writer.Reader()
	if err != nil {
		OutLog("GetIndexInfo err=%v", err)
		return ret, err
	}
	defer func() {
		reader.Close()
	}()
	t, err := reader.Count()
	if err != nil {
		OutLog("GetIndexInfo err=%v", err)
		return ret, err
	}
	f, err := reader.Fields()
	if err != nil {
		OutLog("GetIndexInfo err=%v", err)
		return ret, err
	}
	f = append(f, "score")
	ret.Duration = b.indexer.duration.String()
	ret.Total = t
	ret.Fields = f
	return ret, nil
}

type SearchResult struct {
	Hit      uint64
	Duration string
	MaxScore float64
	Logs     []*LogEnt
	ErrorMsg string
	View     string
}

var regGeo = regexp.MustCompile(`\s*geo:(\S+)`)

func (b *App) SearchLog(q string, limit int) SearchResult {
	OutLog("SearchLog q=%#v", q)
	view := "timeonly"
	if et := b.findExtractorType(); et != nil {
		view = et.View
	}
	ret := SearchResult{
		Logs: []*LogEnt{},
		View: view,
	}
	reader, err := b.indexer.writer.Reader()
	if err != nil {
		OutLog("SearchLog err=%v", err)
		ret.ErrorMsg = err.Error()
		return ret
	}
	if ret.View == "timeonly" {
		if fields, err := reader.Fields(); err == nil {
			for _, f := range fields {
				if f == "winEventID" {
					ret.View = "windows"
				}
			}
		}
	}
	defer func() {
		reader.Close()
	}()
	geo := ""
	if gl := regGeo.FindAllStringSubmatch(q, -1); len(gl) > 0 {
		if len(gl) != 1 && len(gl[0]) != 2 {
			ret.ErrorMsg = fmt.Sprintf("位置検索条件が正しくありません len=%d", len(gl))
			return ret
		}
		q = strings.ReplaceAll(q, gl[0][0], "")
		geo = gl[0][1]
		OutLog("geo=%s", geo)
	}
	q = strings.TrimSpace(q)
	if q == "" {
		// 空欄は全件検索にする
		q = "*"
	}
	qo := querystr.DefaultOptions()
	query, err := querystr.ParseQueryString(q, qo)
	if err != nil {
		OutLog("SearchLog err=%v", err)
		ret.ErrorMsg = err.Error()
		return ret
	}
	if geo != "" {
		a := strings.Split(geo, ",")
		if len(a) < 4 {
			OutLog("invalid geo formar=%s", geo)
			ret.ErrorMsg = fmt.Sprintf("位置検索条件が正しくありません%s", geo)
			return ret
		}
		lat, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			OutLog("SearchLog err=%v", err)
			ret.ErrorMsg = fmt.Sprintf("位置検索条件が正しくありません err=%v", err)
			return ret
		}
		long, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			OutLog("SearchLog err=%v", err)
			ret.ErrorMsg = fmt.Sprintf("位置検索条件が正しくありません err=%v", err)
			return ret
		}
		gq := bluge.NewGeoDistanceQuery(long, lat, a[3]).SetField(a[0])
		OutLog("GeoDistanceQuery err=%#v", gq)
		query = bluge.NewBooleanQuery().AddMust(gq, query)
	}

	OutLog("query=%#+v", query)
	req := bluge.NewTopNSearch(limit, query).WithStandardAggregations().SortBy([]string{"time"})
	dmi, err := reader.Search(b.ctx, req)
	if err != nil {
		OutLog("SearchLog err=%v", err)
		ret.ErrorMsg = err.Error()
		return ret
	}
	ret.Hit = dmi.Aggregations().Count()
	ret.MaxScore = dmi.Aggregations().Metric("max_score")
	ret.Duration = dmi.Aggregations().Duration().String()
	for {
		match, err := dmi.Next()
		if err != nil {
			OutLog("SearchLog err=%v", err)
			ret.ErrorMsg = err.Error()
			return ret
		}
		if match != nil {
			if b.config.InMemory {
				match.VisitStoredFields(func(field string, value []byte) bool {
					if field == "_id" {
						if l, ok := b.indexer.logMap[string(value)]; ok {
							l.KeyValue["score"] = match.Score
							ret.Logs = append(ret.Logs, l)
						}
						return false
					}
					return true
				})
			} else {
				l := LogEnt{
					KeyValue: make(map[string]interface{}),
				}
				l.KeyValue["score"] = match.Score
				match.VisitStoredFields(func(field string, value []byte) bool {
					switch field {
					case "_id":
						l.ID = string(value)
					case "_all":
						l.All = string(value)
					case "time":
						if t, err := bluge.DecodeDateTime(value); err == nil {
							l.Time = t.UnixNano()
						}
					case "delta":
						if f, err := bluge.DecodeNumericFloat64(value); err == nil {
							l.KeyValue["delta"] = f
						}
					}
					return true
				})
				b.parseLogEnt(&l)
				ret.Logs = append(ret.Logs, &l)
			}
		} else {
			return ret
		}
	}
}

func (b *App) parseLogEnt(l *LogEnt) {
	if b.processConf.Extractor == nil {
		return
	}
	values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l.All)
	if err != nil {
		OutLog("parseLogEnt err=%v", err)
	}
	for k, v := range values {
		if k == "TWLOGAIAN" {
			continue
		}
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			l.KeyValue[k] = fv
		} else {
			l.KeyValue[k] = v
		}
	}
	if b.config.GeoIP {
		for _, f := range b.processConf.GeoFields {
			if ip, ok := l.KeyValue[f]; ok {
				if e := b.findGeo(ip.(string)); e != nil {
					l.KeyValue[f+"_geo"] = e
					l.KeyValue[f+"_country"] = e.Country
					l.KeyValue[f+"_city"] = e.City
					l.KeyValue[f+"_latlong"] = fmt.Sprintf("%0.3f,%0.3f", e.Lat, e.Long)
				}
			}
		}
	}
	if b.config.HostName {
		for _, f := range b.processConf.HostFields {
			if ip, ok := l.KeyValue[f]; ok {
				if e := b.findHost(ip.(string)); e != "" {
					l.KeyValue[f+"_host"] = e
				}
			}
		}
	}
	if b.config.VendorName {
		for _, f := range b.processConf.MACFields {
			if ip, ok := l.KeyValue[f]; ok {
				if e := b.findVendor(ip.(string)); e != "" {
					l.KeyValue[f+"_vendor"] = e
				}
			}
		}
	}
}
