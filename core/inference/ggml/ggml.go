// Package ggml provides Go bindings for the GGML tensor library
// GGML is the foundation for llama.cpp and other efficient inference engines
package ggml

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR}/../../../libs -lggml-base -lggml -lggml-cpu -lm -lpthread
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../libs/arm64-v8a

#include "ggml.h"
#include <stdlib.h>
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
// GGML TYPES
// =============================================================================

// Type represents GGML tensor data types
type Type int

const (
	TypeF32     Type = C.GGML_TYPE_F32
	TypeF16     Type = C.GGML_TYPE_F16
	TypeQ4_0    Type = C.GGML_TYPE_Q4_0
	TypeQ4_1    Type = C.GGML_TYPE_Q4_1
	TypeQ5_0    Type = C.GGML_TYPE_Q5_0
	TypeQ5_1    Type = C.GGML_TYPE_Q5_1
	TypeQ8_0    Type = C.GGML_TYPE_Q8_0
	TypeQ8_1    Type = C.GGML_TYPE_Q8_1
	TypeQ2_K    Type = C.GGML_TYPE_Q2_K
	TypeQ3_K    Type = C.GGML_TYPE_Q3_K
	TypeQ4_K    Type = C.GGML_TYPE_Q4_K
	TypeQ5_K    Type = C.GGML_TYPE_Q5_K
	TypeQ6_K    Type = C.GGML_TYPE_Q6_K
	TypeQ8_K    Type = C.GGML_TYPE_Q8_K
	TypeI8      Type = C.GGML_TYPE_I8
	TypeI16     Type = C.GGML_TYPE_I16
	TypeI32     Type = C.GGML_TYPE_I32
	TypeI64     Type = C.GGML_TYPE_I64
	TypeF64     Type = C.GGML_TYPE_F64
	TypeBF16    Type = C.GGML_TYPE_BF16
)

// String returns the name of the type
func (t Type) String() string {
	return C.GoString(C.ggml_type_name(C.enum_ggml_type(t)))
}

// Size returns the size in bytes of a single element of this type
func (t Type) Size() int {
	return int(C.ggml_type_size(C.enum_ggml_type(t)))
}

// BackendType represents GGML backend types
type BackendType int

const (
	BackendTypeCPU      BackendType = C.GGML_BACKEND_TYPE_CPU
	BackendTypeGPU      BackendType = C.GGML_BACKEND_TYPE_GPU
	BackendTypeGPUSplit BackendType = C.GGML_BACKEND_TYPE_GPU_SPLIT
)

// Op represents GGML operations
type Op int

const (
	OpNone       Op = C.GGML_OP_NONE
	OpDup        Op = C.GGML_OP_DUP
	OpAdd        Op = C.GGML_OP_ADD
	OpSub        Op = C.GGML_OP_SUB
	OpMul        Op = C.GGML_OP_MUL
	OpDiv        Op = C.GGML_OP_DIV
	OpSqr        Op = C.GGML_OP_SQR
	OpSqrt       Op = C.GGML_OP_SQRT
	OpLog        Op = C.GGML_OP_LOG
	OpSin        Op = C.GGML_OP_SIN
	OpCos        Op = C.GGML_OP_COS
	OpSum        Op = C.GGML_OP_SUM
	OpMean       Op = C.GGML_OP_MEAN
	OpArgmax     Op = C.GGML_OP_ARGMAX
	OpNorm       Op = C.GGML_OP_NORM
	OpRMSNorm    Op = C.GGML_OP_RMS_NORM
	OpMulMat     Op = C.GGML_OP_MUL_MAT
	OpSoftMax    Op = C.GGML_OP_SOFT_MAX
	OpRope       Op = C.GGML_OP_ROPE
	OpFlashAttn  Op = C.GGML_OP_FLASH_ATTN_EXT
)

// String returns the name of the operation
func (o Op) String() string {
	return C.GoString(C.ggml_op_name(C.enum_ggml_op(o)))
}

// Status represents GGML operation status
type Status int

