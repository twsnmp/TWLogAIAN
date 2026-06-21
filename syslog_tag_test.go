package main

import (
	"testing"
)

func TestSyslogTagParsing(t *testing.T) {
	// Initialize default extractor types
	makeDefalutLogTypes()

	app := NewApp()
	app.config.Extractor = "syslog"

	err := app.setExtractor()
	if err != nil {
		t.Fatalf("failed to set extractor: %v", err)
	}

	if app.processConf.View != "syslog" {
		t.Errorf("expected View to be 'syslog', got '%s'", app.processConf.View)
	}

	// Test case 1: with program and pid
	l1 := LogEnt{
		KeyValue: make(map[string]interface{}),
		All:      "Jun  9 10:11:12 host program[1234]: message body",
	}
	app.parseLogEnt(&l1)

	tag1, ok := l1.KeyValue["tag"].(string)
	if !ok {
		t.Errorf("tag is not found or not a string: %+v", l1.KeyValue)
	} else if tag1 != "program[1234]" {
		t.Errorf("expected tag to be 'program[1234]', got '%s'", tag1)
	}

	// Test case 2: with program and no pid
	l2 := LogEnt{
		KeyValue: make(map[string]interface{}),
		All:      "Jun  9 10:11:12 host program: message body",
	}
	app.parseLogEnt(&l2)

	tag2, ok := l2.KeyValue["tag"].(string)
	if !ok {
		t.Errorf("tag is not found or not a string: %+v", l2.KeyValue)
	} else if tag2 != "program" {
		t.Errorf("expected tag to be 'program', got '%s'", tag2)
	}
}
