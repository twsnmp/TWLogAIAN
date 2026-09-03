package datastore

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/parquet-go/parquet-go"
)

type ParquetLogRecord struct {
	Time     int64  `parquet:"time,snappy"`
	Hash     string `parquet:"hash,dict,snappy"`
	Line     int32  `parquet:"line,snappy"`
	Log      string `parquet:"log,zstd"`
	Delta    int64  `parquet:"delta,snappy"`
	HasDelta bool   `parquet:"has_delta,snappy"`
}

type ParquetDataStore struct {
	path  string
	isDir bool
}

func NewParquetDataStore() *ParquetDataStore {
	return &ParquetDataStore{}
}

func (s *ParquetDataStore) Type() EngineType {
	return EngineParquet
}

func (s *ParquetDataStore) Open(path string) error {
	s.path = path
	fi, err := os.Stat(path)
	if err == nil && !fi.IsDir() {
		// Existing single file
		s.isDir = false
		return nil
	}
	// Default to directory structure for multi-part writes
	s.isDir = true
	if err := os.MkdirAll(s.path, 0755); err != nil {
		return fmt.Errorf("create parquet directory: %w", err)
	}
	return nil
}

func (s *ParquetDataStore) Close() error {
	return nil
}

func (s *ParquetDataStore) SaveLogs(entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	records := make([]ParquetLogRecord, len(entries))
	for i, e := range entries {
		records[i] = ParquetLogRecord{
			Time:     e.Time,
			Hash:     e.Hash,
			Line:     int32(e.Line),
			Log:      e.Log,
			Delta:    e.Delta,
			HasDelta: e.HasDelta,
		}
	}

	var randBytes [4]byte
	_, _ = rand.Read(randBytes[:])
	randVal := binary.BigEndian.Uint32(randBytes[:])

	var filePath string
	if s.isDir {
		fileName := fmt.Sprintf("part_%016x_%08x.parquet", time.Now().UnixNano(), randVal)
		filePath = filepath.Join(s.path, fileName)
	} else {
		filePath = s.path
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create parquet file: %w", err)
	}
	defer file.Close()

	writer := parquet.NewGenericWriter[ParquetLogRecord](file)
	if _, err := writer.Write(records); err != nil {
		return fmt.Errorf("write parquet records: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close parquet writer: %w", err)
	}

	return nil
}

func (s *ParquetDataStore) getFileList() ([]string, error) {
	if !s.isDir {
		return []string{s.path}, nil
	}
	files, err := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	if err != nil {
		return nil, err
	}
	pqFiles, _ := filepath.Glob(filepath.Join(s.path, "*.pq"))
	files = append(files, pqFiles...)
	sort.Strings(files)
	return files, nil
}

func (s *ParquetDataStore) Count() (int64, error) {
	files, err := s.getFileList()
	if err != nil {
		return 0, err
	}
	var total int64
	for _, fpath := range files {
		f, err := os.Open(fpath)
		if err != nil {
			continue
		}
		fi, err := f.Stat()
		if err != nil || fi.Size() == 0 {
			f.Close()
			continue
		}
		pf, err := parquet.OpenFile(f, fi.Size())
		if err != nil {
			f.Close()
			continue
		}
		total += pf.NumRows()
		f.Close()
	}
	return total, nil
}

func (s *ParquetDataStore) ForEach(st, et int64, fn ScanCallback) error {
	files, err := s.getFileList()
	if err != nil {
		return err
	}

	for _, filePath := range files {
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		fi, err := f.Stat()
		if err != nil || fi.Size() == 0 {
			f.Close()
			continue
		}

		pf, err := parquet.OpenFile(f, fi.Size())
		if err != nil {
			f.Close()
			continue
		}

		reader := parquet.NewGenericReader[ParquetLogRecord](pf)
		buf := make([]ParquetLogRecord, 1024)
		stop := false

		for {
			n, err := reader.Read(buf)
			if n > 0 {
				for i := 0; i < n; i++ {
					rec := &buf[i]
					if st > 0 && rec.Time < st {
						continue
					}
					if et > 0 && rec.Time > et {
						continue
					}
					entry := &LogEntry{
						Time:     rec.Time,
						Hash:     rec.Hash,
						Line:     int(rec.Line),
						Log:      rec.Log,
						Delta:    rec.Delta,
						HasDelta: rec.HasDelta,
					}
					if !fn(entry) {
						stop = true
						break
					}
				}
			}
			if stop || err != nil {
				if err != nil && err != io.EOF {
					// reading error
				}
				break
			}
		}
		reader.Close()
		f.Close()

		if stop {
			break
		}
	}
	return nil
}
