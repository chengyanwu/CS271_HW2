package main

import (
	"fmt"
	"net"
	"os"
	"time"
	// "strconv"
	"math/rand"
	"bufio"
	"strings"
	// "sync"
	client "example/users/client/client_interface"
)

const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"
	A           = "5000"
	B           = "5001"
	C           = "5002"
	D           = "5003"
	E           = "5004"
)

var myInfo client.ClientInfo
var connectedClients []client.ConnectedClient
var port string
var generator = rand.NewSource(myInfo.ProcessId)
var r = rand.New(generator)

func main() {
	processId := int64(os.Getpid())
	fmt.Println("My process ID:", processId)
	myInfo.ProcessId = processId

	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go [client name (A-E)] ")
		os.Exit(1)
	}

	myInfo.ClientName = os.Args[1]
	myInfo.LoseChance = 0
	myInfo.Token = false

	if myInfo.ClientName == "A" {
		port = A
		connectedClients = []client.ConnectedClient{
			client.ConnectedClient{ClientID: B, ConnectionType: client.BIDIRECTIONAL},
			// client.ConnectedClient{ClientID: D, ConnectionType: client.INCOMING},
			client.ConnectedClient{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: D, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "B" {
		port = B
		connectedClients = []client.ConnectedClient{
			client.ConnectedClient{ClientID: A, ConnectionType: client.BIDIRECTIONAL},
			// client.ConnectedClient{ClientID: C, ConnectionType: client.INCOMING},
			client.ConnectedClient{ClientID: D, ConnectionType: client.BIDIRECTIONAL},
			// client.ConnectedClient{ClientID: E, ConnectionType: client.INCOMING},
			client.ConnectedClient{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "C" {
		port = C
		connectedClients = []client.ConnectedClient{
			client.ConnectedClient{ClientID: B, ConnectionType: client.OUTGOING},
			// client.ConnectedClient{ClientID: D, ConnectionType: client.INCOMING},
			client.ConnectedClient{ClientID: D, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: E, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: A, ConnectionType: client.SNAPSHOTONLY},
		}
	} else if myInfo.ClientName == "D" {
		port = D
		connectedClients = []client.ConnectedClient{
			client.ConnectedClient{ClientID: A, ConnectionType: client.OUTGOING},
			client.ConnectedClient{ClientID: C, ConnectionType: client.OUTGOING},
			client.ConnectedClient{ClientID: E, ConnectionType: client.BIDIRECTIONAL},
			client.ConnectedClient{ClientID: B, ConnectionType: client.BIDIRECTIONAL},
		}
	} else if myInfo.ClientName == "E" {
		port = E
		connectedClients = []client.ConnectedClient{
			client.ConnectedClient{ClientID: B, ConnectionType: client.OUTGOING},
			client.ConnectedClient{ClientID: D, ConnectionType: client.BIDIRECTIONAL},
			client.ConnectedClient{ClientID: C, ConnectionType: client.SNAPSHOTONLY},
			client.ConnectedClient{ClientID: A, ConnectionType: client.SNAPSHOTONLY},
		}
	}

	// for _, connectedClient := range connectedClients {
		// if connectedClient.ConnectionType == 3 || connectedClient.ConnectionType == 1 {
			// clientPorts = append(clientPorts, connectedClient.ClientID)
		// }
		// connecttedClientInfo := fmt.Sprintf("Client %s: Connection Type: %d", connectedClient.ClientID, connectedClient.ConnectionType)
		// fmt.Println(connecttedClientInfo)
	// }

	// fmt.Printf("Client %s is outbound to ports: ", myInfo.ClientName)
	// fmt.Println(clientPorts)

	// start server to listen to other client connections
	go startServer(port, myInfo.ClientName)

	// wait for all clients to be set up
	fmt.Println("Press \"enter\" AFTER all clients' servers are set up to connect to them")
	fmt.Scanln()

	// TODO: establish outbound client connections
	establishClientConnections(connectedClients, myInfo.ClientName)

	// TODO handle user input
	// 1) request snapshot using Chandy-Lamport
	// 2) set loss probability
	// 3) initiate token transfer process

	// TODO test: print out inbound and outbound connections once all inbound connections are processed
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
	
			connInfo := client.ConnectionInfo{connection, string(clientName[:len(clientName)-1]), clientConnection.ConnectionType}
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

	// PROTOCOL: [TOKEN]
	writeToConnection(randomChannel.Connection, "TOKEN\n")
	fmt.Println("Passing token to client", randomChannel.ClientName)
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

		nameSlice := strings.Split(string(clientName[:len(clientName) - 1]), ":")

		if len(nameSlice) == 1 {
			go processInboundChannel(inboundChannel, nameSlice[0], client.INCOMING)
		} else {
			go processInboundChannel(inboundChannel, nameSlice[0], client.SNAPSHOTONLY)
		}
	}
}

// TODO: Handle inbound channel connection
func processInboundChannel(connection net.Conn, clientName string, connectionType client.ConnectionType) {
	fmt.Printf("Inbound client %s connected\n", clientName)
	inboundChannelInfo := client.ConnectionInfo{connection, clientName, connectionType}

	// Add new connection to self data structure
	myInfo.InboundChannels = append(myInfo.InboundChannels, inboundChannelInfo)

	// Case 1: receive TOKEN
	// 1) Wait 3 seconds after receive
	// 2) Calculate chances of losing token
	// 3) Either lose token or pass it on to a random outbound channel for the current process after waiting 1 second - create a new goroutine for this lol

	// Case 2: receive MARKER

	// Case 3: receive local snapshot
	for {
		action, err := bufio.NewReader(connection).ReadBytes('\n')

		handleError(err, fmt.Sprintf("Error receiving TOKEN or MARKER from client %s", clientName), connection)
		time.Sleep(3 * time.Second)

		actionInfoSlice := strings.Split(string(action[:len(action) - 1]), ":")
		
		if actionInfoSlice[0] == "TOKEN" {
			fmt.Println("Received TOKEN from client", clientName)
			
			myInfo.Token = true
			// lose token here
			if uint(r.Intn(100)) < myInfo.LoseChance {
				fmt.Println("TOKEN lost")
				myInfo.Token = false
			} else {
				go startTokenPassing()
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
