// Copyright (c) 2025 The mouser developers. All rights reserved.
// Project site: https://github.com/apidepot/mouser
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package mouser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	apiURL         = "https://api.mouser.com/"
	accessTokenURL = "https://api.mouser.com/v1/oauth2/token"
	grantType      = "client_credentials"
)

// Client models a client to consume the Mouser API.
type Client struct {
	customerID     string
	baseURL        string
	accessTokenURL string
	clientID       string
	secret         string
	accessToken    string
	tokenType      string
	localeSite     string
	localeLanguage string
	localeCurrency string
	tokenExpiresAt time.Time
	httpClient     *http.Client
	rateLimiter    *rate.Limiter
	mu             sync.RWMutex
	logger         *slog.Logger
}

// Error represents an IEX API error
type Error struct {
	StatusCode int
	Message    string
}

// ClientOption applies an option to the client.
type ClientOption func(*Client)

// Error implements the error interface
func (e Error) Error() string {
	return fmt.Sprintf("%d %s: %s", e.StatusCode, http.StatusText(e.StatusCode), e.Message)
}

// NewClient creates a mouser client with the given client ID and client
// secret.
func NewClient(customerID, clientID, secret string, opts ...ClientOption) (*Client, error) {
	c := &Client{
		customerID:     customerID,
		clientID:       clientID,
		secret:         secret,
		localeSite:     "US",
		localeLanguage: "en",
		localeCurrency: "USD",
		httpClient:     &http.Client{Timeout: time.Second * 60},
		tokenExpiresAt: time.Now(),
		logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),

		// Set default values, which may be overridden by user options.
		baseURL:        apiURL,
		accessTokenURL: accessTokenURL,
		rateLimiter:    rate.NewLimiter(rate.Every(time.Second), 100),
	}

	// Apply options using the functional option pattern.
	for _, opt := range opts {
		opt(c)
	}

	// Get the access token.
	if _, err := c.getAccessToken(); err != nil {
		return nil, err
	}

	c.logger.Info("mouser client created", "base url", c.baseURL)

	return c, nil
}

// WithSandbox sets the baseURL to the default sandbox URL.
func WithDefaultSandbox() ClientOption {
	return func(client *Client) {
		client.baseURL = sandboxURL
		client.accessTokenURL = sandboxTokenURL
	}
}

// WithBaseURL changes the baseURL for a new IEX Client from the default
// Mouser API base URL to the given base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

// WithRateLimiter sets the rate limiter instead of using the default rate
// limiter.
func WithRateLimiter(duration time.Duration, numRequests int) ClientOption {
	return func(client *Client) {
		client.rateLimiter = rate.NewLimiter(rate.Every(duration), numRequests)
	}
}

// Get retrieves the JSON data from the given endpoint and unmarshals the data
// into v.
func (c *Client) Get(ctx context.Context, endpoint string, v any) error {
	// Create the full URL from the given endpoint.
	u, err := c.url(endpoint, nil)
	if err != nil {
		return err
	}
	return c.GetFromURL(ctx, u, v)
}

// GetWithQueryParams gets the JSON data from the given endpoint with the query
// parameters attached and unmarshals the JSON data to v.
func (c *Client) GetWithQueryParams(ctx context.Context,
	endpoint string, queryParams map[string]string, v any) error {
	queryParams["token"] = c.accessToken

	// Create the full URL from the given endpoint and query parameters.
	u, err := c.url(endpoint, queryParams)
	if err != nil {
		return err
	}
	return c.GetFromURL(ctx, u, v)
}

// GetFromURL fetches the JSON data from the given URL and unmarshals it into
// v.
func (c *Client) GetFromURL(ctx context.Context, u *url.URL, v any) error {
	data, err := c.getBytes(ctx, u.String())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// GetJSONWithoutToken gets the JSON data from the given endpoint without
// adding a token to the URL.
func (c *Client) GetJSONWithoutToken(ctx context.Context, endpoint string, v any) error {
	u, err := c.url(endpoint, nil)
	if err != nil {
		return err
	}
	return c.GetFromURL(ctx, u, v)
}

func (c *Client) getBytes(ctx context.Context, address string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	// Debug: print request details
	fmt.Println("Request URL:", req.URL)
	fmt.Println("Request Method:", req.Method)
	fmt.Println("Request Headers:", req.Header)

	err = c.rateLimiter.Wait(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return []byte{}, err
	}
	defer func(body io.Closer) {
		if err = body.Close(); err != nil {
			slog.Error("body close", "func", "getBytes", "error", err)
		}
	}(resp.Body)

	// Even if GET didn't return an error, check the status code to make sure
	// everything was ok.
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		msg := ""

		if err == nil {
			msg = string(b)
		}

		return []byte{}, Error{StatusCode: resp.StatusCode, Message: msg}
	}
	return io.ReadAll(resp.Body)
}

// Returns a URL object that points to the endpoint with optional query parameters.
func (c *Client) url(endpoint string, queryParams map[string]string) (*url.URL, error) {
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, err
	}

	if queryParams != nil {
		q := u.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
	}
	return u, nil
}
