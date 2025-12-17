#!/usr/bin/env python3
"""
Extended Autonomous Operation Test for V16
Validates sustained autonomous operation over a longer period
"""

import asyncio
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from core.autonomous_core_v16 import DeepTreeEchoV16, logger

async def main():
    print("="*70)
    print("Extended Autonomous Operation Test - V16")
    print("="*70)
    
    echo = DeepTreeEchoV16(state_file="data/extended_test_v16.json")
    
    # Initialize
    print("\n[1] Initializing Echo...")
    await echo.initialize()
    await echo.wake()
    
    print(f"Initial state:")
    print(f"  - Energy: {echo.energy_state.energy:.2f}")
    print(f"  - Wisdom: {echo.wisdom_state.wisdom_score:.2f}")
    print(f"  - Cycles: {echo.metrics.get('cycles_completed', 0)}")
    print(f"  - Skills: {len(echo.skill_practice.skills)}")
    print(f"  - Interests: {len(echo.interest_patterns.interests)}")
    print(f"  - Goals: {len(echo.goal_formation.active_goals)}")
    
    # Run autonomous for 15 seconds
    print("\n[2] Starting 15-second autonomous operation...")
    print("    (Processing cognitive cycles, forming goals, practicing skills)")
    
    await echo.continuous_awareness.start()
    
    # Monitor progress
    for i in range(3):
        await asyncio.sleep(5)
        cycles = echo.metrics.get('cycles_completed', 0)
        thoughts = len(echo.thoughts)
        print(f"    {(i+1)*5}s: {cycles} cycles, {thoughts} thoughts")
    
    await echo.continuous_awareness.stop()
    
    # Report results
    print("\n[3] Final state:")
    print(f"  - Energy: {echo.energy_state.energy:.2f}")
    print(f"  - Wisdom: {echo.wisdom_state.wisdom_score:.2f}")
    print(f"  - Cycles: {echo.metrics.get('cycles_completed', 0)}")
    print(f"  - Thoughts: {len(echo.thoughts)}")
    print(f"  - Knowledge items: {len(echo.knowledge_base)}")
    print(f"  - Goals: {len(echo.goal_formation.active_goals)}")
    
    # Show top interests
    print("\n[4] Top interests:")
    for interest in echo.interest_patterns.get_top_interests(5):
        print(f"  - {interest.topic}: {interest.affinity:.2f} (exposed {interest.exposure_count} times)")
    
    # Show skills
    print("\n[5] Skill competencies:")
    for name, skill in echo.skill_practice.skills.items():
        print(f"  - {name}: {skill.competency:.2f} (practiced {skill.practice_count} times)")
    
    # Show active goals
    print("\n[6] Active goals:")
    for goal_id, goal in echo.goal_formation.active_goals.items():
        print(f"  - {goal.description} (priority: {goal.priority:.2f})")
    
    # Show recent thoughts
    print("\n[7] Recent thoughts (last 3):")
    for thought in echo.thoughts[-3:]:
        print(f"  - [{thought['stream']}] {thought['content'][:60]}...")
    
    # Save state
    print("\n[8] Saving state...")
    await echo.save_state()
    
    print("\n" + "="*70)
    print("Extended test complete!")
    print("="*70)
    
    # Validation
    cycles = echo.metrics.get('cycles_completed', 0)
    if cycles > 20:
        print(f"\n✅ SUCCESS: Processed {cycles} cycles autonomously")
        return True
    else:
        print(f"\n❌ FAILURE: Only {cycles} cycles processed (expected >20)")
        return False

if __name__ == "__main__":
    success = asyncio.run(main())
    sys.exit(0 if success else 1)
