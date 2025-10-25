package main

import (
	proto "ITUServer/grpc" //make connection
	"context"              //make connection - the context of the connection
	"errors"               //create custom errors
	"fmt"
	"log"          //logs - used to keep track of messages
	"net"          //make connection to net
	"unicode/utf8" //used to verify number of chars

	"google.golang.org/grpc"
)

type ChitChatServer struct {
	//part of the proto - we are creating an Unimplemented server
	proto.UnimplementedChitChatServer
	clients      map[int32]*Client //map of active clients todo: should maybe be list
	lamportClock int32             //todo: implement
	clientNextId int32
}

type Client struct {
	ClientId         int32
	BroadcastChannel chan *proto.BroadcastMessage
	Stream           proto.ChitChat_BroadcastServer
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) { //if we want something with an empty in it
	//validating length of Message is <128 using RuneCountInString
	if utf8.RuneCountInString(in.MessageContent) > 128 {
		return nil, errors.New("message is too long, must be under 128 characters")
	}
	return &proto.Empty{}, nil //returns the pointer to the student in memory - what we are encourages to do [should be able to see this from the function declaration (?)]
}

func (s *ChitChatServer) JoinSystem(ctx context.Context, in *proto.ClientInformation) (*proto.ClientId, error) {
	//todo: assign client ID and make message that it returns
	clientId := s.clientNextId
	s.clientNextId++
	//add client to clients list
	client := &Client{
		ClientId:         clientId,
		BroadcastChannel: make(chan *proto.BroadcastMessage, 10), //buffered channel with size 10
	}

	s.clients[clientId] = client

	return &proto.ClientId{Id: clientId}, nil
}

func (s *ChitChatServer) LeaveSystem(ctx context.Context, in *proto.ClientInformation) (*proto.Empty, error) {
	//remove client ID from list
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) Broadcast(ctx context.Context, in *proto.ClientId, stream proto.ChitChat_BroadcastServer) error {
	client, ok := s.clients[in.Id]
	if !ok {
		log.Fatalf("client not found %v", in.Id)
	}
	client.Stream = stream
	//todo: start go routine
	go s.start_client(client)
	//todo: log message //todo: both to terminal and file

	//todo: broadcast message: "Participant X joined Chit Chat at logical time L". //todo: should be in its own message
	message := fmt.Sprintf("Participant %d joined Chit Chat at logical time %d", client.ClientId, s.lamportClock)
	s.BroadcastService(message)
	return nil
}

func (s *ChitChatServer) BroadcastService(broadcastMessage string) { //todo: beware the new client does not receive
	for ClientId, client := range s.clients {
		client.BroadcastChannel <- &proto.BroadcastMessage{MessageContent: broadcastMessage}
		log.Printf("[Server] Sent to %s: %s", ClientId, broadcastMessage)
	}
	//find logical time
	//find active clients and then corresponding goroutines
	//send message (broadcastMessage) to those goroutines
}

// listens for broadcast messages and sends it through the stream to the client
func (s *ChitChatServer) start_client(c *Client) {
	for {
		select {
		case message := <-c.BroadcastChannel:
			err := c.Stream.Send(message)
			if err != nil {
				log.Fatalf("could not send message to client: %v", err)
			}
		}
	}
}

func main() {
	server := &ChitChatServer{clients: make(map[int32]*Client), lamportClock: 0, clientNextId: 1}
	//creation of clients here
	server.start_server()
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()              //creates new gRPC server instance
	listener, err := net.Listen("tcp", ":5050") //listens on TCP port using net.Listen
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChitChatServer(grpcServer, s) //registers the server implementation with gRPC
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("Did not work")
	}
}

/* Everytime something is changed in this file, run the following command in the terminal:
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto.proto
*/
