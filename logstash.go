package loges

import (
	"encoding/json"
	u "github.com/araddon/gou"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
)

var (
	esIndex    string
	transforms        = make([]LineTransform, 0)
	hostName   string = "Unknown"
)

func init() {
	if host, err := os.Hostname(); err == nil {
		hostName = host
	}
}

// Representing data about a line from FluentD
type LineEvent struct {
	Data     []byte
	DataType string
	Source   string
	Offset   uint64
	Item     interface{}
}
type LineTransform func(*LineEvent) *Event

func TransformRegister(txform LineTransform) {
	u.Debug("setting foramtter")
	transforms = append(transforms, txform)
}

// update the index occasionally
func UpdateLogstashIndex() {
	esIndex = "logstash-" + time.Now().Format("2006.01.02")
	c := time.Tick(1 * time.Minute)
	for now := range c {
		esIndex = "logstash-" + now.Format("2006.01.02")
	}
}

// A Logstash formatted event
type Event struct {
	id        string
	Source    string                 `json:"@source"`
	Type      string                 `json:"@type"`
	Timestamp time.Time              `json:"@timestamp"`
	Message   string                 `json:"@message"`
	Tags      []string               `json:"@tags"`
	Fields    map[string]interface{} `json:"@fields"`
}

func NewEvent(eventType, source, message string) *Event {
	return &Event{Type: eventType,
		Timestamp: time.Now(),
		Message:   message,
		Source:    source,
	}
}

// New event using existing time stamp
func NewTsEvent(eventType, source, message string, t time.Time) *Event {
	return &Event{Type: eventType,
		Timestamp: t,
		Message:   message,
		Source:    source,
	}
}

// Set the id instead of using mongo objectid as Id
func (e *Event) SetId(id string) {
	e.id = id
}
func (e *Event) Id() string {
	if len(e.id) < 1 {
		e.id = bson.NewObjectId().Hex()
	}
	return e.id
}

func (e *Event) Index() string {
	return "logstash-" + e.Timestamp.Format("2006.01.02")
}

func (e *Event) String() string {
	b, _ := json.Marshal(e)
	return string(b)
}
