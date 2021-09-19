package main

import (
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (client *Client) ReadMessage() {
	// ticker := time.NewTicker(time.Second * 3)
	defer func() {
		client.Conn.Close()
		client.Server.Remove <- client
		message := recover()
		if message != nil {
			println(string(MarshalThis(message)))
		}
	}()
	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetCloseHandler(func(code int, text string) error {
		// this is close handler to be called whenever a close message from the client is sent.
		// send a unregister me message to the main service so that the Unregister function in
		// Main service will close teh closed connection and let the rest.
		return nil
	})
	// client.Conns[key].Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error { /*client.Conns[key].Conn.SetReadDeadline(time.Now().Add(pongWait)); */ return nil })
	for {
		message := []byte{}
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseInternalServerErr, websocket.CloseMandatoryExtension, websocket.CloseMessage, websocket.CloseProtocolError, websocket.CloseUnsupportedData) {
				println("Internal Server Error ")
				return
			}
			break
		}
		if message == nil {
			continue
		}
		val := string(message)
		var mess Message
		if strings.HasPrefix(val, "e:") {
			val = strings.TrimPrefix(val, "e:")
			mess = Message{
				From:     client.ID,
				Type:     EndToEnd,
				Message:  val,
				Username: client.Username,
			}
		} else if strings.HasPrefix(val, "g:") {
			val = strings.TrimPrefix(val, "g:")
			mess = Message{
				From:     client.ID,
				Type:     Group,
				Message:  val,
				Username: client.Username,
			}
		}
		if &mess != nil {
			client.Server.Message <- mess
		}
	}
}

func (client *Client) WriteMessage() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		// --
		client.Conn.Close()
		client.Server.Remove <- client
		message := recover()
		if message != nil {
			println(string(MarshalThis(message)))
		}
	}()
	for {
		select {
		case mess, ok := <-client.Message:
			{
				if !ok {
					client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				client.Conn.WriteMessage(websocket.TextMessage, append(mess, '\n'))
			}
		case <-ticker.C:
			{
				if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}
