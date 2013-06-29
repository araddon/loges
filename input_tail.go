package loges

import (
	"github.com/ActiveState/tail"
	u "github.com/araddon/gou"
)

var (
	_ = u.DEBUG
)

func TailFile(filename string, config tail.Config, done chan bool, msgChan chan *LineEvent) {
	u.Debug("Watching file ", filename, config)
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
	if err := t.Stop(); err != nil {
		u.Info(err)
	}
}
