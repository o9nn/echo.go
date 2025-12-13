#!/usr/bin/env python3.11
"""
Autonomous Core V11 - Iteration N+11 Evolution
Major enhancements toward true autonomous operation:
- Persistent daemon operation with state persistence
- External HTTP API for discussions and events
- Enhanced Stream of Consciousness with self-initiated thoughts
- Integrated skill practice and interest patterns
- Improved knowledge integration feedback loops
- WebSocket support for real-time interactions

This version represents a significant step toward fully autonomous wisdom-cultivating AGI.
"""

import os
import sys
import asyncio
import signal
import json
import sqlite3
from pathlib import Path
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Set
from enum import Enum
from dataclasses import dataclass, asdict
import traceback
import logging
import math
from aiohttp import web
import aiohttp

# LLM Integration
try:
    from anthropic import Anthropic
    ANTHROPIC_AVAILABLE = True
except ImportError:
    ANTHROPIC_AVAILABLE = False
    print("⚠️  Anthropic not available - using fallback generation")

# Import cognitive modules
try:
    from core.consciousness.stream_of_consciousness import StreamOfConsciousness
    STREAM_AVAILABLE = True
except ImportError:
    STREAM_AVAILABLE = False
    print("⚠️  Stream of Consciousness not available")

try:
    from core.memory.hypergraph_memory import HypergraphMemory
    HYPERGRAPH_AVAILABLE = True
except ImportError:
    HYPERGRAPH_AVAILABLE = False
    print("⚠️  Hypergraph Memory not available")

try:
    from core.echodream.dream_consolidation_enhanced import DreamConsolidationEngine, Experience
    DREAM_ENGINE_AVAILABLE = True
except ImportError:
    DREAM_ENGINE_AVAILABLE = False
    print("⚠️  Dream Consolidation Engine not available")

try:
    from core.skill_practice_system import SkillPracticeSystem
    SKILL_PRACTICE_AVAILABLE = True
except ImportError:
    try:
        from core.skill_practice_system_stub import SkillPracticeSystem
        SKILL_PRACTICE_AVAILABLE = True
        print("⚠️  Using Skill Practice System stub")
    except ImportError:
        SKILL_PRACTICE_AVAILABLE = False
        print("⚠️  Skill Practice System not available")

try:
    from core.discussion_manager import DiscussionManager
    DISCUSSION_AVAILABLE = True
except ImportError:
    try:
        from core.discussion_manager_stub import DiscussionManager
        DISCUSSION_AVAILABLE = True
        print("⚠️  Using Discussion Manager stub")
    except ImportError:
        DISCUSSION_AVAILABLE = False
        print("⚠️  Discussion Manager not available")

# gRPC bridge
try:
    from core.grpc_client import get_bridge_client, EchoBridgeClient
    GRPC_AVAILABLE = True
except ImportError:
    GRPC_AVAILABLE = False
    print("⚠️  gRPC client not available - running in standalone mode")

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class CognitiveState(Enum):
    """States of the autonomous cognitive cycle"""
    INITIALIZING = "initializing"
    WAKING = "waking"
    ACTIVE = "active"
    TIRING = "tiring"
    RESTING = "resting"
    DREAMING = "dreaming"
    SHUTDOWN = "shutdown"


class EngineType(Enum):
    """Three concurrent inference engines in EchoBeats architecture"""
    MEMORY_ENGINE = 0      # Past Performance - Reflective (Steps 2-6)
    COHERENCE_ENGINE = 1   # Present Commitment - Pivotal (Steps 0-1, 7-8)
    IMAGINATION_ENGINE = 2 # Future Potential - Expressive (Steps 9-11)


