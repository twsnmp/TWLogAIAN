package main

import (
	"testing"
	"time"
)

func TestApp_GetSigmaPacks(t *testing.T) {
	app := NewApp()
	packs := app.GetSigmaPacks()
	if len(packs) < 10 {
		t.Fatalf("expected at least 10 packs, got %d", len(packs))
	}
	foundCompliance := false
	for _, p := range packs {
		if p.Name == "wazuh-compliance" {
			foundCompliance = true
			if p.RuleCount == 0 {
				t.Errorf("expected wazuh-compliance to have rules, got %d", p.RuleCount)
			}
		}
	}
	if !foundCompliance {
		t.Errorf("wazuh-compliance pack not found")
	}
}

func TestApp_AnalyzeSigmaLogs(t *testing.T) {
	app := NewApp()
	testLogs := []*LogEnt{
		{
			Time: time.Now().UnixNano(),
			All:  `{"Event":{"System":{"Channel":"Security","Computer":"WIN-DC","EventID":4625,"Level":0},"EventData":{"TargetUserName":"admin","WorkstationName":"DESKTOP-1"}}}`,
		},
		{
			Time: time.Now().Add(time.Second).UnixNano(),
			All:  `{"hostname":"server1","content":"Failed password for invalid user admin from 192.168.1.100 port 45678 ssh2"}`,
		},
		{
			Time: time.Now().Add(2 * time.Second).UnixNano(),
			All:  `192.168.1.50 - - [10/Sep/2026:10:00:00 +0000] "GET /?id=1 UNION SELECT null,username,password FROM users HTTP/1.1" 200 4523`,
		},
		{
			Time: time.Now().Add(3 * time.Second).UnixNano(),
			All:  `{"hostname":"server1","content":"systemd-journald: Stopped"}`,
		},
	}

	res, err := app.AnalyzeSigmaLogs(testLogs, nil, "")
	if err != nil {
		t.Fatalf("AnalyzeSigmaLogs error: %v", err)
	}

	if res.Stats.TotalLogs != 4 {
		t.Errorf("expected 4 total logs, got %d", res.Stats.TotalLogs)
	}
	if res.Stats.HitLogs < 3 {
		t.Errorf("expected at least 3 hit logs, got %d", res.Stats.HitLogs)
	}
	if len(res.Items) < 3 {
		t.Errorf("expected at least 3 detected items, got %d", len(res.Items))
	}
	if res.Stats.ComplianceHits == 0 {
		t.Errorf("expected at least 1 compliance hit, got %d", res.Stats.ComplianceHits)
	}
}
