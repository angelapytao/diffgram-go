package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client is a thin wrapper over an amqp091 Connection that exposes a Channel()
// factory. Each consumer opens its own Channel for QoS isolation.
type Client struct {
	conn *amqp.Connection
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Channel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Ping() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	return ch.Close()
}
