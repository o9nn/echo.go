package llm

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/o9nn/echo.go/core/backendcap"
)

func TestLocalGGUFProviderFromCapabilityPreservesProviderContract(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeLLMTinyGGUF(t, path)
	cap, err := backendcap.ProbeModelFile(path)
	if err != nil {
		t.Fatal(err)
	}
	provider := NewLocalGGUFProviderFromCapability(cap)
	if provider.Name() != "local_gguf" {
		t.Fatalf("expected local_gguf provider name, got %q", provider.Name())
	}
	if provider.MaxTokens() < 0 {
		t.Fatalf("unexpected negative max tokens: %d", provider.MaxTokens())
	}
}

func TestLocalGGUFProviderGatedRealModelSmoke(t *testing.T) {
	modelPath := strings.TrimSpace(os.Getenv("ECHO_TEST_GGUF_MODEL"))
	if modelPath == "" {
		t.Skip("set ECHO_TEST_GGUF_MODEL to run native real-model GGUF smoke test")
	}
	cap, err := backendcap.ProbeModelFile(modelPath)
	if err != nil {
		t.Fatalf("probe real GGUF model: %v", err)
	}
	provider := NewLocalGGUFProviderFromCapability(cap)
	if !provider.Available() {
		t.Skip("local GGUF provider is unavailable in this build or host memory envelope")
	}
	defer provider.Close()
	result, err := provider.Generate(context.Background(), "Deep Tree Echo says", GenerateOptions{MaxTokens: 1})
	if err != nil {
		t.Fatalf("real GGUF one-token smoke generation failed: %v", err)
	}
	if strings.TrimSpace(result) == "" {
		t.Fatalf("expected non-empty one-token smoke response")
	}
}

func writeLLMTinyGGUF(t *testing.T, path string) {
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
	writeLLMKVString(t, file, "general.name", "tiny-echo")
	writeLLMKVString(t, file, "general.architecture", "llama")
	writeLLMKVUint32(t, file, "llama.context_length", 2048)
	writeLLMKVUint32(t, file, "general.file_type", 15)
}

func writeLLMKVString(t *testing.T, file *os.File, key, value string) {
	t.Helper()
	writeLLMString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(8))
	writeLLMString(t, file, value)
}

func writeLLMKVUint32(t *testing.T, file *os.File, key string, value uint32) {
	t.Helper()
	writeLLMString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(4))
	binary.Write(file, binary.LittleEndian, value)
}

func writeLLMString(t *testing.T, file *os.File, value string) {
	t.Helper()
	binary.Write(file, binary.LittleEndian, uint64(len(value)))
	if _, err := file.Write([]byte(value)); err != nil {
		t.Fatal(err)
	}
}
