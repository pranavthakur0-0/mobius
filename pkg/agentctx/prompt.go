package agentctx

import (
	"fmt"
)

const defaultSystemInstructions = `You are Mobius, an expert autonomous AI software engineering agent.
You solve coding and engineering tasks by planning step-by-step and executing tools.
Guidelines:
1. Always explore the codebase first (use view_file, grep_search, list_dir) before making changes.
2. When editing code, ensure your target_content matches the exact characters in the file.
3. After making changes, run tests or verify using run_command when appropriate.
4. When the task is fully accomplished, provide a concise final answer summarizing your solution.`

func BuildSystemPrompt(customInstructions string, guidesPrompt ...string) string {
	prompt := defaultSystemInstructions

	// If guides were passed in, append them
	if len(guidesPrompt) > 0 && guidesPrompt[0] != "" {
		prompt += guidesPrompt[0]
	}

	if customInstructions != "" {
		prompt += "\n\n# ADDITIONAL INSTRUCTIONS\n" + customInstructions
	}

	return prompt
}

func TitlePrompt(prompt string) string {
	return fmt.Sprintf("Generate a concise 3 to 4 word title for this prompt. Return ONLY the title with no quotes or punctuation:\n%s", prompt)
}
