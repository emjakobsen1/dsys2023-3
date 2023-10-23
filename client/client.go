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
var stream gRPC.ChatService_SayHiClient

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

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
	//start the biding

}

func establishStream() {
	var err error
	stream, err = server.SayHi(context.Background()) // This establishes the stream
	if err != nil {
		log.Fatalf("Failed to establish stream: %v", err)
	}
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

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewChatServiceClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
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

		// Use the globally established stream for sending messages
		if err := stream.Send(&gRPC.Greeting{ClientName: *clientsName, Message: input}); err != nil {
			log.Println("Failed to send message:", err)
			return
		}
		// farewell, err := stream.Recv()
		// if err != nil {
		// 	log.Println(err)
		// 	return
		// }
		// log.Println("server says: ", farewell)
	}
}
func listenForMessages(stream gRPC.ChatService_SayHiClient) {
	for {
		farewell, err := stream.Recv()
		if err == io.EOF {
			log.Println("Server closed the connection")
			return
		}
		if err != nil {
			log.Println("Error receiving from server:", err)
			return
		}
		log.Println("server says:", farewell)
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
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
