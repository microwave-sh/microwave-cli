package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const apiVersion = "2026-05-27"

type Client struct {
	baseURL string
	token   string
	version string
	debug   bool
	http    *http.Client
}

func New(baseURL, token, version string, debug bool) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		version: version,
		debug:   debug,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type APIError struct {
	Status  int
	Type    string `json:"type"`
	Message string `json:"message"`
	Errors  []struct {
		Field   string `json:"field"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		parts := make([]string, len(e.Errors))
		for i, fe := range e.Errors {
			parts[i] = fmt.Sprintf("%s: %s", fe.Field, fe.Message)
		}
		return fmt.Sprintf("%s (%s)", e.Message, strings.Join(parts, "; "))
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("request failed with status %d", e.Status)
}

func (c *Client) Do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(raw)
		if c.debug {
			log.Printf("→ %s %s\n%s", method, path, raw)
		}
	} else if c.debug {
		log.Printf("→ %s %s", method, path)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", apiVersion)
	req.Header.Set("User-Agent", "microwave-cli/"+c.version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if c.debug {
		log.Printf("← %d\n%s", resp.StatusCode, raw)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{Status: resp.StatusCode}
		_ = json.Unmarshal(raw, apiErr)
		return apiErr
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
