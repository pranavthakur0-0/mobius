package agentctx

import (
	"unicode/utf8"

	"mobius/pkg/llm"
)

// Heuristic token estimation constants
const (
	CharsPerToken         = 4       // ~4 characters per token for code & text
	BlockOverhead         = 4       // Structural framing overhead per tool call / result
	RoleOverhead          = 4       // Overhead per message role tag
	DefaultMaxTokens      = 128_000 // Standard 128k context window
	DefaultWatermarkRatio = 0.80    // Trigger compaction at 80% capacity
	DefaultRetainRatio    = 0.20    // Retain recent 20% of context intact
)

func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	charCount := utf8.RuneCountInString(text)
	return (charCount + CharsPerToken - 1) / CharsPerToken
}

func EstimateMessageTokens(msg llm.Message) int {
	tokens := RoleOverhead + EstimateTokens(string(msg.Role)) + EstimateTokens(msg.Content)
	if msg.ToolCallID != "" {
		tokens += EstimateTokens(msg.ToolCallID)
	}
	for _, tc := range msg.ToolCalls {
		tokens += BlockOverhead
		tokens += EstimateTokens(tc.Function.Name)
		tokens += EstimateTokens(tc.Function.Arguments)
	}

	return tokens
}

func EstimateConversationTokens(messages []llm.Message) int {
	total := 0
	for _, m := range messages {
		total += EstimateMessageTokens(m)
	}
	return total
}

// IsAboveWatermark checks if the current token count has reached or exceeded the watermark threshold.
func IsAboveWatermark(currentTokens int, maxTokens int, watermarkRatio float64) bool {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	if watermarkRatio <= 0 || watermarkRatio > 1.0 {
		watermarkRatio = DefaultWatermarkRatio
	}
	threshold := int(float64(maxTokens) * watermarkRatio)
	return currentTokens >= threshold
}
