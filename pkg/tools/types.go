package tools

type Tool interface {
	// Name returns the unique identifier for the tool
	Name() string
	// Description explains to the LLM what the tool does and how to use it
	Description() string
	// Execute runs the tool with the given input arguments and returns (output, error)
	Execute(args string) (string, error)
}
