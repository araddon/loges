package loges

import (
	u "github.com/araddon/gou"
	"os"
	"strings"
)

var (
	_ = os.DevNull
)

// read the message channel and send to elastic search
func ToStdout(msgChan chan *LineEvent, colorize bool) {
	pos := 0

	logLevel := u.LogColor[u.DEBUG]

	// Find next square bracket, break loop when none was found.
	for in := range msgChan {
		line := string(in.Data)

		pos = strings.IndexRune(line, '[')
		if pos == -1 {
			logLevel = u.LogColor[u.DEBUG]
		} else {
			if len(line) < pos+1 {
				continue
			}
			switch line[pos+1 : pos+5] {
			case "DEBU", "DEBG":
				logLevel = u.LogColor[u.DEBUG]
			case "INFO":
				logLevel = u.LogColor[u.INFO]
			case "WARN":
				logLevel = u.LogColor[u.WARN]
			case "ERRO":
				logLevel = u.LogColor[u.ERROR]
			default:
				//logLevel := u.LogColor[u.ERROR]
				//println("level not recognized? " + line[pos+1:pos+5])
			}
		}
		//u.Debugf("%sabout to print ln:  '%s' len=%d", "\033[0m", line[pos+1:pos+5], len(line))
		os.Stdout.WriteString(logLevel + line + "\033[0m")
	}
}
