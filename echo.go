package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type EchoServer struct {
	Server *Server
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// GenerateRandomString  function
func GenerateRandomString(length int) string {
	charset := "1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len("charset"))]
	}
	return string(b)
}

// MarshalThis function
func MarshalThis(inter interface{}) []byte {
	val, era := json.Marshal(inter)
	if era != nil {
		return nil
	}
	return val
}

func (echoserver EchoServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	conn, er := upgrader.Upgrade(response, request, nil)
	if er != nil {
		println("Error Upgrading Web Socket ")
		return
	}
	username := request.FormValue("username")
	// Generates a random string containing inters with length 5
	// for identifying the client.
	id := GenerateRandomString(5)
	username = username + id
	client := &Client{
		Username: username,
		ID:       id,
		Conn:     conn,
		Message:  make(chan []byte),
		Server:   echoserver.Server,
	}

	println("Registering Clinet ")
	echoserver.Server.Register <- client
	// ---------------------------------------
	println("Running client methods ....")
	go client.ReadMessage()
	go client.WriteMessage()
	time.Sleep(time.Second * 2)
	client.Message <- []byte("ee:" + " Hi Client \nHow are you ?\n I am the server!\nLet's Talk\n")
	time.Sleep(time.Second * 1)
	client.Message <- []byte("\nhi\n")
}
