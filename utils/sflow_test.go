package utils

import (
	"fmt"
	"testing"
)

func TestWokerLow(t *testing.T) {
	wk, e := newWorkerLow(100)
	if e != nil {
		fmt.Println("newWorker error:", e.Error())
		return
	}
	for i := 0; i < 1; i++ {
		tid := wk.nextIDLow()
		fmt.Println("NextID is", tid)
		wid := WorkerIDLow(tid)
		fmt.Println("workerID is", wid)
	}
}
