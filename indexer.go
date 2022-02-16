package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

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
	Score    float64
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
	b.indexer.logCh = make(chan *LogEnt, 1000)
	b.wg.Add(1)
	go b.logIndexer()
	return nil
}

func (b *App) logIndexer() {
	defer b.wg.Done()
	wails.LogDebug(b.ctx, "start logindexer")
	st := time.Now()
	timer := time.NewTicker(time.Millisecond * 200)
	b.indexer.logBuffer = []*LogEnt{}
	for {
		select {
		case l, ok := <-b.indexer.logCh:
			if !ok {
				timer.Stop()
				b.addLogToIndex()
				b.indexer.logBuffer = []*LogEnt{}
				b.processStat.Done = true
				b.indexer.duration = time.Since(st)
				wails.LogDebug(b.ctx, "stop logindexer")
				return
			}
			b.indexer.logBuffer = append(b.indexer.logBuffer, l)
		case <-timer.C:
			if len(b.indexer.logBuffer) > 10000 {
				// Index作成
				b.addLogToIndex()
				b.indexer.logBuffer = []*LogEnt{}
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
	numCount := 0
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
				numCount++
				doc.AddField(bluge.NewNumericField(k, v))
			case *GeoEnt:
				l.KeyValue[k+"_country"] = v.Country
				l.KeyValue[k+"_city"] = v.City
				l.KeyValue[k+"_latlong"] = fmt.Sprintf("%0.3f,%0.3f", v.Lat, v.Long)
				doc.AddField(bluge.NewTextField(k+"_country", v.Country))
				doc.AddField(bluge.NewTextField(k+"_city", v.City))
				doc.AddField(bluge.NewGeoPointField(k+"_latlong", v.Lat, v.Long))
			default:
				// Unknown Type
				wails.LogError(b.ctx, fmt.Sprintf("unknown type %s=%v", k, v))
				continue
			}
		}
		batch.Insert(doc)
		batch_len++
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("batch len=%d %s", batch_len, time.Since(st)))
	if err := b.indexer.writer.Batch(batch); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("error executing batch: %v", err))
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("end batch %s numCount=%d", time.Since(st), numCount))
}

type IndexInfo struct {
	Total    uint64
	Fields   []string
	Duration string
}

func (b *App) GetIndexInfo() (IndexInfo, error) {
	wails.LogDebug(b.ctx, "GetIndexInfo")
	ret := IndexInfo{}
	reader, err := b.indexer.writer.Reader()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	defer func() {
		reader.Close()
	}()
	t, err := reader.Count()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	f, err := reader.Fields()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
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

func (b *App) SearchLog(q string, limit int) SearchResult {
	wails.LogDebug(b.ctx, "SearchLog q="+q)
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
		wails.LogError(b.ctx, err.Error())
		ret.ErrorMsg = err.Error()
		return ret
	}
	defer func() {
		reader.Close()
	}()
	a := strings.SplitN(q, "geo:", 2)
	geo := ""
	if len(a) > 1 {
		q = a[0]
		geo = a[1]
	}
	qo := querystr.DefaultOptions()
	//  TODO:オプションの考える
	query, err := querystr.ParseQueryString(q, qo)
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		ret.ErrorMsg = err.Error()
		return ret
	}
	if geo != "" {
		a = strings.Split(geo, ",")
		if len(a) < 4 {
			ret.ErrorMsg = fmt.Sprintf("invalid geo format=%s", geo)
			return ret
		}
		lat, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			return ret
		}
		long, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			ret.ErrorMsg = err.Error()
			return ret
		}
		gq := bluge.NewGeoDistanceQuery(lat, long, a[2]).SetField(a[0])
		query = bluge.NewBooleanQuery().AddMust(query, gq)
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("query=%#+v", query))
	req := bluge.NewTopNSearch(limit, query).WithStandardAggregations().SortBy([]string{"time"})
	dmi, err := reader.Search(b.ctx, req)
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		ret.ErrorMsg = err.Error()
		return ret
	}
	ret.Hit = dmi.Aggregations().Count()
	ret.MaxScore = dmi.Aggregations().Metric("max_score")
	ret.Duration = dmi.Aggregations().Duration().String()
	for {
		match, err := dmi.Next()
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			ret.ErrorMsg = err.Error()
			return ret
		}
		if match != nil {
			if b.config.InMemory {
				match.VisitStoredFields(func(field string, value []byte) bool {
					if field == "_id" {
						if l, ok := b.indexer.logMap[string(value)]; ok {
							l.Score = match.Score
							ret.Logs = append(ret.Logs, l)
						}
						return false
					}
					return true
				})
			} else {
				l := LogEnt{
					Score: match.Score,
				}
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
					}
					return true
				})
				ret.Logs = append(ret.Logs, &l)
			}
		} else {
			return ret
		}
	}
}
