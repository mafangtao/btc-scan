package utils

import (
	"sync"
	"github.com/liyue201/go-logger"
	"runtime/debug"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("WaitGroupWrapper Wrap %s", err)
				debug.PrintStack()
			}
			w.Done()
		}()
		cb()
	}()
}
