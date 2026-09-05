package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"blitiri.com.ar/go/spf"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/knadh/go-pop3"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func (b *App) readLogFromEmail(lf *LogFile) error {
	src := lf.LogSrc
	if src.Type == "pop3" || strings.HasPrefix(src.Server, "pop:") || strings.HasPrefix(src.Server, "pop3:") {
		return b.readLogFromPOP3(lf)
	}
	return b.readLogFromIMAP(lf)
}

func (b *App) readLogFromIMAP(lf *LogFile) error {
	src := lf.LogSrc
	c, err := b.getIMAPClient(src)
	if err != nil {
		return err
	}
	defer c.Logout()

	if err := c.Login(src.User, src.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP login failed: %w", err)
	}

	mbox := "INBOX"
	if src.Folder != "" {
		mbox = src.Folder
	} else if u, err := url.Parse(src.Server); err == nil && len(u.Path) > 1 {
		mbox = u.Path[1:]
	}

	status, err := c.Select(mbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("IMAP select %s failed: %w", mbox, err)
	}
	if status.NumMessages == 0 {
		return nil
	}

	var seqSet imap.SeqSet
	seqSet.AddRange(1, status.NumMessages)
	headerSection := imap.FetchItemBodySection{Specifier: imap.PartSpecifierHeader}
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{&headerSection},
	}
	fetchCmd := c.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	for {
		if b.stopProcess {
			break
		}
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		buf, err := msg.Collect()
		if err != nil {
			OutLog("IMAP collect message err=%v", err)
			continue
		}
		headerData := buf.FindBodySection(&headerSection)
		if len(headerData) < 1 {
			continue
		}

		ilf := &LogFile{
			Name:   fmt.Sprintf("%s#%d", mbox, msg.SeqNum),
			Path:   fmt.Sprintf("%s/%s#%d", src.Server, mbox, msg.SeqNum),
			Size:   int64(len(headerData)),
			Read:   0,
			Send:   0,
			LogSrc: src,
		}
		b.processStat.IntLogFiles = append(b.processStat.IntLogFiles, ilf)
		b.readLogFromEMailFile(ilf, bytes.NewReader(headerData))
		lf.Read += ilf.Read
		lf.Send += ilf.Send
	}
	return nil
}

