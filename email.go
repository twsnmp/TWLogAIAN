package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/knadh/go-pop3"
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
