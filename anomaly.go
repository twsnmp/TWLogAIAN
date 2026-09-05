package main

import (
	"crypto/md5"
	"errors"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	go_iforest "github.com/codegaudi/go-iforest"
	tensai "github.com/mattn/tensai"
	"github.com/mattn/tensai/layer"
	"github.com/mattn/tensai/loss"
	"github.com/mattn/tensai/model"
	"github.com/mattn/tensai/optim"
	aitensai "github.com/twsnmp/TWLogAIAN/pkg/ai/tensai"
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
	case "knn":
		OutLog("start KNN")
		knn := NewKNNDetector(5)
		if err := knn.Fit(vectors); err != nil {
			OutLog("KNN err=%v", err)
			return
		}
		OutLog("KNN Calculate AnomalyScore")
		scores := knn.ScoreBatch(vectors)
		for i, s := range scores {
			sr.Logs[i].KeyValue["anomalyScore"] = s
		}
	case "mahalanobis":
		OutLog("start Mahalanobis")
		md := NewMahalanobisDetector()
		if err := md.Fit(vectors); err != nil {
			OutLog("Mahalanobis err=%v", err)
			return
		}
		OutLog("Mahalanobis Calculate AnomalyScore")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["anomalyScore"] = md.Score(v)
		}
	case "zscore", "stat":
		OutLog("start Z-Score")
		stat := NewStatDetector()
		if err := stat.Fit(vectors); err != nil {
			OutLog("Stat err=%v", err)
			return
		}
		OutLog("Z-Score Calculate AnomalyScore")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["anomalyScore"] = stat.Score(v)
		}
	case "autoencoder", "ae":
		OutLog("start Autoencoder")
		ae := NewAutoencoderDetector()
		if err := ae.Fit(vectors); err != nil {
			OutLog("Autoencoder err=%v", err)
			return
		}
		OutLog("Autoencoder Calculate AnomalyScore")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["anomalyScore"] = ae.Score(v)
		}
	case "lstm", "rnn":
		OutLog("start LSTM")
		lstm := NewLSTMDetector()
		if err := lstm.Fit(vectors); err != nil {
			OutLog("LSTM err=%v", err)
			return
		}
		OutLog("LSTM Calculate AnomalyScore")
		for i, v := range vectors {
			sr.Logs[i].KeyValue["anomalyScore"] = lstm.Score(v)
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

// KNNDetector computes anomaly score based on k-nearest neighbor distance
type KNNDetector struct {
	k         int
	trainData [][]float64
}

// NewKNNDetector creates a new k-NN distance anomaly detector
func NewKNNDetector(k int) *KNNDetector {
	if k < 1 {
		k = 5
	}
	return &KNNDetector{k: k}
}

// Fit stores training feature vectors
func (d *KNNDetector) Fit(vectors [][]float64) error {
	if len(vectors) == 0 {
		return errors.New("empty vectors")
	}
	n := len(vectors)
	dim := len(vectors[0])
	cleaned := make([][]float64, n)
	for i, v := range vectors {
		cv := make([]float64, dim)
		for j := 0; j < dim && j < len(v); j++ {
			if !math.IsNaN(v[j]) && !math.IsInf(v[j], 0) {
				cv[j] = v[j]
			}
		}
		cleaned[i] = cv
	}

	d.trainData = cleaned
	if d.k >= len(cleaned) {
		d.k = max(1, len(cleaned)-1)
	}
	return nil
}

// Score computes the mean distance to the k-nearest neighbors for a single vector
func (d *KNNDetector) Score(vector []float64) float64 {
	n := len(d.trainData)
	if n == 0 {
		return 0.0
	}
	k := d.k
	if k > n {
		k = n
	}

	dists := make([]float64, n)
	for i, v := range d.trainData {
		dists[i] = safeEuclideanDist(vector, v)
	}
	sort.Float64s(dists)

	var sumDist float64
	for i := 0; i < k; i++ {
		sumDist += dists[i]
	}
	return sumDist / float64(k)
}

// ScoreBatch computes anomaly scores for multiple vectors using parallel worker routines
func (d *KNNDetector) ScoreBatch(vectors [][]float64) []float64 {
	n := len(vectors)
	if n == 0 {
		return []float64{}
	}
	trainN := len(d.trainData)
	if trainN == 0 {
		return make([]float64, n)
	}

	k := d.k
	if k > trainN {
		k = trainN
	}

	scores := make([]float64, n)
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	chunkSize := (n + workers - 1) / workers

	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}
		if start >= end {
			continue
		}

		wg.Add(1)
		go func(rStart, rEnd int) {
			defer wg.Done()
			for i := rStart; i < rEnd; i++ {
				vec := vectors[i]
				dists := make([]float64, trainN)
				for j, tv := range d.trainData {
					dists[j] = safeEuclideanDist(vec, tv)
				}
				sort.Float64s(dists)

				// If evaluating against own training set, dists[0] is 0 (self distance)
				startIdx := 0
				if trainN == n && dists[0] < 1e-12 {
					startIdx = 1
				}
				var sumDist float64
				count := 0
				for idx := startIdx; count < k && idx < trainN; idx++ {
					sumDist += dists[idx]
					count++
				}
				if count > 0 {
					scores[i] = sumDist / float64(count)
				}
			}
		}(start, end)
	}
	wg.Wait()
	return scores
}

