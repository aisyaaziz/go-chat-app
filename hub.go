// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "log"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	// clients map[*Client]bool
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients.
	// broadcast chan []byte
	broadcast chan message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			room := h.rooms[client.roomID]
			if room == nil {
				// First client in the room, create a new one
				room = make(map[*Client]bool)
				h.rooms[client.roomID] = room
			}
			room[client] = true
		case client := <-h.unregister:
			room := h.rooms[client.roomID]
			if room != nil {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)
					if len(room) == 0 {
						// This was last client in the room, delete the room
						delete(h.rooms, client.roomID)
					}
				}
			}
		case message := <-h.broadcast:
			room := h.rooms[message.roomID]
			if room != nil {
				for client := range room {
					select {
					case client.send <- message.content:
						log.Print(client.uuid)
						log.Print(" : ")
						log.Println(message.content)
					default:
						close(client.send)
						delete(room, client)
					}
				}
				if len(room) == 0 {
					// The room was emptied while broadcasting to the room.  Delete the room.
					delete(h.rooms, message.roomID)
				}
			}
		}
		// select {
		// case client := <-h.register:
		// 	h.clients[client] = true
		// case client := <-h.unregister:
		// 	if _, ok := h.clients[client]; ok {
		// 		delete(h.clients, client)
		// 		close(client.send)
		// 	}
		// case message := <-h.broadcast:
		// 	for client := range h.clients {
		// 		select {
		// 		case client.send <- message:
		// 		default:
		// 			close(client.send)
		// 			delete(h.clients, client)
		// 		}
		// 	}
		// }
	}
}
