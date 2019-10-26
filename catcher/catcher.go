package catcher

import "runtime/debug"

func CatchError() {
	if err := recover(); err != nil {
		debug.PrintStack()
	}
}

