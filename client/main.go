package main

import (
	"fmt"
	"net"
	"os"
	// "strconv"
	"time"
	"bufio"
	// "strings"
	// "sync"
	"example/users/client/client_interface"
)

const (
	SERVER_HOST = "localhost"
	SERVER_TYPE = "tcp"
)

var myInfo client.ClientInfo

func main() {
	processId := int64(os.Getpid())
	fmt.Println("My process ID:", processId)
	myInfo.ProcessId = processId

	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go [client name (A-E)] [port number] [client ports to connect to] ...")
		os.Exit(1)
	}

	myInfo.ClientName = os.Args[1]
	port := os.Args[2]
	clientPorts := os.Args[3:]

	// start server to listen to other client connections
	go startServer(port, myInfo.ClientName)

	// wait for all clients to be set up
	fmt.Println("Press \"enter\" AFTER all clients' servers are set up to connect to them")
	fmt.Scanln()

	// TODO: establish outbound client connections
	establishClientConnections(clientPorts, myInfo.ClientName)

	// TODO handle user input:
	// 1) request snapshot using Chandy-Lamport
	// 2) set loss probability
	// 3) initiate token transfer process

	// TODO test: print out inbound and outbound connections once all inbound connections are processed
	takeUserInput()
}

func establishClientConnections(clientPorts []string, name string) {
	var outboundChannels []client.ConnectionInfo

	for _, port := range clientPorts {
		connection, err := net.Dial(SERVER_TYPE, SERVER_HOST + ":" + port)

		handleError(err, fmt.Sprintf("Couldn't connect to client's server with port: %s\n", port), connection)
		// PROTOCOL: receive name from connection
		clientName, err := bufio.NewReader(connection).ReadBytes('\n')
		handleError(err, fmt.Sprintf("Didn't receive client's server's response on port: %s\n", port), connection)

		connInfo := client.ConnectionInfo{connection, string(clientName[:len(clientName) - 1])}
		fmt.Println("Successfully established outbound channel to client with name:", connInfo.ClientName)

		outboundChannels = append(outboundChannels, connInfo)

		// PROTOCOL: send self name to connection
		writeToConnection(connection, name + "\n")

		// TODO: handle outbound channels
	}

	myInfo.OutboundChannels = outboundChannels
}

func takeUserInput() {
	fmt.Println("All outbound connections established")
	var action string

	for {
		_, err := fmt.Scanln(&action)

		if err != nil {
			fmt.Println("Error occurred when scanning input")
		} else if action == "p" {
			fmt.Println(myInfo)
		} else {
			fmt.Println(action)
		}
	}
}

func startServer(port string, name string) {
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST + ":" + port)
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
		writeToConnection(inboundChannel, name + "\n")

		// PROTOCOL: receive inbound client name
		clientName, err := bufio.NewReader(inboundChannel).ReadBytes('\n')

		handleError(err, "Didn't receive connected client's name.", inboundChannel)

		// handle inbound channel
		go processInboundChannel(inboundChannel, string(clientName[:len(clientName) - 1]))
	}
}

// TODO: handle inbound channel connection
func processInboundChannel(connection net.Conn, clientName string) {
	fmt.Printf("Inbound client %s connected\n", clientName)
	inboundChannelInfo := client.ConnectionInfo{connection, clientName}

	// add new connection to self data structure
	myInfo.InboundChannels = append(myInfo.InboundChannels, inboundChannelInfo)
}

func handleError(err error, message string, connection net.Conn) {
	if err != nil {
		fmt.Println(message, err.Error())
		connection.Close()

		os.Exit(1)
	}
}

func writeToConnection(connection net.Conn, message string) {
	time.Sleep(3 * time.Second)
	_, err := connection.Write([]byte(message))

	handleError(err, "Error writing.", connection)
}

