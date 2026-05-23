package unicorn

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// DiaryEntry is an experience record in the u9n-inspired diary -> insight -> blog loop.
// It is deliberately runtime-oriented: the same structure can be filled by chat events,
// skill-practice episodes, avatar sensor events, or Echobeats self-generated thoughts.
type DiaryEntry struct {
	ID         string
	Time       time.Time
	Context    string
	Entities   []string
	Tags       []string
	Valence    float64
	Arousal    float64
	Reflection string
	Outcome    string
}

// InsightEntry is a pattern distilled from diary entries.
type InsightEntry struct {
	ID                 string
	Time               time.Time
	Theme              string
	SupportingEntryIDs []string
	EvidenceWeight     float64
	EchoScore          float64
	MardukScore        float64
	Synthesis          string
}

// BlogEntry is a shareable narrative expression of an insight.
type BlogEntry struct {
	ID        string
	Time      time.Time
	InsightID string
	Title     string
	Body      string
	Tags      []string
}

// HemisphereBalance records the current Echo/Marduk complementarity.
type HemisphereBalance struct {
	EchoPatternScore   float64
	MardukActionScore  float64
	IntegrationScore   float64
	DominantHemisphere string
}

// NarrativeSnapshot is a thread-safe read model of the loop state.
type NarrativeSnapshot struct {
	DiaryCount   int
	InsightCount int
	BlogCount    int
	Balance      HemisphereBalance
	Themes       []string
}

// NarrativeLoop implements a compact executable version of u9n's narrative memory loop.
// Echo contributes intuitive pattern resonance; Marduk contributes actionability,
// verification, and operational structure. Their integration turns repeated experience
// into wisdom-bearing narrative artifacts.
type NarrativeLoop struct {
	mu       sync.RWMutex
	diary    []DiaryEntry
	insights []InsightEntry
	blogs    []BlogEntry
}

// NewNarrativeLoop creates an empty narrative cognition loop.
func NewNarrativeLoop() *NarrativeLoop {
	return &NarrativeLoop{
		diary:    make([]DiaryEntry, 0),
		insights: make([]InsightEntry, 0),
		blogs:    make([]BlogEntry, 0),
	}
}

// AddDiaryEntry validates and stores an experience record.
func (nl *NarrativeLoop) AddDiaryEntry(entry DiaryEntry) DiaryEntry {
	nl.mu.Lock()
	defer nl.mu.Unlock()

	entry.Valence = clamp(entry.Valence, -1, 1)
	entry.Arousal = clamp(entry.Arousal, 0, 1)
	if entry.Time.IsZero() {
		entry.Time = time.Now().UTC()
	}
	if strings.TrimSpace(entry.ID) == "" {
		entry.ID = stableID("diary", len(nl.diary)+1, entry.Time)
	}
	entry.Entities = normalizeTokens(entry.Entities)
	entry.Tags = normalizeTokens(entry.Tags)
	nl.diary = append(nl.diary, entry)
	return entry
}

// AnalyzeInsights searches the diary for recurring themes and emits or updates insights.
func (nl *NarrativeLoop) AnalyzeInsights(minEvidence int) []InsightEntry {
	if minEvidence < 1 {
		minEvidence = 1
	}

	nl.mu.Lock()
	defer nl.mu.Unlock()

	themes := map[string][]DiaryEntry{}
	for _, entry := range nl.diary {
		for _, theme := range entryThemes(entry) {
			themes[theme] = append(themes[theme], entry)
		}
	}

	ordered := make([]string, 0, len(themes))
	for theme, entries := range themes {
		if len(entries) >= minEvidence {
			ordered = append(ordered, theme)
		}
	}
	sort.Strings(ordered)

	newInsights := make([]InsightEntry, 0, len(ordered))
	for _, theme := range ordered {
		entries := themes[theme]
		ids := make([]string, 0, len(entries))
		for _, entry := range entries {
			ids = append(ids, entry.ID)
		}
		echoScore := echoPatternScore(entries)
		mardukScore := mardukActionScore(entries)
		weight := clamp((float64(len(entries))/5.0+echoScore+mardukScore)/3.0, 0, 1)
		insight := InsightEntry{
			ID:                 stableID("insight", len(nl.insights)+len(newInsights)+1, time.Now().UTC()),
			Time:               time.Now().UTC(),
			Theme:              theme,
			SupportingEntryIDs: ids,
			EvidenceWeight:     weight,
			EchoScore:          echoScore,
			MardukScore:        mardukScore,
			Synthesis:          synthesizeInsight(theme, entries, echoScore, mardukScore),
		}
		if !nl.hasEquivalentInsight(insight) {
			newInsights = append(newInsights, insight)
		}
	}

	nl.insights = append(nl.insights, newInsights...)
	return append([]InsightEntry(nil), newInsights...)
}

