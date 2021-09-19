package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var Clients []*Client

func main() {

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}
	println(port)

	server := Server{
		Register: make(chan *Client),
		Remove:   make(chan *Client),
		Message:  make(chan Message),
	}
	go server.Handle()
	wshandler := EchoServer{Server: &server}
	http.Handle("/ws/", wshandler)
	http.ListenAndServe(":"+port, nil)
}

// Server representing the single point fo failure for handling client registration
// deletion and closing of their web socket connection and message forwarding.
type Server struct {
	// This server will be the only one which will have an access to Clients clice.
	Register chan *Client // This method of Server client handling is used for dead lock prevention.
	Remove   chan *Client
	Message  chan Message
}

func (server *Server) Handle() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			{
				println("Ticking ." + strconv.Itoa(len(Clients)))
			}
		case val := <-server.Message:
			{
				server.SendMessage(val)
			}
		case client := <-server.Register:
			{
				Clients = append(Clients, client)
			}
		case client := <-server.Remove:
			{
				for i, cl := range Clients {
					if cl.ID == client.ID {
						if i == 0 {
							Clients = Clients[0:]
						} else {
							Clients = append(Clients[0:i], Clients[i+1:]...)
						}
						break
					}
				}
			}
		}
	}
}

// BroadcastMessage to braodcast the message to the group.
func (server Server) SendMessage(message Message) {
	for _, client := range Clients {
		if message.Type == EndToEnd && client.ID == message.From || message.Type == Group {
			mess := Message{
				From: message.From,
				Type: message.Type,
				Message: func() string {
					if message.Type == EndToEnd {
						if client.ID == message.From {
							return "ey:" + message.Message
						} else {
							return "ee:" + message.Message
						}
					} else {
						if client.ID == message.From {
							return "gy:" + "You:" + message.Message
						} else {
							return "gm:" + message.Username + message.Message
						}
					}
				}(),
			}

			val := mess.Message
			if message.Type == EndToEnd {
				client.Message <- []byte(val)
			}
			time.Sleep(time.Second * 1)
			if message.Type == EndToEnd {
				client.Message <- []byte("ee:" + " This was your Message \n<<" + strings.TrimPrefix(mess.Message, "ey:") + ">>")
			}
			time.Sleep(time.Second * 1)
			client.Message <- []byte("\nhi\n")
		}
	}
}
