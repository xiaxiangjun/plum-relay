package protocol

import "net"

type RelayStream struct {
	conn net.Conn
}

func NewRelayStream(conn net.Conn) *RelayStream {
	stream := &RelayStream{
		conn: conn,
	}

	stream.init()

	return stream
}

func (self *RelayStream) init() {
	
}
