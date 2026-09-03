package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/twsnmp/twsnmpfc/client"
)

func (b *App) readLogFromTWSNMP(lf *LogFile) error {
	src := lf.LogSrc
	u, err := url.Parse(src.Server)
	scheme := "http"
	host := src.Server
	if err == nil && u.Host != "" {
		host = u.Host
		if u.Scheme != "" {
			scheme = u.Scheme
		}
	}
	if src.TLS {
		scheme = "https"
	}
	serverURL := fmt.Sprintf("%s://%s", scheme, host)
	c := client.NewClient(serverURL)
	c.InsecureSkipVerify = src.InsecureSkip
	c.Timeout = 60

	user := src.User
	if user == "" {
		user = "twsnmp"
	}
	pass := src.Password
	if pass == "" {
		pass = "twsnmp"
	}

	if err := c.Login(user, pass); err != nil {
		return fmt.Errorf("failed to login to TWSNMP FC: %w", err)
	}

	st, et := b.getLogSourceTimeRange(src)
	logType := src.SubTarget
	if logType == "" {
		logType = "syslog"
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		switch logType {
		case "eventlog":
			f := &client.EventLogFilter{
				StartDate: time.Unix(0, st).Format("2006-01-02"),
				StartTime: time.Unix(0, st).Format("15:04"),
				EndDate:   time.Unix(0, et).Format("2006-01-02"),
				EndTime:   time.Unix(0, et).Format("15:04"),
			}
			r, err := c.GetEventLogs(f)
			if err != nil {
				OutLog("TWSNMP GetEventLogs err=%v", err)
				return
			}
			for _, l := range r.EventLogs {
				if b.stopProcess {
					return
				}
				sl := fmt.Sprintf("%s %s '%s' %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), l.Type, l.NodeName, l.Event)
				pw.Write([]byte(sl))
			}

		case "trap":
			f := &client.SnmpTrapFilter{
				StartDate: time.Unix(0, st).Format("2006-01-02"),
				StartTime: time.Unix(0, st).Format("15:04"),
				EndDate:   time.Unix(0, et).Format("2006-01-02"),
				EndTime:   time.Unix(0, et).Format("15:04"),
			}
			traps, err := c.GetSnmpTraps(f)
			if err != nil {
				OutLog("TWSNMP GetSnmpTraps err=%v", err)
				return
			}
			for _, l := range traps {
				if b.stopProcess {
					return
				}
				sl := fmt.Sprintf("%s %s %s %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), l.FromAddress, l.TrapType, l.Variables)
				pw.Write([]byte(sl))
			}

		case "netflow", "ipfix":
			ipfix := logType == "ipfix"
			f := &client.NetflowFilter{
				NextTime: st,
				Filter:   0,
			}
			for ct := st; ct >= 0 && ct < et; {
				if b.stopProcess {
					return
				}
				f.NextTime = ct
				var r *client.NetflowWebAPI
				var err error
				if ipfix {
					r, err = c.GetIPFIX(f)
				} else {
					r, err = c.GetNetFlow(f)
				}
				if err != nil {
					OutLog("TWSNMP GetNetflow/IPFIX err=%v", err)
					return
				}
				for _, l := range r.Logs {
					j, err := json.Marshal(&l)
					if err != nil {
						continue
					}
					sl := fmt.Sprintf("%s %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), string(j))
					pw.Write([]byte(sl))
					ct = l.Time
				}
				if r.NextTime == 0 {
					break
				}
			}

		case "sflow":
			f := &client.SFlowFilter{
				NextTime: st,
				Filter:   0,
			}
			for ct := st; ct >= 0 && ct < et; {
				if b.stopProcess {
					return
				}
				f.NextTime = ct
				r, err := c.GetSFlow(f)
				if err != nil {
					OutLog("TWSNMP GetSFlow err=%v", err)
					return
				}
				for _, l := range r.Logs {
					j, err := json.Marshal(&l)
					if err != nil {
						continue
					}
					sl := fmt.Sprintf("%s %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), string(j))
					pw.Write([]byte(sl))
					ct = l.Time
				}
				if r.NextTime == 0 {
					break
				}
			}

		case "arp":
			f := &client.ArpFilter{
				StartDate: time.Unix(0, st).Format("2006-01-02"),
				StartTime: time.Unix(0, st).Format("15:04"),
				EndDate:   time.Unix(0, et).Format("2006-01-02"),
				EndTime:   time.Unix(0, et).Format("15:04"),
			}
			arpLogs, err := c.GetArpLogs(f)
			if err != nil {
				OutLog("TWSNMP GetArpLogs err=%v", err)
				return
			}
			for _, l := range arpLogs {
				if b.stopProcess {
					return
				}
				j, err := json.Marshal(&l)
				if err != nil {
					continue
				}
				sl := fmt.Sprintf("%s %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), string(j))
				pw.Write([]byte(sl))
			}

		default: // "syslog"
			f := &client.SyslogFilter{
				NextTime: st,
				Filter:   0,
			}
			for ct := st; ct >= 0 && ct < et; {
				if b.stopProcess {
					return
				}
				f.NextTime = ct
				r, err := c.GetSyslogs(f)
				if err != nil {
					OutLog("TWSNMP GetSyslogs err=%v", err)
					return
				}
				for _, l := range r.Logs {
					sl := fmt.Sprintf("%s %s %s: %s\n", time.Unix(0, l.Time).Format(time.RFC3339Nano), l.Type, l.Tag, l.Message)
					pw.Write([]byte(sl))
					ct = l.Time
				}
				if r.NextTime == 0 {
					break
				}
			}
		}
	}()

	b.readOneLogFile(lf, pr)
	return nil
}
