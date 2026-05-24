package llm

import (
	"encoding/binary"
	"os"
	"path/filepath"
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
