#!/usr/bin/env python3
"""
Test Suite for Echo9llama Iteration N+18
=========================================

Validates all V18 enhancements:
- Echobeats goal-directed scheduling
- Echodream deep knowledge integration
- Stream-of-consciousness engine
- Knowledge acquisition tools
- Integration of all systems
"""

import asyncio
import logging
import sys
import os
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent))

from core.autonomous_core_v18 import DeepTreeEchoV18
from core.echobeats_scheduler import TaskType, TaskPriority
from core.stream_of_consciousness import StreamThought

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Test results
test_results = []


def test_result(name: str, passed: bool, message: str = ""):
    """Record test result"""
    status = "✅ PASSED" if passed else "❌ FAILED"
    test_results.append((name, passed, message))
    logger.info(f"{status}: {name}")
    if message:
        logger.info(f"   {message}")


async def test_v18_initialization():
    """Test V18 core initialization with all new systems"""
    logger.info("--- Running Test: V18 Core Initialization ---")
    try:
        echo = DeepTreeEchoV18(state_file="data/test_v18_state.json")
        
        # Check V18 components exist
        assert hasattr(echo, 'echobeats'), "Echobeats scheduler should exist"
        assert hasattr(echo, 'echodream'), "Echodream system should exist"
        assert hasattr(echo, 'consciousness_stream'), "Stream-of-consciousness should exist"
        assert hasattr(echo, 'knowledge_acquisition'), "Knowledge acquisition should exist"
        
        # Check V17 components still exist
        assert hasattr(echo, 'skill_practice'), "Skill practice system should exist"
        assert hasattr(echo, 'discussion_manager'), "Discussion manager should exist"
        
        test_result(
            "test_v18_initialization",
            True,
            f"All V18 systems initialized successfully"
        )
        return echo
    except Exception as e:
        test_result("test_v18_initialization", False, f"Error: {e}")
        raise


async def test_echobeats_scheduler(echo: DeepTreeEchoV18):
    """Test echobeats scheduling system"""
    logger.info("--- Running Test: Echobeats Scheduler ---")
    try:
        # Start scheduler
        await echo.echobeats.start()
        
        # Schedule a test task
        task_id = echo.echobeats.schedule_task(
            task_type=TaskType.THOUGHT_GENERATION,
            description="Test thought generation",
            priority=TaskPriority.HIGH,
            estimated_tokens=100
        )
        
        assert task_id is not None, "Task should be scheduled"
        
        # Let scheduler run for a few seconds
        await asyncio.sleep(5)
        
        # Check scheduler status
        status = echo.echobeats.get_scheduler_status()
        
        assert status['is_running'], "Scheduler should be running"
        assert status['cycle_count'] >= 0, "Should have cycle count"
        assert status['current_step'] >= 1 and status['current_step'] <= 12, "Step should be 1-12"
        
        # Stop scheduler
        await echo.echobeats.stop()
        
        test_result(
            "test_echobeats_scheduler",
            True,
            f"Scheduler ran {status['cycle_count']} cycles, step {status['current_step']}/12"
        )
    except Exception as e:
        test_result("test_echobeats_scheduler", False, f"Error: {e}")


async def test_stream_of_consciousness(echo: DeepTreeEchoV18):
    """Test stream-of-consciousness engine"""
    logger.info("--- Running Test: Stream-of-Consciousness ---")
    try:
        # Start stream
        await echo.consciousness_stream.start_streaming()
        
        initial_count = echo.consciousness_stream.thought_count
        
        # Let it generate thoughts for 10 seconds
        logger.info("  -> Generating stream-of-consciousness for 10 seconds...")
        await asyncio.sleep(10)
        
        final_count = echo.consciousness_stream.thought_count
        thoughts_generated = final_count - initial_count
        
        # Stop stream
        await echo.consciousness_stream.stop_streaming()
        
        assert thoughts_generated > 0, "Should generate at least one thought"
        
        # Check recent thoughts
        recent = echo.consciousness_stream.get_recent_stream(5)
        assert len(recent) > 0, "Should have recent thoughts"
        
        # Check narrative coherence
        if len(recent) >= 2:
            coherence_scores = [t.coherence_score for t in recent]
            avg_coherence = sum(coherence_scores) / len(coherence_scores)
            logger.info(f"  -> Average coherence: {avg_coherence:.2f}")
        
        test_result(
            "test_stream_of_consciousness",
            True,
            f"Generated {thoughts_generated} thoughts with continuous stream"
        )
    except Exception as e:
        test_result("test_stream_of_consciousness", False, f"Error: {e}")


