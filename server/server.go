package main

import (
	proto "ITUServer/grpc" //make connection
	"bufio"
	"context" //make connection - the context of the connection
	"errors"  //create custom errors
	"fmt"
	"io"
	"log" //logs - used to keep track of messages
	"net" //make connection to net
	"os"
	"unicode/utf8" //used to verify number of chars

	"google.golang.org/grpc"
)

type ChitChatServer struct {
	//part of the proto - we are creating an Unimplemented server
	proto.UnimplementedChitChatServer
	clients      map[int32]*Client //map of active clients
	lamportClock int32
	clientNextId int32
	logFile      *os.File
}

type Client struct {
	ClientId         int32
	BroadcastChannel chan *proto.BroadcastMessage
	Stream           proto.ChitChat_JoinSystemServer
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) {
	//validating length of Message is <128 using RuneCountInString
	if utf8.RuneCountInString(in.MessageContent) > 128 {
		return nil, errors.New("message is too long, must be under 128 characters")
	}
	lamport := s.updateLamportClockOnReceive(in.LamportClock) //check max and update local lamport
	//include server lamport in broadcast st. clients can check for max
	s.BroadcastService(fmt.Sprintf("Client %d published the message: \"%s\" at logical time %d", in.ClientId, in.MessageContent, lamport))
	return &proto.Empty{}, nil //returns the pointer to the client in memory
}

func (s *ChitChatServer) RegisterClient(ctx context.Context, in *proto.ClientInformation) (*proto.ClientId, error) {
	s.updateLamportClockOnReceive(in.LamportClock) //check max lamport

	// Assign client id
	clientId := s.clientNextId
	s.clientNextId++

	//Local event: add client to clients list
	s.lamportClock++
	client := &Client{
		ClientId:         clientId,
		BroadcastChannel: make(chan *proto.BroadcastMessage, 10), //buffered channel with size 10
	}
	s.clients[clientId] = client

	// Return message containing the assigned id
	return &proto.ClientId{Id: clientId}, nil
}

func (s *ChitChatServer) LeaveSystem(ctx context.Context, in *proto.ClientInformation) (*proto.Empty, error) {
	lamport := s.updateLamportClockOnReceive(in.LamportClock)
	//Local event: remove client from list
	s.lamportClock++
	delete(s.clients, in.ClientId)

	s.BroadcastService(
		fmt.Sprintf("Client %d left Chit Chat at logical time %d", in.ClientId, lamport))
	return &proto.Empty{}, nil
}

func (s *ChitChatServer) JoinSystem(in *proto.ClientInformation, stream proto.ChitChat_JoinSystemServer) error { //todo, is this broadcast only for when joining?
	client, ok := s.clients[in.ClientId]
	if !ok {
		log.Fatalf("client not found %v", in.ClientId)
	}
	client.Stream = stream
	//start goRoutine
	go s.start_client(client)

	//join announcement
	lamport := s.updateLamportClockOnReceive(in.LamportClock) // checks max(S,0) + 1
	s.BroadcastService(
		fmt.Sprintf("Client %d joined Chit Chat at logical time %d", client.ClientId, lamport))

	// Keep the stream open until the client disconnects - And then the stream is closed
	<-stream.Context().Done()
	return nil
}

// loops over clients in server,pushes message to clients broadcast channels and then prints 'sent to..'
func (s *ChitChatServer) BroadcastService(broadcastMessage string) {
	//local event: send broadcast to client go routines that sends to streams
	s.lamportClock++
	for ClientId, client := range s.clients {
		client.BroadcastChannel <- &proto.BroadcastMessage{MessageContent: broadcastMessage, LamportClock: s.lamportClock}
		log.Printf("[Server][Broadcast] to %d: %s", ClientId, broadcastMessage)
	}
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
		// The go routine returns when the client leaves the system
		case <-c.Stream.Context().Done():
			return
		}
	}
}

func main() {
	server := &ChitChatServer{clients: make(map[int32]*Client), lamportClock: 0, clientNextId: 1}
	//creation of clients here
	server.start_server()
}

func (s *ChitChatServer) start_server() {
	//local event: server start
	s.lamportClock++
	//setup logging
	s.setupLogging()
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("failed to close log file: %v", err)
		}
	}(s.logFile) //closes the log file when the main function ends

	grpcServer := grpc.NewServer()              //creates new gRPC server instance
	proto.RegisterChitChatServer(grpcServer, s) //registers the server implementation with gRPC

	listener, err := net.Listen("tcp", ":5050") //listens on TCP port using net.Listen
	if err != nil {
		log.Fatalf("Did not work")
	}

	if err != nil {
		log.Fatalf("Did not work")
	}

	log.Println("[Server] ChitChat server started, listening on :5050")
	//go function for stopping server
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if scanner.Text() == "stop" {
				grpcServer.GracefulStop()
				return
			}
		}
	}()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	err = grpcServer.Serve(listener)
	//local event: server stop
	s.lamportClock++
	log.Println("[Server] ChitChat server stopped")
}

// sets up logging both into a file and the terminal - same as in client
func (s *ChitChatServer) setupLogging() {
	//creates the file (or overwrites it if it already exists)
	logFile, err := os.OpenFile("logFile.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}

	s.logFile = logFile //saves the logfile to the client struct

	log.SetOutput(io.MultiWriter(os.Stdout, logFile)) //sets the output to both the terminal and the file
}

// utility method that on message receive checks max lamport and updates local
func (s *ChitChatServer) updateLamportClockOnReceive(remoteLamport int32) int32 {
	if remoteLamport > s.lamportClock {
		s.lamportClock = remoteLamport
	}
	s.lamportClock++
	return s.lamportClock
}
