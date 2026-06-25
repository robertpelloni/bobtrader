package main

import (
	"fmt"
	"log"
	"time"
	"github.com/gorilla/websocket"
)

func main() {
	url := "wss://stream.binance.us:9443/ws/btcusdt@ticker"
	log.Printf("Connecting to %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Printf("Connected. Waiting for message...")
	c.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, message, err := c.ReadMessage()
	if err != nil {
		log.Fatal("read:", err)
	}
	fmt.Printf("Received: %s\n", message)
}
