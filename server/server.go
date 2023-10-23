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
	// name                             string // Not required but useful if you want to name your server
	// port                             string // Not required but useful if your server needs to know what port it's listening to
	clients map[gRPC.ChatService_SayHiServer]bool
	// incrementValue int64      // value that clients can increment.
	mutex sync.Mutex // used to lock the server to avoid race conditions.
}

func main() {

	fmt.Println(".:server is starting:.")
	launchServer()
}

func launchServer() {
	fmt.Printf("Server: Attempts to create listener \n")

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:5400"))
	if err != nil {
		fmt.Printf("Server: Failed to listen on port 5400") //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		clients: make(map[gRPC.ChatService_SayHiServer]bool),
	}

	gRPC.RegisterChatServiceServer(grpcServer, server) //Registers the server to the gRPC server.

	fmt.Printf("Server: Listening at %v\n", list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func (s *Server) SayHi(msgStream gRPC.ChatService_SayHiServer) error {
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
		log.Printf("Received message from %s: %s", msg.ClientName, msg.Message)
		log.Printf("Clients: %d", len(s.clients))
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
