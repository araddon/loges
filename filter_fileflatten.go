package loges

import (
	"bytes"
	u "github.com/araddon/gou"
	"io/ioutil"
	"strings"
)

// This formatter reads go files and performs:
//  1.  Squashes multiple lines into one (as needed), Tries to squash panics(go) into one line
//  2.  Reads out the LineType/Level [DEBUG,INFO,METRIC] into a field
//
// This expects log files in this format
//   2013-05-25 13:25:32.475 authctx.go:169: [DEBUG] sink       Building sink for kafka from factory method
func MakeFileFlattener(filename string, msgChan chan *LineEvent) func(string) {
	// Builder used to build the colored string.
	buf := new(bytes.Buffer)

	startsNumeric := false
	pos := 0
	posEnd := 0
	var dataType []byte

	return func(line string) {

		if len(line) < 1 {
			return
		}

		startsNumeric = false

		firstRune := line[0]
		if firstRune >= '0' && firstRune <= '9' {
			startsNumeric = true
		}
		// Find first square bracket
		pos = strings.IndexRune(line, '[')

		//u.Debugf("pos=%d datatype=%s num?=%v", pos, dataType, startsNumeric)
		//u.Infof("numeric?=%v pos=%d len=%d", startsNumeric, pos, len(line))
		if pos == -1 && !startsNumeric {
			// accumulate in buffer, probably/possibly a panic?
			buf.WriteString(line)
			buf.WriteByte('\n')
		} else if !startsNumeric {
			// accumulate in buffer
			buf.WriteString(line)
			buf.WriteByte('\n')
		} else {
			// Line had [STUFF] AND had numeric at start
			if buf.Len() > 0 {
				// we already have previous stuff in buffer
				data, err := ioutil.ReadAll(buf)
				if err == nil {
					pos = bytes.IndexRune(data, '[')
					posEnd = bytes.IndexRune(data, ']')
					if pos > 0 && posEnd > 0 {
						dataType = data[pos+1 : posEnd]
					} else {
						dataType = []byte("NA")
					}
					//u.Debugf("dt=%s  data=%s", string(dataType), string(data))
					msgChan <- &LineEvent{Data: data, DataType: string(dataType), Source: filename}
				} else {
					u.Error(err)
				}
			}
			buf.WriteString(line)
		}
	}
}