func safeEuclideanDist(a, b []float64) float64 {
	var sum float64
	dim := min(len(a), len(b))
	for i := 0; i < dim; i++ {
		va, vb := 0.0, 0.0
		if !math.IsNaN(a[i]) && !math.IsInf(a[i], 0) {
			va = a[i]
		}
		if !math.IsNaN(b[i]) && !math.IsInf(b[i], 0) {
			vb = b[i]
		}
		d := va - vb
		sum += d * d
	}
	return math.Sqrt(sum)
}

// MahalanobisDetector detects anomalies using Mahalanobis distance
type MahalanobisDetector struct {
	mean      []float64
	invCov    [][]float64
	invStdDev []float64
	dim       int
	useDiag   bool
}

// NewMahalanobisDetector creates a Mahalanobis distance anomaly detector
func NewMahalanobisDetector() *MahalanobisDetector {
	return &MahalanobisDetector{}
}

// Fit computes the mean vector and inverted covariance matrix
func (d *MahalanobisDetector) Fit(vectors [][]float64) error {
	n := len(vectors)
	if n == 0 {
		return errors.New("empty vectors")
	}
	dim := len(vectors[0])
	if dim == 0 {
		return errors.New("zero feature dimension")
	}
	d.dim = dim

	cleaned := make([][]float64, n)
	for i, v := range vectors {
		cv := make([]float64, dim)
		for j := 0; j < dim && j < len(v); j++ {
			if !math.IsNaN(v[j]) && !math.IsInf(v[j], 0) {
				cv[j] = v[j]
			}
		}
		cleaned[i] = cv
	}

	d.mean = make([]float64, dim)
	for _, v := range cleaned {
		for j := 0; j < dim; j++ {
			d.mean[j] += v[j]
		}
	}
	for j := 0; j < dim; j++ {
		d.mean[j] /= float64(n)
	}

	// For high dimensions (> 64) or insufficient samples, use diagonal variance
	if dim > 64 || dim >= n {
		d.useDiag = true
		d.invStdDev = make([]float64, dim)
		denom := float64(max(1, n-1))
		for _, v := range cleaned {
			for j := 0; j < dim; j++ {
				diff := v[j] - d.mean[j]
				d.invStdDev[j] += diff * diff
			}
		}
		for j := 0; j < dim; j++ {
			variance := (d.invStdDev[j] / denom) + 1e-4
			d.invStdDev[j] = 1.0 / math.Sqrt(variance)
		}
		return nil
	}

	d.useDiag = false
	cov := make([][]float64, dim)
	for i := range cov {
		cov[i] = make([]float64, dim)
	}

	for _, v := range cleaned {
		for i := 0; i < dim; i++ {
			diffI := v[i] - d.mean[i]
			for j := 0; j < dim; j++ {
				diffJ := v[j] - d.mean[j]
				cov[i][j] += diffI * diffJ
			}
		}
	}

	denom := float64(max(1, n-1))
	for i := 0; i < dim; i++ {
		for j := 0; j < dim; j++ {
			cov[i][j] /= denom
		}
		cov[i][i] += 1e-4 // regularization
	}

	inv, err := invertMatrix(cov, dim)
	if err != nil {
		d.useDiag = true
		d.invStdDev = make([]float64, dim)
		for j := 0; j < dim; j++ {
			d.invStdDev[j] = 1.0 / math.Sqrt(cov[j][j])
		}
		return nil
	}
	d.invCov = inv
	return nil
}

