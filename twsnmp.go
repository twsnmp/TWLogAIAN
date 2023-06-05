package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type twsnmpLogin struct {
	UserID   string `json:"UserID"`
	Password string `json:"Password"`
}

type twsnmpSyslogFilter struct {
	StartDate string
	StartTime string
	EndDate   string
	EndTime   string
	Level     string
	Type      string
	Host      string
	Tag       string
	Message   string
	Extractor string
	NextTime  int64
	Filter    int
}

type twsnmpSyslogResp struct {
	Logs          []*twsnmpSyslogLogEnt
	ExtractHeader []string
	ExtractDatas  [][]string
	NextTime      int64
	Process       int
	Filter        int
	Limit         int
}

type twsnmpSyslogLogEnt struct {
	Time     int64
	Level    string
	Host     string
	Type     string
	Tag      string
	Message  string
	Severity int
	Facility int
}

func (b *App) readLogFromTWSNMP(lf *LogFile) error {
	token, err := b.loginToTWSNMP(lf)
	if err != nil {
		return err
	}
	return b.getLogFromTWSNMP(lf, token)
}

func (b *App) loginToTWSNMP(lf *LogFile) (string, error) {
	login := new(twsnmpLogin)
	login.UserID = lf.LogSrc.User
	login.Password = lf.LogSrc.Password

	login_json, _ := json.Marshal(login)

	res, err := http.Post(lf.LogSrc.Server+"login", "application/json", bytes.NewBuffer(login_json))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	respMap := make(map[string]string)
	err = json.Unmarshal(r, &respMap)
	if err != nil {
		return "", err
	}
	return respMap["token"], nil
}

func (b *App) getLogFromTWSNMP(lf *LogFile, token string) error {

	filter := new(twsnmpSyslogFilter)
	if lf.LogSrc.Start != "" {
		if a := strings.SplitN(lf.LogSrc.Start, "T", 2); len(a) == 2 && a[1] != "" {
			filter.StartDate = a[0]
			filter.StartTime = a[1]
		}
	}
	if lf.LogSrc.End != "" {
		if a := strings.SplitN(lf.LogSrc.End, "T", 2); len(a) == 2 && a[1] != "" {
			filter.EndDate = a[0]
			filter.EndTime = a[1]
		}
	}
	filter.Host = lf.LogSrc.Host
	filter.Tag = lf.LogSrc.Tag
	filter.Message = lf.LogSrc.Pattern

	filter_json, _ := json.Marshal(filter)

	req, err := http.NewRequest("POST", lf.LogSrc.Server+"api/log/syslog", bytes.NewBuffer(filter_json))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	r, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	syslogResp := new(twsnmpSyslogResp)
	err = json.Unmarshal(r, &syslogResp)
	if err != nil {
		return err
	}
	logs := []string{}
	for _, l := range syslogResp.Logs {
		ts := time.Unix(0, l.Time)
		logs = append(logs, fmt.Sprintf("%s %s %s %s", ts.Format(time.RFC3339Nano), l.Host, l.Tag, l.Message))
	}
	b.readOneLogFile(lf, strings.NewReader(strings.Join(logs, "\n")))
	return nil
}
