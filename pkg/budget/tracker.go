package budget

import (
	"fmt"
	"mobius/pkg/llm"
)

const USDToINRRate = 95

type CostTracker interface {
	Add(u llm.Usage)
	Check() error
	Status() string
	CurrentCostUSD() float64
	UpdatePrices(pricePrompt, priceComp float64)
}

// Tracker keeps a running total of tokens and calculates USD costs.
type Tracker struct {
	TotalPromptTokens     int
	TotalCompletionTokens int
	MaxCost               float64

	PricePerMillionPrompt     float64
	PricePerMillionCompletion float64
}

// NewTracker initializes a fresh tracker starting at 0 tokens.
func NewTracker(maxCost, pricePrompt, priceComp float64) *Tracker {
	return &Tracker{
		MaxCost:                   maxCost,
		PricePerMillionPrompt:     pricePrompt,
		PricePerMillionCompletion: priceComp,
	}
}

// Add tallies up the tokens from a single LLM generation.
func (t *Tracker) Add(u llm.Usage) {
	t.TotalPromptTokens += u.PromptTokens
	t.TotalCompletionTokens += u.CompletionTokens
}

// CurrentCostUSD calculates the running total spend in USD.
func (t *Tracker) CurrentCostUSD() float64 {
	promptCost := (float64(t.TotalPromptTokens) / 1_000_000.0) * t.PricePerMillionPrompt
	compCost := (float64(t.TotalCompletionTokens) / 1_000_000.0) * t.PricePerMillionCompletion
	return promptCost + compCost
}

// CurrentCostINR converts the current USD spend into Indian Rupees.
func (t *Tracker) CurrentCostINR() float64 {
	return t.CurrentCostUSD() * USDToINRRate
}

// Check aborts if we've spent more than the max allowed cost.
func (t *Tracker) Check() error {
	cost := t.CurrentCostUSD()
	if cost > t.MaxCost {
		return fmt.Errorf("budget exceeded: current cost $%.4f (₹%.2f) > max cost $%.2f", cost, t.CurrentCostINR(), t.MaxCost)
	}
	return nil
}

// Status formats a summary for the terminal showing both USD and INR.
func (t *Tracker) Status() string {
	return fmt.Sprintf("[Budget] Tokens: %d In, %d Out | Cost: $%.5f (₹%.4f)",
		t.TotalPromptTokens, t.TotalCompletionTokens, t.CurrentCostUSD(), t.CurrentCostINR())
}

func (t *Tracker) UpdatePrices(pricePrompt, priceComp float64) {
	t.PricePerMillionPrompt = pricePrompt
	t.PricePerMillionCompletion = priceComp
}
