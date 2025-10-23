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

// Increments the lamport clock by one
func (c *Client) IncrementLamport() {
	c.LamportClock++
}

// Calls the IncrementLamport function, then returns ClientInformation
func (c *Client) ClientInformation() *proto.ClientInformation {
	c.IncrementLamport()
	return &proto.ClientInformation{
		ClientId:     c.ClientId,
		LamportClock: c.LamportClock,
	}
}

func main() {
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

	//listens to the terminal for new commands
	go listenCommand()

	//method to listen for broadcasts?

}

func listenCommand() {
	var input string //todo: use var
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'leave' or 'message ...'")

	for scanner.Scan() {
		
	}
}
