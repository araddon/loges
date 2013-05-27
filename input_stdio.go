package loges

import (
	"bufio"
	u "github.com/araddon/gou"
	"os"
)

// sends messages from stdin for consumption
func StdinPruducer(msgChan chan *LineEvent) {
	b := bufio.NewReader(os.Stdin)
	u.Debug("reading from stdin with lines defined by newline")
	for {
		if s, e := b.ReadString('\n'); e == nil {
			msgChan <- &LineEvent{Data: []byte(s)}
		}
	}
}
