package main

import (
	"sort"
	"time"
)

type Memo struct {
	Time int64
	Diff string
	Memo string //
	Type string // "","error","warn","normal"
	Log  string
}

// AddMemo : メモを追加する
func (b *App) AddMemo(memo Memo) {
	for _, m := range b.memos {
		if m.Log == memo.Log && m.Time == memo.Time {
			return
		}
	}
	b.memos = append(b.memos, memo)
	b.sortAndCalcDiffMemo()
}

// SetMemo : メモを更新する
func (b *App) SetMemo(memo Memo) {
	for i, m := range b.memos {
		if m.Log == memo.Log && m.Time == memo.Time {
			b.memos[i].Memo = memo.Memo
			b.memos[i].Type = memo.Type
			return
		}
	}
}

// GetMemos : メモのリストを取得する
func (b *App) GetMemos() []Memo {
	return b.memos
}

// DeleteMemo : メモを削除する
func (b *App) DeleteMemo(memo Memo) {
	tmp := []Memo{}
	for _, m := range b.memos {
		if m.Log == memo.Log && m.Time == memo.Time {
			continue
		}
		tmp = append(tmp, m)
	}
	b.memos = tmp
	b.sortAndCalcDiffMemo()
}

// sortAndCalcDiffMemo : メモを時刻順に並べて時差を再計算する
func (b *App) sortAndCalcDiffMemo() {
	if len(b.memos) > 1 {
		sort.Slice(b.memos, func(i, j int) bool {
			return b.memos[i].Time < b.memos[j].Time
		})
	}
	if len(b.memos) > 0 {
		b.memos[0].Diff = ""
		for i := 1; i < len(b.memos); i++ {
			ts := time.Unix(0, b.memos[i-1].Time)
			te := time.Unix(0, b.memos[i].Time)
			b.memos[i].Diff = te.Sub(ts).String()
		}
	}
}

// ClearMemos : すべてのメモを削除する
func (b *App) ClearMemos() {
	b.memos = []Memo{}
}
