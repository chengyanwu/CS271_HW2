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

// connectionType: 1: outgoing 2: incoming 3: bi-directional
type ConnectedClient struct {
	ClientID string
	ConnectionType int64
}

func (c ClientInfo) String() string {
	// TODO
	return "TODO"
}