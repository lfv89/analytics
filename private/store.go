package private

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/kelseyhightower/envconfig"
	"github.com/lfv89/analytics/configs"
)

var c configs.Config

const searchAll = `
	"query" : { "match_all" : {} },
	"size" : 100
`

const searchMatch = `
	"query" : { "match_all" : {} },
	"size" : 100
`

type Hit struct {
	ID        string `json:"_id"`
	Source    string `json:"source"`
	UserAgent string `json:"userAgent"`
}

type Store struct {
	indexName string
	es        *elasticsearch.Client
}

type StoreConfig struct {
	IndexName string
	Client    *elasticsearch.Client
}

type SearchResults struct {
	Hits  []*Hit `json:"hits"`
	Total int    `json:"total"`
}

func init() {
	envconfig.Process("analytics", &c)
}

func BuildStore() *Store {
	cfg := elasticsearch.Config{
		Addresses: []string{
			c.Elastic.URL,
		},
	}

	client, _ := elasticsearch.NewClient(cfg)
	eventStore, _ := newStore(StoreConfig{Client: client})

	return eventStore
}

func newStore(c StoreConfig) (*Store, error) {
	indexName := c.IndexName

	if indexName == "" {
		indexName = "events"
	}

	return &Store{es: c.Client, indexName: indexName}, nil
}

func (s *Store) buildQuery(query string, after ...string) io.Reader {
	var b strings.Builder

	b.WriteString("{\n")

	if query == "" {
		b.WriteString(searchAll)
	} else {
		b.WriteString(fmt.Sprintf(searchMatch, query))
	}

	if len(after) > 0 && after[0] != "" && after[0] != "null" {
		b.WriteString(",\n")
		b.WriteString(fmt.Sprintf(`	"search_after": %s`, after))
	}

	b.WriteString("\n}")

	return strings.NewReader(b.String())
}

func (s *Store) Index(indexName string, body []byte) *esapi.Response {
	result, _ := s.es.Index("events", bytes.NewReader(body))

	return result
}

func (s *Store) Search(query string, after ...string) (*SearchResults, error) {
	var results SearchResults

	res, err := s.es.Search(
		s.es.Search.WithIndex(s.indexName),
		s.es.Search.WithBody(s.buildQuery(query, after...)),
	)
	if err != nil {
		return &results, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return &results, err
		}
		return &results, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	type envelopeResponse struct {
		Took int
		Hits struct {
			Total struct {
				Value int
			}
			Hits []struct {
				ID     string          `json:"_id"`
				Source json.RawMessage `json:"_source"`
			}
		}
	}

	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return &results, err
	}

	results.Total = r.Hits.Total.Value

	if len(r.Hits.Hits) < 1 {
		results.Hits = []*Hit{}
		return &results, nil
	}

	for _, hit := range r.Hits.Hits {
		var h Hit
		h.ID = hit.ID

		if err := json.Unmarshal(hit.Source, &h); err != nil {
			return &results, err
		}

		results.Hits = append(results.Hits, &h)
	}

	return &results, nil
}
