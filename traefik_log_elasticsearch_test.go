package traefik_log_elasticsearch_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"

	traefik_log_elasticsearch "github.com/cmdbg/traefik-log-elasticsearch-plugin"
)

func TestLogElasticsearch(t *testing.T) {

	cfg := traefik_log_elasticsearch.CreateConfig()
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

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("next handler"))
	})

	elasticsearchLog := &traefik_log_elasticsearch.ElasticsearchLog{
		Next:             next,
		Name:             "test",
		Message:          cfg.Message,
		ElasticsearchURL: cfg.ElasticsearchURL,
		IndexName:        cfg.IndexName,
		Username:         cfg.Username,
		Password:         cfg.Password,
		VerifyTLS:        cfg.VerifyTLS,
	}

	req := httptest.NewRequest(http.MethodGet, "http://test.com/foo", nil)
	w := httptest.NewRecorder()

	elasticsearchLog.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatalf("Could not read response: %v", err)
	}

	if string(body) != "next handler" {
		t.Errorf("Handler did not chain to the next middleware. Got: %s", body)
	}
}
