// Package llama provides Go bindings for llama.cpp inference
// llama.cpp is a high-performance LLM inference library
package llama

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../ggml
#cgo LDFLAGS: -L${SRCDIR}/../../../libs -lllama -lggml-base -lggml -lggml-cpu -lm -lpthread -lstdc++
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../libs/arm64-v8a

#include "llama.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// =============================================================================
// LLAMA TYPES
// =============================================================================

// Token represents a llama token ID
type Token = int32

// Pos represents a position in the sequence
type Pos = int32

// SeqID represents a sequence ID
type SeqID = int32

// VocabType represents vocabulary types
type VocabType int

const (
	VocabTypeNone VocabType = C.LLAMA_VOCAB_TYPE_NONE
	VocabTypeSPM  VocabType = C.LLAMA_VOCAB_TYPE_SPM
	VocabTypeBPE  VocabType = C.LLAMA_VOCAB_TYPE_BPE
	VocabTypeWPM  VocabType = C.LLAMA_VOCAB_TYPE_WPM
	VocabTypeUGM  VocabType = C.LLAMA_VOCAB_TYPE_UGM
	VocabTypeRWKV VocabType = C.LLAMA_VOCAB_TYPE_RWKV
)

// RopeType represents RoPE types
type RopeType int

const (
	RopeTypeNone RopeType = C.LLAMA_ROPE_TYPE_NONE
	RopeTypeNorm RopeType = C.LLAMA_ROPE_TYPE_NORM
	RopeTypeNeox RopeType = C.LLAMA_ROPE_TYPE_NEOX
)

// SplitMode represents model split modes
type SplitMode int

const (
	SplitModeNone  SplitMode = C.LLAMA_SPLIT_MODE_NONE
	SplitModeLayer SplitMode = C.LLAMA_SPLIT_MODE_LAYER
	SplitModeRow   SplitMode = C.LLAMA_SPLIT_MODE_ROW
)

// PoolingType represents pooling types
type PoolingType int

const (
	PoolingTypeUnspecified PoolingType = C.LLAMA_POOLING_TYPE_UNSPECIFIED
	PoolingTypeNone        PoolingType = C.LLAMA_POOLING_TYPE_NONE
	PoolingTypeMean        PoolingType = C.LLAMA_POOLING_TYPE_MEAN
	PoolingTypeCLS         PoolingType = C.LLAMA_POOLING_TYPE_CLS
	PoolingTypeLast        PoolingType = C.LLAMA_POOLING_TYPE_LAST
	PoolingTypeRank        PoolingType = C.LLAMA_POOLING_TYPE_RANK
)

// AttentionType represents attention types
type AttentionType int

const (
	AttentionTypeUnspecified AttentionType = C.LLAMA_ATTENTION_TYPE_UNSPECIFIED
	AttentionTypeCausal      AttentionType = C.LLAMA_ATTENTION_TYPE_CAUSAL
	AttentionTypeNonCausal   AttentionType = C.LLAMA_ATTENTION_TYPE_NON_CAUSAL
)

// =============================================================================
// LLAMA BACKEND
// =============================================================================

var backendInitOnce sync.Once
var backendInitialized bool

// BackendInit initializes the llama backend
func BackendInit() {
	backendInitOnce.Do(func() {
		C.llama_backend_init()
		backendInitialized = true
	})
}

// BackendFree frees the llama backend
func BackendFree() {
	if backendInitialized {
		C.llama_backend_free()
		backendInitialized = false
	}
}

// MaxDevices returns the maximum number of devices
func MaxDevices() int {
	return int(C.llama_max_devices())
}

// =============================================================================
// LLAMA MODEL PARAMS
// =============================================================================

// ModelParams configures model loading
type ModelParams struct {
	NGPULayers int       // Number of layers to offload to GPU
	SplitMode  SplitMode // How to split the model across GPUs
	MainGPU    int       // Main GPU index
	VocabOnly  bool      // Only load vocabulary
	UseMmap    bool      // Use memory mapping
	UseMlock   bool      // Lock memory
}

// DefaultModelParams returns default model parameters
func DefaultModelParams() ModelParams {
	cParams := C.llama_model_default_params()
	return ModelParams{
		NGPULayers: int(cParams.n_gpu_layers),
		SplitMode:  SplitMode(cParams.split_mode),
		MainGPU:    int(cParams.main_gpu),
		VocabOnly:  bool(cParams.vocab_only),
		UseMmap:    bool(cParams.use_mmap),
		UseMlock:   bool(cParams.use_mlock),
	}
}

