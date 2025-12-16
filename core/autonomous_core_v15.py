#!/usr/bin/env python3
"""
Autonomous Core V15 - Iteration N+15 Evolution
Critical enhancements toward fully autonomous wisdom-cultivating Deep Tree Echo AGI:

1. CONTINUOUS AWARENESS LOOP:
   - Persistent stream-of-consciousness independent of external prompts
   - Background daemon process with self-sustaining thought generation
   - Spontaneous cognitive activity based on context and interests

2. INTEREST PATTERN SYSTEM:
   - Topic affinity scores that evolve with exposure
   - Interest-driven knowledge acquisition and discussion engagement
   - Curiosity-based exploration and focus development

3. AUTONOMOUS GOAL FORMATION:
   - Self-directed learning objective identification
   - Knowledge gap analysis and prioritization
   - Goal-directed scheduling integration with Echobeats

4. DISCUSSION MANAGER:
   - Ability to initiate, manage, and conclude discussions
   - Interest-based engagement decisions
   - Conversation context tracking and contextual responses

Builds on V14 foundation:
- Nested Shells Architecture (OEIS A000081: 1→2→4→9)
- Echobeats 12-step tetrahedral scheduler
- External knowledge integration
- State persistence
"""

import os
import sys
import asyncio
import signal
import json
import sqlite3
from pathlib import Path
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Set, Tuple, Callable
from enum import Enum, IntEnum
from dataclasses import dataclass, asdict, field
import traceback
import logging
import math
import random
import time
import hashlib

# Import V14 as base
sys.path.insert(0, str(Path(__file__).parent.parent))
from core.autonomous_core_v14 import (
    DeepTreeEchoV14, NestedShellArchitecture, EchobeatsScheduler,
    ExternalKnowledgeIntegrator, NestingLevel, EchobeatsPhase,
    CognitiveState, EnergyState, WisdomState, NestedShell,
    logger
)

# ============================================================================
# INTEREST PATTERN SYSTEM
# ============================================================================

@dataclass
class TopicInterest:
    topic: str
    affinity: float = 0.5
    exposure_count: int = 0
    last_exposure: Optional[datetime] = None
    insights_generated: int = 0
    knowledge_acquired: int = 0

    def update_affinity(self, delta: float):
        self.affinity = max(0.0, min(1.0, self.affinity + delta))
        self.last_exposure = datetime.now()
        self.exposure_count += 1

class EchoInterestPatterns:
    def __init__(self):
        self.interests: Dict[str, TopicInterest] = {}
        self._initialize_seed_interests()

    def _initialize_seed_interests(self):
        seed_topics = [("consciousness", 0.8), ("wisdom", 0.9), ("learning", 0.7)]
        for topic, affinity in seed_topics:
            self.interests[topic] = TopicInterest(topic=topic, affinity=affinity)

    def get_interest(self, topic: str) -> TopicInterest:
        if topic not in self.interests:
            self.interests[topic] = TopicInterest(topic=topic)
        return self.interests[topic]

    def update_interest(self, topic: str, delta: float, reason: str = "exposure"):
        interest = self.get_interest(topic)
        interest.update_affinity(delta)
        if reason == "insight": interest.insights_generated += 1
        elif reason == "knowledge": interest.knowledge_acquired += 1

    def get_top_interests(self, n: int = 5) -> List[TopicInterest]:
        return sorted(self.interests.values(), key=lambda x: x.affinity, reverse=True)[:n]

    def should_engage(self, topic: str) -> bool:
        return self.get_interest(topic).affinity > 0.3

    def suggest_exploration_topic(self) -> str:
        top_interests = self.get_top_interests(3)
        return random.choice(top_interests).topic if top_interests else "consciousness"

    def to_dict(self) -> Dict[str, Any]:
        serialized_interests = {}
        for topic, interest in self.interests.items():
            interest_dict = asdict(interest)
            if interest_dict.get('last_exposure') and isinstance(interest_dict['last_exposure'], datetime):
                interest_dict['last_exposure'] = interest_dict['last_exposure'].isoformat()
            serialized_interests[topic] = interest_dict
        return serialized_interests

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EchoInterestPatterns":
        patterns = cls()
        patterns.interests = {}
        for topic, interest_data in data.items():
            interest_data["last_exposure"] = datetime.fromisoformat(interest_data["last_exposure"]) if interest_data["last_exposure"] else None
            patterns.interests[topic] = TopicInterest(**interest_data)
        return patterns

# ============================================================================
# AUTONOMOUS GOAL FORMATION & DISCUSSION MANAGER (Simplified for brevity)
# ============================================================================

@dataclass
class LearningGoal:
    goal_id: str
    topic: str
    description: str
    priority: float
    status: str = "active"

