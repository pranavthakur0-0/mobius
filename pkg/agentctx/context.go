package agentctx

import (
	"mobius/pkg/llm"

)


type ConversationContext struct {
	messages []llm.Message
}


func NewConversationContext(systemPrompt string) *ConversationContext {
	ctx := &ConversationContext{
		messages: make([]llm.Message, 0),
	}

	if systemPrompt != "" {
		ctx.messages = append(ctx.messages, llm.Message{
			Role: llm.RoleSystem,
			Content: systemPrompt,
		})
	}
	return ctx
}

func (c *ConversationContext) AddUserMessage(content string) {
	c.messages = append(c.messages, llm.Message{
		Role: llm.RoleUser,
		Content: content,
	})
}


func (c *ConversationContext) AddAssistantMessage(content string, toolCalls []llm.ToolCall) {
	c.messages = append(c.messages, llm.Message{
		Role:      llm.RoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
	})
}

// AddToolResult appends the output of an executed tool.
func (c *ConversationContext) AddToolResult(toolCallID, output string) {
	c.messages = append(c.messages, llm.Message{
		Role:       llm.RoleTool,
		Content:    output,
		ToolCallID: toolCallID,
	})
}


func (c *ConversationContext) Messages() []llm.Message {
	return c.messages
}

// SetMessages replaces the active message slice (used during context compaction).
func (c *ConversationContext) SetMessages(messages []llm.Message) {
	c.messages = messages
}