// =============================================================================
// LLAMA MODEL
// =============================================================================

// Model represents a loaded llama model
type Model struct {
	model *C.struct_llama_model
	mu    sync.RWMutex
}

// LoadModel loads a model from a file
func LoadModel(path string, params ModelParams) (*Model, error) {
	BackendInit()
	
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	
	cParams := C.llama_model_default_params()
	cParams.n_gpu_layers = C.int32_t(params.NGPULayers)
	cParams.split_mode = C.enum_llama_split_mode(params.SplitMode)
	cParams.main_gpu = C.int32_t(params.MainGPU)
	cParams.vocab_only = C.bool(params.VocabOnly)
	cParams.use_mmap = C.bool(params.UseMmap)
	cParams.use_mlock = C.bool(params.UseMlock)
	
	model := C.llama_model_load_from_file(cPath, cParams)
	if model == nil {
		return nil, fmt.Errorf("failed to load model from %s", path)
	}
	
	m := &Model{model: model}
	runtime.SetFinalizer(m, (*Model).Free)
	return m, nil
}

// Free releases the model resources
func (m *Model) Free() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.model != nil {
		C.llama_model_free(m.model)
		m.model = nil
	}
}

// NCtxTrain returns the training context size
func (m *Model) NCtxTrain() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int(C.llama_model_n_ctx_train(m.model))
}

// NEmbd returns the embedding dimension
func (m *Model) NEmbd() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int(C.llama_model_n_embd(m.model))
}

// NLayer returns the number of layers
func (m *Model) NLayer() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int(C.llama_model_n_layer(m.model))
}

// NHead returns the number of attention heads
func (m *Model) NHead() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int(C.llama_model_n_head(m.model))
}

// NVocab returns the vocabulary size
func (m *Model) NVocab() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int(C.llama_model_n_vocab(m.model))
}

// RopeType returns the RoPE type
func (m *Model) RopeType() RopeType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return RopeType(C.llama_model_rope_type(m.model))
}

// HasEncoder returns true if the model has an encoder
func (m *Model) HasEncoder() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return bool(C.llama_model_has_encoder(m.model))
}

// HasDecoder returns true if the model has a decoder
func (m *Model) HasDecoder() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return bool(C.llama_model_has_decoder(m.model))
}

// IsRecurrent returns true if the model is recurrent
func (m *Model) IsRecurrent() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return bool(C.llama_model_is_recurrent(m.model))
}

// Description returns the model description
func (m *Model) Description() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	buf := make([]byte, 256)
	n := C.llama_model_desc(m.model, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
	if n <= 0 {
		return ""
	}
	return string(buf[:n])
}

// ChatTemplate returns the chat template
func (m *Model) ChatTemplate(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var cName *C.char
	if name != "" {
		cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
	}
	
	tmpl := C.llama_model_chat_template(m.model, cName)
	if tmpl == nil {
		return ""
	}
	return C.GoString(tmpl)
}

// AddBosToken returns true if BOS token should be added
func (m *Model) AddBosToken() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return bool(C.llama_add_bos_token(m.model))
}

// AddEosToken returns true if EOS token should be added
func (m *Model) AddEosToken() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return bool(C.llama_add_eos_token(m.model))
}

// =============================================================================
// LLAMA VOCAB
// =============================================================================

// Vocab represents the model vocabulary
type Vocab struct {
	vocab *C.struct_llama_vocab
	model *Model
}

// GetVocab returns the model's vocabulary
func (m *Model) GetVocab() *Vocab {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	vocab := C.llama_model_get_vocab(m.model)
	return &Vocab{vocab: vocab, model: m}
}

// NTokens returns the number of tokens in the vocabulary
func (v *Vocab) NTokens() int {
	return int(C.llama_vocab_n_tokens(v.vocab))
}

// Type returns the vocabulary type
func (v *Vocab) Type() VocabType {
	return VocabType(C.llama_vocab_type(v.vocab))
}

// BOS returns the beginning-of-sequence token
func (v *Vocab) BOS() Token {
	return Token(C.llama_vocab_bos(v.vocab))
}

