package loges

import (
	//"bytes"
	//"encoding/json"
	u "github.com/araddon/gou"
)

func GraphiteTransform() LineTransform {
	return func(d *LineEvent) *Event {
		if d.DataType == "METRIC" || d.DataType == "METR" {
			u.Info("Should be sending to Graphite! ")
		}
		return nil
	}
}
