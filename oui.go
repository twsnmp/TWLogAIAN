package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/hex"
	"io"
	"strings"
)

//go:embed conf/oui.csv
var ouiList []byte

func (b *App) setMACFields() {
	b.processConf.MACFields = []string{}
	if !b.config.VendorName {
		return
	}
	for _, f := range strings.Split(b.config.MACFields, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			b.processConf.MACFields = append(b.processConf.MACFields, f)
		}
	}
	b.loadOUIMap()
}

// OUI Map
// https://maclookup.app/downloads/csv-database

// LoadOUIMap : Load OUI Data
func (b *App) loadOUIMap() {
	b.processConf.OuiMap = make(map[string]string)
	f := bytes.NewBuffer(ouiList)
	r := csv.NewReader(f)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if len(record) < 2 {
			continue
		}
		oui := record[0]
		if !strings.Contains(oui, ":") {
			continue
		}
		oui = strings.TrimSpace(oui)
		oui = strings.ReplaceAll(oui, ":", "")
		b.processConf.OuiMap[oui] = record[1]
	}
}

// findVendor : Find Vendor Name from MAC Address
func (b *App) findVendor(mac string) string {
	if mac == "" {
		return ""
	}
	mac = strings.TrimSpace(mac)
	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	if len(mac) > 6 {
		mac = strings.ToUpper(mac)
		if n, ok := b.processConf.OuiMap[mac[:6]]; ok {
			return n
		}
		if h, err := hex.DecodeString(mac); err == nil {
			if (h[0] & 0x02) == 0x02 {
				h[0] = h[0] & 0xfd
				mac = strings.ToUpper(hex.EncodeToString(h))
				if n, ok := b.processConf.OuiMap[mac[:6]]; ok {
					return n + "(Local)"
				}
				return "Local"
			}
		}
	}
	return "Unknown"
}
