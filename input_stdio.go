package loges

import (
	"bufio"
	u "github.com/araddon/gou"
	"io"
	"os"
)

// sends messages from stdin for consumption
func StdinPruducer(msgChan chan *LineEvent) {
	b := bufio.NewReader(os.Stdin)
	lineHandler := MakeFileFlattener("stdin", msgChan)
	u.Debug("reading from stdin with lines defined by newline")
	for {
		if s, e := b.ReadString('\n'); e == nil {
			//u.Info(s)
			lineHandler(s)
		} else if e == io.EOF {
			return
		}
	}
}
