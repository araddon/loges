package loges

import (
	"bytes"
	"encoding/json"
	u "github.com/araddon/gou"
	"time"
)

// Fluentd format [date source jsonmessage] parser
func FluentdFormatter(logstashType string, tags []string) LineFormatter {
	return func(d *LineEvent) *Event {
		//2012-11-22 05:07:51 +0000 lio.home.ubuntu.log.collect.log.vm2: {"message":"runtime error: close of closed channel"}
		if lineParts := bytes.SplitN(d.Data, []byte{':', ' '}, 2); len(lineParts) > 1 {
			if len(lineParts[0]) > 26 {
				u.Debug("%s %s\n", string(lineParts[0]), string(lineParts[1]))
				bsrc := lineParts[0][26:]
				bdate := lineParts[0][0:25]
				var msg map[string]interface{}
				if err := json.Unmarshal(lineParts[1], &msg); err == nil {
					if t, err := time.Parse("2006-01-02 15:04:05 -0700", string(bdate)); err == nil {
						evt := NewTsEvent(logstashType, string(bsrc), "", t)
						if msgi, ok := msg["message"]; ok {
							if msgS, ok := msgi.(string); ok {
								evt.Message = msgS
								delete(msg, "message")
							}
						}
						evt.Tags = tags
						evt.Fields = msg
						return evt
					} else {
						u.Debug("%v", err)
						return NewEvent(logstashType, string(bsrc), string(lineParts[1]))
					}

				} else {
					u.Warn("bad message? %v", err)
				}

			}
		}
		return nil
	}
}
