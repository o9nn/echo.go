package core

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// InterestPattern tracks interest in specific topics
type InterestPattern struct {
	Topic       string
	Strength    float64   // 0.0 - 1.0
	LastEngaged time.Time
	DecayRate   float64   // Per-hour decay rate
	Engagements uint64    // Number of times engaged
}

// DiscussionManager manages autonomous discussion and interest patterns
type DiscussionManager struct {
	mu               sync.RWMutex
	interests        map[string]*InterestPattern
	defaultDecayRate float64
	engagementThreshold float64
	maxInterests     int
}

// NewDiscussionManager creates a new discussion manager
func NewDiscussionManager(defaultDecayRate, engagementThreshold float64, maxInterests int) *DiscussionManager {
	return &DiscussionManager{
		interests:           make(map[string]*InterestPattern),
		defaultDecayRate:    defaultDecayRate,
		engagementThreshold: engagementThreshold,
		maxInterests:        maxInterests,
	}
}

// UpdateInterest updates or creates an interest pattern
func (dm *DiscussionManager) UpdateInterest(topic string, strengthDelta float64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	topic = strings.ToLower(strings.TrimSpace(topic))

	interest, exists := dm.interests[topic]
	if !exists {
		interest = &InterestPattern{
			Topic:       topic,
			Strength:    0.0,
			LastEngaged: time.Now(),
			DecayRate:   dm.defaultDecayRate,
			Engagements: 0,
		}
		dm.interests[topic] = interest
	}

	// Update strength
	interest.Strength += strengthDelta
	if interest.Strength > 1.0 {
		interest.Strength = 1.0
	}
	if interest.Strength < 0.0 {
		interest.Strength = 0.0
	}

	interest.LastEngaged = time.Now()
	interest.Engagements++

	// Trim if too many interests
	if len(dm.interests) > dm.maxInterests {
		dm.trimWeakestInterests()
	}
}

// GetInterest returns the current interest strength for a topic
func (dm *DiscussionManager) GetInterest(topic string) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	topic = strings.ToLower(strings.TrimSpace(topic))
	interest, exists := dm.interests[topic]
	if !exists {
		return 0.0
	}

	// Apply decay based on time since last engagement
	return dm.calculateDecayedStrength(interest)
}

// ShouldEngage determines if the agent should engage with a topic
func (dm *DiscussionManager) ShouldEngage(topic string, fatigue float64, cognitiveLoad float64) bool {
	// Get interest level
	interest := dm.GetInterest(topic)

	// Don't engage if too fatigued or overloaded
	if fatigue > 0.8 || cognitiveLoad > 0.9 {
		return false
	}

	// Adjust threshold based on current state
	adjustedThreshold := dm.engagementThreshold
	if fatigue > 0.5 {
		adjustedThreshold += 0.2 // Require higher interest when tired
	}

	// Engage if interest exceeds threshold
	return interest >= adjustedThreshold
}

// DecayInterests applies time-based decay to all interests
func (dm *DiscussionManager) DecayInterests() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for topic, interest := range dm.interests {
		// Calculate decay
		hoursSince := now.Sub(interest.LastEngaged).Hours()
		decay := interest.DecayRate * hoursSince

		interest.Strength -= decay
		if interest.Strength <= 0.0 {
			toDelete = append(toDelete, topic)
		}
	}

	// Remove interests that have decayed to zero
	for _, topic := range toDelete {
		delete(dm.interests, topic)
	}
}

// GetTopInterests returns the top N interests by strength
func (dm *DiscussionManager) GetTopInterests(n int) []InterestPattern {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Convert to slice
	interests := make([]InterestPattern, 0, len(dm.interests))
	for _, interest := range dm.interests {
		// Calculate current strength with decay
		currentStrength := dm.calculateDecayedStrength(interest)
		if currentStrength > 0.0 {
			interestCopy := *interest
			interestCopy.Strength = currentStrength
			interests = append(interests, interestCopy)
		}
	}

	// Sort by strength (simple bubble sort for small lists)
	for i := 0; i < len(interests)-1; i++ {
		for j := i + 1; j < len(interests); j++ {
			if interests[i].Strength < interests[j].Strength {
				interests[i], interests[j] = interests[j], interests[i]
			}
		}
	}

	// Return top N
	if n > len(interests) {
		n = len(interests)
	}
	return interests[:n]
}

