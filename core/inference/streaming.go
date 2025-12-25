// Package inference provides token streaming for real-time output
// This implements streaming output for the echobeats inference engine
package inference

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// TOKEN STREAM
// =============================================================================

// Token represents a single generated token
type Token struct {
	ID        int32     // Token ID
	Text      string    // Decoded text
	Logprob   float32   // Log probability
	TopLogprobs []TokenLogprob // Top alternative tokens
	Timestamp time.Time // Generation timestamp
	StreamID  StreamID  // Source stream
	Step      int       // Cognitive step
	Position  int       // Position in sequence
	IsSpecial bool      // Whether this is a special token (BOS, EOS, etc.)
	IsFinal   bool      // Whether this is the final token
}

// TokenLogprob represents a token with its log probability
type TokenLogprob struct {
	ID      int32
	Text    string
	Logprob float32
}

// TokenStream represents a stream of tokens
type TokenStream struct {
	tokens    chan *Token
	done      chan struct{}
	err       atomic.Value
	closed    atomic.Bool
	mu        sync.RWMutex
	
	// Metadata
	streamID  StreamID
	step      int
	startTime time.Time
	
	// Statistics
	tokenCount int64
	totalTime  time.Duration
}

// NewTokenStream creates a new token stream
func NewTokenStream(streamID StreamID, step int, bufferSize int) *TokenStream {
	if bufferSize <= 0 {
		bufferSize = 256
	}
	return &TokenStream{
		tokens:    make(chan *Token, bufferSize),
		done:      make(chan struct{}),
		streamID:  streamID,
		step:      step,
		startTime: time.Now(),
	}
}

// Send sends a token to the stream
func (ts *TokenStream) Send(token *Token) error {
	if ts.closed.Load() {
		return errors.New("stream closed")
	}
	
	token.StreamID = ts.streamID
	token.Step = ts.step
	token.Timestamp = time.Now()
	
	select {
	case ts.tokens <- token:
		atomic.AddInt64(&ts.tokenCount, 1)
		return nil
	case <-ts.done:
		return errors.New("stream done")
	}
}

