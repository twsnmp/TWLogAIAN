package main

import (
	"testing"
)

func TestKNNDetector(t *testing.T) {
	// 5 regular points around (1, 1), 1 anomaly at (10, 10)
	normal := [][]float64{
		{1.0, 1.0},
		{1.1, 0.9},
		{0.9, 1.1},
		{1.0, 1.2},
		{1.2, 1.0},
	}
	anomaly := []float64{10.0, 10.0}

	knn := NewKNNDetector(3)
	if err := knn.Fit(normal); err != nil {
		t.Fatalf("KNN Fit failed: %v", err)
	}

	normScore := knn.Score([]float64{1.0, 1.05})
	anomScore := knn.Score(anomaly)

	if anomScore <= normScore {
		t.Errorf("anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
	}

	all := append(normal, anomaly)
	knnAll := NewKNNDetector(3)
	_ = knnAll.Fit(all)
	batchScores := knnAll.ScoreBatch(all)
	if len(batchScores) != len(all) {
		t.Fatalf("expected %d batch scores, got %d", len(all), len(batchScores))
	}
	lastScore := batchScores[len(batchScores)-1]
	for i := 0; i < len(normal); i++ {
		if batchScores[i] >= lastScore {
			t.Errorf("normal point %d score (%f) >= anomaly score (%f)", i, batchScores[i], lastScore)
		}
	}
}

func TestMahalanobisDetector(t *testing.T) {
	// Points correlated along y = 2x
	vectors := [][]float64{
		{1.0, 2.0},
		{2.0, 4.0},
		{3.0, 6.0},
		{4.0, 8.0},
		{5.0, 10.0},
		{6.0, 12.0},
		{7.0, 14.0},
	}
	md := NewMahalanobisDetector()
	if err := md.Fit(vectors); err != nil {
		t.Fatalf("Mahalanobis Fit failed: %v", err)
	}

	// Normal point on trend line
	normScore := md.Score([]float64{4.0, 8.0})
	// Anomaly breaking the trend
	anomScore := md.Score([]float64{4.0, 1.0})

	if anomScore <= normScore {
		t.Errorf("anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
	}
}

func TestStatDetector(t *testing.T) {
	vectors := [][]float64{
		{10.0, 100.0},
		{10.5, 101.0},
		{9.5, 99.0},
		{10.2, 100.5},
		{9.8, 99.5},
	}
	sd := NewStatDetector()
	if err := sd.Fit(vectors); err != nil {
		t.Fatalf("Stat Fit failed: %v", err)
	}

	normScore := sd.Score([]float64{10.1, 100.2})
	anomScore := sd.Score([]float64{50.0, 500.0})

	if anomScore <= normScore {
		t.Errorf("anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
	}
}

func TestSetAnomalyScoreApp(t *testing.T) {
	app := &App{}
	sr := &SearchResult{
		Logs: []*LogEnt{
			{KeyValue: map[string]interface{}{"val": 1.0, "val2": 2.0}},
			{KeyValue: map[string]interface{}{"val": 1.1, "val2": 2.1}},
			{KeyValue: map[string]interface{}{"val": 0.9, "val2": 1.9}},
			{KeyValue: map[string]interface{}{"val": 1.05, "val2": 2.05}},
			{KeyValue: map[string]interface{}{"val": 10.0, "val2": 20.0}}, // Outlier
		},
	}

	for _, algo := range []string{"knn", "mahalanobis", "zscore", "autoencoder", "lstm"} {
		app.setAnomalyScore(algo, "", sr)
		outlierScore, ok := sr.Logs[4].KeyValue["anomalyScore"].(float64)
		if !ok {
			t.Fatalf("algo %s did not set anomalyScore", algo)
		}
		normalScore, _ := sr.Logs[0].KeyValue["anomalyScore"].(float64)
		if outlierScore <= normalScore && algo != "lstm" { // lstm is sequence-based
			t.Errorf("algo %s: outlier score %f <= normal score %f", algo, outlierScore, normalScore)
		}
	}
}

func TestAutoencoderDetector(t *testing.T) {
	vectors := [][]float64{
		{1.0, 2.0, 3.0, 4.0},
		{1.1, 2.1, 3.1, 4.1},
		{0.9, 1.9, 2.9, 3.9},
		{1.05, 2.05, 3.05, 4.05},
		{1.02, 2.02, 3.02, 4.02},
	}
	ae := NewAutoencoderDetector()
	if err := ae.Fit(vectors); err != nil {
		t.Fatalf("AE Fit failed: %v", err)
	}

	normScore := ae.Score([]float64{1.0, 2.0, 3.0, 4.0})
	anomScore := ae.Score([]float64{50.0, 100.0, 150.0, 200.0})

	if anomScore <= normScore {
		t.Errorf("AE anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
	}
}

func TestLSTMDetector(t *testing.T) {
	vectors := [][]float64{
		{1.0, 1.0},
		{2.0, 2.0},
		{3.0, 3.0},
		{4.0, 4.0},
		{5.0, 5.0},
		{6.0, 6.0},
	}
	lstm := NewLSTMDetector()
	if err := lstm.Fit(vectors); err != nil {
		t.Fatalf("LSTM Fit failed: %v", err)
	}

	// Score calls sequentially
	_ = lstm.Score([]float64{1.0, 1.0})
	scoreNextExpected := lstm.Score([]float64{2.0, 2.0}) // normal step
	scoreUnexpected := lstm.Score([]float64{99.0, 99.0}) // sudden jump

	if scoreUnexpected <= scoreNextExpected {
		t.Errorf("LSTM unexpected jump score (%f) should be higher than expected step score (%f)", scoreUnexpected, scoreNextExpected)
	}
}