// EOS returns the end-of-sequence token
func (v *Vocab) EOS() Token {
	return Token(C.llama_vocab_eos(v.vocab))
}

// EOT returns the end-of-turn token
func (v *Vocab) EOT() Token {
	return Token(C.llama_vocab_eot(v.vocab))
}

// SEP returns the separator token
func (v *Vocab) SEP() Token {
	return Token(C.llama_vocab_sep(v.vocab))
}

// NL returns the newline token
func (v *Vocab) NL() Token {
	return Token(C.llama_vocab_nl(v.vocab))
}

// PAD returns the padding token
func (v *Vocab) PAD() Token {
	return Token(C.llama_vocab_pad(v.vocab))
}

// IsEOG returns true if the token is end-of-generation
func (v *Vocab) IsEOG(token Token) bool {
	return bool(C.llama_vocab_is_eog(v.vocab, C.llama_token(token)))
}

// IsControl returns true if the token is a control token
func (v *Vocab) IsControl(token Token) bool {
	return bool(C.llama_vocab_is_control(v.vocab, C.llama_token(token)))
}

// Tokenize converts text to tokens
func (v *Vocab) Tokenize(text string, addSpecial, parseSpecial bool) ([]Token, error) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	
	// First call to get required size
	nTokens := C.llama_tokenize(v.vocab, cText, C.int32_t(len(text)), nil, 0, C.bool(addSpecial), C.bool(parseSpecial))
	if nTokens < 0 {
		nTokens = -nTokens
	}
	
	tokens := make([]Token, nTokens)
	if nTokens == 0 {
		return tokens, nil
	}
	
	n := C.llama_tokenize(v.vocab, cText, C.int32_t(len(text)),
		(*C.llama_token)(unsafe.Pointer(&tokens[0])), C.int32_t(nTokens),
		C.bool(addSpecial), C.bool(parseSpecial))
	
	if n < 0 {
		return nil, fmt.Errorf("tokenization failed: %d", n)
	}
	
	return tokens[:n], nil
}

// TokenToPiece converts a token to its string representation
func (v *Vocab) TokenToPiece(token Token, special bool) string {
	buf := make([]byte, 128)
	n := C.llama_token_to_piece(v.vocab, C.llama_token(token),
		(*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)), 0, C.bool(special))
	
	if n < 0 {
		// Buffer too small, try again with larger buffer
		buf = make([]byte, -n)
		n = C.llama_token_to_piece(v.vocab, C.llama_token(token),
			(*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)), 0, C.bool(special))
	}
	
	if n <= 0 {
		return ""
	}
	return string(buf[:n])
}

// Detokenize converts tokens back to text
func (v *Vocab) Detokenize(tokens []Token, removeSpecial, unparseSpecial bool) string {
	if len(tokens) == 0 {
		return ""
	}
	
	buf := make([]byte, len(tokens)*32)
	n := C.llama_detokenize(v.vocab,
		(*C.llama_token)(unsafe.Pointer(&tokens[0])), C.int32_t(len(tokens)),
		(*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)),
		C.bool(removeSpecial), C.bool(unparseSpecial))
	
	if n < 0 {
		// Buffer too small
		buf = make([]byte, -n)
		n = C.llama_detokenize(v.vocab,
			(*C.llama_token)(unsafe.Pointer(&tokens[0])), C.int32_t(len(tokens)),
			(*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)),
			C.bool(removeSpecial), C.bool(unparseSpecial))
	}
	
	if n <= 0 {
		return ""
	}
	return string(buf[:n])
}

// =============================================================================
// LLAMA CONTEXT PARAMS
// =============================================================================

// ContextParams configures context creation
type ContextParams struct {
	NCtx          uint32        // Context size
	NBatch        uint32        // Batch size for prompt processing
	NUBatch       uint32        // Micro-batch size
	NSeqMax       uint32        // Maximum number of sequences
	NThreads      int32         // Number of threads for generation
	NThreadsBatch int32         // Number of threads for batch processing
	PoolingType   PoolingType   // Pooling type for embeddings
	AttentionType AttentionType // Attention type
	RopeFreqBase  float32       // RoPE base frequency
	RopeFreqScale float32       // RoPE frequency scaling
	LogitsAll     bool          // Return logits for all tokens
	Embeddings    bool          // Enable embeddings mode
	FlashAttn     bool          // Use flash attention
	OffloadKQV    bool          // Offload KQV to GPU
}

