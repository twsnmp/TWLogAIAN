package main

import (
	"strings"
	"testing"

	"go.etcd.io/bbolt"
)

func FuzzFindSplunkPat(f *testing.F) {
	f.Add(" number=1 ")
	f.Add(" string=hehehe ")
	f.Fuzz(func(t *testing.T, td string) {
		rmap := make(map[string]string)
		findSplunkPat(td, rmap)

		if len(rmap) > 0 {
			for k := range rmap {
				if !strings.Contains(td, k) {
					t.Fatalf("failed td='%s' k='%s' rmap=%+v", td, k, rmap)
				}
			}
		}
	})
}

func FuzzFindGrok(f *testing.F) {
	f.Add(" 192.168.1.1 ")
	f.Add(" http://www.twise.co.jp ")
	f.Add(" 01:02:03:04:05:06 ")
	f.Add(" twsmmp@gmail.com ")
	f.Fuzz(func(t *testing.T, td string) {
		rmap := make(map[string]string)
		for f, ps := range grokTestMap {
			findGrok(f, td, ps, rmap)
		}
		if len(rmap) > 0 {
			for k := range rmap {
				if !strings.Contains(td, k) {
					t.Fatalf("failed td='%s' k='%s' rmap=%+v", td, k, rmap)
				}
			}
		}
	})
}

func TestLegacyWorkDirConfigLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create legacy DB with InMemory = true and NO StorageEngine
	dbPath := tempDir + "/twlogaian.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bbolt: %v", err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return err
		}
		legacyConfigJSON := []byte(`{"Filter":"test","Extractor":"syslog","InMemory":true}`)
		return bkt.Put([]byte("config"), legacyConfigJSON)
	})
	if err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}
	db.Close()

	// Open with SetWorkDir
	app := NewApp()
	res := app.SetWorkDir(tempDir)
	if res != "" {
		t.Fatalf("SetWorkDir failed: %s", res)
	}

	cfg := app.GetConfig()
	if cfg.StorageEngine != "bluge" {
		t.Errorf("expected StorageEngine 'bluge' for legacy DB, got '%s'", cfg.StorageEngine)
	}
	if !cfg.InMemory {
		t.Errorf("expected InMemory to remain true for legacy DB, got false")
	}

	if app.db != nil {
		app.db.Close()
	}
}

