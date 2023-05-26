package traefik_log_elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Config struct {
	ElasticsearchURL string `json:"elasticsearchURL,omitempty"`
	IndexName        string `json:"indexName,omitempty"`
	Message          string `json:"message,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type ElasticsearchLog struct {
	next             http.Handler
	name             string
	message          string
	elasticsearchURL string
	indexName        string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.ElasticsearchURL) == 0 {
		return nil, errors.New("missing Elasticsearch URL")
	}
	if len(config.IndexName) == 0 {
		return nil, errors.New("missing Elasticsearch index name")
	}
	if len(config.Message) == 0 {
		return nil, errors.New("missing Elasticsearch message")
	}
	return &ElasticsearchLog{
		elasticsearchURL: config.ElasticsearchURL,
		indexName:        config.IndexName,
		next:             next,
		name:             name,
	}, nil
}

func (e *ElasticsearchLog) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Create a client
	es, _ := elasticsearch.NewDefaultClient()

	// Set up the request object directly
	req = esapi.IndexRequest{
		Index:      e.indexName,
		DocumentID: strconv.Itoa(1),
		Body:       strings.NewReader(e.message),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), 1)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}

	e.next.ServeHTTP(rw, req)
}
