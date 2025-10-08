package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ---- Các biến toàn cục ----
var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// broadcast Message structs to all clients
var broadcast = make(chan Message)

// store message history so clients can reload all messages
var messages []Message
var messagesMu sync.RWMutex

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// ---- Hàm nhận kết nối mới ----
func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Lỗi upgrade:", err)
        return
    }
    defer ws.Close()

    // register client (thread-safe)
    clientsMu.Lock()
    clients[ws] = true
    clientsMu.Unlock()

    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            // remove client on error
            clientsMu.Lock()
            delete(clients, ws)
            clientsMu.Unlock()
            break
        }

        // try to parse incoming message as JSON Message
        var m Message
        if err := json.Unmarshal(msg, &m); err != nil {
            // if it's not JSON, treat the whole payload as content and default sender
            m = Message{Sender: "noname", Content: string(msg)}
        }

        // append to history
        messagesMu.Lock()
        messages = append(messages, m)
        messagesMu.Unlock()

        // broadcast Message to all connected clients
        broadcast <- m
    }
}

// ---- Hàm gửi tin nhắn tới tất cả client ----
func handleMessages() {
    for {
        msg := <-broadcast
        // take a snapshot of clients to avoid holding lock while writing
        clientsMu.Lock()
        conns := make([]*websocket.Conn, 0, len(clients))
        for c := range clients {
            conns = append(conns, c)
        }
        clientsMu.Unlock()

        // marshal Message to JSON before sending
        payload, err := json.Marshal(msg)
        if err != nil {
            // skip on marshal error
            continue
        }

        for _, client := range conns {
            err := client.WriteMessage(websocket.TextMessage, payload)
            if err != nil {
                client.Close()
                // remove from map
                clientsMu.Lock()
                delete(clients, client)
                clientsMu.Unlock()
            }
        }
    }
}
