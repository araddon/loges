package loges

import (
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/core"
	log "github.com/ngmoco/timber"
)

// read the message channel and send to elastic search
func toEs(msgChan chan *LineEvent, esType, esHost string) {
	for in := range msgChan {
		msg := formatter(in)
		if msg != nil {
			// TODO:  improve this, if we open this up in go routine then we get these errors
			//[2012-11-22 22:54:19.734] EROR dial tcp 192.168.1.26:9200: too many open files
			//[2012-11-22 22:54:19.734] EROR lookup hostname: no such host
			if _, err := core.Index(true, msg.Index(), esType, msg.Id(), msg); err != nil {
				log.Error("%v", err)
			}
		}
	}
}

// read the message channel and send to elastic search
func ToElasticSearch(msgChan chan *LineEvent, esType, esHost string) {
	// set elasticsearch host
	api.Domain = esHost
	// open 10 writers?   
	// TODO:  es lib should limit?  
	for i := 0; i < 10; i++ {
		go toEs(msgChan, esType, esHost)
	}
}
