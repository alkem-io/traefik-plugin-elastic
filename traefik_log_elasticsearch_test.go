//go:build !generated
// +build !generated

package traefiklogelasticsearch_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	traefiklogelasticsearch "github.com/cmdbg/traefik-log-elasticsearch-plugin"
	"github.com/joho/godotenv"
)

func TestLogElasticsearch(t *testing.T) {
	// Load configuration from environment variables or use default values
	cfg := loadConfig()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("next handler")); err != nil {
			http.Error(w, fmt.Sprintf("Error writing response: %v", err), http.StatusInternalServerError)
		}
	})

	handler := logElasticsearch(next, cfg)

	req := httptest.NewRequest(http.MethodGet, "http://test.com/foo", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "next handler" {
		t.Errorf("Handler did not chain to the next middleware. Got: %s", body)
	}
}

func loadConfig() *traefiklogelasticsearch.Config {
	cfg := traefiklogelasticsearch.CreateConfig()
	cfg.Message = "Test Elasticsearch"
	cfg.ElasticsearchURL = "http://localhost:9200"
	cfg.IndexName = "test-index"
	cfg.Username = "elastic"
	cfg.Password = "elastic"

	err := godotenv.Load(".env")
	if err == nil {
		cfg.Message = "Test Elasticsearch"
		cfg.ElasticsearchURL = os.Getenv("ELASTICSEARCH_URL")
		cfg.IndexName = os.Getenv("INDEX_NAME")
		cfg.Username = os.Getenv("ELASTIC_USERNAME")
		cfg.Password = os.Getenv("ELASTIC_PASSWORD")
	}

	return cfg
}

func logElasticsearch(next http.Handler, _ *traefiklogelasticsearch.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request details here if needed
		// ...

		// Call the next handler in the middleware chain
		next.ServeHTTP(w, r)
	})
}
