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
	"os"

	"github.com/elastic/go-elasticsearch/v7"
)

type Event struct {
	Source    string `json:"source"`
	UserAgent string `json:"userAgent"`
}

func main() {
	fmt.Println("Listening...")
	listener, _ := net.Listen("tcp", getPort())

	for {
		conn, _ := listener.Accept()
		fmt.Println("Handling...")
		go handler(conn)
	}
}

func getPort() string {
	value := os.Getenv("PORT")

	if len(value) == 0 {
		return "4001"
	}

	return value
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

	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elastic:9200",
		},
	}

	body, _ := json.Marshal(rawBody)
	es, _ := elasticsearch.NewClient(cfg)
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
