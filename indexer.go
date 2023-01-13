package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blugelabs/bluge"
	querystr "github.com/blugelabs/query_string"
	"github.com/vjeantet/grok"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type LogIndexer struct {
	config    bluge.Config
	writer    *bluge.Writer
	logBuffer []*LogEnt
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
		b.indexer.logMap = make(map[string]*LogEnt)
	}
	b.wg.Add(1)
	go b.logIndexer()
	return nil
}

// 作業ディレクトリにインデックスがあるか？
func (b *App) HasIndex() bool {
	if b.config.InMemory {
		return b.indexer.writer != nil
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
		if b.config.InMemory {
			b.clearProcessStat()
		}
		return err
	}
	return nil
}

// 作業ディレクトリのインデックスを削除
func (b *App) ClearIndex(title, message string) string {
	result, err := wails.MessageDialog(b.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil || result == "No" {
		return "No"
	}
	b.CloseIndexor()
	b.clearProcessStat()
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
		case l, ok := <-b.logCh:
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
	Total     uint64
	Fields    []string
	Duration  string
	StartTime int64
	EndTime   int64
}

func (b *App) GetIndexInfo() (IndexInfo, error) {
	OutLog("GetIndexInfo")
	ret := IndexInfo{}
	if b.indexer.writer == nil {
		return ret, nil
	}
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
	ret.StartTime = b.processStat.StartTime
	ret.EndTime = b.processStat.EndTime
	return ret, nil
}

type SearchRequest struct {
	Mode       string
	Query      string
	TimeFilter string
	GeoFilter  string
	Anomaly    string
	Vector     string
	Extractor  string
	Limit      int
}

type SearchResult struct {
	Hit        uint64
	Duration   string
	MaxScore   float64
	Logs       []*LogEnt
	Fields     []string
	ExFields   []string
	ErrorMsg   string
	View       string
	AnomalyDur int64
}

func (b *App) SearchLog(r SearchRequest) SearchResult {
	OutLog("SearchLog r=%#v", r)
	view := "timeonly"
	if b.config.Extractor == "auto" {
		view = "auto"
	} else if et, ok := extractorTypes[b.config.Extractor]; ok {
		view = et.View
	}
	ret := SearchResult{
		Logs:     []*LogEnt{},
		View:     view,
		Fields:   []string{},
		ExFields: []string{},
	}
	if b.indexer.writer == nil {
		return ret
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
	query, err := b.makeQuery(r)
	if err != nil {
		OutLog("makeQuery err=%v", err)
		ret.ErrorMsg = err.Error()
		return ret
	}
	OutLog("query=%#+v", query)
	req := bluge.NewTopNSearch(r.Limit, query).WithStandardAggregations().SortBy([]string{"time"})
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
			ret.Fields, _ = reader.Fields()
			if r.Anomaly != "" {
				b.setAnomalyScore(r.Anomaly, r.Vector, &ret)
				ret.Fields = append(ret.Fields, "anomalyScore")
			}
			if r.Extractor != "" {
				b.grokParseLogs(r.Extractor, &ret)
			}
			setFields(&ret)
			return ret
		}
	}
}

// 検索を作る
func (b *App) makeQuery(r SearchRequest) (bluge.Query, error) {
	var q bluge.Query
	var err error
	qs := r.Query
	qs = strings.TrimSpace(qs)
	qo := querystr.DefaultOptions()
	switch r.Mode {
	case "regexp":
		q, err = b.makeRegexpQuery(qs)
	case "full":
		if qs == "" {
			// 空欄は全件検索にする
			qs = "*"
		}
		q, err = querystr.ParseQueryString(qs, qo)
	default:
		q, err = b.makeSimpleQuery(qs)
	}
	if err != nil {
		return nil, err
	}
	if r.TimeFilter != "" {
		OutLog("time filter=%s", r.TimeFilter)
		qt, err := querystr.ParseQueryString(r.TimeFilter, qo)
		if err != nil {
			return nil, err
		}
		q = bluge.NewBooleanQuery().AddMust(qt, q)
	}

	geo := r.GeoFilter
	geo = strings.TrimSpace(geo)
	if geo != "" {
		a := strings.Split(geo, ",")
		if len(a) < 4 {
			return nil, fmt.Errorf("invalid geo format=%s", geo)
		}
		lat, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid geo format=%s err=%v", geo, err)
		}
		long, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid geo format=%s err=%v", geo, err)
		}
		gq := bluge.NewGeoDistanceQuery(long, lat, a[3]).SetField(a[0])
		OutLog("GeoDistanceQuery gq=%#v", gq)
		q = bluge.NewBooleanQuery().AddMust(gq, q)
	}
	return q, nil
}

