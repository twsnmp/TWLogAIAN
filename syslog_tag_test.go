package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/twsnmp/TWLogAIAN/pkg/datastore"
)

func TestSyslogTagParsing(t *testing.T) {
	// Initialize default extractor types
	makeDefalutLogTypes()

	app := NewApp()
	app.config.Extractor = "syslog"

	err := app.setExtractor()
	if err != nil {
		t.Fatalf("failed to set extractor: %v", err)
	}

	if app.processConf.View != "syslog" {
		t.Errorf("expected View to be 'syslog', got '%s'", app.processConf.View)
	}

	// Test case 1: with program and pid
	l1 := LogEnt{
		KeyValue: make(map[string]interface{}),
		All:      "Jun  9 10:11:12 host program[1234]: message body",
	}
	app.parseLogEnt(&l1)

	tag1, ok := l1.KeyValue["tag"].(string)
	if !ok {
		t.Errorf("tag is not found or not a string: %+v", l1.KeyValue)
	} else if tag1 != "program[1234]" {
		t.Errorf("expected tag to be 'program[1234]', got '%s'", tag1)
	}

	// Test case 2: with program and no pid
	l2 := LogEnt{
		KeyValue: make(map[string]interface{}),
		All:      "Jun  9 10:11:12 host program: message body",
	}
	app.parseLogEnt(&l2)

	tag2, ok := l2.KeyValue["tag"].(string)
	if !ok {
		t.Errorf("tag is not found or not a string: %+v", l2.KeyValue)
	} else if tag2 != "program" {
		t.Errorf("expected tag to be 'program', got '%s'", tag2)
	}
}

func TestReopenParquetSyslog(t *testing.T) {
	makeDefalutLogTypes()

	tmpDir, err := os.MkdirTemp("", "test_reopen_parquet_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	if errStr := app.SetWorkDir(tmpDir); errStr != "" {
		t.Fatalf("failed to set workdir: %s", errStr)
	}
	app.config.StorageEngine = "parquet"
	app.config.Extractor = "auto"

	logLine := "Dec 10 06:55:46 LabSZ sshd[24200]: Invalid user webmaster from 173.234.31.186"
	detected := app.autoSetExtractor(logLine)
	if detected != "syslog(BSD)" {
		t.Fatalf("expected autoSetExtractor to detect 'syslog(BSD)', got '%s'", detected)
	}
	if app.processConf.View != "syslog" {
		t.Fatalf("expected View to be 'syslog', got '%s'", app.processConf.View)
	}
	if app.config.Extractor != "syslog" {
		t.Fatalf("expected config.Extractor to be updated to 'syslog', got '%s'", app.config.Extractor)
	}

	app.processStat.LogFiles = append(app.processStat.LogFiles, &LogFile{
		Path:   "test.log",
		ETName: detected,
	})

	if errStr := app.Start(app.config, true); errStr != "" {
		t.Fatalf("failed to start log indexer: %s", errStr)
	}

	// Save log entry to Parquet data store
	err = app.ds.SaveLogs([]*datastore.LogEntry{
		{
			Time:  time.Now().UnixNano(),
			Hash:  "00",
			Line:  1,
			Log:   logLine,
			Delta: 0,
		},
	})
	if err != nil {
		t.Fatalf("failed to save logs to parquet: %v", err)
	}

	// Close work dir (simulate user ending analysis)
	_ = app.saveSettingsToDB()
	_ = app.saveResultToDB()
	_ = app.CloseIndexor()
	if app.db != nil {
		_ = app.db.Close()
		app.db = nil
	}

	// Reopen folder with fresh app instance (simulate reopening folder)
	app2 := NewApp()
	if errStr := app2.SetWorkDir(tmpDir); errStr != "" {
		t.Fatalf("failed to reopen workdir: %s", errStr)
	}

	// Simulate user clicking "検索画面へ" button (Start with noRead = true)
	if errStr := app2.Start(app2.config, true); errStr != "" {
		t.Fatalf("failed to start noRead search: %s", errStr)
	}

	// Perform search
	sr := app2.SearchLog(SearchRequest{Limit: 10})
	if sr.ErrorMsg != "" {
		t.Fatalf("search returned error: %s", sr.ErrorMsg)
	}
	if sr.View != "syslog" {
		t.Errorf("expected SearchResult.View to be 'syslog', got '%s'", sr.View)
	}
	if len(sr.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(sr.Logs))
	}

	l := sr.Logs[0]
	tag, _ := l.KeyValue["tag"].(string)
	if tag != "sshd[24200]" {
		t.Errorf("expected tag 'sshd[24200]', got '%v'", l.KeyValue["tag"])
	}
	src, _ := l.KeyValue["logsource"].(string)
	if src != "LabSZ" {
		t.Errorf("expected logsource 'LabSZ', got '%v'", l.KeyValue["logsource"])
	}
	msg, _ := l.KeyValue["message"].(string)
	if !strings.Contains(msg, "Invalid user webmaster") {
		t.Errorf("expected message to contain 'Invalid user webmaster', got '%v'", msg)
	}
}

