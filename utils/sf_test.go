package utils

import (
	"fmt"
	"testing"
)

func TestWoker(t *testing.T) {
	wk, e := newWorker(100)
	if e != nil {
		fmt.Println("newWorker error:", e.Error())
		return
	}
	tid := wk.nextID()
	fmt.Println("NextID is", tid)
	wid := WorkerID(tid)
	fmt.Println("workerID is", wid)
}
