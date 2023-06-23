//go:build !generated
// +build !generated

// Package traefik_plugin_elastic provides a Traefik middleware plugin
// that logs HTTP request details to an Elasticsearch instance.
package traefik_plugin_elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/google/uuid"
)

// Config is a structure that holds the configuration needed for the Elasticsearch plugin in Traefik.
type Config struct {
	// ElasticsearchURL is the URL of the Elasticsearch instance that the plugin should interact with.
	ElasticsearchURL string
	// IndexName is the name of the Elasticsearch index that the plugin should write logs to.
	IndexName string
	// Message is the default log message that will be used if no specific message is provided in the log entry.
	Message string
	// APIKey is used for authentication with the Elasticsearch instance. This should be used if Username and Password are not provided.
	APIKey string
	// Username is the username to be used for authentication with the Elasticsearch instance. This is an alternative to APIKey.
	Username string
	// Password is the password to be used for authentication with the Elasticsearch instance. This is an alternative to APIKey.
	Password string
	// VerifyTLS determines whether the plugin should verify the TLS certificate of the Elasticsearch instance.
	// It is recommended to set this to true in production to prevent man-in-the-middle attacks.
	VerifyTLS bool
}

// CreateConfig returns a pointer to a Config struct with its fields initialized to zero values.
// This is a convenient way to create a new Config instance.
func CreateConfig() *Config {
	return &Config{}
}

// ElasticsearchLog is a middleware handler that logs HTTP requests to an Elasticsearch instance.
type ElasticsearchLog struct {
	// Next is the next handler to be called in the middleware chain. The ElasticsearchLog handler will call this after logging the request.
	Next http.Handler
	// Name is the name of the handler. This is mainly used for identification and debugging purposes.
	Name string
	// Message is the default message to be logged to Elasticsearch if no specific message is provided in the log entry.
	Message string
	// ElasticsearchURL is the URL of the Elasticsearch instance where the logs should be written to.
	ElasticsearchURL string
	// IndexName is the name of the Elasticsearch index where the logs should be written to.
	IndexName string
	// APIKey is used for authentication with the Elasticsearch instance. This should be used if Username and Password are not provided.
	APIKey string
	// Username is the username to be used for authentication with the Elasticsearch instance. This is an alternative to APIKey.
	Username string
	// Password is the password to be used for authentication with the Elasticsearch instance. This is an alternative to APIKey.
	Password string
	// VerifyTLS determines whether the middleware should verify the TLS certificate of the Elasticsearch instance.
	// It is recommended to set this to true in production to prevent man-in-the-middle attacks.
	VerifyTLS bool
}

// New creates a new ElasticsearchLog middleware instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
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
		tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec

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
			// Note: VerifyTLS is set to true by default when using the elasticsearch.Config struct.
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

	res, err := esReq.Do(req.Context(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Fatalf("Error closing the response body: %s", err)
		}
	}()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), 1)
		log.Printf("%d", res.StatusCode)
		return
	}

	// Deserialize the response into a map.
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
		return
	}

	version, ok := r["_version"].(float64)
	if !ok {
		log.Printf("Error: expected '_version' to be a float64")
		return
	}

	log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(version))

	e.Next.ServeHTTP(rw, req)
}
