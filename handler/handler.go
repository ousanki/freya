package handler

import (
	"freya/common"
	"github.com/kataras/iris/context"
)

type TcpHandler struct {
	MsgId    uint16
	Handler  IHandler
	Proto    interface{}
}

type IHandler interface {
	Handler(common.TcpHeader, interface{})
}

type HttpHandler struct {
	Path       string
	Method     string
	Handlers   []context.Handler
}