// GetAllInterests returns all current interests
func (dm *DiscussionManager) GetAllInterests() map[string]float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	result := make(map[string]float64)
	for topic, interest := range dm.interests {
		result[topic] = dm.calculateDecayedStrength(interest)
	}
	return result
}

// calculateDecayedStrength calculates current strength with time decay
func (dm *DiscussionManager) calculateDecayedStrength(interest *InterestPattern) float64 {
	hoursSince := time.Since(interest.LastEngaged).Hours()
	decay := interest.DecayRate * hoursSince
	strength := interest.Strength - decay

	if strength < 0.0 {
		return 0.0
	}
	return strength
}

// trimWeakestInterests removes the weakest interests to stay under maxInterests
func (dm *DiscussionManager) trimWeakestInterests() {
	if len(dm.interests) <= dm.maxInterests {
		return
	}

	// Find weakest interests
	type interestWithStrength struct {
		topic    string
		strength float64
	}

	interests := make([]interestWithStrength, 0, len(dm.interests))
	for topic, interest := range dm.interests {
		interests = append(interests, interestWithStrength{
			topic:    topic,
			strength: dm.calculateDecayedStrength(interest),
		})
	}

	// Sort by strength
	for i := 0; i < len(interests)-1; i++ {
		for j := i + 1; j < len(interests); j++ {
			if interests[i].strength > interests[j].strength {
				interests[i], interests[j] = interests[j], interests[i]
			}
		}
	}

	// Delete weakest
	toDelete := len(dm.interests) - dm.maxInterests
	for i := 0; i < toDelete; i++ {
		delete(dm.interests, interests[i].topic)
	}
}

// GenerateDiscussionPrompt generates a prompt for autonomous discussion
func (dm *DiscussionManager) GenerateDiscussionPrompt() string {
	topInterests := dm.GetTopInterests(3)

	if len(topInterests) == 0 {
		return "What topics or questions are worth exploring right now?"
	}

	if len(topInterests) == 1 {
		return fmt.Sprintf("I'm interested in %s. What aspects of this are worth exploring further?",
			topInterests[0].Topic)
	}

	topics := make([]string, len(topInterests))
	for i, interest := range topInterests {
		topics[i] = interest.Topic
	}

	return fmt.Sprintf("I'm curious about the connections between %s. How might these relate?",
		strings.Join(topics, ", "))
}

// AnalyzeTopicRelevance analyzes how relevant a topic is to current interests
func (dm *DiscussionManager) AnalyzeTopicRelevance(topic string) float64 {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	topic = strings.ToLower(strings.TrimSpace(topic))
	topicWords := strings.Fields(topic)

	// Check direct match
	if interest, exists := dm.interests[topic]; exists {
		return dm.calculateDecayedStrength(interest)
	}

	// Check partial matches
	maxRelevance := 0.0
	for interestTopic, interest := range dm.interests {
		interestWords := strings.Fields(interestTopic)

		// Count common words
		commonWords := 0
		for _, tw := range topicWords {
			for _, iw := range interestWords {
				if tw == iw && len(tw) > 3 { // Only meaningful words
					commonWords++
				}
			}
		}

		if commonWords > 0 {
			// Calculate relevance based on word overlap
			relevance := dm.calculateDecayedStrength(interest) * float64(commonWords) / float64(len(interestWords))
			if relevance > maxRelevance {
				maxRelevance = relevance
			}
		}
	}

	return maxRelevance
}

// ExtractTopicsFromText extracts potential topics from text
func ExtractTopicsFromText(text string) []string {
	// Simple topic extraction - find meaningful phrases
	words := strings.Fields(strings.ToLower(text))
	topics := make([]string, 0)

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"be": true, "been": true, "are": true, "were": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
		"this": true, "that": true, "these": true, "those": true,
	}

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}") 
		if len(word) > 4 && !stopWords[word] {
			topics = append(topics, word)
		}
	}

	return topics
}
