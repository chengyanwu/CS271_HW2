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
	OutboundChannels []ConnectionInfo // all outbound channels
	InboundChannels []ConnectionInfo // connections from other clients to self
	TokenOutChannels []ConnectionInfo // for purposes for token passing and snapshot algorithm
	LoseChance uint // chance to lose the token after receiving it
	Token bool // false
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
	return fmt.Sprintf("\n===== Client Info =====\nProcessID: %d\nClientName: %s\nOutbound Channels: %s\nInbound Channels: %s\nToken/Marker Channels: %s\nLose Chance: %d\n", c.ProcessId, c.ClientName, c.OutboundChannels, c.InboundChannels, c.TokenOutChannels, c.LoseChance)
}

func (c ConnectionInfo) String() string {
	if c.ConnectionType != SNAPSHOTONLY {
		return c.ClientName
	} else {
		return c.ClientName + "'"
	}
}