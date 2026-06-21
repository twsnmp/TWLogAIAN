package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Helper to generate a dummy log line
func generateDummyLogLine(i int) string {
	t := time.Date(2026, 6, 21, 11, 22, 33, 0, time.UTC).Add(time.Second * time.Duration(i))
	return fmt.Sprintf("%s host program[%d]: log message number %d with some extra text to make it longer", t.Format("2006/01/02 15:04:05"), 1000+i%10, i)
}

func TestParallelParsingCorrectness(t *testing.T) {
	// Initialize default extractor types
	makeDefalutLogTypes()

	app := NewApp()
	app.config.Extractor = "syslog"
	app.config.InMemory = true
	app.workdir = t.TempDir()
	app.wg = &sync.WaitGroup{}

	err := app.setExtractor()
	if err != nil {
		t.Fatalf("failed to set extractor: %v", err)
	}
	err = app.setTimeGrinder()
	if err != nil {
		t.Fatalf("failed to set timegrinder: %v", err)
	}

	// Generate 1000 log lines
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString(generateDummyLogLine(i) + "\n")
	}

	lf := &LogFile{
		Name: "dummy.log",
		Path: "dummy.log",
		LogSrc: &LogSource{
			Type: "file",
		},
	}

	// Setup channels and stats like Start does
	app.setupProcess(false)
	app.logCh = make(chan *LogEnt, 1000)

	// Start indexer in goroutine
	if err := app.StartLogIndexer(); err != nil {
		t.Fatalf("failed to start indexer: %v", err)
	}

	// Read logs using parallel readOneLogFile
	app.readOneLogFile(lf, &buf)

	// Close channel and wait for indexer
	close(app.logCh)
	app.wg.Wait()

	// Verify stats
	if app.processStat.ReadLines != 1000 {
		t.Errorf("expected 1000 ReadLines, got %d", app.processStat.ReadLines)
	}
	if app.processStat.SkipLines != 0 {
		t.Errorf("expected 0 SkipLines, got %d", app.processStat.SkipLines)
	}

	// Verify index
	info, err := app.GetIndexInfo()
	if err != nil {
		t.Fatalf("failed to get index info: %v", err)
	}
	if info.Total != 1000 {
		t.Errorf("expected 1000 indexed logs, got %d", info.Total)
	}

	app.CloseIndexor()
}

func BenchmarkLogImport(b *testing.B) {
	makeDefalutLogTypes()

	// Create temporary log file with 10,000 lines
	tempDir, err := os.MkdirTemp("", "twlogaian-bench")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "bench.log")
	f, err := os.Create(logFilePath)
	if err != nil {
		b.Fatalf("failed to create temp log file: %v", err)
	}
	for i := 0; i < 50000; i++ {
		f.WriteString(generateDummyLogLine(i) + "\n")
	}
	f.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app := NewApp()
		app.config.Extractor = "syslog"
		app.config.InMemory = true
		app.workdir = filepath.Join(tempDir, fmt.Sprintf("work-%d", i))
		app.wg = &sync.WaitGroup{}

		err := app.setExtractor()
		if err != nil {
			b.Fatalf("failed to set extractor: %v", err)
		}
		err = app.setTimeGrinder()
		if err != nil {
			b.Fatalf("failed to set timegrinder: %v", err)
		}

		lf := &LogFile{
			Name: "bench.log",
			Path: logFilePath,
			LogSrc: &LogSource{
				Type: "file",
			},
		}

		app.setupProcess(false)
		app.logCh = make(chan *LogEnt, 50000)

		if err := app.StartLogIndexer(); err != nil {
			b.Fatalf("failed to start indexer: %v", err)
		}

		file, err := os.Open(logFilePath)
		if err != nil {
			b.Fatalf("failed to open file: %v", err)
		}

		app.readOneLogFile(lf, file)
		file.Close()

		close(app.logCh)
		app.wg.Wait()
		app.CloseIndexor()
	}
}