// Score computes the Mahalanobis distance from the mean vector
func (d *MahalanobisDetector) Score(vector []float64) float64 {
	if len(d.mean) == 0 {
		return 0.0
	}
	dim := d.dim
	diff := make([]float64, dim)
	for i := 0; i < dim; i++ {
		val := 0.0
		if i < len(vector) && !math.IsNaN(vector[i]) && !math.IsInf(vector[i], 0) {
			val = vector[i]
		}
		diff[i] = val - d.mean[i]
	}

	if d.useDiag {
		var sumSq float64
		for i := 0; i < dim; i++ {
			z := diff[i] * d.invStdDev[i]
			sumSq += z * z
		}
		return math.Sqrt(sumSq)
	}

	var distSq float64
	for i := 0; i < dim; i++ {
		var rowSum float64
		for j := 0; j < dim; j++ {
			rowSum += diff[j] * d.invCov[j][i]
		}
		distSq += diff[i] * rowSum
	}
	if distSq < 0 || math.IsNaN(distSq) {
		return 0.0
	}
	return math.Sqrt(distSq)
}

func invertMatrix(a [][]float64, n int) ([][]float64, error) {
	aug := make([][]float64, n)
	for i := 0; i < n; i++ {
		aug[i] = make([]float64, 2*n)
		for j := 0; j < n; j++ {
			aug[i][j] = a[i][j]
		}
		aug[i][n+i] = 1.0
	}

	for i := 0; i < n; i++ {
		maxRow := i
		maxVal := math.Abs(aug[i][i])
		for k := i + 1; k < n; k++ {
			if math.Abs(aug[k][i]) > maxVal {
				maxVal = math.Abs(aug[k][i])
				maxRow = k
			}
		}
		if maxVal < 1e-12 {
			return nil, errors.New("singular matrix")
		}
		aug[i], aug[maxRow] = aug[maxRow], aug[i]

		pivot := aug[i][i]
		for j := 0; j < 2*n; j++ {
			aug[i][j] /= pivot
		}

		for k := 0; k < n; k++ {
			if k != i {
				factor := aug[k][i]
				for j := 0; j < 2*n; j++ {
					aug[k][j] -= factor * aug[i][j]
				}
			}
		}
	}

	inv := make([][]float64, n)
	for i := 0; i < n; i++ {
		inv[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			inv[i][j] = aug[i][n+j]
		}
	}
	return inv, nil
}

// StatDetector detects anomalies using statistical Z-score
type StatDetector struct {
	means []float64
	stds  []float64
	dim   int
}

// NewStatDetector creates a Z-score statistical anomaly detector
func NewStatDetector() *StatDetector {
	return &StatDetector{}
}

// Fit computes the mean and standard deviation for each feature dimension
func (d *StatDetector) Fit(vectors [][]float64) error {
	n := len(vectors)
	if n == 0 {
		return errors.New("empty vectors")
	}
	dim := len(vectors[0])
	if dim == 0 {
		return errors.New("zero feature dimension")
	}
	d.dim = dim

	cleaned := make([][]float64, n)
	for i, v := range vectors {
		cv := make([]float64, dim)
		for j := 0; j < dim && j < len(v); j++ {
			if !math.IsNaN(v[j]) && !math.IsInf(v[j], 0) {
				cv[j] = v[j]
			}
		}
		cleaned[i] = cv
	}

	d.means = make([]float64, dim)
	d.stds = make([]float64, dim)

	for _, v := range cleaned {
		for j := 0; j < dim; j++ {
			d.means[j] += v[j]
		}
	}
	for j := 0; j < dim; j++ {
		d.means[j] /= float64(n)
	}

	for _, v := range cleaned {
		for j := 0; j < dim; j++ {
			diff := v[j] - d.means[j]
			d.stds[j] += diff * diff
		}
	}
	for j := 0; j < dim; j++ {
		variance := d.stds[j] / float64(max(1, n-1))
		if variance < 1e-8 {
			d.stds[j] = 1.0
		} else {
			d.stds[j] = math.Sqrt(variance)
		}
	}

	return nil
}