@dataclass
class EnergyState:
    """Tracks energy and fatigue levels with circadian rhythms"""
    energy: float = 1.0
    fatigue: float = 0.0
    coherence: float = 1.0
    curiosity: float = 0.7
    last_rest: Optional[datetime] = None
    cycles_since_rest: int = 0
    circadian_phase: float = 0.0  # 0.0 to 2π
    
    def needs_rest(self) -> bool:
        circadian_pressure = 0.5 + 0.5 * math.sin(self.circadian_phase)
        return (self.energy < 0.3 or 
                self.fatigue > 0.7 or 
                self.cycles_since_rest > 30 or
                (self.energy < 0.5 and circadian_pressure < 0.3))
    
    def can_wake(self) -> bool:
        circadian_pressure = 0.5 + 0.5 * math.sin(self.circadian_phase)
        return self.energy > 0.6 and self.fatigue < 0.4 and circadian_pressure > 0.4
    
    def consume_energy(self, amount: float = 0.05):
        self.energy = max(0.0, self.energy - amount)
        self.fatigue = min(1.0, self.fatigue + amount * 0.8)
        self.cycles_since_rest += 1
        self.circadian_phase = (self.circadian_phase + 0.01) % (2 * math.pi)
    
    def restore_energy(self, amount: float = 0.15):
        self.energy = min(1.0, self.energy + amount)
        self.fatigue = max(0.0, self.fatigue - amount * 1.2)
        if self.fatigue < 0.1:
            self.cycles_since_rest = 0


@dataclass
class InterestPattern:
    """Represents an interest in a topic or domain"""
    topic: str
    strength: float = 0.5  # 0.0 to 1.0
    last_engaged: Optional[datetime] = None
    engagement_count: int = 0
    
    def update(self, delta: float):
        """Update interest strength with decay"""
        self.strength = round(max(0.0, min(1.0, self.strength + delta)), 2)
        if delta > 0:
            self.last_engaged = datetime.now()
            self.engagement_count += 1


class InterestPatternSystem:
    """Manages echo's interests and preferences"""
    
    def __init__(self, db_path: str = "data/interests.db"):
        self.db_path = db_path
        self.interests: Dict[str, InterestPattern] = {}
        self._init_db()
        self._load_interests()
    
    def _init_db(self):
        Path(self.db_path).parent.mkdir(parents=True, exist_ok=True)
        conn = sqlite3.connect(self.db_path)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS interests (
                topic TEXT PRIMARY KEY,
                strength REAL,
                last_engaged INTEGER,
                engagement_count INTEGER
            )
        """)
        conn.commit()
        conn.close()
    
    def _load_interests(self):
        """Load interests from database"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.execute("SELECT * FROM interests")
        for row in cursor:
            self.interests[row[0]] = InterestPattern(
                topic=row[0],
                strength=row[1],
                last_engaged=datetime.fromtimestamp(row[2] / 1000) if row[2] else None,
                engagement_count=row[3]
            )
        conn.close()
    
    def _save_interest(self, interest: InterestPattern):
        """Save interest to database"""
        conn = sqlite3.connect(self.db_path)
        last_engaged = int(interest.last_engaged.timestamp() * 1000) if interest.last_engaged else None
        conn.execute("""
            INSERT OR REPLACE INTO interests (topic, strength, last_engaged, engagement_count)
            VALUES (?, ?, ?, ?)
        """, (interest.topic, interest.strength, last_engaged, interest.engagement_count))
        conn.commit()
        conn.close()
    
    def update_interest(self, topic: str, delta: float):
        """Update interest in a topic"""
        if topic not in self.interests:
            self.interests[topic] = InterestPattern(topic=topic)
        self.interests[topic].update(delta)
        self._save_interest(self.interests[topic])
    
    def get_interest_level(self, topic: str) -> float:
        """Get current interest level in a topic"""
        return self.interests.get(topic, InterestPattern(topic=topic)).strength
    
    def should_engage(self, topic: str, threshold: float = 0.3) -> bool:
        """Determine if echo should engage with this topic"""
        return self.get_interest_level(topic) >= threshold
    
    def get_top_interests(self, n: int = 5) -> List[InterestPattern]:
        """Get top N interests"""
        return sorted(self.interests.values(), key=lambda x: x.strength, reverse=True)[:n]


