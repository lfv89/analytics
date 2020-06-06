package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	store "analytics/private"

	"github.com/elastic/go-elasticsearch/v7"
)

func main() {
	http.HandleFunc("/", Index)
	http.ListenAndServe(fmt.Sprintf(":%s", getPort()), nil)
}

func getPort() string {
	value := os.Getenv("PORT")

	if len(value) == 0 {
		return "4002"
	}

	return value
}

func Index(w http.ResponseWriter, r *http.Request) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://elastic:9200",
		},
	}

	es, _ := elasticsearch.NewClient(cfg)
	config := store.StoreConfig{Client: es}
	myStore, _ := store.NewStore(config)
	search := Search{store: myStore}
	results, err := search.getResults("")

	if err != nil {
		log.Fatal(err)
	}

	tmpl, _ := template.ParseFiles("web/index.html")
	tmpl.Execute(w, results)
}

type Search struct {
	store *store.Store
}

func (s *Search) getResults(query string) (*store.SearchResults, error) {
	return s.store.Search(query)
}
