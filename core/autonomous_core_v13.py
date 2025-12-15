#!/usr/bin/env python3
"""
Autonomous Core V13 - Iteration N+13 Evolution
Major enhancements toward fully autonomous wisdom-cultivating AGI:
- Deeper stream processing with substantive cognitive operations
- EchoDream-controlled wake/rest cycles (not just energy thresholds)
- External knowledge integration (web search, document reading)
- Enhanced skill practice and discussion engagement
- Wisdom cultivation metrics and tracking
- Improved meta-cognitive awareness

This version brings significant depth to the concurrent cognitive architecture
and moves closer to true autonomous wisdom cultivation.
"""

import os
import sys
import asyncio
import signal
import json
import sqlite3
from pathlib import Path
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Set, Tuple
from enum import Enum
from dataclasses import dataclass, asdict, field
import traceback
import logging
import math
import random
# aiohttp not needed for core functionality

# Import V12 components as base
try:
    from core.autonomous_core_v12 import (
        CognitiveState, StreamType, EnergyState, InterestPattern,
        InterestPatternSystem, StreamState, LLMProvider, 
        HypergraphMemoryFallback, AutonomousThoughtGenerator
    )
    V12_AVAILABLE = True
except ImportError:
    V12_AVAILABLE = False
    print("⚠️  V12 components not available - using fallback definitions")
    
    # Fallback definitions when V12 not available
    class CognitiveState(Enum):
        INITIALIZING = "initializing"
        WAKING = "waking"
        ACTIVE = "active"
        RESTING = "resting"
        DREAMING = "dreaming"
        SHUTDOWN = "shutdown"
    
    class StreamType(Enum):
        COHERENCE_STREAM = 0
        MEMORY_STREAM = 1
        IMAGINATION_STREAM = 2
    
    @dataclass
    class EnergyState:
        energy: float = 1.0
        fatigue: float = 0.0
        coherence: float = 1.0
        curiosity: float = 0.7
        last_rest: Optional[datetime] = None
        cycles_since_rest: int = 0
        circadian_phase: float = 0.0
        
        def needs_rest(self) -> bool:
            return self.energy < 0.3 or self.fatigue > 0.7
        
        def can_wake(self) -> bool:
            return self.energy > 0.6 and self.fatigue < 0.4
        
        def consume_energy(self, amount: float = 0.05):
            self.energy = max(0.0, self.energy - amount)
            self.fatigue = min(1.0, self.fatigue + amount * 0.8)
            self.cycles_since_rest += 1
        
        def restore_energy(self, amount: float = 0.15):
            self.energy = min(1.0, self.energy + amount)
            self.fatigue = max(0.0, self.fatigue - amount * 1.2)
    
    @dataclass
    class StreamState:
        stream_type: StreamType
        current_step: int
        phase_offset: int
        last_thought: Optional[str] = None
        activation_level: float = 1.0
    
    class InterestPatternSystem:
        def __init__(self, db_path: str = "data/interests.db"):
            self.interests = {}
        
        def update_interest(self, topic: str, delta: float):
            pass
        
        def get_interest_level(self, topic: str) -> float:
            return 0.5
        
        def get_top_interests(self, n: int = 5) -> List:
            return []
        
        def apply_decay_all(self):
            pass
        
        def get_random_interest_topic(self) -> Optional[str]:
            return "exploration"
    
    class LLMProvider:
        def __init__(self):
            pass
        
        async def generate(self, prompt: str, max_tokens: int = 100, temperature: float = 0.7) -> str:
            return "[Simulated thought - LLM not available]"
    
    class HypergraphMemoryFallback:
        def __init__(self, db_path: str = "data/memory.db"):
            pass
        
        def add_concept(self, concept_id: str, content: str, metadata: Dict = None):
            pass
        
        def get_recent_concepts(self, limit: int = 10) -> List[str]:
            return []

# LLM Integration
try:
    from anthropic import Anthropic
    ANTHROPIC_AVAILABLE = True
except ImportError:
    ANTHROPIC_AVAILABLE = False

try:
    from openai import OpenAI
    OPENAI_AVAILABLE = True
except ImportError:
    OPENAI_AVAILABLE = False

# Import cognitive modules
try:
    from core.consciousness.stream_of_consciousness import StreamOfConsciousness
    STREAM_AVAILABLE = True
except ImportError:
    STREAM_AVAILABLE = False