func (b *App) makeSimpleQuery(qs string) (bluge.Query, error) {
	if qs == "" {
		return bluge.NewMatchAllQuery(), nil
	}
	q := bluge.NewBooleanQuery()
	a := strings.Split(qs, " ")
	for _, s := range a {
		s = strings.TrimSpace(s)
		not := false
		if strings.HasPrefix(s, "!") {
			s = s[1:]
			not = true
		}
		if strings.HasSuffix(s, "*") && len(s) > 1 {
			s = s[:len(s)-1]
			if not {
				q = q.AddMustNot(bluge.NewPrefixQuery(s))
			} else {
				q = q.AddMust(bluge.NewPrefixQuery(s))
			}
		} else {
			if not {
				q = q.AddMustNot(bluge.NewMatchQuery(s))
			} else {
				q = q.AddMust(bluge.NewMatchQuery(s))
			}
		}
	}
	return q, nil
}

func (b *App) makeRegexpQuery(qs string) (bluge.Query, error) {
	if qs == "" {
		return bluge.NewMatchAllQuery(), nil
	}
	q := bluge.NewRegexpQuery(qs)
	if q == nil {
		return nil, fmt.Errorf("invalut regexp query=%s", qs)
	}
	return bluge.NewBooleanQuery().AddMust(q), nil
}

// 検索結果に追加のフィールドを設定する
func setFields(sr *SearchResult) {
	if len(sr.Logs) < 1 {
		return
	}
	fmap := make(map[string]bool)
	for _, f := range sr.Fields {
		// 内部利用のフィールドは除外,geoフィールドも除外
		if !strings.HasPrefix(f, "_") && !strings.HasSuffix(f, "_geo") {
			fmap[f] = true
		}
	}
	for k := range sr.Logs[0].KeyValue {
		if _, ok := fmap[k]; !ok {
			fmap[k] = true
		}
	}
	sr.Fields = []string{}
	for k := range fmap {
		sr.Fields = append(sr.Fields, k)
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
					l.KeyValue[f+"_geo_country"] = e.Country
					l.KeyValue[f+"_geo_city"] = e.City
					l.KeyValue[f+"_geo_latlong"] = fmt.Sprintf("%0.3f,%0.3f", e.Lat, e.Long)
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

func (b *App) grokParseLogs(extractor string, sr *SearchResult) {
	OutLog("start grokParseLogs")
	if sr == nil || len(sr.Logs) < 1 {
		return
	}
	st := time.Now()
	et, ok := extractorTypes[extractor]
	if !ok {
		OutLog("grokParseLogs %s not found", extractor)
		return
	}
	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TWLOGAIAN"] = et.Grok
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		OutLog("%#v err=%v", config, err)
		return
	}
	exFieldMap := make(map[string]bool)
	for _, l := range sr.Logs {
		values, err := g.Parse("%{TWLOGAIAN}", l.All)
		if err != nil {
			OutLog("parseLogEnt err=%v", err)
			continue
		}
		for k, v := range values {
			if k == "TWLOGAIAN" {
				continue
			}
			exFieldMap[k] = true
			if fv, err := strconv.ParseFloat(v, 64); err == nil {
				l.KeyValue[k] = fv
			} else {
				l.KeyValue[k] = v
			}
		}
		if b.config.GeoIP {
			for _, f := range strings.Split(et.IPFields, ",") {
				if ip, ok := l.KeyValue[f]; ok {
					if e := b.findGeo(ip.(string)); e != nil {
						l.KeyValue[f+"_geo"] = e
						l.KeyValue[f+"_geo_country"] = e.Country
						l.KeyValue[f+"_geo_city"] = e.City
						l.KeyValue[f+"_geo_latlong"] = fmt.Sprintf("%0.3f,%0.3f", e.Lat, e.Long)
						exFieldMap[f+"_geo"] = true
						exFieldMap[f+"_geo_country"] = true
						exFieldMap[f+"_geo_city"] = true
						exFieldMap[f+"_geo_latlong"] = true
					} else {
						l.KeyValue[f+"_geo"] = &GeoEnt{}
						l.KeyValue[f+"_geo_country"] = ""
						l.KeyValue[f+"_geo_city"] = ""
						l.KeyValue[f+"_geo_latlong"] = ""
					}
				}
			}
		}
		if b.config.HostName {
			for _, f := range strings.Split(et.IPFields, ",") {
				if ip, ok := l.KeyValue[f]; ok {
					if e := b.findHost(ip.(string)); e != "" {
						l.KeyValue[f+"_host"] = e
						exFieldMap[f+"_host"] = true
					} else {
						l.KeyValue[f+"_host"] = ""
					}
				}
			}
		}
		if b.config.VendorName {
			for _, f := range strings.Split(et.MACFields, ",") {
				if ip, ok := l.KeyValue[f]; ok {
					if e := b.findVendor(ip.(string)); e != "" {
						l.KeyValue[f+"_vendor"] = e
						exFieldMap[f+"_vendor"] = true
					} else {
						l.KeyValue[f+"_vendor"] = ""
					}
				}
			}
		}
	}
	fieldMap := make(map[string]bool)
	for _, f := range sr.Fields {
		fieldMap[f] = true
	}
	for k := range exFieldMap {
		if _, ok := fieldMap[k]; !ok {
			sr.Fields = append(sr.Fields, k)
		}
		sr.ExFields = append(sr.ExFields, k)
	}
	OutLog("end grokParseLogs dur=%v", time.Since(st))
}
