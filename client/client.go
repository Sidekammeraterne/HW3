package main

import (
	proto "ITUServer/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Creates a client struct, keeping track of the clients id and the Lamportclock
type Client struct {
	Client       proto.ChitChatClient
	ClientId     int32
	Conn         *grpc.ClientConn
	LamportClock int32
}

func main() { //todo: should we split up into methods
	// creates a connection related to a client who can speak to the specified port, grpc client can ask for a service
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials())) //security, don't think much about it. "Boilerplate".
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

	log.Printf("[Client] Joined Server: id=%d (L=%d)", c.ClientId, c.LamportClock) //todo: is this where to log?

	//method to listen for broadcasts?
	//call the broadcast rpc call to open stream to receive messages from server
	stream, erro := client.Broadcast(context.Background(), &proto.ClientId{Id: c.ClientId})
	if erro != nil {
		log.Fatalf("did not recieve anything or failed to send %v", erro)
	}

	//Listens on stream Read in a goRoutione, loops forever todo: if it can should probably be in its own method
	go func() {
		for {
			Message, err := stream.Recv()
			if err != nil {
				log.Printf("failed to receive message: %v", err)
				return
			}
			//When receiving check max lamport
			c.updateLamportOnReceive(Message.LamportClock)
			log.Printf("[Client][Deliver]Lamport Clock: %d, Message: %s", c.LamportClock, Message)
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

		switch input {
		case "leave":
			c.leave()
			return

		case "message":
			text := input[8:] //holds everything after "message "
			//if message is empty
			if text == "" {
				fmt.Println("message is empty, try again")
				continue
			}
			c.sendMessage(text)
		case "":
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

// Calls the IncrementLamport function, then returns ClientInformation //todo: rename? is it what happend before joining system?
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
	log.Printf("[Client] requested leave (L=%d)", c.LamportClock) //todo
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
	log.Printf("[Client] sent: %q (L=%d)", text, c.LamportClock)
}
