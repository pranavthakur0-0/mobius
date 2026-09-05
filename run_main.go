package main

import (
	"fmt"
	"mobius/pkg/guides"
)

func main() {
	fmt.Println("=== RUNNING MOBIUS TESTBENCH ===")
	gs, err := guides.LoadFromWorkSpace("config/guides.toml", ".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 🔍 Print the exact markdown prompt generated from your workspace:
	fmt.Println(gs.RenderSystemPrompt())
}