// DefaultContextParams returns default context parameters
func DefaultContextParams() ContextParams {
	cParams := C.llama_context_default_params()
	return ContextParams{
		NCtx:          uint32(cParams.n_ctx),
		NBatch:        uint32(cParams.n_batch),
		NUBatch:       uint32(cParams.n_ubatch),
		NSeqMax:       uint32(cParams.n_seq_max),
		NThreads:      int32(cParams.n_threads),
		NThreadsBatch: int32(cParams.n_threads_batch),
		PoolingType:   PoolingType(cParams.pooling_type),
		AttentionType: AttentionType(cParams.attention_type),
		RopeFreqBase:  float32(cParams.rope_freq_base),
		RopeFreqScale: float32(cParams.rope_freq_scale),
		LogitsAll:     bool(cParams.logits_all),
		Embeddings:    bool(cParams.embeddings),
		FlashAttn:     bool(cParams.flash_attn),
		OffloadKQV:    bool(cParams.offload_kqv),
	}
}

// =============================================================================
// LLAMA CONTEXT
// =============================================================================

// Context represents a llama inference context
type Context struct {
	ctx   *C.struct_llama_context
	model *Model
	mu    sync.RWMutex
}

// NewContext creates a new context from a model
func (m *Model) NewContext(params ContextParams) (*Context, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	cParams := C.llama_context_default_params()
	cParams.n_ctx = C.uint32_t(params.NCtx)
	cParams.n_batch = C.uint32_t(params.NBatch)
	cParams.n_ubatch = C.uint32_t(params.NUBatch)
	cParams.n_seq_max = C.uint32_t(params.NSeqMax)
	cParams.n_threads = C.int32_t(params.NThreads)
	cParams.n_threads_batch = C.int32_t(params.NThreadsBatch)
	cParams.pooling_type = C.enum_llama_pooling_type(params.PoolingType)
	cParams.attention_type = C.enum_llama_attention_type(params.AttentionType)
	cParams.rope_freq_base = C.float(params.RopeFreqBase)
	cParams.rope_freq_scale = C.float(params.RopeFreqScale)
	cParams.logits_all = C.bool(params.LogitsAll)
	cParams.embeddings = C.bool(params.Embeddings)
	cParams.flash_attn = C.bool(params.FlashAttn)
	cParams.offload_kqv = C.bool(params.OffloadKQV)
	
	ctx := C.llama_init_from_model(m.model, cParams)
	if ctx == nil {
		return nil, errors.New("failed to create context")
	}
	
	c := &Context{ctx: ctx, model: m}
	runtime.SetFinalizer(c, (*Context).Free)
	return c, nil
}

// Free releases the context resources
func (c *Context) Free() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.ctx != nil {
		C.llama_free(c.ctx)
		c.ctx = nil
	}
}

// Model returns the underlying model
func (c *Context) Model() *Model {
	return c.model
}

// NCtx returns the context size
func (c *Context) NCtx() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return uint32(C.llama_n_ctx(c.ctx))
}

// NBatch returns the batch size
func (c *Context) NBatch() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return uint32(C.llama_n_batch(c.ctx))
}

// NSeqMax returns the maximum number of sequences
func (c *Context) NSeqMax() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return uint32(C.llama_n_seq_max(c.ctx))
}

// =============================================================================
// LLAMA BATCH
// =============================================================================

// Batch represents a batch of tokens for processing
type Batch struct {
	batch C.struct_llama_batch
}

// NewBatch creates a new batch
func NewBatch(nTokens, embd, nSeqMax int32) *Batch {
	batch := C.llama_batch_init(C.int32_t(nTokens), C.int32_t(embd), C.int32_t(nSeqMax))
	b := &Batch{batch: batch}
	runtime.SetFinalizer(b, (*Batch).Free)
	return b
}

// Free releases the batch resources
func (b *Batch) Free() {
	C.llama_batch_free(b.batch)
}

// NTokens returns the number of tokens in the batch
func (b *Batch) NTokens() int32 {
	return int32(b.batch.n_tokens)
}

