package loges

import (
	u "github.com/araddon/gou"
)

var (
	_ = u.DEBUG
)

//just run the transforms
func RunTransforms(parallel int, msgChan chan *LineEvent) {

	for i := 0; i < parallel; i++ {
		go func(mc chan *LineEvent, txforms []LineTransform) {
			// TODO, refactor this and elasticsearch sink one into a "Router"
			for in := range mc {
				for _, transform := range txforms {
					transform(in)
				}
			}
		}(msgChan, transforms)
	}
}
