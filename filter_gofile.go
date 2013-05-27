package loges

import (
	//"bytes"
	//"encoding/json"
	u "github.com/araddon/gou"
	"strings"
	"time"
)

// go file format [date source jsonmessage] parser
func GoFileFormatter(logstashType string, tags []string) LineFormatter {
	return func(d *LineEvent) *Event {
		// 2013/05/26 13:07:47.606937 rw.go:70: [INFO] RW service is up
		// 2013/05/26 13:07:47.607 [DEBG] sink       Building sink for kafka from factory method
		line := string(d.Data)
		//u.Warn(line)
		if len(line) < 10 {
			u.Warn(line)
			return nil
		} else {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) > 2 {
				datePart := parts[0] + " " + parts[1]
				// "2006/01/02 15:04:05.000000"
				if len(datePart) > 24 {
					if t, err := time.Parse("2006/01/02 15:04:05.000000", datePart); err == nil {
						evt := NewTsEvent(logstashType, d.Source, parts[2], t)
						evt.Fields = make(map[string]interface{})
						evt.Fields["host"] = hostName
						//evt.Fields = msg
						//evt.Source = d.Source
						u.Info(evt.String())
						return evt
					}
				} else {
					if t, err := time.Parse("2006/01/02 15:04:05", datePart); err == nil {
						evt := NewTsEvent(logstashType, d.Source, parts[2], t)
						evt.Fields = make(map[string]interface{})
						evt.Fields["host"] = hostName
						//evt.Fields = msg
						//evt.Source = d.Source
						u.Info(evt.String())
						return evt
					}
				}
				u.Warnf("cant parse %v", err)
				evt := NewTsEvent(logstashType, d.Source, line, time.Now())
				evt.Fields = make(map[string]interface{})
				evt.Fields["host"] = hostName
				//evt.Fields = msg
				//evt.Source = d.Source
				u.Info(evt.String())
				return evt
			} else {
				u.Warn("bad? ", line)
			}
		}

		return nil
	}
}