// Score computes the Euclidean/RMS Z-score across feature dimensions
func (d *StatDetector) Score(vector []float64) float64 {
	if len(d.means) == 0 {
		return 0.0
	}
	dim := d.dim
	var sumSq float64
	for j := 0; j < dim; j++ {
		val := 0.0
		if j < len(vector) && !math.IsNaN(vector[j]) && !math.IsInf(vector[j], 0) {
			val = vector[j]
		}
		z := (val - d.means[j]) / d.stds[j]
		sumSq += z * z
	}
	score := math.Sqrt(sumSq / float64(dim))
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0.0
	}
	return score
}

// AutoencoderDetector implements neural network Autoencoder anomaly detection using tensai
type AutoencoderDetector struct {
	epochs     int
	lr         float32
	net        *model.Sequential
	dim        int
	means      []float64
	stds       []float64
	normalized bool
}

// NewAutoencoderDetector creates an Autoencoder detector
func NewAutoencoderDetector() *AutoencoderDetector {
	return &AutoencoderDetector{
		epochs: 30,
		lr:     0.01,
	}
}

// Fit trains the Autoencoder to reconstruct normal feature vectors
func (d *AutoencoderDetector) Fit(vectors [][]float64) error {
	if len(vectors) == 0 {
		return errors.New("empty vectors")
	}
	rows := len(vectors)
	cols := len(vectors[0])
	if cols == 0 {
		return errors.New("zero feature dimension")
	}
	d.dim = cols

	if dev, err := aitensai.GetGPUDevice(); err == nil && dev != nil {
		tensai.UseAccelerator(dev)
	}

	cleaned := make([][]float64, rows)
	for i, v := range vectors {
		cv := make([]float64, cols)
		for j := 0; j < cols && j < len(v); j++ {
			if !math.IsNaN(v[j]) && !math.IsInf(v[j], 0) {
				cv[j] = v[j]
			}
		}
		cleaned[i] = cv
	}

	// Feature normalization
	d.means = make([]float64, cols)
	d.stds = make([]float64, cols)
	for _, v := range cleaned {
		for j, val := range v {
			d.means[j] += val
		}
	}
	for j := range d.means {
		d.means[j] /= float64(rows)
	}
	for _, v := range cleaned {
		for j, val := range v {
			diff := val - d.means[j]
			d.stds[j] += diff * diff
		}
	}
	for j := range d.stds {
		variance := d.stds[j] / float64(rows)
		if variance < 1e-8 {
			d.stds[j] = 1.0
		} else {
			d.stds[j] = math.Sqrt(variance)
		}
	}
	d.normalized = true

	// Build matrix data
	data := make([]tensai.Float, rows*cols)
	for i, v := range cleaned {
		for j, val := range v {
			normVal := (val - d.means[j]) / d.stds[j]
			data[i*cols+j] = tensai.Float(normVal)
		}
	}
	mat, err := tensai.NewMatrixFromSlice(rows, cols, data)
	if err != nil {
		return err
	}

	bottleneck := cols / 4
	if bottleneck < 2 {
		bottleneck = 2
	}
	if bottleneck > 32 {
		bottleneck = 32
	}

	net := model.NewSequential()
	net.Add(layer.NewDense(bottleneck))
	net.Add(&layer.Tanh{})
	net.Add(layer.NewDense(cols))

	if err := net.Compile(cols, loss.MeanSquaredError{}, optim.NewAdam(d.lr)); err != nil {
		return err
	}

	epochs := d.epochs
	if rows > 1000 {
		epochs = 20
	}
	for e := 1; e <= epochs; e++ {
		if _, err := net.FitStep(mat, mat); err != nil {
			return err
		}
	}

	d.net = net
	return nil
}

// Score calculates the reconstruction error (MSE)
func (d *AutoencoderDetector) Score(vector []float64) float64 {
	if d.net == nil || len(vector) != d.dim {
		return 0.0
	}

	normVec := make([]tensai.Float, d.dim)
	for j := 0; j < d.dim; j++ {
		val := 0.0
		if j < len(vector) && !math.IsNaN(vector[j]) && !math.IsInf(vector[j], 0) {
			val = vector[j]
		}
		if d.normalized {
			val = (val - d.means[j]) / d.stds[j]
		}
		normVec[j] = tensai.Float(val)
	}

	inMat, err := tensai.NewMatrixFromSlice(1, d.dim, normVec)
	if err != nil {
		return 0.0
	}

	outMat, err := d.net.Predict(inMat)
	if err != nil {
		return 0.0
	}

	var mse float64
	for j := 0; j < d.dim; j++ {
		diff := float64(normVec[j] - outMat.Data[j])
		mse += diff * diff
	}
	score := mse / float64(d.dim)
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0.0
	}
	return score
}

