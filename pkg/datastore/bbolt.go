package datastore

import (
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/bbolt"
)

type BboltDataStore struct {
	db *bbolt.DB
}

func NewBboltDataStore() *BboltDataStore {
	return &BboltDataStore{}
}

func (s *BboltDataStore) Type() EngineType {
	return EngineBbolt
}

func (s *BboltDataStore) Open(path string) error {
	var err error
	s.db, err = bbolt.Open(path, 0600, &bbolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("logs")); err != nil {
			return fmt.Errorf("create logs bucket: %w", err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("delta")); err != nil {
			return fmt.Errorf("create delta bucket: %w", err)
		}
		return nil
	})
}

func (s *BboltDataStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *BboltDataStore) SaveLogs(entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return s.db.Batch(func(tx *bbolt.Tx) error {
		bl := tx.Bucket([]byte("logs"))
		bd := tx.Bucket([]byte("delta"))
		for _, e := range entries {
			id := []byte(e.ID())
			if err := bl.Put(id, []byte(e.Log)); err != nil {
				return err
			}
			if e.HasDelta {
				if err := bd.Put(id, []byte(fmt.Sprintf("%d", e.Delta))); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *BboltDataStore) Count() (int64, error) {
	var count int64
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		count = int64(b.Stats().KeyN)
		return nil
	})
	return count, err
}

func (s *BboltDataStore) ForEach(st, et int64, fn ScanCallback) error {
	sk := ""
	if st > 0 {
		sk = fmt.Sprintf("%016x:", st)
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		bl := tx.Bucket([]byte("logs"))
		bd := tx.Bucket([]byte("delta"))
		if bl == nil {
			return nil
		}
		c := bl.Cursor()
		var k, v []byte
		if sk != "" {
			k, v = c.Seek([]byte(sk))
		} else {
			k, v = c.First()
		}
		for ; k != nil; k, v = c.Next() {
			t, hash, line, err := ParseID(string(k))
			if err != nil {
				continue
			}
			if et > 0 && t > et {
				break
			}
			entry := &LogEntry{
				Time: t,
				Hash: hash,
				Line: line,
				Log:  string(v),
			}
			if bd != nil {
				if vd := bd.Get(k); vd != nil {
					if d, err := strconv.ParseInt(string(vd), 10, 64); err == nil {
						entry.Delta = d
						entry.HasDelta = true
					}
				}
			}
			if !fn(entry) {
				break
			}
		}
		return nil
	})
}