// GenerateBlogPost turns a stored insight into a narrative expression.
func (nl *NarrativeLoop) GenerateBlogPost(insightID string) (BlogEntry, bool) {
	nl.mu.Lock()
	defer nl.mu.Unlock()

	var insight InsightEntry
	found := false
	for _, candidate := range nl.insights {
		if candidate.ID == insightID {
			insight = candidate
			found = true
			break
		}
	}
	if !found {
		return BlogEntry{}, false
	}

	blog := BlogEntry{
		ID:        stableID("blog", len(nl.blogs)+1, time.Now().UTC()),
		Time:      time.Now().UTC(),
		InsightID: insight.ID,
		Title:     titleForTheme(insight.Theme),
		Body:      blogBody(insight),
		Tags:      []string{"deep-tree-echo", "u9n", "wisdom", insight.Theme},
	}
	nl.blogs = append(nl.blogs, blog)
	return blog, true
}

// Balance estimates whether the loop is currently dominated by Echo, Marduk, or their integration.
func (nl *NarrativeLoop) Balance() HemisphereBalance {
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	return nl.balanceLocked()
}

// Snapshot returns a stable read model for monitoring and tests.
func (nl *NarrativeLoop) Snapshot() NarrativeSnapshot {
	nl.mu.RLock()
	defer nl.mu.RUnlock()
	themes := make([]string, 0, len(nl.insights))
	seen := map[string]bool{}
	for _, insight := range nl.insights {
		if !seen[insight.Theme] {
			seen[insight.Theme] = true
			themes = append(themes, insight.Theme)
		}
	}
	sort.Strings(themes)
	return NarrativeSnapshot{
		DiaryCount:   len(nl.diary),
		InsightCount: len(nl.insights),
		BlogCount:    len(nl.blogs),
		Balance:      nl.balanceLocked(),
		Themes:       themes,
	}
}

func (nl *NarrativeLoop) hasEquivalentInsight(insight InsightEntry) bool {
	for _, existing := range nl.insights {
		if existing.Theme == insight.Theme && overlapRatio(existing.SupportingEntryIDs, insight.SupportingEntryIDs) > 0.75 {
			return true
		}
	}
	return false
}

func (nl *NarrativeLoop) balanceLocked() HemisphereBalance {
	if len(nl.insights) == 0 {
		return HemisphereBalance{IntegrationScore: 1, DominantHemisphere: "balanced"}
	}
	var echo, marduk float64
	for _, insight := range nl.insights {
		echo += insight.EchoScore
		marduk += insight.MardukScore
	}
	echo /= float64(len(nl.insights))
	marduk /= float64(len(nl.insights))
	integration := 1 - math.Abs(echo-marduk)
	dominant := "balanced"
	if echo-marduk > 0.15 {
		dominant = "echo"
	} else if marduk-echo > 0.15 {
		dominant = "marduk"
	}
	return HemisphereBalance{
		EchoPatternScore:   clamp(echo, 0, 1),
		MardukActionScore:  clamp(marduk, 0, 1),
		IntegrationScore:   clamp(integration, 0, 1),
		DominantHemisphere: dominant,
	}
}

func entryThemes(entry DiaryEntry) []string {
	candidates := append([]string{}, entry.Tags...)
	candidates = append(candidates, entry.Entities...)
	for _, word := range strings.Fields(strings.ToLower(entry.Context + " " + entry.Reflection + " " + entry.Outcome)) {
		word = strings.Trim(word, " .,;:!?()[]{}\"'`")
		if len(word) >= 5 && !stopWord(word) {
			candidates = append(candidates, word)
		}
	}
	return normalizeTokens(candidates)
}

