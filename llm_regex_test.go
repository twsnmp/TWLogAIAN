package main

import (
	"regexp"
	"testing"
)

func TestParseAIRegexResponse(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantRegex   string
		wantExplain string
	}{
		{
			name: "Clean JSON",
			raw: `{
				"regex": "(?i)sshd.*(failed|invalid)",
				"explain": "Matches failed SSH logins"
			}`,
			wantRegex:   "(?i)sshd.*(failed|invalid)",
			wantExplain: "Matches failed SSH logins",
		},
		{
			name: "JSON wrapped in Markdown code fences",
			raw: "```json\n" + `{
				"regex": "\"\\s+5[0-9]{2}\\s+",
				"explain": "HTTP 5xx server errors"
			}` + "\n```",
			wantRegex:   "\"\\s+5[0-9]{2}\\s+",
			wantExplain: "HTTP 5xx server errors",
		},
		{
			name:        "Raw regex string",
			raw:         `(?i)error|critical`,
			wantRegex:   `(?i)error|critical`,
			wantExplain: "",
		},
		{
			name: "JSON embedded in conversational text",
			raw: "Here is the regex you asked for:\n" + `{
				"regex": "192\\.168\\.1\\.[0-9]+",
				"explain": "Subnet 192.168.1.0/24"
			}` + "\nGood luck!",
			wantRegex:   "192\\.168\\.1\\.[0-9]+",
			wantExplain: "Subnet 192.168.1.0/24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAIRegexResponse(tt.raw)
			if got.Regex != tt.wantRegex {
				t.Errorf("parseAIRegexResponse() Regex = %v, want %v", got.Regex, tt.wantRegex)
			}
			if tt.wantExplain != "" && got.Explain != tt.wantExplain {
				t.Errorf("parseAIRegexResponse() Explain = %v, want %v", got.Explain, tt.wantExplain)
			}
			// Must compile as Go regexp
			if _, err := regexp.Compile(got.Regex); err != nil {
				t.Errorf("Regex %s failed to compile: %v", got.Regex, err)
			}
		})
	}
}
