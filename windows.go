package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/0xrawsec/golang-evtx/evtx"
)

func (b *App) IsWindows() bool {
	return runtime.GOOS == "windows"
}

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
		if b.stopProcess {
			return nil
		}
		if e == nil {
			b.processStat.SkipLines++
			continue
		}
		l := string(evtx.ToJSON(e))
		leng := int64(len(l))
		lf.Read += leng
		b.processStat.ReadLines++
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			continue
		}
		eid, err1 := e.GetInt(&evtx.EventIDPath)
		if err1 != nil {
			eid, err = e.GetInt(&evtx.EventIDPath2)
			if err != nil {
				OutLog("no eventID err1=%v,err2=%v", err1, err)
				b.processStat.SkipLines++
				continue
			}
		}
		erid, err := e.GetInt(&evtx.EventRecordIDPath)
		if err != nil {
			OutLog("evtx get recordid err=%v", err)
			b.processStat.SkipLines++
			continue
		}
		t, err := e.GetTime(&evtx.SystemTimePath)
		if err != nil {
			OutLog("evtx gettime err=%v", err)
			b.processStat.SkipLines++
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
		user, err := e.GetString(&evtx.UserIDPath)
		if err != nil {
			user = ""
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
		log.KeyValue["winUserID"] = user
		if b.processConf.Extractor != nil {
			values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
			if err != nil {
				OutLog("evtx grok err=%v:%s", err, l)
				b.processStat.SkipLines++
				continue
			}
			if b.config.Strict && len(values) < 1 {
				b.processStat.SkipLines++
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
		b.processStat.ReadLines++
		timeH := log.Time / (1000 * 1000 * 1000 * 3600)
		if _, ok := b.processStat.TimeLine[timeH]; !ok {
			b.processStat.TimeLine[timeH] = 0
		}
		b.processStat.TimeLine[timeH]++
		if log.Time < b.processStat.StartTime {
			b.processStat.StartTime = log.Time
		} else if log.Time > b.processStat.EndTime {
			b.processStat.EndTime = log.Time
		}
		b.logCh <- &log
		lf.Send += leng
	}
	return nil
}

var reSystem = regexp.MustCompile(`<System.+System>`)

type System struct {
	Provider struct {
		Name string `xml:"Name,attr"`
	}
	EventID       int    `xml:"EventID"`
	Level         int    `xml:"Level"`
	EventRecordID int64  `xml:"EventRecordID"`
	Channel       string `xml:"Channel"`
	Computer      string `xml:"Computer"`
	Security      struct {
		UserID string `xml:"UserID,attr"`
	}
	TimeCreated struct {
		SystemTime string `xml:"SystemTime,attr"`
	}
}

func (b *App) readLogFromWinEventLog(lf *LogFile) error {
	end := time.Now()
	start := end.Add(time.Hour * -1)
	if lf.LogSrc.Start != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.Start+" JST"); err == nil {
			start = t
		}
	}
	if lf.LogSrc.End != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.End+" JST"); err == nil {
			end = t
		}
	}
	filter := fmt.Sprintf(`/q:*[System[TimeCreated[@SystemTime>='%s' and @SystemTime<='%s']]]`, start.UTC().Format("2006-01-02T15:04:05"), end.UTC().Format("2006-01-02T15:04:05"))
	params := []string{"qe", lf.LogSrc.Channel, filter}
	if lf.LogSrc.Server != "" {
		params = append(params, "/r:"+lf.LogSrc.Server)
		params = append(params, "/u:"+lf.LogSrc.User)
		params = append(params, "/p:"+lf.LogSrc.Password)
		if lf.LogSrc.Auth != "" {
			params = append(params, "/a:"+lf.LogSrc.Auth)
		}
	}
	out, err := exec.Command("wevtutil.exe", params...).Output()
	if err != nil {
		OutLog("readLogFromWinEventLog c=%s filter=%s err=%v", lf.LogSrc.Channel, filter, err)
		return err
	}
	if len(out) < 5 {
		OutLog("readLogFromWinEventLog not output")
		return nil
	}
	s := new(System)
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\n", ""), "</Event>") {
		if b.stopProcess {
			return nil
		}
		b.processStat.ReadLines++
		l := strings.TrimSpace(l)
		leng := int64(len(l))
		lf.Read += leng
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			b.processStat.SkipLines++
			continue
		}
		if leng < 10 {
			b.processStat.SkipLines++
			continue
		}
		lsys := reSystem.FindString(l)
		err := xml.Unmarshal([]byte(lsys), s)
		if err != nil {
			OutLog("xml.Unmarshal err=%v", err)
			continue
		}
		t := getEventTime(s.TimeCreated.SystemTime)
		log := LogEnt{
			ID:       fmt.Sprintf("%s:%s:%d", s.Computer, s.Channel, s.EventRecordID),
			KeyValue: make(map[string]interface{}),
			Time:     t.UnixNano(),
			All:      l,
		}
		log.KeyValue["winEventID"] = float64(s.EventID)
		log.KeyValue["winEventRecordID"] = float64(s.EventRecordID)
		log.KeyValue["winChannel"] = s.Channel
		log.KeyValue["winProvider"] = s.Provider.Name
		log.KeyValue["winLevel"] = float64(s.Level)
		log.KeyValue["winComputer"] = s.Computer
		log.KeyValue["winUserID"] = s.Security.UserID
		if err := b.setKeyValuesToLogEnt(l, &log); err != nil {
			b.processStat.SkipLines++
			continue
		}
		timeH := log.Time / (1000 * 1000 * 1000 * 3600)
		if _, ok := b.processStat.TimeLine[timeH]; !ok {
			b.processStat.TimeLine[timeH] = 0
		}
		b.processStat.TimeLine[timeH]++
		if log.Time < b.processStat.StartTime {
			b.processStat.StartTime = log.Time
		} else if log.Time > b.processStat.EndTime {
			b.processStat.EndTime = log.Time
		}
		b.logCh <- &log
		lf.Send += leng
	}
	return nil
}

func getEventTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		OutLog("getEventTime err=%v", err)
		return time.Now()
	}
	return t
}

func (b *App) setKeyValuesToLogEnt(l string, log *LogEnt) error {
	if b.processConf.Extractor != nil {
		values, err := b.processConf.Extractor.Parse("%{TWLOGAIAN}", l)
		if err != nil {
			return nil
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
	return nil
}