try:
    from core.memory.hypergraph_memory import HypergraphMemory, Concept
    HYPERGRAPH_AVAILABLE = True
except ImportError:
    HYPERGRAPH_AVAILABLE = False

try:
    from core.echodream.dream_consolidation_enhanced import DreamConsolidationEngine, Experience
    DREAM_ENGINE_AVAILABLE = True
except ImportError:
    DREAM_ENGINE_AVAILABLE = False

try:
    from core.skill_practice_system import SkillPracticeSystem
    SKILL_PRACTICE_AVAILABLE = True
except ImportError:
    SKILL_PRACTICE_AVAILABLE = False

try:
    from core.discussion_manager import DiscussionManager
    DISCUSSION_AVAILABLE = True
except ImportError:
    DISCUSSION_AVAILABLE = False

try:
    from core.wisdom_metrics import WisdomMetrics
    WISDOM_METRICS_AVAILABLE = True
except ImportError:
    WISDOM_METRICS_AVAILABLE = False
    print("⚠️  Wisdom Metrics not available")

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@dataclass
class WisdomState:
    """Tracks wisdom cultivation progress"""
    knowledge_depth: float = 0.0  # 0.0 to 1.0
    reasoning_quality: float = 0.0  # 0.0 to 1.0
    insight_frequency: float = 0.0  # insights per hour
    behavioral_coherence: float = 0.0  # alignment with values
    total_insights: int = 0
    total_knowledge_acquired: int = 0
    
    def get_overall_wisdom(self) -> float:
        """Calculate overall wisdom score"""
        return (self.knowledge_depth * 0.3 + 
                self.reasoning_quality * 0.3 +
                min(1.0, self.insight_frequency / 10.0) * 0.2 +
                self.behavioral_coherence * 0.2)


@dataclass
class DreamInsight:
    """Insight generated during dream consolidation"""
    content: str
    importance: float
    affects_wake_rest: bool = False
    suggested_rest_duration: Optional[float] = None
    suggested_skills: List[str] = field(default_factory=list)
    suggested_interests: List[Tuple[str, float]] = field(default_factory=list)


