package traefik_log_elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type Config struct {
	ElasticsearchURL string
	IndexName        string
	Message          string
	APIKey           string
	Username         string
	Password         string
	VerifyTLS        bool
}

func CreateConfig() *Config {
	return &Config{}
}

type ElasticsearchLog struct {
	Next             http.Handler
	Name             string
	Message          string
	ElasticsearchURL string
	IndexName        string
	APIKey           string
	Username         string
	Password         string
	VerifyTLS        bool
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
	if (len(config.APIKey) == 0) && (len(config.Username) == 0 || len(config.Password) == 0) {
		return nil, errors.New("missing Elasticsearch credentials")
	}

	elasticsearchLog := &ElasticsearchLog{
		ElasticsearchURL: config.ElasticsearchURL,
		IndexName:        config.IndexName,
		Next:             next,
		Name:             name,
		Username:         config.Username,
		Password:         config.Password,
		APIKey:           config.APIKey,
		VerifyTLS:        config.VerifyTLS,
	}

	return elasticsearchLog, nil
}

func convertToJSON(data map[string]interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}
	return string(jsonData)
}

func (e *ElasticsearchLog) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var cfg elasticsearch.Config
	if !e.VerifyTLS {
		// Create a TLS config that skips certificate verification.
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		// Create a transport to use our TLS config.
		transport := &http.Transport{TLSClientConfig: tlsConfig}

		cfg = elasticsearch.Config{
			Addresses: []string{
				e.ElasticsearchURL,
			},
			Transport: transport,
			Username:  e.Username,
			Password:  e.Password,
			APIKey:    e.APIKey,
		}
	} else {
		cfg = elasticsearch.Config{
			Addresses: []string{
				e.ElasticsearchURL,
			},
			Username: e.Username,
			Password: e.Password,
			APIKey:   e.APIKey,
		}
	}

	// Create a client
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	id := uuid.New().String()

	msg := map[string]interface{}{
		"message": e.Message,
	}

	// Set up the Elasticsearch request object directly
	esReq := esapi.IndexRequest{
		Index:      e.IndexName,
		DocumentID: id,
		Body:       strings.NewReader(convertToJSON(msg)),
		Refresh:    "true",
	}

	res, err := esReq.Do(context.Background(), es)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), 1)
		log.Printf("%d", res.StatusCode)
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

	e.Next.ServeHTTP(rw, req)
}
