package main

import (
	proto "ITUServer/grpc"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	// creates a connection related to a client who can speak to the specified port, grpc client can ask for a service
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials())) //security, don't think much about it."Boilerpalte<"
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	client := proto.NewChitChatClient(conn) //creates a new client and opens for the connection

	//students, err := client.GetStudents(context.Background(), &proto.Empty{}) //creates an array or slice with the repeated strings of stufents
	//if err != nil {
	//log.Fatalf("we did not recieve anything or failed to send %v", err)
	//	}

	//courses, err := client.GetCourses(context.Background(), &proto.Empty{})
	//if err != nil {
	//	log.Fatalf("we did not recieve anything or failed to send %v", err)
	//}

	//prints all the students
	//for _, student := range students.Students {
	//	log.Println(" - " + student)
	//}

	//prints all the courses
	//for _, course := range courses.Courses {
	//	log.Println(" - " + course)
	//}
}