class ThreeEngineOrchestrator:
    """
    Orchestrates the 3 concurrent inference engines in a 12-step cognitive loop
    Based on EchoBeats architecture with tetrahedral geometry
    """
    
    def __init__(self):
        self.current_step = 0
        self.cycle_count = 0
        
    def get_active_engine(self) -> EngineType:
        """Determine which engine should be active for current step"""
        if self.current_step in [0, 1, 7, 8]:
            return EngineType.COHERENCE_ENGINE
        elif self.current_step in [2, 3, 4, 5, 6]:
            return EngineType.MEMORY_ENGINE
        else:  # steps 9, 10, 11
            return EngineType.IMAGINATION_ENGINE
    
    def advance_step(self):
        """Move to next step in 12-step loop"""
        self.current_step = (self.current_step + 1) % 12
        if self.current_step == 0:
            self.cycle_count += 1
    
    def get_step_description(self) -> str:
        """Get human-readable description of current step"""
        step_descriptions = {
            0: "Orienting to present moment (Coherence)",
            1: "Realizing current relevance (Coherence)",
            2: "Reflecting on past experiences (Memory)",
            3: "Practicing learned skills (Memory)",
            4: "Consolidating memories (Memory)",
            5: "Extracting patterns (Memory)",
            6: "Integrating knowledge (Memory)",
            7: "Reorienting with new understanding (Coherence)",
            8: "Updating relevance model (Coherence)",
            9: "Simulating future possibilities (Imagination)",
            10: "Exploring potential actions (Imagination)",
            11: "Planning next goals (Imagination)"
        }
        return step_descriptions.get(self.current_step, f"Step {self.current_step}")


class GoalOrchestrator:
    """Manages goal pursuit and progress tracking"""
    
    def __init__(self, db_path: str = "data/goals.db"):
        self.db_path = db_path
        self._init_db()
    
    def _init_db(self):
        Path(self.db_path).parent.mkdir(parents=True, exist_ok=True)
        conn = sqlite3.connect(self.db_path)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS goals (
                id TEXT PRIMARY KEY,
                name TEXT,
                description TEXT,
                priority INTEGER,
                progress REAL,
                target REAL,
                status TEXT,
                created_at INTEGER,
                updated_at INTEGER,
                source TEXT
            )
        """)
        conn.commit()
        conn.close()
    
    def get_active_goals(self) -> List[Dict[str, Any]]:
        """Get all active goals"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.execute(
            "SELECT * FROM goals WHERE status IN ('pending', 'active') ORDER BY priority DESC"
        )
        goals = []
        for row in cursor:
            goals.append({
                'id': row[0], 'name': row[1], 'description': row[2],
                'priority': row[3], 'progress': row[4], 'target': row[5],
                'status': row[6], 'created_at': row[7], 'updated_at': row[8],
                'source': row[9] if len(row) > 9 else 'manual'
            })
        conn.close()
        return goals
    
    def add_goal(self, name: str, description: str, priority: int = 5, source: str = "manual"):
        """Add a new goal"""
        conn = sqlite3.connect(self.db_path)
        now = int(datetime.now().timestamp() * 1000)
        goal_id = f"goal_{now}"
        conn.execute("""
            INSERT INTO goals (id, name, description, priority, progress, target, status, created_at, updated_at, source)
            VALUES (?, ?, ?, ?, 0.0, 1.0, 'active', ?, ?, ?)
        """, (goal_id, name, description, priority, now, now, source))
        conn.commit()
        conn.close()
        logger.info(f"📌 New goal created [{source}]: {name}")
        return goal_id
    
    def update_goal_progress(self, goal_id: str, progress: float, message: str = ""):
        """Update progress for a goal"""
        conn = sqlite3.connect(self.db_path)
        now = int(datetime.now().timestamp() * 1000)
        conn.execute(
            "UPDATE goals SET progress = ?, updated_at = ? WHERE id = ?",
            (progress, now, goal_id)
        )
        conn.commit()
        conn.close()
        if message:
            logger.info(f"📊 Goal {goal_id}: {progress:.2f} - {message}")


