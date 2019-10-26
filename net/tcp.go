package net

import (
	"errors"
	"fmt"
	"github.com/ousanki/freya/decoder"
	"github.com/ousanki/freya/encoder"
	"github.com/ousanki/freya/global"
	"github.com/ousanki/freya/handler"
	"github.com/ousanki/freya/log"
	"io"
	"net"
	"sync"
	"time"
)

type UGwMap struct {
	sync.RWMutex
	UGw  map[uint32]net.Conn
}

func (ugw *UGwMap) save(clientId uint32, conn net.Conn) {
	ugw.Lock()
	defer ugw.Unlock()

	ugw.UGw[clientId] = conn
}

func (ugw *UGwMap) load(clientId uint32) net.Conn {
	ugw.RLock()
	ugw.RUnlock()

	return ugw.UGw[clientId]
}

type TcpServer struct {
	handlers    map[uint16]handler.IHandler
	Ugw         *UGwMap
}

var tcpServer *TcpServer

func InitTcpServer(handlers ...handler.TcpHandler) {
	tcpServer = newTcpServer()
	tcpServer.regTcpHandlers(handlers...)
}

func StartTcpServer() {
	go tcpServer.listenTcp()
}

func newTcpServer() *TcpServer {
	return &TcpServer{
		handlers: make(map[uint16]handler.IHandler),
		Ugw:      &UGwMap{
			UGw: make(map[uint32]net.Conn),
		},
	}
}

func (s *TcpServer) regTcpHandlers(handlers ...handler.TcpHandler) {
	for _, h := range handlers {
		s.handlers[h.MsgId] = h.Handler
	}
}

func (s *TcpServer) listenTcp() {
	if global.G.TcpPort == 0 {
		return
	}

	port := fmt.Sprintf(":%d", global.G.TcpPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	if err != nil {
		panic(err.Error())
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err.Error())
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.GetLogger().Errorf("listenTcp | AcceptTCP | err:%s", err.Error())
			continue
		}
		// handler conn
		go s.workConn(conn)
	}
}

func (s *TcpServer) workConn(conn net.Conn) {
	for {
		h, err := decoder.De.DecodeHeader(conn)
		if err != nil && err == io.EOF {
			log.GetLogger().Errorf("workConn | DecodeHeader | EOF ip:%s", conn.RemoteAddr())
			break
		}
		if err != nil {
			log.GetLogger().Errorf("workConn | DecodeHeader | err:%s", err.Error())
			continue
		}
		if h.Length == 0 {
			log.GetLogger().Errorf("workConn | DecodeHeader | h.Length is 0")
			continue
		}
		s.Ugw.save(h.ClientID, conn)

		body := make([]byte, h.Length)
		_, err = io.ReadFull(conn, body)
		if err != nil && err == io.EOF {
			log.GetLogger().Errorf("workConn | ReadFull | EOF ip:%s", conn.RemoteAddr())
			break
		}
		if err != nil {
			log.GetLogger().Errorf("workConn | ReadFull | err:%s", err.Error())
			continue
		}
		// 解析体
		inf, err := decoder.De.DecodeBody(h.MsgId, body)
		if err != nil {
			log.GetLogger().Errorf("workConn | DecodeBody | body:%+v, err:%s", body, err.Error())
			continue
		}
		if worker, has := s.handlers[h.MsgId]; has {
			worker.Handler(h, inf)
		} else {
			log.GetLogger().Warningf("workConn | DecodeBody | Msg:%d not find handler", h.MsgId)
		}
	}
}

func Send(msgId uint16, proxyId uint64, clientId uint32, delay int, data interface{}) error {
	msg, err := encoder.Encode(msgId, proxyId, clientId, data)
	if err != nil {
		return err
	}
	conn := tcpServer.Ugw.load(clientId)
	if conn == nil {
		return errors.New(fmt.Sprintf("freya | tcp | Send not find conn, clientId:%d", clientId))
	}
	fn := func() {
		var offset int
		for {
			n, err := conn.Write(msg[offset:])
			if err != nil {
				return
			}
			if offset + n >= len(msg) {
				break
			}
			offset += n
		}
	}

	if delay > 0 {
		go func() {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			fn()
		}()
	} else {
		fn()
	}

	return nil
}
