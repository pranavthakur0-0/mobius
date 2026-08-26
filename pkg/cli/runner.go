package cli

import (
	"context"
	"fmt"
	"mobius/pkg/session"
)

func RunGoal(s *session.Session, userInstruction string) error {
	ctx := context.Background()
	// If this is the session's first query, generate a title
	if !s.Started {
		s.Name = s.Agent.GenerateTitle(ctx, userInstruction)
		s.Started = true
	}
	result, err := s.Agent.Run(ctx, s.Context, userInstruction)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	fmt.Println("\n" + result + "\n")
	return nil
}
