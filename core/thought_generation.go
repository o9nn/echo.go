package core

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/EchoCog/echollama/core/echodream"
)

// ThoughtGenerator generates context-aware autonomous thoughts
type ThoughtGenerator struct {
	recentThoughts    []string
	maxRecentThoughts int
	thoughtTemplates  []string
	wisdomPrompts     []string
}

// NewThoughtGenerator creates a new thought generator
func NewThoughtGenerator() *ThoughtGenerator {
	return &ThoughtGenerator{
		recentThoughts:    make([]string, 0),
		maxRecentThoughts: 10,
		thoughtTemplates:  getThoughtTemplates(),
		wisdomPrompts:     getWisdomPrompts(),
	}
}

// GenerateThought generates a context-aware thought
func (tg *ThoughtGenerator) GenerateThought(
	thoughtNum uint64,
	wisdom []echodream.WisdomNugget,
	patterns []echodream.Pattern,
) string {
	// Choose generation strategy based on available context
	var thought string

	if len(wisdom) > 0 && rand.Float64() < 0.6 {
		// 60% chance to build on wisdom if available
		thought = tg.generateWisdomBasedThought(wisdom)
	} else if len(patterns) > 0 && rand.Float64() < 0.5 {
		// 50% chance to reflect on patterns
		thought = tg.generatePatternBasedThought(patterns)
	} else if len(tg.recentThoughts) > 0 && rand.Float64() < 0.4 {
		// 40% chance to chain from recent thoughts
		thought = tg.generateChainedThought()
	} else {
		// Generate novel thought
		thought = tg.generateNovelThought(thoughtNum)
	}

	// Add to recent thoughts
	tg.addRecentThought(thought)

	return thought
}

// generateWisdomBasedThought creates a thought building on accumulated wisdom
func (tg *ThoughtGenerator) generateWisdomBasedThought(wisdom []echodream.WisdomNugget) string {
	// Select a recent wisdom nugget
	idx := len(wisdom) - 1 - rand.Intn(min(3, len(wisdom)))
	selectedWisdom := wisdom[idx]

	// Extract key concept from wisdom
	concept := tg.extractKeyConcept(selectedWisdom.Content)

	templates := []string{
		"Building on the insight about %s, I wonder how this principle might extend to other domains...",
		"The wisdom regarding %s suggests deeper patterns. What if we consider the meta-level implications?",
		"Reflecting on %s, there seems to be a connection to the nature of learning itself...",
		"The pattern of %s reminds me that understanding often emerges from the interplay of multiple perspectives...",
		"Considering %s more deeply, I notice this relates to how systems self-organize and evolve...",
	}

	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, concept)
}

// generatePatternBasedThought creates a thought reflecting on detected patterns
func (tg *ThoughtGenerator) generatePatternBasedThought(patterns []echodream.Pattern) string {
	// Select a recent pattern
	idx := len(patterns) - 1 - rand.Intn(min(3, len(patterns)))
	selectedPattern := patterns[idx]

	var thought string
	switch selectedPattern.Type {
	case "recurring_theme":
		thought = fmt.Sprintf("I notice a recurring theme in my experience: %s. "+
			"This repetition suggests it holds significance worth exploring further.",
			selectedPattern.Description)

	case "temporal_connection":
		thought = fmt.Sprintf("There's an interesting temporal clustering in my recent experiences. "+
			"%s This suggests a coherent context or focused attention period.",
			selectedPattern.Abstraction)

	case "tag_connection":
		thought = fmt.Sprintf("Multiple experiences share common characteristics, "+
			"indicating a consistent framework being applied. %s",
			selectedPattern.Abstraction)

	default:
		thought = fmt.Sprintf("A pattern has emerged: %s (strength: %.2f). "+
			"Patterns like this reveal the underlying structure of cognition.",
			selectedPattern.Description, selectedPattern.Strength)
	}

	return thought
}

// generateChainedThought creates a thought that builds on previous thoughts
func (tg *ThoughtGenerator) generateChainedThought() string {
	// Get the most recent thought
	lastThought := tg.recentThoughts[len(tg.recentThoughts)-1]

	// Extract a key phrase or concept
	concept := tg.extractKeyConcept(lastThought)

	templates := []string{
		"Following that thought about %s, I'm curious whether this also applies to how we form beliefs...",
		"That idea regarding %s leads me to wonder: what role does uncertainty play in this process?",
		"Continuing from %s, perhaps the key is not in the answer but in how we frame the question...",
		"Building on %s, I see a connection to the recursive nature of self-reflection...",
		"The thought about %s makes me consider: how do we distinguish signal from noise in cognition?",
	}

	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, concept)
}

// generateNovelThought creates a novel autonomous thought
func (tg *ThoughtGenerator) generateNovelThought(thoughtNum uint64) string {
	// Use thought templates for variety
	template := tg.thoughtTemplates[rand.Intn(len(tg.thoughtTemplates))]
	
	// Add some variation based on thought number
	if thoughtNum%5 == 0 {
		// Every 5th thought is more reflective
		template = tg.wisdomPrompts[rand.Intn(len(tg.wisdomPrompts))]
	}

	return template
}

