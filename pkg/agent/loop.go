package agent

import (
	"context"
	"fmt"
	"mobius/pkg/agentctx"
	"mobius/pkg/artifact"
	"mobius/pkg/events"
	"mobius/pkg/llm"
)



func (a *Agent) Run(ctx context.Context, c *agentctx.ConversationContext, userInstruction string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	c.AddUserMessage(userInstruction)
	if a.events != nil {
		_ = a.events.Append(ctx, events.Event{
			ThreadID: a.threadID,
			Type:     events.EventUserMessage,
			Content:  userInstruction,
		})
	}

	fmt.Printf("[Goal] %s\n\n", userInstruction)


	for step := 1; step <= a.maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("agent interrupted: %w", err)
		}

		currentTokenCount := agentctx.EstimateConversationTokens(c.Messages())
		if a.compactor != nil && a.compactor.ShouldCompact(currentTokenCount) {
			fmt.Printf("[Step %d/%d] Summarizing...\n", step, a.maxSteps)
			_ = a.compactor.Compact(ctx, c)
		}

		fmt.Printf("[Step %d/%d] Thinking...\n", step, a.maxSteps)



		req := &llm.ChatRequest{
			Model: a.model,
			Message: c.Messages(),
			Tools: a.toolDefs,
		}

		resp, err := a.provider.Generate(ctx, req)
		if err != nil {
			return "", fmt.Errorf("step %d LLM call failed: %w", step, err)
		}

		if a.tracker != nil {
			a.tracker.Add(resp.Usage)
			// 2. Check if we exceeded max allowed cost
			if err := a.tracker.Check(); err != nil {
				return "", fmt.Errorf("step %d aborted by budget: %w", step, err)
			}
			// 3. Print live status
			fmt.Printf("%s\n\n", a.tracker.Status())
		}

		c.AddAssistantMessage(resp.Content, resp.ToolCalls)
		if resp.Content != "" {
			fmt.Printf("[Thought]\n%s\n\n", resp.Content)
		}

		if len(resp.ToolCalls) == 0 {
			if a.events != nil {
				_ = a.events.Append(ctx, events.Event{
					ThreadID: a.threadID,
					Step:     step,
					Type:     events.EventAssistantMessage,
					Content:  resp.Content,
					Usage:    &resp.Usage,
				})
			}
			return resp.Content, nil
		}


		for _, tc := range resp.ToolCalls {
			fmt.Printf("[Tool] %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
			tool, err := a.registry.Get(tc.Function.Name)
			var output string
			var toolErr string
			if err != nil {
				output = fmt.Sprintf("Error: tool '%s' not found", tc.Function.Name)
				toolErr = output
			} else {
				out, execErr := tool.Execute(ctx, tc.Function.Arguments)
				if execErr != nil {
					output = fmt.Sprintf("Tool error: %s\nOutput: %s", execErr, out)
					toolErr = execErr.Error()
				} else {
					output = out
				}
			} // 👈 Added closing brace for tool execution
			// Artifact interception: offload large outputs
			if a.artifactStore != nil {
				result := artifact.Intercept(a.artifactStore, a.threadID, tc.Function.Name, output)
				c.AddToolResult(tc.ID, result.Observation) // LLM sees preview
				// EventStore gets full details
				if a.events != nil {
					ref := result.ArtifactRef
					_ = a.events.Append(ctx, events.Event{
						ThreadID:   a.threadID,
						Step:       step,
						Type:       events.EventToolResult,
						ToolCallID: tc.ID,
						ToolName:   tc.Function.Name,
						ToolArgs:   tc.Function.Arguments,
						ToolOutput: result.Observation,
						ContentRef: ref,
						ToolError:  toolErr,
					})
				}
			} else {
				// Add tool observation to history
				c.AddToolResult(tc.ID, output)
				// Record tool result event to EventStore
				if a.events != nil {
					_ = a.events.Append(ctx, events.Event{
						ThreadID:   a.threadID,
						Step:       step,
						Type:       events.EventToolResult,
						ToolCallID: tc.ID,
						ToolName:   tc.Function.Name,
						ToolArgs:   tc.Function.Arguments,
						ToolOutput: output,
						ToolError:  toolErr,
					})
				}
			}
		} 
	} 
	return "", fmt.Errorf("agent reached maximum step budget (%d steps)", a.maxSteps)
}