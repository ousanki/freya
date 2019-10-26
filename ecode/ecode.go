package ecode

import (
	"fmt"
	"strconv"
)

var (
	codes = map[int]struct{}{}
)

type Code int

func (e Code) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

func New(e int) Code {
	return add(e)
}

func add(e int) Code {
	if _, ok := codes[e]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", e))
	}
	codes[e] = struct{}{}
	return Int(e)
}

func Int(i int) Code { return Code(i) }