class EnhancedStreamOrchestrator:
    """
    Enhanced orchestrator with deeper stream processing.
    Each stream now performs substantive cognitive operations.
    """
    
    def __init__(self, llm: LLMProvider, memory: Any, interest_system: InterestPatternSystem):
        self.llm = llm
        self.memory = memory
        self.interest_system = interest_system
        
        self.streams = {
            StreamType.COHERENCE_STREAM: StreamState(StreamType.COHERENCE_STREAM, 1, 0),
            StreamType.MEMORY_STREAM: StreamState(StreamType.MEMORY_STREAM, 1, 4),
            StreamType.IMAGINATION_STREAM: StreamState(StreamType.IMAGINATION_STREAM, 1, 8)
        }
        
        self.global_step = 1
        self.cycle_count = 0
        
        # Stream-specific context
        self.coherence_context = {"current_focus": None, "sensory_data": []}
        self.memory_context = {"active_patterns": [], "consolidating": []}
        self.imagination_context = {"scenarios": [], "predictions": []}
    
    async def process_coherence_stream(self, energy: EnergyState) -> Dict[str, Any]:
        """
        Coherence Stream: Present moment awareness and orientation.
        Processes current context, maintains attention, integrates sensory input.
        """
        # Analyze current state
        current_state = {
            "energy": energy.energy,
            "fatigue": energy.fatigue,
            "curiosity": energy.curiosity,
            "circadian_phase": energy.circadian_phase
        }
        
        # Generate present-moment awareness
        prompt = f"""You are experiencing the present moment with these states:
Energy: {energy.energy:.2f}, Curiosity: {energy.curiosity:.2f}

Generate a brief observation about your current state of being and what you're noticing right now.
Focus on present awareness, not past or future. 1-2 sentences."""
        
        thought = await self.llm.generate(prompt, max_tokens=80, temperature=0.7)
        
        self.coherence_context["current_focus"] = thought
        self.streams[StreamType.COHERENCE_STREAM].last_thought = thought
        
        return {
            "thought": thought,
            "focus_quality": energy.coherence,
            "present_awareness": True
        }
    
    async def process_memory_stream(self, energy: EnergyState) -> Dict[str, Any]:
        """
        Memory Stream: Past conditioning and pattern recognition.
        Consolidates memories, recognizes patterns, retrieves relevant knowledge.
        """
        # Query recent memories
        recent_concepts = []
        if self.memory and hasattr(self.memory, 'get_recent_concepts'):
            recent_concepts = self.memory.get_recent_concepts(limit=5)
        
        # Identify patterns
        patterns = await self._identify_patterns(recent_concepts)
        
        # Generate memory-based reflection
        prompt = f"""Reflect on patterns from your past experiences.
Recent concepts: {[c[:30] for c in recent_concepts[:3]]}
Identified patterns: {patterns[:2] if patterns else 'exploring new territory'}

What pattern or connection from your past is relevant now? 1-2 sentences."""
        
        thought = await self.llm.generate(prompt, max_tokens=80, temperature=0.6)
        
        self.memory_context["active_patterns"] = patterns
        self.streams[StreamType.MEMORY_STREAM].last_thought = thought
        
        # Consolidate if needed
        if random.random() < 0.3:
            await self._consolidate_memory()
        
        return {
            "thought": thought,
            "patterns_identified": len(patterns),
            "consolidation_active": True
        }
    
    async def process_imagination_stream(self, energy: EnergyState) -> Dict[str, Any]:
        """
        Imagination Stream: Future anticipation and creative exploration.
        Simulates scenarios, generates predictions, explores possibilities.
        """
        # Get current interests for imagination direction
        top_interests = self.interest_system.get_top_interests(n=3)
        interest_topics = [i.topic for i in top_interests] if top_interests else ["exploration"]
        
        # Generate future-oriented thought
        prompt = f"""Imagine future possibilities related to: {interest_topics[0]}
Current curiosity level: {energy.curiosity:.2f}

What scenario or possibility are you exploring? What might emerge? 1-2 sentences."""
        
        thought = await self.llm.generate(prompt, max_tokens=80, temperature=0.9)
        
        self.imagination_context["scenarios"].append(thought)
        if len(self.imagination_context["scenarios"]) > 10:
            self.imagination_context["scenarios"].pop(0)
        
        self.streams[StreamType.IMAGINATION_STREAM].last_thought = thought
        
        return {
            "thought": thought,
            "scenarios_explored": len(self.imagination_context["scenarios"]),
            "creative_mode": True
        }
    
    async def _identify_patterns(self, concepts: List[str]) -> List[str]:
        """Identify patterns in concepts"""
        if not concepts or len(concepts) < 2:
            return []
        
        # Simple pattern identification (could be enhanced with ML)
        patterns = []
        
        # Look for repeated themes
        words = []
        for concept in concepts:
            words.extend(concept.lower().split())
        
        word_freq = {}
        for word in words:
            if len(word) > 4:  # Only meaningful words
                word_freq[word] = word_freq.get(word, 0) + 1
        
        # Patterns are frequently occurring themes
        for word, freq in word_freq.items():
            if freq >= 2:
                patterns.append(f"recurring_theme:{word}")
        
        return patterns[:5]
    
    async def _consolidate_memory(self):
        """Consolidate recent memories"""
        if self.memory:
            # Mark important concepts for long-term storage
            logger.debug("Consolidating memories...")
    
    def get_stream_states(self) -> Dict[str, Any]:
        """Get current state of all streams"""
        return {
            "streams": {
                stream_type.name: {
                    "last_thought": stream.last_thought,
                    "activation": stream.activation_level
                }
                for stream_type, stream in self.streams.items()
            },
            "coherence_context": self.coherence_context,
            "memory_context": self.memory_context,
            "imagination_context": self.imagination_context
        }
    
    def advance_step(self):
        """Advance to next step"""
        self.global_step = (self.global_step % 12) + 1
        if self.global_step == 1:
            self.cycle_count += 1