func (b *App) getIMAPClient(src *LogSource) (*imapclient.Client, error) {
	u, err := url.Parse(src.Server)
	server := src.Server
	scheme := "imap"
	if err == nil && u.Host != "" {
		server = u.Host
		scheme = u.Scheme
	}
	if !strings.Contains(server, ":") {
		if scheme == "imaps" || src.TLS {
			server += ":993"
		} else {
			server += ":143"
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: src.InsecureSkip,
	}

	opts := &imapclient.Options{
		TLSConfig: tlsConfig,
	}

	if scheme == "imaps" || src.TLS || strings.HasSuffix(server, ":993") {
		return imapclient.DialTLS(server, opts)
	}
	return imapclient.DialInsecure(server, opts)
}

func (b *App) readLogFromPOP3(lf *LogFile) error {
	src := lf.LogSrc
	conn, err := b.getPOP3Conn(src)
	if err != nil {
		return err
	}
	defer conn.Quit()

	if err := conn.Auth(src.User, src.Password); err != nil {
		return fmt.Errorf("POP3 auth failed: %w", err)
	}
	count, _, err := conn.Stat()
	if err != nil {
		return fmt.Errorf("POP3 stat failed: %w", err)
	}
	if count == 0 {
		return nil
	}

	for i := 1; i <= count; i++ {
		if b.stopProcess {
			break
		}
		msg, err := conn.Retr(i)
		if err != nil {
			OutLog("POP3 retr %d err=%v", i, err)
			continue
		}
		a := []string{}
		for k, v := range msg.Header.Map() {
			for _, val := range v {
				a = append(a, fmt.Sprintf("%s: %s", k, val))
			}
		}
		rawHeader := strings.Join(a, "\r\n") + "\r\n"
		ilf := &LogFile{
			Name:   fmt.Sprintf("POP3#%d", i),
			Path:   fmt.Sprintf("%s#%d", src.Server, i),
			Size:   int64(len(rawHeader)),
			Read:   0,
			Send:   0,
			LogSrc: src,
		}
		b.processStat.IntLogFiles = append(b.processStat.IntLogFiles, ilf)
		b.readLogFromEMailFile(ilf, strings.NewReader(rawHeader))
		lf.Read += ilf.Read
		lf.Send += ilf.Send
	}
	return nil
}

func (b *App) getPOP3Conn(src *LogSource) (*pop3.Conn, error) {
	u, err := url.Parse(src.Server)
	server := src.Server
	port := 110
	isTLS := src.TLS

	if err == nil && u.Hostname() != "" {
		server = u.Hostname()
		if u.Port() != "" {
			if p, err := strconv.Atoi(u.Port()); err == nil {
				port = p
			}
		} else if u.Scheme == "pop3s" || isTLS {
			port = 995
			isTLS = true
		}
	} else if strings.Contains(server, ":") {
		h, p, err := net.SplitHostPort(server)
		if err == nil {
			server = h
			if pn, err := strconv.Atoi(p); err == nil {
				port = pn
			}
		}
	}

	p := pop3.New(pop3.Opt{
		Host:       server,
		Port:       port,
		TLSEnabled: isTLS || port == 995,
	})
	return p.NewConn()
}

func (b *App) readLogFromEMailFile(lf *LogFile, r io.Reader) error {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}
	ts, err := msg.Header.Date()
	t := int64(0)
	if err == nil {
		t = ts.UnixNano()
	} else if b.config.NoTimeStamp {
		t = time.Now().UnixNano()
	}
	if b.processStat.TimeLine == nil {
		b.processStat.TimeLine = make(map[int64]int)
	}

	a := []string{}
	for k, va := range msg.Header {
		for _, v := range va {
			a = append(a, fmt.Sprintf("%s: %s", k, v))
		}
	}
	raw := strings.Join(a, "\r\n") + "\r\n"
	lf.Read += int64(len(raw))
	b.processStat.ReadLines += len(a)

	if b.processConf.Filter != nil && !b.processConf.Filter.MatchString(raw) {
		b.processStat.SkipLines += len(a)
		return nil
	}

	log := &LogEnt{
		ID:       fmt.Sprintf("%s:%06d", lf.Path, 1),
		KeyValue: make(map[string]interface{}),
		Time:     t,
		All:      raw,
	}

	info := parseEmailMessage(msg, false)
	if info != nil {
		log.KeyValue["from"] = info.From
		log.KeyValue["to"] = info.To
		log.KeyValue["subject"] = info.Subject
		log.KeyValue["domain"] = info.Domain
		log.KeyValue["senderIP"] = info.SenderIP
		log.KeyValue["relay"] = float64(info.Relay)
		log.KeyValue["delay"] = float64(info.Delay)
	}

	timeH := log.Time / (1000 * 1000 * 1000 * 3600)
	if _, ok := b.processStat.TimeLine[timeH]; !ok {
		b.processStat.TimeLine[timeH] = 0
	}
	b.processStat.TimeLine[timeH]++
	if log.Time < b.processStat.StartTime {
		b.processStat.StartTime = log.Time
	}
	if log.Time > b.processStat.EndTime {
		b.processStat.EndTime = log.Time
	}

	b.logCh <- log
	lf.Send += int64(len(raw))
	return nil
}

// ListIMAPFolders lists mailbox folders from the given IMAP source
func (b *App) ListIMAPFolders(src LogSource) ([]string, string) {
	c, err := b.getIMAPClient(&src)
	if err != nil {
		return nil, err.Error()
	}
	defer c.Logout()
	if err := c.Login(src.User, src.Password).Wait(); err != nil {
		return nil, err.Error()
	}
	mailboxes, err := c.List("", "*", nil).Collect()
	if err != nil {
		return nil, err.Error()
	}
	var res []string
	for _, mbox := range mailboxes {
		res = append(res, mbox.Mailbox)
	}
	return res, ""
}

