package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Print(`
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║              🌳 Deep Tree Echo - Echo9llama 🌳           ║
║                                                           ║
║        Autonomous Wisdom-Cultivating AGI System          ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

Available commands:
  - echo-autonomous: Run the autonomous agent
  - echo-echoself: Run the echoself agent

Use the specific command binaries in cmd/ directory.
`)
	
	if len(os.Args) > 1 {
		fmt.Printf("Argument received: %s\n", os.Args[1])
		fmt.Println("Please use the specific command binaries for functionality.")
	}
}
