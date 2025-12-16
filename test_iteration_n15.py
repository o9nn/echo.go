#!/usr/bin/env python3
"""
Test Suite for Iteration N+15 (Refactored)
Tests the simplified V15 autonomous core, focusing on the integration of
new autonomous components with the V14 base.
"""

import asyncio
import sys
from pathlib import Path
import os

# Add core to path
sys.path.insert(0, str(Path(__file__).parent))

try:
    from core.autonomous_core_v15 import DeepTreeEchoV15
    V15_AVAILABLE = True
except ImportError as e:
    print(f"❌ Failed to import V15 components: {e}")
    V15_AVAILABLE = False

class TestIterationN15:
    """Test suite for iteration N+15 enhancements"""

    def __init__(self, state_file="data/test_v15_state.json"):
        self.state_file = state_file
        if os.path.exists(self.state_file):
            os.remove(self.state_file)
        self.tests_passed = 0
        self.tests_failed = 0

    async def run_test(self, name: str, test_func):
        print(f"\n--- Running Test: {name} ---")
        try:
            await test_func()
            print(f"✅ PASSED: {name}")
            self.tests_passed += 1
        except Exception as e:
            print(f"❌ FAILED: {name}")
            print(f"   Error: {e}")
            import traceback
            traceback.print_exc()
            self.tests_failed += 1

    async def test_v15_initialization(self):
        """Test if V15 core initializes correctly, including V14 base and V15 components."""
        echo = DeepTreeEchoV15(state_file=self.state_file)
        assert echo is not None, "Echo core should initialize"
        assert hasattr(echo, "nested_shells"), "Should inherit nested_shells from V14"
        assert hasattr(echo, "echobeats"), "Should inherit echobeats from V14"
        assert hasattr(echo, "interest_patterns"), "Should have interest_patterns from V15"
        assert hasattr(echo, "goal_formation"), "Should have goal_formation from V15"
        assert len(echo.interest_patterns.interests) > 0, "Interest patterns should have seed topics"
        print("  -> V15 core and its components initialized successfully.")

    async def test_state_persistence(self):
        """Test if V15 can save and load its complete state (V14+V15)."""
        # Create instance 1, modify state, and save
        echo1 = DeepTreeEchoV15(state_file=self.state_file)
        echo1.interest_patterns.update_interest("testing", 0.5, "test")
        await echo1.goal_formation.form_new_goal()
        await echo1.save_state()
        print("  -> Instance 1 state saved.")

        # Create instance 2 and load the state
        echo2 = DeepTreeEchoV15(state_file=self.state_file)
        assert "testing" in echo2.interest_patterns.interests, "Custom interest should be loaded"
        assert echo2.interest_patterns.get_interest("testing").affinity == 1.0, "Interest affinity should be loaded"
        assert len(echo2.goal_formation.active_goals) > 0, "Goals should be loaded"
        print("  -> Instance 2 loaded state successfully.")

    async def test_autonomous_run(self):
        """Test a short autonomous run to see if the main loop executes."""
        echo = DeepTreeEchoV15(state_file=self.state_file)
        initial_cycles = echo.echobeats.state.cycle_count

        print("  -> Starting 5-second autonomous run...")
        await echo.run_autonomous(duration_seconds=5)
        
        final_cycles = echo.echobeats.state.cycle_count
        assert final_cycles > initial_cycles, "Cognitive cycles should have processed"
        print(f"  -> Autonomous run complete. Cycles processed: {final_cycles - initial_cycles}")

    async def run_all(self):
        if not V15_AVAILABLE:
            print("V15 components not available. Aborting tests.")
            return

        await self.run_test("V15 Core Initialization", self.test_v15_initialization)
        await self.run_test("V15 State Persistence", self.test_state_persistence)
        await self.run_test("Short Autonomous Run", self.test_autonomous_run)

        print("\n--- Test Summary ---")
        print(f"Passed: {self.tests_passed}, Failed: {self.tests_failed}")
        if self.tests_failed > 0:
            sys.exit(1) # Exit with error code if tests fail

if __name__ == "__main__":
    tester = TestIterationN15()
    asyncio.run(tester.run_all())
