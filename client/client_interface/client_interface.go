package client

import (
	"net"
	"fmt"
)

type ConnectionType int64

const (
	OUTGOING ConnectionType        = 0
	INCOMING ConnectionType        = 1
	BIDIRECTIONAL ConnectionType   = 2
	SNAPSHOTONLY ConnectionType    = 3
)

type ClientInfo struct {
	ProcessId int64 // process ID identifier
	ClientName string // name identifier
	OutboundChannels []ConnectionInfo // connections to other clients for purposes of token passing and snapshot algorithm
	InboundChannels []ConnectionInfo // connections from other clients to self
	ConsolidateChannels []ConnectionInfo // channels used to send local state to initiator
}

type ConnectionInfo struct {
	Connection net.Conn
	ClientName string
	ConnectionType ConnectionType
}

type ConnectedClient struct {
	ClientID string
	ConnectionType ConnectionType
}

func (c ClientInfo) String() string {
	return fmt.Sprintf("\n===== Client Info =====\nProcessID: %d\nClientName: %s\nOutbound Channels: %s\nInbound Channels: %s\n", c.ProcessId, c.ClientName, c.OutboundChannels, c.InboundChannels)
}

func (c ConnectionInfo) String() string {
	if c.ConnectionType != SNAPSHOTONLY {
		return c.ClientName
	} else {
		return c.ClientName + "'"
	}
}