type EmailInfo struct {
	Time       int64  `json:"Time"`
	From       string `json:"From"`
	To         string `json:"To"`
	Subject    string `json:"Subject"`
	ReturnPath string `json:"ReturnPath"`
	Domain     string `json:"Domain"`
	SenderIP   string `json:"SenderIP"`
	SPF        string `json:"SPF"`
	SPFError   string `json:"SPFError"`
	Relay      int    `json:"Relay"`
	Delay      int    `json:"Delay"`
	Log        string `json:"Log"`
}

type EmailCount struct {
	Name  string `json:"Name"`
	Count int    `json:"Count"`
}

type EmailMatrix struct {
	From  string `json:"From"`
	To    string `json:"To"`
	Count int    `json:"Count"`
}

type EmailReportStats struct {
	TotalEmails int            `json:"TotalEmails"`
	SPFPass     int            `json:"SPFPass"`
	SPFFail     int            `json:"SPFFail"`
	SPFSoftFail int            `json:"SPFSoftFail"`
	SPFNone     int            `json:"SPFNone"`
	SPFOther    int            `json:"SPFOther"`
	AvgDelay    float64        `json:"AvgDelay"`
	MaxDelay    int            `json:"MaxDelay"`
	AvgRelay    float64        `json:"AvgRelay"`
	MaxRelay    int            `json:"MaxRelay"`
	TopFrom     []*EmailCount  `json:"TopFrom"`
	TopTo       []*EmailCount  `json:"TopTo"`
	TopDomain   []*EmailCount  `json:"TopDomain"`
	TopSenderIP []*EmailCount  `json:"TopSenderIP"`
	Matrix      []*EmailMatrix `json:"Matrix"`
}

type EmailReportResult struct {
	Stats EmailReportStats `json:"Stats"`
	Items []*EmailInfo     `json:"Items"`
}

var (
	emailSPFMap   = make(map[string]string)
	emailSPFMapMu sync.RWMutex
	mimeWordDec   = &mime.WordDecoder{
		CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
			switch strings.ToLower(charset) {
			case "iso-2022-jp":
				return transform.NewReader(input, japanese.ISO2022JP.NewDecoder()), nil
			case "shift_jis", "shift-jis", "sjis":
				return transform.NewReader(input, japanese.ShiftJIS.NewDecoder()), nil
			case "euc-jp":
				return transform.NewReader(input, japanese.EUCJP.NewDecoder()), nil
			}
			return nil, fmt.Errorf("unhandled charset %q", charset)
		},
	}
)

func getMimeDecodedWord(s string) string {
	r, err := mimeWordDec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return r
}

func extractDomain(addr string) string {
	if strings.Contains(addr, "@") {
		parts := strings.Split(addr, "@")
		domain := strings.Trim(parts[len(parts)-1], "> ")
		return domain
	}
	return addr
}

func extractIP(received string) net.IP {
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindAllStringSubmatch(received, -1)
	for _, match := range matches {
		if len(match) > 1 {
			ip := net.ParseIP(match[1])
			if ip != nil && ip.IsGlobalUnicast() {
				return ip
			}
		}
	}
	return nil
}

func extractHelo(received string) string {
	re := regexp.MustCompile(`(?i)from\s+([^\s(]+)`)
	matches := re.FindStringSubmatch(received)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractTimestamp(header string) (time.Time, error) {
	parts := strings.Split(header, ";")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("; not found")
	}
	dateStr := strings.TrimSpace(parts[len(parts)-1])
	re := regexp.MustCompile(`\s*\(.*\)$`)
	dateStr = re.ReplaceAllString(dateStr, "")
	t, err := mail.ParseDate(dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %v", err)
	}
	return t, nil
}