// SetNTokens sets the number of tokens
func (b *Batch) SetNTokens(n int32) {
	b.batch.n_tokens = C.int32_t(n)
}

// SetToken sets a token at position i
func (b *Batch) SetToken(i int, token Token, pos Pos, seqID SeqID, logits bool) {
	b.batch.token[i] = C.llama_token(token)
	b.batch.pos[i] = C.llama_pos(pos)
	b.batch.n_seq_id[i] = 1
	b.batch.seq_id[i][0] = C.llama_seq_id(seqID)
	if logits {
		b.batch.logits[i] = 1
	} else {
		b.batch.logits[i] = 0
	}
}

// Clear resets the batch
func (b *Batch) Clear() {
	b.batch.n_tokens = 0
}

// AddToken adds a token to the batch
func (b *Batch) AddToken(token Token, pos Pos, seqID SeqID, logits bool) {
	i := int(b.batch.n_tokens)
	b.SetToken(i, token, pos, seqID, logits)
	b.batch.n_tokens++
}

// =============================================================================
// LLAMA DECODE
// =============================================================================

// Decode processes a batch of tokens
func (c *Context) Decode(batch *Batch) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	ret := C.llama_decode(c.ctx, batch.batch)
	if ret != 0 {
		return fmt.Errorf("decode failed with code %d", ret)
	}
	return nil
}

// Encode processes a batch for encoder models
func (c *Context) Encode(batch *Batch) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	ret := C.llama_encode(c.ctx, batch.batch)
	if ret != 0 {
		return fmt.Errorf("encode failed with code %d", ret)
	}
	return nil
}

// =============================================================================
// LLAMA LOGITS AND EMBEDDINGS
// =============================================================================

// GetLogits returns the logits for the last token
func (c *Context) GetLogits() []float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ptr := C.llama_get_logits(c.ctx)
	if ptr == nil {
		return nil
	}
	
	nVocab := c.model.NVocab()
	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), nVocab)
}

// GetLogitsIth returns the logits for token at index i
func (c *Context) GetLogitsIth(i int32) []float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ptr := C.llama_get_logits_ith(c.ctx, C.int32_t(i))
	if ptr == nil {
		return nil
	}
	
	nVocab := c.model.NVocab()
	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), nVocab)
}

// GetEmbeddings returns the embeddings
func (c *Context) GetEmbeddings() []float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ptr := C.llama_get_embeddings(c.ctx)
	if ptr == nil {
		return nil
	}
	
	nEmbd := c.model.NEmbd()
	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), nEmbd)
}

// GetEmbeddingsIth returns the embeddings for token at index i
func (c *Context) GetEmbeddingsIth(i int32) []float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	ptr := C.llama_get_embeddings_ith(c.ctx, C.int32_t(i))
	if ptr == nil {
		return nil
	}
	
	nEmbd := c.model.NEmbd()
	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), nEmbd)
}

// =============================================================================
// LLAMA KV CACHE
// =============================================================================

// KVCacheClear clears the KV cache
func (c *Context) KVCacheClear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_clear(c.ctx)
}

// KVCacheSeqRm removes a sequence from the KV cache
func (c *Context) KVCacheSeqRm(seqID SeqID, p0, p1 Pos) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return bool(C.llama_kv_cache_seq_rm(c.ctx, C.llama_seq_id(seqID), C.llama_pos(p0), C.llama_pos(p1)))
}

// KVCacheSeqCp copies a sequence in the KV cache
func (c *Context) KVCacheSeqCp(seqIDSrc, seqIDDst SeqID, p0, p1 Pos) {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_seq_cp(c.ctx, C.llama_seq_id(seqIDSrc), C.llama_seq_id(seqIDDst), C.llama_pos(p0), C.llama_pos(p1))
}

// KVCacheSeqKeep keeps only the specified sequence
func (c *Context) KVCacheSeqKeep(seqID SeqID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_seq_keep(c.ctx, C.llama_seq_id(seqID))
}

// KVCacheSeqAdd adds a position delta to a sequence
func (c *Context) KVCacheSeqAdd(seqID SeqID, p0, p1, delta Pos) {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_seq_add(c.ctx, C.llama_seq_id(seqID), C.llama_pos(p0), C.llama_pos(p1), C.llama_pos(delta))
}

