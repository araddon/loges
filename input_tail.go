package loges

import (
	u "github.com/araddon/gou"
	tail "github.com/fw42/go-tail"
)

var (
	_ = u.DEBUG
)

func TailFile(filename string, config tail.Config, done chan bool, msgChan chan *LineEvent) {
	u.Debug("Watching file ", filename)
	t, err := tail.TailFile(filename, config)
	if err != nil {
		u.Error(err)
		return
	}
	//defer func() { done <- true }()

	lineHandler := MakeFileFlattener(filename, msgChan)
	for line := range t.Lines {
		lineHandler(line.Text)
	}
	err = t.Wait()
	if err != nil {
		u.Error(err)
	}
}
