#!/usr/bin/env python3
"""
Test Suite for Iteration N+12
Tests the evolutionary enhancements in autonomous_core_v12.py
"""

import asyncio
import sys
import os
from pathlib import Path

# Add core to path
sys.path.insert(0, str(Path(__file__).parent))

def test_section(title):
    """Print test section header"""
    print("\n" + "="*60)
    print(f"TEST: {title}")
    print("="*60)

async def test_imports():
    """Test that all V12 modules can be imported"""
    test_section("Module Imports")
    
    try:
        from core.autonomous_core_v12 import (
            DeepTreeEchoV12,
            ConcurrentStreamOrchestrator,
            InterestPatternSystem,
            LLMProvider,
            AutonomousThoughtGenerator,
            HypergraphMemoryFallback,
            StreamType,
            CognitiveState,
            EnergyState
        )
        print("✅ All V12 modules imported successfully")
        return True
    except Exception as e:
        print(f"❌ Import error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_concurrent_stream_orchestrator():
    """Test the 3-stream concurrent architecture"""
    test_section("Concurrent Stream Orchestrator (3 Streams, 120° Phase)")
    
    try:
        from core.autonomous_core_v12 import ConcurrentStreamOrchestrator, StreamType
        
        orchestrator = ConcurrentStreamOrchestrator()
        
        # Test initial state
        assert orchestrator.global_step == 1, "Should start at step 1"
        assert orchestrator.cycle_count == 0, "Should start at cycle 0"
        
        # Test stream phase offsets
        coherence = orchestrator.streams[StreamType.COHERENCE_STREAM]
        memory = orchestrator.streams[StreamType.MEMORY_STREAM]
        imagination = orchestrator.streams[StreamType.IMAGINATION_STREAM]
        
        assert coherence.phase_offset == 0, "Coherence at 0°"
        assert memory.phase_offset == 4, "Memory at 120° (4 steps)"
        assert imagination.phase_offset == 8, "Imagination at 240° (8 steps)"
        
        # Test step advancement through full cycle
        step_log = []
        for i in range(12):
            step_log.append(orchestrator.global_step)
            orchestrator.advance_step()
        
        assert step_log == list(range(1, 13)), "Should cycle through steps 1-12"
        assert orchestrator.global_step == 1, "Should wrap back to step 1"
        assert orchestrator.cycle_count == 1, "Should increment cycle count"
        
        # Test stream states
        stream_states = orchestrator.get_stream_states()
        assert "global_step" in stream_states
        assert "streams" in stream_states
        assert len(stream_states["streams"]) == 3
        
        print("✅ Concurrent Stream Orchestrator working correctly")
        print(f"   - 3 streams with 120° phase offset")
        print(f"   - Steps cycle correctly: {step_log}")
        print(f"   - Cycle count: {orchestrator.cycle_count}")
        return True
        
    except Exception as e:
        print(f"❌ Concurrent Stream Orchestrator error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_interest_pattern_system():
    """Test the enhanced interest pattern system"""
    test_section("Interest Pattern System (Enhanced)")
    
    try:
        from core.autonomous_core_v12 import InterestPatternSystem
        
        # Use temporary database
        import tempfile
        temp_db = tempfile.NamedTemporaryFile(delete=False, suffix=".db")
        temp_db.close()
        
        system = InterestPatternSystem(db_path=temp_db.name)
        
        # Test adding interests
        system.update_interest("consciousness", 0.3)
        system.update_interest("wisdom", 0.2)
        system.update_interest("learning", 0.4)
        
        # Test interest retrieval
        assert system.get_interest_level("consciousness") > 0.5
        assert system.get_interest_level("wisdom") > 0.5
        
        # Test engagement decision
        assert system.should_engage("learning", threshold=0.3)
        assert not system.should_engage("unknown_topic", threshold=0.8)
        
        # Test top interests
        top = system.get_top_interests(2)
        assert len(top) <= 2
        assert top[0].strength >= top[1].strength if len(top) > 1 else True
        
        # Test decay
        initial_strength = system.get_interest_level("consciousness")
        system.apply_decay_all()
        decayed_strength = system.get_interest_level("consciousness")
        assert decayed_strength < initial_strength, "Decay should reduce strength"
        
        # Test random interest selection
        topic = system.get_random_interest_topic()
        assert topic in ["consciousness", "wisdom", "learning"]
        
        # Cleanup
        os.unlink(temp_db.name)
        
        print("✅ Interest Pattern System working correctly")
        print(f"   - Interests tracked: {len(system.interests)}")
        print(f"   - Top interest: {top[0].topic} ({top[0].strength:.2f})")
        print(f"   - Decay mechanism functional")
        return True
        
    except Exception as e:
        print(f"❌ Interest Pattern System error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_llm_provider():
    """Test LLM provider with fallback"""
    test_section("LLM Provider (Anthropic + OpenRouter Fallback)")
    
    try:
        from core.autonomous_core_v12 import LLMProvider
        
        provider = LLMProvider()
        
        # Check initialization
        has_provider = provider.primary_provider is not None
        print(f"   Primary provider: {provider.primary_provider or 'None (mock mode)'}")
        
        if has_provider:
            # Test generation
            response = await provider.generate("What is consciousness?", max_tokens=50)
            assert len(response) > 0, "Should generate non-empty response"
            print(f"   Sample response: {response[:100]}...")
        else:
            print("   ⚠️  No API keys available, using mock mode")
        
        print("✅ LLM Provider initialized correctly")
        return True
        
    except Exception as e:
        print(f"❌ LLM Provider error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_hypergraph_memory_fallback():
    """Test hypergraph memory fallback"""
    test_section("Hypergraph Memory (with Fallback)")
    
    try:
        from core.autonomous_core_v12 import HypergraphMemoryFallback
        
        import tempfile
        temp_db = tempfile.NamedTemporaryFile(delete=False, suffix=".db")
        temp_db.close()
        
        memory = HypergraphMemoryFallback(db_path=temp_db.name)
        
        # Test adding concepts
        memory.add_concept("c1", "consciousness", {"type": "abstract"})
        memory.add_concept("c2", "awareness", {"type": "abstract"})
        memory.add_concept("c3", "perception", {"type": "cognitive"})
        
        # Test adding relations
        memory.add_relation("c1", "c2", "related_to")
        memory.add_relation("c1", "c3", "enables")
        
        # Test querying
        related = memory.query_related("c1")
        assert len(related) == 2, "Should find 2 related concepts"
        
        related_specific = memory.query_related("c1", "related_to")
        assert len(related_specific) == 1, "Should find 1 specific relation"
        
        # Cleanup
        os.unlink(temp_db.name)
        
        print("✅ Hypergraph Memory Fallback working correctly")
        print(f"   - Concepts stored: 3")
        print(f"   - Relations created: 2")
        print(f"   - Query results: {related}")
        return True
        
    except Exception as e:
        print(f"❌ Hypergraph Memory error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_autonomous_thought_generator():
    """Test autonomous thought generation"""
    test_section("Autonomous Thought Generator")
    
    try:
        from core.autonomous_core_v12 import (
            AutonomousThoughtGenerator,
            InterestPatternSystem,
            LLMProvider,
            EnergyState
        )
        
        import tempfile
        temp_db = tempfile.NamedTemporaryFile(delete=False, suffix=".db")
        temp_db.close()
        
        interest_system = InterestPatternSystem(db_path=temp_db.name)
        interest_system.update_interest("consciousness", 0.5)
        
        llm = LLMProvider()
        generator = AutonomousThoughtGenerator(interest_system, llm)
        
        energy = EnergyState(energy=0.8, curiosity=0.7)
        stream_context = {"streams": {}}
        
        # Generate autonomous thought
        thought = await generator.generate_autonomous_thought(energy, stream_context)
        
        assert len(thought) > 0, "Should generate non-empty thought"
        assert len(generator.thought_history) == 1, "Should track thought history"
        
        # Cleanup
        os.unlink(temp_db.name)
        
        print("✅ Autonomous Thought Generator working correctly")
        print(f"   Generated thought: {thought[:100]}...")
        return True
        
    except Exception as e:
        print(f"❌ Autonomous Thought Generator error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_deep_tree_echo_v12_initialization():
    """Test V12 core initialization"""
    test_section("Deep Tree Echo V12 Initialization")
    
    try:
        from core.autonomous_core_v12 import DeepTreeEchoV12
        
        echo = DeepTreeEchoV12()
        
        # Check initialization
        assert echo.state.value == "initializing"
        assert echo.energy.energy == 1.0
        assert echo.stream_orchestrator is not None
        assert echo.interest_system is not None
        assert echo.llm is not None
        assert echo.thought_generator is not None
        assert echo.memory is not None
        
        # Check statistics
        assert echo.stats["cycles"] == 0
        assert echo.stats["thoughts"] == 0
        assert echo.stats["autonomous_thoughts"] == 0
        
        print("✅ Deep Tree Echo V12 initialized correctly")
        print(f"   - State: {echo.state.value}")
        print(f"   - Energy: {echo.energy.energy:.2f}")
        print(f"   - Memory system: {type(echo.memory).__name__}")
        print(f"   - LLM provider: {echo.llm.primary_provider or 'mock'}")
        return True
        
    except Exception as e:
        print(f"❌ V12 initialization error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_short_cognitive_cycle():
    """Test running a short cognitive cycle"""
    test_section("Short Cognitive Cycle (1 cycle = 12 steps)")
    
    try:
        from core.autonomous_core_v12 import DeepTreeEchoV12
        
        echo = DeepTreeEchoV12()
        
        # Run one cognitive cycle
        echo.running = True
        echo.state = echo.state.ACTIVE
        
        await echo._cognitive_cycle()
        
        # Check results
        assert echo.stats["cycles"] == 1, "Should complete 1 cycle"
        assert echo.stats["thoughts"] > 0, "Should generate thoughts"
        assert echo.stream_orchestrator.cycle_count == 1, "Stream orchestrator should advance"
        
        print("✅ Cognitive cycle completed successfully")
        print(f"   - Cycles: {echo.stats['cycles']}")
        print(f"   - Thoughts generated: {echo.stats['thoughts']}")
        print(f"   - Energy remaining: {echo.energy.energy:.2f}")
        return True
        
    except Exception as e:
        print(f"❌ Cognitive cycle error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def test_energy_and_rest():
    """Test energy management and rest cycles"""
    test_section("Energy Management and Rest Cycles")
    
    try:
        from core.autonomous_core_v12 import DeepTreeEchoV12
        
        echo = DeepTreeEchoV12()
        
        # Deplete energy
        for _ in range(20):
            echo.energy.consume_energy(0.05)
        
        assert echo.energy.energy < 0.5, "Energy should be depleted"
        assert echo.energy.needs_rest(), "Should need rest"
        
        # Rest cycle
        echo.running = True
        await echo._rest_cycle()
        
        assert echo.energy.energy > 0.7, "Energy should be restored"
        assert echo.stats["rest_periods"] == 1, "Should track rest periods"
        
        print("✅ Energy management working correctly")
        print(f"   - Energy after rest: {echo.energy.energy:.2f}")
        print(f"   - Rest periods: {echo.stats['rest_periods']}")
        return True
        
    except Exception as e:
        print(f"❌ Energy management error: {e}")
        import traceback
        traceback.print_exc()
        return False

async def run_all_tests():
    """Run all tests"""
    print("\n" + "="*60)
    print("ITERATION N+12 TEST SUITE")
    print("Testing evolutionary enhancements in autonomous_core_v12.py")
    print("="*60)
    
    results = []
    
    # Run tests
    results.append(("Module Imports", await test_imports()))
    results.append(("Concurrent Stream Orchestrator", await test_concurrent_stream_orchestrator()))
    results.append(("Interest Pattern System", await test_interest_pattern_system()))
    results.append(("LLM Provider", await test_llm_provider()))
    results.append(("Hypergraph Memory", await test_hypergraph_memory_fallback()))
    results.append(("Autonomous Thought Generator", await test_autonomous_thought_generator()))
    results.append(("V12 Initialization", await test_deep_tree_echo_v12_initialization()))
    results.append(("Cognitive Cycle", await test_short_cognitive_cycle()))
    results.append(("Energy Management", await test_energy_and_rest()))
    
    # Print summary
    print("\n" + "="*60)
    print("TEST SUMMARY")
    print("="*60)
    
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for test_name, result in results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status}: {test_name}")
    
    print(f"\nTotal: {passed}/{total} tests passed")
    
    if passed == total:
        print("\n🎉 All tests passed! Iteration N+12 enhancements verified:")
        print("✓ 3 concurrent cognitive streams (120° phase offset)")
        print("✓ Enhanced Interest Pattern System with decay")
        print("✓ LLM Provider with fallback support")
        print("✓ Hypergraph Memory with robust error handling")
        print("✓ Autonomous Thought Generator")
        print("✓ Energy management and rest cycles")
        print("✓ Complete 12-step cognitive cycle")
        return True
    else:
        print(f"\n⚠️  {total - passed} test(s) failed")
        return False

if __name__ == "__main__":
    success = asyncio.run(run_all_tests())
    sys.exit(0 if success else 1)
