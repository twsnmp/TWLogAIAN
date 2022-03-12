package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/0xrawsec/golang-evtx/evtx"
)

func (b *App) readWindowsEvtx(lf *LogFile) error {
	r, err := os.Open(lf.Path)
	if err != nil {
		OutLog("readWindowsEvtx err=%v", err)
		return err
	}
	defer r.Close()
	return b.readWindowsEvtxInt(lf, r)
}

func (b *App) readWindowsEvtxInt(lf *LogFile, r io.ReadSeeker) error {
	defer func() {
		err := recover()
		if err != nil {
			OutLog("readWindowsEvtxInt recover=%v", err)
		}
	}()
	ef, err := evtx.New(r)
	if err == nil {
		err = ef.Header.Verify()
	}
	if err != nil {
		OutLog("readWindowsEvtxInt err=%v", err)
		err = ef.Header.Repair(r)
		if err != nil {
			OutLog("readWindowsEvtxInt err=%v", err)
			return err
		}
	}
	comPath := evtx.Path("/Event/System/Computer")
	levelPath := evtx.Path("/Event/System/Level")
	providerPath := evtx.Path("/Event/System/Provider/Name")
	for e := range ef.FastEvents() {
		if e == nil {
			continue
		}
		l := string(evtx.ToJSON(e))
		len := int64(len(l))
		lf.Read += len
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			continue
		}
		eid, err1 := e.GetInt(&evtx.EventIDPath)
		if err1 != nil {
			eid, err = e.GetInt(&evtx.EventIDPath2)
			if err != nil {
				OutLog("no eventID err1=%v,err2=%v", err1, err)
				continue
			}
		}
		erid, err := e.GetInt(&evtx.EventRecordIDPath)
		if err != nil {
			OutLog("evtx get recordid err=%v", err)
			continue
		}
		t, err := e.GetTime(&evtx.SystemTimePath)
		if err != nil {
			OutLog("evtx gettime err=%v", err)
			continue
		}
		ch, err := e.GetString(&evtx.ChannelPath)
		if err != nil {
			ch = ""
		}
		level, err := e.GetInt(&levelPath)
		if err != nil {
			level = 0
		}
		provider, err := e.GetString(&providerPath)
		if err != nil {
			provider = ""
		}
		com, err := e.GetString(&comPath)
		if err != nil {
			com = ""
		}
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%s:%d", lf.Path, ch, erid),
			KeyValue: make(map[string]interface{}),
			Time:     t.UnixNano(),
			All:      l,
		}
		log.KeyValue["winEventID"] = float64(eid)
		log.KeyValue["winEventRecordID"] = float64(erid)
		log.KeyValue["winChannel"] = ch
		log.KeyValue["winProvider"] = provider
		log.KeyValue["winLevel"] = float64(level)
		log.KeyValue["winComputer"] = com
		if b.processConf.Extractor != nil {
			values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				OutLog("evtx grok err=%v:%s", err, l)
				continue
			}
			for k, v := range values {
				if k == "TWLOGAIAN" {
					continue
				}
				// 数値に変換可能な場合は数値として保存
				if fv, err := strconv.ParseFloat(v, 64); err == nil {
					log.KeyValue[k] = fv
				} else {
					log.KeyValue[k] = v
				}
			}
		}
		if b.config.GeoIP {
			for _, f := range b.processConf.GeoFields {
				if ip, ok := log.KeyValue[f]; ok {
					if e := b.findGeo(ip.(string)); e != nil {
						log.KeyValue[f+"_geo"] = e
					}
				}
			}
		}
		if b.config.HostName {
			for _, f := range b.processConf.HostFields {
				if ip, ok := log.KeyValue[f]; ok {
					if e := b.findHost(ip.(string)); e != "" {
						log.KeyValue[f+"_host"] = e
					}
				}
			}
		}
		if b.config.VendorName {
			for _, f := range b.processConf.MACFields {
				if ip, ok := log.KeyValue[f]; ok {
					if e := b.findVendor(ip.(string)); e != "" {
						log.KeyValue[f+"_vendor"] = e
					}
				}
			}
		}
		b.indexer.logCh <- &log
		lf.Send += len
	}
	return nil
}