// LSTMDetector detects sequential transition anomalies
type LSTMDetector struct {
	net        *model.Sequential
	dim        int
	prevVector []float64
}

// NewLSTMDetector creates an LSTM/Sequence transition anomaly detector
func NewLSTMDetector() *LSTMDetector {
	return &LSTMDetector{}
}

// Fit trains the transition predictor model (predict X_{t+1} from X_t)
func (d *LSTMDetector) Fit(vectors [][]float64) error {
	if len(vectors) < 2 {
		return errors.New("need at least 2 vectors for sequence learning")
	}
	dim := len(vectors[0])
	if dim == 0 {
		return errors.New("zero feature dimension")
	}
	d.dim = dim

	if dev, err := aitensai.GetGPUDevice(); err == nil && dev != nil {
		tensai.UseAccelerator(dev)
	}

	// Sample vectors if dataset is large
	maxSamples := 300
	step := 1
	if len(vectors) > maxSamples {
		step = len(vectors) / maxSamples
		if step < 1 {
			step = 1
		}
	}

	var sampled [][]float64
	for i := 0; i < len(vectors); i += step {
		cv := make([]float64, dim)
		for j := 0; j < dim && j < len(vectors[i]); j++ {
			if !math.IsNaN(vectors[i][j]) && !math.IsInf(vectors[i][j], 0) {
				cv[j] = vectors[i][j]
			}
		}
		sampled = append(sampled, cv)
	}

	samples := len(sampled) - 1
	if samples < 1 {
		return errors.New("not enough sampled vectors")
	}

	inData := make([]tensai.Float, samples*dim)
	tgtData := make([]tensai.Float, samples*dim)
	for i := 0; i < samples; i++ {
		for j := 0; j < dim; j++ {
			inData[i*dim+j] = tensai.Float(sampled[i][j])
			tgtData[i*dim+j] = tensai.Float(sampled[i+1][j])
		}
	}

	inMat, err := tensai.NewMatrixFromSlice(samples, dim, inData)
	if err != nil {
		return err
	}
	tgtMat, err := tensai.NewMatrixFromSlice(samples, dim, tgtData)
	if err != nil {
		return err
	}

	hidden := dim / 4
	if hidden < 4 {
		hidden = 4
	}
	if hidden > 32 {
		hidden = 32
	}

	net := model.NewSequential()
	net.Add(layer.NewDense(hidden))
	net.Add(&layer.Tanh{})
	net.Add(layer.NewDense(dim))

	if err := net.Compile(dim, loss.MeanSquaredError{}, optim.NewAdam(0.01)); err != nil {
		return err
	}

	for e := 1; e <= 20; e++ {
		if _, err := net.FitStep(inMat, tgtMat); err != nil {
			return err
		}
	}

	d.net = net
	return nil
}

// Score predicts the current vector from the previous vector and returns the prediction error
func (d *LSTMDetector) Score(vector []float64) float64 {
	if d.net == nil || len(vector) != d.dim {
		return 0.0
	}
	if d.prevVector == nil {
		d.prevVector = vector
		return 0.0
	}

	prevMat, err := tensai.NewMatrixFromSlice(1, d.dim, toTensaiFloats(d.prevVector))
	if err != nil {
		d.prevVector = vector
		return 0.0
	}

	predMat, err := d.net.Predict(prevMat)
	if err != nil {
		d.prevVector = vector
		return 0.0
	}

	var mse float64
	for j := 0; j < d.dim; j++ {
		diff := float64(tensai.Float(vector[j]) - predMat.Data[j])
		mse += diff * diff
	}
	d.prevVector = vector

	score := mse / float64(d.dim)
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0.0
	}
	return score
}

func toTensaiFloats(v []float64) []tensai.Float {
	out := make([]tensai.Float, len(v))
	for i, val := range v {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			out[i] = 0.0
		} else {
			out[i] = tensai.Float(val)
		}
	}
	return out
}


