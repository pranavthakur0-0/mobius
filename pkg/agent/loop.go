package agent

import (
	"context"
	"fmt"
	contextPkg "mobius/pkg/context"
	"mobius/pkg/llm"
	"mobius/pkg/tracer"
)



func (a *Agent) Run(ctx context.Context, c *contextPkg.ConversationContext, userInstruction string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	c.AddUserMessage(userInstruction)

	fmt.Printf("[Goal] %s\n\n", userInstruction)


	for step := 1; step <= a.maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("agent interrupted: %w", err)
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

		c.AddAssistantMessage(resp.Content, resp.ToolCalls)
		if resp.Content != "" {
			fmt.Printf("[Thought]\n%s\n\n", resp.Content)
		}

		if len(resp.ToolCalls) == 0 {
			_ = tracer.AppendLog(tracer.LogEntry{
				ThreadID: a.threadID,
				Step: step,
				Thought: resp.Content,
			})
			return resp.Content, nil
		}

		for _, tc := range resp.ToolCalls {
			fmt.Printf("[Tool] %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
			tool, err := a.registry.Get(tc.Function.Name)
			var output string
			if err != nil {
				output = fmt.Sprintf("Error: tool '%s' not found", tc.Function.Name)
			} else {
				out, execErr := tool.Execute(ctx, tc.Function.Arguments)
				if execErr != nil {
					output = fmt.Sprintf("Tool error: %s\nOutput: %s", execErr, out)
				} else {
					output = out
				}
			}
			// Add tool observation to history
			c.AddToolResult(tc.ID, output)

			// Adding this tracer to get the log of what are we sending to llm 
			_ = tracer.AppendLog(tracer.LogEntry{
				ThreadID: a.threadID,
				Step: step,
				Thought: resp.Content,
				ToolName: tc.Function.Name,
				Output: output,
			})
		}
		
	}
	
	return "", fmt.Errorf("agent reached maximum step budget (%d steps)", a.maxSteps)

}