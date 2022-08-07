package config

import "plum-relay/protocol"

type Parameter struct {
	Stream *protocol.RelayStream
	Store  *MemoryStore
}
