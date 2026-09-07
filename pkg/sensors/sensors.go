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

func (r *Registry) Count() int {
	if r == nil {
		return 0
	}
	return len(r.sensor)
}

func (r *Registry) RunAll(ctx context.Context) []Verdict {
	if r == nil {
		return nil
	}
	var verdicts []Verdict
	for _, s := range r.sensor {
		v, err := s.Check(ctx)
		if err != nil {
			verdicts = append(verdicts, Verdict{
				SensorName: s.Name(),
				Passed:     false,
				Output:     fmt.Sprintf("%s error: %v", s.Name(), err),
				Retryable:  true,
			})
		} else {
			if v.SensorName == "" {
				v.SensorName = s.Name()
			}
			verdicts = append(verdicts, v)
		}
	}
	return verdicts
}
