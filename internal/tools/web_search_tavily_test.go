package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTavilySearchProvider_Search(t *testing.T) {
	var gotAuth string
	var gotContentType string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/search" {
			t.Fatalf("path = %q, want /search", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{
					"title":   "Tavily Result",
					"url":     "https://example.com/tavily",
					"content": "Snippet from Tavily",
				},
			},
		})
	}))
	defer srv.Close()

	provider := newTavilySearchProvider("tvly-test")
	provider.client = srv.Client()
	provider.endpoint = srv.URL + "/search"

	results, err := provider.Search(context.Background(), searchParams{Query: "golang", Count: 3})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotAuth != "Bearer tvly-test" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer tvly-test")
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["query"] != "golang" {
		t.Fatalf("query = %#v, want %q", gotBody["query"], "golang")
	}
	if gotBody["max_results"] != float64(3) {
		t.Fatalf("max_results = %#v, want %d", gotBody["max_results"], 3)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Title != "Tavily Result" || results[0].URL != "https://example.com/tavily" || results[0].Description != "Snippet from Tavily" {
		t.Fatalf("unexpected result = %#v", results[0])
	}
}

func TestTavilySearchProvider_SearchErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":{"error":"rate limit"}}`, http.StatusTooManyRequests)
	}))
	defer srv.Close()

	provider := newTavilySearchProvider("tvly-test")
	provider.client = srv.Client()
	provider.endpoint = srv.URL

	_, err := provider.Search(context.Background(), searchParams{Query: "golang", Count: 3})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "tavily API returned status 429") {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestNewWebSearchTool_PriorityAndDefaults(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		BraveEnabled:     true,
		BraveAPIKey:      "brave-key",
		BraveMaxResults:  7,
		TavilyEnabled:    true,
		TavilyAPIKey:     "tvly-key",
		TavilyMaxResults: 4,
		DDGEnabled:       true,
		DDGMaxResults:    3,
		CacheTTL:         time.Minute,
	})
	if tool == nil {
		t.Fatal("expected tool")
	}
	if len(tool.providers) != 3 {
		t.Fatalf("len(providers) = %d, want 3", len(tool.providers))
	}
	if tool.providers[0].provider.Name() != "brave" || tool.providers[1].provider.Name() != "tavily" || tool.providers[2].provider.Name() != "duckduckgo" {
		t.Fatalf("unexpected provider order: %s, %s, %s", tool.providers[0].provider.Name(), tool.providers[1].provider.Name(), tool.providers[2].provider.Name())
	}
	if tool.defaultSearchCount != 7 {
		t.Fatalf("defaultSearchCount = %d, want 7", tool.defaultSearchCount)
	}
}

func TestNewWebSearchTool_TavilyDefaultWhenPrimary(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "tvly-key",
		TavilyMaxResults: 4,
		DDGEnabled:       true,
		DDGMaxResults:    3,
		CacheTTL:         time.Minute,
	})
	if tool == nil {
		t.Fatal("expected tool")
	}
	if len(tool.providers) != 2 {
		t.Fatalf("len(providers) = %d, want 2", len(tool.providers))
	}
	if tool.providers[0].provider.Name() != "tavily" || tool.providers[1].provider.Name() != "duckduckgo" {
		t.Fatalf("unexpected provider order: %s, %s", tool.providers[0].provider.Name(), tool.providers[1].provider.Name())
	}
	if tool.defaultSearchCount != 4 {
		t.Fatalf("defaultSearchCount = %d, want 4", tool.defaultSearchCount)
	}
}

func TestWebSearchTool_ExecuteUsesProviderSpecificMaxResults(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		BraveEnabled:     true,
		BraveAPIKey:      "brave-key",
		BraveMaxResults:  7,
		TavilyEnabled:    true,
		TavilyAPIKey:     "tvly-key",
		TavilyMaxResults: 4,
		DDGEnabled:       true,
		DDGMaxResults:    3,
		CacheTTL:         time.Minute,
	})
	if tool == nil {
		t.Fatal("expected tool")
	}

	braveCalls := 0
	tool.providers[0] = searchProviderEntry{
		provider: stubSearchProvider{
			name: "brave",
			searchFn: func(_ context.Context, params searchParams) ([]searchResult, error) {
				braveCalls++
				if params.Count != 7 {
					t.Fatalf("brave count = %d, want 7", params.Count)
				}
				if params.Country != "US" {
					t.Fatalf("brave country = %q, want %q", params.Country, "US")
				}
				return nil, assertErr{"brave failed"}
			},
		},
		maxResults:           7,
		supportsSearchParams: true,
	}
	tool.providers[1] = searchProviderEntry{
		provider: stubSearchProvider{
			name: "tavily",
			searchFn: func(_ context.Context, params searchParams) ([]searchResult, error) {
				if params.Count != 4 {
					t.Fatalf("tavily count = %d, want 4", params.Count)
				}
				if params.Country != "" || params.SearchLang != "" || params.UILang != "" || params.Freshness != "" {
					t.Fatalf("tavily params should clear provider-specific search fields: %#v", params)
				}
				return []searchResult{{Title: "ok", URL: "https://example.com"}}, nil
			},
		},
		maxResults: 4,
	}

	result := tool.Execute(context.Background(), map[string]any{
		"query":       "golang",
		"count":       float64(9),
		"country":     "US",
		"search_lang": "en",
		"ui_lang":     "en",
		"freshness":   "pw",
	})
	if result == nil || result.IsError {
		t.Fatalf("expected success result, got %#v", result)
	}
	if braveCalls != 1 {
		t.Fatalf("brave calls = %d, want 1", braveCalls)
	}
}

type stubSearchProvider struct {
	name     string
	searchFn func(context.Context, searchParams) ([]searchResult, error)
}

func (s stubSearchProvider) Name() string { return s.name }

func (s stubSearchProvider) Search(ctx context.Context, params searchParams) ([]searchResult, error) {
	return s.searchFn(ctx, params)
}

type assertErr struct{ msg string }

func (e assertErr) Error() string { return e.msg }
