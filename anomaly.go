package main

import (
	"crypto/md5"
	"strings"
	"time"

	go_iforest "github.com/codegaudi/go-iforest"
	"github.com/twsnmp/golof/lof"
	"github.com/twsnmp/tfidf"
	"github.com/twsnmp/tfidf/seg"
)

var oscmdKeys = []string{
	"rm%20", "cat%20", "wget%20",
	"curl%20", "sudo%20", "ssh%20",
	"usermod%20", "useradd%20", "grep%20", "ls%20",
	";", "|", "&",
	"/bin", "/dev", "/home", "/lib", "/misc", "/opt",
	"/root", "/tftpboot", "/usr", "/boot", "/etc", "/initrd",
	"/lost+found", "/mnt", "/proc", "/sbin", "/tmp", "/var",
}

var dirTraversalKeys = []string{
	"../", "..\\", ":\\",
	"/bin", "/dev", "/home", "/lib", "/misc", "/opt",
	"/root", "/tftpboot", "/usr", "/boot", "/etc/", "/initrd",
	"/lost+found", "/mnt", "/proc", "/sbin", "/tmp", "/var",
}

var sqlKeys = []string{
	"&#039", "*", ";", "%20", "--",
	"select", "delete", "create", "drop", "alter",
	"insert", "update", "set", "from", "where",
	"union", "all", "like",
	"and", "&", "or", "|",
	"user", "username", "passwd", "id", "admin", "information_schema",
}

func (b *App) setAnomalyScore(algo, vmode string, sr *SearchResult) {
	OutLog("start setAnomalyScore")
	if sr == nil || len(sr.Logs) < 1 {
		return
	}
	st := time.Now()
	var vectors [][]float64
	switch vmode {
	case "walu":
		OutLog("start make vector form Walu log")
		for _, l := range sr.Logs {
			v := getWaluVector(&l.All)
			if len(v) > 20 {
				vectors = append(vectors, v)
			} else {
				OutLog("v=%v %s", v, l.All)
			}
		}
	case "sql":
		for _, l := range sr.Logs {
			v := getKeywordsVector(&l.All, &sqlKeys)
			if len(v) == len(sqlKeys) {
				vectors = append(vectors, v)
			}
		}
	case "oscmd":
		for _, l := range sr.Logs {
			v := getKeywordsVector(&l.All, &oscmdKeys)
			if len(v) == len(oscmdKeys) {
				vectors = append(vectors, v)
			}
		}
	case "dirt":
		for _, l := range sr.Logs {
			v := getKeywordsVector(&l.All, &dirTraversalKeys)
			if len(v) == len(dirTraversalKeys) {
				vectors = append(vectors, v)
			}
		}
	case "tfidf":
		OutLog("start tfidf")
		lines := []string{}
		for _, l := range sr.Logs {
			lines = append(lines, l.All)
		}
		f := tfidf.NewTokenizer(seg.NewLogTokenizer(true))
		f.AddDocs(lines...)
		vectors = f.GetTFIDF(100, lines...)
		OutLog("tfidf lines=%d docs=%d words=%d", len(lines), f.GetDocumentCount(), len(f.GetAllTerms()))
	default:
		OutLog("start make vector form number fields")
		numKeys := []string{}
		strKeys := []string{}
		addStr := vmode == "all" || vmode == "alltime"
		for k, v := range sr.Logs[0].KeyValue {
			if _, ok := v.(float64); ok {
				numKeys = append(numKeys, string(k))
			} else if addStr {
				if _, ok := v.(string); ok {
					strKeys = append(strKeys, string(k))
				}
			}
		}
		addTime := vmode == "time" || vmode == "alltime"
		for _, l := range sr.Logs {
			vector := []float64{}
			for _, key := range numKeys {
				if f, ok := l.KeyValue[key].(float64); ok {
					vector = append(vector, f)
				} else {
					vector = append(vector, 0.0)
				}
			}
			for _, key := range strKeys {
				if s, ok := l.KeyValue[key].(string); ok {
					vector = append(vector, strToFloat(s))
				} else {
					vector = append(vector, 0.0)
				}
			}
			if addTime {
				ts := time.Unix(0, l.Time).Local()
				vector = append(vector, float64(ts.Day()))
				vector = append(vector, float64(ts.Weekday()))
				vector = append(vector, float64(ts.Hour()))
			}
			vectors = append(vectors, vector)
		}
	}
	switch algo {
	case "iforest":
		OutLog("start IForest")
		iforest, err := go_iforest.NewIForest(vectors, 1000, 256)
		if err != nil {
			OutLog("NewIForest err=%v", err)
			return
		}
		OutLog("IForest Calculate AnomalyScore")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["anomalyScore"] = iforest.CalculateAnomalyScore(v)
		}
	case "lof":
		OutLog("start LOF")
		samples := lof.GetSamplesFromFloat64s(vectors)
		lofGetter := lof.NewLOF(5)
		OutLog("LOF Train")
		if err := lofGetter.Train(samples); err != nil {
			OutLog("LOF err=%v", err)
			return
		}
		OutLog("LOF Calculate AnomalyScore")
		for i, s := range samples {
			sr.Logs[i].KeyValue["anomalyScore"] = lofGetter.GetLOF(s, "fast")
		}
	case "sum":
		OutLog("start sum")
		for i, v := range vectors {
			sum := 0.0
			for _, e := range v {
				sum += e
			}
			sr.Logs[i].KeyValue["anomalyScore"] = sum
		}
	default:
		OutLog("Other set vector")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["vector"] = v
		}
	}
	sr.AnomalyDur = time.Now().UnixMilli() - st.UnixMilli()
	OutLog("end setAnomalyScore dur=%v", time.Since(st))
}

