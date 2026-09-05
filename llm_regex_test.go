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
		{
			name: "Phrase with space and case insensitivity",
			raw: `{
				"regex": "(?i)connection\\s+refused",
				"explain": "connection refusedを含むログを検索"
			}`,
			wantRegex:   "(?i)connection\\s+refused",
			wantExplain: "connection refusedを含むログを検索",
		},
		{
			name: "Simple keyword contains",
			raw: `{
				"regex": "(?i)timeout",
				"explain": "timeoutを含むログを検索"
			}`,
			wantRegex:   "(?i)timeout",
			wantExplain: "timeoutを含むログを検索",
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

func TestCleanSearchPrompt(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Invalid userを含む", "Invalid user"},
		{"timeoutを含むログ", "timeout"},
		{"エラーのログ", "エラー"},
		{"パスワード認証失敗を検索", "パスワード認証失敗"},
		{"contains Invalid user", "Invalid user"},
		{"logs with error", "error"},
		{"192.168.1.1からのアクセス", "192.168.1.1"},
		{"SSHのログイン失敗", "SSHのログイン失敗"},
		{"", ""},
	}

	for _, tt := range tests {
		got := cleanSearchPrompt(tt.input)
		if got != tt.want {
			t.Errorf("cleanSearchPrompt(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCleanGeneratedRegex(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"(?i)invalid userを含む", "(?i)invalid user"},
		{"(?i)invalid\\s+userを含むログ", "(?i)invalid\\s+user"},
		{"(?i)timeout", "(?i)timeout"},
		{"192\\.168\\.1\\.1", "192\\.168\\.1\\.1"},
	}

	for _, tt := range tests {
		got := cleanGeneratedRegex(tt.input)
		if got != tt.want {
			t.Errorf("cleanGeneratedRegex(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
