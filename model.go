package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	EndToEnd = iota
	Group
)
const (
	// writeWait Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// pongWait Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// pmessagegPeriod Send pmessagegs to peer with this period. Must be less than pongWait.
	pmessagegPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 99999999999
)

/* Message
\ Message Formats for Sample /
This is for both end to end and group messages
ee:"message"  	-- for end to end (echo) message
ey:"message"  	-- for end to end (echo) message when source is YOU.
gm:"message" 	-- for group message.
gy:"message"    -- for group message when the sender is YOU.
*/
type Message struct {
	From     string `json:"From"`
	Type     int    `json:"type"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

// Client a class representing application client.
type Client struct {
	Username string
	ID       string
	Conn     *websocket.Conn
	Message  chan []byte
	Server   *Server
	
}
