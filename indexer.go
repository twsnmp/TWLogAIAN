package main

import (
	"fmt"
	"os"
	"path/filepath"
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
}

type LogEnt struct {
	ID       string
	Time     time.Time
	Raw      string
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
	timer := time.NewTicker(time.Second * 10)
	b.indexer.logBuffer = []*LogEnt{}
	for {
		select {
		case l, ok := <-b.indexer.logCh:
			if !ok {
				timer.Stop()
				b.addLogToIndex()
				b.process.Done = true
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
			doc.AddField(bluge.NewTextField("raw", l.Raw))
			doc.AddField(bluge.NewDateTimeField("time", l.Time))
			b.indexer.logMap[l.ID] = l
		} else {
			doc.AddField(bluge.NewTextField("raw", l.Raw).StoreValue())
			doc.AddField(bluge.NewDateTimeField("time", l.Time).StoreValue())
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

func (b *App) SearchLog(q string) ([]*LogEnt, error) {
	wails.LogDebug(b.ctx, "SearchLog q="+q)
	ret := []*LogEnt{}
	reader, err := b.indexer.writer.Reader()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	defer func() {
		reader.Close()
	}()
	qo := querystr.DefaultOptions()
	//  TODO:オプションの考える
	query, err := querystr.ParseQueryString(q, qo)
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	c, err := reader.Count()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	f, err := reader.Fields()
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
	wails.LogDebug(b.ctx, fmt.Sprintf("c=%d,f=%v", c, f))
	reader.Count()
	request := bluge.NewTopNSearch(100, query).WithStandardAggregations()
	dmi, err := reader.Search(b.ctx, request)
	if err != nil {
		wails.LogError(b.ctx, err.Error())
		return ret, err
	}
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
							ret = append(ret, l)
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
					case "raw":
						l.Raw = string(value)
					case "time":
						if t, err := bluge.DecodeDateTime(value); err == nil {
							l.Time = t
						}
					}
					return true
				})
				ret = append(ret, &l)
			}
		} else {
			wails.LogDebug(b.ctx, fmt.Sprintf("search ret=%v", ret))
			return ret, nil
		}
	}
}
