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

var clientsName = flag.Int("name", 0, "Senders ID")
var logger *log.Logger
var server gRPC.ChatServiceClient
var ServerConn *grpc.ClientConn
var stream gRPC.ChatService_MessageClient

var clock []int32

func main() {
	file, err := os.OpenFile("../log/output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	logger = log.New(file, "", 0)

	flag.Parse()
	log.SetFlags(0)
	clock = make([]int32, 3)

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
	conn, err := grpc.Dial(fmt.Sprintf(":%s", "5400"), opts...)
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

	clock[*clientsName]++
	if err := stream.Send(&gRPC.Request{
		ClientName: int32(*clientsName),
		Message:    "",
		Type:       gRPC.MessageType_JOIN,
		Clock:      clock,
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
			log.Printf("Client %d: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		if input == "/leave" {
			clock[*clientsName] = clock[*clientsName] + 1
			if err := stream.Send(&gRPC.Request{
				ClientName: int32(*clientsName),
				Message:    "",
				Type:       gRPC.MessageType_LEAVE,
				Clock:      clock,
			}); err != nil {
				log.Println("Failed to send leaving message:", err)
			}

			return
		}
		clock[*clientsName]++
		if err := stream.Send(&gRPC.Request{
			ClientName: int32(*clientsName),
			Message:    input,
			Type:       gRPC.MessageType_PUBLISH,
			Clock:      clock,
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
		if len(message.Clock) == 3 && len(clock) == 3 {
			for j := 0; j < len(clock); j++ {
				clock[j] = max(clock[j], message.Clock[j])
			}
			clock[*clientsName]++
		}
		switch message.Type {
		case gRPC.MessageType_PUBLISH:
			log.Printf("T %v: Participant %d publishes %s", clock, message.ClientName, message.Message)
			logger.Printf("Client: %d T: %v Participant %d publishes %s", *clientsName, clock, message.ClientName, message.Message)
		case gRPC.MessageType_JOIN:
			log.Printf("T: %v Participant %d joins", clock, message.ClientName)
			logger.Printf("Client %d T: %v Client %d joins", *clientsName, clock, message.ClientName)
		case gRPC.MessageType_LEAVE:
			log.Printf("T: %v Participant %d leaves", clock, message.ClientName)
			logger.Printf("Client %d T: %v Participant %d leaves", *clientsName, clock, message.ClientName)
		}

	}
}

func conReady(s gRPC.ChatServiceClient) bool {
	return ServerConn.GetState().String() == "READY"
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
