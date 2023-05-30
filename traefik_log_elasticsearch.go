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

// Config is a structure that holds the configuration needed for the Elasticsearch plugin in Traefik.
// It contains the following fields:
// - ElasticsearchURL: This is the URL of the Elasticsearch instance that the plugin should interact with.
// - IndexName: This is the name of the Elasticsearch index that the plugin should write logs to.
// - Message: This is the default log message that will be used if no specific message is provided in the log entry.
// - APIKey: If provided, this will be used for authentication with the Elasticsearch instance. This should be used if Username and Password are not provided.
// - Username: This is the username to be used for authentication with the Elasticsearch instance. This is used in conjunction with Password and is an alternative to APIKey.
// - Password: This is the password to be used for authentication with the Elasticsearch instance. This is used in conjunction with Username and is an alternative to APIKey.
// - VerifyTLS: If true, the plugin will verify the TLS certificate of the Elasticsearch instance. It is recommended to always set this to true in production to prevent man-in-the-middle attacks.
type Config struct {
	ElasticsearchURL string
	IndexName        string
	Message          string
	APIKey           string
	Username         string
	Password         string
	VerifyTLS        bool
}

// CreateConfig returns a pointer to a Config struct with its fields initialized to zero values.
// This is a convenient way to create a new Config instance.
func CreateConfig() *Config {
	return &Config{}
}

// ElasticsearchLog is a middleware handler that logs HTTP requests to an Elasticsearch instance.
// It contains the following fields:
// - Next: The next handler to be called in the middleware chain. The ElasticsearchLog handler will call this after logging the request.
// - Name: The name of the handler. This is mainly used for identification and debugging purposes.
// - Message: The default message to be logged to Elasticsearch if no specific message is provided in the log entry.
// - ElasticsearchURL: The URL of the Elasticsearch instance where the logs should be written to.
// - IndexName: The name of the Elasticsearch index where the logs should be written to.
// - APIKey: If provided, this will be used for authentication with the Elasticsearch instance. This should be used if Username and Password are not provided.
// - Username: The username to be used for authentication with the Elasticsearch instance. This is used in conjunction with Password and is an alternative to APIKey.
// - Password: The password to be used for authentication with the Elasticsearch instance. This is used in conjunction with Username and is an alternative to APIKey.
// - VerifyTLS: If true, the middleware will verify the TLS certificate of the Elasticsearch instance. It is recommended to always set this to true in production to prevent man-in-the-middle attacks.
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

// New creates a new ElasticsearchLog middleware instance. The 'next' parameter specifies the
// handler to be executed after the middleware, and 'config' specifies the configuration settings.
// The 'name' parameter is a string identifier for the middleware instance.
// It returns the created middleware instance as an http.Handler and an error, if any.
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
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			version, ok := r["_version"].(float64)
			if !ok {
				log.Printf("Error: expected '_version' to be a float64")
				return
			}
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(version))
		}
	}

	e.Next.ServeHTTP(rw, req)
}