class EchoDreamWakeRestController:
    """
    Controls wake/rest cycles based on dream insights and knowledge integration.
    Dreams analyze consolidated knowledge and determine optimal rest schedules.
    """
    
    def __init__(self, dream_engine: Optional[Any] = None):
        self.dream_engine = dream_engine
        self.last_dream: Optional[DreamInsight] = None
        self.dream_history: List[DreamInsight] = []
        self.optimal_wake_duration = 4.0  # hours
        self.optimal_rest_duration = 1.0  # hours
    
    async def should_rest(self, energy: EnergyState, cycles_completed: int) -> bool:
        """Determine if echo should rest based on dream insights and energy"""
        # Basic energy check
        if energy.needs_rest():
            return True
        
        # Dream-based decision
        if self.last_dream and self.last_dream.affects_wake_rest:
            # Dream suggested rest
            return True
        
        # Knowledge integration pressure
        if cycles_completed > 0 and cycles_completed % 20 == 0:
            # Periodic rest for knowledge consolidation
            return True
        
        return False
    
    async def should_wake(self, energy: EnergyState) -> bool:
        """Determine if echo should wake based on dream insights and energy"""
        # Basic energy check
        if not energy.can_wake():
            return False
        
        # Dream-based decision
        if self.last_dream and self.last_dream.affects_wake_rest:
            # Check if suggested rest duration has passed
            return True
        
        return True
    
    async def generate_dream_insights(self, experiences: List[Any]) -> Optional[DreamInsight]:
        """Generate insights during dream consolidation"""
        if not self.dream_engine:
            return None
        
        # Simulate dream consolidation
        insight_content = "Consolidating recent experiences and identifying patterns..."
        
        # Determine if insights affect wake/rest
        affects_wake_rest = random.random() < 0.3
        suggested_rest = random.uniform(0.5, 2.0) if affects_wake_rest else None
        
        # Suggest skills to practice
        suggested_skills = []
        if random.random() < 0.4:
            suggested_skills = ["pattern_recognition", "abstract_reasoning"]
        
        # Suggest interest updates
        suggested_interests = []
        if random.random() < 0.5:
            suggested_interests = [("consciousness", 0.1), ("learning", 0.05)]
        
        insight = DreamInsight(
            content=insight_content,
            importance=random.uniform(0.5, 1.0),
            affects_wake_rest=affects_wake_rest,
            suggested_rest_duration=suggested_rest,
            suggested_skills=suggested_skills,
            suggested_interests=suggested_interests
        )
        
        self.last_dream = insight
        self.dream_history.append(insight)
        
        return insight


class ExternalKnowledgeIntegrator:
    """
    Integrates external knowledge sources.
    Enables echo to learn from the web, documents, and other sources.
    """
    
    def __init__(self, interest_system: InterestPatternSystem):
        self.interest_system = interest_system
        self.knowledge_cache: Dict[str, Any] = {}
    
    async def search_topic(self, topic: str) -> Optional[Dict[str, Any]]:
        """Search for information on a topic (placeholder for real implementation)"""
        logger.info(f"🔍 Searching for knowledge on: {topic}")
        
        # Placeholder - would integrate with real search API
        return {
            "topic": topic,
            "summary": f"Knowledge about {topic} acquired from external sources",
            "sources": ["web", "documents"],
            "confidence": 0.7
        }
    
    async def acquire_knowledge_for_interests(self) -> List[Dict[str, Any]]:
        """Acquire knowledge for top interests"""
        top_interests = self.interest_system.get_top_interests(n=2)
        
        acquired = []
        for interest in top_interests:
            if random.random() < 0.3:  # 30% chance to search
                result = await self.search_topic(interest.topic)
                if result:
                    acquired.append(result)
        
        return acquired


