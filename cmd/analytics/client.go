package main

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{
		Scheme: "ws",
		Path:   "/subscribe",
		Host:   "localhost:4002",
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		log.Fatalln(err)
	}

	go Receive(c)
	Send(c)
}

func Send(c *websocket.Conn) {
	// message := []byte("I am sending this...")

	for {

// 		err := c.WriteMessage(websocket.TextMessage, message)

// 		if err != nil {
// 			log.Fatalln(err)
// 		}

		time.Sleep(100 * time.Millisecond)
	}
}

func Receive(c *websocket.Conn) {
	for {
		_, msg, err := c.ReadMessage()

		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("recv: %s", msg)
	}
}
