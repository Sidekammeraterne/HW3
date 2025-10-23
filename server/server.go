package main

import (
	proto "ITUServer/grpc" //make connection
	"context"              //make connection - the context of the connection
	"google.golang.org/grpc"
	"log" //logs - used to keep track of messages
	"net" //make connection to net
)

type ChitChatServer struct {
	//part of the proto - we are creating an Unimplemented server
	proto.UnimplementedChitChatServer
	//todo: define what we want the struct to contain, if it should contain anything
}

// Method only be activated if called through proto
// (s*ITU_databaseServer) the reciever - it points to the struct ITU_databaseServer that we want to use (it is the same as parsing the student as a parameter)
// GetStudents(ctx context.Context, in *proto.Empty) function name and parameters of the function
// (*proto.Students, error) is the return type
func (s *ChitChatServer) publishMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) { //if we want something with an empty in it
	return &proto.Empty{}, nil //returns the pointer to the student in memory - what we are encourages to do [should be able to see this from the function declaration (?)]
}

func main() {
	//creation of clients
	server := &ChitChatServer{}
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
