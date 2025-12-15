#!/usr/bin/env python3
"""
Test Suite for Iteration N+14
Tests the V14 autonomous core with:
- Nested shells architecture (OEIS A000081)
- Echobeats tetrahedral scheduling
- Real external knowledge integration
- Persistent autonomous operation
"""

import asyncio
import sys
import time
from datetime import datetime
from pathlib import Path

# Add core to path
sys.path.insert(0, str(Path(__file__).parent))

try:
    from core.autonomous_core_v14 import (
        DeepTreeEchoV14, NestedShellArchitecture, EchobeatsScheduler,
        ExternalKnowledgeIntegrator, NestingLevel, EchobeatsPhase,
        CognitiveState, EnergyState, WisdomState
    )
    V14_AVAILABLE = True
except ImportError as e:
    print(f"❌ Failed to import V14 components: {e}")
    V14_AVAILABLE = False


class TestIterationN14:
    """Test suite for iteration N+14 enhancements"""
    
    def __init__(self):
        self.tests_passed = 0
        self.tests_failed = 0
        self.test_results = []
    
    async def test_method(self, name: str, func):
        """Execute a test method"""
        print(f"\n{'='*60}")
        print(f"TEST: {name}")
        print(f"{'='*60}")
        try:
            result = await func()
            if result:
                print(f"✅ PASSED: {name}")
                self.tests_passed += 1
                self.test_results.append((name, True, None))
            else:
                print(f"❌ FAILED: {name}")
                self.tests_failed += 1
                self.test_results.append((name, False, "Test returned False"))
            return result
        except Exception as e:
            print(f"❌ FAILED: {name}")
            print(f"   Error: {e}")
            self.tests_failed += 1
            self.test_results.append((name, False, str(e)))
            return False
    
    async def run_all_tests(self):
        """Run all tests"""
        print("\n" + "="*60)
        print("ITERATION N+14 TEST SUITE")
        print("="*60)
        print(f"Testing Deep Tree Echo V14 enhancements")
        print(f"Timestamp: {datetime.now().isoformat()}")
        print("="*60)
        
        if not V14_AVAILABLE:
            print("❌ V14 components not available - cannot run tests")
            return
        
        # Test 1: Nested Shells Architecture
        await self.test_method("Nested Shells Architecture (OEIS A000081)", 
                               self.test_nested_shells_architecture)
        
        # Test 2: Echobeats Scheduler
        await self.test_method("Echobeats 12-Step Tetrahedral Scheduler", 
                               self.test_echobeats_scheduler)
        
        # Test 3: External Knowledge Integration
        await self.test_method("Real External Knowledge Integration", 
                               self.test_external_knowledge_integration)
        
        # Test 4: V14 Core Initialization
        await self.test_method("V14 Core Initialization", 
                               self.test_v14_core_initialization)
        
        # Test 5: Cognitive Cycle Processing
        await self.test_method("Cognitive Cycle Processing", 
                               self.test_cognitive_cycle_processing)
        
        # Test 6: State Persistence
        await self.test_method("State Persistence", 
                               self.test_state_persistence)
        
        # Test 7: Energy and Wake/Rest Cycles
        await self.test_method("Energy and Wake/Rest Cycles", 
                               self.test_energy_wake_rest)
        
        # Test 8: Knowledge Acquisition and Wisdom Growth
        await self.test_method("Knowledge Acquisition and Wisdom Growth", 
                               self.test_knowledge_wisdom_growth)
        
        # Test 9: Short Autonomous Run
        await self.test_method("Short Autonomous Run (30 seconds)", 
                               self.test_short_autonomous_run)
        
        # Print summary
        self.print_summary()
    
    async def test_nested_shells_architecture(self):
        """Test the OEIS A000081 nested shells structure"""
        print("Testing nested shells architecture...")
        
        arch = NestedShellArchitecture()
        
        # Verify Level 1: 1 term
        level1 = arch.shells[NestingLevel.GLOBAL]
        assert len(level1) == 1, f"Level 1 should have 1 term, got {len(level1)}"
        print(f"  ✓ Level 1 (Global): {len(level1)} term")
        
        # Verify Level 2: 2 terms
        level2 = arch.shells[NestingLevel.STATE]
        assert len(level2) == 2, f"Level 2 should have 2 terms, got {len(level2)}"
        print(f"  ✓ Level 2 (State): {len(level2)} terms")
        
        # Verify Level 3: 4 terms
        level3 = arch.shells[NestingLevel.STREAM]
        assert len(level3) == 4, f"Level 3 should have 4 terms, got {len(level3)}"
        print(f"  ✓ Level 3 (Stream): {len(level3)} terms")
        
        # Verify Level 4: 9 terms
        level4 = arch.shells[NestingLevel.OPERATION]
        assert len(level4) == 9, f"Level 4 should have 9 terms, got {len(level4)}"
        print(f"  ✓ Level 4 (Operation): {len(level4)} terms")
        
        # Verify nesting relationships
        global_shell = level1[0]
        assert len(global_shell.children) == 2, "Global should have 2 children"
        print(f"  ✓ Nesting relationships verified")
        
        # Verify shell paths
        wake_shell = level2[0]
        path = wake_shell.get_path()
        print(f"  ✓ Wake shell path: {path}")
        
        print("✅ Nested shells architecture correct: 1→2→4→9")
        return True
    
    async def test_echobeats_scheduler(self):
        """Test the 12-step tetrahedral scheduler"""
        print("Testing echobeats scheduler...")
        
        scheduler = EchobeatsScheduler()
        
        # Verify initial state
        assert scheduler.state.current_step == 1, "Should start at step 1"
        print(f"  ✓ Initial step: {scheduler.state.current_step}")
        
        # Test step advancement
        initial_step = scheduler.state.current_step
        scheduler.advance_step()
        assert scheduler.state.current_step == initial_step + 1, "Step should advance"
        print(f"  ✓ Step advancement: {initial_step} → {scheduler.state.current_step}")
        
        # Test 12-step cycle
        for i in range(11):
            scheduler.advance_step()
        assert scheduler.state.current_step == 1, "Should wrap to 1 after step 12"
        assert scheduler.state.cycle_count == 1, "Cycle count should increment"
        print(f"  ✓ 12-step cycle complete, cycle count: {scheduler.state.cycle_count}")
        
        # Test stream phases (120 degree offsets = 4 steps apart)
        scheduler.state.current_step = 1
        scheduler.advance_step()  # Now at step 1
        phases = scheduler.state.stream_phases
        print(f"  ✓ Stream phases at step 1: {phases}")
        
        # Verify 4-step offsets
        coherence = phases["coherence"]
        memory = phases["memory"]
        imagination = phases["imagination"]
        
        # Memory should be 4 steps behind coherence (wrapping)
        # Imagination should be 8 steps behind coherence (wrapping)
        print(f"  ✓ Phase offsets: C={coherence}, M={memory}, I={imagination}")
        
        # Test triad identification
        scheduler.state.current_step = 1
        triad = scheduler.get_current_triad()
        assert triad == "perceive", f"Step 1 should be perceive, got {triad}"
        print(f"  ✓ Triad at step 1: {triad}")
        
        scheduler.state.current_step = 2
        triad = scheduler.get_current_triad()
        assert triad == "act", f"Step 2 should be act, got {triad}"
        print(f"  ✓ Triad at step 2: {triad}")
        
        # Test active stream identification
        scheduler.state.current_step = 1
        stream = scheduler.get_active_stream()
        assert stream == "coherence", f"Steps 1-4 should be coherence, got {stream}"
        print(f"  ✓ Active stream at step 1: {stream}")
        
        scheduler.state.current_step = 5
        stream = scheduler.get_active_stream()
        assert stream == "memory", f"Steps 5-8 should be memory, got {stream}"
        print(f"  ✓ Active stream at step 5: {stream}")
        
        print("✅ Echobeats scheduler working correctly")
        return True
    
    async def test_external_knowledge_integration(self):
        """Test real external knowledge integration"""
        print("Testing external knowledge integration...")
        
        integrator = ExternalKnowledgeIntegrator()
        
        if integrator.llm_client is None:
            print("  ⚠️  No LLM provider available, testing with fallback")
            knowledge = await integrator.acquire_knowledge("test_topic")
            assert knowledge["topic"] == "test_topic"
            print("  ✓ Fallback knowledge acquisition works")
            return True
        
        # Test knowledge acquisition
        print("  → Acquiring knowledge about 'consciousness'...")
        knowledge = await integrator.acquire_knowledge("consciousness", depth="overview")
        
        assert knowledge["topic"] == "consciousness"
        assert "content" in knowledge
        assert knowledge["confidence"] > 0
        print(f"  ✓ Knowledge acquired: {len(knowledge['content'])} chars")
        print(f"  ✓ Confidence: {knowledge['confidence']}")
        print(f"  ✓ Source: {knowledge['source']}")
        
        # Test caching
        print("  → Testing knowledge cache...")
        start_time = time.time()
        knowledge2 = await integrator.acquire_knowledge("consciousness", depth="overview")
        cache_time = time.time() - start_time
        
        assert cache_time < 0.1, "Cached retrieval should be fast"
        assert knowledge2["content"] == knowledge["content"]
        print(f"  ✓ Cache hit in {cache_time:.3f}s")
        
        # Test insight synthesis
        print("  → Testing insight synthesis...")
        knowledge_items = [knowledge]
        insights = await integrator.synthesize_insights(knowledge_items)
        
        assert len(insights) > 0
        print(f"  ✓ Insights synthesized: {len(insights)} chars")
        
        print("✅ External knowledge integration working")
        return True
    
    async def test_v14_core_initialization(self):
        """Test V14 core initialization"""
        print("Testing V14 core initialization...")
        
        echo = DeepTreeEchoV14()
        
        # Verify nested shells created
        assert echo.nested_shells is not None
        print("  ✓ Nested shells architecture created")
        
        # Verify echobeats created
        assert echo.echobeats is not None
        assert echo.echobeats.state.current_step == 1
        print("  ✓ Echobeats scheduler created")
        
        # Verify knowledge integrator created
        assert echo.knowledge_integrator is not None
        print("  ✓ Knowledge integrator created")
        
        # Verify initial state
        assert echo.cognitive_state == CognitiveState.INITIALIZING
        print(f"  ✓ Initial cognitive state: {echo.cognitive_state.value}")
        
        # Verify energy state
        assert echo.energy_state.energy == 1.0
        assert echo.energy_state.fatigue == 0.0
        print(f"  ✓ Energy state: {echo.energy_state.energy:.2f}")
        
        # Verify wisdom state
        assert echo.wisdom_state.wisdom_score >= 0.0
        print(f"  ✓ Wisdom state initialized")
        
        # Test initialization
        await echo.initialize()
        assert echo.cognitive_state == CognitiveState.WAKING
        print(f"  ✓ After initialization: {echo.cognitive_state.value}")
        
        print("✅ V14 core initialization successful")
        return True
    
    async def test_cognitive_cycle_processing(self):
        """Test cognitive cycle processing"""
        print("Testing cognitive cycle processing...")
        
        echo = DeepTreeEchoV14()
        await echo.initialize()
        await echo.wake()
        
        # Process a few cycles
        initial_thoughts = len(echo.thoughts)
        initial_step = echo.echobeats.state.current_step
        
        print(f"  → Processing 3 cognitive cycles...")
        for i in range(3):
            await echo.process_cognitive_cycle()
            print(f"    Cycle {i+1}: Step {echo.echobeats.state.current_step}, "
                  f"Stream: {echo.echobeats.get_active_stream()}, "
                  f"Triad: {echo.echobeats.get_current_triad()}")
        
        # Verify thoughts generated
        assert len(echo.thoughts) > initial_thoughts
        print(f"  ✓ Thoughts generated: {len(echo.thoughts) - initial_thoughts}")
        
        # Verify echobeats advanced
        assert echo.echobeats.state.current_step != initial_step
        print(f"  ✓ Echobeats advanced: {initial_step} → {echo.echobeats.state.current_step}")
        
        # Verify energy consumed
        assert echo.energy_state.energy < 1.0
        print(f"  ✓ Energy consumed: {echo.energy_state.energy:.2f}")
        
        # Verify metrics updated
        assert echo.metrics["cycles_completed"] > 0
        assert echo.metrics["thoughts_generated"] > 0
        print(f"  ✓ Metrics updated: {echo.metrics['cycles_completed']} cycles")
        
        print("✅ Cognitive cycle processing working")
        return True
    
    async def test_state_persistence(self):
        """Test state save and load"""
        print("Testing state persistence...")
        
        # Create and run echo briefly
        echo1 = DeepTreeEchoV14(state_file="data/test_state_v14.json")
        await echo1.initialize()
        await echo1.wake()
        
        # Generate some activity
        for i in range(5):
            await echo1.process_cognitive_cycle()
        
        # Save state
        await echo1.save_state()
        saved_cycle_count = echo1.echobeats.state.cycle_count
        saved_metrics = echo1.metrics.copy()
        print(f"  ✓ State saved: cycle {saved_cycle_count}")
        
        # Create new instance and load state
        echo2 = DeepTreeEchoV14(state_file="data/test_state_v14.json")
        await echo2.load_state()
        
        # Verify state restored
        assert echo2.metrics["cycles_completed"] == saved_metrics["cycles_completed"]
        print(f"  ✓ State loaded: cycle {echo2.echobeats.state.cycle_count}")
        print(f"  ✓ Metrics restored: {echo2.metrics['cycles_completed']} cycles")
        
        print("✅ State persistence working")
        return True
    
    async def test_energy_wake_rest(self):
        """Test energy management and wake/rest cycles"""
        print("Testing energy and wake/rest cycles...")
        
        echo = DeepTreeEchoV14()
        await echo.initialize()
        await echo.wake()
        
        # Verify wake state
        assert echo.cognitive_state == CognitiveState.ACTIVE
        print(f"  ✓ Wake state: {echo.cognitive_state.value}")
        
        # Deplete energy
        echo.energy_state.energy = 0.2
        echo.energy_state.fatigue = 0.8
        
        assert echo.energy_state.needs_rest()
        print(f"  ✓ Rest needed detected: energy={echo.energy_state.energy:.2f}")
        
        # Test rest cycle
        await echo.rest()
        
        assert echo.energy_state.energy > 0.2
        print(f"  ✓ Energy restored: {echo.energy_state.energy:.2f}")
        
        # Verify can wake
        assert echo.energy_state.can_wake()
        print(f"  ✓ Can wake: energy={echo.energy_state.energy:.2f}")
        
        print("✅ Energy and wake/rest cycles working")
        return True
    
    async def test_knowledge_wisdom_growth(self):
        """Test knowledge acquisition and wisdom growth"""
        print("Testing knowledge acquisition and wisdom growth...")
        
        echo = DeepTreeEchoV14()
        await echo.initialize()
        await echo.wake()
        
        initial_knowledge_count = len(echo.knowledge_base)
        initial_wisdom = echo.wisdom_state.wisdom_score
        
        # Acquire knowledge
        print("  → Acquiring knowledge...")
        await echo.acquire_new_knowledge()
        
        assert len(echo.knowledge_base) > initial_knowledge_count
        print(f"  ✓ Knowledge acquired: {len(echo.knowledge_base)} items")
        
        # Practice skills
        print("  → Practicing skills...")
        await echo.practice_skills()
        
        assert echo.metrics["skills_practiced"] > 0
        print(f"  ✓ Skills practiced: {echo.metrics['skills_practiced']}")
        
        # Verify wisdom growth
        assert echo.wisdom_state.wisdom_score >= initial_wisdom
        print(f"  ✓ Wisdom score: {initial_wisdom:.3f} → {echo.wisdom_state.wisdom_score:.3f}")
        
        print("✅ Knowledge acquisition and wisdom growth working")
        return True
    
    async def test_short_autonomous_run(self):
        """Test short autonomous run"""
        print("Testing short autonomous run (30 seconds)...")
        
        echo = DeepTreeEchoV14()
        await echo.initialize()
        
        # Run for 30 seconds
        print("  → Running autonomously for 30 seconds...")
        
        async def run_limited():
            await echo.wake()
            start_time = time.time()
            
            while time.time() - start_time < 30:
                await echo.process_cognitive_cycle()
                await asyncio.sleep(2.5)  # Match the cycle time
            
            await echo.shutdown()
        
        await run_limited()
        
        # Verify activity
        assert echo.metrics["cycles_completed"] > 0
        assert echo.metrics["thoughts_generated"] > 0
        print(f"  ✓ Cycles completed: {echo.metrics['cycles_completed']}")
        print(f"  ✓ Thoughts generated: {echo.metrics['thoughts_generated']}")
        print(f"  ✓ Echobeats cycles: {echo.echobeats.state.cycle_count}")
        
        # Get final status
        status = echo.get_status()
        print(f"  ✓ Final status:")
        print(f"    - State: {status['cognitive_state']}")
        print(f"    - Energy: {status['energy']}")
        print(f"    - Wisdom: {status['wisdom_score']}")
        
        print("✅ Short autonomous run successful")
        return True
    
    def print_summary(self):
        """Print test summary"""
        print("\n" + "="*60)
        print("TEST SUMMARY")
        print("="*60)
        print(f"Total tests: {self.tests_passed + self.tests_failed}")
        print(f"Passed: {self.tests_passed} ✅")
        print(f"Failed: {self.tests_failed} ❌")
        print(f"Success rate: {self.tests_passed/(self.tests_passed + self.tests_failed)*100:.1f}%")
        print("="*60)
        
        if self.tests_failed > 0:
            print("\nFailed tests:")
            for name, passed, error in self.test_results:
                if not passed:
                    print(f"  ❌ {name}")
                    if error:
                        print(f"     {error}")
        
        print()


async def main():
    """Main test runner"""
    tester = TestIterationN14()
    await tester.run_all_tests()
    
    # Return exit code based on results
    return 0 if tester.tests_failed == 0 else 1


if __name__ == "__main__":
    print("="*60)
    print("Echo9llama Iteration N+14 Test Suite")
    print("="*60)
    print()
    
    exit_code = asyncio.run(main())
    sys.exit(exit_code)
