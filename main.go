package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
    http.HandleFunc("/ws", handleConnections)
    http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        messagesMu.RLock()
        defer messagesMu.RUnlock()
        json.NewEncoder(w).Encode(messages)
    })
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/", fs)

    go handleMessages()

    fmt.Println("Server đang chạy tại :8080")
    http.ListenAndServe(":8080", nil)
}
