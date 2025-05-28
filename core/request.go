package core

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
)

type Headers struct {
	Authorization string `json:"Authorization,omitempty"`
	ContentType   string `json:"Content-Type,omitempty"`
	Custom        map[string]string
}

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type HttpRequestParams struct {
	Method  HTTPMethod
	URL     string
	Headers Headers
	Body    []byte
}

func SendHttpRequest(params HttpRequestParams) (*http.Response, error) {
	req, err := http.NewRequest(string(params.Method), params.URL, bytes.NewReader(params.Body))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar a requisição: %v", err)
	}

	if params.Headers.Authorization != "" {
		req.Header.Add("Authorization", params.Headers.Authorization)
	}

	if params.Headers.ContentType != "" {
		req.Header.Add("Content-Type", params.Headers.ContentType)
	} else {
		req.Header.Add("Content-Type", "application/json")
	}

	for key, value := range params.Headers.Custom {
		req.Header.Add(key, value)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar a requisição: %v", err)
	}

	return resp, nil
}
