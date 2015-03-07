package loges

import (
	"bytes"
	"github.com/araddon/dateparse"
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

	startsDate := false
	pos := 0
	posEnd := 0
	var dataType []byte
	var loglevel string
	var dateStr string
	lineCt := 0

	return func(line string) {
		lineCt++
		if len(line) < 8 {
			buf.WriteString(line)
			return
		}

		startsDate = false
		spaceCt := 0

		// 2014/07/10 11:04:20.653185 filter_fluentd.go:16: [DEBUG] %s %s
		for i := 0; i < len(line); i++ {
			r := line[i]
			if r == ' ' {
				if spaceCt == 1 {
					dateStr = string(line[:i])
					if _, err := dateparse.ParseAny(dateStr); err == nil {
						startsDate = true
					}
					break
				}
				spaceCt++
			}
		}

		// Find first square bracket wrapper:   [WARN]
		pos = strings.IndexRune(line, '[')
		posEnd = strings.IndexRune(line, ']')
		if pos > 0 && posEnd > 0 && pos < posEnd && len(line) > pos && len(line) > posEnd {
			loglevel = line[pos+1 : posEnd]
			if _, ok := expectedLevels[loglevel]; !ok {
				buf.WriteString(line)
				return
			}
		}

		//u.Debugf("pos=%d datatype=%s num?=%v", pos, dataType, startsDate)
		//u.Infof("starts with date?=%v pos=%d lvl=%s short[]%v len=%d buf.len=%d", startsDate, pos, loglevel, (posEnd-pos) < 8, len(line), buf.Len())
		if pos == -1 {
			// accumulate in buffer, probably/possibly a panic?
			buf.WriteString(line)
			buf.WriteString(" \n")
			return
		} else if !startsDate {
			// accumulate in buffer
			buf.WriteString(line)
			buf.WriteString(" \n")
			return
		} else if posEnd-8 > pos {
			// position of [block]  too long, so ignore
			buf.WriteString(line)
			buf.WriteString(" \n")
			return
		} else if pos > 80 {
			// [WARN] should be at beginning of line
			buf.WriteString(line)
			buf.WriteString(" \n")
			return
		} else {
			// Line had [STUFF] AND startsDate at start

			if buf.Len() == 0 {
				// lets buffer it, ensuring we have the completion of this line
				buf.WriteString(line)
				return
			}

			// we already have previous line in buffer
			data, err := ioutil.ReadAll(buf)
			if err == nil {
				pos = bytes.IndexRune(data, '[')
				posEnd = bytes.IndexRune(data, ']')
				if posEnd-8 > pos {
					//u.Warnf("level:%s  \n\nline=%s", string(data[pos+1:posEnd]), string(data))
					//buf.WriteString(line)
					return
				} else if pos > 0 && posEnd > 0 && pos < posEnd && len(data) > pos && len(data) > posEnd {
					dataType = data[pos+1 : posEnd]
				} else {
					dataType = []byte("NA")
					//u.Warnf("level:%s  \n\nline=%s", string(data[pos+1:posEnd]), string(data))
				}
				// if !bytes.HasPrefix(data, datePrefix) {
				// 	u.Warnf("ct=%d level:%s  \n\nline=%s", lineCt, string(data[pos+1:posEnd]), string(data))
				// }
				//u.Debugf("dt='%s'  data=%s", string(dataType), string(data[0:20]))
				msgChan <- &LineEvent{Data: data, DataType: string(dataType), Source: filename}
			} else {
				u.Error(err)
			}

			// now write this line for next analysis
			buf.WriteString(line)
		}
	}
}
