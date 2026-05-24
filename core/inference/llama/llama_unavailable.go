//go:build !llama_legacy

// Package llama exposes the legacy core/inference llama API when the native
// direct-link wrapper is unavailable.
//
// The real, maintained source-based llama.cpp binding for the repository lives
// at github.com/o9nn/echo.go/llama. The legacy core/inference wrapper is kept
// behind the explicit `llama_legacy` build tag because it links against
// prebuilt libs/libllama and libs/libggml* artifacts that are not guaranteed to
// exist in normal development, CI, or CPU-only sandboxes.
package llama

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnavailable reports that the legacy direct-link llama backend is not part
// of this build. Callers should query backend capabilities and route inference
// to an available backend rather than assuming native llama is always present.
var ErrUnavailable = errors.New("legacy core/inference/llama backend unavailable: build with -tags llama_legacy after validating native libs")

// Available reports whether this package was built with a usable native backend.
func Available() bool { return false }

// BackendInit is a no-op when the legacy backend is unavailable.
func BackendInit() {}

// BackendFree is a no-op when the legacy backend is unavailable.
func BackendFree() {}

// MaxDevices returns zero because no native llama devices are available.
func MaxDevices() int { return 0 }

// Token represents a llama token ID.
type Token = int32

// Pos represents a position in the sequence.
type Pos = int32

// SeqID represents a sequence ID.
type SeqID = int32

// VocabType represents vocabulary types.
type VocabType int

const (
	VocabTypeNone VocabType = iota
	VocabTypeSPM
	VocabTypeBPE
	VocabTypeWPM
	VocabTypeUGM
	VocabTypeRWKV
)

// RopeType represents RoPE types.
type RopeType int

const (
	RopeTypeNone RopeType = -1
	RopeTypeNorm RopeType = 0
	RopeTypeNeox RopeType = 2
)

// SplitMode represents model split modes.
type SplitMode int

const (
	SplitModeNone SplitMode = iota
	SplitModeLayer
	SplitModeRow
)

// PoolingType represents pooling types.
type PoolingType int

const (
	PoolingTypeUnspecified PoolingType = -1
	PoolingTypeNone        PoolingType = 0
	PoolingTypeMean        PoolingType = 1
	PoolingTypeCLS         PoolingType = 2
	PoolingTypeLast        PoolingType = 3
	PoolingTypeRank        PoolingType = 4
)

// AttentionType represents attention types.
type AttentionType int

const (
	AttentionTypeUnspecified AttentionType = -1
	AttentionTypeCausal      AttentionType = 0
	AttentionTypeNonCausal   AttentionType = 1
)

// ModelParams configures model loading.
type ModelParams struct {
	NGPULayers int
	SplitMode  SplitMode
	MainGPU    int
	VocabOnly  bool
	UseMmap    bool
	UseMlock   bool
}

// DefaultModelParams returns conservative CPU defaults even when unavailable.
func DefaultModelParams() ModelParams {
	return ModelParams{SplitMode: SplitModeNone, UseMmap: true}
}

// Model represents an unavailable legacy llama model handle.
type Model struct{}

// LoadModel returns ErrUnavailable because the native legacy backend is absent.
func LoadModel(path string, params ModelParams) (*Model, error) {
	return nil, fmt.Errorf("%w: cannot load %q", ErrUnavailable, path)
}

func (m *Model) Free()                           {}
func (m *Model) NCtxTrain() int                  { return 0 }
func (m *Model) NEmbd() int                      { return 0 }
func (m *Model) NLayer() int                     { return 0 }
func (m *Model) NHead() int                      { return 0 }
func (m *Model) NVocab() int                     { return 0 }
func (m *Model) RopeType() RopeType              { return RopeTypeNone }
func (m *Model) HasEncoder() bool                { return false }
func (m *Model) HasDecoder() bool                { return false }
func (m *Model) IsRecurrent() bool               { return false }
func (m *Model) Description() string             { return "" }
func (m *Model) ChatTemplate(name string) string { return "" }
func (m *Model) AddBosToken() bool               { return false }
func (m *Model) AddEosToken() bool               { return false }
func (m *Model) GetVocab() *Vocab                { return &Vocab{model: m} }

