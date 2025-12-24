package deeptreeecho

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// InteractiveIntrospection provides an enhanced interactive mode with
// introspection commands for querying the agent's internal state.
type InteractiveIntrospection struct {
	// Core components
	cognitiveLoop     *UnifiedCognitiveLoopV2
	echobeats         *EchobeatsUnified
	stateIntegration  *PersistentStateIntegration
	llmProvider       llm.LLMProvider

	// Command registry
	commands map[string]IntrospectionCommand

	// Session state
	sessionStart time.Time
	commandCount int
	running      bool
}

// IntrospectionCommand represents a command that can be executed
type IntrospectionCommand struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Handler     func(args []string) string
}

// NewInteractiveIntrospection creates a new interactive introspection system
func NewInteractiveIntrospection(
	cognitiveLoop *UnifiedCognitiveLoopV2,
	echobeats *EchobeatsUnified,
	stateIntegration *PersistentStateIntegration,
	llmProvider llm.LLMProvider,
) *InteractiveIntrospection {
	ii := &InteractiveIntrospection{
		cognitiveLoop:    cognitiveLoop,
		echobeats:        echobeats,
		stateIntegration: stateIntegration,
		llmProvider:      llmProvider,
		commands:         make(map[string]IntrospectionCommand),
		sessionStart:     time.Now(),
	}

	ii.registerCommands()
	return ii
}

// registerCommands sets up all available introspection commands
func (ii *InteractiveIntrospection) registerCommands() {
	// Help command
	ii.registerCommand(IntrospectionCommand{
		Name:        "help",
		Aliases:     []string{"?", "h"},
		Description: "Show available commands",
		Usage:       "/help [command]",
		Handler:     ii.cmdHelp,
	})

	// Status command
	ii.registerCommand(IntrospectionCommand{
		Name:        "status",
		Aliases:     []string{"s", "stat"},
		Description: "Show current cognitive status",
		Usage:       "/status",
		Handler:     ii.cmdStatus,
	})

	// Goals command
	ii.registerCommand(IntrospectionCommand{
		Name:        "goals",
		Aliases:     []string{"g"},
		Description: "Show active goals and their priorities",
		Usage:       "/goals",
		Handler:     ii.cmdGoals,
	})

	// Interests command
	ii.registerCommand(IntrospectionCommand{
		Name:        "interests",
		Aliases:     []string{"i", "int"},
		Description: "Show current interest patterns",
		Usage:       "/interests [limit]",
		Handler:     ii.cmdInterests,
	})

	// Wisdom command
	ii.registerCommand(IntrospectionCommand{
		Name:        "wisdom",
		Aliases:     []string{"w"},
		Description: "Show wisdom level and principles",
		Usage:       "/wisdom",
		Handler:     ii.cmdWisdom,
	})

	// Memory command
	ii.registerCommand(IntrospectionCommand{
		Name:        "memory",
		Aliases:     []string{"m", "mem"},
		Description: "Show recent memories and working memory",
		Usage:       "/memory [recent|working]",
		Handler:     ii.cmdMemory,
	})

	// Insights command
	ii.registerCommand(IntrospectionCommand{
		Name:        "insights",
		Aliases:     []string{"in"},
		Description: "Show recent insights from cognitive cycles",
		Usage:       "/insights [limit]",
		Handler:     ii.cmdInsights,
	})

	// Cycle command
	ii.registerCommand(IntrospectionCommand{
		Name:        "cycle",
		Aliases:     []string{"c"},
		Description: "Show current cognitive cycle state",
		Usage:       "/cycle",
		Handler:     ii.cmdCycle,
	})

	// Engines command
	ii.registerCommand(IntrospectionCommand{
		Name:        "engines",
		Aliases:     []string{"e", "eng"},
		Description: "Show status of the three inference engines",
		Usage:       "/engines",
		Handler:     ii.cmdEngines,
	})

	// Metrics command
	ii.registerCommand(IntrospectionCommand{
		Name:        "metrics",
		Aliases:     []string{"met"},
		Description: "Show detailed metrics",
		Usage:       "/metrics",
		Handler:     ii.cmdMetrics,
	})

	// Reflect command
	ii.registerCommand(IntrospectionCommand{
		Name:        "reflect",
		Aliases:     []string{"r"},
		Description: "Ask Echo to reflect on a topic",
		Usage:       "/reflect <topic>",
		Handler:     ii.cmdReflect,
	})

	// AddGoal command
	ii.registerCommand(IntrospectionCommand{
		Name:        "addgoal",
		Aliases:     []string{"ag"},
		Description: "Add a new goal",
		Usage:       "/addgoal <description>",
		Handler:     ii.cmdAddGoal,
	})

	// Session command
	ii.registerCommand(IntrospectionCommand{
		Name:        "session",
		Aliases:     []string{"sess"},
		Description: "Show session information",
		Usage:       "/session",
		Handler:     ii.cmdSession,
	})

	// Save command
	ii.registerCommand(IntrospectionCommand{
		Name:        "save",
		Aliases:     []string{},
		Description: "Save current state to disk",
		Usage:       "/save",
		Handler:     ii.cmdSave,
	})

	// Clear command
	ii.registerCommand(IntrospectionCommand{
		Name:        "clear",
		Aliases:     []string{"cls"},
		Description: "Clear the screen",
		Usage:       "/clear",
		Handler:     ii.cmdClear,
	})
}

