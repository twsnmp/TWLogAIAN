package datastore

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

type BadgerDataStore struct {
	db *badger.DB
}

func NewBadgerDataStore() *BadgerDataStore {
	return &BadgerDataStore{}
}

func (s *BadgerDataStore) Type() EngineType {
	return EngineBadger
}

func (s *BadgerDataStore) Open(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create badger directory: %w", err)
	}
	opts := badger.DefaultOptions(path).
		WithLoggingLevel(badger.WARNING)
	var err error
	s.db, err = badger.Open(opts)
	if err != nil {
		return fmt.Errorf("open badger: %w", err)
	}
	return nil
}

func (s *BadgerDataStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *BadgerDataStore) SaveLogs(entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	wb := s.db.NewWriteBatch()
	defer wb.Cancel()

	for _, e := range entries {
		id := e.ID()
		if err := wb.Set([]byte("l/"+id), []byte(e.Log)); err != nil {
			return err
		}
		if e.HasDelta {
			if err := wb.Set([]byte("d/"+id), []byte(fmt.Sprintf("%d", e.Delta))); err != nil {
				return err
			}
		}
	}
	return wb.Flush()
}

func (s *BadgerDataStore) Count() (int64, error) {
	txn := s.db.NewTransaction(false)
	defer txn.Discard()

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	var count int64
	prefix := []byte("l/")
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		count++
	}
	return count, nil
}

func (s *BadgerDataStore) ForEach(st, et int64, fn ScanCallback) error {
	txn := s.db.NewTransaction(false)
	defer txn.Discard()

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	opts.PrefetchSize = 100
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte("l/")
	seekKey := []byte("l/")
	if st > 0 {
		seekKey = []byte(fmt.Sprintf("l/%016x:", st))
	}

	for it.Seek(seekKey); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		k := string(item.Key())
		id := strings.TrimPrefix(k, "l/")
		t, hash, line, err := ParseID(id)
		if err != nil {
			continue
		}
		if et > 0 && t > et {
			break
		}

		var logVal []byte
		err = item.Value(func(v []byte) error {
			logVal = append([]byte(nil), v...)
			return nil
		})
		if err != nil {
			continue
		}

		entry := &LogEntry{
			Time: t,
			Hash: hash,
			Line: line,
			Log:  string(logVal),
		}

		// check delta
		if dItem, err := txn.Get([]byte("d/" + id)); err == nil {
			_ = dItem.Value(func(v []byte) error {
				if d, err := strconv.ParseInt(string(v), 10, 64); err == nil {
					entry.Delta = d
					entry.HasDelta = true
				}
				return nil
			})
		}

		if !fn(entry) {
			break
		}
	}
	return nil
}
