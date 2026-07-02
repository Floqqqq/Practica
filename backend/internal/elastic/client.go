package elastic

import (
	"context"
	"fmt"
	"net/http"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

type Client struct {
	es *elasticsearch.Client
}

func NewClient(address string) (*Client, error) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	return &Client{
		es: esClient,
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	response, err := c.es.Ping(
		c.es.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("ping elasticsearch: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch ping failed with status: %s", response.Status())
	}

	return nil
}

func (c *Client) Raw() *elasticsearch.Client {
	return c.es
}
