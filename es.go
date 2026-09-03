package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type esSearchHit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Source map[string]interface{} `json:"_source"`
	Sort   []interface{}          `json:"sort"`
}

type esSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []esSearchHit `json:"hits"`
	} `json:"hits"`
	Error *struct {
		Reason string `json:"reason"`
		Type   string `json:"type"`
	} `json:"error,omitempty"`
}

func (b *App) readLogFromES(lf *LogFile) error {
	src := lf.LogSrc
	serverStr := strings.TrimSpace(src.Server)
	if strings.HasPrefix(serverStr, "htt://") {
		serverStr = "http://" + strings.TrimPrefix(serverStr, "htt://")
	} else if strings.HasPrefix(serverStr, "htts://") {
		serverStr = "https://" + strings.TrimPrefix(serverStr, "htts://")
	} else if !strings.Contains(serverStr, "://") {
		serverStr = "http://" + serverStr
	}

	u, err := url.Parse(serverStr)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	scheme := "http"
	if u.Scheme == "ess" || u.Scheme == "opensearchs" || u.Scheme == "https" || src.TLS || strings.HasPrefix(serverStr, "https:") {
		scheme = "https"
	}
	host := u.Host
	if host == "" {
		host = u.Path
	}

	targetIndex := src.Index
	if targetIndex == "" {
		p := strings.Trim(u.Path, "/")
		p = strings.TrimSuffix(p, "/_search")
		p = strings.TrimSuffix(p, "_search")
		p = strings.Trim(p, "/")
		if p != "" {
			targetIndex = p
		} else {
			targetIndex = "*"
		}
	} else {
		targetIndex = strings.TrimSuffix(targetIndex, "/_search")
		targetIndex = strings.TrimSuffix(targetIndex, "_search")
		targetIndex = strings.Trim(targetIndex, "/")
		if targetIndex == "" {
			targetIndex = "*"
		}
	}

	searchURL := fmt.Sprintf("%s://%s/%s/_search", scheme, host, targetIndex)

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: src.InsecureSkip,
			},
		},
	}

	timeField := src.TimeField
	if timeField == "" {
		timeField = "@timestamp"
	}

	queryStr := src.Query
	if queryStr == "" && src.Pattern != "" {
		queryStr = src.Pattern
	}
	if queryStr == "" {
		queryStr = "*"
	}

	hasTimeRange := src.Start != "" || src.End != ""
	var st, et int64
	if hasTimeRange {
		st, et = b.getLogSourceTimeRange(src)
	}

	size := 1000
	pr, pw := io.Pipe()

	var fetchErr error
	go func() {
		defer func() {
			if fetchErr != nil {
				pw.CloseWithError(fetchErr)
			} else {
				pw.Close()
			}
		}()
		var lastSort []interface{}

		for !b.stopProcess {
			reqMap := map[string]interface{}{
				"size": size,
				"sort": []interface{}{
					map[string]interface{}{
						timeField: map[string]interface{}{
							"order":         "asc",
							"unmapped_type": "date",
						},
					},
					map[string]interface{}{
						"_id": map[string]interface{}{
							"order": "asc",
						},
					},
				},
			}

			var queryObj map[string]interface{}
			if strings.HasPrefix(strings.TrimSpace(queryStr), "{") {
				if err := json.Unmarshal([]byte(queryStr), &queryObj); err != nil {
					queryObj = map[string]interface{}{
						"query_string": map[string]interface{}{
							"query": queryStr,
						},
					}
				}
			} else {
				queryObj = map[string]interface{}{
					"query_string": map[string]interface{}{
						"query": queryStr,
					},
				}
			}

			var filterList []map[string]interface{}
			if hasTimeRange && (st > 0 || et > 0) {
				rangeMap := map[string]interface{}{}
				if st > 0 {
					rangeMap["gte"] = st / 1e6
				}
				if et > 0 {
					rangeMap["lte"] = et / 1e6
				}
				rangeMap["format"] = "epoch_millis"
				filterList = append(filterList, map[string]interface{}{
					"range": map[string]interface{}{
						timeField: rangeMap,
					},
				})
			}

			reqMap["query"] = map[string]interface{}{
				"bool": map[string]interface{}{
					"must":   queryObj,
					"filter": filterList,
				},
			}

			if len(lastSort) > 0 {
				reqMap["search_after"] = lastSort
			}

			reqBytes, err := json.Marshal(reqMap)
			if err != nil {
				fetchErr = fmt.Errorf("ES req json marshal err: %w", err)
				OutLog("%v", fetchErr)
				return
			}

			req, err := http.NewRequest("POST", searchURL, bytes.NewReader(reqBytes))
			if err != nil {
				fetchErr = fmt.Errorf("ES new request err: %w", err)
				OutLog("%v", fetchErr)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			if src.User != "" || src.Password != "" {
				req.SetBasicAuth(src.User, src.Password)
			} else if u.User != nil {
				pass, _ := u.User.Password()
				req.SetBasicAuth(u.User.Username(), pass)
			}
			apiKey := src.APIKey
			if apiKey == "" {
				apiKey = src.Token
			}
			if apiKey != "" {
				req.Header.Set("Authorization", "ApiKey "+apiKey)
			}

			resp, err := client.Do(req)
			if err != nil {
				fetchErr = fmt.Errorf("ES request failed: %w", err)
				OutLog("%v", fetchErr)
				return
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fetchErr = fmt.Errorf("ES read response err: %w", err)
				OutLog("%v", fetchErr)
				return
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				fetchErr = fmt.Errorf("ES/OpenSearch error [%d]: %s", resp.StatusCode, string(body))
				OutLog("%v", fetchErr)
				return
			}

			var searchResp esSearchResponse
			if err := json.Unmarshal(body, &searchResp); err != nil {
				fetchErr = fmt.Errorf("ES unmarshal searchResp err: %w", err)
				OutLog("%v", fetchErr)
				return
			}

			if searchResp.Error != nil {
				fetchErr = fmt.Errorf("ES search error: %s (%s)", searchResp.Error.Reason, searchResp.Error.Type)
				OutLog("%v", fetchErr)
				return
			}

			hits := searchResp.Hits.Hits
			if len(hits) == 0 {
				break
			}

			for _, hit := range hits {
				if b.stopProcess {
					return
				}
				lastSort = hit.Sort

				t := int64(0)
				if rawTs, ok := hit.Source[timeField]; ok {
					t = parseESTimestamp(rawTs)
				}
				if t == 0 {
					for _, tf := range []string{"timestamp", "time", "datetime", "created_at", "@time", "date", "Date"} {
						if rawTs, ok := hit.Source[tf]; ok {
							if parsed := parseESTimestamp(rawTs); parsed > 0 {
								t = parsed
								break
							}
						}
					}
				}
				if t == 0 {
					t = time.Now().UnixNano()
				}

				jsonBytes, err := json.Marshal(hit.Source)
				var logContent string
				if err != nil {
					logContent = fmt.Sprintf("%s %v\n", time.Unix(0, t).Format(time.RFC3339Nano), hit.Source)
				} else {
					logContent = fmt.Sprintf("%s %s\n", time.Unix(0, t).Format(time.RFC3339Nano), string(jsonBytes))
				}

				if _, err := pw.Write([]byte(logContent)); err != nil {
					return
				}
			}

			if len(hits) < size {
				break
			}
		}
	}()

	b.readOneLogFile(lf, pr)
	return fetchErr
}

func parseESTimestamp(val interface{}) int64 {
	switch v := val.(type) {
	case string:
		if t, err := dateparse.ParseAny(v); err == nil {
			return t.UnixNano()
		}
	case float64:
		if v > 1e16 {
			return int64(v)
		} else if v > 1e11 {
			return int64(v) * int64(time.Millisecond)
		}
		return int64(v) * int64(time.Second)
	case int64:
		if v > 1e16 {
			return v
		} else if v > 1e11 {
			return v * 1e6
		}
		return v * 1e9
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return parseESTimestamp(f)
		}
	}
	return 0
}
