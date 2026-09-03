package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type lokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func (b *App) readLogFromLoki(lf *LogFile) error {
	src := lf.LogSrc
	u, err := url.Parse(src.Server)
	if err != nil {
		return err
	}

	scheme := "http"
	if u.Scheme == "lokis" || u.Scheme == "https" || src.TLS || strings.HasPrefix(src.Server, "https:") {
		scheme = "https"
	}
	host := u.Host
	if host == "" {
		host = u.Path
	}

	lokiURL := fmt.Sprintf("%s://%s/loki/api/v1/query_range", scheme, host)
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	query := src.Query
	if query == "" && src.Pattern != "" {
		query = src.Pattern
	}
	if query == "" && u.Path != "" && strings.Contains(u.Path, "{") {
		p := strings.Trim(u.Path, "/")
		if strings.HasPrefix(p, "{") {
			query = p
		}
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: src.InsecureSkip,
			},
		},
	}

	st, et := b.getLogSourceTimeRange(src)
	if query == "" {
		query = resolveLokiQuery(client, baseURL, src, st, et)
	}

	const maxWindowNanos = int64(24 * time.Hour)
	limit := 5000

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for currentStart := st; currentStart < et && !b.stopProcess; {
			reqEnd := currentStart + maxWindowNanos
			if reqEnd > et {
				reqEnd = et
			}

			params := url.Values{}
			params.Set("query", query)
			params.Set("start", strconv.FormatInt(currentStart, 10))
			params.Set("end", strconv.FormatInt(reqEnd, 10))
			params.Set("limit", strconv.Itoa(limit))
			params.Set("direction", "FORWARD")

			reqURL := fmt.Sprintf("%s?%s", lokiURL, params.Encode())
			req, err := http.NewRequest("GET", reqURL, nil)
			if err != nil {
				OutLog("Loki new request err=%v", err)
				return
			}

			if src.User != "" || src.Password != "" {
				req.SetBasicAuth(src.User, src.Password)
			} else if u.User != nil {
				pass, _ := u.User.Password()
				req.SetBasicAuth(u.User.Username(), pass)
			}
			token := src.Token
			if token == "" {
				token = src.APIKey
			}
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			if src.OrgID != "" {
				req.Header.Set("X-Scope-OrgID", src.OrgID)
			}

			resp, err := client.Do(req)
			if err != nil {
				OutLog("Loki Do request err=%v", err)
				return
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				OutLog("Loki ReadAll err=%v", err)
				return
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				OutLog("Loki error response [%d]: %s", resp.StatusCode, string(body))
				return
			}

			var lokiResp lokiQueryResponse
			if err := json.Unmarshal(body, &lokiResp); err != nil {
				OutLog("Loki json unmarshal err=%v", err)
				return
			}

			type valEnt struct {
				ts     int64
				log    string
				stream map[string]string
			}
			var batchEntries []valEnt

			for _, res := range lokiResp.Data.Result {
				for _, v := range res.Values {
					if len(v) < 2 {
						continue
					}
					ts, err := strconv.ParseInt(v[0], 10, 64)
					if err != nil {
						continue
					}
					batchEntries = append(batchEntries, valEnt{ts: ts, log: v[1], stream: res.Stream})
				}
			}

			if len(batchEntries) == 0 {
				currentStart = reqEnd
				if currentStart >= et {
					break
				}
				continue
			}

			maxTs := currentStart
			for _, ent := range batchEntries {
				if b.stopProcess {
					return
				}
				t := ent.ts
				if t > maxTs {
					maxTs = t
				}

				logMap := make(map[string]interface{})
				for k, v := range ent.stream {
					logMap[k] = v
				}
				logMap["message"] = ent.log
				jsonBytes, err := json.Marshal(logMap)
				var l string
				if err != nil {
					l = fmt.Sprintf("%s %s\n", time.Unix(0, t).Format(time.RFC3339Nano), ent.log)
				} else {
					l = fmt.Sprintf("%s %s\n", time.Unix(0, t).Format(time.RFC3339Nano), string(jsonBytes))
				}
				if _, err := pw.Write([]byte(l)); err != nil {
					return
				}
			}

			if len(batchEntries) >= limit {
				if maxTs <= currentStart {
					currentStart++
				} else {
					currentStart = maxTs + 1
				}
			} else {
				currentStart = reqEnd
			}
		}
	}()

	b.readOneLogFile(lf, pr)
	return nil
}

func resolveLokiQuery(client *http.Client, baseURL string, src *LogSource, st, et int64) string {
	labelsURL := fmt.Sprintf("%s/loki/api/v1/labels?start=%d&end=%d", baseURL, st, et)
	req, err := http.NewRequest("GET", labelsURL, nil)
	if err == nil {
		if src.User != "" || src.Password != "" {
			req.SetBasicAuth(src.User, src.Password)
		}
		token := src.Token
		if token == "" {
			token = src.APIKey
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if src.OrgID != "" {
			req.Header.Set("X-Scope-OrgID", src.OrgID)
		}

		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var lResp struct {
					Status string   `json:"status"`
					Data   []string `json:"data"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&lResp); err == nil && len(lResp.Data) > 0 {
					preferred := []string{"job", "app", "service_name", "container", "filename", "stream"}
					for _, pref := range preferred {
						for _, l := range lResp.Data {
							if l == pref {
								return fmt.Sprintf("{%s=~\".+\"}", l)
							}
						}
					}
					for _, l := range lResp.Data {
						if l != "" && !strings.HasPrefix(l, "__") {
							return fmt.Sprintf("{%s=~\".+\"}", l)
						}
					}
				}
			}
		}
	}
	return `{job=~".+"}`
}
