package unicorn

import (
	"strings"
	"testing"
	"time"
)

func TestNarrativeLoopDiaryInsightBlog(t *testing.T) {
	loop := NewNarrativeLoop()
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)

	loop.AddDiaryEntry(DiaryEntry{
		Time:       now,
		Context:    "Echo should practice physical balance before avatar action.",
		Entities:   []string{"Avatar", "PhyMotion"},
		Tags:       []string{"practice", "embodiment"},
		Valence:    0.4,
		Arousal:    0.7,
		Reflection: "The next step is to test motion feasibility.",
		Outcome:    "Plan a grounded practice loop.",
	})
	loop.AddDiaryEntry(DiaryEntry{
		Time:       now.Add(time.Minute),
		Context:    "Echo should preserve a diary insight before publishing wisdom.",
		Entities:   []string{"U9N", "Avatar"},
		Tags:       []string{"practice", "narrative"},
		Valence:    0.6,
		Arousal:    0.6,
		Reflection: "Practice requires memory and action.",
		Outcome:    "Build a diary to insight to blog cycle.",
	})

	insights := loop.AnalyzeInsights(2)
	if len(insights) == 0 {
		t.Fatalf("expected at least one recurring insight")
	}

	var practice InsightEntry
	foundPractice := false
	for _, insight := range insights {
		if insight.Theme == "practice" {
			practice = insight
			foundPractice = true
			break
		}
	}
	if !foundPractice {
		t.Fatalf("expected practice theme, got %#v", insights)
	}
	if practice.EvidenceWeight <= 0 || practice.EchoScore <= 0 || practice.MardukScore <= 0 {
		t.Fatalf("expected non-zero evidence and hemisphere scores, got %#v", practice)
	}

	blog, ok := loop.GenerateBlogPost(practice.ID)
	if !ok {
		t.Fatalf("expected blog generation for insight %s", practice.ID)
	}
	if !strings.Contains(blog.Body, "practice") || blog.InsightID != practice.ID {
		t.Fatalf("blog body should reference the insight theme and id, got %#v", blog)
	}

	snapshot := loop.Snapshot()
	if snapshot.DiaryCount != 2 || snapshot.InsightCount == 0 || snapshot.BlogCount != 1 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if snapshot.Balance.IntegrationScore <= 0 {
		t.Fatalf("expected a positive Echo/Marduk integration score")
	}
}

func TestNarrativeLoopDoesNotDuplicateEquivalentInsights(t *testing.T) {
	loop := NewNarrativeLoop()
	loop.AddDiaryEntry(DiaryEntry{Tags: []string{"skill"}, Reflection: "learn skill practice", Outcome: "test skill"})
	loop.AddDiaryEntry(DiaryEntry{Tags: []string{"skill"}, Reflection: "practice skill again", Outcome: "verify skill"})

	first := loop.AnalyzeInsights(2)
	second := loop.AnalyzeInsights(2)
	if len(first) == 0 {
		t.Fatalf("expected first analysis to create an insight")
	}
	if len(second) != 0 {
		t.Fatalf("expected duplicate analysis to create no new insights, got %#v", second)
	}
}
