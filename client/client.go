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

var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.ChatServiceClient
var ServerConn *grpc.ClientConn
var stream gRPC.ChatService_MessageClient

func main() {
	flag.Parse()
	log.SetFlags(0)

	ConnectToServer()
	defer ServerConn.Close()
	establishStream()
	go listenForMessages(stream)
	parseInput()

}

func ConnectToServer() {
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
	if err := stream.Send(&gRPC.Request{
		ClientName: *clientsName,
		Message:    "",
		Type:       gRPC.MessageType_JOIN,
	}); err != nil {
		log.Println("Failed to send join message:", err)
		return
	}
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)
		if len([]rune(input)) > 128 {
			continue
		}

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		if input == "/leave" {
			if err := stream.Send(&gRPC.Request{
				ClientName: *clientsName,
				Message:    "",
				Type:       gRPC.MessageType_LEAVE,
			}); err != nil {
				log.Println("Failed to send leaving message:", err)
			}
			return
		}

		if err := stream.Send(&gRPC.Request{
			ClientName: *clientsName,
			Message:    input,
			Type:       gRPC.MessageType_PUBLISH,
		}); err != nil {
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

		switch message.Type {
		case gRPC.MessageType_PUBLISH:
			log.Printf("Client %s publishes %s", message.ClientName, message.Message)
		case gRPC.MessageType_JOIN:
			log.Printf("Client %s joins", message.ClientName)
		case gRPC.MessageType_LEAVE:
			log.Printf("Client %s leaves", message.ClientName)
		}
	}
}

func conReady(s gRPC.ChatServiceClient) bool {
	return ServerConn.GetState().String() == "READY"
}
