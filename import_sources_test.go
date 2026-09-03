package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseESTimestamp(t *testing.T) {
	// ISO string
	isoStr := "2024-01-15T10:30:00.123456789Z"
	expected, _ := time.Parse(time.RFC3339Nano, isoStr)
	res := parseESTimestamp(isoStr)
	if res != expected.UnixNano() {
		t.Errorf("parseESTimestamp(%s) = %d, expected %d", isoStr, res, expected.UnixNano())
	}

	// Milliseconds as float64
	millisFloat := float64(1705314600123)
	expectedNanos := int64(1705314600123 * 1e6)
	res = parseESTimestamp(millisFloat)
	if res != expectedNanos {
		t.Errorf("parseESTimestamp(%v) = %d, expected %d", millisFloat, res, expectedNanos)
	}

	// Seconds as float64
	secFloat := float64(1705314600)
	expectedNanos = int64(1705314600 * 1e9)
	res = parseESTimestamp(secFloat)
	if res != expectedNanos {
		t.Errorf("parseESTimestamp(%v) = %d, expected %d", secFloat, res, expectedNanos)
	}

	// Nanoseconds as int64
	nanosInt := int64(1705314600123456789)
	res = parseESTimestamp(nanosInt)
	if res != nanosInt {
		t.Errorf("parseESTimestamp(%d) = %d, expected %d", nanosInt, res, nanosInt)
	}
}

func TestGetLogSourceTimeRange(t *testing.T) {
	app := &App{}

	// When Start and End are specified
	src := &LogSource{
		Start: "2024-01-01T00:00",
		End:   "2024-01-02T00:00",
	}
	st, et := app.getLogSourceTimeRange(src)
	tStart, _ := time.Parse("2006-01-02T15:04", "2024-01-01T00:00")
	tEnd, _ := time.Parse("2006-01-02T15:04", "2024-01-02T00:00")
	if st != tStart.UnixNano() || et != tEnd.UnixNano() {
		t.Errorf("getLogSourceTimeRange mismatch: got (%d, %d), expected (%d, %d)", st, et, tStart.UnixNano(), tEnd.UnixNano())
	}

	// When empty, should default to within 24h
	srcEmpty := &LogSource{}
	st2, et2 := app.getLogSourceTimeRange(srcEmpty)
	if st2 <= 0 || et2 <= 0 || et2 < st2 {
		t.Errorf("getLogSourceTimeRange empty fallback invalid: st=%d, et=%d", st2, et2)
	}
}

func TestMakeLogFileListNewSources(t *testing.T) {
	app := &App{
		logSources: []*LogSource{
			{No: 1, Type: "ftp", Server: "192.168.1.1:21", Path: "/logs/*.log"},
			{No: 2, Type: "loki", Server: "http://localhost:3100", Query: `{job="app"}`},
			{No: 3, Type: "es", Server: "http://localhost:9200", Index: "app-*", Query: "*"},
			{No: 4, Type: "imap", Server: "imap.example.com:993", Folder: "INBOX"},
			{No: 5, Type: "twlogeye", Server: "localhost:8081", Target: "logs", SubTarget: "syslog"},
		},
	}
	errStr := app.makeLogFileList()
	if errStr != "" {
		t.Fatalf("makeLogFileList returned error: %s", errStr)
	}
	if len(app.processStat.LogFiles) != 5 {
		t.Fatalf("expected 5 log files, got %d", len(app.processStat.LogFiles))
	}
	if app.processStat.LogFiles[0].LogSrc.Type != "ftp" {
		t.Errorf("expected ftp type, got %s", app.processStat.LogFiles[0].LogSrc.Type)
	}
	if app.processStat.LogFiles[1].LogSrc.Type != "loki" {
		t.Errorf("expected loki type, got %s", app.processStat.LogFiles[1].LogSrc.Type)
	}
	if app.processStat.LogFiles[2].LogSrc.Type != "es" {
		t.Errorf("expected es type, got %s", app.processStat.LogFiles[2].LogSrc.Type)
	}
	if app.processStat.LogFiles[3].LogSrc.Type != "imap" {
		t.Errorf("expected imap type, got %s", app.processStat.LogFiles[3].LogSrc.Type)
	}
	if app.processStat.LogFiles[4].LogSrc.Type != "twlogeye" {
		t.Errorf("expected twlogeye type, got %s", app.processStat.LogFiles[4].LogSrc.Type)
	}
}

func TestEMailFileParsing(t *testing.T) {
	app := &App{
		logCh: make(chan *LogEnt, 100),
		config: Config{
			Extractor: "none",
		},
	}
	app.processConf.TimeGrinder = nil

	rawEMail := "Date: Mon, 15 Jan 2024 10:30:00 +0000\r\nFrom: user@example.com\r\nTo: admin@example.com\r\nSubject: Security Alert\r\n\r\nAlert body"
	lf := &LogFile{
		Name:   "test.eml",
		Path:   "test.eml",
		LogSrc: &LogSource{Type: "file"},
	}

	go func() {
		app.readLogFromEMailFile(lf, strings.NewReader(rawEMail))
		close(app.logCh)
	}()

	count := 0
	for l := range app.logCh {
		count++
		if !strings.Contains(l.All, "Security Alert") {
			t.Errorf("expected email header content, got: %s", l.All)
		}
	}
	if count == 0 {
		t.Errorf("expected at least 1 log entry from email parsing")
	}
}

