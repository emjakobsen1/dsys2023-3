###DSYS3: Chat Service

The chat service is a distributed system that allows clients to chat with each other.

In a terminal, start the server from the /server repository with with 

```sh
    go run server.go
```



For the clients, start them in seperate terminals from the /client repository with 

```sh
    go run client.go -name 0
```

```sh
    go run client.go -name 1
```

```sh
    go run client.go -name 2
```



The server can handle maximum 3 clients.

Now clients can publish messages by typing in the terminal e.g:

```sh
    Hey!
```

And the server will reply the broadcasted message to all the clients (the publisher included). The client prints its local timestamp along with recieved message:

```sh
    -> T [0 4 0]: Participant 1 publishes Hey!
```


Whenever a client wants to leave the chat, the client can type /leave in its terminal. 

The client writes all their received messages to the "logs/output.txt", and the server will write RPC's to the "logs/service_log.txt". Be aware that the contents of "output.txt" has to be manually deleted each time the system is started or the system will just append to it, however this is not the case for "service_log.txt" which do this automatically. 
