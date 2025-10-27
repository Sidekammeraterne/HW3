package main

import (
	proto "ITUServer/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
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

	//todo: does it need some way to stop broadcast stream when leaving? Marie: it should probably close the stream - if it can as the stream was created by the server - done in server
}

func main() { //todo: should we split up into methods
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
	}

	//Allows the client to join the server.
	Message, err := client.JoinSystem(context.Background(), c.ClientInformation())
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

	//Listens on stream Read in a goRoutine, loops forever todo: if it can should probably be in its own method
	go func() {
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
	}()

	//listens to the terminal for new commands
	c.listenCommand() //todo: should be go?

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

// Increments the lamport clock by one
func (c *Client) IncrementLamport() int32 {
	c.LamportClock++
	return c.LamportClock
}

// Calls the IncrementLamport function, then returns ClientInformation //todo: rename since it is what happens only before joining system?
func (c *Client) ClientInformation() *proto.ClientInformation {
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
func (c *Client) leave() {
	//increment lamport
	c.IncrementLamport()
	//todo: these following lines were from CHATGPT
	_, _ = c.Client.LeaveSystem(context.Background(), &proto.ClientInformation{
		ClientId:     c.ClientId,
		LamportClock: c.LamportClock,
	})
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
