package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"bufio"
	"math/rand"
	"strconv"
	"strings"

	// "sync"
	client "example/users/client/client_interface"
)

const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"

	A = "6000"
	B = "6001"
	C = "6002"
	D = "6003"
	E = "6004"
)

var myInfo client.ClientInfo
var connectedClients []client.ConnectedClient
var port string
var generator rand.Source
var r = rand.New(generator)
var snapshotString string // TODO: very important - consolidate messages to initiator here
var counter = 0
var globalSnapShot []string

func main() {
	processId := int64(os.Getpid())
	generator = rand.NewSource(processId)
	fmt.Println("My process ID:", processId)
	myInfo.ProcessId = processId

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go [client name (A-E)] ")
		os.Exit(1)
	}

	myInfo.ClientName = os.Args[1]
	myInfo.LoseChance = 0
	myInfo.Token = false
	myInfo.Recording = false

	if myInfo.ClientName == "A" {
		port = A
		connectedClients = []client.ConnectedClient{
			{ClientID: B, ConnectionType: client.BIDIRECTIONAL},
			{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: D, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "B" {
		port = B
		connectedClients = []client.ConnectedClient{
			{ClientID: A, ConnectionType: client.BIDIRECTIONAL},
			{ClientID: D, ConnectionType: client.BIDIRECTIONAL},
			{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "C" {
		port = C
		connectedClients = []client.ConnectedClient{
			{ClientID: B, ConnectionType: client.OUTGOING},
			{ClientID: D, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: A, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "D" {
		port = D
		connectedClients = []client.ConnectedClient{
			{ClientID: A, ConnectionType: client.OUTGOING},
			{ClientID: C, ConnectionType: client.OUTGOING},
			{ClientID: E, ConnectionType: client.BIDIRECTIONAL},
			{ClientID: B, ConnectionType: client.BIDIRECTIONAL},
		}
	} else if myInfo.ClientName == "E" {
		port = E
		connectedClients = []client.ConnectedClient{
			{ClientID: B, ConnectionType: client.OUTGOING},
			{ClientID: D, ConnectionType: client.BIDIRECTIONAL},
			{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			{ClientID: A, ConnectionType: client.SNAPSHOTONLY},
		}
	}

	// start server to listen to other client connections
	go startServer(port, myInfo.ClientName)

	// wait for all clients to be set up
	fmt.Println("Press \"enter\" AFTER all clients' servers are set up to connect to them")
	fmt.Scanln()

	// TODO: establish outbound client connections
	establishClientConnections(connectedClients, myInfo.ClientName)

	// 1) request snapshot using Chandy-Lamport
	takeUserInput()
}

// Establish outbound communication channels to other clients
func establishClientConnections(clientPorts []client.ConnectedClient, name string) {
	var outboundChannels []client.ConnectionInfo
	var tokenChannels []client.ConnectionInfo

	// loop through client connections
	for _, clientConnection := range clientPorts {
		connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+clientConnection.ClientID)

		if clientConnection.ConnectionType == client.BIDIRECTIONAL || clientConnection.ConnectionType == client.OUTGOING || clientConnection.ConnectionType == client.SNAPSHOTONLY {
			handleError(err, fmt.Sprintf("Couldn't connect to client's server with port: %s\n", port), connection)

			// PROTOCOL: receive name from connection
			clientName, err := bufio.NewReader(connection).ReadBytes('\n')
			handleError(err, fmt.Sprintf("Didn't receive client's server's response on port: %s\n", port), connection)

			connInfo := client.ConnectionInfo{connection, string(clientName[:len(clientName)-1]), clientConnection.ConnectionType, true, []client.Message{}}
			fmt.Println("Successfully established outbound channel to client with name:", connInfo.ClientName)

			if clientConnection.ConnectionType != client.SNAPSHOTONLY {
				tokenChannels = append(tokenChannels, connInfo)
			}

			outboundChannels = append(outboundChannels, connInfo)

			// PROTOCOL: send self name to connection, along with indicator if the channel is solely for sending local state
			if clientConnection.ConnectionType == client.SNAPSHOTONLY {
				writeToConnection(connection, name+":"+"snapshot_only"+"\n")
			} else {
				writeToConnection(connection, name+"\n")
			}
		}
	}

	myInfo.OutboundChannels = outboundChannels
	myInfo.TokenOutChannels = tokenChannels
}

func takeUserInput() {
	fmt.Println("All outbound connections established")
	var action string
	var loseChance uint

	fmt.Println("===== Actions =====\np - print client info\nl - set chance of losing token\ns - initiate snapshot\nt - start token passing\n===================")
	for {
		_, err := fmt.Scanln(&action)

		if err != nil {
			fmt.Println("Error occurred when scanning input")
		} else if action == "p" {
			fmt.Println(myInfo)
		} else if action == "l" {
			fmt.Print("Enter chance of losing token (0 - 100): ")
			_, err := fmt.Scanln(&loseChance)
			if err != nil {
				fmt.Println("Error occurred when scanning input to set chance of losing token")
			}
			if loseChance > 100 {
				fmt.Println("Please use a number between 0 and 100, inclusive")
			} else {
				myInfo.LoseChance = loseChance
			}
		} else if action == "t" {
			// start token passing process
			startTokenPassing() // list some tokens
		} else if action == "s" {
			// TODO: Ian
			myInfo.Initiator = myInfo.ClientName
			myInfo.TokenForSnapshot = myInfo.Token
			startSnapShot(myInfo.Initiator)
			myInfo.Recording = true
		} else {
			fmt.Println("Invalid action:", action)
		}
	}
}

func startTokenPassing() {
	time.Sleep(1 * time.Second)
	// pick random outbound channel
	randomIndex := r.Intn(len(myInfo.TokenOutChannels))
	randomChannel := myInfo.TokenOutChannels[randomIndex]

	// pass token to random channel
	myInfo.Token = false

	// PROTOCOL: [TOKEN, initiator name]
	writeToConnection(randomChannel.Connection, "TOKEN\n")
	fmt.Println("Passing TOKEN to client", randomChannel.ClientName)
}

func startSnapShot(initiator string) {
	// record its own local state

	// send marker message (Initiator Sends MARKER) on all outgoing channels
	// PROTOCOL: [MARKER, initiator]
	markerMessage := "MARKER: " + initiator + "\n"
	for _, channel := range myInfo.TokenOutChannels {
		writeToConnection(channel.Connection, markerMessage)
		fmt.Println("Snapshot in progress: Sending MARKER to client", channel.ClientName)
	}
}

func startServer(port string, name string) {
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err.Error())

		os.Exit(1)
	}

	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + port)

	for {
		// inbound connection
		inboundChannel, err := server.Accept()
		handleError(err, "Error accepting client.", inboundChannel)

		// PROTOCOL: broadcast self name to connection
		writeToConnection(inboundChannel, name+"\n")

		// PROTOCOL: receive inbound client name
		clientName, err := bufio.NewReader(inboundChannel).ReadBytes('\n')
		handleError(err, "Didn't receive connected client's name.", inboundChannel)

		nameSlice := strings.Split(string(clientName[:len(clientName)-1]), ":")

		if len(nameSlice) == 1 {
			go processInboundChannel(inboundChannel, nameSlice[0], client.INCOMING)
		} else {
			go processInboundChannel(inboundChannel, nameSlice[0], client.SNAPSHOTONLY)
		}
	}
}

// Handle inbound channel connection
func processInboundChannel(connection net.Conn, clientName string, connectionType client.ConnectionType) {
	fmt.Printf("Inbound client %s connected\n", clientName)
	inboundChannelInfo := client.ConnectionInfo{connection, clientName, connectionType, true, []client.Message{}}

	// Add new connection to self data structure
	myInfo.InboundChannels = append(myInfo.InboundChannels, &inboundChannelInfo)

	for {
		action, err := bufio.NewReader(connection).ReadBytes('\n')
		handleError(err, fmt.Sprintf("Error receiving TOKEN or MARKER from client %s", clientName), connection)

		actionInfoSlice := strings.Split(string(action[:len(action)-1]), " ")
		time.Sleep(3 * time.Second)

		// Case 1: receive TOKEN
		// 1) Calculate chances of losing token
		// 2) Either lose token or pass it on to a random outbound channel for the current process after waiting 1 second - create a new goroutine for this lol
		if actionInfoSlice[0] == "TOKEN" {
			fmt.Println("Received TOKEN from client", clientName)

			// save TOKEN message on channel if channel is recording
			if myInfo.Recording && inboundChannelInfo.Recording {
				fmt.Println("Snapshot in progress: Recording TOKEN message from client", clientName)
				inboundChannelInfo.IncomingMessages = append(inboundChannelInfo.IncomingMessages, client.Message{clientName, "TOKEN"})
			}

			myInfo.Token = true
			// lose token
			if uint(r.Intn(100)) < myInfo.LoseChance {
				fmt.Println("TOKEN lost")
				myInfo.Token = false
			} else {
				go startTokenPassing()
			}

			// Case 2: receive MARKER
		} else if actionInfoSlice[0] == "MARKER:" {
			// Scenario 1: receive MARKER for the first time, mark channel empty and sends MARKER to all outbound channels
			if !myInfo.Recording {
				inboundChannelInfo.Recording = false
				myInfo.Initiator = actionInfoSlice[1]
				fmt.Printf("Snapshot starting: Received MARKER from initiator client %s\n", myInfo.Initiator)
				// myInfo.TokenForSnapshot = myInfo.Token

				// markerMessage := "MARKER\n"
				// for _, channel := range myInfo.OutboundChannels {
				// 	if channel.ConnectionType == client.OUTGOING || channel.ConnectionType == client.BIDIRECTIONAL {
				// 		writeToConnection(channel.Connection, markerMessage)
				// 		fmt.Println("Markers sent to:", channel.ClientName)
				// 	}
				// }
				myInfo.TokenForSnapshot = myInfo.Token
				go startSnapShot(myInfo.Initiator)
				myInfo.Recording = true
				// Scenario 2: if process already received MARKER but receives another marker on this channel, stops recording on this channel
			} else {
				fmt.Println("Snapshot in progress: Received MARKER from client", clientName)
				// stop recording - channel state is finalized
				inboundChannelInfo.Recording = false
			}

			// check if MARKERS have been received on all incoming channels, and send messages to initiator if true
			go snapshotTermination()
		} else if actionInfoSlice[0] == "SNAPSHOT" {
			if myInfo.Initiator != myInfo.ClientName {
				panic(fmt.Sprintf("Only the initiator %s, not %s, should be receiving SNAPSHOT messages", myInfo.Initiator, myInfo.ClientName))
			}

			counter += 1

			message := clientName + ": " + actionInfoSlice[1]
			globalSnapShot = append(globalSnapShot, message)

			fmt.Println("Received new SNAPSHOT from client", clientName)
			fmt.Println("Local state:", actionInfoSlice[1])

			if counter == 4 {
				counter = 0
				fmt.Println("Snaptshot Completed")
				message := myInfo.ClientName + ": " + strconv.FormatBool(myInfo.TokenForSnapshot)
				globalSnapShot = append(globalSnapShot, message)
				fmt.Println("Global Snapshot: ", globalSnapShot)
				globalSnapShot = nil

			}

		}

		// TODO: Case 3: receive local snapshot (don't need to sleep), panic if initiator of the client received isn't self
	}
}

// send local snapshot to initiator if MARKERS have been received on all incoming channels, clear inbound channel messages
// set myInfo.Recording to false and all inboundChannel.Recording to true!
func snapshotTermination() {
	snapshotComplete := true
	// check all inbound channels that aren't for local state passing
	for _, inboundChannel := range myInfo.InboundChannels {
		// if channel is still recording, that means it is still waiting for MARKER message
		if inboundChannel.ConnectionType == client.INCOMING && inboundChannel.Recording {
			snapshotComplete = false
		}
	}

	// PROTOCOL: [SNAPSHOT:TOKEN STATE:SENDER,MESSAGE:SENDER,MESSAGE:...]
	var snapshotInfo = []byte("SNAPSHOT ")
	if snapshotComplete {
		if !myInfo.Recording {
			panic("Snapshot can't be complete if process hasn't started recording channel information.")
		} else {
			myInfo.Recording = false

			snapshotInfo = fmt.Appendf(snapshotInfo, ":%v", myInfo.TokenForSnapshot)
			for _, inboundChannel := range myInfo.InboundChannels {
				if inboundChannel.ConnectionType == client.INCOMING {
					inboundChannel.Recording = true

					for _, message := range inboundChannel.IncomingMessages {
						snapshotInfo = fmt.Appendf(snapshotInfo, ":%s,%s", message.SenderName, message.Message)
					}

					// clear incoming messages from channel
					inboundChannel.IncomingMessages = []client.Message{}
				}
			}
			snapshotInfo = fmt.Appendf(snapshotInfo, "\n")

			// send message to initiator
			if myInfo.Initiator != myInfo.ClientName {
				for _, channel := range myInfo.OutboundChannels {
					if myInfo.Initiator == channel.ClientName {
						writeToConnection(channel.Connection, string(snapshotInfo))
						fmt.Println("Sending SnapshotInfo: ", string(snapshotInfo))
						break
					}
				}
			}
		}
	}
}

func handleError(err error, message string, connection net.Conn) {
	if err != nil {
		fmt.Println(message, err.Error())
		connection.Close()

		os.Exit(1)
	}
}

func writeToConnection(connection net.Conn, message string) {
	// time.Sleep(3 * time.Second)
	_, err := connection.Write([]byte(message))

	handleError(err, "Error writing.", connection)
}

// not sure
// func markChannelEmpty(clientName string) {
// 	for _, channel := range myInfo.InboundChannels {
// 		if channel.ClientName == clientName {
// 			// mark channel as empty
// 			channel.State = 0
// 			fmt.Printf("Inbound Channel %s Marked as Empty", channel.ClientName)
// 		}
// 	}
// }

// // not sure
// func checkChannelActive(clientName string) bool {
// 	for _, channel := range myInfo.InboundChannels {
// 		if channel.ClientName == clientName {
// 			if channel.State == 0 {
// 				return false
// 			}
// 		}

// 	}
// 	return true
// }
