package tools

import (
	"context"
	"testing"
)

func TestNewWebSearchTool_TavilyConfigured(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "tavily-key",
		TavilyMaxResults: 7,
	})
	if tool == nil {
		t.Fatal("expected tool to be created")
	}
	if tool.provider == nil {
		t.Fatal("expected Tavily provider to be configured")
	}
	if got := tool.provider.Name(); got != searchProviderTavily {
		t.Fatalf("provider name = %q, want %q", got, searchProviderTavily)
	}
	if got := tool.provider.maxResults; got != 7 {
		t.Fatalf("provider maxResults = %d, want 7", got)
	}
}

func TestNewWebSearchTool_NoTavilyConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  WebSearchConfig
	}{
		{
			name: "disabled",
			cfg: WebSearchConfig{TavilyEnabled: false, TavilyAPIKey: "tavily-key"},
		},
		{
			name: "missing key",
			cfg: WebSearchConfig{TavilyEnabled: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWebSearchTool(tt.cfg)
			if tool == nil {
				t.Fatal("expected tool shell")
			}
			if tool.provider != nil {
				t.Fatal("expected provider to be absent")
			}
		})
	}
}

func TestWebSearchTool_UpdateConfig(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "old-key",
		TavilyMaxResults: 5,
	})
	if tool == nil {
		t.Fatal("expected initial tool")
	}

	tool.UpdateConfig(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "new-key",
		TavilyMaxResults: 9,
	})
	if tool.provider == nil {
		t.Fatal("expected provider after update")
	}
	if got := tool.provider.apiKey; got != "new-key" {
		t.Fatalf("provider apiKey = %q, want new-key", got)
	}
	if got := tool.provider.maxResults; got != 9 {
		t.Fatalf("provider maxResults = %d, want 9", got)
	}

	tool.UpdateConfig(WebSearchConfig{})
	if tool.provider != nil {
		t.Fatal("expected provider to be cleared when Tavily config is unavailable")
	}
}

func TestWebSearchTool_ResolveProvider_UsesBuiltinOverride(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "global-key",
		TavilyMaxResults: 5,
	})
	ctx := WithBuiltinToolSettings(context.Background(), BuiltinToolSettings{
		"web_search": []byte(`{"tavily":{"enabled":true,"max_results":8}}`),
	})

	provider := tool.resolveProvider(ctx)
	if provider == nil {
		t.Fatal("expected override provider")
	}
	if got := provider.maxResults; got != 8 {
		t.Fatalf("provider maxResults = %d, want 8", got)
	}
	if got := provider.apiKey; got != "global-key" {
		t.Fatalf("provider apiKey = %q, want global-key", got)
	}
}

func TestWebSearchTool_ResolveProvider_DisableOverride(t *testing.T) {
	tool := NewWebSearchTool(WebSearchConfig{
		TavilyEnabled:    true,
		TavilyAPIKey:     "global-key",
		TavilyMaxResults: 5,
	})
	ctx := WithBuiltinToolSettings(context.Background(), BuiltinToolSettings{
		"web_search": []byte(`{"tavily":{"enabled":false}}`),
	})

	if provider := tool.resolveProvider(ctx); provider != nil {
		t.Fatal("expected override to disable provider")
	}
}

func TestClampProviderResultCount(t *testing.T) {
	tests := []struct {
		requested, max, want int
	}{
		{5, 10, 5},
		{15, 10, 10},
		{0, 10, defaultSearchCount},
		{5, 0, 5},
	}
	for _, tt := range tests {
		got := clampProviderResultCount(tt.requested, tt.max)
		if got != tt.want {
			t.Errorf("clampProviderResultCount(%d, %d) = %d, want %d", tt.requested, tt.max, got, tt.want)
		}
	}
}

func TestNormalizeProviderMaxResults(t *testing.T) {
	tests := []struct {
		input, want int
	}{
		{0, defaultSearchCount},
		{-1, defaultSearchCount},
		{7, 7},
		{15, maxSearchCount},
	}
	for _, tt := range tests {
		got := normalizeProviderMaxResults(tt.input)
		if got != tt.want {
			t.Errorf("normalizeProviderMaxResults(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
