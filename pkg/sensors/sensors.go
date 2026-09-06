package sensors

import (
	"context"
	"fmt"
)

type Registry struct {
	sensor map[string]Sensor
}

func NewRegistry() *Registry {
	return &Registry{
		sensor: make(map[string]Sensor),
	}
}

func (r *Registry) Register(s Sensor) error {
	if _, exists := r.sensor[s.Name()]; exists {
		return fmt.Errorf("Sensor already loaded")
	}
	r.sensor[s.Name()] = s
	return nil
}

func (r *Registry) RunAll(ctx context.Context) []Verdict {
	var verdicts []Verdict
	for _, s := range r.sensor {
		v, err := s.Check(ctx)
		if err != nil {
			verdicts = append(verdicts, Verdict{
				Passed: false,
				Output: fmt.Sprintf("%s error: %v", s.Name(), err),
			})
		} else {
			verdicts = append(verdicts, v)
		}
	}
	return verdicts
}
