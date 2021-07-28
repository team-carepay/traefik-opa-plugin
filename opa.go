package traefik_opa_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Config the plugin configuration.
type Config struct {
	URL        string `json:"url,omitempty"`
	AllowField string `json:"allow-field,omitempty"`
}

// CreateConfig creates a new OPA Config
func CreateConfig() *Config {
	return &Config{}
}

// Opa contains the runtime config
type Opa struct {
	next       http.Handler
	url        string
	allowField string
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
		next:       next,
		url:        config.URL,
		allowField: config.AllowField,
	}, nil
}

func (opaConfig *Opa) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	err := opaConfig.ServeHTTPInternal(rw, request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (opaConfig *Opa) ServeHTTPInternal(rw http.ResponseWriter, request *http.Request) error {
	authPayloadAsJSON, err := json.Marshal(toOPAPayload(request))
	if err != nil {
		return err
	}
	authResponse, err := http.Post(opaConfig.url, "application/json", bytes.NewBuffer(authPayloadAsJSON))
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(authResponse.Body)
	if err != nil {
		return err
	}
	var result Response
	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}
	if len(result.Result) == 0 {
		return fmt.Errorf("OPA result invalid")
	}
	fieldResult, ok := result.Result[opaConfig.allowField]
	if !ok {
		return fmt.Errorf("OPA result missing: %v", opaConfig.allowField)
	}
	var allow bool
	if err = json.Unmarshal(fieldResult, &allow); err != nil {
		return err
	}
	if allow == true {
		opaConfig.next.ServeHTTP(rw, request)
	} else {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte(body))
	}
	return nil
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
