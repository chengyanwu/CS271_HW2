package client

import (
	"fmt"
	"net"
)

type ConnectionType int64

const (
	OUTGOING      ConnectionType = 0
	INCOMING      ConnectionType = 1
	BIDIRECTIONAL ConnectionType = 2
	SNAPSHOTONLY  ConnectionType = 3
)

type ClientInfo struct {
	ProcessId        int64            // process ID identifier
	ClientName       string           // name identifier
	OutboundChannels []ConnectionInfo // all outbound channels
	InboundChannels  []*ConnectionInfo // connections from other clients to self
	TokenOutChannels []ConnectionInfo // for purposes for token passing and snapshot algorithm
	LoseChance       uint             // chance to lose the token after receiving it (0 - 100)
	Token            bool             // whether process has token
	TokenForSnapshot bool             // whether process has token when most recent snapshot is taken
	Recording        bool             // whether channel is recording incoming messages on all incoming channels - ex: has received a MARKER for the first time or is initiator
	Initiator        string           // client name of the initiator of the process
}

type ConnectionInfo struct {
	Connection       net.Conn
	ClientName       string         // client name on the other end of the connection
	ConnectionType   ConnectionType
	Recording        bool           // state of channel: defaults to true, set to false when P receives a MARKER for the first time over the channel
	IncomingMessages []Message      // messages on the incoming channel for the purposes of TOKEN passing, should be emptied when snapshot terminates
}

// incoming messages on the channel when snapshot is initiated
type Message struct {
	SenderName   string
	Message      string
	// ReceiverName string
}

type ConnectedClient struct {
	ClientID       string
	ConnectionType ConnectionType
}

func (c ClientInfo) String() string {
	return fmt.Sprintf("\n===== Client Info =====\nProcessID: %d\nClientName: %s\nOutbound Channels: %s\nInbound Channels: %s\nToken/Marker Channels: %s\nLose Chance: %d\nRecording: %v", c.ProcessId, c.ClientName, c.OutboundChannels, c.InboundChannels, c.TokenOutChannels, c.LoseChance, c.Recording)
}

func (c ConnectionInfo) String() string {
	if c.ConnectionType != SNAPSHOTONLY {
		return c.ClientName
	} else {
		return c.ClientName + "'"
	}
}