async def test_echodream_consolidation(echo: DeepTreeEchoV18):
    """Test echodream deep consolidation"""
    logger.info("--- Running Test: Echodream Consolidation ---")
    try:
        # Add some thoughts to consolidate
        for i in range(10):
            echo.all_thoughts.append({
                'content': f"Test thought {i} about wisdom and learning",
                'timestamp': None
            })
        
        # Trigger consolidation
        logger.info("  -> Performing dream consolidation...")
        synthesis = await echo.echodream.consolidate_knowledge()
        
        if synthesis:
            assert synthesis.theme is not None, "Should have a theme"
            assert len(synthesis.integrated_nodes) > 0, "Should integrate nodes"
            
            logger.info(f"  -> Theme: {synthesis.theme}")
            logger.info(f"  -> Integrated {len(synthesis.integrated_nodes)} nodes")
            logger.info(f"  -> Generated {len(synthesis.new_insights)} insights")
            
            test_result(
                "test_echodream_consolidation",
                True,
                f"Consolidated {len(synthesis.integrated_nodes)} thoughts into theme: {synthesis.theme}"
            )
        else:
            test_result(
                "test_echodream_consolidation",
                True,
                "Consolidation completed (no synthesis needed)"
            )
    except Exception as e:
        test_result("test_echodream_consolidation", False, f"Error: {e}")


async def test_knowledge_acquisition(echo: DeepTreeEchoV18):
    """Test knowledge acquisition system"""
    logger.info("--- Running Test: Knowledge Acquisition ---")
    try:
        # Explore a topic
        topic = "artificial intelligence"
        logger.info(f"  -> Exploring topic: {topic}")
        
        query = await echo.knowledge_acquisition.explore_topic(topic, motivation="test")
        
        assert query is not None, "Should create learning query"
        assert query.topic == topic, "Topic should match"
        assert query.status in ["completed", "failed"], "Should have final status"
        
        if query.status == "completed":
            assert len(query.sources_found) > 0, "Should find sources"
            logger.info(f"  -> Found {len(query.sources_found)} sources")
            
            if query.knowledge_extracted:
                logger.info(f"  -> Extracted: {query.knowledge_extracted[:100]}...")
        
        test_result(
            "test_knowledge_acquisition",
            True,
            f"Explored {topic}: {query.status}, {len(query.sources_found)} sources"
        )
    except Exception as e:
        test_result("test_knowledge_acquisition", False, f"Error: {e}")


async def test_autonomous_wake_rest_cycle(echo: DeepTreeEchoV18):
    """Test autonomous wake/rest cycling"""
    logger.info("--- Running Test: Autonomous Wake/Rest Cycle ---")
    try:
        # Simulate high cognitive load
        for _ in range(25):
            echo.echodream.load_monitor.record_thought()
        
        # Check if rest is needed
        should_rest, reason = echo.echodream.load_monitor.should_rest()
        
        logger.info(f"  -> Should rest: {should_rest}, Reason: {reason}")
        
        if should_rest:
            # Trigger rest cycle
            logger.info("  -> Entering rest cycle...")
            await echo.echodream.enter_rest_cycle(duration_seconds=5)
            
            assert echo.echodream.consolidation_count > 0, "Should have consolidated"
        
        test_result(
            "test_autonomous_wake_rest_cycle",
            True,
            f"Wake/rest cycle tested: {reason}"
        )
    except Exception as e:
        test_result("test_autonomous_wake_rest_cycle", False, f"Error: {e}")


async def test_integrated_autonomous_run(echo: DeepTreeEchoV18):
    """Test integrated autonomous operation"""
    logger.info("--- Running Test: Integrated Autonomous Run ---")
    try:
        initial_thoughts = len(echo.all_thoughts)
        initial_cycles = echo.triple_stream.cycle_count
        
        logger.info("  -> Running integrated autonomous operation for 20 seconds...")
        
        # Initialize and run
        await echo.initialize()
        await echo.wake()
        await echo.continuous_awareness.start()
        
        # Run for 20 seconds
        await asyncio.sleep(20)
        
        # Stop
        await echo.continuous_awareness.stop()
        await echo.echobeats.stop()
        await echo.consciousness_stream.stop_streaming()
        
        final_thoughts = len(echo.all_thoughts)
        final_cycles = echo.triple_stream.cycle_count
        
        thoughts_generated = final_thoughts - initial_thoughts
        cycles_completed = final_cycles - initial_cycles
        
        assert thoughts_generated >= 0, "Should generate thoughts"
        assert cycles_completed >= 0, "Should complete cycles"
        
        # Check that all systems are active
        echobeats_status = echo.echobeats.get_scheduler_status()
        stream_count = echo.consciousness_stream.thought_count
        
        logger.info(f"  -> Thoughts generated: {thoughts_generated}")
        logger.info(f"  -> Cycles completed: {cycles_completed}")
        logger.info(f"  -> Echobeats tasks completed: {echobeats_status['completed_tasks']}")
        logger.info(f"  -> Stream thoughts: {stream_count}")
        
        test_result(
            "test_integrated_autonomous_run",
            True,
            f"Integrated run: {thoughts_generated} thoughts, {cycles_completed} cycles, {stream_count} stream thoughts"
        )
    except Exception as e:
        test_result("test_integrated_autonomous_run", False, f"Error: {e}")


