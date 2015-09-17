package loges

import (
	u "github.com/araddon/gou"
	"github.com/bmizerany/assert"
	"strings"
	"testing"
	"time"
)

func init() {
	u.SetupLogging("debug")
}

func TestFileFlattenNoLevel(t *testing.T) {
	msgChan := make(chan *LineEvent, 1000)

	events := make([]*Event, 0)
	lineHandler := MakeFileFlattener("etcd", msgChan)
	TransformRegister(FileFormatter("golog", nil))

	msgCt := 0
	go func() {
		for in := range msgChan {
			//u.Debugf("GotMsg %s", string(in.Data))
			for _, transform := range transforms {
				if msg := transform(in); msg != nil {
					msgCt += 1
					events = append(events, msg)
				} else {
					//These are ok, just means its not destined for ElasticSearch
					//u.Debugf("bad es? %v", in)
				}
			}
		}
	}()

	lineHandler("2015/09/15 22:02:23 etcdserver: member dir = /home/user/.local/etcd/member")
	lineHandler("2015/09/15 22:02:23 etcdserver: heartbeat = 100ms")
	time.Sleep(time.Millisecond * 20)
	assert.Tf(t, msgCt == 1, "has message: %v", msgCt)
	assert.T(t, strings.Contains(events[0].Message, "/home/user/.local/etcd/member"))

	lineHandler(`2015/09/17 16:48:55.532259 workmgmt.go:110: [INFO] aid=12, method=GET
	 url=/api/work/abcdef?%3Aid=4444444&showcompleted=true, userid=123456 Reading work`)
	lineHandler(`2015/09/17 16:48:57.298668 ast.go:863: [WARN] If a column is by field, must be by in both:  
		<col stream="segment_users" as="fbuid" aid="1474" by?true op="latest" indextype="string" datatype="string" />
		!=<col stream="segmentio" as="fbuid" aid="1474" by?false op="latest" indextype="string" datatype="string" />`)
	time.Sleep(time.Millisecond * 20)
	assert.Tf(t, msgCt == 3, "has message: %v", msgCt)
	assert.T(t, strings.Contains(events[2].Message, "userid=123456 Reading work"))

	lineHandler(`2014/08/17 04:27:45.269905 sink_elasticsearch.go:43: [ERROR] 2014-08-17 04:27:45.269885863 +0000 UTC: Error [Failed to derive xcontent from org.elasticsearch.common.bytes.BytesArray@0] Status [<nil>] [400]
panic: Too many errors in ES

goroutine 42 [running]:
runtime.panic(0x679580, 0xc20850f7c0)
    /usr/local/go/src/pkg/runtime/panic.c:279 +0xf5
github.com/araddon/loges.func·009()
    /home/aaron/Dropbox/go/root/src/github.com/araddon/loges/sink_elasticsearch.go:35 +0xc6
created by github.com/araddon/loges.ToElasticSearch
    /home/aaron/Dropbox/go/root/src/github.com/araddon/loges/sink_elasticsearch.go:39 +0x382

goroutine 16 [chan receive, 33 minutes]:
main.main()
    /home/aaron/Dropbox/go/root/src/github.com/araddon/loges/loges/main.go:166 +0xc4c

goroutine 19 [finalizer wait, 33 minutes]:
runtime.park(0x414f20, 0x91fc10, 0x91de89)
    /usr/local/go/src/pkg/runtime/proc.c:1369 +0x89
runtime.parkunlock(0x91fc10, 0x91de89)
    /usr/local/go/src/pkg/runtime/proc.c:1385 +0x3b
runfinq()
    /usr/local/go/src/pkg/runtime/mgc0.c:2644 +0xcf
runtime.goexit()
    /usr/local/go/src/pkg/runtime/proc.c:1445`)
	lineHandler("2015/09/15 22:02:23 etcdserver: heartbeat = 100ms")
	time.Sleep(time.Millisecond * 20)
	assert.Tf(t, msgCt == 5, "has message: %v", msgCt)
	assert.T(t, strings.Contains(events[4].Message, "goroutine 19 [finalizer wait, 33 minutes]:"))
}
