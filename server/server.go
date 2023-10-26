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
	gRPC.UnimplementedChatServiceServer // You need this line if you have a server
	clients                             map[gRPC.ChatService_MessageServer]bool
	mutex                               sync.Mutex // used to lock the server to avoid race conditions.
}

func main() {
	fmt.Println(".:server is starting:.")
	launchServer()
}

func launchServer() {
	fmt.Printf("Server: Attempts to create listener \n")
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:5400"))
	if err != nil {
		fmt.Printf("Server: Failed to listen on port 5400")
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	server := &Server{
		clients: make(map[gRPC.ChatService_MessageServer]bool),
	}

	gRPC.RegisterChatServiceServer(grpcServer, server)

	fmt.Printf("Server: Listening at %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) Message(msgStream gRPC.ChatService_MessageServer) error {
	// Add the client to the map
	s.mutex.Lock()
	s.clients[msgStream] = true
	s.mutex.Unlock()

	// Ensure the client is removed when the function exits
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
		log.Printf("Participant %s: %s", msg.ClientName, msg.Message)
		// Broadcast to all clients
		msgToClient := msg.Message
		s.mutex.Lock()
		for client := range s.clients {
			if err := client.Send(&gRPC.Farewell{Message: msgToClient}); err != nil {
				log.Printf("Error sending to client: %v", err)
			}
		}
		s.mutex.Unlock()
	}

	return nil
}
