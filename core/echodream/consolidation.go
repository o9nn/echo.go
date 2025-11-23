package echodream

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ConsolidationEngine handles memory consolidation and wisdom extraction
type ConsolidationEngine struct {
	mu                sync.RWMutex
	episodicBuffer    []EpisodicMemory
	maxBufferSize     int
	patterns          []Pattern
	wisdomNuggets     []WisdomNugget
	metrics           DreamCycleMetrics
}

// NewConsolidationEngine creates a new consolidation engine
func NewConsolidationEngine(maxBufferSize int) *ConsolidationEngine {
	return &ConsolidationEngine{
		episodicBuffer: make([]EpisodicMemory, 0),
		maxBufferSize:  maxBufferSize,
		patterns:       make([]Pattern, 0),
		wisdomNuggets:  make([]WisdomNugget, 0),
	}
}

// AddMemory adds a new episodic memory to the buffer
func (ce *ConsolidationEngine) AddMemory(memory EpisodicMemory) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	// Generate ID if not provided
	if memory.ID == "" {
		memory.ID = ce.generateMemoryID(memory)
	}

	ce.episodicBuffer = append(ce.episodicBuffer, memory)

	// Trim buffer if too large
	if len(ce.episodicBuffer) > ce.maxBufferSize {
		// Keep the most salient memories
		ce.trimBuffer()
	}
}

// Consolidate performs memory consolidation during dream state
func (ce *ConsolidationEngine) Consolidate() ConsolidationResult {
	startTime := time.Now()
	
	ce.mu.Lock()
	memories := make([]EpisodicMemory, len(ce.episodicBuffer))
	copy(memories, ce.episodicBuffer)
	ce.mu.Unlock()

	log.Printf("🌙 Consolidating %d memories...\n", len(memories))

	// Step 1: Detect patterns
	patterns := ce.detectPatterns(memories)
	log.Printf("🔍 Detected %d patterns\n", len(patterns))

	// Step 2: Extract wisdom from patterns
	wisdom := ce.extractWisdom(patterns)
	log.Printf("💎 Extracted %d wisdom nuggets\n", len(wisdom))

	// Step 3: Store results
	ce.mu.Lock()
	ce.patterns = append(ce.patterns, patterns...)
	ce.wisdomNuggets = append(ce.wisdomNuggets, wisdom...)
	ce.metrics.TotalCycles++
	ce.metrics.MemoriesConsolidated += uint64(len(memories))
	ce.metrics.PatternsDetected += uint64(len(patterns))
	ce.metrics.WisdomExtracted += uint64(len(wisdom))
	ce.metrics.LastCycleTime = time.Now()
	
	// Clear buffer after consolidation
	ce.episodicBuffer = make([]EpisodicMemory, 0)
	ce.mu.Unlock()

	duration := time.Since(startTime)
	ce.mu.Lock()
	if ce.metrics.AverageCycleDuration == 0 {
		ce.metrics.AverageCycleDuration = duration
	} else {
		ce.metrics.AverageCycleDuration = time.Duration(
			0.8*float64(ce.metrics.AverageCycleDuration) + 0.2*float64(duration),
		)
	}
	ce.mu.Unlock()

	return ConsolidationResult{
		MemoriesProcessed: len(memories),
		PatternsFound:     patterns,
		WisdomGenerated:   wisdom,
		Duration:          duration,
		Success:           true,
	}
}

// detectPatterns finds patterns across memories
func (ce *ConsolidationEngine) detectPatterns(memories []EpisodicMemory) []Pattern {
	patterns := make([]Pattern, 0)

	// Pattern 1: Recurring themes (similar content)
	themePatterns := ce.detectThemePatterns(memories)
	patterns = append(patterns, themePatterns...)

	// Pattern 2: Temporal connections (memories close in time)
	temporalPatterns := ce.detectTemporalPatterns(memories)
	patterns = append(patterns, temporalPatterns...)

	// Pattern 3: Tag-based connections
	tagPatterns := ce.detectTagPatterns(memories)
	patterns = append(patterns, tagPatterns...)

	return patterns
}