const (
	StatusAllocFailed Status = C.GGML_STATUS_ALLOC_FAILED
	StatusFailed      Status = C.GGML_STATUS_FAILED
	StatusSuccess     Status = C.GGML_STATUS_SUCCESS
	StatusAborted     Status = C.GGML_STATUS_ABORTED
)

// =============================================================================
// GGML CONTEXT
// =============================================================================

// Context represents a GGML computation context
type Context struct {
	ctx *C.struct_ggml_context
	mu  sync.RWMutex
}

// InitParams configures context initialization
type InitParams struct {
	MemSize   int  // Memory size in bytes
	MemBuffer []byte // Optional pre-allocated buffer
	NoAlloc   bool // Don't allocate tensor data
}

// NewContext creates a new GGML context
func NewContext(params InitParams) (*Context, error) {
	cParams := C.struct_ggml_init_params{
		mem_size:   C.size_t(params.MemSize),
		mem_buffer: nil,
		no_alloc:   C.bool(params.NoAlloc),
	}
	
	if len(params.MemBuffer) > 0 {
		cParams.mem_buffer = unsafe.Pointer(&params.MemBuffer[0])
	}
	
	ctx := C.ggml_init(cParams)
	if ctx == nil {
		return nil, errors.New("failed to initialize GGML context")
	}
	
	c := &Context{ctx: ctx}
	runtime.SetFinalizer(c, (*Context).Free)
	return c, nil
}

// Free releases the context resources
func (c *Context) Free() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.ctx != nil {
		C.ggml_free(c.ctx)
		c.ctx = nil
	}
}

// UsedMem returns the amount of memory used by the context
func (c *Context) UsedMem() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int(C.ggml_used_mem(c.ctx))
}

// =============================================================================
// GGML TENSOR
// =============================================================================

// Tensor represents a GGML tensor
type Tensor struct {
	tensor *C.struct_ggml_tensor
	ctx    *Context
}

// NewTensor1D creates a 1D tensor
func (c *Context) NewTensor1D(dtype Type, ne0 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	t := C.ggml_new_tensor_1d(c.ctx, C.enum_ggml_type(dtype), C.int64_t(ne0))
	return &Tensor{tensor: t, ctx: c}
}

// NewTensor2D creates a 2D tensor
func (c *Context) NewTensor2D(dtype Type, ne0, ne1 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	t := C.ggml_new_tensor_2d(c.ctx, C.enum_ggml_type(dtype), C.int64_t(ne0), C.int64_t(ne1))
	return &Tensor{tensor: t, ctx: c}
}

// NewTensor3D creates a 3D tensor
func (c *Context) NewTensor3D(dtype Type, ne0, ne1, ne2 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	t := C.ggml_new_tensor_3d(c.ctx, C.enum_ggml_type(dtype), C.int64_t(ne0), C.int64_t(ne1), C.int64_t(ne2))
	return &Tensor{tensor: t, ctx: c}
}

// NewTensor4D creates a 4D tensor
func (c *Context) NewTensor4D(dtype Type, ne0, ne1, ne2, ne3 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	t := C.ggml_new_tensor_4d(c.ctx, C.enum_ggml_type(dtype), C.int64_t(ne0), C.int64_t(ne1), C.int64_t(ne2), C.int64_t(ne3))
	return &Tensor{tensor: t, ctx: c}
}

// NElements returns the total number of elements in the tensor
func (t *Tensor) NElements() int64 {
	return int64(C.ggml_nelements(t.tensor))
}

// NRows returns the number of rows in the tensor
func (t *Tensor) NRows() int64 {
	return int64(C.ggml_nrows(t.tensor))
}

// NBytes returns the size in bytes of the tensor data
func (t *Tensor) NBytes() int {
	return int(C.ggml_nbytes(t.tensor))
}

// NDims returns the number of dimensions
func (t *Tensor) NDims() int {
	return int(C.ggml_n_dims(t.tensor))
}

// Data returns a pointer to the tensor data
func (t *Tensor) Data() unsafe.Pointer {
	return C.ggml_get_data(t.tensor)
}

// DataF32 returns the tensor data as a float32 slice
func (t *Tensor) DataF32() []float32 {
	ptr := C.ggml_get_data_f32(t.tensor)
	if ptr == nil {
		return nil
	}
	n := t.NElements()
	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), n)
}

