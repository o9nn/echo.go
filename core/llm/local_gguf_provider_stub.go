//go:build !cgo || nollama
// +build !cgo nollama

package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/o9nn/echo.go/core/backendcap"
)

// LocalGGUFProviderConfig configures the local GGUF provider. In no-cgo or
// nollama builds it is retained so callers can compile while native inference is disabled.
type LocalGGUFProviderConfig struct {
	Name           string
	ModelPath      string
	ContextSize    int
	BatchSize      int
	Threads        int
	GPULayers      int
	FlashAttention bool
	KVCacheType    string
	Seed           uint32
	Capability     backendcap.Capability
}

// LocalGGUFProvider is a compile-safe stub when the maintained llama binding is unavailable.
type LocalGGUFProvider struct {
	config LocalGGUFProviderConfig
}

// NewLocalGGUFProvider creates a stub local provider.
func NewLocalGGUFProvider(modelPath string) *LocalGGUFProvider {
	return NewLocalGGUFProviderWithConfig(LocalGGUFProviderConfig{Name: "local_gguf", ModelPath: modelPath})
}

// NewLocalGGUFProviderFromCapability creates a stub provider from discovered metadata.
func NewLocalGGUFProviderFromCapability(cap backendcap.Capability) *LocalGGUFProvider {
	return NewLocalGGUFProviderWithConfig(LocalGGUFProviderConfig{Name: "local_gguf", ModelPath: cap.ModelPath, ContextSize: cap.ContextLength, Capability: cap})
}

// NewLocalGGUFProviderWithConfig creates a stub provider with explicit config.
func NewLocalGGUFProviderWithConfig(config LocalGGUFProviderConfig) *LocalGGUFProvider {
	if config.Name == "" {
		config.Name = "local_gguf"
	}
	return &LocalGGUFProvider{config: config}
}

func (lgp *LocalGGUFProvider) Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return "", fmt.Errorf("local GGUF support unavailable in this build (requires cgo and no nollama tag)")
}

func (lgp *LocalGGUFProvider) StreamGenerate(ctx context.Context, prompt string, opts GenerateOptions) (<-chan StreamChunk, error) {
	out := make(chan StreamChunk, 1)
	out <- StreamChunk{Error: fmt.Errorf("local GGUF support unavailable in this build (requires cgo and no nollama tag)")}
	close(out)
	return out, nil
}

func (lgp *LocalGGUFProvider) Name() string {
	if lgp.config.Name != "" {
		return lgp.config.Name
	}
	return "local_gguf"
}

func (lgp *LocalGGUFProvider) Available() bool {
	if lgp.config.ModelPath == "" {
		return false
	}
	_, _ = os.Stat(lgp.config.ModelPath)
	return false
}

func (lgp *LocalGGUFProvider) MaxTokens() int {
	return 0
}

// Loaded reports false when native GGUF support is unavailable.
func (lgp *LocalGGUFProvider) Loaded() bool {
	return false
}

// LoadError reports no retained native load error for the no-cgo stub.
func (lgp *LocalGGUFProvider) LoadError() error {
	return nil
}

func (lgp *LocalGGUFProvider) Close() error {
	return nil
}
