package config

import "testing"

func TestConfigSecrets_TavilyLifecycle(t *testing.T) {
	cfg := Default()
	cfg.Tools.Web.Tavily.APIKey = "tvly-secret"

	masked := cfg.MaskedCopy()
	if masked.Tools.Web.Tavily.APIKey != secretMask {
		t.Fatalf("masked Tavily key = %q, want %q", masked.Tools.Web.Tavily.APIKey, secretMask)
	}

	extracted := cfg.ExtractDBSecrets()
	if extracted["tools.web.tavily.api_key"] != "tvly-secret" {
		t.Fatalf("extracted Tavily key = %q", extracted["tools.web.tavily.api_key"])
	}

	cfg.StripSecrets()
	if cfg.Tools.Web.Tavily.APIKey != "" {
		t.Fatalf("StripSecrets Tavily key = %q, want empty", cfg.Tools.Web.Tavily.APIKey)
	}

	cfg.Tools.Web.Tavily.APIKey = secretMask
	cfg.StripMaskedSecrets()
	if cfg.Tools.Web.Tavily.APIKey != "" {
		t.Fatalf("StripMaskedSecrets Tavily key = %q, want empty", cfg.Tools.Web.Tavily.APIKey)
	}

	cfg.ApplyDBSecrets(map[string]string{"tools.web.tavily.api_key": "tvly-from-db"})
	if cfg.Tools.Web.Tavily.APIKey != "tvly-from-db" {
		t.Fatalf("ApplyDBSecrets Tavily key = %q, want %q", cfg.Tools.Web.Tavily.APIKey, "tvly-from-db")
	}
}
