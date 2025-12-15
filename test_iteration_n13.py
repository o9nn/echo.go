#!/usr/bin/env python3
"""
Test Suite for Iteration N+13
Tests the enhanced autonomous core V13 with deeper stream processing,
EchoDream-controlled wake/rest, and external knowledge integration.
"""

import asyncio
import sys
import time
from datetime import datetime
from pathlib import Path

# Add core to path
sys.path.insert(0, str(Path(__file__).parent))

try:
    from core.autonomous_core_v13 import (
        DeepTreeEchoV13, WisdomState, EchoDreamWakeRestController,
        ExternalKnowledgeIntegrator, EnhancedStreamOrchestrator,
        DreamInsight, StreamType
    )
    from core.autonomous_core_v12 import (
        EnergyState, InterestPatternSystem, LLMProvider,
        HypergraphMemoryFallback
    )
    V13_AVAILABLE = True
except ImportError as e:
    print(f"❌ Failed to import V13 components: {e}")
    V13_AVAILABLE = False


class TestIterationN13:
    """Test suite for iteration N+13 enhancements"""
    
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
        print("ITERATION N+13 TEST SUITE")
        print("="*60)
        print(f"Testing Deep Tree Echo V13 enhancements")
        print(f"Timestamp: {datetime.now().isoformat()}")
        print("="*60)
        
        if not V13_AVAILABLE:
            print("❌ V13 components not available - cannot run tests")
            return
        
        # Test 1: Wisdom State Tracking
        await self.test_method("Wisdom State Tracking", self.test_wisdom_state_tracking)
        
        # Test 2: Enhanced Stream Orchestrator
        await self.test_method("Enhanced Stream Orchestrator", self.test_enhanced_stream_orchestrator)
        
        # Test 3: EchoDream Wake/Rest Controller
        await self.test_method("EchoDream Wake/Rest Controller", self.test_echodream_wake_rest_controller)
        
        # Test 4: External Knowledge Integrator
        await self.test_method("External Knowledge Integrator", self.test_external_knowledge_integrator)
        
        # Test 5: Dream Insights Generation
        await self.test_method("Dream Insights Generation", self.test_dream_insights_generation)
        
        # Test 6: V13 Core Initialization
        await self.test_method("V13 Core Initialization", self.test_v13_core_initialization)
        
        # Test 7: Single Cognitive Cycle
        await self.test_method("Single Cognitive Cycle", self.test_single_cognitive_cycle)
        
        # Test 8: Wisdom Metrics Update
        await self.test_method("Wisdom Metrics Update", self.test_wisdom_metrics_update)
        
        # Test 9: Integration Test
        await self.test_method("Integration Test - Multiple Cycles", self.test_integration)
        
        # Print summary
        self.print_summary()
    
    async def test_wisdom_state_tracking(self):
        """Test wisdom state tracking and metrics"""
        wisdom = WisdomState()
        
        # Initial state
        assert wisdom.knowledge_depth == 0.0
        assert wisdom.total_insights == 0
        
        # Simulate learning
        wisdom.total_knowledge_acquired = 50
        wisdom.knowledge_depth = 0.5
        wisdom.total_insights = 10
        wisdom.insight_frequency = 2.0
        wisdom.reasoning_quality = 0.6
        wisdom.behavioral_coherence = 0.7
        
        # Calculate overall wisdom
        overall = wisdom.get_overall_wisdom()
        print(f"Overall wisdom score: {overall:.3f}")
        
        assert 0.0 <= overall <= 1.0
        assert overall > 0.5  # Should be decent with these values
        
        return True
    
    async def test_enhanced_stream_orchestrator(self):
        """Test enhanced stream orchestrator with deeper processing"""
        interest_system = InterestPatternSystem()
        llm = LLMProvider()
        memory = HypergraphMemoryFallback()
        
        orchestrator = EnhancedStreamOrchestrator(llm, memory, interest_system)
        
        # Test stream initialization
        assert len(orchestrator.streams) == 3
        assert StreamType.COHERENCE_STREAM in orchestrator.streams
        assert StreamType.MEMORY_STREAM in orchestrator.streams
        assert StreamType.IMAGINATION_STREAM in orchestrator.streams
        
        # Test stream processing
        energy = EnergyState()
        
        print("Testing coherence stream...")
        coherence_result = await orchestrator.process_coherence_stream(energy)
        assert "thought" in coherence_result
        assert coherence_result["present_awareness"] == True
        print(f"Coherence thought: {coherence_result['thought'][:60]}...")
        
        print("Testing memory stream...")
        memory_result = await orchestrator.process_memory_stream(energy)
        assert "thought" in memory_result
        assert "patterns_identified" in memory_result
        print(f"Memory thought: {memory_result['thought'][:60]}...")
        
        print("Testing imagination stream...")
        imagination_result = await orchestrator.process_imagination_stream(energy)
        assert "thought" in imagination_result
        assert imagination_result["creative_mode"] == True
        print(f"Imagination thought: {imagination_result['thought'][:60]}...")
        
        # Test stream states
        states = orchestrator.get_stream_states()
        assert "streams" in states
        assert "coherence_context" in states
        
        return True
    
    async def test_echodream_wake_rest_controller(self):
        """Test EchoDream-based wake/rest control"""
        controller = EchoDreamWakeRestController()
        
        # Test with low energy (should rest)
        energy_low = EnergyState(energy=0.2, fatigue=0.8)
        should_rest = await controller.should_rest(energy_low, cycles_completed=10)
        print(f"Should rest with low energy: {should_rest}")
        assert should_rest == True
        
        # Test with high energy (should not rest)
        energy_high = EnergyState(energy=0.9, fatigue=0.1)
        should_rest = await controller.should_rest(energy_high, cycles_completed=5)
        print(f"Should rest with high energy: {should_rest}")
        
        # Test wake decision
        should_wake = await controller.should_wake(energy_high)
        print(f"Should wake with high energy: {should_wake}")
        assert should_wake == True
        
        should_wake = await controller.should_wake(energy_low)
        print(f"Should wake with low energy: {should_wake}")
        assert should_wake == False
        
        return True
    
    async def test_external_knowledge_integrator(self):
        """Test external knowledge integration"""
        interest_system = InterestPatternSystem()
        interest_system.update_interest("consciousness", 0.8)
        interest_system.update_interest("learning", 0.7)
        
        integrator = ExternalKnowledgeIntegrator(interest_system)
        
        # Test topic search
        result = await integrator.search_topic("consciousness")
        assert result is not None
        assert "topic" in result
        assert result["topic"] == "consciousness"
        print(f"Search result: {result}")
        
        # Test knowledge acquisition for interests
        acquired = await integrator.acquire_knowledge_for_interests()
        print(f"Acquired {len(acquired)} knowledge items")
        
        return True
    
    async def test_dream_insights_generation(self):
        """Test dream insight generation"""
        controller = EchoDreamWakeRestController()
        
        # Generate dream insights
        insight = await controller.generate_dream_insights([])
        
        if insight:
            print(f"Dream insight: {insight.content}")
            print(f"Importance: {insight.importance:.2f}")
            print(f"Affects wake/rest: {insight.affects_wake_rest}")
            
            if insight.suggested_skills:
                print(f"Suggested skills: {insight.suggested_skills}")
            
            if insight.suggested_interests:
                print(f"Suggested interests: {insight.suggested_interests}")
            
            assert isinstance(insight, DreamInsight)
            assert 0.0 <= insight.importance <= 1.0
            
            # Check dream history
            assert len(controller.dream_history) == 1
            assert controller.last_dream == insight
        
        return True
    
    async def test_v13_core_initialization(self):
        """Test V13 core initialization"""
        echo = DeepTreeEchoV13()
        
        # Check components initialized
        assert echo.energy is not None
        assert echo.wisdom is not None
        assert echo.interest_system is not None
        assert echo.llm is not None
        assert echo.memory is not None
        assert echo.stream_orchestrator is not None
        assert echo.wake_rest_controller is not None
        assert echo.knowledge_integrator is not None
        
        # Check initial state
        assert echo.state.value == "initializing"
        assert echo.running == False
        assert echo.stats["cycles"] == 0
        
        # Check wisdom state
        assert echo.wisdom.get_overall_wisdom() == 0.0
        
        print("✅ All V13 components initialized successfully")
        
        return True
    
    async def test_single_cognitive_cycle(self):
        """Test a single cognitive cycle execution"""
        echo = DeepTreeEchoV13()
        echo.running = True
        
        initial_cycles = echo.stats["cycles"]
        initial_thoughts = echo.stats["thoughts"]
        
        print("Executing one cognitive cycle...")
        await echo._cognitive_cycle()
        
        # Check that cycle completed
        assert echo.stats["cycles"] == initial_cycles + 1
        assert echo.stats["thoughts"] > initial_thoughts
        
        print(f"Cycle completed: {echo.stats['cycles']} total cycles")
        print(f"Thoughts generated: {echo.stats['thoughts']}")
        
        echo.running = False
        
        return True
    
    async def test_wisdom_metrics_update(self):
        """Test wisdom metrics updating"""
        echo = DeepTreeEchoV13()
        
        # Simulate activity
        echo.stats["cycles"] = 50
        echo.stats["thoughts"] = 200
        echo.stats["insights"] = 10
        echo.stats["knowledge_acquired"] = 20
        
        echo.wisdom.total_insights = 10
        echo.wisdom.total_knowledge_acquired = 20
        
        # Update metrics
        echo._update_wisdom_metrics()
        
        # Check metrics updated
        assert echo.wisdom.reasoning_quality > 0.0
        assert echo.wisdom.behavioral_coherence > 0.0
        assert echo.wisdom.knowledge_depth > 0.0
        
        overall_wisdom = echo.wisdom.get_overall_wisdom()
        print(f"Overall wisdom: {overall_wisdom:.3f}")
        print(f"Knowledge depth: {echo.wisdom.knowledge_depth:.3f}")
        print(f"Reasoning quality: {echo.wisdom.reasoning_quality:.3f}")
        
        assert overall_wisdom > 0.0
        
        return True
    
    async def test_integration(self):
        """Integration test with multiple cycles"""
        echo = DeepTreeEchoV13()
        echo.running = True
        
        print("Running 3 cognitive cycles...")
        
        for i in range(3):
            print(f"\nCycle {i+1}/3")
            await echo._cognitive_cycle()
            await asyncio.sleep(0.5)
        
        # Verify execution
        assert echo.stats["cycles"] == 3
        assert echo.stats["thoughts"] > 0
        
        # Test rest cycle
        print("\nTesting rest cycle...")
        await echo._rest_cycle()
        
        assert echo.stats["rest_periods"] == 1
        assert echo.energy.energy > 0.5  # Should be restored
        
        print(f"\nFinal stats:")
        print(f"  Cycles: {echo.stats['cycles']}")
        print(f"  Thoughts: {echo.stats['thoughts']}")
        print(f"  Insights: {echo.stats['insights']}")
        print(f"  Rest periods: {echo.stats['rest_periods']}")
        print(f"  Overall wisdom: {echo.wisdom.get_overall_wisdom():.3f}")
        
        echo.running = False
        
        return True
    
    def print_summary(self):
        """Print test summary"""
        print("\n" + "="*60)
        print("TEST SUMMARY")
        print("="*60)
        print(f"Tests passed: {self.tests_passed}")
        print(f"Tests failed: {self.tests_failed}")
        print(f"Total tests: {self.tests_passed + self.tests_failed}")
        print(f"Success rate: {self.tests_passed / max(1, self.tests_passed + self.tests_failed) * 100:.1f}%")
        print("="*60)
        
        if self.tests_failed > 0:
            print("\nFailed tests:")
            for name, passed, error in self.test_results:
                if not passed:
                    print(f"  ❌ {name}")
                    if error:
                        print(f"     {error}")
        
        print("\n" + "="*60)
        if self.tests_failed == 0:
            print("🎉 ALL TESTS PASSED!")
        else:
            print(f"⚠️  {self.tests_failed} test(s) failed")
        print("="*60)


async def main():
    """Main test runner"""
    test_suite = TestIterationN13()
    await test_suite.run_all_tests()
    
    # Return exit code based on results
    sys.exit(0 if test_suite.tests_failed == 0 else 1)


if __name__ == "__main__":
    asyncio.run(main())