// SetF32 sets all elements to a value
func (t *Tensor) SetF32(value float32) {
	C.ggml_set_f32(t.tensor, C.float(value))
}

// GetF32_1D gets a single float32 value at index i
func (t *Tensor) GetF32_1D(i int) float32 {
	return float32(C.ggml_get_f32_1d(t.tensor, C.int(i)))
}

// SetF32_1D sets a single float32 value at index i
func (t *Tensor) SetF32_1D(i int, value float32) {
	C.ggml_set_f32_1d(t.tensor, C.int(i), C.float(value))
}

// SameShape checks if two tensors have the same shape
func (t *Tensor) SameShape(other *Tensor) bool {
	return bool(C.ggml_are_same_shape(t.tensor, other.tensor))
}

// =============================================================================
// TENSOR OPERATIONS
// =============================================================================

// Add performs element-wise addition: a + b
func (c *Context) Add(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_add(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Sub performs element-wise subtraction: a - b
func (c *Context) Sub(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_sub(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Mul performs element-wise multiplication: a * b
func (c *Context) Mul(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_mul(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Div performs element-wise division: a / b
func (c *Context) Div(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_div(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Sqr computes element-wise square: a^2
func (c *Context) Sqr(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_sqr(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Sqrt computes element-wise square root
func (c *Context) Sqrt(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_sqrt(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Log computes element-wise natural logarithm
func (c *Context) Log(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_log(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Sin computes element-wise sine
func (c *Context) Sin(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_sin(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Cos computes element-wise cosine
func (c *Context) Cos(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_cos(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Sum computes the sum of all elements
func (c *Context) Sum(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_sum(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Mean computes the mean of all elements
func (c *Context) Mean(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_mean(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Argmax returns the index of the maximum element
func (c *Context) Argmax(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_argmax(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Abs computes element-wise absolute value
func (c *Context) Abs(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_abs(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// MATRIX OPERATIONS
// =============================================================================

// MulMat performs matrix multiplication: a @ b
func (c *Context) MulMat(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_mul_mat(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// OutProd computes outer product
func (c *Context) OutProd(a, b *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_out_prod(c.ctx, a.tensor, b.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// NORMALIZATION
// =============================================================================

// Norm computes layer normalization
func (c *Context) Norm(a *Tensor, eps float32) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_norm(c.ctx, a.tensor, C.float(eps))
	return &Tensor{tensor: t, ctx: c}
}

// RMSNorm computes RMS normalization
func (c *Context) RMSNorm(a *Tensor, eps float32) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_rms_norm(c.ctx, a.tensor, C.float(eps))
	return &Tensor{tensor: t, ctx: c}
}

// GroupNorm computes group normalization
func (c *Context) GroupNorm(a *Tensor, nGroups int, eps float32) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_group_norm(c.ctx, a.tensor, C.int(nGroups), C.float(eps))
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// ACTIVATION FUNCTIONS
// =============================================================================

// Relu computes ReLU activation
func (c *Context) Relu(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_relu(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Gelu computes GELU activation
func (c *Context) Gelu(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_gelu(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Silu computes SiLU (Swish) activation
func (c *Context) Silu(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_silu(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// LeakyRelu computes Leaky ReLU activation
func (c *Context) LeakyRelu(a *Tensor, negativeSlope float32, inplace bool) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_leaky_relu(c.ctx, a.tensor, C.float(negativeSlope), C.bool(inplace))
	return &Tensor{tensor: t, ctx: c}
}

// SoftMax computes softmax
func (c *Context) SoftMax(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_soft_max(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// TENSOR MANIPULATION
// =============================================================================

// Reshape1D reshapes tensor to 1D
func (c *Context) Reshape1D(a *Tensor, ne0 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_reshape_1d(c.ctx, a.tensor, C.int64_t(ne0))
	return &Tensor{tensor: t, ctx: c}
}

// Reshape2D reshapes tensor to 2D
func (c *Context) Reshape2D(a *Tensor, ne0, ne1 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_reshape_2d(c.ctx, a.tensor, C.int64_t(ne0), C.int64_t(ne1))
	return &Tensor{tensor: t, ctx: c}
}

// Reshape3D reshapes tensor to 3D
func (c *Context) Reshape3D(a *Tensor, ne0, ne1, ne2 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_reshape_3d(c.ctx, a.tensor, C.int64_t(ne0), C.int64_t(ne1), C.int64_t(ne2))
	return &Tensor{tensor: t, ctx: c}
}

// Reshape4D reshapes tensor to 4D
func (c *Context) Reshape4D(a *Tensor, ne0, ne1, ne2, ne3 int64) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_reshape_4d(c.ctx, a.tensor, C.int64_t(ne0), C.int64_t(ne1), C.int64_t(ne2), C.int64_t(ne3))
	return &Tensor{tensor: t, ctx: c}
}

// View1D creates a 1D view of a tensor
func (c *Context) View1D(a *Tensor, ne0 int64, offset int) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_view_1d(c.ctx, a.tensor, C.int64_t(ne0), C.size_t(offset))
	return &Tensor{tensor: t, ctx: c}
}

// Permute permutes tensor dimensions
func (c *Context) Permute(a *Tensor, axis0, axis1, axis2, axis3 int) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_permute(c.ctx, a.tensor, C.int(axis0), C.int(axis1), C.int(axis2), C.int(axis3))
	return &Tensor{tensor: t, ctx: c}
}

// Transpose transposes the tensor
func (c *Context) Transpose(a *Tensor) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_transpose(c.ctx, a.tensor)
	return &Tensor{tensor: t, ctx: c}
}

// Concat concatenates tensors along a dimension
func (c *Context) Concat(a, b *Tensor, dim int) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_concat(c.ctx, a.tensor, b.tensor, C.int(dim))
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// POSITIONAL ENCODING
// =============================================================================

// Rope applies rotary positional embedding
func (c *Context) Rope(a, pos *Tensor, nDims, mode int) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_rope(c.ctx, a.tensor, pos.tensor, C.int(nDims), C.int(mode))
	return &Tensor{tensor: t, ctx: c}
}

// Arange creates a tensor with values from start to stop with step
func (c *Context) Arange(start, stop, step float32) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := C.ggml_arange(c.ctx, C.float(start), C.float(stop), C.float(step))
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// ATTENTION
// =============================================================================

// FlashAttention computes flash attention
func (c *Context) FlashAttention(q, k, v, mask *Tensor, scale, maxBias, logitSoftcap float32) *Tensor {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var maskTensor *C.struct_ggml_tensor
	if mask != nil {
		maskTensor = mask.tensor
	}
	
	t := C.ggml_flash_attn_ext(c.ctx, q.tensor, k.tensor, v.tensor, maskTensor,
		C.float(scale), C.float(maxBias), C.float(logitSoftcap))
	return &Tensor{tensor: t, ctx: c}
}

// =============================================================================
// COMPUTATION GRAPH
// =============================================================================

// Graph represents a GGML computation graph
type Graph struct {
	graph *C.struct_ggml_cgraph
	ctx   *Context
}

// NewGraph creates a new computation graph
func (c *Context) NewGraph() *Graph {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := C.ggml_new_graph(c.ctx)
	return &Graph{graph: g, ctx: c}
}

// NewGraphCustom creates a new computation graph with custom size
func (c *Context) NewGraphCustom(size int, grads bool) *Graph {
	c.mu.Lock()
	defer c.mu.Unlock()
	g := C.ggml_new_graph_custom(c.ctx, C.size_t(size), C.bool(grads))
	return &Graph{graph: g, ctx: c}
}

// BuildForward adds a tensor to the graph for forward computation
func (g *Graph) BuildForward(tensor *Tensor) {
	g.ctx.mu.Lock()
	defer g.ctx.mu.Unlock()
	C.ggml_build_forward_expand(g.graph, tensor.tensor)
}

// NNodes returns the number of nodes in the graph
func (g *Graph) NNodes() int {
	return int(C.ggml_graph_n_nodes(g.graph))
}

// =============================================================================
// BACKEND
// =============================================================================

// Backend represents a GGML compute backend
type Backend struct {
	backend *C.struct_ggml_backend
	name    string
}

// BackendRegCount returns the number of registered backends
func BackendRegCount() int {
	return int(C.ggml_backend_reg_count())
}

// BackendDevCount returns the number of available backend devices
func BackendDevCount() int {
	return int(C.ggml_backend_dev_count())
}

// LoadBackend loads a backend from a shared library path
func LoadBackend(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	
	if !C.ggml_backend_load(cPath) {
		return fmt.Errorf("failed to load backend from %s", path)
	}
	return nil
}

// LoadAllBackends loads all available backends
func LoadAllBackends() int {
	return int(C.ggml_backend_load_all())
}

// LoadAllBackendsFromPath loads all backends from a directory
func LoadAllBackendsFromPath(path string) int {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	return int(C.ggml_backend_load_all_from_path(cPath))
}

// InitBackendByName initializes a backend by name
func InitBackendByName(name, params string) (*Backend, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	
	var cParams *C.char
	if params != "" {
		cParams = C.CString(params)
		defer C.free(unsafe.Pointer(cParams))
	}
	
	b := C.ggml_backend_init_by_name(cName, cParams)
	if b == nil {
		return nil, fmt.Errorf("failed to initialize backend: %s", name)
	}
	
	return &Backend{backend: b, name: name}, nil
}

// InitBackendByType initializes a backend by type
func InitBackendByType(btype BackendType, params string) (*Backend, error) {
	var cParams *C.char
	if params != "" {
		cParams = C.CString(params)
		defer C.free(unsafe.Pointer(cParams))
	}
	
	b := C.ggml_backend_init_by_type(C.enum_ggml_backend_type(btype), cParams)
	if b == nil {
		return nil, fmt.Errorf("failed to initialize backend type: %d", btype)
	}
	
	return &Backend{backend: b, name: fmt.Sprintf("type_%d", btype)}, nil
}

// InitBestBackend initializes the best available backend
func InitBestBackend() (*Backend, error) {
	b := C.ggml_backend_init_best()
	if b == nil {
		return nil, errors.New("failed to initialize best backend")
	}
	return &Backend{backend: b, name: "best"}, nil
}

// =============================================================================
// BACKEND BUFFER
// =============================================================================

// Buffer represents a GGML backend buffer
type Buffer struct {
	buffer *C.struct_ggml_backend_buffer
}

// AllocBuffer allocates a buffer on the backend
func (b *Backend) AllocBuffer(size int) (*Buffer, error) {
	buf := C.ggml_backend_alloc_buffer(b.backend, C.size_t(size))
	if buf == nil {
		return nil, fmt.Errorf("failed to allocate buffer of size %d", size)
	}
	return &Buffer{buffer: buf}, nil
}

// Free releases the buffer
func (buf *Buffer) Free() {
	if buf.buffer != nil {
		C.ggml_backend_buffer_free(buf.buffer)
		buf.buffer = nil
	}
}

// Size returns the buffer size
func (buf *Buffer) Size() int {
	return int(C.ggml_backend_buffer_get_size(buf.buffer))
}

// Name returns the buffer name
func (buf *Buffer) Name() string {
	return C.GoString(C.ggml_backend_buffer_name(buf.buffer))
}

// Alignment returns the buffer alignment
func (buf *Buffer) Alignment() int {
	return int(C.ggml_backend_buffer_get_alignment(buf.buffer))
}

// IsHost returns true if the buffer is on the host
func (buf *Buffer) IsHost() bool {
	return bool(C.ggml_backend_buffer_is_host(buf.buffer))
}

// Clear clears the buffer with a value
func (buf *Buffer) Clear(value byte) {
	C.ggml_backend_buffer_clear(buf.buffer, C.uint8_t(value))
}

// =============================================================================
// MEMORY UTILITIES
// =============================================================================

// AlignedMalloc allocates aligned memory
func AlignedMalloc(size int) unsafe.Pointer {
	return C.ggml_aligned_malloc(C.size_t(size))
}

// AlignedFree frees aligned memory
func AlignedFree(ptr unsafe.Pointer) {
	C.ggml_aligned_free(ptr)
}
