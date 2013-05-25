package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	u "github.com/araddon/gou"
	"io"
	"os"
	"strings"
)

func args2config() tail.Config {
	config := tail.Config{Follow: true}
	flag.IntVar(&config.Location, "n", 0, "tail from the last Nth location")
	flag.BoolVar(&config.Follow, "f", false, "wait for additional data to be appended to the file")
	flag.BoolVar(&config.ReOpen, "F", false, "follow, and track file rename/rotation")
	flag.Parse()
	if config.ReOpen {
		config.Follow = true
	}
	return config
}

func main() {
	config := args2config()
	if flag.NFlag() < 1 {
		fmt.Println("need one or more files as arguments")
		os.Exit(1)
	}

	done := make(chan bool)
	for _, filename := range flag.Args() {
		go tailFile(filename, config, done)
	}

	for _, _ = range flag.Args() {
		<-done
	}
}

func MakeStderrColorized() func(string) {
	// Builder used to build the colored string.
	buf := new(bytes.Buffer)

	startsNumeric := false
	//hasBrackets := false
	pos := 0

	logLevel := u.LogColor[u.DEBUG]

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

		// Find next square bracket, break loop when none was found.
		pos = strings.IndexRune(line, '[')
		if pos == -1 && !startsNumeric {
			// accumulate in buffer, probably/possibly a panic?
			logLevel = u.LogColor[u.ERROR]
			buf.WriteString(line)
		} else if !startsNumeric {
			// accumulate in buffer
			buf.WriteString(line)
		} else {
			if buf.Len() > 0 {
				// we already have previous stuff in buffer
				os.Stderr.WriteString(logLevel)
				io.Copy(os.Stderr, buf)
				os.Stderr.WriteString("\033[0m\n")
			}
			buf.WriteString(line)
			if len(line) < pos+6 {
				return
			}
			switch line[pos+1 : pos+5] {
			case "DEBU":
				logLevel = u.LogColor[u.DEBUG]
			case "INFO":
				logLevel = u.LogColor[u.INFO]
			case "WARN":
				logLevel = u.LogColor[u.WARN]
			case "ERRO":
				logLevel = u.LogColor[u.ERROR]
			default:
				//logLevel := u.LogColor[u.ERROR]
			}
			os.Stderr.WriteString(logLevel)
			io.Copy(os.Stderr, buf)
			os.Stderr.WriteString("\033[0m\n")
		}
	}
}

func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	lineHandler := MakeStderrColorized()
	for line := range t.Lines {
		lineHandler(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