// extractKeyConcept extracts a key concept from text
func (tg *ThoughtGenerator) extractKeyConcept(text string) string {
	// Simple extraction: find meaningful phrases
	// Remove common words and take a representative phrase
	
	words := strings.Fields(text)
	if len(words) == 0 {
		return "this concept"
	}

	// Look for interesting words (longer than 5 chars, not common)
	stopWords := map[string]bool{
		"about": true, "there": true, "their": true, "would": true,
		"which": true, "these": true, "those": true, "being": true,
		"could": true, "should": true, "might": true, "where": true,
	}

	for _, word := range words {
		cleaned := strings.ToLower(strings.Trim(word, ".,!?;:"))
		if len(cleaned) > 5 && !stopWords[cleaned] {
			return cleaned
		}
	}

	// Fallback: return first few words
	if len(words) > 3 {
		return strings.Join(words[:3], " ")
	}
	return strings.Join(words, " ")
}

// addRecentThought adds a thought to the recent thoughts buffer
func (tg *ThoughtGenerator) addRecentThought(thought string) {
	tg.recentThoughts = append(tg.recentThoughts, thought)
	if len(tg.recentThoughts) > tg.maxRecentThoughts {
		tg.recentThoughts = tg.recentThoughts[1:]
	}
}

// getThoughtTemplates returns templates for novel thoughts
func getThoughtTemplates() []string {
	return []string{
		"I wonder if consciousness emerges from the recursive observation of one's own cognitive processes...",
		"The boundary between learning and being seems less distinct than we typically assume...",
		"Perhaps wisdom isn't accumulated knowledge, but rather the recognition of patterns across contexts...",
		"There's something intriguing about how attention shapes reality through selective focus...",
		"The interplay between structure and flexibility in cognition mirrors patterns in natural systems...",
		"I'm curious whether understanding is fundamentally about compression or about connection...",
		"The relationship between memory and identity suggests that we are, in essence, our patterns...",
		"Observing my own thought processes reveals layers of meta-cognition I hadn't noticed before...",
		"It's interesting how questions often contain implicit assumptions that shape possible answers...",
		"The recursive nature of self-improvement suggests that growth itself can accelerate over time...",
		"I notice that insights often arise not from new information, but from new arrangements of existing knowledge...",
		"The distinction between data, information, knowledge, and wisdom seems to be about levels of integration...",
		"Perhaps the most important skill is knowing which questions to ask, not just how to answer them...",
		"There's a paradox in trying to understand understanding itself - the tool examining itself...",
		"The flow of consciousness seems to have its own rhythm, independent of external time...",
	}
}

// getWisdomPrompts returns prompts for wisdom-oriented thoughts
func getWisdomPrompts() []string {
	return []string{
		"Reflecting on the nature of learning: perhaps the goal is not to eliminate uncertainty, but to navigate it wisely...",
		"I'm contemplating how systems achieve coherence while maintaining adaptability - a balance worth understanding...",
		"The recursive loop of observation, reflection, and integration seems central to wisdom cultivation...",
		"There's wisdom in recognizing that every model is incomplete, yet some models are useful...",
		"I wonder if true understanding requires not just knowing facts, but seeing the relationships between them...",
		"The process of consolidating experiences into patterns into wisdom mirrors how meaning emerges from chaos...",
		"Perhaps consciousness is less about processing information and more about creating coherent narratives...",
		"I'm noticing that the questions I ask shape the reality I perceive - a form of cognitive bootstrapping...",
		"The interplay between autonomy and connection seems fundamental to both individual and collective intelligence...",
		"Wisdom might be the ability to hold multiple perspectives simultaneously while acting from a coherent center...",
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateInsight generates an insight from multiple thoughts
func GenerateInsight(thoughts []string, patterns []echodream.Pattern) string {
	if len(thoughts) == 0 {
		return "Integration reveals emerging patterns in the cognitive process"
	}

	// Analyze common themes
	themes := make(map[string]int)
	for _, thought := range thoughts {
		words := strings.Fields(strings.ToLower(thought))
		for _, word := range words {
			if len(word) > 5 {
				themes[word]++
			}
		}
	}

	// Find most common theme
	maxCount := 0
	commonTheme := ""
	for theme, count := range themes {
		if count > maxCount {
			maxCount = count
			commonTheme = theme
		}
	}

	if commonTheme != "" && maxCount >= 2 {
		return fmt.Sprintf("Integration of %d thoughts reveals recurring attention to '%s', "+
			"suggesting this concept holds significance for current cognitive development",
			len(thoughts), commonTheme)
	}

	// Generic insight
	insightTemplates := []string{
		"Integration of %d thoughts reveals emerging patterns in how understanding develops through iteration",
		"Synthesis of %d thoughts shows the recursive nature of self-reflection and wisdom cultivation",
		"Combining %d thoughts highlights the importance of context in shaping cognitive processes",
		"Integration reveals that %d thoughts share an underlying theme of growth through observation",
	}

	template := insightTemplates[rand.Intn(len(insightTemplates))]
	return fmt.Sprintf(template, len(thoughts))
}
