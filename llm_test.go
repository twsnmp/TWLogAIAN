package main

import (
	"strings"
	"testing"
)

func TestMaskPII(t *testing.T) {
	sampleLog := "Dec 10 06:55:46 LabSZ sshd[24200]: reverse mapping checking getaddrinfo for ns.marryaldkfaczcz.com [173.234.31.186] failed - POSSIBLE BREAK-IN ATTEMPT! user=admin password=secret123 from user testuser with email test@example.com mac 00:11:22:33:44:55, duplicate ip 173.234.31.186"

	masked := MaskPII(sampleLog)

	if strings.Contains(masked, "173.234.31.186") {
		t.Errorf("expected IP to be masked, got %s", masked)
	}
	if !strings.Contains(masked, "[IP_1]") {
		t.Errorf("expected [IP_1] in masked output, got %s", masked)
	}
	// Check consistent replacement for duplicate IP
	if strings.Count(masked, "[IP_1]") != 2 {
		t.Errorf("expected [IP_1] to appear twice for identical IP, got %s", masked)
	}
	if strings.Contains(masked, "secret123") {
		t.Errorf("expected password to be masked, got %s", masked)
	}
	if !strings.Contains(masked, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in masked output, got %s", masked)
	}
	if strings.Contains(masked, "test@example.com") {
		t.Errorf("expected email to be masked, got %s", masked)
	}
	if !strings.Contains(masked, "[EMAIL_1]") {
		t.Errorf("expected [EMAIL_1] in masked output, got %s", masked)
	}
	if strings.Contains(masked, "00:11:22:33:44:55") {
		t.Errorf("expected MAC to be masked, got %s", masked)
	}
	if !strings.Contains(masked, "[MAC_1]") {
		t.Errorf("expected [MAC_1] in masked output, got %s", masked)
	}
	if strings.Contains(masked, "ns.marryaldkfaczcz.com") {
		t.Errorf("expected domain to be masked, got %s", masked)
	}
	if !strings.Contains(masked, "[HOST_1]") {
		t.Errorf("expected [HOST_1] in masked output, got %s", masked)
	}
}

func TestNormalizeLog(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "2026-09-05 14:00:00 server error connection refused from 192.168.1.50:8080",
			want:  "#TIMESTAMP# server error connection refused from #IP#:#NUM#",
		},
		{
			input: "User test@example.com with uuid 12345678-1234-1234-1234-123456789abc and mac aa:bb:cc:dd:ee:ff failed 3 times",
			want:  "User #EMAIL# with uuid #UUID# and mac #MAC# failed #NUM# times",
		},
	}

	for _, tt := range tests {
		got := normalizeLog(tt.input)
		if got != tt.want {
			t.Errorf("normalizeLog(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAnalyzeLogsAggregation(t *testing.T) {
	app := &App{}
	logs := []*LogEnt{
		{
			Time: 1700000000000000000,
			All:  "2026-09-05 10:00:00 [ERROR] failed login for user admin from 10.0.0.1",
			KeyValue: map[string]interface{}{
				"severity_str": "ERROR",
			},
		},
		{
			Time: 1700000060000000000,
			All:  "2026-09-05 10:01:00 [ERROR] failed login for user admin from 10.0.0.2",
			KeyValue: map[string]interface{}{
				"severity_str": "ERROR",
			},
		},
		{
			Time: 1700000120000000000,
			All:  "2026-09-05 10:02:00 [WARN] disk space low 92%",
			KeyValue: map[string]interface{}{
				"severity_str": "WARN",
			},
		},
		{
			Time: 1700000180000000000,
			All:  "2026-09-05 10:03:00 [INFO] backup completed successfully",
			KeyValue: map[string]interface{}{
				"severity_str": "INFO",
			},
		},
	}

	req := AIAnalysisReportRequest{
		Lang:       "ja",
		MaskPII:    true,
		TopNError:  5,
		SampleSize: 10,
	}

	// Test without LLM configured (should aggregate correctly and return error about LLM provider)
	resp := app.AnalyzeLogsWithAI(req, logs)
	if resp.Summary.TotalEntries != 4 {
		t.Errorf("expected 4 total entries, got %d", resp.Summary.TotalEntries)
	}
	if resp.Summary.ErrorCount != 2 {
		t.Errorf("expected 2 error count, got %d", resp.Summary.ErrorCount)
	}
	if resp.Summary.WarningCount != 1 {
		t.Errorf("expected 1 warning count, got %d", resp.Summary.WarningCount)
	}
	if len(resp.Summary.TopErrors) != 1 {
		t.Fatalf("expected 1 normalized error pattern group, got %d", len(resp.Summary.TopErrors))
	}
	if resp.Summary.TopErrors[0].Count != 2 {
		t.Errorf("expected pattern count 2, got %d", resp.Summary.TopErrors[0].Count)
	}
}

