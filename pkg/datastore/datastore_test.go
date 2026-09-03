package datastore

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testDataStoreImplementation(t *testing.T, ds DataStore, path string) {
	if err := ds.Open(path); err != nil {
		t.Fatalf("failed to open datastore: %v", err)
	}
	defer ds.Close()

	now := time.Now().UnixNano()
	entries := []*LogEntry{
		{
			Time:     now,
			Hash:     "01",
			Line:     1,
			Log:      "first log entry",
			Delta:    1000,
			HasDelta: true,
		},
		{
			Time:     now + 1000,
			Hash:     "01",
			Line:     2,
			Log:      "second log entry",
			Delta:    1000,
			HasDelta: true,
		},
		{
			Time:     now + 2000,
			Hash:     "02",
			Line:     1,
			Log:      "third log entry with error",
			Delta:    1000,
			HasDelta: true,
		},
	}

	if err := ds.SaveLogs(entries); err != nil {
		t.Fatalf("failed to save logs: %v", err)
	}

	count, err := ds.Count()
	if err != nil {
		t.Fatalf("failed to count logs: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 entries, got %d", count)
	}

	var scanned []*LogEntry
	err = ds.ForEach(0, 0, func(e *LogEntry) bool {
		scanned = append(scanned, e)
		return true
	})
	if err != nil {
		t.Fatalf("failed to scan logs: %v", err)
	}
	if len(scanned) != 3 {
		t.Errorf("expected 3 scanned entries, got %d", len(scanned))
	}

	// Range scan
	var rangeScanned []*LogEntry
	err = ds.ForEach(now+500, now+1500, func(e *LogEntry) bool {
		rangeScanned = append(rangeScanned, e)
		return true
	})
	if err != nil {
		t.Fatalf("failed to range scan logs: %v", err)
	}
	if len(rangeScanned) != 1 || rangeScanned[0].Line != 2 {
		t.Errorf("expected 1 entry in range [now+500, now+1500], got %d", len(rangeScanned))
	}
}

func TestParquetDataStore(t *testing.T) {
	tempDir := t.TempDir()
	pqPath := filepath.Join(tempDir, "test.parquet")
	ds := NewParquetDataStore()
	testDataStoreImplementation(t, ds, pqPath)
}

func TestBboltDataStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	ds := NewBboltDataStore()
	testDataStoreImplementation(t, ds, dbPath)
}

func TestBadgerDataStore(t *testing.T) {
	tempDir := t.TempDir()
	badgerPath := filepath.Join(tempDir, "test.badger")
	ds := NewBadgerDataStore()
	testDataStoreImplementation(t, ds, badgerPath)
}

func TestDetectEngineType(t *testing.T) {
	if DetectEngineType("test.parquet") != EngineParquet {
		t.Errorf("expected parquet")
	}
	if DetectEngineType("test.pq") != EngineParquet {
		t.Errorf("expected parquet")
	}
	if DetectEngineType("test.badger") != EngineBadger {
		t.Errorf("expected badger")
	}
	if DetectEngineType("test.db") != EngineBbolt {
		t.Errorf("expected bbolt")
	}

	// Dir check
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "test.parquet"), []byte("test"), 0644)
	if DetectEngineType(tempDir) != EngineParquet {
		t.Errorf("expected dir with parquet to be detected as parquet")
	}
}
