package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TimEngleSF/remote-hue-server/internal/service"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader with custom settings
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		return true
	},
}

// GroupsStateMessage represents the structure of the group state messages.
type GroupsStateMessage struct {
	Type string `json:"type"`
	Data struct {
		Groups service.Groups `json:"groups"`
	} `json:"data"`
}

func (gs *GroupsStateMessage) Unmarshal(data []byte) error {
	return json.Unmarshal(data, gs)
}

// handleWSConnections handles WebSocket connections and processes incoming messages.
func (app *application) handleWSConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	defer conn.Close()

	// Set the WebSocket connection for the application.
	app.wsConnection = conn
	log.Println("Client connected:", r.RemoteAddr)

	// Loop to read and process incoming messages from the client.
	for {
		fmt.Println("Waiting for message from client")
		var msg JSONMessage
		// Read a JSON message from the client.
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message from client:", err)
			break
		}

		// Dispatch the message based on its type.
		app.dispatchMessage(msg)
	}
	log.Println("Client disconnected")
}

// dispatchMessage routes the message to the appropriate handler or channel.
func (app *application) dispatchMessage(msg JSONMessage) {
	app.responseMu.Lock()
	defer app.responseMu.Unlock()

	if ch, exists := app.responseMap[msg.Type]; exists {
		ch <- msg
	} else {
		switch msg.Type {
		case "group_state":
			err := app.GroupStateMessageHandler(msg)
			if err != nil {
				app.logger.Error("Error handling group state message:", "error", err)
			}
		default:
			app.logger.Warn("unknown message type:", "type", msg.Type)
		}
	}
}

// GroupStateMessageHandler processes group state messages and updates the application state.
func (app *application) GroupStateMessageHandler(msg JSONMessage) error {
	// Convert the generic Data field to JSON.
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var msgData GroupsStateMessage
	err = msgData.Unmarshal(data)
	if err != nil {
		return err
	}

	groups := service.Groups(msgData.Data.Groups)
	// Update the application state with the new groups data.
	app.groupsState = &groups
	app.groupNames = []string{}
	for _, group := range groups {
		app.groupNames = append(app.groupNames, group.Name)
	}
	return nil
}

func (app *application) SetGroupsStateField() error {
	msg := JSONMessage{
		Type: "status",
		Data: nil,
	}

	responseChan := make(chan JSONMessage)

	app.responseMu.Lock()
	app.responseMap["group_state"] = responseChan
	app.responseMu.Unlock()

	err := app.wsConnection.WriteJSON(msg)
	if err != nil {
		return err
	}

	select {
	case response := <-responseChan:
		// Handle the response
		var groupsMessage GroupsStateMessage
		data, err := json.Marshal(response)
		if err != nil {
			return err
		}

		err = groupsMessage.Unmarshal(data)
		if err != nil {
			return err
		}

		app.groupsState = &groupsMessage.Data.Groups

	case <-time.After(5 * time.Second):
		// Timeout after 5 seconds
		app.logger.Error("timeout waiting for group_state response")
		return fmt.Errorf("timeout waiting for group_state response")

	}
	app.responseMu.Lock()
	delete(app.responseMap, "group_state")
	app.responseMu.Unlock()
	return nil
}
