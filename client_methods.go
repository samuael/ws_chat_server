package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// ReadMessage is a goroutine which runs all the time until
// the socket connection between the client and the server is closed.
func (client *Client) ReadMessage(IP string) {
	defer func() {
		if client.Devices[IP].Conn != nil {
			client.Devices[IP].Conn.Close()
			client.Server.Remove <- UniqueAddress{IP: IP, ID: client.ID}
		}
		message := recover()
		if message != nil {
			println("Write ", IP, string(MarshalThis(message)))
		}
	}()
	// This is the packages message ...
	// SetReadLimit sets the maximum size in bytes for a message read from the peer.
	// If a message exceeds the limit, the connection sends a close message to the peer and
	// returns ErrReadLimit to the application.
	//
	// In Our Case i have set the length of the Message size to be a big number.but , if you really want to set it to limitted size
	// you can do so by using the 'maxMessageSize' variable.
	client.Devices[IP].Conn.SetReadLimit(maxMessageSize)
	client.Devices[IP].Conn.SetCloseHandler(
		func(code int, text string) error {
			// This is close handler to be called whenever a close message from the client is sent.
			// send a unregister me message to the main service so that the Unregister function in
			// Main service will close teh closed connection and let the rest.
			return nil
		})
	/*	Using this method below named SetReadDeadLine you  can set the maximum time to wait before the next message arrives or sent.
		the diration is set using the pong wait variable.
	*/
	// client.Devices[IP].Conn.SetReadDeadline(time.Now().Add(pongWait))

	client.Devices[IP].Conn.SetPongHandler(func(string) error { /*client.Conns[key].Conn.SetReadDeadline(time.Now().Add(pongWait)); */ return nil })
	for {
		// This slice of byte is used to get a message from the socket connection.
		message := &XChangeMessage{}
		err := client.Devices[IP].Conn.ReadJSON(message)
		if err != nil {
			log.Println("ERROR : ", err.Error())
			if websocket.IsUnexpectedCloseError(err, 1006, websocket.CloseInternalServerErr, websocket.CloseMessage) {
				return
			}
			continue
		}
		if message == nil || message.Body == nil {
			continue
		}
		message.SenderID = client.ID

		if message.Type == EndToEndClientMessage {
			serverMessage := &XChangeMessage{}
			serverMessage.SenderID = client.ID
			println((message.Body))
			if mess := func() (msg string) {
				msg = ""
				defer func() {
					// Recover incase an exception happened ...
				}()
				return message.Body.(map[string]interface{})["msg"].(string)
			}(); mess != "" {
				body := ClientEchoMessage{Message: mess}
				message.Body = &body
				message.Type = EndToEndClientMessage

				// Here I am Sending the client with the server echo message .
				serverMessage.Body = &ServerEchoMessage{ClientID: client.ID, Message: "\nYou Said '" + body.Message + "'"}
				serverMessage.Type = EndToEndServerReply
				client.Server.Message <- message
				time.Sleep(time.Second * 1)
				// println("Before sending ... ")
				client.Server.Message <- serverMessage
			}
		} else if message.Type == BroadcastMessageType {
			// When the message is a broadcast message ...
			// try getting the message body
			if msg := func() (msg string) {
				// Casting the message from client input is needed to be implemented this way because the clients may send un invalid for message and if error happens
				// in this go ruting due to that message the go routine may return with an excption.
				// threfore , checking errors this way is the best way to prevent such casting errors.
				msg = ""
				defer func() {
					// recover incase an exception happened.
					recover()
				}()
				return message.Body.(map[string]interface{})["msg"].(string)
			}(); msg != "" {
				client.BroadcastHandler.LastMessageNumber++
				body := &BroadcastMessage{Username: client.Username, ID: client.ID, Time: time.Now(), No: client.BroadcastHandler.LastMessageNumber}
				client.BroadcastHandler.Messages = append(client.BroadcastHandler.Messages, *body)
				body.Message = msg
				client.BroadcastHandler.Users[client.ID] = client
				message.Body = body
				/*
					Before Sending the message we need to add the broadcast message to the list of broadcast messages in the BroadcastChat Instance.
				*/
				client.Server.Message <- message
				time.Sleep(time.Millisecond * 30)
			}
		} else if message.Type == BroadcastStopTypingMessage || message.Type == BroadcastTypingMessage {
			// Check whether the client is active of not.
			if client := Clients[client.ID]; client != nil {
				// Instantiating the Broadcast Typign message and pass it to the Xcahnge message..
				body := &BroadcastTyping{Username: client.Username, ID: client.ID}
				message.Body = body
				client.Server.Message <- message
				time.Sleep(time.Millisecond * 30)
			}
		}
	}
}

func (client *Client) WriteMessage(IP string) {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		if client.Devices[IP] != nil && client.Devices[IP].Conn != nil {
			client.Devices[IP].Conn.Close()
		}
		client.Server.Remove <- UniqueAddress{IP: IP, ID: client.ID}
		message := recover()
		if message != nil {
			println("Write ", IP, string(MarshalThis(message)))
		}
	}()
	// Continuously running loop for checking the connection and a
	// ticker ticking in each 10 seconds to check for cheching whether the connection is live or not.
	for {
		select {
		// message channel
		case mess, ok := <-client.Devices[IP].Message:
			{
				if !ok {
					client.Devices[IP].Conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				client.Devices[IP].Conn.WriteMessage(websocket.TextMessage, mess.Body)
			}
		case <-ticker.C:
			{
				if err := client.Devices[IP].Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}
