package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrade http to websocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// client with unique uuid and websocket connection
type Client struct {
	ID   string
	Conn *websocket.Conn
}

// message json format
type Message struct {
	From string `json:"from"`
	To   string `json:"to"`
	Msg  string `json:"msg"`
}

// map to store client id as key and client connection as value
var clients = make(map[string]*Client)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// handler function
func (m *Manager) handleConnectionData(w http.ResponseWriter, r *http.Request) {
	//upgrade http to websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	//close the connection
	defer func() {
		ws.Close()
		log.Println("client disconnected")
	}()

	clientID := uuid.New().String()           //generate new uuid
	client := &Client{ID: clientID, Conn: ws} //create new client
	clients[clientID] = client                //save new client in map

	log.Println("new client connected with id:", clientID)

	ws.WriteJSON(map[string]string{
		"connection_id": clientID,
	})

	for {
		_, receivedMsg, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(receivedMsg, &msg); err != nil {
			log.Println("invalid input received:", string(receivedMsg))
			continue
		}

		msg.From = clientID
		fmt.Printf("Message from %s to %s: %s\n", msg.From, msg.To, msg.Msg)

		if target, ok := clients[msg.To]; ok {
			err = target.Conn.WriteJSON(msg)
			if err != nil {
				log.Println("write error:", err)
			}
		} else {
			ws.WriteJSON(map[string]string{
				"error": "receiver client not found",
			})
		}
	}
}

func main() {
	manager := NewManager()

	http.HandleFunc("/ws", manager.handleConnectionData)

	fmt.Println("server started at ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
