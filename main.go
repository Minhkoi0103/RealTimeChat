package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	// Endpoint WebSocket
	http.HandleFunc("/ws", handleConnections)

	// Endpoint API để xem danh sách tin nhắn
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		messagesMu.RLock()
		defer messagesMu.RUnlock()
		json.NewEncoder(w).Encode(messages)
	})

	// Cung cấp file tĩnh (HTML, JS, CSS)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Goroutine xử lý gửi tin nhắn tới client
	go handleMessages()

	// --- 🔹 SỬA Ở ĐÂY: Lấy port động từ Railway ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback khi chạy local
	}

	fmt.Println("✅ Server đang chạy tại port:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("❌ Lỗi khởi động server:", err)
	}
}