func getSPF(ip net.IP, helo, sender string) (string, string) {
	k := fmt.Sprintf("%s/%s/%s", ip, helo, sender)
	emailSPFMapMu.RLock()
	if v, ok := emailSPFMap[k]; ok {
		emailSPFMapMu.RUnlock()
		a := strings.SplitN(v, "\t", 2)
		if len(a) == 2 {
			return a[0], a[1]
		}
		return v, ""
	}
	emailSPFMapMu.RUnlock()

	r, err := spf.CheckHostWithSender(ip, helo, sender)
	emailSPFMapMu.Lock()
	defer emailSPFMapMu.Unlock()
	if err != nil {
		emailSPFMap[k] = fmt.Sprintf("%s\t%v", r, err)
		return string(r), err.Error()
	}
	emailSPFMap[k] = string(r)
	return string(r), ""
}

func parseEmailMessage(msg *mail.Message, checkSPF bool) *EmailInfo {
	if msg == nil {
		return nil
	}
	r := &EmailInfo{}
	subject := getMimeDecodedWord(msg.Header.Get("Subject"))
	r.Subject = subject

	from := ""
	if list, err := msg.Header.AddressList("From"); err == nil {
		for i, a := range list {
			name := getMimeDecodedWord(a.Name)
			if i > 0 {
				from += ", "
			}
			if name != "" {
				from += fmt.Sprintf("%s <%s>", name, a.Address)
			} else {
				from += a.Address
			}
		}
	}
	r.From = from

	to := ""
	if list, err := msg.Header.AddressList("To"); err == nil {
		for i, a := range list {
			name := getMimeDecodedWord(a.Name)
			if i > 0 {
				to += ", "
			}
			if name != "" {
				to += fmt.Sprintf("%s <%s>", name, a.Address)
			} else {
				to += a.Address
			}
		}
	}
	r.To = to

	returnPath := msg.Header.Get("Return-Path")
	if returnPath == "" {
		if list, err := msg.Header.AddressList("From"); err == nil && len(list) > 0 {
			returnPath = list[0].Address
		}
	}
	returnPath = strings.Trim(returnPath, "<> ")
	r.ReturnPath = returnPath
	domain := extractDomain(returnPath)
	r.Domain = domain

	mt, errdate := msg.Header.Date()
	if errdate == nil {
		r.Time = mt.UnixNano()
	}
	r.Delay = 0
	r.Relay = 0

	for _, v := range msg.Header["Received"] {
		if errdate == nil {
			if rt, err := extractTimestamp(v); err == nil {
				dt := int(rt.Unix() - mt.Unix())
				if dt > r.Delay {
					r.Delay = dt
				}
			}
		}
		if !strings.Contains(v, "qmail") && !strings.Contains(v, "invoked") && !strings.Contains(v, "with l") {
			r.Relay++
		}
		senderIP := extractIP(v)
		if senderIP != nil && r.SenderIP == "" {
			r.SenderIP = senderIP.String()
			helo := extractHelo(v)
			if helo == "" {
				helo = domain
			}
			if checkSPF {
				spfResult, spfErr := getSPF(senderIP, helo, returnPath)
				r.SPF = spfResult
				r.SPFError = spfErr
			}
		}
	}
	return r
}

func parseEmailText(raw string, checkSPF bool) *EmailInfo {
	msg, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		return nil
	}
	info := parseEmailMessage(msg, checkSPF)
	if info != nil {
		info.Log = raw
	}
	return info
}

