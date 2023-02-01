package client

import (
	"net"
)

type ClientInfo struct {
	ProcessId int64 // process ID identifier
	ClientName string // name identifier
	OutboundChannels []ConnectionInfo // connections to other clients
	InboundChannels []ConnectionInfo // connections from other clients to self
}

type ConnectionInfo struct {
	Connection net.Conn
	ClientName string
}

func (c ClientInfo) String() string {
	// TODO
	return "TODO"
}