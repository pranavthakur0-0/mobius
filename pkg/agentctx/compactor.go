package agentctx

import (
	"context"
	"fmt"
	"mobius/pkg/llm"
	"strings"
)

// CompactorConfig holds settings for automatic context compaction.
type CompactorConfig struct {
	MaxTokens      int
	WatermarkRatio float64
	RetainRatio    float64
}

// DefaultCompactorConfig returns standard production settings.
func DefaultCompactorConfig() CompactorConfig {
	return CompactorConfig{
		MaxTokens:      DefaultMaxTokens,
		WatermarkRatio: DefaultWatermarkRatio,
		RetainRatio:    DefaultRetainRatio,
	}
}

// Compactor manages token metering, checkpoint summarization, and safe history pruning.
type Compactor struct {
	config   CompactorConfig
	provider llm.Provider
	model    string
}

// NewCompactor initializes a fresh context compactor with fallback defaults.
func NewCompactor(cfg CompactorConfig, provider llm.Provider, model string) *Compactor {
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}
	if cfg.WatermarkRatio <= 0 || cfg.WatermarkRatio > 1.0 {
		cfg.WatermarkRatio = DefaultWatermarkRatio
	}
	if cfg.RetainRatio <= 0 || cfg.RetainRatio >= cfg.WatermarkRatio {
		cfg.RetainRatio = DefaultRetainRatio
	}

	return &Compactor{
		config:   cfg,
		provider: provider,
		model:    model,
	}
}

// ShouldCompact checks if the compactor should trigger based on the watermark threshold.
func (cp *Compactor) ShouldCompact(currentTokens int) bool {
	return IsAboveWatermark(currentTokens, cp.config.MaxTokens, cp.config.WatermarkRatio)
}

// Compact performs the full checkpoint summarization and in-place history reduction.
func (cp *Compactor) Compact(ctx context.Context, c *ConversationContext) error {
	messages := c.Messages()
	if len(messages) <= 3 {
		return fmt.Errorf("conversation history too short to compact (%d messages)", len(messages))
	}

	// 1. Calculate how many tokens we want to retain in the recent tail
	retainTokenBudget := int(float64(cp.config.MaxTokens) * cp.config.RetainRatio)

	// 2. Find the safe cut point that preserves recent turns and tool pairings
	cutIndex := cp.findSafeCutIndex(messages, retainTokenBudget)
	if cutIndex <= 1 {
		return fmt.Errorf("unable to find a safe compaction cut point")
	}

	// Older messages to summarize (indices 1 to cutIndex-1, skipping 0 which is system prompt)
	olderMessages := messages[1:cutIndex]
	olderTokens := EstimateConversationTokens(olderMessages)

	// 3. Build summarization payload and call the LLM
	summarizerPayload := BuildCompactionMessages(olderMessages)
	req := &llm.ChatRequest{
		Model:   cp.model,
		Message: summarizerPayload,
	}

	resp, err := cp.provider.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("summarization LLM call failed: %w", err)
	}

	rawSummary := strings.TrimSpace(resp.Content)
	if rawSummary == "" {
		return fmt.Errorf("summarization returned empty output")
	}

	// 4. Shrink Check: Verify the summary is strictly smaller than original older history
	summaryTokens := EstimateTokens(rawSummary)
	if summaryTokens >= olderTokens {
		return fmt.Errorf("shrink check failed: summary (%d tokens) >= original (%d tokens)", summaryTokens, olderTokens)
	}

	// 5. In-Place History Replacement:
	// New history = [System Prompt (0)] + [Compacted Checkpoint (1)] + [Recent Tail (cutIndex...end)]
	checkpointMsg := llm.Message{
		Role:    llm.RoleUser,
		Content: FrameSummary(rawSummary),
	}

	recentTail := messages[cutIndex:]
	newMessages := make([]llm.Message, 0, 2+len(recentTail))
	newMessages = append(newMessages, messages[0])   // Keep system prompt
	newMessages = append(newMessages, checkpointMsg) // Add checkpoint
	newMessages = append(newMessages, recentTail...) // Add recent turns

	c.SetMessages(newMessages)
	return nil
}

// findSafeCutIndex scans backwards from the end of history to keep recent turns within retainTokenBudget,
// while ensuring we NEVER cut between an assistant tool_call and its tool results.
func (cp *Compactor) findSafeCutIndex(messages []llm.Message, retainTokenBudget int) int {
	tailTokens := 0
	cutIndex := len(messages) - 1

	// 1. Scan backwards accumulating tokens until we reach the retain budget
	for i := len(messages) - 1; i >= 1; i-- {
		tailTokens += EstimateMessageTokens(messages[i])
		if tailTokens >= retainTokenBudget {
			cutIndex = i
			break
		}
	}

	// 2. Safety check: ensure cutIndex does not separate a tool result from its tool call.
	// If messages[cutIndex] is a RoleTool message, move backwards past all tool results
	for cutIndex > 1 && messages[cutIndex].Role == llm.RoleTool {
		cutIndex--
	}

	// If messages[cutIndex] is the RoleAssistant that initiated tool calls, move backwards before it
	if cutIndex > 1 && messages[cutIndex].Role == llm.RoleAssistant && len(messages[cutIndex].ToolCalls) > 0 {
		cutIndex--
	}

	return cutIndex
}

// UpdateModel updates the model and provider when switching models dynamically.
func (cp *Compactor) UpdateModel(model string, provider llm.Provider) {
	cp.model = model
	cp.provider = provider
}
