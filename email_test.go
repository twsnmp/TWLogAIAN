package main

import (
	"testing"
	"time"
)

func TestAnalyzeEmailLogs(t *testing.T) {
	app := &App{}

	rawMail := `From: "Sender User" <sender@example.com>
To: "Recipient User" <recipient@example.org>
Subject: Test Subject Email
Date: Mon, 01 Jan 2024 10:00:00 +0000
Return-Path: <sender@example.com>
Received: from mail.example.com (mail.example.com [198.51.100.1])
	by mx.example.org with ESMTP id ABC12345
	for <recipient@example.org>; Mon, 01 Jan 2024 10:00:05 +0000

This is the body.
`

	logs := []*LogEnt{
		{
			ID:   "mail1",
			Time: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).UnixNano(),
			All:  rawMail,
			KeyValue: map[string]interface{}{
				"from": "sender@example.com",
				"to":   "recipient@example.org",
			},
		},
	}

	res, err := app.AnalyzeEmailLogs(logs, false)
	if err != nil {
		t.Fatalf("AnalyzeEmailLogs failed: %v", err)
	}
	if res == nil {
		t.Fatalf("AnalyzeEmailLogs returned nil result")
	}

	if res.Stats.TotalEmails != 1 {
		t.Errorf("expected 1 email, got %d", res.Stats.TotalEmails)
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res.Items))
	}

	item := res.Items[0]
	if item.Subject != "Test Subject Email" {
		t.Errorf("expected Subject 'Test Subject Email', got '%s'", item.Subject)
	}
	if item.Domain != "example.com" {
		t.Errorf("expected Domain 'example.com', got '%s'", item.Domain)
	}
	if item.Delay != 5 {
		t.Errorf("expected Delay 5s, got %d", item.Delay)
	}
	if item.Relay != 1 {
		t.Errorf("expected Relay 1, got %d", item.Relay)
	}
	if item.SenderIP != "198.51.100.1" {
		t.Errorf("expected SenderIP 198.51.100.1, got '%s'", item.SenderIP)
	}
}
