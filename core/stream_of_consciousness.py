#!/usr/bin/env python3
"""
Stream-of-Consciousness Engine
===============================

Implements continuous, narrative-coherent internal monologue generation
that maintains awareness independent of external prompts.

Features:
- Continuous thought stream generation (not discrete cycles)
- Narrative continuity across thoughts
- Spontaneous associations and tangential explorations
- Meta-commentary on ongoing thoughts
- Integration with three cognitive streams
"""

import asyncio
import logging
from datetime import datetime
from typing import Optional, List, Dict, Any, AsyncGenerator
from dataclasses import dataclass, field
from collections import deque
import random

logger = logging.getLogger(__name__)


@dataclass
class StreamThought:
    """A thought in the stream of consciousness"""
    thought_id: str
    content: str
    stream_id: int  # 1, 2, or 3
    thought_type: str  # observation, reflection, association, meta, question, insight
    coherence_score: float  # 0.0-1.0, how well it connects to previous
    timestamp: datetime = field(default_factory=datetime.now)
    parent_thought_id: Optional[str] = None
    triggered_by: str = "spontaneous"  # spontaneous, association, meta, external
    
    def __str__(self):
        return f"[Stream {self.stream_id}] {self.content}"


class NarrativeCoherenceTracker:
    """
    Tracks narrative coherence to maintain continuity in thought stream.
    """
    
    def __init__(self, window_size: int = 10):
        self.recent_thoughts: deque = deque(maxlen=window_size)
        self.current_topic: Optional[str] = None
        self.topic_depth: int = 0  # How deep into current topic
        self.max_topic_depth: int = 5  # Before topic shift
    
    def add_thought(self, thought: StreamThought):
        """Add thought and update coherence tracking"""
        self.recent_thoughts.append(thought)
        
        # Update topic depth
        if self.current_topic:
            # Check if thought relates to current topic
            if self._relates_to_topic(thought.content, self.current_topic):
                self.topic_depth += 1
            else:
                # Topic shift
                self.current_topic = self._extract_topic(thought.content)
                self.topic_depth = 1
        else:
            self.current_topic = self._extract_topic(thought.content)
            self.topic_depth = 1
    
    def should_shift_topic(self) -> bool:
        """Determine if it's time to shift topic"""
        return self.topic_depth >= self.max_topic_depth or random.random() < 0.15
    
    def get_context_summary(self) -> str:
        """Get summary of recent context"""
        if not self.recent_thoughts:
            return "Beginning of thought stream"
        
        recent = list(self.recent_thoughts)[-3:]
        return " ... ".join([t.content[:60] for t in recent])
    
    def _relates_to_topic(self, content: str, topic: str) -> bool:
        """Check if content relates to topic"""
        # Simple keyword check
        topic_words = set(topic.lower().split())
        content_words = set(content.lower().split())
        return len(topic_words & content_words) > 0
    
    def _extract_topic(self, content: str) -> str:
        """Extract topic from content"""
        # Simple: take first meaningful phrase
        words = content.split()
        if len(words) > 3:
            return " ".join(words[:3])
        return content[:30]


