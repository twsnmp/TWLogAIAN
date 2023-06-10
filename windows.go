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
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
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
	lf.ETName = "Windows"
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
		if l, ok := winSecLevelMap[int(eid)]; ok {
			log.KeyValue["winLevel"] = float64(l)
		} else {
			log.KeyValue["winLevel"] = float64(level)
		}
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
	tz := time.Now().Format("MST")
	if lf.LogSrc.Start != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.Start+" "+tz); err == nil {
			start = t
		}
	}
	if lf.LogSrc.End != "" {
		if t, err := time.Parse("2006-01-02T15:04 MST", lf.LogSrc.End+" "+tz); err == nil {
			end = t
		}
	}
	OutLog("tz=%s start=%v end=%v", tz, start, end)
	lf.ETName = "Windows"
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
	st := time.Now()
	tr := japanese.ShiftJIS.NewDecoder()
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\n", ""), "</Event>") {
		if b.stopProcess {
			return nil
		}
		b.processStat.ReadLines++
		l = strings.TrimSpace(l)
		if lf.LogSrc.ShiftJIS {
			l, _, err = transform.String(tr, l)
			if err != nil {
				OutLog("shift-jis to utf8 error err=%v", err)
				continue
			}
		}
		leng := int64(len(l))
		lf.Read += leng
		lf.Size += leng
		if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(l) {
			b.processStat.SkipLines++
			continue
		}
		if leng < 10 {
			b.processStat.SkipLines++
			continue
		}
		lsys := reSystem.FindString(l)
		err = xml.Unmarshal([]byte(lsys), s)
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
		if l, ok := winSecLevelMap[int(s.EventID)]; ok {
			log.KeyValue["winLevel"] = float64(l)
		} else {
			log.KeyValue["winLevel"] = float64(s.Level)
		}
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
	lf.Duration = time.Since(st).String()
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

