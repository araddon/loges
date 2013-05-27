package loges

import (
	u "github.com/araddon/gou"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
)

// read the message channel and send to elastic search
// Uses the background Bulk Indexor, which has the **Possibility** of losing
// data if app panic/quits (but is much faster than non-bulk)
func ToElasticSearch(msgChan chan *LineEvent, esType, esHost string) {
	// set elasticsearch host which is a global
	u.Warnf("Starting elasticsearch on %s", esHost)
	api.Domain = esHost
	done := make(chan bool)
	core.BulkIndexorGlobalRun(100, done)

	for in := range msgChan {
		msg := formatter(in)
		if msg != nil {
			//u.Info(msg.String())
			//IndexBulk(index string, _type string, id string, date *time.Time, data interface{})
			if err := core.IndexBulk(msg.Index(), esType, msg.Id(), nil, msg); err != nil {
				u.Error("%v", err)
			}
		}
	}
}
