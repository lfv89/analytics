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
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"github.com/lfv89/analytics/configs"
	"github.com/lfv89/analytics/private"
)

var c configs.Config
var eventStore *private.Store

func init() {
	eventStore = private.BuildStore()
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

	id, _ := strconv.Atoi(req.URL.Query()["clientId"][0])

	rawBody := &private.Event{
		Source:    req.Referer(),
		UserAgent: req.UserAgent(),
		ClientID:  id,
	}

	body, _ := json.Marshal(rawBody)
	res := eventStore.Index("events", body)

	http.Post(c.Web.NotifyURL, "text/plain", bytes.NewBuffer(body))

	fmt.Println(res)
	fmt.Println(err)

	response := []byte("HTTP/1.1 200 OK\nContent-Type: image/jpeg\nContent-Length: 3\nAccess-Control-Allow-Origin: *\n\nimg")
	if _, err := conn.Write(response); err != nil {
		log.Fatalln("Unable to write data")
	}

	conn.Close()
}
