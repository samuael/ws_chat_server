package main

import (
	"log"
	"net/http"
	"os"
)

var Clients map[string]*Client

func init() {
	Clients = map[string]*Client{}
}
func main() {
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8070"
		log.Printf("defaulting to PORT %s", port)
	}
	broadcast := &BroadcastChat{
		LastMessageNumber: 1,
		Messages:          []BroadcastMessage{},
		Users:             map[string]*Client{},
	}
	server := Server{
		Register:      make(chan *Client),
		Remove:        make(chan UniqueAddress),
		Message:       make(chan *XChangeMessage),
		BroadcastChat: broadcast,
	}
	// Start a the main service handler function.
	go server.Handle()
	wshandler := &server
	http.Handle("/ws/", wshandler)
	http.HandleFunc("/api/messages/", broadcast.GetListOfMessages)
	http.HandleFunc("/api/users/", broadcast.GetUsers)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Server representing the single point fo failure for handling client registration
// deletion and closing of their web socket connection and message forwarding.
type Server struct {
	// This server will be the only one which will have an access to Clients clice.
	Register      chan *Client // This method of Server client handling is used for dead lock prevention.
	Remove        chan UniqueAddress
	Message       chan *XChangeMessage
	BroadcastChat *BroadcastChat
}

// Handle .. the main goo routin for running cotinuously
//
func (server *Server) Handle() {
	// Use this ticker if you want to check any thing related this server and it connectins.
	// ticker := time.NewTicker(time.Second * 10)
	defer func() {
		// ticker.Stop()
		recover()
		// Recover from faults or closing of this thread by errors .
		// The only thing that has to stop this thread is KILL SIGNAL.
		go server.Handle()
	}()
	/*
		Handling all the data access (especially, any updates or non retrival requests) are always has to be handled through this
		function to prevent multiple a contension of other critical section problems.
		But , using such techniques makes the server a single point of failure . meaning ,if this function returns all the connections will sit IDLE
		and no message transaction will be held.
		So To Prevent this to happen you should Use a
		1. Distributed servers connected through RPC calls.
		2. Using Deadlock prevention techniques to prevent such Problems Lick locking.
		3. ...
	*/
	for {
		select {
		case val := <-server.Message:
			{
				server.SendMessage(val)
			}
		case client := <-server.Register:
			{
				server.RegisterClient(client)
			}
		case uniqueAddress := <-server.Remove:
			{
				server.UnRegisterClient(uniqueAddress.ID, uniqueAddress.IP)
			}
			/* case <-ticker.C: 	{
				log.Println(len(Clients))
			} */
		}
	}
}

// SendMessage : this function send a message depending on the message type.
// for example if the mesage type is group message then broadcast the message and if it is end to end it will send the message to target device.
func (server *Server) SendMessage(message *XChangeMessage) {
	defer func() {
		message := recover()
		if message != nil {
			log.Println(" HERE : ", message.(string))
		}
	}()
	for _, client := range Clients {
		if ((message.Type == EndToEndClientMessage || message.Type == EndToEndServerReply) && client.ID == message.SenderID) ||
			message.Type == BroadcastMessageType ||
			message.Type == BroadcastStopTypingMessage ||
			message.Type == BroadcastTypingMessage {
			// Message to be sent out.
			outMessage := &OutMessage{
				Body: MarshalThis(message),
			}
			// Loop over the devices and send the message for each devices.
			for _, device := range client.Devices {
				// This device instance is a pointer to the device in the list of the Devices in the client.
				device.Message <- outMessage
			}
			if message.Type == EndToEndClientMessage {
				break
			}
		}
	}
}

// RegisterClient : this method adds the newly connected device to the map of client in this application
// if device with same id is already registered, then this function adds the client's Device to the list of existing devices.
func (server *Server) RegisterClient(client *Client) {
	// check whether the client is available or not.
	// if so append the device in teh client devices list else use this newly generated client instance.
	if clnt := Clients[client.ID]; clnt != nil {
		// Loop Over each clients device and add it to the priviously created Client instance Devices List.
		for ip, device := range client.Devices {
			// Add the new device to the list of devices attached with this client object.
			clnt.Devices[ip] = device
		}
	} else {
		// Just use the newly Created Client Instance.
		Clients[client.ID] = client
	}

}

// UnRegisterClient is a function to remove a client from the cached list of clients.
// This funciton takes an argument of client's ID and clients IP address.
// ID to identify the client object and IP to identify the Device of the client.
func (server *Server) UnRegisterClient(ID, IP string) {
	// When we want to unregister the client we need to pass this two parameters to the server inside the UniqueAddress instance with the Unregister
	// channel and the main server filters the client object with this id and a device with thsi id to delete the device if there are
	// a number of active devices connected using this id. but , if the number of connected devices is only one , then the client object will be deleted too.
	if client := Clients[ID]; client != nil && len(client.Devices) > 1 {
		// If the Length of the Clients is greater than 1 meaning there is other devices that are too connected with this client account
		// then , delete that specified client Device from the devices list.
		if client.Devices != nil && client.Devices[IP] != nil && client.Devices[IP].Conn != nil {
			client.Devices[IP].Conn.Close()
		}
		delete(client.Devices, IP)
	} else if client != nil && len(client.Devices) <= 1 {
		// since teh length of device is 1 or less unregistering that device is also Unregistering the clinet instance.
		if len(client.Devices) == 0 && (len(client.Devices) == 1 && client.Devices[IP] != nil) {
			delete(Clients, ID)
			delete(server.BroadcastChat.Users, ID)
		}
	}
}