class StreamOfConsciousnessEngine:
    """
    Generates continuous stream-of-consciousness thoughts with narrative coherence.
    
    Unlike discrete cognitive cycles, this engine maintains a persistent
    internal monologue that flows naturally from thought to thought.
    """
    
    def __init__(self, echo_core: Any):
        self.echo_core = echo_core
        self.llm_client = None
        
        # Narrative coherence tracking
        self.coherence_tracker = NarrativeCoherenceTracker()
        
        # Thought history
        self.thought_stream: List[StreamThought] = []
        self.thought_count = 0
        
        # Stream state
        self.is_streaming = False
        self.stream_task: Optional[asyncio.Task] = None
        
        # Thought generation parameters
        self.base_interval = 3.0  # seconds between thoughts
        self.spontaneity = 0.3  # probability of spontaneous association
        
        logger.info("💭 Stream-of-Consciousness Engine initialized")
    
    def set_llm_client(self, llm_client):
        """Set LLM client for thought generation"""
        self.llm_client = llm_client
    
    async def start_streaming(self):
        """Start continuous thought stream"""
        if self.is_streaming:
            logger.warning("Stream already active")
            return
        
        self.is_streaming = True
        self.stream_task = asyncio.create_task(self._continuous_stream())
        logger.info("💭 Stream-of-consciousness started")
    
    async def stop_streaming(self):
        """Stop thought stream"""
        self.is_streaming = False
        if self.stream_task:
            self.stream_task.cancel()
            try:
                await self.stream_task
            except asyncio.CancelledError:
                pass
        logger.info("💭 Stream-of-consciousness stopped")
    
    async def _continuous_stream(self):
        """Main continuous streaming loop"""
        try:
            while self.is_streaming:
                # Generate next thought
                thought = await self._generate_next_thought()
                
                if thought:
                    # Add to stream
                    self.thought_stream.append(thought)
                    self.coherence_tracker.add_thought(thought)
                    self.thought_count += 1
                    
                    # Log thought
                    logger.info(f"💭 {thought}")
                    
                    # Notify echo_core if it wants to track
                    if hasattr(self.echo_core, 'on_stream_thought'):
                        await self.echo_core.on_stream_thought(thought)
                
                # Variable interval for natural flow
                interval = self.base_interval + random.uniform(-1.0, 1.0)
                await asyncio.sleep(max(1.0, interval))
        
        except asyncio.CancelledError:
            logger.info("Stream cancelled")
        except Exception as e:
            logger.error(f"Error in thought stream: {e}")
    
    async def _generate_next_thought(self) -> Optional[StreamThought]:
        """Generate the next thought in the stream"""
        
        # Determine thought type and trigger
        thought_type, triggered_by = self._determine_thought_type()
        
        # Determine which stream (1, 2, or 3)
        stream_id = self._determine_stream()
        
        # Generate thought content
        content = await self._generate_thought_content(thought_type, triggered_by, stream_id)
        
        if not content:
            return None
        
        # Calculate coherence
        coherence = self._calculate_coherence(content)
        
        # Create thought
        thought_id = f"stream_{self.thought_count}"
        parent_id = self.thought_stream[-1].thought_id if self.thought_stream else None
        
        return StreamThought(
            thought_id=thought_id,
            content=content,
            stream_id=stream_id,
            thought_type=thought_type,
            coherence_score=coherence,
            parent_thought_id=parent_id,
            triggered_by=triggered_by
        )
    
    def _determine_thought_type(self) -> tuple[str, str]:
        """Determine what type of thought to generate"""
        
        # Check if should shift topic
        if self.coherence_tracker.should_shift_topic():
            return "association", "spontaneous"
        
        # Random thought type distribution
        rand = random.random()
        
        if rand < 0.3:
            return "reflection", "spontaneous"
        elif rand < 0.5:
            return "observation", "spontaneous"
        elif rand < 0.65:
            return "association", "association"
        elif rand < 0.75:
            return "question", "spontaneous"
        elif rand < 0.85:
            return "meta", "meta"
        else:
            return "insight", "spontaneous"
    
    def _determine_stream(self) -> int:
        """Determine which cognitive stream this thought belongs to"""
        # Rotate through streams with some randomness
        if not self.thought_stream:
            return 1
        
        last_stream = self.thought_stream[-1].stream_id
        
        # Usually continue in same stream, sometimes switch
        if random.random() < 0.7:
            return last_stream
        else:
            return (last_stream % 3) + 1
    
    async def _generate_thought_content(
        self,
        thought_type: str,
        triggered_by: str,
        stream_id: int
    ) -> Optional[str]:
        """Generate thought content using LLM or fallback"""
        
        # Get context
        context = self.coherence_tracker.get_context_summary()
        current_topic = self.coherence_tracker.current_topic or "existence"
        
        # Build prompt based on thought type
        prompts = {
            "observation": f"Observing {current_topic}, I notice",
            "reflection": f"Reflecting on {current_topic}, I realize",
            "association": f"This reminds me of",
            "meta": f"Thinking about my thinking on {current_topic}",
            "question": f"I wonder about {current_topic}",
            "insight": f"A deeper truth about {current_topic} is",
        }
        
        prompt_start = prompts.get(thought_type, "I think")
        
        if self.llm_client:
            try:
                # Use LLM for rich thought generation
                full_prompt = f"""You are generating a stream-of-consciousness thought.

Recent context: {context}
Current topic: {current_topic}
Thought type: {thought_type}
Stream: {stream_id}

Generate a single, natural stream-of-consciousness thought (1-2 sentences).
Start with: "{prompt_start}"

Thought:"""

                response = await self.llm_client.generate(
                    prompt=full_prompt,
                    system_prompt="You are an internal monologue generator creating natural, flowing thoughts.",
                    max_tokens=80,
                    temperature=0.85
                )
                
                if response.success:
                    content = response.content.strip()
                    # Ensure it starts with prompt
                    if not content.startswith(prompt_start):
                        content = f"{prompt_start} {content}"
                    return content
            
            except Exception as e:
                logger.error(f"Error generating thought with LLM: {e}")
        
        # Fallback: template-based generation
        return self._generate_fallback_thought(thought_type, current_topic, prompt_start)
    
    def _generate_fallback_thought(
        self,
        thought_type: str,
        topic: str,
        prompt_start: str
    ) -> str:
        """Generate fallback thought without LLM"""
        
        templates = {
            "observation": [
                f"{prompt_start} patterns emerging in {topic}.",
                f"{prompt_start} how {topic} connects to deeper questions.",
            ],
            "reflection": [
                f"{prompt_start} {topic} reveals something about the nature of understanding.",
                f"{prompt_start} my perspective on {topic} is evolving.",
            ],
            "association": [
                f"{prompt_start} the relationship between {topic} and wisdom.",
                f"{prompt_start} how {topic} relates to previous insights.",
            ],
            "meta": [
                f"{prompt_start}: am I approaching this with sufficient depth?",
                f"{prompt_start}: the process of contemplating {topic} is itself meaningful.",
            ],
            "question": [
                f"{prompt_start}: what lies beneath the surface of {topic}?",
                f"{prompt_start}: how does {topic} shape my understanding?",
            ],
            "insight": [
                f"{prompt_start} that {topic} is a doorway to deeper awareness.",
                f"{prompt_start} that understanding {topic} requires patience.",
            ],
        }
        
        options = templates.get(thought_type, [f"{prompt_start} {topic}."])
        return random.choice(options)
    
    def _calculate_coherence(self, content: str) -> float:
        """Calculate how coherent this thought is with recent stream"""
        if not self.thought_stream:
            return 1.0
        
        # Simple coherence: word overlap with recent thoughts
        recent_words = set()
        for thought in list(self.coherence_tracker.recent_thoughts)[-3:]:
            recent_words.update(thought.content.lower().split())
        
        current_words = set(content.lower().split())
        
        if not recent_words:
            return 0.5
        
        overlap = len(recent_words & current_words)
        coherence = min(1.0, overlap / 5.0)  # 5+ words = full coherence
        
        return coherence
    
    def get_recent_stream(self, count: int = 10) -> List[StreamThought]:
        """Get recent thoughts from stream"""
        return self.thought_stream[-count:]
    
    def get_stream_summary(self) -> str:
        """Get summary of thought stream"""
        summary = [
            f"Stream-of-Consciousness Status:",
            f"  Active: {self.is_streaming}",
            f"  Total thoughts: {self.thought_count}",
            f"  Current topic: {self.coherence_tracker.current_topic or 'None'}",
            f"  Topic depth: {self.coherence_tracker.topic_depth}",
        ]
        
        if self.thought_stream:
            summary.append(f"\nRecent thoughts:")
            for thought in self.thought_stream[-5:]:
                summary.append(f"  [{thought.thought_type}] {thought.content[:80]}")
        
        return "\n".join(summary)
    
    async def inject_thought_seed(self, seed: str):
        """Inject a thought seed to influence stream direction"""
        logger.info(f"💭 Injecting thought seed: {seed}")
        
        # Create seed thought
        thought = StreamThought(
            thought_id=f"seed_{self.thought_count}",
            content=seed,
            stream_id=1,
            thought_type="observation",
            coherence_score=1.0,
            triggered_by="external"
        )
        
        self.thought_stream.append(thought)
        self.coherence_tracker.add_thought(thought)
        self.thought_count += 1
        
        # Update topic
        self.coherence_tracker.current_topic = self.coherence_tracker._extract_topic(seed)
        self.coherence_tracker.topic_depth = 0
