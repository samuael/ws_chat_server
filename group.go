package main

import (
	"net/http"
)

// BroadcastChat ...
// A centrac repo for holding the list of messages that are broadcasted.
type BroadcastChat struct {
	// Messages holds a list of messages that are passed transffered between the clients.
	Messages []BroadcastMessage
	// Users holds a list of Uses who have sent at least 1 message to the group.
	Users map[string]*Client
	// LastMessageNumber
	LastMessageNumber int
}

// GetListOfMessages this is a REST handler function for getting a list of messages broadcasted
// this response starts at the recent message and if  you increate the limit then you will foind the oldest messages.
func (bchat *BroadcastChat) GetListOfMessages(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	resp := struct {
		// LastMessageNumber int `json:"last_message_number"`
		Messages []BroadcastMessage
	}{
		// LastMessageNumber: bchat.LastMessageNumber,
		Messages: bchat.Messages,
	}
	response.Write(MarshalThis(resp))
}

// GetUsers clients that had sent at least one message.
func (bchat *BroadcastChat) GetUsers(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.Write(MarshalThis(bchat.Users))
}
