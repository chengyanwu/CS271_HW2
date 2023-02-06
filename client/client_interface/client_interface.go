package client

import (
	"net"
	"fmt"
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
	return fmt.Sprintf("\n===== Client Info =====\nProcessID: %d\nClientName: %s\nOutbound Channels: %s\nInbound Channels: %s\n", c.ProcessId, c.ClientName, c.OutboundChannels, c.InboundChannels)
}

func (c ConnectionInfo) String() string {
	return c.ClientName
}