func TestEVTXFileParsing(t *testing.T) {
	if _, err := os.Stat("testlog/winevent.evtx"); os.IsNotExist(err) {
		t.Skip("testlog/winevent.evtx not found")
	}

	app := &App{
		logCh: make(chan *LogEnt, 1000),
		config: Config{
			Extractor: "none",
		},
	}
	lf := &LogFile{
		Name:   "winevent.evtx",
		Path:   "testlog/winevent.evtx",
		LogSrc: &LogSource{Type: "file"},
	}

	go func() {
		_ = app.readWindowsEvtx(lf)
		close(app.logCh)
	}()

	count := 0
	for l := range app.logCh {
		count++
		if l.Time <= 0 {
			t.Errorf("expected valid timestamp from evtx, got %d", l.Time)
		}
	}
	if count == 0 {
		t.Errorf("expected parsed events from winevent.evtx")
	}
}

func TestMultilineLogParsing(t *testing.T) {
	if _, err := os.Stat("testlog/multi.log"); os.IsNotExist(err) {
		t.Skip("testlog/multi.log not found")
	}

	raw, err := os.ReadFile("testlog/multi.log")
	if err != nil {
		t.Fatal(err)
	}

	// Test with MLStart (matches timestamp at beginning of line)
	app := &App{
		logCh: make(chan *LogEnt, 100),
		config: Config{
			Extractor: "none",
			MLStart:   `^\d{4}/\d{2}/\d{2}`,
		},
	}
	lf := &LogFile{
		Name:   "multi.log",
		Path:   "testlog/multi.log",
		LogSrc: &LogSource{Type: "file"},
	}

	go func() {
		app.readOneLogFile(lf, bytes.NewReader(raw))
		close(app.logCh)
	}()

	var entries []*LogEnt
	for l := range app.logCh {
		entries = append(entries, l)
	}

	if len(entries) == 0 {
		t.Fatalf("expected multiline entries, got 0")
	}
}

func TestFTPErrorHandling(t *testing.T) {
	app := &App{}
	lf := &LogFile{
		Name: "invalid-ftp",
		Path: "127.0.0.1:9999/nonexistent.log",
		LogSrc: &LogSource{
			Type:   "ftp",
			Server: "127.0.0.1:9999",
			Path:   "/nonexistent.log",
		},
	}
	err := app.readLogFromFTP(lf)
	if err == nil {
		t.Errorf("expected error connecting to invalid ftp server, got nil")
	}
}

func TestOpenSearchWithoutTimestamp(t *testing.T) {
	server := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := `{
  "took" : 70,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "my-first-index",
        "_id" : "1",
        "_score" : 1.0,
        "_source":{"title": "OpenSearch Test", "message": "Docker setup completed successfully!"}
      }
    ]
  }
}`
		w.Write([]byte(resp))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	httpServer := &http.Server{Handler: server}
	go httpServer.Serve(listener)
	defer httpServer.Close()

	addr := listener.Addr().String()

	app := &App{
		logCh: make(chan *LogEnt, 100),
		config: Config{
			Extractor: "none",
		},
	}
	lf := &LogFile{
		Name: "opensearch",
		Path: "http://" + addr + "/my-first-index",
		LogSrc: &LogSource{
			Type:      "es",
			Server:    "htt://" + addr, // test scheme normalization
			Index:     "my-first-index",
			TimeField: "@timestamp",
		},
	}

	go func() {
		err := app.readLogFromES(lf)
		if err != nil {
			t.Errorf("readLogFromES err = %v", err)
		}
		close(app.logCh)
	}()

	count := 0
	for l := range app.logCh {
		count++
		if !strings.Contains(l.All, "Docker setup completed successfully!") {
			t.Errorf("expected hit message in log content, got %s", l.All)
		}
		if l.Time <= 0 {
			t.Errorf("expected positive fallback timestamp, got %d", l.Time)
		}
	}
	if count != 1 {
		t.Errorf("expected 1 log hit, got %d", count)
	}
}

func TestTwLogEyeReportJSONParsing(t *testing.T) {
	app := &App{
		logCh: make(chan *LogEnt, 100),
		config: Config{
			Extractor: "none",
		},
	}
	lf := &LogFile{
		Name:   "report.log",
		Path:   "report.log",
		LogSrc: &LogSource{Type: "file"},
	}

	rawLog := `{"Time":"2026-09-03T05:45:00.303143+09:00","Normal":389,"Warn":57,"Error":5,"Patterns":36,"ErrPatterns":1}`

	go func() {
		app.readOneLogFile(lf, strings.NewReader(rawLog))
		close(app.logCh)
	}()

	var entries []*LogEnt
	for l := range app.logCh {
		entries = append(entries, l)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	ent := entries[0]
	if ent.KeyValue["Normal"] != float64(389) {
		t.Errorf("expected Normal 389, got %v", ent.KeyValue["Normal"])
	}
	if ent.KeyValue["Error"] != float64(5) {
		t.Errorf("expected Error 5, got %v", ent.KeyValue["Error"])
	}
}



