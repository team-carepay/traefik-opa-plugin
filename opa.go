package traefik_opa_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	OpaURL string `json:"opa-url,omitempty"`
	Field  string `json:"field,omitempty"`
}

// CreateConfig creates a new OPA Config
func CreateConfig() *Config {
	return &Config{}
}

// Opa contains the runtime config
type Opa struct {
	next  http.Handler
	url   string
	field string
}

// PayloadInput is the input payload
type PayloadInput struct {
	Host       string              `json:"host"`
	Method     string              `json:"method"`
	Path       []string            `json:"path"`
	Parameters url.Values          `json:"parameters"`
	Headers    map[string][]string `json:"headers"`
}

// Payload for OPA requests
type Payload struct {
	Input *PayloadInput `json:"input"`
}

// Response from OPA
type Response struct {
	Result map[string]json.RawMessage `json:"result"`
}

// New creates a new plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Opa{
		next:  next,
		url:   config.OpaURL,
		field: config.Field,
	}, nil
}

func (opaConfig *Opa) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	authPayloadAsJSON, err := json.Marshal(toOPAPayload(request))
	if err == nil {
		authResponse, err := http.Post(opaConfig.url, "application/json", bytes.NewBuffer(authPayloadAsJSON))
		if err == nil {
			body, err := ioutil.ReadAll(authResponse.Body)
			if err == nil {
				var result Response
				err := json.Unmarshal(body, &result)
				if err == nil {
					var allow bool
					err := json.Unmarshal(result.Result[opaConfig.field], &allow)
					if err == nil {
						if allow == true {
							opaConfig.next.ServeHTTP(rw, request)
						} else {
							rw.WriteHeader(http.StatusForbidden)
							rw.Write([]byte(body))
						}
						return
					}
				}
			}
		}
	}
	http.Error(rw, err.Error(), http.StatusInternalServerError)
}

func toOPAPayload(request *http.Request) *Payload {
	return &Payload{
		Input: &PayloadInput{
			Host:       request.Host,
			Method:     request.Method,
			Path:       strings.Split(request.URL.Path, "/")[1:],
			Parameters: request.URL.Query(),
			Headers:    request.Header,
		},
	}
}
