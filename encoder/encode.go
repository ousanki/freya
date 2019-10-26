package encoder

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/ousanki/freya/common"
)

func Encode(msgId uint16, proxyId uint64, clientId uint32, data interface{}) ([]byte, error) {
	var header common.TcpHeader
	body, err := json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}
	header.Length = uint16(len(body))
	header.MsgId = msgId
	header.ProxyID = proxyId
	header.ClientID = clientId

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		return []byte{}, err
	}
	err = binary.Write(buf, binary.BigEndian, body)
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
