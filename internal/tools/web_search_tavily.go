package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const tavilyMaxResponseBodyBytes = 1 << 20

type tavilySearchProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func newTavilySearchProvider(apiKey string) *tavilySearchProvider {
	return &tavilySearchProvider{
		apiKey:   apiKey,
		endpoint: tavilySearchEndpoint,
		client:   &http.Client{Timeout: time.Duration(searchTimeoutSeconds) * time.Second},
	}
}

func (p *tavilySearchProvider) Name() string { return "tavily" }

func (p *tavilySearchProvider) Search(ctx context.Context, params searchParams) ([]searchResult, error) {
	requestBody := struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results,omitempty"`
	}{
		Query:      params.Query,
		MaxResults: params.Count,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, tavilyMaxResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily API returned status %d", resp.StatusCode)
	}

	var tavilyResp struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &tavilyResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]searchResult, 0, len(tavilyResp.Results))
	for _, r := range tavilyResp.Results {
		results = append(results, searchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Content,
		})
	}

	return results, nil
}
