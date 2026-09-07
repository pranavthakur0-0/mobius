package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"mobius/pkg/agentctx"
	"mobius/pkg/llm"
	"mobius/pkg/sensors"
	"mobius/pkg/tools"
)

type mockSensor struct {
	name   string
	passed bool
	output string
}

func (m *mockSensor) Name() string { return m.name }
func (m *mockSensor) Type() string { return sensors.TypeComputational }
func (m *mockSensor) Check(ctx context.Context) (sensors.Verdict, error) {
	return sensors.Verdict{
		SensorName: m.name,
		Passed:     m.passed,
		Output:     m.output,
		Duration:   5 * time.Millisecond,
	}, nil
}

type dummyTool struct {
	name string
}

func (d *dummyTool) Name() string                               { return d.name }
func (d *dummyTool) Description() string                        { return "dummy tool" }
func (d *dummyTool) Schema() tools.ToolSchema                  { return tools.ToolSchema{Type: "object"} }
func (d *dummyTool) Execute(ctx context.Context, args string) (string, error) {
	return "File edited successfully", nil
}

type mockProvider struct {
	turns []*llm.ChatResponse
	index int
}

func (m *mockProvider) Name() string { return "mock" }
func (m *mockProvider) Generate(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.index >= len(m.turns) {
		return &llm.ChatResponse{Content: "Done"}, nil
	}
	resp := m.turns[m.index]
	m.index++
	return resp, nil
}

func TestFormatSensorFailures(t *testing.T) {
	verdicts := []sensors.Verdict{
		{
			SensorName: "build",
			Passed:     false,
			Output:     "main.go:10: undefined: foo",
		},
	}

	formatted := formatSensorFailures(verdicts)
	if !strings.Contains(formatted, "[Automatic Sensor Check Failed]") {
		t.Errorf("expected header in %s", formatted)
	}
	if !strings.Contains(formatted, "Sensor: build") {
		t.Errorf("expected sensor name in %s", formatted)
	}
	if !strings.Contains(formatted, "main.go:10: undefined: foo") {
		t.Errorf("expected compiler error in %s", formatted)
	}
}

type mutableSensor struct {
	name      string
	callCount int
}

func (m *mutableSensor) Name() string { return m.name }
func (m *mutableSensor) Type() string { return sensors.TypeComputational }
func (m *mutableSensor) Check(ctx context.Context) (sensors.Verdict, error) {
	m.callCount++
	if m.callCount == 1 {
		return sensors.Verdict{
			SensorName: m.name,
			Passed:     false,
			Output:     "compile error: expected ';'",
		}, nil
	}
	return sensors.Verdict{
		SensorName: m.name,
		Passed:     true,
		Output:     "build successful",
	}, nil
}

func TestSelfHealingLoop_SensorFailureInjected(t *testing.T) {
	// Turn 1: Model introduces buggy edit -> Sensor fails
	// Turn 2: Model receives compiler error, edits file to fix it -> Sensor passes
	// Turn 3: Model concludes with final message
	mockP := &mockProvider{
		turns: []*llm.ChatResponse{
			{
				Content: "Making initial change...",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "edit_file",
							Arguments: `{"path": "main.go"}`,
						},
					},
				},
			},
			{
				Content: "Fixing compiler syntax error...",
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_2",
						Type: "function",
						Function: llm.FunctionCall{
							Name:      "edit_file",
							Arguments: `{"path": "main.go"}`,
						},
					},
				},
			},
			{
				Content: "I have fixed the issue and tests pass.",
			},
		},
	}

	reg := tools.NewRegistry()
	reg.Register(&dummyTool{name: "edit_file"})

	ag := &Agent{
		threadID: "test-thread",
		provider: mockP,
		registry: reg,
		model:    "test-model",
		maxSteps: 5,
		timeout:  5 * time.Second,
	}

	sensorReg := sensors.NewRegistry()
	_ = sensorReg.Register(&mutableSensor{name: "build"})
	ag.SetSensors(sensorReg)

	ctx := context.Background()
	conv := agentctx.NewConversationContext("system prompt")

	res, err := ag.Run(ctx, conv, "Fix the bug")
	if err != nil {
		t.Fatalf("expected run to succeed, got error: %v", err)
	}

	if res != "I have fixed the issue and tests pass." {
		t.Errorf("unexpected final result: %s", res)
	}

	// Verify that the conversation history contains the sensor failure injected message
	foundSensorFailure := false
	for _, m := range conv.Messages() {
		if strings.Contains(m.Content, "[Automatic Sensor Check Failed]") && strings.Contains(m.Content, "compile error: expected ';'") {
			foundSensorFailure = true
			break
		}
	}

	if !foundSensorFailure {
		t.Errorf("expected sensor failure message to be injected into conversation history")
	}
}
