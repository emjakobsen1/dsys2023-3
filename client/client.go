package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	gRPC "github.com/emjakobsen1/dsys2023-3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChatServiceClient //the server
var ServerConn *grpc.ClientConn   //the server connection
var stream gRPC.ChatService_MessageClient

func main() {
	//parse flag/arguments
	flag.Parse()
	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()
	establishStream()
	go listenForMessages(stream) // Start listening immediately after the stream is established

	parseInput()

}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	server = gRPC.NewChatServiceClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func establishStream() {
	var err error
	stream, err = server.Message(context.Background()) // This establishes the stream
	if err != nil {
		log.Fatalf("Failed to establish stream: %v", err)
	}
	if err := stream.Send(&gRPC.Request{ClientName: *clientsName, Message: "joins."}); err != nil {
		log.Println("Failed to send message:", err)
		return
	}
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		if input == "/leave" {
			if err := stream.Send(&gRPC.Request{ClientName: *clientsName, Message: "leaves."}); err != nil {
				log.Println("Failed to send leaving message:", err)
			}
			return
		}

		// Use the globally established stream for sending messages
		if err := stream.Send(&gRPC.Request{ClientName: *clientsName, Message: input}); err != nil {
			log.Println("Failed to send message:", err)
			return
		}

	}
}
func listenForMessages(stream gRPC.ChatService_MessageClient) {
	for {
		message, err := stream.Recv()
		if err == io.EOF {
			log.Println("Server closed the connection")
			return
		}
		if err != nil {
			log.Println("Error receiving from server:", err)
			return
		}
		log.Printf("Client %s publishes %s", message.ClientName, message.Message)
	}
}

func conReady(s gRPC.ChatServiceClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
// func setLog() *os.File {
// 	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
// 	if err != nil {
// 		log.Fatalf("error opening file: %v", err)
// 	}
// 	log.SetOutput(f)
// 	return f
// }
