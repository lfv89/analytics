package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"analytics/configs"
	store "analytics/private"
	"analytics/private/socket"

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
}

func Index(w http.ResponseWriter, r *http.Request) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			c.Elastic.URL,
		},
	}

	client, _ := elasticsearch.NewClient(cfg)
	config := store.StoreConfig{Client: client}

	myStore, _ := store.NewStore(config)
	results, resultsErr := myStore.Search("")

	if resultsErr != nil {
		log.Fatal(resultsErr)
	}

	tmpl, _ := template.ParseFiles("web/index.html")
	tmpl.Execute(w, results)
}

func Notify(w http.ResponseWriter, r *http.Request) {
	clientId := 1
	clients := hub.Clients

	for client, _ := range clients {
		if client.Id == clientId {
			client.Send <- []byte("Only for client with ID == 1")
		}
	}
}

func Subscribe(hub *socket.Hub, w http.ResponseWriter, r *http.Request) {
	clientId := len(hub.Clients) + 1
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
