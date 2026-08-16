package cli
import (
	"context"
	"fmt"
	"mobius/pkg/agent"
)
func RunGoal(a *agent.Agent, goal string) error {
	result, err := a.Run(context.Background(), goal)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	fmt.Println("\n" + result + "\n")
	return nil
}
