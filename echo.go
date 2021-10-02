package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// seededRand ...
var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// GenerateRandomString  function
func GenerateRandomString(length int) string {
	charset := "1234567890abcdefhijklmnopqrstuvwxyz"
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

// ServeHTTP a handler function to create a socket connection with the client and to create a Client instance
// which holds the socket client connection  instance and thre related information related to the client it may be either a username an ID
// or IP address and related information.
func (server Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	conn, er := upgrader.Upgrade(response, request, nil)
	if er != nil {
		println("Error Upgrading Web Socket ")
		return
	}
	username := request.FormValue("username")
	// Generates a random string containing inters with length 5
	// for identifying the client.
	id := request.FormValue("id")
	if er != nil {
		id = GenerateRandomString(5)
	}
	if username == "" {
		username = "Unknown"
	}
	// This may be the IP ADDRESS , MAC ADDRESS , or other Device Uniquely identifying number
	// in this case i have used randomly generated String.
	ip := GenerateRandomString(5)
	device := &Device{
		Conn:    conn,
		IP:      ip,
		Message: make(chan *OutMessage)}
	client := Client{
		Username:         username,
		ID:               id,
		BroadcastHandler: server.BroadcastChat,
		// Conn:     conn,
		// Message:  make(chan *OutMessage),
		Devices: map[string]*Device{ip: device},
		Server:  &server,
	}
	server.Register <- client
	go client.ReadMessage(ip)
	go client.WriteMessage(ip)

	time.Sleep(time.Millisecond * 200)
	// serverMessage
	serverMessage := &XChangeMessage{}
	xchangebody := &ServerEchoMessage{Message: "Hi " + strings.Title(username) + "!\nWelcome To echo chat!\n"}
	// Here I am Sending the client "id" with the first message of the server so that the client will be aware of his/her ID  even if the Client ID
	// not created by the client ( Because  , In this case the client "id" can be  passed by the client in the request parameter ).
	xchangebody.ClientID = client.ID
	serverMessage.Type = EndToEndServerReply
	serverMessage.SenderID = client.ID
	serverMessage.Body = xchangebody

	time.Sleep(time.Millisecond * 200)
	// This message is to be seen by the newly connected device of the client
	// but, other devices of the connected devices will not receive this message.
	device.Message <- &OutMessage{
		Body: MarshalThis(serverMessage),
	}
}
