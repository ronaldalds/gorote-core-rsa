package core

import (
	"bytes"
	"fmt"
	"net/http"
)

// Definindo um tipo personalizado para os métodos HTTP
type HTTPMethod string

// Definindo constantes para os métodos HTTP válidos
const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
)

type Headers struct {
	Authorization string            `json:"Authorization,omitempty"` // Header de autenticação
	ContentType   string            `json:"Content-Type,omitempty"`  // Tipo do conteúdo
	Custom        map[string]string // Headers personalizados adicionais
}

// GraphQLRequest encapsula a requisição GraphQL
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// HttpRequestParams encapsula os parâmetros para a requisição
type HttpRequestParams struct {
	Method  HTTPMethod // Método HTTP: GET, POST, PUT, DELETE.
	URL     string     // URL da API
	Headers Headers    // Headers personalizados
	Body    []byte     // Corpo da requisição (opcional)
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar a requisição: %v", err)
	}

	return resp, nil
}
