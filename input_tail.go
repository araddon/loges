package loges

import (
	"bytes"
	"fmt"
	u "github.com/araddon/gou"
	tail "github.com/fw42/go-tail"
	"io/ioutil"
	"strings"
)

var (
	_ = u.DEBUG
)

func TailFile(filename string, config tail.Config, done chan bool, msgChan chan *LineEvent) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	lineHandler := MakeFileParser(filename, msgChan)
	for line := range t.Lines {
		lineHandler(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

// This formatter reads go files and performs:
//  1.  Squashes multiple lines into one
//  2.  Tries to squash panics into one line
//  3.  Reads out the [DEBUG,INFO] into a field
//
// This expects log files in this format
//   2013-05-25 13:25:32.475 authctx.go:169: [DEBUG] sink       Building sink for kafka from factory method
func MakeFileParser(filename string, msgChan chan *LineEvent) func(string) {
	// Builder used to build the colored string.
	buf := new(bytes.Buffer)

	startsNumeric := false
	//hasBrackets := false
	pos := 0

	return func(line string) {

		if len(line) < 1 {
			return
		}

		startsNumeric = false
		//hasBrackets = false

		firstRune := line[0]
		if firstRune >= '0' && firstRune <= '9' {
			startsNumeric = true
		}

		// Find first square bracket
		pos = strings.IndexRune(line, '[')

		if pos == -1 && !startsNumeric {
			// accumulate in buffer, probably/possibly a panic?
			buf.WriteString(line)
			//u.Warnf("||| 1 not bracket or numeric %s", line)
		} else if !startsNumeric {
			// accumulate in buffer
			buf.WriteString(line)
			//u.Warnf("||| 2 not numeric %s", line)
		} else {
			// Line had [] AND had numeric at start
			if buf.Len() > 0 {
				// we already have previous stuff in buffer
				data, err := ioutil.ReadAll(buf)
				//u.Warnf("||| %s\n", string(data))
				if err == nil {
					msgChan <- &LineEvent{Data: data, Source: filename}
				}
			}
			//u.Warnf("|| 3 %s", line)
			//msgChan <- &LineEvent{Data: []byte(line)}
			buf.WriteString(line)
		}
	}
}
