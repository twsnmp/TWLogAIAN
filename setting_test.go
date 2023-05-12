package main

import (
	"strings"
	"testing"
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
