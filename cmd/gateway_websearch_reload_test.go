package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/nextlevelbuilder/goclaw/internal/agent"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

func TestReloadWebSearchTool_ReplacesRegistryAndInvalidatesAgents(t *testing.T) {
	cfg := config.Default()
	cfg.Tools.Web.Brave.Enabled = false
	cfg.Tools.Web.Brave.APIKey = ""
	cfg.Tools.Web.Tavily.Enabled = true
	cfg.Tools.Web.Tavily.APIKey = "tvly-secret"
	cfg.Tools.Web.DuckDuckGo.Enabled = false

	toolsReg := tools.NewRegistry()
	toolsReg.Register(tools.NewWebSearchTool(tools.WebSearchConfig{
		DDGEnabled:    true,
		DDGMaxResults: 5,
	}))

	router := agent.NewRouter()
	resolverCalls := 0
	router.SetResolver(func(_ context.Context, agentKey string) (agent.Agent, error) {
		resolverCalls++
		return nil, errors.New("resolver invoked")
	})
	if _, err := router.Get(t.Context(), "test-agent"); err == nil {
		t.Fatal("expected resolver error before reload")
	}
	if resolverCalls != 1 {
		t.Fatalf("resolver calls before reload = %d, want 1", resolverCalls)
	}

	reloadWebSearchTool(toolsReg, cfg, router)

	tool, ok := toolsReg.Get("web_search")
	if !ok {
		t.Fatal("web_search not registered after reload")
	}
	webSearch, ok := tool.(*tools.WebSearchTool)
	if !ok {
		t.Fatalf("web_search has unexpected type %T", tool)
	}
	result := webSearch.Execute(t.Context(), map[string]any{"query": "golang", "count": float64(10)})
	if result == nil {
		t.Fatal("web_search execute returned nil result")
	}
	if _, err := router.Get(t.Context(), "test-agent"); err == nil {
		t.Fatal("expected resolver error after invalidation")
	}
	if resolverCalls != 2 {
		t.Fatalf("resolver calls after reload = %d, want 2", resolverCalls)
	}
}

func TestReloadWebSearchTool_UnregistersWhenNoProvidersConfigured(t *testing.T) {
	cfg := config.Default()
	cfg.Tools.Web.Brave.Enabled = false
	cfg.Tools.Web.Brave.APIKey = ""
	cfg.Tools.Web.Tavily.Enabled = false
	cfg.Tools.Web.Tavily.APIKey = ""
	cfg.Tools.Web.DuckDuckGo.Enabled = false

	toolsReg := tools.NewRegistry()
	toolsReg.Register(tools.NewWebSearchTool(tools.WebSearchConfig{
		DDGEnabled:    true,
		DDGMaxResults: 5,
	}))

	reloadWebSearchTool(toolsReg, cfg, nil)

	if _, ok := toolsReg.Get("web_search"); ok {
		t.Fatal("web_search should be unregistered when all providers are disabled")
	}
}
