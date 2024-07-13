package main

import (
	"log"
	"net/http"

	"github.com/amimof/huego"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		return true
	},
}

type GroupsStateMessage struct {
	Type string        `json:"type"`
	Data []huego.Group `json:"data"`
}

func (app *application) handleWSConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	defer conn.Close()

	log.Println("Client connected:", r.RemoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to server")); err != nil {
		app.logger.Error(err.Error())
		return
	}

	for {
		var msg GroupsStateMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading message from client:", err)
			break
		}

		app.groupsState = &msg.Data
		log.Printf("Received message: %+v\n", msg)

		// Echo the message back to the client
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("Error writing message to client:", err)
			break
		}
	}
	log.Println("Client disconnected")
}

// func (app *application) handleWSConnections(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		app.logger.Error("Error upgrading websocket connection:", err)
// 		return
// 	}
// 	defer conn.Close()
//
// 	app.wsConnection = conn
// 	app.logger.Info("Client connected", "ip", r.RemoteAddr)
// 	if err := conn.WriteMessage(websocket.TextMessage, []byte("Connected to server")); err != nil {
// 		app.logger.Error("Error writing initial message to client:", err)
// 		return
// 	}
//
// 	for {
// 		// read message from client
// 		var msg Message
// 		err := conn.ReadJSON(&msg)
// 		if err != nil {
// 			app.logger.Error("Error reading message from client:", err)
// 			break
// 		}
// 		fmt.Println(msg)
//
// 		// Example: Echo the message back to the client
// 		if err := conn.WriteJSON(msg); err != nil {
// 			app.logger.Error("Error writing message to client:", err)
// 			break
// 		}
// 	}
// 	app.logger.Info("Client disconnected")
// }
