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

type LatLong struct {
	Lat  float64
	Long float64
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
	timer := time.NewTicker(time.Second * 10)
	b.indexer.logBuffer = []*LogEnt{}
	for {
		select {
		case l, ok := <-b.indexer.logCh:
			if !ok {
				timer.Stop()
				b.addLogToIndex()
				b.process.Done = true
				b.indexer.duration = time.Since(st)
				wails.LogDebug(b.ctx, "stop logindexer")
				return
			}
			b.indexer.logBuffer = append(b.indexer.logBuffer, l)
		case <-timer.C:
			if len(b.indexer.logBuffer) > 0 {
				// Index作成
				b.addLogToIndex()
			}
		}
	}
}

func (b *App) addLogToIndex() {
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
				doc.AddField(bluge.NewNumericField(k, v))
			case LatLong:
				doc.AddField(bluge.NewGeoPointField(k, v.Lat, v.Long))
			default:
				// Unknown Type
				wails.LogError(b.ctx, fmt.Sprintf("unnown type %s=%v", k, v))
				continue
			}
		}
		batch.Insert(doc)
		batch_len++
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("batch len=%d", batch_len))
	if err := b.indexer.writer.Batch(batch); err != nil {
		wails.LogError(b.ctx, fmt.Sprintf("error executing batch: %v", err))
	}
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
}

func (b *App) SearchLog(q string, limit int) (SearchResult, error) {
	wails.LogDebug(b.ctx, "SearchLog q="+q)
	ret := SearchResult{
		Logs: []*LogEnt{},
	}
	reader, err := b.indexer.writer.Reader()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
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
		return ret, err
	}
	if geo != "" {
		a = strings.Split(geo, ",")
		if len(a) < 4 {
			return ret, fmt.Errorf("invalid geo format=%s", geo)
		}
		lat, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			return ret, err
		}
		long, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			return ret, err
		}
		gq := bluge.NewGeoDistanceQuery(lat, long, a[2]).SetField(a[0])
		query = bluge.NewBooleanQuery().AddMust(query, gq)
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("query=%#+v", query))
	req := bluge.NewTopNSearch(limit, query).WithStandardAggregations().SortBy([]string{"time"})
	dmi, err := reader.Search(b.ctx, req)
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	ret.Hit = dmi.Aggregations().Count()
	ret.MaxScore = dmi.Aggregations().Metric("max_score")
	ret.Duration = dmi.Aggregations().Duration().String()
	for {
		match, err := dmi.Next()
		if err != nil {
			wails.LogError(b.ctx, err.Error())
			return ret, err
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
			return ret, nil
		}
	}
}
