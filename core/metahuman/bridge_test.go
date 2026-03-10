package metahuman

import (
	"math"
	"testing"
	"time"

	"github.com/o9nn/echo.go/core/deeptreeecho"
)

func TestNewFACSState(t *testing.T) {
	facs := NewFACSState()
	if facs == nil {
		t.Fatal("NewFACSState returned nil")
	}
	// All AUs should be zero
	for au := ActionUnit(0); au < AUCount; au++ {
		if facs.Get(au) != 0.0 {
			t.Errorf("AU%d should be 0.0, got %f", au, facs.Get(au))
		}
	}
}

func TestFACSSetGet(t *testing.T) {
	facs := NewFACSState()
	facs.Set(AU12, 0.8)
	if got := facs.Get(AU12); got != 0.8 {
		t.Errorf("Expected AU12=0.8, got %f", got)
	}
	// Test clamping
	facs.Set(AU6, 1.5)
	if got := facs.Get(AU6); got != 1.0 {
		t.Errorf("Expected AU6 clamped to 1.0, got %f", got)
	}
	facs.Set(AU4, -0.5)
	if got := facs.Get(AU4); got != 0.0 {
		t.Errorf("Expected AU4 clamped to 0.0, got %f", got)
	}
}

func TestFACSAdd(t *testing.T) {
	facs := NewFACSState()
	facs.Set(AU12, 0.5)
	facs.Add(AU12, 0.3)
	if got := facs.Get(AU12); got != 0.8 {
		t.Errorf("Expected AU12=0.8 after add, got %f", got)
	}
	// Test clamping on add
	facs.Add(AU12, 0.5)
	if got := facs.Get(AU12); got != 1.0 {
		t.Errorf("Expected AU12 clamped to 1.0, got %f", got)
	}
}

func TestFACSSnapshot(t *testing.T) {
	facs := NewFACSState()
	facs.Set(AU6, 0.7)
	facs.Set(AU12, 0.9)
	snap := facs.Snapshot()
	if len(snap) != 2 {
		t.Errorf("Expected 2 active AUs, got %d", len(snap))
	}
	if snap[AU6] != 0.7 {
		t.Errorf("Expected AU6=0.7 in snapshot, got %f", snap[AU6])
	}
}

func TestFACSToMorphTargets(t *testing.T) {
	facs := NewFACSState()
	facs.Set(AU12, 0.8)
	facs.Set(AU6, 0.6)
	targets := facs.ToMorphTargets()
	if targets["CTRL_mouth_cornerPull"] != 0.8 {
		t.Errorf("Expected CTRL_mouth_cornerPull=0.8, got %f", targets["CTRL_mouth_cornerPull"])
	}
	if targets["CTRL_cheek_raise"] != 0.6 {
		t.Errorf("Expected CTRL_cheek_raise=0.6, got %f", targets["CTRL_cheek_raise"])
	}
}

func TestFACSReset(t *testing.T) {
	facs := NewFACSState()
	facs.Set(AU12, 0.8)
	facs.Set(AU6, 0.6)
	facs.Reset()
	snap := facs.Snapshot()
	if len(snap) != 0 {
		t.Errorf("Expected 0 active AUs after reset, got %d", len(snap))
	}
}

func TestLorenzAttractor(t *testing.T) {
	la := NewLorenzAttractor()
	if la == nil {
		t.Fatal("NewLorenzAttractor returned nil")
	}
	// Step several times
	for i := 0; i < 1000; i++ {
		x, y, z := la.Step()
		if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) {
			t.Fatalf("NaN at step %d", i)
		}
	}
	if !la.IsHealthy() {
		t.Error("Attractor should be healthy after 1000 steps")
	}
}

func TestLorenzLyapunov(t *testing.T) {
	la := NewLorenzAttractor()
	// Run enough steps for Lyapunov estimation
	for i := 0; i < 5000; i++ {
		la.Step()
	}
	lyap := la.LyapunovExponent()
	// Lorenz attractor should have positive Lyapunov exponent (~0.9)
	if lyap <= 0 {
		t.Errorf("Expected positive Lyapunov exponent, got %f", lyap)
	}
}

func TestLorenzMicroExpressions(t *testing.T) {
	la := NewLorenzAttractor()
	facs := NewFACSState()
	// Warm up attractor
	for i := 0; i < 100; i++ {
		la.Step()
	}
	la.ApplyMicroExpressions(facs)
	snap := facs.Snapshot()
	if len(snap) == 0 {
		t.Error("Expected some AU activations from micro-expressions")
	}
}

