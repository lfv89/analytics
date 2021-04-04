package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/lfv89/analytics/configs"
	"github.com/lfv89/analytics/private"
	"github.com/lfv89/analytics/private/socket"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

var hub *socket.Hub
var c configs.Config

func init() {
	envconfig.Process("analytics", &c)
}

func main() {
	hub = socket.NewHub()
	go hub.Run()

	http.HandleFunc("/", Index)
	http.HandleFunc("/notify", Notify)

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		Subscribe(hub, w, r)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", configs.GetPort("4002")), nil)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	cfg := elasticsearch.Config{
		Addresses: []string{
			c.Elastic.URL,
		},
	}

	client, _ := elasticsearch.NewClient(cfg)
	config := private.StoreConfig{Client: client}

	myStore, _ := private.NewStore(config)
	results, resultsErr := myStore.Search("")

	if resultsErr != nil {
		log.Fatal(resultsErr)
	}

	json.NewEncoder(w).Encode(results)
}

func Notify(w http.ResponseWriter, r *http.Request) {
	clients := hub.Clients

	// read JSON from body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// unmarshal JSON to Event
	event := private.Event{}
	err = json.Unmarshal(b, &event)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	for client, _ := range clients {
		if client.Id == event.ClientID {
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(event)

			client.Send <- reqBodyBytes.Bytes()
		}
	}
}

func Subscribe(hub *socket.Hub, w http.ResponseWriter, r *http.Request) {
	clientId := 123 // get from session later
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &socket.Client{Id: clientId, Hub: hub, Conn: conn, Send: make(chan []byte, 256)}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
