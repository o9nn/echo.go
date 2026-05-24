package backendcap

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestProbeHostMemoryReturnsTier(t *testing.T) {
	probe := ProbeHostMemory()
	if probe.Tier == "" {
		t.Fatal("expected host memory tier")
	}
}

func TestReadGGUFModelMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeTinyGGUF(t, path)

	metadata, err := ReadGGUFModelMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "tiny-echo" {
		t.Fatalf("unexpected model name %q", metadata.Name)
	}
	if metadata.Architecture != "llama" {
		t.Fatalf("unexpected architecture %q", metadata.Architecture)
	}
	if metadata.ContextLength != 2048 {
		t.Fatalf("unexpected context length %d", metadata.ContextLength)
	}
	if metadata.Quantization != "Q4_K_M" {
		t.Fatalf("unexpected quantization %q", metadata.Quantization)
	}
	if metadata.EstimatedMemoryBytes == 0 {
		t.Fatal("expected non-zero memory estimate")
	}
}

func TestDiscoverModelCapabilities(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.gguf")
	writeTinyGGUF(t, path)

	caps := DiscoverModelCapabilities([]string{dir})
	if len(caps) != 1 {
		t.Fatalf("expected one capability, got %d", len(caps))
	}
	if caps[0].ContextLength != 2048 || caps[0].Quantization != "Q4_K_M" {
		t.Fatalf("unexpected capability metadata: %+v", caps[0])
	}
}

func writeTinyGGUF(t *testing.T, path string) {
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
	writeKVString(t, file, "general.name", "tiny-echo")
	writeKVString(t, file, "general.architecture", "llama")
	writeKVUint32(t, file, "llama.context_length", 2048)
	writeKVUint32(t, file, "general.file_type", 15)
}

func writeKVString(t *testing.T, file *os.File, key, value string) {
	t.Helper()
	writeString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(8))
	writeString(t, file, value)
}

func writeKVUint32(t *testing.T, file *os.File, key string, value uint32) {
	t.Helper()
	writeString(t, file, key)
	binary.Write(file, binary.LittleEndian, uint32(4))
	binary.Write(file, binary.LittleEndian, value)
}

func writeString(t *testing.T, file *os.File, value string) {
	t.Helper()
	binary.Write(file, binary.LittleEndian, uint64(len(value)))
	if _, err := file.Write([]byte(value)); err != nil {
		t.Fatal(err)
	}
}