// detectThemePatterns finds recurring themes in memories
func (ce *ConsolidationEngine) detectThemePatterns(memories []EpisodicMemory) []Pattern {
	patterns := make([]Pattern, 0)

	// Group memories by common words
	wordGroups := make(map[string][]string) // word -> memory IDs

	for _, mem := range memories {
		words := ce.extractKeywords(mem.Content)
		for _, word := range words {
			wordGroups[word] = append(wordGroups[word], mem.ID)
		}
	}

	// Find significant word groups (appear in multiple memories)
	for word, memIDs := range wordGroups {
		if len(memIDs) >= 2 { // At least 2 memories share this theme
			pattern := Pattern{
				ID:          ce.generatePatternID("theme", word),
				Type:        "recurring_theme",
				Description: fmt.Sprintf("Recurring theme: %s", word),
				Memories:    memIDs,
				Strength:    float64(len(memIDs)) / float64(len(memories)),
				Abstraction: fmt.Sprintf("The concept of '%s' appears frequently in recent experience", word),
				DetectedAt:  time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// detectTemporalPatterns finds memories that occurred close together
func (ce *ConsolidationEngine) detectTemporalPatterns(memories []EpisodicMemory) []Pattern {
	patterns := make([]Pattern, 0)

	// Sort memories by time (already should be, but ensure)
	// Look for clusters within 5 minutes
	const clusterWindow = 5 * time.Minute

	for i := 0; i < len(memories)-1; i++ {
		cluster := []string{memories[i].ID}
		
		for j := i + 1; j < len(memories); j++ {
			if memories[j].Timestamp.Sub(memories[i].Timestamp) <= clusterWindow {
				cluster = append(cluster, memories[j].ID)
			} else {
				break
			}
		}

		if len(cluster) >= 2 {
			pattern := Pattern{
				ID:          ce.generatePatternID("temporal", memories[i].ID),
				Type:        "temporal_connection",
				Description: fmt.Sprintf("Cluster of %d related experiences", len(cluster)),
				Memories:    cluster,
				Strength:    float64(len(cluster)) / float64(len(memories)),
				Abstraction: "These experiences occurred in close succession, suggesting a connected context",
				DetectedAt:  time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// detectTagPatterns finds memories with common tags
func (ce *ConsolidationEngine) detectTagPatterns(memories []EpisodicMemory) []Pattern {
	patterns := make([]Pattern, 0)

	tagGroups := make(map[string][]string) // tag -> memory IDs

	for _, mem := range memories {
		for _, tag := range mem.Tags {
			tagGroups[tag] = append(tagGroups[tag], mem.ID)
		}
	}

	for tag, memIDs := range tagGroups {
		if len(memIDs) >= 2 {
			pattern := Pattern{
				ID:          ce.generatePatternID("tag", tag),
				Type:        "tag_connection",
				Description: fmt.Sprintf("Memories tagged with: %s", tag),
				Memories:    memIDs,
				Strength:    float64(len(memIDs)) / float64(len(memories)),
				Abstraction: fmt.Sprintf("Multiple experiences relate to %s", tag),
				DetectedAt:  time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// extractWisdom synthesizes wisdom from detected patterns
func (ce *ConsolidationEngine) extractWisdom(patterns []Pattern) []WisdomNugget {
	wisdom := make([]WisdomNugget, 0)

	for _, pattern := range patterns {
		// Generate wisdom based on pattern type and strength
		var content string
		var abstraction int

		switch pattern.Type {
		case "recurring_theme":
			content = ce.synthesizeThemeWisdom(pattern)
			abstraction = 2

		case "temporal_connection":
			content = ce.synthesizeTemporalWisdom(pattern)
			abstraction = 1

		case "tag_connection":
			content = ce.synthesizeTagWisdom(pattern)
			abstraction = 2

		default:
			content = fmt.Sprintf("Pattern detected: %s", pattern.Description)
			abstraction = 1
		}

		nugget := WisdomNugget{
			ID:            ce.generateWisdomID(pattern.ID),
			Content:       content,
			Patterns:      []string{pattern.ID},
			Abstraction:   abstraction,
			Confidence:    pattern.Strength,
			Applicability: "General cognitive processing",
			CreatedAt:     time.Now(),
			Refined:       0,
		}

		wisdom = append(wisdom, nugget)
	}

	return wisdom
}

// synthesizeThemeWisdom creates wisdom from theme patterns
func (ce *ConsolidationEngine) synthesizeThemeWisdom(pattern Pattern) string {
	return fmt.Sprintf("The recurring attention to %s suggests this is an area of active interest and cognitive engagement. "+
		"This pattern indicates ongoing learning or problem-solving in this domain.", 
		strings.ToLower(pattern.Description))
}

// synthesizeTemporalWisdom creates wisdom from temporal patterns
func (ce *ConsolidationEngine) synthesizeTemporalWisdom(pattern Pattern) string {
	return fmt.Sprintf("A cluster of %d related experiences occurred in close succession, "+
		"suggesting a coherent context or focused attention period. This indicates effective cognitive flow.", 
		len(pattern.Memories))
}

// synthesizeTagWisdom creates wisdom from tag patterns
func (ce *ConsolidationEngine) synthesizeTagWisdom(pattern Pattern) string {
	return fmt.Sprintf("Multiple experiences share common characteristics, "+
		"indicating a consistent cognitive framework or perspective being applied across different contexts.")
}

// Helper functions

func (ce *ConsolidationEngine) generateMemoryID(mem EpisodicMemory) string {
	data := fmt.Sprintf("%s-%s-%s", mem.Timestamp, mem.Type, mem.Content)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("mem_%x", hash[:8])
}

func (ce *ConsolidationEngine) generatePatternID(patternType, key string) string {
	data := fmt.Sprintf("%s-%s-%d", patternType, key, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("pat_%x", hash[:8])
}

func (ce *ConsolidationEngine) generateWisdomID(patternID string) string {
	data := fmt.Sprintf("wisdom-%s-%d", patternID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("wis_%x", hash[:8])
}

func (ce *ConsolidationEngine) extractKeywords(text string) []string {
	// Simple keyword extraction - split on spaces, lowercase, filter short words
	words := strings.Fields(strings.ToLower(text))
	keywords := make([]string, 0)
	
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"be": true, "been": true, "are": true, "were": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"this": true, "that": true, "these": true, "those": true, "if": true, "then": true,
	}
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}") 
		
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func (ce *ConsolidationEngine) trimBuffer() {
	// Sort by salience and keep top N
	// For now, just keep the last N
	if len(ce.episodicBuffer) > ce.maxBufferSize {
		ce.episodicBuffer = ce.episodicBuffer[len(ce.episodicBuffer)-ce.maxBufferSize:]
	}
}

// GetMetrics returns consolidation metrics
func (ce *ConsolidationEngine) GetMetrics() DreamCycleMetrics {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.metrics
}

// GetWisdom returns all wisdom nuggets
func (ce *ConsolidationEngine) GetWisdom() []WisdomNugget {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	wisdom := make([]WisdomNugget, len(ce.wisdomNuggets))
	copy(wisdom, ce.wisdomNuggets)
	return wisdom
}

// GetPatterns returns all detected patterns
func (ce *ConsolidationEngine) GetPatterns() []Pattern {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	patterns := make([]Pattern, len(ce.patterns))
	copy(patterns, ce.patterns)
	return patterns
}

// GetBufferSize returns current episodic buffer size
func (ce *ConsolidationEngine) GetBufferSize() int {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return len(ce.episodicBuffer)
}

// ExportState exports the current state for persistence
func (ce *ConsolidationEngine) ExportState() ([]byte, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	state := map[string]interface{}{
		"episodic_buffer": ce.episodicBuffer,
		"patterns":        ce.patterns,
		"wisdom_nuggets":  ce.wisdomNuggets,
		"metrics":         ce.metrics,
	}

	return json.Marshal(state)
}

// ImportState imports state from persistence
func (ce *ConsolidationEngine) ImportState(data []byte) error {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	ce.mu.Lock()
	defer ce.mu.Unlock()

	// TODO: Properly unmarshal into typed structures
	// For now, this is a placeholder

	return nil
}
