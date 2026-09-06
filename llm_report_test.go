package main

import (
	"strings"
	"testing"
)

func TestBuildReportPrompt(t *testing.T) {
	req := LLMExplainReportRequest{
		ReportType: "histogram",
		Title:      "Histogram Test",
		Field:      "bytes",
		Columns:    []string{"DateTime", "bytes"},
		Data: [][]interface{}{
			{"2026-09-06 10:00:00", 1234},
			{"2026-09-06 10:01:00", 5678},
		},
		Lang: "ja",
	}

	sys, user := buildReportPrompt(req, "Japanese")

	// Verify system prompt is English-based instructions
	if !strings.Contains(sys, "You are an expert in cybersecurity") {
		t.Errorf("expected English expert intro in system prompt, got: %s", sys)
	}
	if !strings.Contains(sys, "Analyze value distributions") {
		t.Errorf("expected histogram focus in system prompt, got: %s", sys)
	}
	if !strings.Contains(sys, "strictly in Japanese") {
		t.Errorf("expected Japanese output requirement, got: %s", sys)
	}

	// Verify user prompt
	if !strings.Contains(user, "| DateTime | bytes |") {
		t.Errorf("expected markdown table header, got: %s", user)
	}
	if !strings.Contains(user, "| 2026-09-06 10:00:00 | 1234 |") {
		t.Errorf("expected data row, got: %s", user)
	}
	if !strings.Contains(user, "written in Japanese") {
		t.Errorf("expected Japanese section requirements, got: %s", user)
	}

	// Test English output
	sysEN, userEN := buildReportPrompt(req, "English")
	if !strings.Contains(sysEN, "strictly in English") {
		t.Errorf("expected English output requirement, got: %s", sysEN)
	}
	if !strings.Contains(userEN, "written in English") {
		t.Errorf("expected English section requirements, got: %s", userEN)
	}
}
