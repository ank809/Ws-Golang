package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Room struct {
	roomId     string
	clientList map[*Client]bool
	broadcast  chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var roomList = make(map[string]*Room)

// func to create room
func getOrcreateRoom(roomId string) *Room {
	room, exists := roomList[roomId]
	if !exists {
		room = &Room{
			roomId:     roomId,
			clientList: make(map[*Client]bool),
			broadcast:  make(chan []byte),
		}
		roomList[roomId] = room
		fmt.Printf("Room created with roomId %v\n", roomId) // Added newline for clarity
		go room.run()
	}
	return room
}

// func to add create user
func addUser(conn *websocket.Conn, roomId string) {
	client := &Client{
		conn: conn,
		send: make(chan []byte),
	}
	room := getOrcreateRoom(roomId)
	room.clientList[client] = true
	go client.readMessage(room)
	go client.writeMessage()
}

func (c *Client) readMessage(room *Room) {
	defer func() {
		// room.unregisterClient(c)
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		room.broadcast <- message
	}
}

func (c *Client) writeMessage() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func (r *Room) run() {
	for {
		select {
		case message := <-r.broadcast:
			for client := range r.clientList {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clientList, client)
				}
			}
		}
	}
}

func webserver(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	roomId := r.URL.Query().Get("roomId")
	if roomId == "" {
		log.Println("roomId is required")
		conn.Close()
		return
	}
	addUser(conn, roomId)
}