func TestReopenParquetSyslogFromDataStoreSample(t *testing.T) {
	makeDefalutLogTypes()

	tmpDir, err := os.MkdirTemp("", "test_reopen_sample_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	app := NewApp()
	if errStr := app.SetWorkDir(tmpDir); errStr != "" {
		t.Fatalf("failed to set workdir: %s", errStr)
	}
	app.config.StorageEngine = "parquet"
	app.config.Extractor = "auto"

	logLine := "Dec 10 06:55:46 LabSZ sshd[24200]: Invalid user webmaster from 173.234.31.186"

	if errStr := app.Start(app.config, true); errStr != "" {
		t.Fatalf("failed to start log indexer: %s", errStr)
	}

	// Save log entry to Parquet data store
	err = app.ds.SaveLogs([]*datastore.LogEntry{
		{
			Time:  time.Now().UnixNano(),
			Hash:  "00",
			Line:  1,
			Log:   logLine,
			Delta: 0,
		},
	})
	if err != nil {
		t.Fatalf("failed to save logs to parquet: %v", err)
	}

	// Explicitly keep Extractor as "auto" and empty LogFiles in DB to simulate old/existing DB
	app.config.Extractor = "auto"
	app.processStat.LogFiles = []*LogFile{}
	_ = app.saveSettingsToDB()
	_ = app.saveResultToDB()
	_ = app.CloseIndexor()
	if app.db != nil {
		_ = app.db.Close()
		app.db = nil
	}

	// Reopen folder with fresh app instance
	app2 := NewApp()
	if errStr := app2.SetWorkDir(tmpDir); errStr != "" {
		t.Fatalf("failed to reopen workdir: %s", errStr)
	}

	// Start with noRead = true
	if errStr := app2.Start(app2.config, true); errStr != "" {
		t.Fatalf("failed to start noRead search: %s", errStr)
	}

	// Perform search
	sr := app2.SearchLog(SearchRequest{Limit: 10})
	if sr.ErrorMsg != "" {
		t.Fatalf("search returned error: %s", sr.ErrorMsg)
	}
	if sr.View != "syslog" {
		t.Errorf("expected SearchResult.View to be 'syslog', got '%s'", sr.View)
	}
	if len(sr.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(sr.Logs))
	}

	l := sr.Logs[0]
	tag, _ := l.KeyValue["tag"].(string)
	if tag != "sshd[24200]" {
		t.Errorf("expected tag 'sshd[24200]', got '%v'", l.KeyValue["tag"])
	}
	src, _ := l.KeyValue["logsource"].(string)
	if src != "LabSZ" {
		t.Errorf("expected logsource 'LabSZ', got '%v'", l.KeyValue["logsource"])
	}
}