// AnalyzeEmailLogs analyzes email logs, extracting headers, calculating delay/relay, and evaluating SPF
func (b *App) AnalyzeEmailLogs(logs []*LogEnt, checkSPF bool) (*EmailReportResult, error) {
	result := &EmailReportResult{
		Stats: EmailReportStats{},
		Items: []*EmailInfo{},
	}
	if len(logs) == 0 {
		return result, nil
	}

	fromMap := make(map[string]int)
	toMap := make(map[string]int)
	domainMap := make(map[string]int)
	senderIPMap := make(map[string]int)
	matrixMap := make(map[string]int)

	totalDelay := 0
	totalRelay := 0

	for _, l := range logs {
		if l == nil {
			continue
		}
		info := parseEmailText(l.All, checkSPF)
		if info == nil {
			// KeyValue からフォールバック抽出
			from, _ := l.KeyValue["from"].(string)
			to, _ := l.KeyValue["to"].(string)
			subject, _ := l.KeyValue["subject"].(string)
			domain, _ := l.KeyValue["domain"].(string)
			senderIP, _ := l.KeyValue["senderIP"].(string)
			relay, _ := l.KeyValue["relay"].(float64)
			delay, _ := l.KeyValue["delay"].(float64)

			if from == "" && to == "" && subject == "" {
				continue
			}
			info = &EmailInfo{
				Time:     l.Time,
				From:     from,
				To:       to,
				Subject:  subject,
				Domain:   domain,
				SenderIP: senderIP,
				Relay:    int(relay),
				Delay:    int(delay),
				Log:      l.All,
			}
		}

		if info.Time == 0 {
			info.Time = l.Time
		}
		result.Items = append(result.Items, info)

		result.Stats.TotalEmails++
		totalDelay += info.Delay
		if info.Delay > result.Stats.MaxDelay {
			result.Stats.MaxDelay = info.Delay
		}
		totalRelay += info.Relay
		if info.Relay > result.Stats.MaxRelay {
			result.Stats.MaxRelay = info.Relay
		}

		// SPF 集計
		switch strings.ToLower(info.SPF) {
		case "pass":
			result.Stats.SPFPass++
		case "fail":
			result.Stats.SPFFail++
		case "softfail":
			result.Stats.SPFSoftFail++
		case "none", "":
			result.Stats.SPFNone++
		default:
			result.Stats.SPFOther++
		}

		if info.From != "" {
			fromMap[info.From]++
		}
		if info.To != "" {
			toMap[info.To]++
		}
		if info.Domain != "" {
			domainMap[info.Domain]++
		}
		if info.SenderIP != "" {
			senderIPMap[info.SenderIP]++
		}
		if info.From != "" && info.To != "" {
			matrixKey := fmt.Sprintf("%s\t%s", info.From, info.To)
			matrixMap[matrixKey]++
		}
	}

	if result.Stats.TotalEmails > 0 {
		result.Stats.AvgDelay = float64(totalDelay) / float64(result.Stats.TotalEmails)
		result.Stats.AvgRelay = float64(totalRelay) / float64(result.Stats.TotalEmails)
	}

	// Top 集計ソートヘルパー
	sortCounts := func(m map[string]int, limit int) []*EmailCount {
		list := make([]*EmailCount, 0, len(m))
		for k, v := range m {
			list = append(list, &EmailCount{Name: k, Count: v})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].Count > list[j].Count
		})
		if len(list) > limit {
			list = list[:limit]
		}
		return list
	}

	result.Stats.TopFrom = sortCounts(fromMap, 20)
	result.Stats.TopTo = sortCounts(toMap, 20)
	result.Stats.TopDomain = sortCounts(domainMap, 20)
	result.Stats.TopSenderIP = sortCounts(senderIPMap, 20)

	// Matrix ソート
	matrixList := make([]*EmailMatrix, 0, len(matrixMap))
	for k, v := range matrixMap {
		parts := strings.SplitN(k, "\t", 2)
		if len(parts) == 2 {
			matrixList = append(matrixList, &EmailMatrix{
				From:  parts[0],
				To:    parts[1],
				Count: v,
			})
		}
	}
	sort.Slice(matrixList, func(i, j int) bool {
		return matrixList[i].Count > matrixList[j].Count
	})
	if len(matrixList) > 30 {
		matrixList = matrixList[:30]
	}
	result.Stats.Matrix = matrixList

	return result, nil
}
