package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"route256/cart/internal/model"
	"time"
)

type HTTPError struct {
	StatusCode int
	Err        error
}

type Client struct {
	baseURL     string
	client      *http.Client
	rateLimiter *rate.Limiter
}

func NewClient(baseUrl string) (*Client, error) {
	return &Client{
		baseURL:     baseUrl,
		client:      &http.Client{},
		rateLimiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 10),
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, productId int64) (*model.Product, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("request canceled while waiting for rate limit: %w", ctx.Err())
	default:
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit error: %w", err)
		}
	}

	url := fmt.Sprintf("%s/product/%d", c.baseURL, productId)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer([]byte{}))
	resp, err := c.client.Do(req)
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