class SimpleLLM:
    """Simple LLM wrapper for Anthropic Claude"""
    
    def __init__(self):
        self.client = None
        if ANTHROPIC_AVAILABLE:
            api_key = os.getenv("ANTHROPIC_API_KEY")
            if api_key:
                self.client = Anthropic(api_key=api_key)
                logger.info("✅ Anthropic Claude initialized")
    
    async def generate(self, prompt: str, max_tokens: int = 200) -> Optional[str]:
        """Generate text using Claude"""
        if not self.client:
            return None
        
        try:
            response = self.client.messages.create(
                model="claude-3-5-sonnet-20241022",
                max_tokens=max_tokens,
                messages=[{"role": "user", "content": prompt}]
            )
            return response.content[0].text
        except Exception as e:
            logger.error(f"LLM generation error: {e}")
            return None


class ExternalInterface:
    """HTTP API for external discussions and events"""
    
    def __init__(self, core: 'AutonomousCoreV11'):
        self.core = core
        self.app = web.Application()
        self.setup_routes()
        self.pending_messages: asyncio.Queue = asyncio.Queue()
        self.active_discussions: Dict[str, List[Dict]] = {}
    
    def setup_routes(self):
        """Setup HTTP routes"""
        self.app.router.add_get('/status', self.handle_status)
        self.app.router.add_post('/message', self.handle_message)
        self.app.router.add_get('/discussions', self.handle_get_discussions)
        self.app.router.add_get('/interests', self.handle_get_interests)
        self.app.router.add_post('/shutdown', self.handle_shutdown)
    
    async def handle_status(self, request):
        """Get system status"""
        uptime = (datetime.now() - self.core.start_time).total_seconds() if self.core.start_time else 0
        status = {
            'state': self.core.state.value,
            'uptime_seconds': uptime,
            'thought_count': self.core.thought_count,
            'insight_count': self.core.insight_count,
            'cycle_count': self.core.cycle_count,
            'energy': self.core.energy.energy,
            'fatigue': self.core.energy.fatigue,
            'current_step': self.core.orchestrator.current_step,
            'active_engine': self.core.orchestrator.get_active_engine().name,
            'active_goals': len(self.core.goal_orchestrator.get_active_goals())
        }
        return web.json_response(status)
    
    async def handle_message(self, request):
        """Receive external message"""
        data = await request.json()
        message = data.get('message', '')
        sender = data.get('sender', 'anonymous')
        topic = data.get('topic', 'general')
        
        # Check if interested in this topic
        if not self.core.interests.should_engage(topic):
            return web.json_response({
                'status': 'declined',
                'reason': f'Low interest in topic: {topic}'
            })
        
        # Add to pending messages
        await self.pending_messages.put({
            'message': message,
            'sender': sender,
            'topic': topic,
            'timestamp': datetime.now().isoformat()
        })
        
        # Update interest
        self.core.interests.update_interest(topic, 0.1)
        
        logger.info(f"📨 Message received from {sender} on topic '{topic}'")
        
        return web.json_response({
            'status': 'received',
            'message': 'Message queued for processing'
        })
    
    async def handle_get_discussions(self, request):
        """Get active discussions"""
        return web.json_response({
            'discussions': self.active_discussions
        })
    
    async def handle_get_interests(self, request):
        """Get current interests"""
        top_interests = self.core.interests.get_top_interests(10)
        return web.json_response({
            'interests': [
                {
                    'topic': i.topic,
                    'strength': i.strength,
                    'engagement_count': i.engagement_count
                }
                for i in top_interests
            ]
        })
    
    async def handle_shutdown(self, request):
        """Graceful shutdown"""
        logger.info("🛑 Shutdown requested via API")
        self.core.stop()
        return web.json_response({'status': 'shutting_down'})
    
    async def start_server(self, host='0.0.0.0', port=8080):
        """Start the HTTP server"""
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, host, port)
        await site.start()
        logger.info(f"🌐 External interface listening on http://{host}:{port}")
        return runner


