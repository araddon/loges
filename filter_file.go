package loges

import (
	"encoding/json"

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
	return func(d *LineEvent) *Event {

		//u.Infof("%v line event: %v  Metric?%v  json?%v", d.Ts, d.LogLevel, d.IsMetric(), d.IsJson())

		// Don't log out metrics
		if d.IsMetric() {
			return nil
		}
		if len(d.Data) < 10 {
			u.Warn("Invalid line?", string(d.Data))
			return nil
		} else if !d.Ts.IsZero() {
			evt := NewTsEvent(logstashType, d.Source, string(d.Data), d.Ts)
			evt.Fields = make(map[string]interface{})
			evt.Fields["host"] = hostName
			evt.Fields["level"] = d.LogLevel
			evt.Fields["WriteErrs"] = d.WriteErrs

			if d.IsJson() {
				jd := json.RawMessage(d.Data)
				evt.Raw = &jd
			}
			return evt

		}
		return nil
	}
}
