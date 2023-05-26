package traefik_log_elasticsearch_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cmdbg/traefik_log_elasticsearch"
)

func TestLogElasticsearch(t *testing.T) {
	cfg := traefik_log_elasticsearch.CreateConfig()
	cfg.Message = ""
	cfg.ElasticsearchURL = ""
	cfg.IndexName = ""

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	handler, err := traefik_log_elasticsearch.New(ctx, next, cfg, "traefik-log-elasticsearch-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeader(t, req, "X-Host", "localhost")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()

	if req.Header.Get(key) != expected {
		t.Errorf("invalid header value: %s", req.Header.Get(key))
	}
}
