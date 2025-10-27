package main

import (
	proto "ITUServer/grpc"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Creates a client struct, keeping track of the ChitChatClient, the client's id, the connection and the Lamportclock
type Client struct {
	Client       proto.ChitChatClient
	ClientId     int32
	Conn         *grpc.ClientConn
	LamportClock int32
	isActive     bool
	logFile      *os.File
}

func main() {
	// creates a connection related to a client who can speak to the specified port, grpc client can ask for a service
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	client := proto.NewChitChatClient(conn) //creates a new client and opens for the connection

	//creates a new instance of the object client
	c := Client{
		Client:       client,
		ClientId:     0, //starts as 0 to show that it has not gotten its id yet
		Conn:         conn,
		LamportClock: 0,
		isActive:     true,
	}

	//Allows the client to join the server.
	Message, err := client.JoinSystem(context.Background(), c.getClientInformation())
	if err != nil {
		log.Fatalf("did not recieve anything or failed to send %v", err)
	}
	c.ClientId = Message.Id

	//call the broadcast rpc call to open stream to receive messages from server
	stream, erro := client.Broadcast(context.Background(), &proto.ClientId{Id: c.ClientId})
	if erro != nil {
		log.Fatalf("did not recieve anything or failed to send %v", erro)
	}

	log.Printf("[Client] Joined Server: With id = %d at logical time %d", c.ClientId, c.LamportClock) //todo: is this where to log? Nope it is first offecially a part of the system after the rpc call broadcast (I want to change the names) - right place now

	c.setupLogging()
	//todo: den starter med at logge i fil her - den kan flyttes rundt på til hvornår end I vil have den. Bare vær sikker på, at det er efter, at den har fået clientid
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("failed to close log file: %v", err)
		}
	}(c.logFile) //closes the log file when the main function ends //todo: før hed den bare defer logFile.Close(). Resten rundt om er skrevet af GoLand

	//listen to the stream
	go c.listenBroadcast(stream)

	//listens to the terminal for new commands
	c.listenCommand()

}

// sets up logging both into a file and the terminal
func (c *Client) setupLogging() {
	//creates the file address
	clientIdString := strconv.Itoa(int(c.ClientId))
	fileAddress := "client" + clientIdString + ".log"

	//creates the file (or overwrites it if it already exists)
	logFile, err := os.Create(fileAddress)
	if err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}

	c.logFile = logFile //saves the logfile to the client struct

	log.SetOutput(io.MultiWriter(os.Stdout, logFile)) //sets the output to both the terminal and the file

}

// listens for broadcast messages from server
func (c *Client) listenBroadcast(stream grpc.ServerStreamingClient[proto.BroadcastMessage]) {
	//stops the goroutine if the client has left the system
	if !c.isActive {
		return
	}
	log.Printf("[Client] listening for broadcast messages at time %d", c.LamportClock)
	for {
		Message, err := stream.Recv()
		if err != nil {
			log.Printf("failed to receive message: %v", err)
			return
		}
		//When receiving check max lamport
		c.updateLamportOnReceive(Message.LamportClock)
		log.Printf("[Client][Broadcast] Received: %s (server L=%d, local L=%d)", Message.MessageContent, Message.LamportClock, c.LamportClock) //what's received from server
	}
}

// commands:
// message [insert text]. publishes a message
// leave. client exits chat
func (c *Client) listenCommand() {
	var input string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type command 'message ...' or 'leave'")

	for {
		if !scanner.Scan() {
			//client for some reason not responding
			c.leave()
			return
		}
		input = scanner.Text() //holds input

		switch {
		case input == "leave":
			c.leave()
			return

		case strings.HasPrefix(input, "message "):
			text := input[8:] //holds everything after "message "
			//if message is empty
			if text == "" {
				fmt.Println("message is empty, try again")
				continue
			}
			c.sendMessage(text)
		case input == "":
			//if input was empty, were ignoring it
		default:
			fmt.Println("unknown command, try again")
		}

	}

}

// IncrementLamport increments the lamport clock by one
func (c *Client) IncrementLamport() int32 {
	c.LamportClock++
	return c.LamportClock
}

// getClientInformation calls the IncrementLamport function, then returns ClientInformation //todo: rename since it is what happens only before joining system?
func (c *Client) getClientInformation() *proto.ClientInformation {
	c.IncrementLamport()
	return &proto.ClientInformation{
		ClientId:     c.ClientId,
		LamportClock: c.LamportClock,
	}
}

// when receiving from server, check max and increment
func (c *Client) updateLamportOnReceive(serverLamport int32) {
	if serverLamport > c.LamportClock {
		c.LamportClock = serverLamport
	}
	c.IncrementLamport()
}

// Increments the lamport clock, sets isActive to false and calls the leave rpc call
func (c *Client) leave() {
	clientInfo := c.getClientInformation() //increments lamport clock and returns client information
	c.isActive = false
	_, err := c.Client.LeaveSystem(context.Background(), clientInfo)
	if err != nil {
		log.Fatalf("failed to leave the system %v", err)
	}
	log.Printf("[Client] requested leave at time %d", c.LamportClock) //todo
}

// Publish RPC (Lamport++ before send)
func (c *Client) sendMessage(text string) {
	c.IncrementLamport()
	_, err := c.Client.PublishMessage(context.Background(), &proto.Message{
		ClientId:       c.ClientId,
		LamportClock:   c.LamportClock,
		MessageContent: text,
	})
	if err != nil {
		log.Printf("[Client] publish failed: %v", err)
		return
	}
	log.Printf("[Client] sent: %q at time %d)", text, c.LamportClock)
}
