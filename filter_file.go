package loges

import (
	//"bytes"
	//"encoding/json"
	"strings"
	"time"

	u "github.com/araddon/gou"
)

var expectedLevels = map[string]bool{
	"DEBU":   true,
	"DEBG":   true,
	"DEBUG":  true,
	"INFO":   true,
	"ERROR":  true,
	"ERRO":   true,
	"WARN":   true,
	"FATAL":  true,
	"FATA":   true,
	"METRIC": true,
	"METR":   true,
}

// file format [date source jsonmessage] parser
func FileFormatter(logstashType string, tags []string) LineTransform {
	loc := time.UTC
	pos := 0
	posEnd := 0
	logLevel := ""

	return func(d *LineEvent) *Event {
		// 2013/05/26 13:07:47.606937 rw.go:70: [INFO] RW service is up
		// 2013/05/26 13:07:47.607 [DEBG] sink       Building sink for kafka from factory method
		line := string(d.Data)

		// Find first square brackets
		pos = strings.IndexRune(line, '[')
		posEnd = strings.IndexRune(line, ']')
		if pos > 0 && posEnd > 0 && posEnd > pos && len(line) > posEnd {
			logLevel = line[pos+1 : posEnd]
		} else {
			logLevel = "NONE"
		}
		// Don't log out These metrics
		if logLevel == "METRIC" || logLevel == "METR" {
			return nil
		}
		// if _, ok := expectedLevels[logLevel]; ok {
		// 	return nil
		// }
		//u.Debugf("dt='%s' line: %s", d.DataType, line)
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
					if _, err := time.Parse("2006/01/02 15:04:05.000000", datePart); err == nil {
						evt := NewTsEvent(logstashType, d.Source, parts[2], time.Now().In(loc))
						evt.Fields = make(map[string]interface{})
						evt.Fields["host"] = hostName
						evt.Fields["level"] = logLevel
						evt.Fields["WriteErrs"] = d.WriteErrs
						//evt.Fields = msg
						//evt.Source = d.Source
						//u.Debug(evt.String())
						return evt
					}
				} else {
					if _, err := time.Parse("2006/01/02 15:04:05", datePart); err == nil {
						evt := NewTsEvent(logstashType, d.Source, parts[2], time.Now().In(loc))
						evt.Fields = make(map[string]interface{})
						evt.Fields["host"] = hostName
						evt.Fields["level"] = logLevel
						evt.Fields["WriteErrs"] = d.WriteErrs
						//evt.Fields = msg
						//evt.Source = d.Source
						//u.Debug(evt.String())
						return evt
					}
				}
				evt := NewTsEvent(logstashType, d.Source, line, time.Now())
				evt.Fields = make(map[string]interface{})
				evt.Fields["host"] = hostName
				evt.Fields["level"] = logLevel
				evt.Fields["WriteErrs"] = d.WriteErrs
				//evt.Fields = msg
				//evt.Source = d.Source
				//u.Debug(evt.String())
				return evt
			} else {
				//u.Warnf("bad? %s:%s ", logLevel, line)
			}
		}

		return nil
	}
}
