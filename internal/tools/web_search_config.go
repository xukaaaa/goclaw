package tools

import (
	"strings"
	"time"

	"github.com/nextlevelbuilder/goclaw/internal/config"
)

// WebSearchConfig holds configuration for the web search tool.
type WebSearchConfig struct {
	TavilyAPIKey     string
	TavilyEnabled    bool
	TavilyMaxResults int
	CacheTTL         time.Duration
}

// WebSearchConfigFromConfig creates a WebSearchConfig from the global config.
func WebSearchConfigFromConfig(cfg *config.Config) WebSearchConfig {
	return WebSearchConfig{
		TavilyEnabled:    cfg.Tools.Web.Tavily.Enabled,
		TavilyAPIKey:     cfg.Tools.Web.Tavily.APIKey,
		TavilyMaxResults: cfg.Tools.Web.Tavily.MaxResults,
	}
}

// --- Shared provider helpers ---

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func coalesceSearchText(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func clampProviderResultCount(requested, providerMax int) int {
	if requested <= 0 {
		requested = defaultSearchCount
	}
	if providerMax > 0 && requested > providerMax {
		return providerMax
	}
	return requested
}

func normalizeProviderMaxResults(value int) int {
	if value <= 0 {
		return defaultSearchCount
	}
	if value > maxSearchCount {
		return maxSearchCount
	}
	return value
}