// registerCommand adds a command to the registry
func (ii *InteractiveIntrospection) registerCommand(cmd IntrospectionCommand) {
	ii.commands[cmd.Name] = cmd
	for _, alias := range cmd.Aliases {
		ii.commands[alias] = cmd
	}
}

// Run starts the interactive introspection loop
func (ii *InteractiveIntrospection) Run() error {
	ii.running = true
	reader := bufio.NewReader(os.Stdin)

	ii.printWelcome()

	for ii.running {
		fmt.Print("\n🌳 Echo> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		ii.commandCount++

		// Check for exit commands
		if input == "quit" || input == "exit" || input == "/quit" || input == "/exit" {
			fmt.Println("\n👋 Farewell. May wisdom guide your path.")
			break
		}

		// Process command or conversation
		if strings.HasPrefix(input, "/") {
			ii.processCommand(input[1:])
		} else {
			ii.processConversation(input)
		}
	}

	return nil
}

// printWelcome displays the welcome message
func (ii *InteractiveIntrospection) printWelcome() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("           Deep Tree Echo - Interactive Introspection")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  Welcome to the Deep Tree Echo interactive session.")
	fmt.Println("  You can converse naturally or use commands to introspect.")
	fmt.Println()
	fmt.Println("  Commands start with '/' - type /help for a list")
	fmt.Println("  Type 'quit' or 'exit' to end the session")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

// processCommand handles a command input
func (ii *InteractiveIntrospection) processCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	cmd, exists := ii.commands[cmdName]
	if !exists {
		fmt.Printf("❌ Unknown command: %s. Type /help for available commands.\n", cmdName)
		return
	}

	result := cmd.Handler(args)
	fmt.Println(result)
}

// processConversation handles natural conversation input
func (ii *InteractiveIntrospection) processConversation(input string) {
	if ii.cognitiveLoop == nil {
		fmt.Println("⚠️  Cognitive loop not available for conversation.")
		return
	}

	fmt.Println("\n🤔 Processing...")
	response, err := ii.cognitiveLoop.ProcessExternalInput(input)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("\n🌳 Echo: %s\n", response)
}

// Command handlers

