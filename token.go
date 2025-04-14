// Copyright (c) 2025 The mouser developers. All rights reserved.
// Project site: https://github.com/apidepot/mouser
// Use of this source code is governed by a MIT-style license that
// can be found in the LICENSE.txt file for the project.

package mouser

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// accessToken provides the response for a successful access token request.
// ExpiresIn is in seconds.
type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
	Type      string `json:"token_type"`
}

// getAccessToken returns the current access token or refreshes the access
// token using the client ID and client secret.
func (c *Client) getAccessToken() (string, error) {
	c.mu.RLock()
	if time.Now().Before(c.tokenExpiresAt) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	// Token is expred, so refresh.
	return c.refreshToken()
}

func (c *Client) refreshToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Seems like Mouser is doing something atypical, so I'm manually creating
	// the body to be in the form that Mouser expects to request an access
	// token.
	formData := fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=client_credentials",
		c.clientID,
		c.secret,
	)

	req, err := http.NewRequest("POST", c.accessTokenURL, strings.NewReader(formData))
	if err != nil {
		return "", fmt.Errorf("error creating post token request: %w", err)
	}

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error in post request for new access token: %w", err)
	}

	defer func(body io.Closer) {
		if err = body.Close(); err != nil {
			slog.Error("body close", "func", "refreshToken", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"bad status code (%d) from post for new access token: %s",
			resp.StatusCode,
			string(errorBody),
		)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	token := accessToken{}
	if err := json.Unmarshal(responseBody, &token); err != nil {
		return "", fmt.Errorf("error unmarshaling response body: %w", err)
	}

	// Remove one second from the time to expriration to be safe.
	c.accessToken = token.Token
	c.tokenType = token.Type
	c.tokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn - 1))

	return c.accessToken, nil

}
