package main

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	EndToEndClientMessage      = iota + 1
	EndToEndServerReply        // This status code represents the server reply(echo) to end to end message of the client.
	BroadcastMessageType       // This status code represents the client message for a group.
	BroadcastTypingMessage     // This status code represents the Typing Message including the typer ID
	BroadcastStopTypingMessage // This status code represents the Stop Typing Message which also holds the ID of the typing user
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

// XChangeMessage is a message which represents all the messages to be transffered through the web socket channel
// and each message's type and other information should be distinguished by their Status Number
// For all messages the type code is an integer and should represent a message type
// such as :  ClientsEchoMessage  , ServerEchoReplyMessage , Typing Message  ,
// Seen Message   , GroupClientMessage  , EditMessage
// representing those Type codes in integer is needed to save bandwidth cost.
type XChangeMessage struct {
	Type     int         `json:"type"`
	Body     interface{} `json:"body"`
	SenderID string      `json:"-"`
}

// BroadcastMessage  this struct represents the broadcast message body
// When the Type (Status ) code of XChangeMessage is BroadcastMessageType==3 ,
//  then the body will be an instance of this struct.
type BroadcastMessage struct {
	ID       string    `bson:"_id"  	  json:"id"`
	Username string    `bson:"username"  json:"username"`
	Message  string    `bson:"msg"  	  json:"msg"`
	Time     time.Time `bson:"time"  	  json:"time"`
	No       int       `bson:"no"  	  json:"no"`
}

// BroadcastTyping  .. this struct holds information such as Typing or stop typing message.
// The Status Message will define  the body and if the type (status) code is typing message ,then the  body will be an instance
// of this struct.
/*
	Type Codes :
	BroadcastTypingMessage=4
	BroadcastStopTypingMessage=5
*/
type BroadcastTyping struct {
	Username string `json:"username,omitempty"`
	ID       string `json:"id"`
}

// OutMessage is a class representing all the data to be sent out to the client through the connections
// NOTE : All client connections should have a common Format to be accepted by the client connection.
// In this format the Out Message Struct haves a list of parameters that are to be used by the client connection.
// In this demo i will only have a body only ... but for broader applications it can be extended...
// For example if just logged in to your telegram account using different devices... then telegram takes a look at your devices
// and send the message telling other device using this account has logged in
//
//  So : for such features you may need to have a filter telling for which device should the message be send.
//
type OutMessage struct {
	// MACAddress string --- This may tell which device is this message having as a target
	// ForAll bool --- this may tell that should the message be sent to all the devices of this account or specified target.
	Body []byte `json:"body"`
}

// This represents the client echo message that is to be
type ClientEchoMessage struct {
	Message string `json:"msg"`
	// Seen    bool   `json:"seen"`
}

// ServerEchoMessage represernts the servers echo message reply.
// I have separated those two messages because we may have
// some thing different to be included in the two message bodies.
type ServerEchoMessage struct {
	ClientID string `json:"client_id"` // client id represents id of the client.
	// Seen     bool   `json:"seen"`      //`
	Message string `json:"msg"`
}

// Client a class representing application clients.
type Client struct {
	Username string
	ID       string
	// Conn     *websocket.Conn
	// Message  chan *OutMessage
	Server *Server
	// map of string to Device
	// the key represents the newly generated ID which is same as that of IP String in the device instance.
	Devices map[string]*Device
	// BroadcastHandler  hold the Broadcast Chat  handler list
	BroadcastHandler *BroadcastChat
}

// Device object represents a connected device of a user account.
// Let's Say Samuael Connected to this server using a single account and
// different devices ..... the account will be represented by the by the client Object
//  where as the connected devices of the client will be represented by the Device struct
type Device struct {
	// Uniquely Identifying a device  ..
	// the IP is a symbol for IP address , but we can use other signatures to differentiate each device uniquely
	// For example it may be MAC ADDRESS , IP Address , or Randomly Generated Address to Uniquly Identify the Device.
	// a dynamically generated strin to represent each device.
	IP string
	// Conn : web socket connection created by the device
	Conn *websocket.Conn
	// This is a channel through which Messages to be sent for end device will be sent through.
	Message chan *OutMessage
}

// UniqueAddress a struct to hold the ID and IP of a device.
type UniqueAddress struct {
	// ID the id of the client instance.
	ID string //
	// IP the IP Address of the device (  in our case a randmly generated address. )
	IP string
}