// KVCacheSeqPosMax returns the maximum position in a sequence
func (c *Context) KVCacheSeqPosMax(seqID SeqID) Pos {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Pos(C.llama_kv_cache_seq_pos_max(c.ctx, C.llama_seq_id(seqID)))
}

// KVCacheDefrag defragments the KV cache
func (c *Context) KVCacheDefrag() {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_defrag(c.ctx)
}

// KVCacheUpdate updates the KV cache
func (c *Context) KVCacheUpdate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	C.llama_kv_cache_update(c.ctx)
}

// KVCacheTokenCount returns the number of tokens in the KV cache
func (c *Context) KVCacheTokenCount() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int32(C.llama_get_kv_cache_token_count(c.ctx))
}

// KVCacheUsedCells returns the number of used cells in the KV cache
func (c *Context) KVCacheUsedCells() int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int32(C.llama_get_kv_cache_used_cells(c.ctx))
}

// =============================================================================
// LLAMA SAMPLER
// =============================================================================

// Sampler represents a token sampler
type Sampler struct {
	sampler *C.struct_llama_sampler
}

// NewSamplerGreedy creates a greedy sampler
func NewSamplerGreedy() *Sampler {
	s := C.llama_sampler_init_greedy()
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerDist creates a distribution sampler
func NewSamplerDist(seed uint32) *Sampler {
	s := C.llama_sampler_init_dist(C.uint32_t(seed))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerTopK creates a top-k sampler
func NewSamplerTopK(k int32) *Sampler {
	s := C.llama_sampler_init_top_k(C.int32_t(k))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerTopP creates a top-p (nucleus) sampler
func NewSamplerTopP(p float32, minKeep int) *Sampler {
	s := C.llama_sampler_init_top_p(C.float(p), C.size_t(minKeep))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerMinP creates a min-p sampler
func NewSamplerMinP(p float32, minKeep int) *Sampler {
	s := C.llama_sampler_init_min_p(C.float(p), C.size_t(minKeep))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerTypical creates a typical sampler
func NewSamplerTypical(p float32, minKeep int) *Sampler {
	s := C.llama_sampler_init_typical(C.float(p), C.size_t(minKeep))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerTemp creates a temperature sampler
func NewSamplerTemp(t float32) *Sampler {
	s := C.llama_sampler_init_temp(C.float(t))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerMirostat creates a Mirostat sampler
func NewSamplerMirostat(nVocab int32, seed uint32, tau, eta float32, m int32) *Sampler {
	s := C.llama_sampler_init_mirostat(C.int32_t(nVocab), C.uint32_t(seed), C.float(tau), C.float(eta), C.int32_t(m))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerMirostatV2 creates a Mirostat v2 sampler
func NewSamplerMirostatV2(seed uint32, tau, eta float32) *Sampler {
	s := C.llama_sampler_init_mirostat_v2(C.uint32_t(seed), C.float(tau), C.float(eta))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// NewSamplerPenalties creates a penalties sampler
func NewSamplerPenalties(penaltyLastN int32, penaltyRepeat, penaltyFreq, penaltyPresent float32) *Sampler {
	s := C.llama_sampler_init_penalties(
		C.int32_t(penaltyLastN),
		C.float(penaltyRepeat),
		C.float(penaltyFreq),
		C.float(penaltyPresent))
	sampler := &Sampler{sampler: s}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// Free releases the sampler resources
func (s *Sampler) Free() {
	if s.sampler != nil {
		C.llama_sampler_free(s.sampler)
		s.sampler = nil
	}
}

// Reset resets the sampler state
func (s *Sampler) Reset() {
	C.llama_sampler_reset(s.sampler)
}

// Clone creates a copy of the sampler
func (s *Sampler) Clone() *Sampler {
	clone := C.llama_sampler_clone(s.sampler)
	sampler := &Sampler{sampler: clone}
	runtime.SetFinalizer(sampler, (*Sampler).Free)
	return sampler
}

// Name returns the sampler name
func (s *Sampler) Name() string {
	return C.GoString(C.llama_sampler_name(s.sampler))
}

// Accept accepts a token
func (s *Sampler) Accept(token Token) {
	C.llama_sampler_accept(s.sampler, C.llama_token(token))
}

// Sample samples a token from the context
func (s *Sampler) Sample(ctx *Context, idx int32) Token {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return Token(C.llama_sampler_sample(s.sampler, ctx.ctx, C.int32_t(idx)))
}

// =============================================================================
// LLAMA SAMPLER CHAIN
// =============================================================================

// SamplerChain represents a chain of samplers
type SamplerChain struct {
	chain    *C.struct_llama_sampler
	samplers []*Sampler // Keep references to prevent GC
}

// NewSamplerChain creates a new sampler chain
func NewSamplerChain(noPerf bool) *SamplerChain {
	params := C.llama_sampler_chain_default_params()
	params.no_perf = C.bool(noPerf)
	
	chain := C.llama_sampler_chain_init(params)
	sc := &SamplerChain{chain: chain, samplers: make([]*Sampler, 0)}
	runtime.SetFinalizer(sc, (*SamplerChain).Free)
	return sc
}

// Free releases the chain resources
func (sc *SamplerChain) Free() {
	if sc.chain != nil {
		C.llama_sampler_free(sc.chain)
		sc.chain = nil
		sc.samplers = nil
	}
}

// Add adds a sampler to the chain
func (sc *SamplerChain) Add(s *Sampler) {
	C.llama_sampler_chain_add(sc.chain, s.sampler)
	sc.samplers = append(sc.samplers, s)
	// Prevent the sampler from being freed individually
	runtime.SetFinalizer(s, nil)
}

// N returns the number of samplers in the chain
func (sc *SamplerChain) N() int32 {
	return int32(C.llama_sampler_chain_n(sc.chain))
}

// Sample samples a token using the chain
func (sc *SamplerChain) Sample(ctx *Context, idx int32) Token {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return Token(C.llama_sampler_sample(sc.chain, ctx.ctx, C.int32_t(idx)))
}

// Accept accepts a token
func (sc *SamplerChain) Accept(token Token) {
	C.llama_sampler_accept(sc.chain, C.llama_token(token))
}

// Reset resets all samplers in the chain
func (sc *SamplerChain) Reset() {
	C.llama_sampler_reset(sc.chain)
}

// =============================================================================
// LLAMA STATE
// =============================================================================

// GetStateSize returns the size of the context state
func (c *Context) GetStateSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int(C.llama_get_state_size(c.ctx))
}

// CopyStateData copies the context state to a buffer
func (c *Context) CopyStateData(dst []byte) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return int(C.llama_copy_state_data(c.ctx, (*C.uint8_t)(unsafe.Pointer(&dst[0]))))
}

// SetStateData restores the context state from a buffer
func (c *Context) SetStateData(src []byte) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return int(C.llama_set_state_data(c.ctx, (*C.uint8_t)(unsafe.Pointer(&src[0]))))
}

// =============================================================================
// CHAT TEMPLATE
// =============================================================================

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string
	Content string
}

// ApplyChatTemplate applies a chat template to messages
func ApplyChatTemplate(template string, messages []ChatMessage, addAss bool) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}
	
	// Convert messages to C structs
	cMessages := make([]C.struct_llama_chat_message, len(messages))
	cStrings := make([]*C.char, len(messages)*2) // Keep references
	
	for i, msg := range messages {
		cStrings[i*2] = C.CString(msg.Role)
		cStrings[i*2+1] = C.CString(msg.Content)
		cMessages[i].role = cStrings[i*2]
		cMessages[i].content = cStrings[i*2+1]
	}
	
	defer func() {
		for _, s := range cStrings {
			C.free(unsafe.Pointer(s))
		}
	}()
	
	var cTmpl *C.char
	if template != "" {
		cTmpl = C.CString(template)
		defer C.free(unsafe.Pointer(cTmpl))
	}
	
	// First call to get required size
	n := C.llama_chat_apply_template(cTmpl, &cMessages[0], C.size_t(len(messages)), C.bool(addAss), nil, 0)
	if n < 0 {
		return "", fmt.Errorf("failed to apply chat template: %d", n)
	}
	
	buf := make([]byte, n+1)
	n = C.llama_chat_apply_template(cTmpl, &cMessages[0], C.size_t(len(messages)), C.bool(addAss),
		(*C.char)(unsafe.Pointer(&buf[0])), C.int32_t(len(buf)))
	
	if n < 0 {
		return "", fmt.Errorf("failed to apply chat template: %d", n)
	}
	
	return string(buf[:n]), nil
}
