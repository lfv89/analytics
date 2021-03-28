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
	"github.com/kelseyhightower/envconfig"
	"github.com/lfv89/analytics/configs"
	"github.com/lfv89/analytics/private"
)

var c configs.Config

func init() {
	envconfig.Process("analytics", &c)
}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GetPort("4001")))

	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, _ := listener.Accept()
		go Handler(conn)
	}
}

func Handler(conn net.Conn) {
	buf := bufio.NewReader(conn)
	req, err := http.ReadRequest(buf)

	if err == io.EOF {
		log.Println("Client disconnected")
	}

	if err != nil {
		log.Println("An Unexpected error ocurred")
	}

	rawBody := &private.Event{
		Source:    req.Referer(),
		UserAgent: req.UserAgent(),
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			c.Elastic.URL,
		},
	}

	body, _ := json.Marshal(rawBody)
	es, _ := elasticsearch.NewClient(cfg)
	res, err := es.Index("events", bytes.NewReader(body))

	http.Post(c.Web.NotifyURL, "text/plain", bytes.NewBuffer(body))

	fmt.Println(res)
	fmt.Println(err)

	response := []byte("HTTP/1.1 200 OK\nContent-Type: image/jpeg\nContent-Length: 3\nAccess-Control-Allow-Origin: *\n\nimg")
	if _, err := conn.Write(response); err != nil {
		log.Fatalln("Unable to write data")
	}

	conn.Close()
}
