package loges

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	u "github.com/araddon/gou"
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

// Representing data about a log line from file we are tailing
type LineEvent struct {
	Data      []byte
	Prefix    string
	Ts        time.Time // Date string if found
	LogLevel  string    // [METRIC, INFO,  DEBUG]
	Source    string    // Source = filename if file, else monit, etc
	Offset    uint64
	Item      interface{}
	WriteErrs uint16
}

func (l *LineEvent) IsJson() bool {
	return IsJsonObject(l.Data)
}
func (l *LineEvent) IsMetric() bool {
	switch l.LogLevel {
	case "METRIC", "METR":
		return true
	}
	return false
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
	id          string
	Source      string                 `json:"@source"`
	Type        string                 `json:"@type"`
	Timestamp   time.Time              `json:"@timestamp"`
	Message     string                 `json:"@message"`
	Tags        []string               `json:"@tags,omitempty"`
	IndexFields map[string]interface{} `json:"@idx,omitempty"`
	Fields      map[string]interface{} `json:"@fields,omitempty"`
	Raw         *json.RawMessage       `json:"jsonfields,omitempty"`
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
		// We want consistent hashing here, so we auto-dedupe rows
		md5h := md5.New()
		md5h.Write([]byte(fmt.Sprintf("%d:%s", e.Timestamp.UnixNano(), e.Message)))
		e.id = base64.StdEncoding.EncodeToString(md5h.Sum(nil))
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

// Returns the first non-whitespace rune in the byte slice. Returns false if the input is
// not valid UTF-8 or the input contains no runes except whitespace.
// Returns
func FirstNonWhitespace(bs []byte) (r rune, ok bool) {
	for {
		if len(bs) == 0 {
			return 0, false
		}
		r, numBytes := utf8.DecodeRune(bs)
		switch r {
		case '\t', '\n', '\r', ' ':
			bs = bs[numBytes:] // advance past the current whitespace rune and continue
			continue
		case utf8.RuneError: // This is returned when invalid UTF8 is found
			return 0, false
		}
		return r, true
	}
	return 0, false
}

// Determines if the bytes is a json array, only looks at prefix
//  not parsing the entire thing
func IsJsonArray(by []byte) bool {
	firstRune, ok := FirstNonWhitespace(by)
	if !ok {
		return false
	}
	if firstRune == '[' {
		return true
	}
	return false
}

func IsJsonObject(by []byte) bool {
	firstRune, ok := FirstNonWhitespace(by)
	if !ok {
		return false
	}
	if firstRune == '{' {
		return true
	}
	return false
}

func IsJson(by []byte) bool {
	if IsJsonObject(by) {
		return true
	}
	return IsJsonArray(by)
}
