package main

import (
	"strings"
	"testing"

	"github.com/vjeantet/grok"
)

func TestParseAIGrokResponse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantGrok string
		wantTime string
		wantIP   string
		wantMAC  string
	}{
		{
			name: "Clean JSON",
			raw: `{
				"grok": "%{SYSLOGTIMESTAMP:timestamp} %{NOTSPACE:host} %{WORD:process}: %{GREEDYDATA:message}",
				"time_field": "timestamp",
				"ip_fields": "src_ip",
				"mac_fields": "mac"
			}`,
			wantGrok: "%{SYSLOGTIMESTAMP:timestamp} %{NOTSPACE:host} %{WORD:process}: %{GREEDYDATA:message}",
			wantTime: "timestamp",
			wantIP:   "src_ip",
			wantMAC:  "mac",
		},
		{
			name: "JSON wrapped in Markdown code fences",
			raw: "```json\n" + `{
				"grok": "%{TIMESTAMP_ISO8601:timestamp} %{IPV4:client_ip} %{GREEDYDATA:msg}",
				"time_field": "timestamp",
				"ip_fields": "client_ip",
				"mac_fields": ""
			}` + "\n```",
			wantGrok: "%{TIMESTAMP_ISO8601:timestamp} %{IPV4:client_ip} %{GREEDYDATA:msg}",
			wantTime: "timestamp",
			wantIP:   "client_ip",
			wantMAC:  "",
		},
		{
			name: "Markdown code block only grok pattern",
			raw: "Here is the grok pattern you requested:\n```\n%{SYSLOGTIMESTAMP:timestamp} %{WORD:user} from %{IPV4:ip}\n```\nHope this helps!",
			wantGrok: "%{SYSLOGTIMESTAMP:timestamp} %{WORD:user} from %{IPV4:ip}",
			wantTime: "",
			wantIP:   "",
			wantMAC:  "",
		},
		{
			name: "Raw grok pattern line",
			raw: "%{SYSLOGTIMESTAMP:timestamp} %{DATA:process}: Invalid user %{WORD:user} from %{IPV4:src_ip}",
			wantGrok: "%{SYSLOGTIMESTAMP:timestamp} %{DATA:process}: Invalid user %{WORD:user} from %{IPV4:src_ip}",
			wantTime: "",
			wantIP:   "",
			wantMAC:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAIGrokResponse(tt.raw)
			if got.Grok != tt.wantGrok {
				t.Errorf("parseAIGrokResponse() Grok = %v, want %v", got.Grok, tt.wantGrok)
			}
			if got.TimeField != tt.wantTime {
				t.Errorf("parseAIGrokResponse() TimeField = %v, want %v", got.TimeField, tt.wantTime)
			}
			if got.IPFields != tt.wantIP {
				t.Errorf("parseAIGrokResponse() IPFields = %v, want %v", got.IPFields, tt.wantIP)
			}
			if got.MACFields != tt.wantMAC {
				t.Errorf("parseAIGrokResponse() MACFields = %v, want %v", got.MACFields, tt.wantMAC)
			}
		})
	}
}

func TestAIGrokPatternExecution(t *testing.T) {
	pattern := `%{SYSLOGTIMESTAMP:timestamp} %{NOTSPACE:host} %{WORD:process}\[%{INT:pid}\]: Invalid user %{WORD:user} from %{IPV4:src_ip}`
	sample := "Dec 10 06:55:46 LabSZ sshd[24200]: Invalid user webmaster from 173.234.31.186"

	config := grok.Config{
		Patterns:          make(map[string]string),
		NamedCapturesOnly: true,
	}
	config.Patterns["TEST"] = strings.TrimSpace(pattern)
	g, err := grok.NewWithConfig(&config)
	if err != nil {
		t.Fatalf("Failed to create grok: %v", err)
	}

	values, err := g.Parse("%{TEST}", sample)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if values["timestamp"] != "Dec 10 06:55:46" {
		t.Errorf("timestamp = %v, want 'Dec 10 06:55:46'", values["timestamp"])
	}
	if values["user"] != "webmaster" {
		t.Errorf("user = %v, want 'webmaster'", values["user"])
	}
	if values["src_ip"] != "173.234.31.186" {
		t.Errorf("src_ip = %v, want '173.234.31.186'", values["src_ip"])
	}
}

func TestCleanAndValidateAIGrok(t *testing.T) {
	app := &App{}
	sample := "Dec 10 06:55:46 LabSZ sshd[24200]: reverse mapping checking getaddrinfo for ns.marryaldkfaczcz.com [173.234.31.186] failed - POSSIBLE BREAK-IN ATTEMPT!"

	// Case 1: Raw log string returned from LLM (exactly what happened in user's report)
	res := AIGrokResp{
		Grok:      sample,
		TimeField: "timestamp",
		IPFields:  "ns.marryaldkfaczcz.com",
		MACFields: "173.234.31.186",
	}
	app.cleanAndValidateAIGrok(&res, sample)

	if res.Grok == sample || !strings.Contains(res.Grok, "%{") {
		t.Errorf("Expected Grok pattern with %%{...}, got raw log: %s", res.Grok)
	}
	if strings.Contains(res.IPFields, "marryaldkfaczcz") {
		t.Errorf("IPFields should not contain domain value: %s", res.IPFields)
	}
	if strings.Contains(res.MACFields, "173.234.31.186") {
		t.Errorf("MACFields should not contain IP value: %s", res.MACFields)
	}

	// Verify the resulting pattern parses the sample without error
	testRes := app.TestGrok(res.Grok, sample)
	if testRes.ErrorMsg != "" {
		t.Errorf("TestGrok returned error for cleaned grok: %v", testRes.ErrorMsg)
	}
	if len(testRes.Data) == 0 {
		t.Errorf("TestGrok should have matched sample data")
	}
}

