package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/pelletier/go-toml/v2"
)

type ProviderConfig struct {
    Name    string   `toml:"name"`
    Type    string   `toml:"type"`
    BaseURL string   `toml:"base_url"`
    EnvKey  string   `toml:"env_key"`
    Models  []string `toml:"models"`
}


type Config struct {
	ActiveModel  string                    `toml:"default_model"`
	Providers    map[string]ProviderConfig `toml:"providers"`
}



// Load the config provider

func LoadConfig(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("Config.toml path is required")
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
		return nil, fmt.Errorf("Failed to read config file at %s: %w", path, err)
	}

	var cfg Config
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
		return &cfg, nil
}



func (c *Config) ListAllModels() ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var allModels []string
	for _, provider := range c.Providers {
		allModels = append(allModels, provider.Models...)
	}

	if len(allModels) == 0 {
		return nil, fmt.Errorf("no models found in config")
	}

	return allModels, nil
}




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

