package product

import (
	"encoding/json"
	"fmt"
	"net/http"
	"route256/cart/internal/model"
)

type HTTPError struct {
	StatusCode int
	Err        error
}

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseUrl string) (*Client, error) {
	return &Client{
		baseURL: baseUrl,
		client:  &http.Client{},
	}, nil
}

func (c *Client) GetProduct(productId int64) (*model.Product, error) {
	url := fmt.Sprintf("%s/product/%d", c.baseURL, productId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error calling product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, model.ErrPreconditionFailed
	}

	var product model.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode product: %w", err)
	}

	return &product, nil
}