func TestLorenzReset(t *testing.T) {
	la := NewLorenzAttractor()
	for i := 0; i < 100; i++ {
		la.Step()
	}
	la.Reset()
	if la.X != 1.0 || la.Y != 1.0 || la.Z != 1.0 {
		t.Error("Reset should restore initial conditions")
	}
}

func TestEndocrineExpressionMapper(t *testing.T) {
	mapper := NewEndocrineExpressionMapper()
	facs := NewFACSState()
	// Simulate high dopamine (phasic) → should produce smile
	concentrations := map[deeptreeecho.Hormone]float64{
		deeptreeecho.DopaminePhasic: 0.9,
	}
	mapper.MapToFACS(concentrations, facs)
	// AU12 (smile) should be activated
	if facs.Get(AU12) < 0.5 {
		t.Errorf("High phasic dopamine should produce strong smile, got AU12=%f", facs.Get(AU12))
	}
	// AU6 (cheek raise) should also be activated
	if facs.Get(AU6) < 0.3 {
		t.Errorf("High phasic dopamine should produce cheek raise, got AU6=%f", facs.Get(AU6))
	}
}

func TestEndocrineCortisolStress(t *testing.T) {
	mapper := NewEndocrineExpressionMapper()
	facs := NewFACSState()
	concentrations := map[deeptreeecho.Hormone]float64{
		deeptreeecho.Cortisol: 0.8,
	}
	mapper.MapToFACS(concentrations, facs)
	// AU4 (brow lowerer) should be strongly activated
	if facs.Get(AU4) < 0.5 {
		t.Errorf("High cortisol should produce brow lowering, got AU4=%f", facs.Get(AU4))
	}
	// AU15 (lip corner depress) should be activated
	if facs.Get(AU15) < 0.2 {
		t.Errorf("High cortisol should produce lip corner depress, got AU15=%f", facs.Get(AU15))
	}
}

func TestCognitiveModePreset(t *testing.T) {
	facs := NewFACSState()
	CognitiveModePreset(deeptreeecho.ModeSocial, facs)
	// Social mode should produce warm expression
	if facs.Get(AU6) < 0.5 {
		t.Errorf("Social mode should produce cheek raise, got AU6=%f", facs.Get(AU6))
	}
	if facs.Get(AU12) < 0.4 {
		t.Errorf("Social mode should produce smile, got AU12=%f", facs.Get(AU12))
	}
}

func TestValenceArousalToFACS(t *testing.T) {
	// Positive valence → smile
	facs := NewFACSState()
	ValenceArousalToFACS(0.8, 0.5, facs)
	if facs.Get(AU12) < 0.4 {
		t.Errorf("Positive valence should produce smile, got AU12=%f", facs.Get(AU12))
	}
	// Negative valence → frown
	facs2 := NewFACSState()
	ValenceArousalToFACS(-0.8, 0.5, facs2)
	if facs2.Get(AU15) < 0.2 {
		t.Errorf("Negative valence should produce lip corner depress, got AU15=%f", facs2.Get(AU15))
	}
}

func TestAestheticParameters(t *testing.T) {
	ap := DefaultAestheticParameters()
	facs := NewFACSState()
	facs.Set(AU4, 0.8) // High brow lowering (stress)
	ap.ConfidencePosture = 0.9
	ap.ApplyToFACS(facs)
	// Confidence should reduce brow lowering
	if facs.Get(AU4) >= 0.8 {
		t.Errorf("Confidence should reduce brow lowering, got AU4=%f", facs.Get(AU4))
	}
	// Should add chin raise
	if facs.Get(AU17) < 0.1 {
		t.Errorf("Confidence should add chin raise, got AU17=%f", facs.Get(AU17))
	}
}

func TestAestheticMaterialParameters(t *testing.T) {
	ap := DefaultAestheticParameters()
	params := ap.MaterialParameters()
	if params["EyeSparkleIntensity"] != ap.EyeSparkle {
		t.Error("EyeSparkleIntensity should match EyeSparkle parameter")
	}
	if params["IrisSpecular"] != ap.EyeSparkle*2.0 {
		t.Error("IrisSpecular should be 2x EyeSparkle")
	}
}

