package agent

import (
	"context"
	"fmt"
	"mobius/pkg/agentctx"
	"mobius/pkg/artifact"
	"mobius/pkg/events"
	"mobius/pkg/llm"
	"mobius/pkg/sensors"
	"strings"
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

	consecutiveSensorFailures := 0
	const maxSensorRetries = 3

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
			Model:   a.model,
			Message: c.Messages(),
			Tools:   a.toolDefs,
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
			// If verification checks are active and we had sensor failures, guard against premature completion
			if a.sensorRegistry != nil && a.sensorRegistry.Count() > 0 && consecutiveSensorFailures > 0 {
				fmt.Printf("[Sensor] Validating before completion...\n")
				verdicts := a.sensorRegistry.RunAll(ctx)
				var stillFailing []sensors.Verdict
				for _, v := range verdicts {
					if !v.Passed {
						stillFailing = append(stillFailing, v)
					}
				}
				if len(stillFailing) > 0 {
					consecutiveSensorFailures++
					if consecutiveSensorFailures <= maxSensorRetries {
						sensorMsg := "[Verification Incomplete] You cannot conclude the task because verification checks are currently failing:\n\n" + formatSensorFailures(stillFailing)
						fmt.Printf("%s\n\n", sensorMsg)
						c.AddUserMessage(sensorMsg)
						continue
					}
				}
			}

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

		fileModified := false
		for _, tc := range resp.ToolCalls {
			if tc.Function.Name == "write_file" || tc.Function.Name == "edit_file" {
				fileModified = true
			}
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
			}
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

		// Self-healing sensor verification
		if fileModified && a.sensorRegistry != nil && a.sensorRegistry.Count() > 0 {
			fmt.Printf("[Sensor] Running verification checks...\n")
			verdicts := a.sensorRegistry.RunAll(ctx)
			var failed []sensors.Verdict
			for _, v := range verdicts {
				if !v.Passed {
					failed = append(failed, v)
				}
			}

			if len(failed) > 0 {
				consecutiveSensorFailures++
				sensorMsg := formatSensorFailures(failed)
				fmt.Printf("%s\n\n", sensorMsg)

				if consecutiveSensorFailures >= maxSensorRetries {
					sensorMsg += fmt.Sprintf("\n\n[Warning] Sensor check has failed %d consecutive times. Please carefully rethink your changes.", consecutiveSensorFailures)
				}

				c.AddUserMessage(sensorMsg)

				if a.events != nil {
					_ = a.events.Append(ctx, events.Event{
						ThreadID: a.threadID,
						Step:     step,
						Type:     events.EventSensorResult,
						Content:  sensorMsg,
					})
				}
			} else {
				consecutiveSensorFailures = 0
				fmt.Printf("[Sensor] All %d verification checks passed.\n\n", len(verdicts))
				if a.events != nil {
					_ = a.events.Append(ctx, events.Event{
						ThreadID: a.threadID,
						Step:     step,
						Type:     events.EventSensorResult,
						Content:  fmt.Sprintf("All %d verification checks passed.", len(verdicts)),
					})
				}
			}
		}
	}
	return "", fmt.Errorf("agent reached maximum step budget (%d steps)", a.maxSteps)
}

func formatSensorFailures(failed []sensors.Verdict) string {
	var sb strings.Builder
	sb.WriteString("[Automatic Sensor Check Failed]\n")
	for _, v := range failed {
		name := v.SensorName
		if name == "" {
			name = "verification"
		}
		sb.WriteString(fmt.Sprintf("Sensor: %s\nOutput:\n%s\n", name, strings.TrimSpace(v.Output)))
	}
	sb.WriteString("\nThe code did not compile or pass verification. Please inspect the compiler/test error above and fix the issue before proceeding.")
	return sb.String()
}
