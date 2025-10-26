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
	clients      map[int32]*Client //map of active clients
	lamportClock int32
	clientNextId int32
	//todo: maybe use a Lock
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
	lamport := s.updateLamportClockOnReceive(in.LamportClock) //check max and update local lamport
	//include server lamport in broadcast st. clients can check for max
	s.BroadcastService(fmt.Sprintf("(%d) %d: %s", lamport, in.ClientId, in.MessageContent), lamport)
	return &proto.Empty{}, nil //returns the pointer to the student in memory - what we are encourages to do [should be able to see this from the function declaration (?)]
}

func (s *ChitChatServer) JoinSystem(ctx context.Context, in *proto.ClientInformation) (*proto.ClientId, error) {
	s.updateLamportClockOnReceive(in.LamportClock) //check max lamport todo: is this wrong

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
	delete(s.clients, in.ClientId)
	lamport := s.updateLamportClockOnReceive(in.LamportClock)
	s.BroadcastService(
		fmt.Sprintf("Client %d left ChitChat at time %d", in.ClientId, lamport), lamport,
	)
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) Broadcast(in *proto.ClientId, stream proto.ChitChat_BroadcastServer) error { //todo, is this broadcast only for when joining?
	client, ok := s.clients[in.Id]
	if !ok {
		log.Fatalf("client not found %v", in.Id) //todo: unsure if fatalf should be used, or error. Fatal exits program
	}
	client.Stream = stream
	//start goRoutine
	go s.start_client(client)

	//todo: comment below code Sara
	// JOIN announcement (server-side local event)
	lamport := s.updateLamportClockOnReceive(0) // checks max(S,0) + 1
	s.BroadcastService(
		fmt.Sprintf("Participant %d joined Chit Chat at logical time %d", client.ClientId, lamport),
		lamport,
	)

	// Keep the stream open until the client disconnects
	<-stream.Context().Done()

	// LEAVE announcement happens LATER, only after disconnect
	delete(s.clients, client.ClientId)
	lamport = s.updateLamportClockOnReceive(0)
	s.BroadcastService(
		fmt.Sprintf("Participant %d left Chit Chat at logical time %d", client.ClientId, lamport),
		lamport,
	)
	return nil
}

func (s *ChitChatServer) BroadcastService(broadcastMessage string, lamport int32) { //todo: beware the new client does not receive
	for ClientId, client := range s.clients {
		client.BroadcastChannel <- &proto.BroadcastMessage{MessageContent: broadcastMessage, LamportClock: lamport}
		log.Printf("[Server] Sent to %d: %s", ClientId, broadcastMessage)
	}
	//todo we should “support multiple concurrent client connections without blocking message delivery.”
	// which means for should maybe be switch with a default case that handles slow client
	// but don't really know what should happen in the case
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
			} //todo: same here, fatal or error?
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
	log.Println("ChitChat server listening on :5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("Did not work")
	}
}

// utility method that on message receive checks max lamport and updates local
func (s *ChitChatServer) updateLamportClockOnReceive(remoteLamport int32) int32 {
	if remoteLamport > s.lamportClock {
		s.lamportClock = remoteLamport
	}
	s.lamportClock++
	return s.lamportClock
}

/* Everytime something is changed in proto file, run the following command in the terminal:
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto.proto
*/
