package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ProviderConfig defines the TOML configuration block for an LLM provider.
type ProviderConfig struct {
	Name    string   `toml:"name"`     // Display name (e.g. "DeepSeek", "OpenAI")
	Type    string   `toml:"type"`     // Protocol type used for factory lookup (e.g. "openai", "anthropic")
	BaseURL string   `toml:"base_url"` // Endpoint base URL
	EnvKey  string   `toml:"env_key"`  // Environment variable containing the API key
	Models  []string `toml:"models"`   // List of model IDs supported by this provider
}

// Config represents the top-level configuration loaded from config.toml.
type Config struct {
	ActiveModel string                    `toml:"default_model"`
	Providers   map[string]ProviderConfig `toml:"providers"`
}


type ModelInfo struct {
	ProviderName	string
	Model			string
}

// LoadConfig reads and unmarshals a TOML configuration file from the specified path.
func LoadConfig(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("config.toml path is required")
	}

	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}
	return &cfg, nil
}

// ListAllModels extracts and returns a flat list of all model names across all configured providers.
func (c *Config) ListAllModels() (map[string]*ModelInfo, error) {
	if c == nil {
		return nil, fmt.Errorf("config is nil")
	}
	allModels := make(map[string]*ModelInfo)
	for _, provider := range c.Providers {
		models := provider.Models
		for _, model := range models {
			allModels[model] = &ModelInfo {
				ProviderName: provider.Name,
				Model: model,
			}
		}
	}

	if len(allModels) == 0 {
		return nil, fmt.Errorf("no models found in config")
	}

	return allModels, nil
}

// GetProviderForModel resolves the model name to its provider configuration and instantiates the Provider via factory.
func (c *Config) GetProviderForModel(modelName string) (Provider, error) {
	if modelName == "" {
		modelName = c.ActiveModel
	}

	for _, prov := range c.Providers {
		for _, m := range prov.Models {
			if modelName == m {
				factory, ok := providerFactories[prov.Type]
				if !ok {
					return nil, fmt.Errorf("unsupported provider type: %s", prov.Type)
				}
				apiKey := os.Getenv(prov.EnvKey)
				return factory(apiKey, prov.BaseURL), nil
			}
		}
	}

	return nil, fmt.Errorf("no provider configured for model %q", modelName)
}
