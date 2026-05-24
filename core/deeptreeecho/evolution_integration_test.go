package deeptreeecho

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/o9nn/echo.go/core/backendcap"
)

func TestDefaultEvolutionSystemConfig(t *testing.T) {
	config := DefaultEvolutionSystemConfig()

	if len(config.PreferredProviders) != 4 {
		t.Errorf("Expected 4 preferred providers, got %d", len(config.PreferredProviders))
	}

	expectedProviders := []string{"local_gguf", "anthropic", "openrouter", "openai"}
	for i, provider := range config.PreferredProviders {
		if provider != expectedProviders[i] {
			t.Errorf("Expected provider %s at index %d, got %s", expectedProviders[i], i, provider)
		}
	}

	if config.Debug {
		t.Error("Expected Debug to be false by default")
	}
}

func TestEvolutionSystemWithoutAPIProvidersUsesContinuityFallback(t *testing.T) {
	// Save and clear environment variables
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	// Restore at end
	defer func() {
		if anthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", anthropicKey)
		}
		if openrouterKey != "" {
			os.Setenv("OPENROUTER_API_KEY", openrouterKey)
		}
		if openaiKey != "" {
			os.Setenv("OPENAI_API_KEY", openaiKey)
		}
	}()

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)

	if err != nil {
		t.Fatalf("expected fallback-backed evolution system, got error: %v", err)
	}

	providers := es.GetStatus()["providers"].([]string)
	foundFallback := false
	for _, provider := range providers {
		if provider == "SimpleFallback" {
			foundFallback = true
		}
	}
	if !foundFallback {
		t.Fatalf("expected SimpleFallback provider in continuity mode, got %v", providers)
	}
}

func TestEvolutionSystemModelPathsSurfaceConcreteCapabilities(t *testing.T) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	openrouterKey := os.Getenv("OPENROUTER_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	modelPaths := os.Getenv("ECHO_MODEL_PATHS")
	localModel := os.Getenv("LOCAL_MODEL_PATH")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ECHO_MODEL_PATHS")
	os.Unsetenv("LOCAL_MODEL_PATH")
	defer func() {
		restoreEnv("ANTHROPIC_API_KEY", anthropicKey)
		restoreEnv("OPENROUTER_API_KEY", openrouterKey)
		restoreEnv("OPENAI_API_KEY", openaiKey)
		restoreEnv("ECHO_MODEL_PATHS", modelPaths)
		restoreEnv("LOCAL_MODEL_PATH", localModel)
	}()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "tiny.gguf")
	writeEvolutionTinyGGUF(t, modelPath)
	config := DefaultEvolutionSystemConfig()
	config.ModelPaths = []string{dir}
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("expected fallback/local-capability backed system, got error: %v", err)
	}
	status := es.GetStatus()
	caps, ok := status["backend_capabilities"].([]backendcap.Capability)
	if !ok {
		t.Fatalf("expected backend capabilities in status, got %T", status["backend_capabilities"])
	}
	found := false
	for _, cap := range caps {
		if cap.ModelPath == modelPath && cap.ContextLength == 2048 && cap.Quantization == "Q4_K_M" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected concrete model capability in status, got %+v", caps)
	}
}

func restoreEnv(name, value string) {
	if value == "" {
		os.Unsetenv(name)
		return
	}
	os.Setenv(name, value)
}

func writeEvolutionTinyGGUF(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.Write([]byte("GGUF"))
	binary.Write(file, binary.LittleEndian, uint32(3))
	binary.Write(file, binary.LittleEndian, uint64(0))
	binary.Write(file, binary.LittleEndian, uint64(4))
	writeEvolutionKVString(t, file, "general.name", "tiny-echo")
	writeEvolutionKVString(t, file, "general.architecture", "llama")
	writeEvolutionKVUint32(t, file, "llama.context_length", 2048)
	writeEvolutionKVUint32(t, file, "general.file_type", 15)
}

func writeEvolutionKVString(t *testing.T, file *os.File, key, value string) {
	t.Helper()
	writeEvolutionString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(8))
	writeEvolutionString(t, file, value)
}

func writeEvolutionKVUint32(t *testing.T, file *os.File, key string, value uint32) {
	t.Helper()
	writeEvolutionString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(4))
	binary.Write(file, binary.LittleEndian, value)
}

func writeEvolutionString(t *testing.T, file *os.File, value string) {
	t.Helper()
	binary.Write(file, binary.LittleEndian, uint64(len(value)))
	if _, err := file.Write([]byte(value)); err != nil {
		t.Fatal(err)
	}
}
func TestEvolutionSystemStatus(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	config.Debug = false

	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	status := es.GetStatus()

	if !status["initialized"].(bool) {
		t.Error("Expected system to be initialized")
	}

	if status["running"].(bool) {
		t.Error("Expected system to not be running initially")
	}

	providers, ok := status["providers"].([]string)
	if !ok || len(providers) == 0 {
		t.Error("Expected at least one provider")
	}
}

func TestEvolutionSystemDiagnostics(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	ctx := context.Background()
	diag, err := es.RunDiagnostics(ctx)
	if err != nil {
		t.Fatalf("Failed to run diagnostics: %v", err)
	}

	if len(diag.Tests) == 0 {
		t.Error("Expected at least one diagnostic test")
	}

	if diag.OverallHealth == "" {
		t.Error("Expected overall health to be set")
	}

	// Check specific tests exist
	testNames := make(map[string]bool)
	for _, test := range diag.Tests {
		testNames[test.Name] = true
	}

	expectedTests := []string{
		"LLM Provider",
		"Stream of Consciousness",
		"Echobeats Scheduler",
		"Echodream Integration",
		"Genetic Traits",
	}

	for _, name := range expectedTests {
		if !testNames[name] {
			t.Errorf("Expected diagnostic test '%s' not found", name)
		}
	}
}

func TestEvolutionSystemGoals(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	goalID := es.AddGoal("Test evolutionary goal", 0.8)
	if goalID == "" {
		t.Error("Expected non-empty goal ID")
	}

	metrics := es.GetSchedulerMetrics()
	if metrics == nil {
		t.Error("Expected non-nil scheduler metrics")
	}
}

func TestEvolutionSystemMetrics(t *testing.T) {
	// This test only runs if at least one API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" &&
		os.Getenv("OPENROUTER_API_KEY") == "" &&
		os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("No API keys set, skipping integration test")
	}

	config := DefaultEvolutionSystemConfig()
	es, err := NewEvolutionSystem(config)
	if err != nil {
		t.Fatalf("Failed to create evolution system: %v", err)
	}

	// Test consciousness metrics
	consMetrics := es.GetConsciousnessMetrics()
	if consMetrics == nil {
		t.Error("Expected non-nil consciousness metrics")
	}

	// Test scheduler metrics
	schedMetrics := es.GetSchedulerMetrics()
	if schedMetrics == nil {
		t.Error("Expected non-nil scheduler metrics")
	}

	// Test dream metrics
	dreamMetrics := es.GetDreamMetrics()
	if dreamMetrics == nil {
		t.Error("Expected non-nil dream metrics")
	}

	// Test evolution history
	history := es.GetEvolutionHistory()
	if history == nil {
		t.Error("Expected non-nil evolution history")
	}
}