// strToFloat : 文字列の識別するための数値を取得する、MD5の上位８バイト
func strToFloat(s string) float64 {
	var r int64
	h := md5.Sum([]byte(s))
	for i := 0; i < 8 && i < len(h); i++ {
		r *= 256
		r += int64(h[i])
	}
	return float64(r)
}

// getKeywordsVector : キーワードのりストから特徴ベクターを作成する
func getKeywordsVector(s *string, keys *[]string) []float64 {
	vector := []float64{}
	for _, k := range *keys {
		vector = append(vector, float64(strings.Count(*s, k)))
	}
	return vector
}

// https://github.com/Kanatoko/Walu
func getWaluVector(s *string) []float64 {
	vector := []float64{}
	a := strings.Split(*s, "\"")
	if len(a) < 2 {
		return vector
	}
	query := ""
	path := ""
	f := strings.Fields(a[1])
	if len(f) > 1 {
		ua := strings.SplitN(f[1], "?", 2)
		if len(ua) > 1 {
			path = ua[0]
			query = ua[1]
		}
	}

	ca := getCharCount(a[1])

	//findex_%
	vector = append(vector, float64(strings.Index(a[1], "%")))

	//findex_:
	vector = append(vector, float64(strings.Index(a[1], ":")))

	// countedCharArray
	for _, c := range []rune{':', '(', ';', '%', '/', '\'', '<', '?', '.', '#'} {
		vector = append(vector, float64(ca[c]))
	}

	//encoded =
	vector = append(vector, float64(strings.Count(a[1], "%3D")+strings.Count(a[1], "%3d")))

	//encoded /
	vector = append(vector, float64(strings.Count(a[1], "%2F")+strings.Count(a[1], "%2f")))

	//encoded \
	vector = append(vector, float64(strings.Count(a[1], "%5C")+strings.Count(a[1], "%5c")))

	//encoded %
	vector = append(vector, float64(strings.Count(a[1], "%25")))

	//%20
	vector = append(vector, float64(strings.Count(a[1], "%20")))

	//POST
	if strings.HasPrefix(a[1], "POST") {
		vector = append(vector, 1)
	} else {
		vector = append(vector, 0)
	}

	//path_nonalnum_count
	vector = append(vector, float64(len(path)-getAlphaNumCount(path)))

	//pvalue_nonalnum_avg
	vector = append(vector, float64(len(query)-getAlphaNumCount(query)))

	//non_alnum_len(max_len)
	vector = append(vector, float64(getMaxNonAlnumLength(a[1])))

	//non_alnum_count
	vector = append(vector, float64(getNonAlnumCount(a[1])))

	for _, p := range []string{"/%", "//", "/.", "..", "=/", "./", "/?"} {
		vector = append(vector, float64(strings.Count(a[1], p)))
	}
	return vector
}

func getCharCount(s string) []int {
	ret := []int{}
	for i := 0; i < 96; i++ {
		ret = append(ret, 0)
	}
	for _, c := range s {
		if 33 <= c && c <= 95 {
			ret[c] += 1
		}
	}
	return ret
}

func getAlphaNumCount(s string) int {
	ret := 0
	for _, c := range s {
		if 65 <= c && c <= 90 {
			ret++
		} else if 97 <= c && c <= 122 {
			ret++
		} else if 48 <= c && c <= 57 {
			ret++
		}
	}
	return ret
}

func getMaxNonAlnumLength(s string) int {
	max := 0
	length := 0
	for _, c := range s {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
			if length > max {
				max = length
			}
			length = 0
		} else {
			length++
		}
	}
	if max < length {
		max = length
	}
	return max
}

func getNonAlnumCount(s string) int {
	ret := 0
	for _, c := range s {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') {
		} else {
			ret++
		}
	}
	return ret
}
