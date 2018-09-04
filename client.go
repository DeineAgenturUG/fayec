package faye

import (
	"github.com/thesyncim/faye/message"
	"github.com/thesyncim/faye/transport"
)

type options struct {
	inExtension  message.Extension
	outExtension message.Extension
	transport    transport.Transport
}

var defaultOpts = options{
	transport: transport.GetTransport("websocket"),
}

//https://faye.jcoglan.com/architecture.html
type client interface {
	Subscribe(subscription string, onMsg func(message *message.Message)) error
	Publish(subscription string, message *message.Message) error
	//todo unsubscribe,etc
}

type Option func(*options)

var _ client = (*Client)(nil)

type Client struct {
	opts options
}

func NewClient(url string, opts ...Option) (*Client, error) {
	var c Client
	c.opts = defaultOpts
	for _, opt := range opts {
		opt(&c.opts)
	}
	return &c, nil
}

func (c *Client) Subscribe(subscription string, onMsg func(message *message.Message)) error {
	panic("not implemented")
}

func (c *Client) Publish(subscription string, message *message.Message) error {
	panic("not implemented")
}
