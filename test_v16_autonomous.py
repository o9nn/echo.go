#!/usr/bin/env python3
"""
Test script for Deep Tree Echo V16 autonomous operation
"""
import asyncio
import logging
import sys
from pathlib import Path

# Add to path
sys.path.insert(0, str(Path(__file__).parent))

from core.autonomous_core_v16 import DeepTreeEchoV16

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

async def test_v16_short_run():
    """Test V16 with a short autonomous run"""
    print("\n" + "=" * 80)
    print("DEEP TREE ECHO V16 - TEST RUN")
    print("=" * 80)
    
    echo = DeepTreeEchoV16(state_file="data/test_v16_state.json")
    
    print("\n📊 Initial State:")
    print(f"   Interest topics: {len(echo.interest_patterns.interests)}")
    print(f"   Active goals: {len(echo.goal_formation.active_goals)}")
    print(f"   Wisdom insights: {len(echo.wisdom_system.wisdom_insights)}")
    
    print("\n🚀 Starting 30-second autonomous run...")
    print("   (Watch for 3-stream consciousness, wisdom extraction, and goal formation)\n")
    
    try:
        await echo.run_autonomous(duration_seconds=30)
    except KeyboardInterrupt:
        print("\n⚠️ Interrupted by user")
    
    print("\n" + "=" * 80)
    print("TEST COMPLETE")
    print("=" * 80)

if __name__ == "__main__":
    asyncio.run(test_v16_short_run())
