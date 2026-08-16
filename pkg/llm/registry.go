package llm



// ProviderFactory is a constructor function that creates a Provider
type ProviderFactory func(apiKey, baseURL string) Provider




// Global factory registry — providers register themselves here
var providerFactories = map[string]ProviderFactory{}



// RegisterProviderFactory lets each provider file register its constructor
func RegisterProviderFactory(typeName string, factory ProviderFactory) {
    providerFactories[typeName] = factory
}