package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SeatUpdateMessage struct {
	ScreeningID string `json:"screening_id"`
	SeatID      string `json:"seat_id"`
	Status      string `json:"status"` // AVAILABLE, LOCKED, BOOKED
}

type Hub struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan SeatUpdateMessage
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	Mutex      sync.Mutex
}

var WSHub *Hub

func InitWSHub() {
	WSHub = &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan SeatUpdateMessage),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
	go WSHub.Run()
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.Mutex.Unlock()
		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
			}
			h.Mutex.Unlock()
		case message := <-h.Broadcast:
			msgBytes, _ := json.Marshal(message)
			h.Mutex.Lock()
			for client := range h.Clients {
				err := client.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					log.Printf("WS Error: %v", err)
					client.Close()
					delete(h.Clients, client)
				}
			}
			h.Mutex.Unlock()
		}
	}
}

func ServeWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	WSHub.Register <- ws
}