// Recv receives a token from the stream
func (ts *TokenStream) Recv() (*Token, error) {
	select {
	case token, ok := <-ts.tokens:
		if !ok {
			if err := ts.Error(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		return token, nil
	case <-ts.done:
		if err := ts.Error(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
}

// RecvWithTimeout receives a token with a timeout
func (ts *TokenStream) RecvWithTimeout(timeout time.Duration) (*Token, error) {
	select {
	case token, ok := <-ts.tokens:
		if !ok {
			if err := ts.Error(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		return token, nil
	case <-ts.done:
		if err := ts.Error(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	case <-time.After(timeout):
		return nil, errors.New("timeout waiting for token")
	}
}

// Close closes the stream
func (ts *TokenStream) Close() {
	if ts.closed.Swap(true) {
		return
	}
	close(ts.done)
	close(ts.tokens)
	ts.totalTime = time.Since(ts.startTime)
}

// CloseWithError closes the stream with an error
func (ts *TokenStream) CloseWithError(err error) {
	ts.err.Store(err)
	ts.Close()
}

// Error returns any error that occurred
func (ts *TokenStream) Error() error {
	if err := ts.err.Load(); err != nil {
		return err.(error)
	}
	return nil
}

// IsClosed returns whether the stream is closed
func (ts *TokenStream) IsClosed() bool {
	return ts.closed.Load()
}

// TokenCount returns the number of tokens sent
func (ts *TokenStream) TokenCount() int64 {
	return atomic.LoadInt64(&ts.tokenCount)
}

// TokensPerSecond returns the generation speed
func (ts *TokenStream) TokensPerSecond() float64 {
	count := atomic.LoadInt64(&ts.tokenCount)
	if count == 0 {
		return 0
	}
	
	var elapsed time.Duration
	if ts.closed.Load() {
		elapsed = ts.totalTime
	} else {
		elapsed = time.Since(ts.startTime)
	}
	
	if elapsed == 0 {
		return 0
	}
	
	return float64(count) / elapsed.Seconds()
}

// =============================================================================
// STREAM MULTIPLEXER
// =============================================================================

// StreamMultiplexer combines multiple token streams
type StreamMultiplexer struct {
	streams   []*TokenStream
	combined  chan *Token
	done      chan struct{}
	closed    atomic.Bool
	wg        sync.WaitGroup
}

// NewStreamMultiplexer creates a multiplexer for multiple streams
func NewStreamMultiplexer(streams ...*TokenStream) *StreamMultiplexer {
	sm := &StreamMultiplexer{
		streams:  streams,
		combined: make(chan *Token, 256),
		done:     make(chan struct{}),
	}
	
	// Start goroutines to forward from each stream
	for _, stream := range streams {
		sm.wg.Add(1)
		go sm.forward(stream)
	}
	
	// Close combined channel when all streams are done
	go func() {
		sm.wg.Wait()
		close(sm.combined)
	}()
	
	return sm
}

func (sm *StreamMultiplexer) forward(stream *TokenStream) {
	defer sm.wg.Done()
	
	for {
		select {
		case <-sm.done:
			return
		default:
			token, err := stream.Recv()
			if err != nil {
				return
			}
			
			select {
			case sm.combined <- token:
			case <-sm.done:
				return
			}
		}
	}
}

// Recv receives the next token from any stream
func (sm *StreamMultiplexer) Recv() (*Token, error) {
	select {
	case token, ok := <-sm.combined:
		if !ok {
			return nil, io.EOF
		}
		return token, nil
	case <-sm.done:
		return nil, io.EOF
	}
}

// Close closes the multiplexer
func (sm *StreamMultiplexer) Close() {
	if sm.closed.Swap(true) {
		return
	}
	close(sm.done)
}

// =============================================================================
// STREAMING INFERENCE
// =============================================================================

// StreamingRequest extends InferenceRequest for streaming
type StreamingRequest struct {
	*InferenceRequest
	
	// Streaming options
	StreamTokens    bool // Enable token streaming
	StreamLogprobs  bool // Include logprobs in stream
	TopLogprobs     int  // Number of top logprobs to include
	StopSequences   []string // Sequences that stop generation
	
	// Callbacks
	OnToken func(*Token) // Called for each token
	OnDone  func()       // Called when generation completes
	OnError func(error)  // Called on error
}

// StreamingResponse wraps the token stream
type StreamingResponse struct {
	Stream       *TokenStream
	RequestID    string
	StreamID     StreamID
	Step         int
	StartTime    time.Time
	
	// Accumulated output
	mu           sync.RWMutex
	tokens       []*Token
	text         string
}

// NewStreamingResponse creates a new streaming response
func NewStreamingResponse(streamID StreamID, step int) *StreamingResponse {
	return &StreamingResponse{
		Stream:    NewTokenStream(streamID, step, 256),
		StreamID:  streamID,
		Step:      step,
		StartTime: time.Now(),
		tokens:    make([]*Token, 0, 256),
	}
}

// AccumulateToken adds a token to the accumulated output
func (sr *StreamingResponse) AccumulateToken(token *Token) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.tokens = append(sr.tokens, token)
	sr.text += token.Text
}

// GetText returns the accumulated text
func (sr *StreamingResponse) GetText() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.text
}

// GetTokens returns all accumulated tokens
func (sr *StreamingResponse) GetTokens() []*Token {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	result := make([]*Token, len(sr.tokens))
	copy(result, sr.tokens)
	return result
}

// ToInferenceResponse converts to a standard response
func (sr *StreamingResponse) ToInferenceResponse() *InferenceResponse {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	tokenIDs := make([]int32, len(sr.tokens))
	for i, t := range sr.tokens {
		tokenIDs[i] = t.ID
	}
	
	return &InferenceResponse{
		StreamID:     sr.StreamID,
		Step:         sr.Step,
		Output:       sr.text,
		Tokens:       tokenIDs,
		LatencyMs:    time.Since(sr.StartTime).Milliseconds(),
		TokensPerSec: sr.Stream.TokensPerSecond(),
	}
}

// =============================================================================
// STREAMING ENGINE INTERFACE
// =============================================================================

// StreamingEngine extends InferenceEngine with streaming support
type StreamingEngine interface {
	InferenceEngine
	
	// InferStream performs streaming inference
	InferStream(ctx context.Context, req *StreamingRequest) (*StreamingResponse, error)
	
	// InferStreamAsync performs async streaming inference
	InferStreamAsync(ctx context.Context, req *StreamingRequest, callback func(*Token)) error
}

// =============================================================================
// STREAMING ECHOBEATS ENGINE
// =============================================================================

// StreamingEchobeatsEngine adds streaming to EchobeatsEngine
type StreamingEchobeatsEngine struct {
	*EchobeatsEngine
	
	// Active streams
	activeStreams   map[string]*StreamingResponse
	activeStreamsMu sync.RWMutex
}

// NewStreamingEchobeatsEngine creates a streaming echobeats engine
func NewStreamingEchobeatsEngine() *StreamingEchobeatsEngine {
	return &StreamingEchobeatsEngine{
		EchobeatsEngine: NewEchobeatsEngine(),
		activeStreams:   make(map[string]*StreamingResponse),
	}
}

// InferStreamOnStream performs streaming inference on a specific stream
func (se *StreamingEchobeatsEngine) InferStreamOnStream(
	ctx context.Context,
	streamID StreamID,
	req *StreamingRequest,
) (*StreamingResponse, error) {
	if !se.initialized.Load() {
		return nil, errors.New("engine not initialized")
	}
	
	// Create streaming response
	resp := NewStreamingResponse(streamID, req.Step)
	requestID := fmt.Sprintf("%d-%d-%d", streamID, req.Step, time.Now().UnixNano())
	resp.RequestID = requestID
	
	// Track active stream
	se.activeStreamsMu.Lock()
	se.activeStreams[requestID] = resp
	se.activeStreamsMu.Unlock()
	
	// Start generation goroutine
	go se.generateStream(ctx, streamID, req, resp)
	
	return resp, nil
}

func (se *StreamingEchobeatsEngine) generateStream(
	ctx context.Context,
	streamID StreamID,
	req *StreamingRequest,
	resp *StreamingResponse,
) {
	defer func() {
		resp.Stream.Close()
		
		// Remove from active streams
		se.activeStreamsMu.Lock()
		delete(se.activeStreams, resp.RequestID)
		se.activeStreamsMu.Unlock()
		
		if req.OnDone != nil {
			req.OnDone()
		}
	}()
	
	// Simulate token generation (in production, this would use actual llama.cpp)
	position := 0
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 256
	}
	
	// Mock tokenized prompt
	promptTokens := []string{"[", "Stream", " ", streamID.String(), "]", " "}
	words := tokenizeSimple(req.Prompt)
	
	// Generate tokens
	for i := 0; i < maxTokens; i++ {
		select {
		case <-ctx.Done():
			resp.Stream.CloseWithError(ctx.Err())
			if req.OnError != nil {
				req.OnError(ctx.Err())
			}
			return
		default:
		}
		
		// Create token
		var text string
		if i < len(promptTokens) {
			text = promptTokens[i]
		} else if i-len(promptTokens) < len(words) {
			text = words[i-len(promptTokens)] + " "
		} else {
			text = "..."
			// Mark as final
			token := &Token{
				ID:        int32(i),
				Text:      text,
				Logprob:   -0.1,
				Position:  position,
				IsFinal:   true,
			}
			
			resp.AccumulateToken(token)
			resp.Stream.Send(token)
			
			if req.OnToken != nil {
				req.OnToken(token)
			}
			break
		}
		
		token := &Token{
			ID:       int32(i),
			Text:     text,
			Logprob:  -0.1 * float32(i+1),
			Position: position,
			IsFinal:  false,
		}
		
		// Add top logprobs if requested
		if req.StreamLogprobs && req.TopLogprobs > 0 {
			token.TopLogprobs = make([]TokenLogprob, req.TopLogprobs)
			for j := 0; j < req.TopLogprobs; j++ {
				token.TopLogprobs[j] = TokenLogprob{
					ID:      int32(i + j + 1),
					Text:    fmt.Sprintf("alt%d", j),
					Logprob: -0.2 * float32(j+1),
				}
			}
		}
		
		resp.AccumulateToken(token)
		
		if err := resp.Stream.Send(token); err != nil {
			if req.OnError != nil {
				req.OnError(err)
			}
			return
		}
		
		if req.OnToken != nil {
			req.OnToken(token)
		}
		
		position++
		
		// Simulate generation latency
		time.Sleep(10 * time.Millisecond)
		
		// Check for stop sequences
		if containsStopSequence(resp.GetText(), req.StopSequences) {
			break
		}
	}
}

// InferStreamConcurrent performs streaming inference on all 3 streams
func (se *StreamingEchobeatsEngine) InferStreamConcurrent(
	ctx context.Context,
	requests [3]*StreamingRequest,
) ([3]*StreamingResponse, error) {
	responses := [3]*StreamingResponse{}
	
	for i := 0; i < 3; i++ {
		if requests[i] != nil {
			resp, err := se.InferStreamOnStream(ctx, StreamID(i), requests[i])
			if err != nil {
				return responses, err
			}
			responses[i] = resp
		}
	}
	
	return responses, nil
}

// GetActiveStreams returns all active streaming responses
func (se *StreamingEchobeatsEngine) GetActiveStreams() []*StreamingResponse {
	se.activeStreamsMu.RLock()
	defer se.activeStreamsMu.RUnlock()
	
	result := make([]*StreamingResponse, 0, len(se.activeStreams))
	for _, resp := range se.activeStreams {
		result = append(result, resp)
	}
	return result
}

// =============================================================================
// STREAM CONSUMER
// =============================================================================

// StreamConsumer consumes tokens from a stream
type StreamConsumer struct {
	stream    *TokenStream
	buffer    []*Token
	bufferMu  sync.RWMutex
	
	// Callbacks
	onToken   func(*Token)
	onDone    func([]*Token)
	onError   func(error)
}

// NewStreamConsumer creates a new stream consumer
func NewStreamConsumer(stream *TokenStream) *StreamConsumer {
	return &StreamConsumer{
		stream: stream,
		buffer: make([]*Token, 0, 256),
	}
}

// OnToken sets the token callback
func (sc *StreamConsumer) OnToken(fn func(*Token)) *StreamConsumer {
	sc.onToken = fn
	return sc
}

// OnDone sets the completion callback
func (sc *StreamConsumer) OnDone(fn func([]*Token)) *StreamConsumer {
	sc.onDone = fn
	return sc
}

// OnError sets the error callback
func (sc *StreamConsumer) OnError(fn func(error)) *StreamConsumer {
	sc.onError = fn
	return sc
}

// Consume starts consuming the stream
func (sc *StreamConsumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			token, err := sc.stream.Recv()
			if err != nil {
				if err == io.EOF {
					if sc.onDone != nil {
						sc.bufferMu.RLock()
						tokens := make([]*Token, len(sc.buffer))
						copy(tokens, sc.buffer)
						sc.bufferMu.RUnlock()
						sc.onDone(tokens)
					}
					return nil
				}
				if sc.onError != nil {
					sc.onError(err)
				}
				return err
			}
			
			sc.bufferMu.Lock()
			sc.buffer = append(sc.buffer, token)
			sc.bufferMu.Unlock()
			
			if sc.onToken != nil {
				sc.onToken(token)
			}
		}
	}
}

// ConsumeAsync starts consuming the stream asynchronously
func (sc *StreamConsumer) ConsumeAsync(ctx context.Context) {
	go sc.Consume(ctx)
}

// GetBuffer returns the buffered tokens
func (sc *StreamConsumer) GetBuffer() []*Token {
	sc.bufferMu.RLock()
	defer sc.bufferMu.RUnlock()
	result := make([]*Token, len(sc.buffer))
	copy(result, sc.buffer)
	return result
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func tokenizeSimple(text string) []string {
	var words []string
	current := ""
	for _, c := range text {
		if c == ' ' || c == '\n' || c == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

func containsStopSequence(text string, stopSequences []string) bool {
	for _, seq := range stopSequences {
		if len(seq) > 0 && len(text) >= len(seq) {
			if text[len(text)-len(seq):] == seq {
				return true
			}
		}
	}
	return false
}
