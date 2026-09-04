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