// Vocab represents an unavailable model vocabulary.
type Vocab struct{ model *Model }

func (v *Vocab) NTokens() int               { return 0 }
func (v *Vocab) Type() VocabType            { return VocabTypeNone }
func (v *Vocab) BOS() Token                 { return -1 }
func (v *Vocab) EOS() Token                 { return -1 }
func (v *Vocab) EOT() Token                 { return -1 }
func (v *Vocab) SEP() Token                 { return -1 }
func (v *Vocab) NL() Token                  { return -1 }
func (v *Vocab) PAD() Token                 { return -1 }
func (v *Vocab) IsEOG(token Token) bool     { return false }
func (v *Vocab) IsControl(token Token) bool { return false }
func (v *Vocab) Tokenize(text string, addSpecial, parseSpecial bool) ([]Token, error) {
	return nil, ErrUnavailable
}
func (v *Vocab) TokenToPiece(token Token, special bool) string { return "" }
func (v *Vocab) Detokenize(tokens []Token, removeSpecial, unparseSpecial bool) string {
	return ""
}

// ContextParams configures context creation.
type ContextParams struct {
	NCtx          uint32
	NBatch        uint32
	NUBatch       uint32
	NSeqMax       uint32
	NThreads      int32
	NThreadsBatch int32
	PoolingType   PoolingType
	AttentionType AttentionType
	RopeFreqBase  float32
	RopeFreqScale float32
	LogitsAll     bool
	Embeddings    bool
	FlashAttn     bool
	OffloadKQV    bool
}

// DefaultContextParams returns minimal CPU-safe defaults.
func DefaultContextParams() ContextParams {
	return ContextParams{NCtx: 2048, NBatch: 512, NUBatch: 512, NSeqMax: 1, NThreads: 1, NThreadsBatch: 1, PoolingType: PoolingTypeUnspecified, AttentionType: AttentionTypeCausal}
}

// Context represents an unavailable inference context.
type Context struct{ model *Model }

func (m *Model) NewContext(params ContextParams) (*Context, error) { return nil, ErrUnavailable }
func (c *Context) Free()                                           {}
func (c *Context) Model() *Model                                   { return c.model }
func (c *Context) NCtx() uint32                                    { return 0 }
func (c *Context) NBatch() uint32                                  { return 0 }
func (c *Context) NSeqMax() uint32                                 { return 0 }

// Batch stores token metadata in pure Go so callers can construct batches even
// when decode/encode is unavailable.
type Batch struct {
	tokens []Token
	pos    []Pos
	seqID  []SeqID
	logits []bool
	n      int32
}

func NewBatch(nTokens, embd, nSeqMax int32) *Batch {
	if nTokens < 0 {
		nTokens = 0
	}
	return &Batch{tokens: make([]Token, nTokens), pos: make([]Pos, nTokens), seqID: make([]SeqID, nTokens), logits: make([]bool, nTokens)}
}
func (b *Batch) Free()          {}
func (b *Batch) NTokens() int32 { return b.n }
func (b *Batch) SetNTokens(n int32) {
	if n >= 0 && int(n) <= len(b.tokens) {
		b.n = n
	}
}
func (b *Batch) SetToken(i int, token Token, pos Pos, seqID SeqID, logits bool) {
	if i < 0 || i >= len(b.tokens) {
		return
	}
	b.tokens[i] = token
	b.pos[i] = pos
	b.seqID[i] = seqID
	b.logits[i] = logits
}
func (b *Batch) Clear() { b.n = 0 }
func (b *Batch) AddToken(token Token, pos Pos, seqID SeqID, logits bool) {
	if int(b.n) >= len(b.tokens) {
		return
	}
	b.SetToken(int(b.n), token, pos, seqID, logits)
	b.n++
}

