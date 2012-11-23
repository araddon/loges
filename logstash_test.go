package loges

import (
	"testing"
)

func init() {
	//TimberSetLogging("[%D %T] %s %L %M", "DEBUG")
}

func TestFluentdFormat(t *testing.T) {
	formatter := FluentdFormatter("testing", []string{"tag1", "tag2"})
	src := `2012-11-22 05:07:51 +0000 lio.home.ubuntu.log.collect.log.vm2: {"message":"runtime error: close of closed channel","part":"val"}`
	e := formatter(&LineEvent{Data: []byte(src)})
	if len(e.Tags) != 2 {
		t.Errorf("Should have found tags %v", e.Tags)
	}
	if e.Timestamp.Unix() != 1353560871 { //  correct time stamp
		t.Errorf("Should have found tags %v", e.Tags)
	}
	if e.Message != "runtime error: close of closed channel" {
		t.Errorf("Should have found tags %v", e.Tags)
	}
	if e.Type != "testing" {
		t.Errorf("Should be type=testing but was %v", e.Type)
	}
	if len(e.Fields) != 1 {
		t.Errorf("Should be one field but was %v", e.Fields)
	}
	if part, ok := e.Fields["part"]; !ok || part.(string) != "val" {
		t.Errorf("Fields have part? but was %v", e.Fields)
	}
	if e.Source != "lio.home.ubuntu.log.collect.log.vm2" {
		t.Errorf("Should be src but was %v", e.Source)
	}
}