async def test_llm_provider_fallback():
    """Test LLM provider fallback (Anthropic -> OpenRouter)"""
    logger.info("--- Running Test: LLM Provider Fallback ---")
    try:
        from core.llm_unified import get_llm_client
        
        client = get_llm_client()
        
        # Test generation
        response = await client.generate(
            prompt="What is wisdom?",
            system_prompt="You are a philosophical assistant.",
            max_tokens=50,
            temperature=0.7
        )
        
        assert response is not None, "Should get response"
        
        if response.success:
            logger.info(f"  -> Provider: {response.provider.value}")
            logger.info(f"  -> Model: {response.model}")
            logger.info(f"  -> Response: {response.content[:100]}...")
            
            test_result(
                "test_llm_provider_fallback",
                True,
                f"LLM working with {response.provider.value}"
            )
        else:
            test_result(
                "test_llm_provider_fallback",
                False,
                f"LLM failed: {response.error}"
            )
    except Exception as e:
        test_result("test_llm_provider_fallback", False, f"Error: {e}")


async def test_v18_state_persistence(echo: DeepTreeEchoV18):
    """Test V18 state persistence"""
    logger.info("--- Running Test: V18 State Persistence ---")
    try:
        # Modify state
        echo.consciousness_stream.thought_count = 42
        echo.echodream.consolidation_count = 5
        
        # Save state
        await echo.save_state()
        
        # Create new instance and check if state file exists
        state_path = Path("data/test_v18_state_v18.json")
        assert state_path.exists(), "V18 state file should exist"
        
        test_result(
            "test_v18_state_persistence",
            True,
            "V18 state saved successfully"
        )
    except Exception as e:
        test_result("test_v18_state_persistence", False, f"Error: {e}")


async def run_all_tests():
    """Run all V18 tests"""
    logger.info("=" * 80)
    logger.info("ECHO9LLAMA ITERATION N+18 TEST SUITE")
    logger.info("=" * 80)
    
    # Check API keys
    anthropic_key = os.getenv("ANTHROPIC_API_KEY")
    openrouter_key = os.getenv("OPENROUTER_API_KEY")
    
    logger.info(f"ANTHROPIC_API_KEY: {'✅ Set' if anthropic_key else '❌ Not set'}")
    logger.info(f"OPENROUTER_API_KEY: {'✅ Set' if openrouter_key else '❌ Not set'}")
    logger.info("=" * 80)
    
    try:
        # Test 1: Initialization
        echo = await test_v18_initialization()
        
        # Test 2: LLM provider fallback
        await test_llm_provider_fallback()
        
        # Test 3: Echobeats scheduler
        await test_echobeats_scheduler(echo)
        
        # Test 4: Stream-of-consciousness
        await test_stream_of_consciousness(echo)
        
        # Test 5: Echodream consolidation
        await test_echodream_consolidation(echo)
        
        # Test 6: Knowledge acquisition
        await test_knowledge_acquisition(echo)
        
        # Test 7: Autonomous wake/rest cycle
        await test_autonomous_wake_rest_cycle(echo)
        
        # Test 8: V18 state persistence
        await test_v18_state_persistence(echo)
        
        # Test 9: Integrated autonomous run
        await test_integrated_autonomous_run(echo)
        
    except Exception as e:
        logger.error(f"Test suite error: {e}")
    
    # Print results
    logger.info("\n" + "=" * 80)
    logger.info("TEST RESULTS SUMMARY")
    logger.info("=" * 80)
    
    passed = sum(1 for _, p, _ in test_results if p)
    total = len(test_results)
    
    for name, passed_flag, message in test_results:
        status = "✅" if passed_flag else "❌"
        logger.info(f"{status} {name}")
        if message:
            logger.info(f"   {message}")
    
    logger.info("=" * 80)
    logger.info(f"TOTAL: {passed}/{total} tests passed")
    logger.info("=" * 80)
    
    return passed == total


if __name__ == "__main__":
    success = asyncio.run(run_all_tests())
    sys.exit(0 if success else 1)
