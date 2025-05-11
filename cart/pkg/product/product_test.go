package product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"route256/cart/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetProduct_Success(t *testing.T) {
	expected := model.Product{
		Id:    123,
		Name:  "Sample Product",
		Price: 1000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/product/123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  server.Client(),
	}

	got, err := client.GetProduct(123)
	require.NoError(t, err)
	require.Equal(t, &expected, got)
}

func TestClient_GetProduct_HTTPError(t *testing.T) {
	// Simulate 400 response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  server.Client(),
	}

	_, err := client.GetProduct(123)
	require.ErrorIs(t, err, model.ErrPreconditionFailed)
}

func TestClient_GetProduct_DecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json"))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  server.Client(),
	}

	_, err := client.GetProduct(123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode product")
}
