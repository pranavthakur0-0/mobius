package sensors

import (
	"context"
	"os/exec"
	"time"
)

type CommandSensor struct {
    name string
    cmd  string
    dir  string
}

func NewCommandSensor(name, cmd, dir string) *CommandSensor {
    return &CommandSensor{
        name: name,
        cmd:  cmd,
        dir:  dir,
    }
}

func (s *CommandSensor) Name() string {
    return s.name
}


func (s *CommandSensor) Type() string {
    return TypeComputational
}


func (s *CommandSensor) Check(ctx context.Context) (Verdict, error) {
	start := time.Now()
	cmd := exec.CommandContext(ctx, "sh", "-c", s.cmd)
	cmd.Dir = s.dir

	out, err := cmd.CombinedOutput()
    duration := time.Since(start)
    // 3. Did it pass?
    if err != nil {
        return Verdict{
            Passed:    false,
            Output:    string(out),
            Retryable: true,
            Duration:  duration,
        }, nil
    }
    return Verdict{
        Passed:   true,
        Output:   string(out),
        Duration: duration,
    }, nil
}