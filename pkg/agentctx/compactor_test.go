package agentctx

import (
	"context"
	"fmt"
	"mobius/pkg/llm"
	"strings"
	"testing"
)

// mockProvider implements llm.Provider for testing compaction.
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) Generate(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{
		Content: m.response,
		Usage: llm.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

func TestTokenEstimation(t *testing.T) {
	// 1. Raw string token estimation
	if got := EstimateTokens("hello world"); got != 3 {
		t.Errorf("EstimateTokens('hello world') = %d; want 3", got)
	}

	// 2. Message estimation
	msg := llm.Message{
		Role:    llm.RoleUser,
		Content: "Fix the bug in auth.go",
	}
	tokens := EstimateMessageTokens(msg)
	if tokens <= 0 {
		t.Errorf("EstimateMessageTokens = %d; want > 0", tokens)
	}

	// 3. Watermark check
	if IsAboveWatermark(100, 1000, 0.80) {
		t.Errorf("IsAboveWatermark(100, 1000, 0.80) = true; want false")
	}
	if !IsAboveWatermark(850, 1000, 0.80) {
		t.Errorf("IsAboveWatermark(850, 1000, 0.80) = false; want true")
	}
}

func TestCompactor_FindSafeCutIndex_ToolPairing(t *testing.T) {
	cfg := CompactorConfig{
		MaxTokens:      1000,
		WatermarkRatio: 0.80,
		RetainRatio:    0.30, // 300 tokens retain budget
	}
	cp := NewCompactor(cfg, nil, "test-model")

	// Construct history with system prompt, old turns, and tool call/result pair
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "System instructions"},
		{Role: llm.RoleUser, Content: strings.Repeat("Long user task ", 20)},
		{Role: llm.RoleAssistant, Content: strings.Repeat("Long reasoning ", 20)},
		{Role: llm.RoleAssistant, Content: "Running tool", ToolCalls: []llm.ToolCall{{ID: "call_1", Function: llm.FunctionCall{Name: "bash", Arguments: "ls"}}}},
		{Role: llm.RoleTool, ToolCallID: "call_1", Content: "file1.go\nfile2.go"},
		{Role: llm.RoleAssistant, Content: "Recent thought"},
	}

	cutIndex := cp.findSafeCutIndex(messages, 20)
	if cutIndex < 1 || cutIndex >= len(messages) {
		t.Fatalf("unexpected cutIndex: %d", cutIndex)
	}

	// Ensure cut index does not slice between tool_call and tool result
	cutMsg := messages[cutIndex]
	if cutMsg.Role == llm.RoleTool {
		t.Errorf("cutIndex landed on RoleTool message (%d), breaking pairing", cutIndex)
	}
}

func TestCompactor_Compact_Success(t *testing.T) {
	mockSummary := `## Primary Request and Intent
- Fix bug in auth.go
## Key Technical Concepts
- Golang, OAuth
## Files and Code
- pkg/auth.go: fixed nil pointer
## Errors and Fixes
- nil pointer exception resolved
## Pending Jobs
- (none)
## Current Work
- Verification
## Next Step
- Run tests
## Critical Context
- All unit tests must pass`

	mock := &mockProvider{
		response: mockSummary,
	}

	cfg := CompactorConfig{
		MaxTokens:      500,
		WatermarkRatio: 0.50,
		RetainRatio:    0.20,
	}
	cp := NewCompactor(cfg, mock, "mock-llm")

	// Create conversation with large history
	conv := NewConversationContext("You are Mobius agent.")
	for i := 1; i <= 6; i++ {
		conv.AddUserMessage(fmt.Sprintf("User instruction step %d with lots of details %s", i, strings.Repeat("detail ", 30)))
		conv.AddAssistantMessage(fmt.Sprintf("Assistant reasoning step %d with lots of analysis %s", i, strings.Repeat("analysis ", 30)), nil)
	}

	initialCount := EstimateConversationTokens(conv.Messages())
	if !cp.ShouldCompact(initialCount) {
		t.Fatalf("expected ShouldCompact to be true, initial count: %d", initialCount)
	}

	err := cp.Compact(context.Background(), conv)
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}

	finalMessages := conv.Messages()
	finalCount := EstimateConversationTokens(finalMessages)

	// Check system prompt preserved
	if finalMessages[0].Role != llm.RoleSystem {
		t.Errorf("first message role = %s; want system", finalMessages[0].Role)
	}

	// Check second message is checkpoint
	if !strings.Contains(finalMessages[1].Content, SummaryOpenTag) {
		t.Errorf("second message does not contain %s", SummaryOpenTag)
	}

	// Check shrink
	if finalCount >= initialCount {
		t.Errorf("final count %d >= initial count %d (failed to shrink)", finalCount, initialCount)
	}
}
