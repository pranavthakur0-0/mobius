package sensors

import (
	"context"
	"time"
)

const (
	TypeComputational = "computational"
	TypeInferential   = "inferential"
)

type Verdict struct {
	Passed    bool
	Output    string
	Retryable bool
	Duration  time.Duration
}

type Sensor interface {
	Name() string
	Type() string
	Check(ctx context.Context) (Verdict, error)
}


