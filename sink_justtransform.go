package loges

import (
	u "github.com/araddon/gou"
)

var (
	_ = u.DEBUG
)

//just run the transforms
func RunTransforms(parallel int, msgChan chan *LineEvent, colorize bool) {

	for i := 0; i < parallel; i++ {
		go func(txforms []LineTransform) {
			// TODO, refactor this and elasticsearch sink one into a "Router"
			for in := range msgChan {
				for _, transform := range txforms {
					transform(in)
				}
			}
		}(transforms)
	}
}
