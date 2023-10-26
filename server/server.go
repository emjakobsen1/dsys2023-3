package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	gRPC "github.com/emjakobsen1/dsys2023-3/proto"
	"google.golang.org/grpc"
)

type Server struct {
	gRPC.UnimplementedChatServiceServer
	clients map[gRPC.ChatService_MessageServer]bool
	mutex   sync.Mutex
}

func main() {
	log.SetFlags(0)
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
			log.Printf("Client %s publishes: %s", msg.ClientName, msg.Message)
		case gRPC.MessageType_JOIN:
			log.Printf("Client %s joins", msg.ClientName)
		case gRPC.MessageType_LEAVE:
			log.Printf("Client %s leaves", msg.ClientName)
		}

		// Broadcast to all clients
		s.mutex.Lock()
		for client := range s.clients {
			if err := client.Send(&gRPC.Reply{Message: msg.Message, ClientName: msg.ClientName, Type: msg.Type}); err != nil {
				log.Printf("Error sending to client: %v", err)
			}
		}
		s.mutex.Unlock()
	}

	return nil
}
