package loms

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/credentials/insecure"
	"route256/cart/vendor-proto/route256/loms"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	host   string
	client proto.LomsServiceClient
	conn   *grpc.ClientConn
}

func NewClient(host string) (*Client, error) {
	return &Client{
		host: host,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Connect() error {
	conn, err := grpc.Dial(
		c.host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = proto.NewLomsServiceClient(conn)
	return nil
}
