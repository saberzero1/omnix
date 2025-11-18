package checks

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/juspay/omnix/pkg/nix"
)

// Caches checks that required caches are configured
type Caches struct {
	// Required is the list of required cache URLs
	Required []string `yaml:"required" json:"required"`
}

// DefaultCaches returns a Caches check with default required caches
func DefaultCaches() Caches {
	return Caches{
		Required: []string{"https://cache.nixos.org"},
	}
}

// Check verifies that all required caches are configured
func (c *Caches) Check(ctx context.Context, nixInfo *nix.Info) []NamedCheck {
	// Get configured substituters from nix config
	configuredCaches := nixInfo.Config.Substituters.Value

	missingCaches := c.getMissingCaches(configuredCaches)

	var result CheckResult
	if len(missingCaches) == 0 {
		result = GreenResult{}
	} else {
		result = RedResult{
			Message: fmt.Sprintf(
				"You are missing some required caches: %s",
				strings.Join(missingCaches, " "),
			),
			Suggestion: fmt.Sprintf(
				"Caches can be added in your %s (see https://nixos.wiki/wiki/Binary_Cache#Using_a_binary_cache). "+
					"Cachix caches can also be added using `nix run nixpkgs#cachix use <name>`.",
				nixInfo.Env.OS.NixConfigLabel(),
			),
		}
	}

	check := Check{
		Title:    "Nix Caches in use",
		Info:     fmt.Sprintf("substituters = %s", strings.Join(configuredCaches, " ")),
		Result:   result,
		Required: true,
	}

	return []NamedCheck{
		{Name: "caches", Check: check},
	}
}

// getMissingCaches returns the subset of required caches not in the configured list
func (c *Caches) getMissingCaches(configured []string) []string {
	var missing []string

	// Normalize configured caches
	configuredSet := make(map[string]bool)
	for _, cache := range configured {
		normalized := normalizeURL(cache)
		configuredSet[normalized] = true
	}

	// Check which required caches are missing
	for _, required := range c.Required {
		normalized := normalizeURL(required)
		if !configuredSet[normalized] {
			missing = append(missing, required)
		}
	}

	return missing
}

// normalizeURL normalizes a URL for comparison
func normalizeURL(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	// Ensure trailing slash is removed for comparison
	return strings.TrimSuffix(parsed.String(), "/")
}

// CachixCache represents a Cachix cache
type CachixCache struct {
	Name string
}

// ParseCachixURL parses a Cachix cache URL and returns the cache name
// Returns nil if the URL is not a Cachix cache
func ParseCachixURL(urlStr string) *CachixCache {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	host := parsed.Hostname()
	if !strings.HasSuffix(host, ".cachix.org") {
		return nil
	}

	// Extract cache name (e.g., "foo" from "foo.cachix.org")
	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return nil
	}

	return &CachixCache{Name: parts[0]}
}