// https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/plan/appendix-l--events-to-monitor
var winSecLevelMap = map[int]int{
	4618: 100, 4649: 100, 4719: 100, 4765: 100, 4766: 100, 4794: 100, 4897: 100, 4964: 100, 5124: 100, 1102: 100,
	4621: 101, 4675: 101, 4692: 101, 4693: 101, 4706: 101, 4713: 101, 4714: 101, 4715: 101, 4716: 101, 4724: 101,
	4727: 101, 4735: 101, 4737: 101, 4739: 101, 4754: 101, 4755: 101, 4764: 101, 4780: 101, 4816: 101,
	4865: 101, 4866: 101, 4867: 101, 4868: 101, 4870: 101, 4882: 101, 4885: 101, 4890: 101, 4892: 101, 4896: 101,
	4906: 101, 4907: 101, 4908: 101, 4912: 101, 4960: 101, 4961: 101, 4962: 101, 4963: 101, 4965: 101, 4976: 101,
	4977: 101, 4978: 101, 4983: 101, 4984: 101, 5027: 101, 5028: 101, 5029: 101, 5030: 101, 5035: 101, 5037: 101,
	5038: 101, 5120: 101, 5121: 101, 5122: 101, 5123: 101, 5376: 101, 5377: 101, 5453: 101, 5480: 101, 5483: 101,
	5484: 101, 5485: 101, 5827: 101, 5828: 101, 6145: 101, 6273: 101, 6274: 101, 6275: 101, 6276: 101, 6277: 101,
	6278: 101, 6279: 101, 6280: 101, 24586: 101, 24592: 101, 24593: 101, 24594: 101, 4608: 102, 4609: 102, 4610: 102,
	4611: 102, 4612: 102, 4614: 102, 4615: 102, 4616: 102, 4622: 102, 4624: 102, 4625: 102, 4634: 102, 4646: 102,
	4647: 102, 4648: 102, 4650: 102, 4651: 102, 4652: 102, 4653: 102, 4654: 102, 4655: 102, 4656: 102, 4657: 102,
	4658: 102, 4659: 102, 4660: 102, 4661: 102, 4662: 102, 4663: 102, 4664: 102, 4665: 102, 4666: 102, 4667: 102,
	4668: 102, 4670: 102, 4671: 102, 4672: 102, 4673: 102, 4674: 102, 4688: 102, 4689: 102, 4690: 102, 4691: 102,
	4694: 102, 4695: 102, 4696: 102, 4697: 102, 4698: 102, 4699: 102, 4700: 102, 4701: 102, 4702: 102, 4704: 102,
	4705: 102, 4707: 102, 4709: 102, 4710: 102, 4711: 102, 4712: 102, 4717: 102, 4718: 102, 4720: 102, 4722: 102,
	4723: 102, 4725: 102, 4726: 102, 4728: 102, 4729: 102, 4730: 102, 4731: 102, 4732: 102, 4733: 102, 4734: 102,
	4738: 102, 4740: 102, 4741: 102, 4742: 102, 4743: 102, 4744: 102, 4745: 102, 4746: 102, 4747: 102, 4748: 102,
	4749: 102, 4750: 102, 4751: 102, 4752: 102, 4753: 102, 4756: 102, 4757: 102, 4758: 102, 4759: 102, 4760: 102,
	4761: 102, 4762: 102, 4767: 102, 4768: 102, 4769: 102, 4770: 102, 4771: 102, 4772: 102, 4774: 102, 4775: 102,
	4776: 102, 4777: 102, 4778: 102, 4779: 102, 4781: 102, 4782: 102, 4783: 102, 4784: 102, 4785: 102, 4786: 102,
	4787: 102, 4788: 102, 4789: 102, 4790: 102, 4793: 102, 4800: 102, 4801: 102, 4802: 102, 4803: 102, 4864: 102,
	4869: 102, 4871: 102, 4872: 102, 4873: 102, 4874: 102, 4875: 102, 4876: 102, 4877: 102, 4878: 102, 4879: 102,
	4880: 102, 4881: 102, 4883: 102, 4884: 102, 4886: 102, 4887: 102, 4888: 102, 4889: 102, 4891: 102, 4893: 102,
	4894: 102, 4895: 102, 4898: 102, 4902: 102, 4904: 102, 4905: 102, 4909: 102, 4910: 102, 4928: 102, 4929: 102,
	4930: 102, 4931: 102, 4932: 102, 4933: 102, 4934: 102, 4935: 102, 4936: 102, 4937: 102, 4944: 102, 4945: 102,
	4946: 102, 4947: 102, 4948: 102, 4949: 102, 4950: 102, 4951: 102, 4952: 102, 4953: 102, 4954: 102, 4956: 102,
	4957: 102, 4958: 102, 4979: 102, 4980: 102, 4981: 102, 4982: 102, 4985: 102, 5024: 102, 5025: 102, 5031: 102,
	5032: 102, 5033: 102, 5034: 102, 5039: 102, 5040: 102, 5041: 102, 5042: 102, 5043: 102, 5044: 102, 5045: 102,
	5046: 102, 5047: 102, 5048: 102, 5050: 102, 5051: 102, 5056: 102, 5057: 102, 5058: 102, 5059: 102, 5060: 102,
	5061: 102, 5062: 102, 5063: 102, 5064: 102, 5065: 102, 5066: 102, 5067: 102, 5068: 102, 5069: 102, 5070: 102,
	5125: 102, 5126: 102, 5127: 102, 5136: 102, 5137: 102, 5138: 102, 5139: 102, 5140: 102, 5141: 102, 5152: 102,
	5153: 102, 5154: 102, 5155: 102, 5156: 102, 5157: 102, 5158: 102, 5159: 102, 5378: 102, 5440: 102, 5441: 102,
	5442: 102, 5443: 102, 5444: 102, 5446: 102, 5447: 102, 5448: 102, 5449: 102, 5450: 102, 5451: 102, 5452: 102,
	5456: 102, 5457: 102, 5458: 102, 5459: 102, 5460: 102, 5461: 102, 5462: 102, 5463: 102, 5464: 102, 5465: 102,
	5466: 102, 5467: 102, 5468: 102, 5471: 102, 5472: 102, 5473: 102, 5474: 102, 5477: 102, 5479: 102, 5632: 102,
	5633: 102, 5712: 102, 5888: 102, 5889: 102, 5890: 102, 6008: 102, 6144: 102, 6272: 102, 24577: 102, 24578: 102,
	24579: 102, 24580: 102, 24581: 102, 24582: 102, 24583: 102, 24584: 102, 24588: 102, 24595: 102, 24621: 102, 5049: 102,
	5478: 102,
}
