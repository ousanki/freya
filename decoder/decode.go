package decoder

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"freya/common"
	"freya/handler"
)

type Decoder struct {
	structs    map[uint16]interface{}
}

var De *Decoder

func InitDecoder(handlers ...handler.TcpHandler) {
	De = newDecoder()
	De.register(handlers...)
}

func newDecoder() *Decoder {
	return &Decoder{
		structs: make(map[uint16]interface{}),
	}
}

func (d *Decoder) register(handlers ...handler.TcpHandler) {
	for _, v := range handlers {
		d.structs[v.MsgId] = v.Proto
	}
}

func (d *Decoder) clone(msgId uint16) (interface{}, error) {
	s, ok := d.structs[msgId]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Clone obj error, msgId:%v", msgId))
	}
	return reflect.New(reflect.TypeOf(s).Elem()).Interface(), nil
}

func (d *Decoder) DecodeHeader(rd io.Reader) (common.TcpHeader, error) {
	h := common.TcpHeader{}
	err := binary.Read(rd, binary.BigEndian, &h)
	return h, err
}

func (d *Decoder) DecodeBody(msgId uint16, data []byte) (interface{}, error) {
	inf, err := d.clone(msgId)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &inf)
	if err != nil {
		return nil, err
	}
	return inf, nil
}

