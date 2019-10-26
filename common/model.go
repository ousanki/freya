package common

type TcpHeader struct {
	MsgId    uint16
	Length   uint16
	GWID     uint16
	ClientID uint32
	ProxyID  uint64
}
