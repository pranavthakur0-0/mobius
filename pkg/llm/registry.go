package llm

// ProviderFactory is a constructor function type that initializes an LLM Provider.
type ProviderFactory func(apiKey, baseURL string) Provider

// providerFactories holds the registered factory functions for each provider protocol type.
var providerFactories = map[string]ProviderFactory{}

// RegisterProviderFactory registers a constructor function under a given provider type name.
func RegisterProviderFactory(typeName string, factory ProviderFactory) {
	providerFactories[typeName] = factory
}