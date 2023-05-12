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
