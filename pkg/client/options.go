package client

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the [Client].
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
//
// Use this to configure custom transport settings, proxies, or TLS configuration.
//
// Example:
//
//	transport := &http.Transport{
//	    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//	}
//	httpClient := &http.Client{Transport: transport}
//	c := client.New(url, client.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout.
//
// The default timeout is 5 minutes.
//
// Example:
//
//	c := client.New(url, client.WithTimeout(60*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}
