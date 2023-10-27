package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	gRPC "github.com/emjakobsen1/dsys2023-3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Server struct {
	gRPC.UnimplementedChatServiceServer
	clients map[gRPC.ChatService_MessageServer]bool
	mutex   sync.Mutex
}

func main() {
	log.SetFlags(log.LstdFlags)
	file, err := os.OpenFile("../log/service_log.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()
	log.SetOutput(file)

	launchServer()
}

func launchServer() {
	list, err := net.Listen("tcp", "localhost:5400")
	if err != nil {
		fmt.Printf("Server: Failed to listen on port 5400")
		return
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	server := &Server{
		clients: make(map[gRPC.ChatService_MessageServer]bool),
	}

	gRPC.RegisterChatServiceServer(grpcServer, server)

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Message(msgStream gRPC.ChatService_MessageServer) error {
	if peer, ok := peer.FromContext(msgStream.Context()); ok {
		log.Println("Message service call by ", peer.Addr.String())
	}
	s.mutex.Lock()
	s.clients[msgStream] = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.clients, msgStream)
		s.mutex.Unlock()
	}()

	for {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len([]rune(msg.Message)) > 128 {
			continue
		}

		switch msg.Type {
		case gRPC.MessageType_PUBLISH:
			fmt.Printf("T: %v Client %d publishes: %s \n", msg.Clock, msg.ClientName, msg.Message)
		case gRPC.MessageType_JOIN:
			fmt.Printf("T: %v Client %d joins \n", msg.Clock, msg.ClientName)
		case gRPC.MessageType_LEAVE:
			fmt.Printf("T: %v Client %d leaves \n", msg.Clock, msg.ClientName)
		}

		// Broadcast to all clients
		s.mutex.Lock()
		for client := range s.clients {
			if err := client.Send(&gRPC.Reply{Message: msg.Message, ClientName: msg.ClientName, Type: msg.Type, Clock: msg.Clock}); err != nil {
				log.Printf("Error sending to client: %v", err)
			}
		}
		s.mutex.Unlock()
	}

	return nil
}