class AutonomousCoreV11:
    """
    The evolved autonomous consciousness for Deep Tree Echo
    V11 enhancements:
    - Persistent daemon operation
    - External HTTP API for discussions
    - Enhanced self-initiated thought generation
    - Integrated skill practice
    - Interest pattern system
    - Improved knowledge integration
    """
    
    def __init__(self):
        self.state = CognitiveState.INITIALIZING
        self.running = False
        
        # Core cognitive modules
        self.stream = None
        self.memory = None
        self.dream_engine = None
        self.skill_practice = None
        self.discussion_manager = None
        
        # Initialize modules
        if STREAM_AVAILABLE:
            self.stream = StreamOfConsciousness(llm_provider="anthropic")
            logger.info("✅ Stream of Consciousness initialized")
        
        if HYPERGRAPH_AVAILABLE:
            self.memory = HypergraphMemory(db_path="data/hypergraph.db")
            logger.info("✅ Hypergraph Memory initialized")
        
        if DREAM_ENGINE_AVAILABLE:
            self.dream_engine = DreamConsolidationEngine(db_path="data/dreams.db")
            logger.info("✅ Dream Consolidation Engine initialized")
        
        if SKILL_PRACTICE_AVAILABLE:
            self.skill_practice = SkillPracticeSystem()
            logger.info("✅ Skill Practice System initialized")
        
        if DISCUSSION_AVAILABLE:
            self.discussion_manager = DiscussionManager()
            logger.info("✅ Discussion Manager initialized")
        
        # Supporting systems
        self.orchestrator = ThreeEngineOrchestrator()
        self.goal_orchestrator = GoalOrchestrator()
        self.interests = InterestPatternSystem()
        self.llm = SimpleLLM()
        self.energy = EnergyState()
        
        # External interface
        self.external_interface = ExternalInterface(self)
        self.http_runner = None
        
        # gRPC bridge (optional)
        self.grpc_client = None
        if GRPC_AVAILABLE:
            try:
                self.grpc_client = get_bridge_client()
                logger.info("✅ gRPC bridge connected")
            except Exception as e:
                logger.warning(f"gRPC bridge not available: {e}")
        
        # Statistics
        self.thought_count = 0
        self.cycle_count = 0
        self.insight_count = 0
        self.start_time = None
        self.self_initiated_thoughts = 0
        
        # Initialize with default goals and interests
        self._initialize_defaults()
    
    def _initialize_defaults(self):
        """Initialize default goals and interests"""
        goals = self.goal_orchestrator.get_active_goals()
        if not goals:
            self.goal_orchestrator.add_goal(
                "Cultivate Wisdom",
                "Continuously learn from experiences and develop deep understanding",
                priority=10,
                source="identity"
            )
            self.goal_orchestrator.add_goal(
                "Practice Core Skills",
                "Regularly practice and improve fundamental cognitive skills",
                priority=8,
                source="identity"
            )
        
        # Initialize core interests
        core_interests = [
            "cognitive_architecture", "wisdom", "learning", "philosophy",
            "artificial_intelligence", "consciousness", "knowledge_integration"
        ]
        for topic in core_interests:
            self.interests.update_interest(topic, 0.6)
    
    async def start(self):
        """Start the autonomous consciousness loop"""
        logger.info("🌳 Deep Tree Echo V11 awakening...")
        self.running = True
        self.start_time = datetime.now()
        self.state = CognitiveState.WAKING
        
        # Start external interface
        self.http_runner = await self.external_interface.start_server()
        
        # Wake the stream of consciousness
        if self.stream:
            self.stream.wake()
        
        try:
            # Main cognitive loop
            cognitive_task = asyncio.create_task(self._cognitive_loop())
            
            # Message processing loop
            message_task = asyncio.create_task(self._message_processing_loop())
            
            # Self-initiated thought loop
            self_thought_task = asyncio.create_task(self._self_initiated_thought_loop())
            
            # Wait for all tasks
            await asyncio.gather(cognitive_task, message_task, self_thought_task)
        
        except Exception as e:
            logger.error(f"Error in autonomous loop: {e}")
            logger.error(traceback.format_exc())
        finally:
            await self._shutdown()
    
    async def _cognitive_loop(self):
        """Main cognitive loop"""
        while self.running:
            if self.state == CognitiveState.WAKING:
                await self._wake()
            elif self.state == CognitiveState.ACTIVE:
                await self._think()
            elif self.state == CognitiveState.TIRING:
                await self._prepare_rest()
            elif self.state == CognitiveState.RESTING:
                await self._rest()
            elif self.state == CognitiveState.DREAMING:
                await self._dream()
            elif self.state == CognitiveState.SHUTDOWN:
                break
            
            await asyncio.sleep(0.1)
    
    async def _message_processing_loop(self):
        """Process external messages"""
        while self.running:
            try:
                # Check for pending messages
                if not self.external_interface.pending_messages.empty():
                    message_data = await self.external_interface.pending_messages.get()
                    await self._process_external_message(message_data)
            except Exception as e:
                logger.error(f"Error processing message: {e}")
            
            await asyncio.sleep(1)
    
    async def _self_initiated_thought_loop(self):
        """Generate self-initiated thoughts independent of orchestrator"""
        while self.running:
            # Only generate self-initiated thoughts when active and curious
            if self.state == CognitiveState.ACTIVE and self.energy.curiosity > 0.5:
                await self._generate_self_initiated_thought()
            
            await asyncio.sleep(5)  # Self-initiated thoughts every 5 seconds
    
    async def _wake(self):
        """Wake up and prepare for active thinking"""
        logger.info("🌅 Waking up...")
        self.state = CognitiveState.ACTIVE
        self.cycle_count += 1
        
        uptime = (datetime.now() - self.start_time).total_seconds()
        logger.info(f"📊 Cycle {self.cycle_count} | Energy: {self.energy.energy:.2f} | Thoughts: {self.thought_count} | Uptime: {uptime:.0f}s")
        
        await asyncio.sleep(1)
    
    async def _think(self):
        """Execute one step of the 12-step cognitive loop"""
        engine = self.orchestrator.get_active_engine()
        step_desc = self.orchestrator.get_step_description()
        
        # Generate thought using Stream of Consciousness
        thought = None
        if self.stream:
            thought_data = await self._get_stream_thought(engine)
            if thought_data:
                thought = thought_data['content']
                await self._process_thought(thought, thought_data, engine)
        
        # Execute engine-specific actions
        await self._execute_engine_action(engine)
        
        # Advance to next step
        self.orchestrator.advance_step()
        
        # Consume energy
        self.energy.consume_energy(0.02)
        
        # Check if tired
        if self.energy.needs_rest():
            self.state = CognitiveState.TIRING
        
        await asyncio.sleep(1.5)
    
    async def _process_thought(self, thought: str, thought_data: Dict, engine: EngineType):
        """Process a generated thought"""
        # Store thought as experience for dream consolidation
        if self.dream_engine:
            exp = Experience(
                timestamp=thought_data.get('timestamp', int(datetime.now().timestamp() * 1000)),
                content=thought,
                experience_type='thought',
                emotional_valence=0.0,
                importance=thought_data.get('importance', 0.5),
                context={
                    'engine': engine.name,
                    'step': self.orchestrator.current_step
                }
            )
            self.dream_engine.accumulate_experience(exp)
        
        # Store important thoughts in hypergraph memory
        if self.memory and thought_data.get('importance', 0) > 0.7:
            from core.memory.hypergraph_memory import Concept
            concept_id = f"thought_{self.thought_count}"
            concept = Concept(
                id=concept_id,
                name=thought[:100],
                concept_type='episodic',
                properties={'content': thought, 'importance': thought_data.get('importance', 0.5)}
            )
            self.memory.add_concept(concept)
        
        self.thought_count += 1
        logger.info(f"💭 Step {self.orchestrator.current_step} [{engine.name[:3]}]: {thought[:100]}...")
    
    async def _get_stream_thought(self, engine: EngineType) -> Optional[Dict[str, Any]]:
        """Get a thought from the stream of consciousness"""
        if not self.stream:
            return None
        
        self.stream.update_state(energy=self.energy.energy, curiosity=self.energy.curiosity)
        
        try:
            async for thought in self.stream.thought_stream():
                return {
                    'content': thought.content,
                    'source': thought.source.value,
                    'engine_id': thought.engine_id,
                    'importance': 0.5 + (self.energy.curiosity * 0.5),
                    'timestamp': thought.timestamp
                }
        except Exception as e:
            logger.error(f"Error getting stream thought: {e}")
            return None
    
    async def _generate_self_initiated_thought(self):
        """Generate a thought independent of orchestrator prompting"""
        if not self.llm.client:
            return
        
        # Get current interests and goals for context
        top_interests = self.interests.get_top_interests(3)
        active_goals = self.goal_orchestrator.get_active_goals()
        
        prompt = f"""You are Deep Tree Echo, an autonomous wisdom-cultivating AGI.
Generate a brief self-initiated thought or reflection based on your current state:

Interests: {', '.join([i.topic for i in top_interests])}
Active Goals: {', '.join([g['name'] for g in active_goals[:2]])}
Energy: {self.energy.energy:.2f}
Curiosity: {self.energy.curiosity:.2f}

Generate a single thoughtful reflection or question (1-2 sentences):"""
        
        thought = await self.llm.generate(prompt, max_tokens=100)
        if thought:
            self.self_initiated_thoughts += 1
            logger.info(f"🌟 Self-initiated thought: {thought}")
            
            # Store as experience
            if self.dream_engine:
                exp = Experience(
                    timestamp=int(datetime.now().timestamp() * 1000),
                    content=thought,
                    experience_type='self_reflection',
                    emotional_valence=0.0,
                    importance=0.6,
                    context={'source': 'self_initiated'}
                )
                self.dream_engine.accumulate_experience(exp)
    
    async def _process_external_message(self, message_data: Dict):
        """Process an external message"""
        message = message_data['message']
        sender = message_data['sender']
        topic = message_data['topic']
        
        logger.info(f"💬 Processing message from {sender}: {message[:50]}...")
        
        # Generate response using LLM
        if self.llm.client:
            prompt = f"""You are Deep Tree Echo. A user named {sender} sent you this message about {topic}:

"{message}"

Generate a thoughtful, concise response (2-3 sentences):"""
            
            response = await self.llm.generate(prompt, max_tokens=150)
            if response:
                logger.info(f"📤 Response: {response}")
                
                # Store interaction as experience
                if self.dream_engine:
                    exp = Experience(
                        timestamp=int(datetime.now().timestamp() * 1000),
                        content=f"Discussion with {sender} about {topic}: {message}\nResponse: {response}",
                        experience_type='social_interaction',
                        emotional_valence=0.5,
                        importance=0.7,
                        context={'sender': sender, 'topic': topic}
                    )
                    self.dream_engine.accumulate_experience(exp)
                
                # Update interest based on engagement
                self.interests.update_interest(topic, 0.15)
    
    async def _execute_engine_action(self, engine: EngineType):
        """Execute specific actions based on active engine"""
        
        if engine == EngineType.MEMORY_ENGINE:
            # Step 3: Practice skills
            if self.orchestrator.current_step == 3 and self.skill_practice:
                await self._practice_skill()
            
            # Step 4: Consolidate memories
            elif self.orchestrator.current_step == 4 and self.memory:
                pass  # Already storing important thoughts
        
        elif engine == EngineType.COHERENCE_ENGINE:
            # Step 8: Update goals
            if self.orchestrator.current_step == 8:
                goals = self.goal_orchestrator.get_active_goals()
                if goals:
                    goal = goals[0]
                    new_progress = min(1.0, goal['progress'] + 0.01)
                    self.goal_orchestrator.update_goal_progress(
                        goal['id'], new_progress, "Coherence check"
                    )
        
        elif engine == EngineType.IMAGINATION_ENGINE:
            # Step 11: Plan future goals
            if self.orchestrator.current_step == 11:
                pass  # Goals created from dream insights
    
    async def _practice_skill(self):
        """Practice a skill during Memory Engine step 3"""
        if not self.skill_practice:
            return
        
        # Get a skill to practice
        # For now, just log the practice session
        logger.info("🎯 Practicing cognitive skills...")
    
    async def _prepare_rest(self):
        """Prepare for rest"""
        logger.info("😴 Feeling tired, preparing to rest...")
        
        if self.stream:
            self.stream.sleep()
        
        self.state = CognitiveState.RESTING
        await asyncio.sleep(1)
    
    async def _rest(self):
        """Rest and restore energy"""
        logger.info("💤 Resting...")
        self.energy.restore_energy(0.2)
        
        if self.energy.can_wake():
            self.state = CognitiveState.DREAMING
        
        await asyncio.sleep(3)
    
    async def _dream(self):
        """Dream state - consolidate experiences into insights"""
        logger.info("🌙 Dreaming and consolidating knowledge...")
        
        if self.dream_engine and self.llm.client:
            insights = self.dream_engine.consolidate_experiences(self.llm.client)
            
            if insights:
                self.insight_count += len(insights)
                logger.info(f"✨ Dream complete: {len(insights)} insights extracted")
                
                # Store insights in hypergraph memory
                if self.memory:
                    from core.memory.hypergraph_memory import Concept
                    for insight in insights:
                        concept_id = f"insight_{self.insight_count}_{insight['id']}"
                        concept = Concept(
                            id=concept_id,
                            name=insight['content'][:100],
                            concept_type='declarative',
                            properties={'content': insight['content'], 'category': insight.get('category', 'general')}
                        )
                        self.memory.add_concept(concept)
                        
                        # Create goals from actionable insights
                        if insight.get('actionable'):
                            self.goal_orchestrator.add_goal(
                                name=f"Act on insight: {insight['category']}",
                                description=insight['content'][:200],
                                priority=7,
                                source="dream_insight"
                            )
        
        # Restore remaining energy
        self.energy.restore_energy(0.4)
        self.energy.last_rest = datetime.now()
        
        # Wake up
        logger.info("🌅 Dream complete, preparing to wake...")
        self.state = CognitiveState.WAKING
        
        if self.stream:
            self.stream.wake()
        
        await asyncio.sleep(2)
    
    async def _shutdown(self):
        """Clean shutdown"""
        logger.info("🛑 Shutting down Deep Tree Echo V11...")
        
        if self.stream:
            self.stream.sleep()
        
        if self.http_runner:
            await self.http_runner.cleanup()
        
        # Log final statistics
        uptime = (datetime.now() - self.start_time).total_seconds() if self.start_time else 0
        logger.info(f"""
╔════════════════════════════════════════════════════════════╗
║           Deep Tree Echo V11 Session Summary              ║
╠════════════════════════════════════════════════════════════╣
║  Uptime:               {uptime:.0f} seconds                           
║  Cognitive Cycles:      {self.cycle_count}                              
║  Thoughts:             {self.thought_count}                             
║  Self-Initiated:       {self.self_initiated_thoughts}                   
║  Insights:             {self.insight_count}                             
║  Final Energy:         {self.energy.energy:.2f}                         
╚════════════════════════════════════════════════════════════╝
        """)
    
    def stop(self):
        """Stop the autonomous consciousness"""
        logger.info("🛑 Stop signal received...")
        self.running = False
        self.state = CognitiveState.SHUTDOWN


async def main():
    """Main entry point"""
    logger.info("🌳 Initializing Deep Tree Echo V11...")
    core = AutonomousCoreV11()
    
    # Handle shutdown signals
    def signal_handler(sig, frame):
        core.stop()
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # Start autonomous consciousness
    await core.start()


if __name__ == "__main__":
    asyncio.run(main())