func (c *Context) Decode(batch *Batch) error                         { return ErrUnavailable }
func (c *Context) Encode(batch *Batch) error                         { return ErrUnavailable }
func (c *Context) GetLogits() []float32                              { return nil }
func (c *Context) GetLogitsIth(i int32) []float32                    { return nil }
func (c *Context) GetEmbeddings() []float32                          { return nil }
func (c *Context) GetEmbeddingsIth(i int32) []float32                { return nil }
func (c *Context) KVCacheClear()                                     {}
func (c *Context) KVCacheSeqRm(seqID SeqID, p0, p1 Pos) bool         { return false }
func (c *Context) KVCacheSeqCp(seqIDSrc, seqIDDst SeqID, p0, p1 Pos) {}
func (c *Context) KVCacheSeqKeep(seqID SeqID)                        {}
func (c *Context) KVCacheSeqAdd(seqID SeqID, p0, p1, delta Pos)      {}
func (c *Context) KVCacheSeqPosMax(seqID SeqID) Pos                  { return 0 }
func (c *Context) KVCacheDefrag()                                    {}
func (c *Context) KVCacheUpdate()                                    {}
func (c *Context) KVCacheTokenCount() int32                          { return 0 }
func (c *Context) KVCacheUsedCells() int32                           { return 0 }

// Sampler represents an unavailable sampler.
type Sampler struct{ name string }

func NewSamplerGreedy() *Sampler                        { return &Sampler{name: "unavailable-greedy"} }
func NewSamplerDist(seed uint32) *Sampler               { return &Sampler{name: "unavailable-dist"} }
func NewSamplerTopK(k int32) *Sampler                   { return &Sampler{name: "unavailable-top-k"} }
func NewSamplerTopP(p float32, minKeep int) *Sampler    { return &Sampler{name: "unavailable-top-p"} }
func NewSamplerMinP(p float32, minKeep int) *Sampler    { return &Sampler{name: "unavailable-min-p"} }
func NewSamplerTypical(p float32, minKeep int) *Sampler { return &Sampler{name: "unavailable-typical"} }
func NewSamplerTemp(t float32) *Sampler                 { return &Sampler{name: "unavailable-temp"} }
func NewSamplerMirostat(nVocab int32, seed uint32, tau, eta float32, m int32) *Sampler {
	return &Sampler{name: "unavailable-mirostat"}
}
func NewSamplerMirostatV2(seed uint32, tau, eta float32) *Sampler {
	return &Sampler{name: "unavailable-mirostat-v2"}
}
func NewSamplerPenalties(penaltyLastN int32, penaltyRepeat, penaltyFreq, penaltyPresent float32) *Sampler {
	return &Sampler{name: "unavailable-penalties"}
}
func (s *Sampler) Free()  {}
func (s *Sampler) Reset() {}
func (s *Sampler) Clone() *Sampler {
	if s == nil {
		return nil
	}
	return &Sampler{name: s.name}
}
func (s *Sampler) Name() string {
	if s == nil {
		return "unavailable"
	}
	return s.name
}
func (s *Sampler) Accept(token Token)                   {}
func (s *Sampler) Sample(ctx *Context, idx int32) Token { return -1 }

// SamplerChain represents an unavailable sampler chain.
type SamplerChain struct{ samplers []*Sampler }

func NewSamplerChain(noPerf bool) *SamplerChain               { return &SamplerChain{samplers: make([]*Sampler, 0)} }
func (sc *SamplerChain) Free()                                { sc.samplers = nil }
func (sc *SamplerChain) Add(s *Sampler)                       { sc.samplers = append(sc.samplers, s) }
func (sc *SamplerChain) N() int32                             { return int32(len(sc.samplers)) }
func (sc *SamplerChain) Sample(ctx *Context, idx int32) Token { return -1 }
func (sc *SamplerChain) Accept(token Token)                   {}
func (sc *SamplerChain) Reset()                               {}

func (c *Context) GetStateSize() int            { return 0 }
func (c *Context) CopyStateData(dst []byte) int { return 0 }
func (c *Context) SetStateData(src []byte) int  { return 0 }

// ChatMessage represents a chat message.
type ChatMessage struct {
	Role    string
	Content string
}

// ApplyChatTemplate fails explicitly when the legacy backend is unavailable.
func ApplyChatTemplate(template string, messages []ChatMessage, addAss bool) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}
	roles := make([]string, 0, len(messages))
	for _, msg := range messages {
		roles = append(roles, msg.Role)
	}
	return "", fmt.Errorf("%w: chat template requested for roles [%s]", ErrUnavailable, strings.Join(roles, ","))
}