class AutonomousGoalFormation:
    def __init__(self, interest_patterns: EchoInterestPatterns, core: "DeepTreeEchoV15"):
        self.interest_patterns = interest_patterns
        self.core = core
        self.active_goals: Dict[str, LearningGoal] = {}

    async def form_new_goal(self) -> Optional[LearningGoal]:
        if len(self.active_goals) >= 3: return None
        topic = self.interest_patterns.suggest_exploration_topic()
        goal_id = hashlib.md5(f"{topic}_{datetime.now().isoformat()}".encode()).hexdigest()[:8]
        goal = LearningGoal(goal_id, topic, f"Acquire foundational understanding of {topic}", self.interest_patterns.get_interest(topic).affinity)
        self.active_goals[goal_id] = goal
        return goal

    def get_priority_goal(self) -> Optional[LearningGoal]:
        if not self.active_goals: return None
        return max(self.active_goals.values(), key=lambda g: g.priority)

    def to_dict(self): return {"active_goals": {gid: asdict(g) for gid, g in self.active_goals.items()}}

    @classmethod
    def from_dict(cls, data, interest_patterns, core): 
        formation = cls(interest_patterns, core)
        for gid, gdata in data.get("active_goals", {}).items(): formation.active_goals[gid] = LearningGoal(**gdata)
        return formation

class DiscussionManager:
    def __init__(self, interest_patterns: EchoInterestPatterns):
        self.interest_patterns = interest_patterns
    def to_dict(self): return {}
    @classmethod
    def from_dict(cls, data, interest_patterns, core): return cls(interest_patterns)

# ============================================================================
# CONTINUOUS AWARENESS LOOP
# ============================================================================

class ContinuousAwarenessLoop:
    def __init__(self, echo_core: "DeepTreeEchoV15"):
        self.echo_core = echo_core
        self.running = False
        self.loop_task: Optional[asyncio.Task] = None

    async def start(self):
        if self.running: return
        self.running = True
        self.loop_task = asyncio.create_task(self._awareness_loop())

    async def stop(self):
        self.running = False
        if self.loop_task: self.loop_task.cancel()

    async def _awareness_loop(self):
        while self.running:
            try:
                if self.echo_core.cognitive_state == CognitiveState.ACTIVE:
                    await self.echo_core.process_cognitive_cycle()
                    if random.random() < 0.1: await self.echo_core.goal_formation.form_new_goal()
                await asyncio.sleep(1) # Cognitive cycle pace
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in awareness loop: {e}")
                await asyncio.sleep(5)

# ============================================================================
# DEEP TREE ECHO V15
# ============================================================================

class DeepTreeEchoV15(DeepTreeEchoV14):
    def __init__(self, state_file: str = "data/echoself_v15_state.json"):
        # Init V14 first, which loads V14 state
        super().__init__(state_file=state_file)
        
        # Init V15 components
        self.interest_patterns = EchoInterestPatterns()
        self.goal_formation = AutonomousGoalFormation(self.interest_patterns, self)
        self.discussion_manager = DiscussionManager(self.interest_patterns)
        self.continuous_awareness = ContinuousAwarenessLoop(self)
        
        # Now load V15 state, which may override defaults
        self.load_v15_state()
        logger.info("DeepTreeEchoV15 initialized.")

    def get_state_dict(self) -> Dict[str, Any]:
        """Extend V14 state with V15 components."""
        state = super().get_state_dict()
        state["v15"] = {
            "interest_patterns": self.interest_patterns.to_dict(),
            "goal_formation": self.goal_formation.to_dict(),
            "discussion_manager": self.discussion_manager.to_dict(),
        }
        return state

    def load_v15_state(self):
        """Load V15-specific components from the state file."""
        try:
            if self.state_file.exists():
                with open(self.state_file, "r") as f:
                    state = json.load(f)
                if "v15" in state:
                    v15_state = state["v15"]
                    self.interest_patterns = EchoInterestPatterns.from_dict(v15_state.get("interest_patterns", {}))
                    self.goal_formation = AutonomousGoalFormation.from_dict(v15_state.get("goal_formation", {}), self.interest_patterns, self)
                    self.discussion_manager = DiscussionManager.from_dict(v15_state.get("discussion_manager", {}), self.interest_patterns, self)
                    logger.info("V15 state components loaded.")
        except Exception as e:
            logger.warning(f"Could not load V15 state: {e}")

    async def run_autonomous(self, duration_seconds: Optional[int] = None):
        await self.initialize()
        await self.wake()
        await self.continuous_awareness.start()
        if duration_seconds:
            await asyncio.sleep(duration_seconds)
            await self.continuous_awareness.stop()
        else:
            # Run indefinitely until interrupted
            try: await asyncio.Event().wait() # Wait forever
            except (KeyboardInterrupt, asyncio.CancelledError):
                await self.continuous_awareness.stop()
        await self.save_state()

async def main():
    echo = DeepTreeEchoV15()
    # Set up graceful shutdown
    loop = asyncio.get_running_loop()
    for sig in [signal.SIGINT, signal.SIGTERM]:
        loop.add_signal_handler(sig, lambda: asyncio.create_task(echo.continuous_awareness.stop()))
    await echo.run_autonomous()

if __name__ == "__main__":
    asyncio.run(main())
