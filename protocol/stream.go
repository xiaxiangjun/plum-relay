package protocol

import "net"

// stream事件回调
type StreamEvent interface {
	OnClose(rs *RelayStream)
	OnRequrest(rs *RelayStream, req []byte) (res []byte, err error)
}

type RelayStream struct {
	conn net.Conn
}

// 创建一个新的stream
func NewRelayStream(conn net.Conn) *RelayStream {
	stream := &RelayStream{
		conn: conn,
	}

	// 初始化stream
	stream.init()

	return stream
}

func (self *RelayStream) init() {

}

func (self *RelayStream) Serve() {

}

func (self *RelayStream) Close() {
	self.conn.Close()
}
