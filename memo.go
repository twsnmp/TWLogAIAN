package main

import "sort"

type Memo struct {
	Time int64
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
	sort.Slice(b.memos, func(i, j int) bool {
		return b.memos[i].Time < b.memos[j].Time
	})
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
	sort.Slice(b.memos, func(i, j int) bool {
		return b.memos[i].Time < b.memos[j].Time
	})
}

// ClearMemos : すべてのメモを削除する
func (b *App) ClearMemos() {
	b.memos = []Memo{}
}
