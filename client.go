// Copyright 2026 Mataki Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type managementClient struct {
	baseURL    string
	apiKey     string
	apiVersion string
	httpClient *http.Client
}

type apiResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (c *managementClient) do(ctx context.Context, method string, path string, body any) (apiResponse, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return apiResponse{}, fmt.Errorf("encode request: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.url(path), reader)
	if err != nil {
		return apiResponse{}, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("API-Version", c.apiVersion)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiResponse{}, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResponse{}, fmt.Errorf("read response: %w", err)
	}
	apiResp := apiResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: data}
	if resp.StatusCode >= 400 {
		return apiResp, fmt.Errorf("%s %s failed: %s", method, path, responseErrorMessage(apiResp))
	}
	return apiResp, nil
}

func (c *managementClient) url(path string) string {
	base := strings.TrimRight(c.baseURL, "/")
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

func responseErrorMessage(resp apiResponse) string {
	if len(resp.Body) == 0 {
		return http.StatusText(resp.StatusCode)
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body, &payload); err == nil {
		for _, key := range []string{"message", "error", "detail"} {
			if value, ok := payload[key].(string); ok && value != "" {
				return value
			}
		}
	}
	return strings.TrimSpace(string(resp.Body))
}

func printAPIResponse(resp apiResponse, jsonOutput bool) error {
	if len(resp.Body) == 0 {
		if jsonOutput {
			fmt.Println("{}")
			return nil
		}
		fmt.Println("ok")
		return nil
	}
	if json.Valid(resp.Body) {
		if jsonOutput {
			var value any
			if err := json.Unmarshal(resp.Body, &value); err != nil {
				return err
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(value)
		}
		var value any
		if err := json.Unmarshal(resp.Body, &value); err == nil {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(value)
		}
	}
	fmt.Println(strings.TrimSpace(string(resp.Body)))
	return nil
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func workspacePath(workspace string, suffix string) string {
	return "/workspaces/" + url.PathEscape(workspace) + suffix
}

func workspaceOrActive(flagValue string, g *Globals) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if active := g.activeWorkspace(); active != "" {
		return active, nil
	}
	project, err := LoadProjectConfig(".")
	if err == nil && project.Workspace != "" {
		return project.Workspace, nil
	}
	return "", fmt.Errorf("workspace required. Pass --workspace or run `microwave workspace use <id>`")
}