func BenchmarkLogImportSequential(b *testing.B) {
	makeDefalutLogTypes()

	// Create temporary log file with 50,000 lines
	tempDir, err := os.MkdirTemp("", "twlogaian-bench-seq")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "bench.log")
	f, err := os.Create(logFilePath)
	if err != nil {
		b.Fatalf("failed to create temp log file: %v", err)
	}
	for i := 0; i < 50000; i++ {
		f.WriteString(generateDummyLogLine(i) + "\n")
	}
	f.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app := NewApp()
		app.config.Extractor = "syslog"
		app.config.InMemory = true
		app.workdir = filepath.Join(tempDir, fmt.Sprintf("work-seq-%d", i))
		app.wg = &sync.WaitGroup{}

		err := app.setExtractor()
		if err != nil {
			b.Fatalf("failed to set extractor: %v", err)
		}
		err = app.setTimeGrinder()
		if err != nil {
			b.Fatalf("failed to set timegrinder: %v", err)
		}

		lf := &LogFile{
			Name: "bench.log",
			Path: logFilePath,
			LogSrc: &LogSource{
				Type: "file",
			},
		}

		app.setupProcess(false)
		app.logCh = make(chan *LogEnt, 50000)

		if err := app.StartLogIndexer(); err != nil {
			b.Fatalf("failed to start indexer: %v", err)
		}

		file, err := os.Open(logFilePath)
		if err != nil {
			b.Fatalf("failed to open file: %v", err)
		}

		app.readOneLogFileSequential(lf, file)
		file.Close()

		close(app.logCh)
		app.wg.Wait()
		app.CloseIndexor()
	}
}

func BenchmarkLogImportWithDNS(b *testing.B) {
	makeDefalutLogTypes()

	tempDir, err := os.MkdirTemp("", "twlogaian-bench-dns")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "bench.log")
	f, err := os.Create(logFilePath)
	if err != nil {
		b.Fatalf("failed to create temp log file: %v", err)
	}
	// We use 100 logs for DNS benchmarks to keep execution time reasonable
	for i := 0; i < 100; i++ {
		// Use unresolvable IP addresses to trigger resolver timeouts
		t := time.Date(2026, 6, 21, 11, 22, 33, 0, time.UTC).Add(time.Second * time.Duration(i))
		line := fmt.Sprintf("%s 192.0.2.%d program[%d]: log message number %d", t.Format("2006/01/02 15:04:05"), i%256, 1000+i%10, i)
		f.WriteString(line + "\n")
	}
	f.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app := NewApp()
		app.config.Extractor = "syslog"
		app.config.InMemory = true
		app.config.HostName = true
		app.config.HostFields = "host"
		app.workdir = filepath.Join(tempDir, fmt.Sprintf("work-dns-%d", i))
		app.wg = &sync.WaitGroup{}

		err := app.setExtractor()
		if err != nil {
			b.Fatalf("failed to set extractor: %v", err)
		}
		err = app.setTimeGrinder()
		if err != nil {
			b.Fatalf("failed to set timegrinder: %v", err)
		}

		lf := &LogFile{
			Name: "bench.log",
			Path: logFilePath,
			LogSrc: &LogSource{
				Type: "file",
			},
		}

		app.setupProcess(false)
		app.logCh = make(chan *LogEnt, 100)

		if err := app.StartLogIndexer(); err != nil {
			b.Fatalf("failed to start indexer: %v", err)
		}

		file, err := os.Open(logFilePath)
		if err != nil {
			b.Fatalf("failed to open file: %v", err)
		}

		app.readOneLogFile(lf, file)
		file.Close()

		close(app.logCh)
		app.wg.Wait()
		app.CloseIndexor()
	}
}

func BenchmarkLogImportWithDNSSequential(b *testing.B) {
	makeDefalutLogTypes()

	tempDir, err := os.MkdirTemp("", "twlogaian-bench-dns-seq")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "bench.log")
	f, err := os.Create(logFilePath)
	if err != nil {
		b.Fatalf("failed to create temp log file: %v", err)
	}
	for i := 0; i < 100; i++ {
		t := time.Date(2026, 6, 21, 11, 22, 33, 0, time.UTC).Add(time.Second * time.Duration(i))
		line := fmt.Sprintf("%s 192.0.2.%d program[%d]: log message number %d", t.Format("2006/01/02 15:04:05"), i%256, 1000+i%10, i)
		f.WriteString(line + "\n")
	}
	f.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app := NewApp()
		app.config.Extractor = "syslog"
		app.config.InMemory = true
		app.config.HostName = true
		app.config.HostFields = "host"
		app.workdir = filepath.Join(tempDir, fmt.Sprintf("work-dns-seq-%d", i))
		app.wg = &sync.WaitGroup{}

		err := app.setExtractor()
		if err != nil {
			b.Fatalf("failed to set extractor: %v", err)
		}
		err = app.setTimeGrinder()
		if err != nil {
			b.Fatalf("failed to set timegrinder: %v", err)
		}

		lf := &LogFile{
			Name: "bench.log",
			Path: logFilePath,
			LogSrc: &LogSource{
				Type: "file",
			},
		}

		app.setupProcess(false)
		app.logCh = make(chan *LogEnt, 100)

		if err := app.StartLogIndexer(); err != nil {
			b.Fatalf("failed to start indexer: %v", err)
		}

		file, err := os.Open(logFilePath)
		if err != nil {
			b.Fatalf("failed to open file: %v", err)
		}

		app.readOneLogFileSequential(lf, file)
		file.Close()

		close(app.logCh)
		app.wg.Wait()
		app.CloseIndexor()
	}
}

