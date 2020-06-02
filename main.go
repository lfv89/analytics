package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
)

type Event struct {
	Source    string `json:"source"`
	UserAgent string `json:"userAgent"`
}

func handler(conn net.Conn) {
	buf := bufio.NewReader(conn)
	req, err := http.ReadRequest(buf)

	if err == io.EOF {
		log.Println("Client disconnected")
	}

	if err != nil {
		log.Println("An Unexpected error ocurred")
	}

	rawBody := &Event{
		Source:    req.Referer(),
		UserAgent: req.UserAgent(),
	}

	body, _ := json.Marshal(rawBody)
	es, _ := elasticsearch.NewDefaultClient()
	res, err := es.Index("events", bytes.NewReader(body))

	fmt.Println(res)
	fmt.Println(err)

	response := []byte("HTTP/1.1 200 OK\nContent-Type: image/jpeg\nContent-Length: 3\nAccess-Control-Allow-Origin: *\n\nimg")
	if _, err := conn.Write(response); err != nil {
		log.Fatalln("Unable to write data")
	}

	fmt.Println("Closing...")
	conn.Close()
}

func main() {
	listener, _ := net.Listen("tcp", ":4001")

	for {
		conn, _ := listener.Accept()
		go handler(conn)
	}
}