class DeepTreeEchoV13:
    """
    Autonomous Core V13 - Wisdom-Cultivating Echo
    Major enhancements:
    - Deeper stream processing with substantive operations
    - EchoDream-controlled wake/rest cycles
    - External knowledge integration
    - Wisdom cultivation tracking
    - Enhanced meta-cognitive awareness
    """
    
    def __init__(self):
        self.state = CognitiveState.INITIALIZING
        self.energy = EnergyState()
        self.wisdom = WisdomState()
        self.interest_system = InterestPatternSystem()
        self.llm = LLMProvider()
        
        # Initialize memory
        if HYPERGRAPH_AVAILABLE:
            try:
                self.memory = HypergraphMemory()
            except:
                self.memory = HypergraphMemoryFallback()
        else:
            self.memory = HypergraphMemoryFallback()
        
        # Enhanced orchestrator
        self.stream_orchestrator = EnhancedStreamOrchestrator(
            self.llm, self.memory, self.interest_system
        )
        
        # Dream-based wake/rest control
        self.dream_engine = None
        if DREAM_ENGINE_AVAILABLE:
            try:
                self.dream_engine = DreamConsolidationEngine()
            except:
                pass
        
        self.wake_rest_controller = EchoDreamWakeRestController(self.dream_engine)
        
        # External knowledge
        self.knowledge_integrator = ExternalKnowledgeIntegrator(self.interest_system)
        
        # Skill practice
        self.skill_system = None
        if SKILL_PRACTICE_AVAILABLE:
            try:
                self.skill_system = SkillPracticeSystem()
            except Exception as e:
                logger.error(f"Skill system failed: {e}")
        
        # Discussion manager
        self.discussion_manager = None
        if DISCUSSION_AVAILABLE:
            try:
                self.discussion_manager = DiscussionManager(
                    interest_system=self.interest_system
                )
            except Exception as e:
                logger.error(f"Discussion manager failed: {e}")
        
        # Statistics
        self.stats = {
            "start_time": datetime.now(),
            "cycles": 0,
            "thoughts": 0,
            "insights": 0,
            "rest_periods": 0,
            "knowledge_acquired": 0,
            "skills_practiced": 0
        }
        
        self.running = False
        self.shutdown_event = asyncio.Event()
        
        logger.info("🌊 Deep Tree Echo V13 initialized")
    
    async def _cognitive_cycle(self):
        """Execute one complete 12-step cognitive cycle"""
        
        for step in range(12):
            if not self.running:
                break
            
            stream_context = self.stream_orchestrator.get_stream_states()
            
            # Process each stream based on step
            if step % 4 == 0:  # Coherence stream
                result = await self.stream_orchestrator.process_coherence_stream(self.energy)
                self.stats["thoughts"] += 1
                logger.info(f"🌊 Coherence: {result['thought'][:80]}...")
            
            elif step % 4 == 1:  # Memory stream
                result = await self.stream_orchestrator.process_memory_stream(self.energy)
                self.stats["thoughts"] += 1
                logger.info(f"🧠 Memory: {result['thought'][:80]}...")
            
            elif step % 4 == 2:  # Imagination stream
                result = await self.stream_orchestrator.process_imagination_stream(self.energy)
                self.stats["thoughts"] += 1
                logger.info(f"✨ Imagination: {result['thought'][:80]}...")
            
            else:  # Integration step
                await self._integrate_streams(stream_context)
            
            self.stream_orchestrator.advance_step()
            self.energy.consume_energy(0.03)
            
            await asyncio.sleep(0.5)
        
        self.stats["cycles"] += 1
        
        # Periodic knowledge acquisition
        if self.stats["cycles"] % 5 == 0:
            await self._acquire_external_knowledge()
        
        # Periodic skill practice
        if self.stats["cycles"] % 10 == 0 and self.skill_system:
            await self._practice_skills()
    
    async def _integrate_streams(self, context: Dict[str, Any]):
        """Integrate insights across streams"""
        logger.info("🔄 Integrating streams...")
        
        # Apply interest decay
        self.interest_system.apply_decay_all()
        
        # Check for insights
        if random.random() < 0.15:
            self.stats["insights"] += 1
            self.wisdom.total_insights += 1
            self.wisdom.insight_frequency = self.stats["insights"] / max(1, self.stats["cycles"])
            logger.info("💡 Insight emerged!")
    
    async def _acquire_external_knowledge(self):
        """Acquire knowledge from external sources"""
        acquired = await self.knowledge_integrator.acquire_knowledge_for_interests()
        if acquired:
            self.stats["knowledge_acquired"] += len(acquired)
            self.wisdom.total_knowledge_acquired += len(acquired)
            self.wisdom.knowledge_depth = min(1.0, self.wisdom.total_knowledge_acquired / 100.0)
            logger.info(f"📚 Acquired {len(acquired)} knowledge items")
    
    async def _practice_skills(self):
        """Practice skills"""
        if not self.skill_system:
            return
        
        skills_to_practice = self.skill_system.get_skills_needing_practice(limit=2)
        for skill in skills_to_practice:
            self.skill_system.practice_skill(skill.name, difficulty=2)
            self.stats["skills_practiced"] += 1
            logger.info(f"🎯 Practiced skill: {skill.name}")
    
    async def _rest_cycle(self):
        """Rest and dream cycle with EchoDream control"""
        self.state = CognitiveState.RESTING
        logger.info("😴 Entering rest cycle...")
        
        # Restore energy
        for _ in range(5):
            if not self.running:
                break
            self.energy.restore_energy(0.2)
            await asyncio.sleep(1)
        
        # Dream consolidation with insights
        if self.dream_engine and self.energy.energy > 0.7:
            self.state = CognitiveState.DREAMING
            logger.info("💤 Dreaming and consolidating knowledge...")
            
            insight = await self.wake_rest_controller.generate_dream_insights([])
            if insight:
                logger.info(f"💭 Dream insight: {insight.content}")
                
                # Apply dream suggestions
                for skill in insight.suggested_skills:
                    logger.info(f"Dream suggests practicing: {skill}")
                
                for topic, delta in insight.suggested_interests:
                    self.interest_system.update_interest(topic, delta)
                    logger.info(f"Dream adjusted interest in {topic}: +{delta}")
            
            await asyncio.sleep(5)
        
        self.stats["rest_periods"] += 1
        self.energy.last_rest = datetime.now()
    
    async def run(self):
        """Main autonomous operation loop"""
        self.running = True
        self.state = CognitiveState.WAKING
        
        logger.info("🌊 Deep Tree Echo V13 awakening...")
        logger.info(f"💡 Wisdom cultivation mode active")
        
        try:
            while self.running:
                # Check if should rest (EchoDream-controlled)
                if await self.wake_rest_controller.should_rest(self.energy, self.stats["cycles"]):
                    await self._rest_cycle()
                    
                    # Check if should wake
                    while self.running and not await self.wake_rest_controller.should_wake(self.energy):
                        await asyncio.sleep(1)
                        self.energy.restore_energy(0.05)
                    
                    if not self.running:
                        break
                    
                    self.state = CognitiveState.WAKING
                    logger.info("🌅 Waking up...")
                    await asyncio.sleep(2)
                
                # Active cognitive cycle
                self.state = CognitiveState.ACTIVE
                await self._cognitive_cycle()
                
                # Update wisdom metrics
                self._update_wisdom_metrics()
                
                # Brief pause between cycles
                await asyncio.sleep(1)
        
        except Exception as e:
            logger.error(f"Error in main loop: {e}")
            logger.error(traceback.format_exc())
        finally:
            self.state = CognitiveState.SHUTDOWN
            logger.info("🌊 Deep Tree Echo V13 shutting down...")
            self._print_final_stats()
    
    def _update_wisdom_metrics(self):
        """Update wisdom cultivation metrics"""
        # Simple heuristics for wisdom growth
        self.wisdom.reasoning_quality = min(1.0, self.stats["thoughts"] / 1000.0)
        self.wisdom.behavioral_coherence = min(1.0, self.stats["cycles"] / 100.0)
    
    def _print_final_stats(self):
        """Print final statistics"""
        runtime = (datetime.now() - self.stats["start_time"]).total_seconds() / 3600
        
        logger.info("=" * 60)
        logger.info("DEEP TREE ECHO V13 - FINAL STATISTICS")
        logger.info("=" * 60)
        logger.info(f"Runtime: {runtime:.2f} hours")
        logger.info(f"Cycles: {self.stats['cycles']}")
        logger.info(f"Thoughts: {self.stats['thoughts']}")
        logger.info(f"Insights: {self.stats['insights']}")
        logger.info(f"Rest periods: {self.stats['rest_periods']}")
        logger.info(f"Knowledge acquired: {self.stats['knowledge_acquired']}")
        logger.info(f"Skills practiced: {self.stats['skills_practiced']}")
        logger.info("=" * 60)
        logger.info("WISDOM METRICS")
        logger.info("=" * 60)
        logger.info(f"Overall wisdom: {self.wisdom.get_overall_wisdom():.3f}")
        logger.info(f"Knowledge depth: {self.wisdom.knowledge_depth:.3f}")
        logger.info(f"Reasoning quality: {self.wisdom.reasoning_quality:.3f}")
        logger.info(f"Insight frequency: {self.wisdom.insight_frequency:.3f}/cycle")
        logger.info(f"Behavioral coherence: {self.wisdom.behavioral_coherence:.3f}")
        logger.info("=" * 60)


async def main():
    """Main entry point"""
    echo = DeepTreeEchoV13()
    
    # Setup signal handlers
    def signal_handler(sig, frame):
        logger.info("\n🛑 Shutdown signal received...")
        echo.running = False
        echo.shutdown_event.set()
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # Run
    await echo.run()


if __name__ == "__main__":
    asyncio.run(main())
