package client

import immudb "github.com/codenotary/immudb/pkg/client"

type ClientOption interface {
	Set(*Client)
}

type ClientOptionFunc func(*Client)

func (f ClientOptionFunc) Set(c *Client) {
	f(c)
}

func Limit(limit uint32) ClientOption {
	return ClientOptionFunc(func(c *Client) {
		c.limit = limit
	})
}

func ClientOptions(options *immudb.Options) ClientOption {
	return ClientOptionFunc(func(c *Client) {
		c.options = options
	})
}

func Verified(value ...bool) ClientOption {
	return ClientOptionFunc(func(c *Client) {
		if len(value) == 0 {
			c.verified = true
		}

		c.verified = value[0]
	})
}
