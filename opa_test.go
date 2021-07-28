package traefik_opa_plugin_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	traefik_opa_plugin "github.com/team-carepay/traefik-opa-plugin"
)

func TestServeHTTPOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data/testok" {
			t.Fatal(fmt.Sprintf("Path incorrect: %s", r.URL.Path))
		}
		param1 := r.URL.Query()["Param1"]
		if len(param1) != 2 || param1[0] != "foo" || param1[1] != "bar" {
			t.Fatal(fmt.Sprintf("Parameters incorrect, expected foo,bar but got %s", strings.Join(param1, ",")))
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{ \"result\": { \"allow\": true } }")
	}))
	defer ts.Close()
	cfg := traefik_opa_plugin.CreateConfig()
	cfg.URL = fmt.Sprintf("%s/v1/data/testok?Param1=foo&Param1=bar", ts.URL)
	cfg.AllowField = "allow"
	ctx := context.Background()
	nextCalled := false
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { nextCalled = true })

	opa, err := traefik_opa_plugin.New(ctx, next, cfg, "test-traefik-opa-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	opa.ServeHTTP(recorder, req)

	if recorder.Code == http.StatusForbidden {
		t.Fatal("Exptected OK")
	}
	if nextCalled == false {
		t.Fatal("next.ServeHTTP was not called")
	}
}

func TestServeHTTPForbidden(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{ \"result\": { \"allow\": false } }")
	}))
	defer ts.Close()
	cfg := traefik_opa_plugin.CreateConfig()
	cfg.URL = ts.URL
	cfg.AllowField = "allow"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { t.Fatal("Should not chain HTTP call") })

	opa, err := traefik_opa_plugin.New(ctx, next, cfg, "test-traefik-opa-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	opa.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatal("Exptected Forbidden")
	}
}

func TestServeHTTPEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "{}")
	}))
	defer ts.Close()
	cfg := traefik_opa_plugin.CreateConfig()
	cfg.URL = ts.URL
	cfg.AllowField = "allow"
	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { t.Fatal("Should not chain HTTP call") })

	opa, err := traefik_opa_plugin.New(ctx, next, cfg, "test-traefik-opa-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	opa.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatal("Exptected InternalServerError")
	}
}