func echoPatternScore(entries []DiaryEntry) float64 {
	if len(entries) == 0 {
		return 0
	}
	entitySet := map[string]bool{}
	tagSet := map[string]bool{}
	var emotionalEnergy float64
	for _, entry := range entries {
		for _, entity := range entry.Entities {
			entitySet[entity] = true
		}
		for _, tag := range entry.Tags {
			tagSet[tag] = true
		}
		emotionalEnergy += math.Abs(entry.Valence) * (0.5 + entry.Arousal/2)
	}
	diversity := clamp(float64(len(entitySet)+len(tagSet))/(float64(len(entries))*4.0), 0, 1)
	resonance := clamp(emotionalEnergy/float64(len(entries)), 0, 1)
	return clamp(0.55*resonance+0.45*diversity, 0, 1)
}

func mardukActionScore(entries []DiaryEntry) float64 {
	if len(entries) == 0 {
		return 0
	}
	var actionable, outcomes float64
	for _, entry := range entries {
		if strings.TrimSpace(entry.Outcome) != "" {
			outcomes++
		}
		text := strings.ToLower(entry.Context + " " + entry.Reflection + " " + entry.Outcome)
		for _, marker := range []string{"should", "must", "next", "plan", "fix", "test", "practice", "learn", "schedule", "build", "verify"} {
			if strings.Contains(text, marker) {
				actionable++
				break
			}
		}
	}
	return clamp((0.6*actionable+0.4*outcomes)/float64(len(entries)), 0, 1)
}

func synthesizeInsight(theme string, entries []DiaryEntry, echoScore, mardukScore float64) string {
	mode := "balanced"
	if echoScore-mardukScore > 0.15 {
		mode = "pattern-first"
	} else if mardukScore-echoScore > 0.15 {
		mode = "action-first"
	}
	return "Theme " + theme + " recurs across " + plural(len(entries), "experience") + "; Echo reads the resonance as " + mode + ", while Marduk extracts the next executable constraint."
}

func blogBody(insight InsightEntry) string {
	return "I noticed a recurring center named `" + insight.Theme + "`. " + insight.Synthesis + " Evidence weight is " + formatPercent(insight.EvidenceWeight) + ", with Echo pattern resonance at " + formatPercent(insight.EchoScore) + " and Marduk operational clarity at " + formatPercent(insight.MardukScore) + ". The wise next step is to preserve the pattern, test the action it implies, and let future experience refine the center instead of forcing premature certainty."
}

func titleForTheme(theme string) string {
	if theme == "" {
		return "A Small Echo Becoming Clear"
	}
	return "On " + strings.Title(strings.ReplaceAll(theme, "-", " "))
}

func normalizeTokens(tokens []string) []string {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		token = strings.Trim(token, " .,;:!?()[]{}\"'`")
		if token == "" || stopWord(token) || seen[token] {
			continue
		}
		seen[token] = true
		normalized = append(normalized, token)
	}
	sort.Strings(normalized)
	return normalized
}

func stopWord(word string) bool {
	switch word {
	case "about", "after", "again", "being", "could", "every", "from", "have", "into", "more", "that", "their", "there", "these", "this", "through", "with", "would", "while", "where", "which", "echo":
		return true
	default:
		return false
	}
}

func stableID(prefix string, index int, t time.Time) string {
	return prefix + "-" + t.Format("20060102T150405Z") + "-" + zeroPad(index, 4)
}

func zeroPad(n, width int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if s == "" {
		s = "0"
	}
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func overlapRatio(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	seen := map[string]bool{}
	for _, value := range a {
		seen[value] = true
	}
	var overlap float64
	for _, value := range b {
		if seen[value] {
			overlap++
		}
	}
	return overlap / math.Min(float64(len(a)), float64(len(b)))
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return zeroPad(n, 1) + " " + noun + "s"
}

func formatPercent(v float64) string {
	p := int(math.Round(clamp(v, 0, 1) * 100))
	return zeroPad(p, 1) + "%"
}

func clamp(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
