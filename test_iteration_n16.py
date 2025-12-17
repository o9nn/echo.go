#!/usr/bin/env python3
"""
Test Suite for Iteration N+16
Tests the enhanced autonomous core V16 with:
- Fixed autonomous cognitive cycling
- Real LLM integration
- Enhanced echodream consolidation
- Skill practice system
- Context-aware memory retrieval
"""

import asyncio
import sys
import os
from pathlib import Path

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent))

from core.autonomous_core_v16 import DeepTreeEchoV16, logger
from core.autonomous_core_v14 import CognitiveState

# Test configuration
TEST_STATE_FILE = "data/test_v16_state.json"

async def test_v16_initialization():
    """Test V16 core initialization"""
    print("--- Running Test: V16 Core Initialization ---")
    
    # Clean state
    if Path(TEST_STATE_FILE).exists():
        Path(TEST_STATE_FILE).unlink()
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Verify V16 components exist
    assert hasattr(echo, 'skill_practice'), "Skill practice system should exist"
    assert hasattr(echo, 'cognitive_operations'), "Cognitive operations should exist"
    assert hasattr(echo, 'echodream_consolidation'), "Echodream consolidation should exist"
    assert hasattr(echo, 'memory_retrieval'), "Memory retrieval should exist"
    
    # Verify LLM setup
    assert echo.llm_client is not None or echo.llm_provider is None, "LLM should be configured or None"
    
    # Verify V15 components still exist
    assert hasattr(echo, 'interest_patterns'), "Interest patterns should exist"
    assert hasattr(echo, 'goal_formation'), "Goal formation should exist"
    assert hasattr(echo, 'discussion_manager'), "Discussion manager should exist"
    
    # Verify skills initialized
    assert len(echo.skill_practice.skills) > 0, "Should have seed skills"
    
    print("  -> V16 core and all components initialized successfully.")
    print(f"  -> LLM provider: {echo.llm_provider or 'None'}")
    print(f"  -> Skills: {list(echo.skill_practice.skills.keys())}")
    print("✅ PASSED: V16 Core Initialization\n")

async def test_v16_state_persistence():
    """Test V16 state persistence"""
    print("--- Running Test: V16 State Persistence ---")
    
    # Create instance 1
    echo1 = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Modify state
    await echo1.skill_practice.practice_skill("pattern_recognition")
    echo1.interest_patterns.update_interest("quantum_computing", 0.3, "exposure")
    
    # Save state
    await echo1.save_state()
    print("  -> Instance 1 state saved.")
    
    # Create instance 2 and load
    echo2 = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Verify state loaded
    assert "pattern_recognition" in echo2.skill_practice.skills, "Skill should be loaded"
    assert echo2.skill_practice.skills["pattern_recognition"].practice_count > 0, "Practice count should persist"
    assert "quantum_computing" in echo2.interest_patterns.interests, "Interest should be loaded"
    
    print("  -> Instance 2 loaded state successfully.")
    print(f"  -> Pattern recognition competency: {echo2.skill_practice.skills['pattern_recognition'].competency:.2f}")
    print(f"  -> Quantum computing affinity: {echo2.interest_patterns.interests['quantum_computing'].affinity:.2f}")
    print("✅ PASSED: V16 State Persistence\n")

async def test_autonomous_cognitive_cycling():
    """Test that autonomous cognitive cycling actually works"""
    print("--- Running Test: Autonomous Cognitive Cycling ---")
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Initialize and wake
    await echo.initialize()
    await echo.wake()
    
    initial_cycles = echo.metrics.get("cycles_completed", 0)
    print(f"  -> Initial cycles: {initial_cycles}")
    
    # Start autonomous operation for 5 seconds
    print("  -> Starting 5-second autonomous run...")
    await echo.continuous_awareness.start()
    await asyncio.sleep(5)
    await echo.continuous_awareness.stop()
    
    final_cycles = echo.metrics.get("cycles_completed", 0)
    print(f"  -> Final cycles: {final_cycles}")
    
    # Save state
    await echo.save_state()
    
    # Verify cycles processed
    cycles_processed = final_cycles - initial_cycles
    assert cycles_processed > 0, f"Cognitive cycles should have processed (got {cycles_processed})"
    
    print(f"  -> Processed {cycles_processed} cycles in 5 seconds")
    print(f"  -> Thoughts generated: {len(echo.thoughts)}")
    print("✅ PASSED: Autonomous Cognitive Cycling\n")

async def test_skill_practice():
    """Test skill practice system"""
    print("--- Running Test: Skill Practice System ---")
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    skill_name = "logical_reasoning"
    initial_competency = echo.skill_practice.skills[skill_name].competency
    
    print(f"  -> Initial competency for {skill_name}: {initial_competency:.2f}")
    
    # Practice the skill
    result = await echo.skill_practice.practice_skill(skill_name)
    
    assert result["success"], "Practice should succeed"
    assert result["competency"] > initial_competency, "Competency should improve"
    
    print(f"  -> After practice: {result['competency']:.2f}")
    print(f"  -> Practice count: {result['practice_count']}")
    print("✅ PASSED: Skill Practice System\n")

