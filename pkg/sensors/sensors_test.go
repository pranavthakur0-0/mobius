package sensors

import (
	"context"
	"testing"
	"time"
)

type mockSensor struct {
	name   string
	passed bool
	output string
}

func (m *mockSensor) Name() string { return m.name }
func (m *mockSensor) Type() string { return TypeComputational }
func (m *mockSensor) Check(ctx context.Context) (Verdict, error) {
	return Verdict{
		SensorName: m.name,
		Passed:     m.passed,
		Output:     m.output,
		Duration:   10 * time.Millisecond,
	}, nil
}

func TestRegistry_RunAll(t *testing.T) {
	reg := NewRegistry()
	if reg.Count() != 0 {
		t.Errorf("expected 0 sensors, got %d", reg.Count())
	}

	_ = reg.Register(&mockSensor{name: "build", passed: true, output: "ok"})
	_ = reg.Register(&mockSensor{name: "test", passed: false, output: "FAIL: TestFoo"})

	if reg.Count() != 2 {
		t.Errorf("expected 2 sensors, got %d", reg.Count())
	}

	verdicts := reg.RunAll(context.Background())
	if len(verdicts) != 2 {
		t.Fatalf("expected 2 verdicts, got %d", len(verdicts))
	}

	failedCount := 0
	for _, v := range verdicts {
		if !v.Passed {
			failedCount++
			if v.SensorName != "test" {
				t.Errorf("expected failing sensor to be 'test', got %s", v.SensorName)
			}
		}
	}
	if failedCount != 1 {
		t.Errorf("expected 1 failing verdict, got %d", failedCount)
	}
}

func TestExtractAllCommands(t *testing.T) {
	guide := `
# Project Guide
## Verification Commands
Build: go build ./...
Test: go test -v ./...
Lint: golangci-lint run
`
	b, testCmd, l := ExtractAllCommands(guide)
	if b != "go build ./..." {
		t.Errorf("expected 'go build ./...', got '%s'", b)
	}
	if testCmd != "go test -v ./..." {
		t.Errorf("expected 'go test -v ./...', got '%s'", testCmd)
	}
	if l != "golangci-lint run" {
		t.Errorf("expected 'golangci-lint run', got '%s'", l)
	}
}