func TestCompositeExpression(t *testing.T) {
	facs := NewFACSState()
	ap := DefaultAestheticParameters()
	ok := ApplyCompositeExpression("GenuineSmile", facs, &ap)
	if !ok {
		t.Error("GenuineSmile should be a valid preset")
	}
	if facs.Get(AU6) < 0.5 {
		t.Error("GenuineSmile should activate AU6")
	}
	if facs.Get(AU12) < 0.5 {
		t.Error("GenuineSmile should activate AU12")
	}
}

func TestDNACognitiveBridge(t *testing.T) {
	bridge := NewDNACognitiveBridge()
	if bridge == nil {
		t.Fatal("NewDNACognitiveBridge returned nil")
	}
	// Simple update without endocrine system
	frame := bridge.UpdateSimple(0.7, 0.5, 0.3)
	if frame.Frame != 1 {
		t.Errorf("Expected frame 1, got %d", frame.Frame)
	}
	if len(frame.MorphTargets) == 0 {
		t.Error("Expected non-empty morph targets")
	}
	if len(frame.MaterialParams) == 0 {
		t.Error("Expected non-empty material params")
	}
}

func TestDNACognitiveBridgeMultipleFrames(t *testing.T) {
	bridge := NewDNACognitiveBridge()
	var lastFrame ExpressionFrame
	for i := 0; i < 100; i++ {
		lastFrame = bridge.UpdateSimple(0.5, 0.5, 0.3)
		time.Sleep(time.Millisecond) // Small delay for delta time
	}
	if lastFrame.Frame != 100 {
		t.Errorf("Expected frame 100, got %d", lastFrame.Frame)
	}
	if !bridge.IsHealthy() {
		t.Error("Bridge should be healthy after 100 frames")
	}
}

func TestDNACognitiveBridgeMetrics(t *testing.T) {
	bridge := NewDNACognitiveBridge()
	bridge.UpdateSimple(0.5, 0.5, 0.3)
	metrics := bridge.Metrics()
	if metrics["frame_count"].(uint64) != 1 {
		t.Error("Expected frame_count=1")
	}
	if metrics["enable_chaos"].(bool) != true {
		t.Error("Expected chaos enabled")
	}
}

func TestDNACognitiveBridgeWithEndocrine(t *testing.T) {
	bridge := NewDNACognitiveBridge()
	ves := deeptreeecho.NewVirtualEndocrineSystem()
	// Signal a reward event to boost dopamine
	ves.SignalEvent(deeptreeecho.EndoRewardReceived, 0.8)
	frame := bridge.Update(ves, deeptreeecho.ModeSocial, 0.3, 0.7, 0.5)
	// Should have smile-related morph targets
	if frame.MorphTargets["CTRL_mouth_cornerPull"] < 0.1 {
		t.Error("Dopamine + social mode should produce smile morph target")
	}
}

func TestActionUnitNames(t *testing.T) {
	if len(ActionUnitNames) != int(AUCount) {
		t.Errorf("Expected %d AU names, got %d", AUCount, len(ActionUnitNames))
	}
}

func TestMetaHumanMorphTargetMapping(t *testing.T) {
	if len(MetaHumanMorphTarget) != int(AUCount) {
		t.Errorf("Expected %d morph target mappings, got %d", AUCount, len(MetaHumanMorphTarget))
	}
	// Verify key mappings
	if MetaHumanMorphTarget[AU12] != "CTRL_mouth_cornerPull" {
		t.Error("AU12 should map to CTRL_mouth_cornerPull")
	}
	if MetaHumanMorphTarget[AU6] != "CTRL_cheek_raise" {
		t.Error("AU6 should map to CTRL_cheek_raise")
	}
}

func BenchmarkBridgeUpdate(b *testing.B) {
	bridge := NewDNACognitiveBridge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.UpdateSimple(0.5, 0.5, 0.3)
	}
}

func BenchmarkLorenzStep(b *testing.B) {
	la := NewLorenzAttractor()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Step()
	}
}

func BenchmarkFACSToMorphTargets(b *testing.B) {
	facs := NewFACSState()
	facs.Set(AU6, 0.7)
	facs.Set(AU12, 0.8)
	facs.Set(AU5, 0.5)
	facs.Set(AU25, 0.3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		facs.ToMorphTargets()
	}
}