async def test_cognitive_operations():
    """Test enhanced cognitive operations"""
    print("--- Running Test: Enhanced Cognitive Operations ---")
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Add some knowledge for operations to work with
    echo.knowledge_base["consciousness"] = {
        "topic": "consciousness",
        "content": "Consciousness is the state of being aware of one's thoughts and surroundings.",
        "confidence": 0.8,
        "timestamp": "2025-12-17T00:00:00"
    }
    echo.knowledge_base["learning"] = {
        "topic": "learning",
        "content": "Learning is the process of acquiring new knowledge or skills.",
        "confidence": 0.8,
        "timestamp": "2025-12-17T00:00:00"
    }
    
    # Test pattern recognition
    pattern_result = await echo.cognitive_operations.pattern_recognition({})
    assert len(pattern_result) > 0, "Pattern recognition should return content"
    print(f"  -> Pattern recognition: {pattern_result[:100]}...")
    
    # Test creative synthesis
    synthesis_result = await echo.cognitive_operations.creative_synthesis({})
    assert len(synthesis_result) > 0, "Creative synthesis should return content"
    print(f"  -> Creative synthesis: {synthesis_result[:100]}...")
    
    # Test future simulation
    future_result = await echo.cognitive_operations.future_simulation({})
    assert len(future_result) > 0, "Future simulation should return content"
    print(f"  -> Future simulation: {future_result[:100]}...")
    
    print("✅ PASSED: Enhanced Cognitive Operations\n")

async def test_memory_retrieval():
    """Test context-aware memory retrieval"""
    print("--- Running Test: Context-Aware Memory Retrieval ---")
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Add diverse knowledge
    echo.knowledge_base["ai"] = {
        "topic": "artificial_intelligence",
        "content": "AI systems can learn and adapt.",
        "confidence": 0.9,
        "depth": 0.7,
        "timestamp": "2025-12-17T00:00:00"
    }
    echo.knowledge_base["wisdom"] = {
        "topic": "wisdom",
        "content": "Wisdom is the ability to make sound judgments.",
        "confidence": 0.8,
        "depth": 0.8,
        "timestamp": "2025-12-17T00:00:00"
    }
    
    # Set active goal
    from core.autonomous_core_v15 import LearningGoal
    goal = LearningGoal("test_goal", "wisdom", "Learn about wisdom", 0.9)
    echo.goal_formation.active_goals["test_goal"] = goal
    
    # Retrieve relevant knowledge
    relevant = echo.memory_retrieval.retrieve_relevant_knowledge({}, top_k=2)
    
    assert len(relevant) > 0, "Should retrieve some knowledge"
    print(f"  -> Retrieved {len(relevant)} relevant items")
    for item in relevant:
        print(f"     - {item['topic']}")
    
    print("✅ PASSED: Context-Aware Memory Retrieval\n")

async def test_echodream_consolidation():
    """Test enhanced echodream consolidation"""
    print("--- Running Test: Enhanced Echodream Consolidation ---")
    
    echo = DeepTreeEchoV16(state_file=TEST_STATE_FILE)
    
    # Create some experiences
    experiences = [
        {"content": "Thinking about consciousness and awareness"},
        {"content": "Learning about pattern recognition"},
        {"content": "Exploring the nature of wisdom"},
        {"content": "Practicing logical reasoning skills"}
    ]
    
    initial_wisdom = echo.wisdom_state.wisdom_score
    print(f"  -> Initial wisdom: {initial_wisdom:.2f}")
    
    # Perform consolidation
    result = await echo.echodream_consolidation.consolidate(experiences)
    
    assert "insights" in result, "Should return insights"
    assert result["quality"] > 0, "Should have quality score"
    
    print(f"  -> Consolidation quality: {result['quality']:.2f}")
    print(f"  -> Insights generated: {len(result['insights'])}")
    if result['insights']:
        print(f"  -> First insight: {result['insights'][0][:100]}...")
    
    final_wisdom = echo.wisdom_state.wisdom_score
    print(f"  -> Final wisdom: {final_wisdom:.2f}")
    
    print("✅ PASSED: Enhanced Echodream Consolidation\n")

async def run_all_tests():
    """Run all tests"""
    tests = [
        test_v16_initialization,
        test_v16_state_persistence,
        test_autonomous_cognitive_cycling,
        test_skill_practice,
        test_cognitive_operations,
        test_memory_retrieval,
        test_echodream_consolidation
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            await test()
            passed += 1
        except AssertionError as e:
            print(f"❌ FAILED: {test.__name__}")
            print(f"   Error: {e}\n")
            failed += 1
        except Exception as e:
            print(f"❌ ERROR in {test.__name__}")
            print(f"   Error: {e}\n")
            failed += 1
    
    print("--- Test Summary ---")
    print(f"Passed: {passed}, Failed: {failed}")
    
    return failed == 0

if __name__ == "__main__":
    success = asyncio.run(run_all_tests())
    sys.exit(0 if success else 1)