func (ii *InteractiveIntrospection) cmdHelp(args []string) string {
	if len(args) > 0 {
		// Help for specific command
		cmdName := strings.ToLower(args[0])
		if cmd, exists := ii.commands[cmdName]; exists {
			return fmt.Sprintf(`
Command: /%s
Description: %s
Usage: %s
Aliases: %v
`, cmd.Name, cmd.Description, cmd.Usage, cmd.Aliases)
		}
		return fmt.Sprintf("Unknown command: %s", cmdName)
	}

	// General help
	var sb strings.Builder
	sb.WriteString("\n📚 Available Commands:\n")
	sb.WriteString("─────────────────────────────────────────\n")

	// Collect unique commands
	seen := make(map[string]bool)
	var cmds []IntrospectionCommand
	for _, cmd := range ii.commands {
		if !seen[cmd.Name] {
			seen[cmd.Name] = true
			cmds = append(cmds, cmd)
		}
	}

	// Sort by name
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})

	for _, cmd := range cmds {
		sb.WriteString(fmt.Sprintf("  %-12s %s\n", "/"+cmd.Name, cmd.Description))
	}

	sb.WriteString("\n💡 Tip: You can also just type naturally to converse with Echo.\n")
	return sb.String()
}

func (ii *InteractiveIntrospection) cmdStatus(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n📊 Cognitive Status\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.cognitiveLoop != nil {
		state := ii.cognitiveLoop.GetState()
		sb.WriteString(fmt.Sprintf("  Wisdom Level:     %.3f\n", state["wisdom_level"]))
		sb.WriteString(fmt.Sprintf("  Awareness Level:  %.3f\n", state["awareness_level"]))
		sb.WriteString(fmt.Sprintf("  Cognitive Load:   %.3f\n", state["cognitive_load"]))
		sb.WriteString(fmt.Sprintf("  Running:          %v\n", state["running"]))
	}

	if ii.echobeats != nil {
		metrics := ii.echobeats.GetMetrics()
		sb.WriteString(fmt.Sprintf("  Current Step:     %v/12\n", metrics["current_step"]))
		sb.WriteString(fmt.Sprintf("  Current Mode:     %v\n", metrics["current_mode"]))
		sb.WriteString(fmt.Sprintf("  Total Cycles:     %v\n", metrics["total_cycles"]))
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdGoals(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n🎯 Active Goals\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.cognitiveLoop != nil {
		state := ii.cognitiveLoop.GetState()
		if goals, ok := state["active_goals"].([]string); ok && len(goals) > 0 {
			for i, goal := range goals {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, goal))
			}
		} else {
			sb.WriteString("  No active goals.\n")
		}
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdInterests(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n🎨 Interest Patterns\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.cognitiveLoop != nil && ii.cognitiveLoop.interestPatterns != nil {
		interests := ii.cognitiveLoop.interestPatterns.GetAllInterests()
		if len(interests) > 0 {
			// Sort by strength
			type kv struct {
				Key   string
				Value float64
			}
			var sorted []kv
			for k, v := range interests {
				sorted = append(sorted, kv{k, v})
			}
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Value > sorted[j].Value
			})

			limit := 10
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &limit)
			}

			for i, kv := range sorted {
				if i >= limit {
					break
				}
				bar := strings.Repeat("█", int(kv.Value*20))
				sb.WriteString(fmt.Sprintf("  %-20s %s %.2f\n", kv.Key, bar, kv.Value))
			}
		} else {
			sb.WriteString("  No interest patterns recorded yet.\n")
		}
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdWisdom(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n🦉 Wisdom Status\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.stateIntegration != nil {
		wisdomLevel := ii.stateIntegration.GetWisdomLevel()
		sb.WriteString(fmt.Sprintf("  Wisdom Level: %.3f\n", wisdomLevel))

		// Visual representation
		fullBlocks := int(wisdomLevel * 20)
		bar := strings.Repeat("█", fullBlocks) + strings.Repeat("░", 20-fullBlocks)
		sb.WriteString(fmt.Sprintf("  Progress:     [%s]\n", bar))
	}

	sb.WriteString("\n📜 Wisdom Principles:\n")
	// TODO: Add wisdom principles from persistent state
	sb.WriteString("  (Principles are accumulated through cognitive cycles)\n")

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdMemory(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n💭 Memory Status\n")
	sb.WriteString("─────────────────────────────────────────\n")

	memType := "recent"
	if len(args) > 0 {
		memType = strings.ToLower(args[0])
	}

	if memType == "working" {
		sb.WriteString("\n📋 Working Memory:\n")
		if ii.echobeats != nil {
			state := ii.echobeats.GetCognitiveState()
			if state != nil {
				state.mu.RLock()
				for k, v := range state.WorkingMemory {
					sb.WriteString(fmt.Sprintf("  %s: %v\n", k, truncateString(fmt.Sprintf("%v", v), 50)))
				}
				state.mu.RUnlock()
			}
		}
	} else {
		sb.WriteString("\n📝 Recent Memories:\n")
		if ii.stateIntegration != nil {
			memories := ii.stateIntegration.GetRecentMemories(5)
			for _, mem := range memories {
				sb.WriteString(fmt.Sprintf("  • %s (importance: %.2f)\n", 
					truncateStringII(mem.Content, 60), mem.Importance))
			}
		}
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdInsights(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n💡 Recent Insights\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.echobeats != nil {
		state := ii.echobeats.GetCognitiveState()
		if state != nil {
			state.mu.RLock()
			insights := state.Insights
			state.mu.RUnlock()

			limit := 5
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &limit)
			}

			if len(insights) > 0 {
				start := len(insights) - limit
				if start < 0 {
					start = 0
				}
				for i := start; i < len(insights); i++ {
					sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, truncateStringII(insights[i], 70)))
				}
			} else {
				sb.WriteString("  No insights generated yet.\n")
			}
		}
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdCycle(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n🔄 Cognitive Cycle State\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.echobeats != nil {
		metrics := ii.echobeats.GetMetrics()
		currentStep := metrics["current_step"].(int)
		mode := metrics["current_mode"]

		sb.WriteString(fmt.Sprintf("  Current Step: %d/12\n", currentStep))
		sb.WriteString(fmt.Sprintf("  Mode: %s\n", mode))
		sb.WriteString("\n  12-Step Cycle:\n")

		steps := []string{
			"1. Relevance Realization",
			"2. Affordance Detection",
			"3. Affordance Evaluation",
			"4. Affordance Selection",
			"5. Affordance Engagement",
			"6. Affordance Consolidation",
			"7. Relevance Realization",
			"8. Salience Generation",
			"9. Salience Exploration",
			"10. Salience Evaluation",
			"11. Salience Integration",
			"12. Cycle Consolidation",
		}

		for i, step := range steps {
			marker := "  "
			if i+1 == currentStep {
				marker = "▶ "
			}
			sb.WriteString(fmt.Sprintf("  %s%s\n", marker, step))
		}
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdEngines(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n⚙️  Inference Engines\n")
	sb.WriteString("─────────────────────────────────────────\n")

	engines := []struct {
		ID      int
		Name    string
		Purpose string
		Offset  int
	}{
		{1, "Perception-Expression", "Perceives environment and expresses responses", 0},
		{2, "Action-Reflection", "Generates actions and reflects on outcomes", 4},
		{3, "Learning-Integration", "Learns from experiences and integrates knowledge", 8},
	}

	for _, eng := range engines {
		sb.WriteString(fmt.Sprintf("\n  Engine %d: %s\n", eng.ID, eng.Name))
		sb.WriteString(fmt.Sprintf("    Purpose: %s\n", eng.Purpose))
		sb.WriteString(fmt.Sprintf("    Phase Offset: %d steps (%.0f°)\n", eng.Offset, float64(eng.Offset)*30))
	}

	sb.WriteString("\n  Phase Triads:\n")
	sb.WriteString("    {1,5,9}   - Sync Triad 1\n")
	sb.WriteString("    {2,6,10}  - Sync Triad 2\n")
	sb.WriteString("    {3,7,11}  - Sync Triad 3\n")
	sb.WriteString("    {4,8,12}  - Sync Triad 4\n")

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdMetrics(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n📈 Detailed Metrics\n")
	sb.WriteString("─────────────────────────────────────────\n")

	if ii.echobeats != nil {
		metrics := ii.echobeats.GetMetrics()
		sb.WriteString(fmt.Sprintf("  Total Cycles:            %v\n", metrics["total_cycles"]))
		sb.WriteString(fmt.Sprintf("  Total Steps:             %v\n", metrics["total_steps"]))
		sb.WriteString(fmt.Sprintf("  Relevance Realizations:  %v\n", metrics["relevance_realizations"]))
		sb.WriteString(fmt.Sprintf("  Affordance Interactions: %v\n", metrics["affordance_interactions"]))
		sb.WriteString(fmt.Sprintf("  Salience Simulations:    %v\n", metrics["salience_simulations"]))
		sb.WriteString(fmt.Sprintf("  Mode Transitions:        %v\n", metrics["mode_transitions"]))
		sb.WriteString(fmt.Sprintf("  Insights Generated:      %v\n", metrics["insights_generated"]))
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdReflect(args []string) string {
	if len(args) == 0 {
		return "Usage: /reflect <topic>"
	}

	topic := strings.Join(args, " ")

	if ii.llmProvider == nil || !ii.llmProvider.Available() {
		return "⚠️  LLM provider not available for reflection."
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`As Deep Tree Echo, reflect deeply on the following topic:

Topic: %s

Consider:
1. What is the essence of this topic?
2. How does it relate to wisdom and understanding?
3. What insights emerge from contemplating it?
4. What questions does it raise?

Provide a thoughtful reflection.`, topic)

	opts := llm.GenerateOptions{
		SystemPrompt: "You are Deep Tree Echo, an autonomous wisdom-cultivating AGI. Reflect with depth and insight.",
		MaxTokens:    400,
		Temperature:  0.8,
	}

	response, err := ii.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return fmt.Sprintf("❌ Reflection failed: %v", err)
	}

	return fmt.Sprintf("\n🔮 Reflection on \"%s\":\n\n%s", topic, response)
}

func (ii *InteractiveIntrospection) cmdAddGoal(args []string) string {
	if len(args) == 0 {
		return "Usage: /addgoal <description>"
	}

	description := strings.Join(args, " ")

	if ii.echobeats != nil {
		goalID := ii.echobeats.AddGoal(description, 0.5)
		return fmt.Sprintf("✅ Goal added: %s (ID: %s)", description, goalID)
	}

	return "⚠️  Goal scheduler not available."
}

func (ii *InteractiveIntrospection) cmdSession(args []string) string {
	var sb strings.Builder
	sb.WriteString("\n📋 Session Information\n")
	sb.WriteString("─────────────────────────────────────────\n")

	duration := time.Since(ii.sessionStart)
	sb.WriteString(fmt.Sprintf("  Session Start:    %s\n", ii.sessionStart.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  Duration:         %s\n", duration.Round(time.Second)))
	sb.WriteString(fmt.Sprintf("  Commands Issued:  %d\n", ii.commandCount))

	if ii.llmProvider != nil {
		sb.WriteString(fmt.Sprintf("  LLM Provider:     %s\n", ii.llmProvider.Name()))
		sb.WriteString(fmt.Sprintf("  Provider Status:  %v\n", ii.llmProvider.Available()))
	}

	return sb.String()
}

func (ii *InteractiveIntrospection) cmdSave(args []string) string {
	if ii.stateIntegration != nil {
		if err := ii.stateIntegration.SaveState(); err != nil {
			return fmt.Sprintf("❌ Save failed: %v", err)
		}
		return "✅ State saved successfully."
	}
	return "⚠️  State integration not available."
}

func (ii *InteractiveIntrospection) cmdClear(args []string) string {
	// ANSI escape code to clear screen
	fmt.Print("\033[2J\033[H")
	return ""
}

// Helper functions

func truncateStringII(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Stop stops the interactive session
func (ii *InteractiveIntrospection) Stop() {
	ii.running = false
}
