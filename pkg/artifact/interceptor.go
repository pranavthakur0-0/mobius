package artifact

import "fmt"

// Thresholds for tool output handling
const (
	SmallThreshold   = 2000  // Under 2k chars: keep full output in context
	PreviewThreshold = 8000  // 2k-8k chars: include a preview
	PreviewLines     = 20    // Number of lines to show in preview
)

// Result holds the processed tool output after interception.
type Result struct {
	Observation string // What goes into the LLM context (full, preview, or ref)
	FullOutput  string // The complete original output
	ArtifactRef string // artifact:// reference if offloaded, empty otherwise
	WasOffloaded bool
}

// Intercept processes tool output and decides whether to offload.
func Intercept(store *Store, threadID, toolName, fullOutput string) Result {
	size := len(fullOutput)

	// Small output: keep everything in context
	if size <= SmallThreshold {
		return Result{
			Observation: fullOutput,
			FullOutput:  fullOutput,
		}
	}

	// Medium or large: offload to artifact store
	_, ref, err := store.Save(threadID, fullOutput)
	if err != nil {
		// Fallback: truncate if artifact save fails
		return Result{
			Observation: fullOutput[:SmallThreshold] + "\n... [truncated]",
			FullOutput:  fullOutput,
		}
	}

	// Build a preview (first N lines)
	preview := buildPreview(fullOutput, PreviewLines)

	observation := fmt.Sprintf("Output saved to %s (%d bytes).\nPreview:\n%s\nUse read_artifact for full content.", ref, size, preview)

	return Result{
		Observation:  observation,
		FullOutput:   fullOutput,
		ArtifactRef:  ref,
		WasOffloaded: true,
	}
}

func buildPreview(content string, maxLines int) string {
	lines := splitLines(content)
	if len(lines) <= maxLines {
		return content
	}
	result := ""
	for i := 0; i < maxLines; i++ {
		result += lines[i] + "\n"
	}
	result += fmt.Sprintf("... (%d more lines)", len(lines)-maxLines)
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
