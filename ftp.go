package main

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

func (b *App) readLogFromFTP(lf *LogFile) error {
	src := lf.LogSrc
	u, err := url.Parse(src.Server)
	host := src.Server
	port := "21"
	isTLS := src.TLS || strings.HasPrefix(src.Server, "ftps:")

	if err == nil && u.Hostname() != "" {
		host = u.Hostname()
		if u.Port() != "" {
			port = u.Port()
		}
		if u.Scheme == "ftps" {
			isTLS = true
		}
	} else if strings.Contains(host, ":") {
		h, p, err := net.SplitHostPort(host)
		if err == nil {
			host = h
			port = p
		}
	}

	user := src.User
	pass := src.Password
	if user == "" {
		user = "anonymous"
	}
	if pass == "" {
		pass = "anonymous@"
	}

	var dialOpts []ftp.DialOption
	dialOpts = append(dialOpts, ftp.DialWithTimeout(15*time.Second))

	if isTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: src.InsecureSkip,
			ServerName:         host,
		}
		dialOpts = append(dialOpts, ftp.DialWithTLS(tlsConfig))
	}

	addr := net.JoinHostPort(host, port)
	client, err := ftp.Dial(addr, dialOpts...)
	if err != nil {
		return fmt.Errorf("FTP failed to dial %s: %w", addr, err)
	}
	defer client.Quit()

	if err := client.Login(user, pass); err != nil {
		return fmt.Errorf("FTP failed to login as %s: %w", user, err)
	}

	targetPath := src.Path
	if targetPath == "" {
		targetPath = "/"
	}

	filter, err := getFileNameFilter(src.Pattern)
	if err != nil {
		return err
	}

	// ディレクトリまたはファイルパターン指定の場合
	if strings.HasSuffix(targetPath, "/") || src.Pattern != "" {
		entries, err := listFTPDirectory(client, targetPath)
		if err != nil {
			return fmt.Errorf("FTP failed to list directory %s: %w", targetPath, err)
		}
		var filePaths []string
		for _, entry := range entries {
			if entry.Type == ftp.EntryTypeFolder {
				continue
			}
			fileName := path.Base(entry.Name)
			if filter != nil && !filter.MatchString(fileName) {
				continue
			}
			if strings.HasPrefix(entry.Name, "/") {
				filePaths = append(filePaths, entry.Name)
			} else {
				filePaths = append(filePaths, path.Join(targetPath, fileName))
			}
		}
		if len(filePaths) == 0 {
			return fmt.Errorf("no matching files in FTP directory %s", targetPath)
		}
		return b.readMultipleFTPFiles(client, lf, filePaths)
	}

	// 単一ファイル取得
	if err := b.fetchFTPFileToLogFile(client, lf, targetPath); err != nil {
		firstErr := err
		// ディレクトリ末尾スラッシュなしの可能性をチェック
		entries, listErr := listFTPDirectory(client, targetPath)
		if listErr == nil && len(entries) > 0 {
			firstBase := path.Base(entries[0].Name)
			if len(entries) > 1 || entries[0].Type == ftp.EntryTypeFolder || firstBase != path.Base(targetPath) {
				var filePaths []string
				for _, entry := range entries {
					if entry.Type == ftp.EntryTypeFolder {
						continue
					}
					fileName := path.Base(entry.Name)
					if filter != nil && !filter.MatchString(fileName) {
						continue
					}
					if strings.HasPrefix(entry.Name, "/") {
						filePaths = append(filePaths, entry.Name)
					} else {
						filePaths = append(filePaths, path.Join(targetPath, fileName))
					}
				}
				if len(filePaths) > 0 {
					return b.readMultipleFTPFiles(client, lf, filePaths)
				}
			}
		}
		return firstErr
	}
	return nil
}

func (b *App) readMultipleFTPFiles(client *ftp.ServerConn, lf *LogFile, filePaths []string) error {
	readCount := 0
	var lastErr error
	for i, fp := range filePaths {
		if b.stopProcess {
			break
		}
		var targetLf *LogFile
		if i == 0 {
			targetLf = lf
		} else {
			targetLf = &LogFile{
				Name:   path.Base(fp),
				Path:   lf.LogSrc.Server + ":" + fp,
				Size:   0,
				Read:   0,
				Send:   0,
				LogSrc: lf.LogSrc,
			}
			b.processStat.IntLogFiles = append(b.processStat.IntLogFiles, targetLf)
		}
		if err := b.fetchFTPFileToLogFile(client, targetLf, fp); err != nil {
			OutLog("fetchFTPFileToLogFile err=%v", err)
			lastErr = err
		} else {
			readCount++
		}
	}
	if readCount == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func listFTPDirectory(client *ftp.ServerConn, dirPath string) ([]*ftp.Entry, error) {
	entries, err := client.List(dirPath)
	if err == nil {
		return entries, nil
	}
	// 先頭スラッシュを除去した相対パスでリトライ
	relPath := strings.TrimPrefix(dirPath, "/")
	if relPath != dirPath && relPath != "" {
		if entries2, err2 := client.List(relPath); err2 == nil {
			return entries2, nil
		}
	}
	return nil, err
}

func (b *App) fetchFTPFileToLogFile(client *ftp.ServerConn, lf *LogFile, filePath string) error {
	resp, err := client.Retr(filePath)
	if err != nil {
		// 先頭スラッシュを除去した相対パスでリトライ
		relPath := strings.TrimPrefix(filePath, "/")
		if relPath != filePath && relPath != "" {
			resp2, err2 := client.Retr(relPath)
			if err2 == nil {
				resp = resp2
				err = nil
			}
		}
	}
	if err != nil {
		return fmt.Errorf("FTP failed to retrieve %s: %w", filePath, err)
	}
	defer resp.Close()

	lf.Name = path.Base(filePath)
	lf.Path = lf.LogSrc.Server + ":" + filePath

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".gz" {
		gzr, err := gzip.NewReader(resp)
		if err != nil {
			return fmt.Errorf("failed to decompress gzip for %s: %w", filePath, err)
		}
		defer gzr.Close()
		b.readOneLogFile(lf, gzr)
	} else {
		b.readOneLogFile(lf, resp)
	}
	return nil
}