// Original sequential implementation for baseline comparison
func (b *App) readOneLogFileSequential(lf *LogFile, reader io.Reader) {
	b.setTimeGrinder()
	if b.processStat.TimeLine == nil {
		b.processStat.TimeLine = make(map[int64]int)
	}
	scanner := bufio.NewScanner(reader)
	autoSetExtractor := b.config.Extractor == "auto"
	if !autoSetExtractor {
		if et, ok := extractorTypes[b.config.Extractor]; ok {
			lf.ETName = et.Name
		}
	}
	ln := 0
	var lastTime int64
	for scanner.Scan() {
		if b.stopProcess {
			return
		}
		l := scanner.Text()
		lf.Read += int64(len(l))
		ln++
		b.processStat.ReadLines++
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			b.processStat.SkipLines++
			continue
		}
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%06d", lf.Path, ln),
			KeyValue: make(map[string]interface{}),
			All:      l,
		}
		if autoSetExtractor {
			lf.ETName = b.autoSetExtractor(l)
			autoSetExtractor = false
		}
		delta := int64(0)
		if b.processConf.Extractor != nil {
			values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				b.processStat.SkipLines++
				continue
			}
			if b.config.Strict && len(values) < 1 {
				b.processStat.SkipLines++
				continue
			}
			for k, v := range values {
				if k == "TWLOGAIAN" {
					continue
				}
				if fv, err := strconv.ParseFloat(v, 64); err == nil {
					log.KeyValue[k] = fv
				} else {
					log.KeyValue[k] = v
				}
			}
			if b.processConf.View == "syslog" {
				if _, ok := log.KeyValue["tag"]; !ok {
					program, _ := log.KeyValue["program"].(string)
					pidStr := ""
					if pid, ok := log.KeyValue["pid"]; ok && pid != "" && pid != nil {
						switch pv := pid.(type) {
						case float64:
							pidStr = fmt.Sprintf("[%.0f]", pv)
						default:
							pidStr = fmt.Sprintf("[%v]", pv)
						}
					}
					if program != "" || pidStr != "" {
						log.KeyValue["tag"] = program + pidStr
					}
				}
			}
			var ts time.Time
			if b.processConf.TimeField != "" {
				tf := ""
				tfi, ok := log.KeyValue[b.processConf.TimeField]
				if !ok {
					if b.config.Strict {
						b.processStat.SkipLines++
						continue
					}
					tf = l
				} else {
					tf, ok = tfi.(string)
					if !ok {
						b.processStat.SkipLines++
						continue
					}
				}
				var ok2 bool
				ts, ok2, err = b.processConf.TimeGrinder.Extract([]byte(tf))
				if err != nil || !ok2 {
					b.processStat.SkipLines++
					continue
				}
			} else {
				var ok2 bool
				ts, ok2, err = b.processConf.TimeGrinder.Extract([]byte(l))
				if err != nil || !ok2 {
					b.processStat.SkipLines++
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
			ts, ok2, err := b.processConf.TimeGrinder.Extract([]byte(l))
			if err != nil {
				if lastTime < 1 {
					b.processStat.SkipLines++
					continue
				}
			} else if ok2 {
				if lastTime > 0 {
					delta = ts.UnixNano() - lastTime
				}
				lastTime = ts.UnixNano()
			} else {
				b.processStat.SkipLines++
				continue
			}
		}
		log.Time = lastTime
		log.KeyValue["delta"] = float64(delta) / (1000.0 * 1000.0 * 1000.0)
		timeH := log.Time / (1000 * 1000 * 1000 * 3600)
		if _, ok := b.processStat.TimeLine[timeH]; !ok {
			b.processStat.TimeLine[timeH] = 0
		}
		b.processStat.TimeLine[timeH]++
		if log.Time < b.processStat.StartTime {
			b.processStat.StartTime = log.Time
		} else if log.Time > b.processStat.EndTime {
			b.processStat.EndTime = log.Time
		}
		b.logCh <- &log
		lf.Send += int64(len(l))
	}
}

