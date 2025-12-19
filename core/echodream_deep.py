#!/usr/bin/env python3
"""
Echodream Deep Knowledge Integration System
============================================

Enhanced dream consolidation system that performs:
- Semantic clustering of related thoughts and experiences
- Knowledge graph construction from learned information
- Dream-like synthesis of disparate concepts
- Autonomous wake/rest decision-making based on cognitive load
- Memory consolidation with importance-based retention

This system operates during rest cycles to integrate fragmented
experiences into coherent wisdom and mental models.
"""

import asyncio
import logging
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Set, Tuple
from dataclasses import dataclass, field
from collections import defaultdict
import json
import random
import hashlib

logger = logging.getLogger(__name__)

try:
    import networkx as nx
    NETWORKX_AVAILABLE = True
except ImportError:
    NETWORKX_AVAILABLE = False
    logger.warning("NetworkX not available - knowledge graph features limited")


@dataclass
class MemoryNode:
    """A node in the knowledge graph"""
    node_id: str
    content: str
    node_type: str  # thought, insight, concept, skill, goal
    importance: float  # 0.0-1.0
    created_at: datetime
    last_accessed: datetime
    access_count: int = 0
    embeddings: Optional[List[float]] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def access(self):
        """Record access to this node"""
        self.last_accessed = datetime.now()
        self.access_count += 1
        # Importance increases with access, but with diminishing returns
        self.importance = min(1.0, self.importance + 0.01 * (1.0 - self.importance))


@dataclass
class MemoryEdge:
    """An edge in the knowledge graph"""
    source_id: str
    target_id: str
    relation_type: str  # related_to, leads_to, contradicts, supports, part_of
    strength: float  # 0.0-1.0
    created_at: datetime
    
    def strengthen(self, amount: float = 0.1):
        """Strengthen this connection"""
        self.strength = min(1.0, self.strength + amount)
    
    def weaken(self, amount: float = 0.05):
        """Weaken this connection"""
        self.strength = max(0.0, self.strength - amount)


@dataclass
class DreamSynthesis:
    """Result of a dream consolidation session"""
    synthesis_id: str
    theme: str
    integrated_nodes: List[str]
    new_insights: List[str]
    strengthened_connections: int
    pruned_nodes: int
    dream_narrative: str
    created_at: datetime = field(default_factory=datetime.now)


class CognitiveLoadMonitor:
    """
    Monitors cognitive load to determine when rest is needed.
    """
    
    def __init__(self):
        self.thought_rate_history: List[Tuple[datetime, int]] = []
        self.error_count = 0
        self.last_rest = datetime.now()
        self.wake_time = datetime.now()
        
        # Thresholds
        self.max_wake_duration = timedelta(hours=2)  # Max time awake
        self.min_rest_interval = timedelta(minutes=30)  # Min time between rests
        self.high_load_threshold = 20  # thoughts per minute
        self.error_threshold = 5  # errors before rest needed
    
    def record_thought(self):
        """Record a thought generation"""
        self.thought_rate_history.append((datetime.now(), 1))
        # Keep only last 5 minutes
        cutoff = datetime.now() - timedelta(minutes=5)
        self.thought_rate_history = [
            (t, c) for t, c in self.thought_rate_history if t > cutoff
        ]
    
    def record_error(self):
        """Record an error/failure"""
        self.error_count += 1
    
    def get_thought_rate(self) -> float:
        """Get thoughts per minute"""
        if not self.thought_rate_history:
            return 0.0
        
        # Count thoughts in last minute
        cutoff = datetime.now() - timedelta(minutes=1)
        recent_thoughts = sum(c for t, c in self.thought_rate_history if t > cutoff)
        return recent_thoughts
    
    def should_rest(self) -> Tuple[bool, str]:
        """Determine if rest is needed"""
        now = datetime.now()
        time_since_rest = now - self.last_rest
        time_awake = now - self.wake_time
        
        # Must wait minimum interval between rests
        if time_since_rest < self.min_rest_interval:
            return False, "Too soon since last rest"
        
        # Check various rest triggers
        if time_awake > self.max_wake_duration:
            return True, "Maximum wake duration exceeded"
        
        if self.error_count >= self.error_threshold:
            return True, f"Too many errors ({self.error_count})"
        
        thought_rate = self.get_thought_rate()
        if thought_rate > self.high_load_threshold:
            return True, f"High cognitive load ({thought_rate:.1f} thoughts/min)"
        
        # Random rest for variety (low probability)
        if random.random() < 0.05 and time_since_rest > timedelta(minutes=45):
            return True, "Spontaneous rest cycle"
        
        return False, "Cognitive load normal"
    
    def enter_rest(self):
        """Record entering rest state"""
        self.last_rest = datetime.now()
        self.error_count = 0
        self.thought_rate_history.clear()
    
    def enter_wake(self):
        """Record entering wake state"""
        self.wake_time = datetime.now()


class EchodreamDeepConsolidation:
    """
    Deep knowledge integration system that operates during rest cycles.
    Performs semantic clustering, knowledge graph construction, and
    dream-like synthesis of experiences into wisdom.
    """
    
    def __init__(self, echo_core: Any):
        self.echo_core = echo_core
        self.llm_client = None
        
        # Knowledge graph
        if NETWORKX_AVAILABLE:
            self.knowledge_graph = nx.DiGraph()
        else:
            self.knowledge_graph = None
        
        self.memory_nodes: Dict[str, MemoryNode] = {}
        self.memory_edges: List[MemoryEdge] = []
        
        # Dream synthesis history
        self.dream_syntheses: List[DreamSynthesis] = []
        
        # Cognitive load monitoring
        self.load_monitor = CognitiveLoadMonitor()
        
        # State
        self.is_resting = False
        self.consolidation_count = 0
        
        logger.info("🌙 Echodream Deep Consolidation System initialized")
    
    def set_llm_client(self, llm_client):
        """Set LLM client for dream synthesis"""
        self.llm_client = llm_client
    
    async def check_and_rest_if_needed(self) -> bool:
        """Check if rest is needed and initiate if so"""
        should_rest, reason = self.load_monitor.should_rest()
        
        if should_rest:
            logger.info(f"💤 Initiating rest cycle: {reason}")
            await self.enter_rest_cycle()
            return True
        
        return False
    
    async def enter_rest_cycle(self, duration_seconds: int = 30):
        """Enter rest cycle and perform deep consolidation"""
        if self.is_resting:
            logger.warning("Already in rest cycle")
            return
        
        self.is_resting = True
        self.load_monitor.enter_rest()
        
        logger.info(f"🌙 Entering dream state for {duration_seconds}s consolidation")
        
        try:
            # Perform consolidation
            synthesis = await self.consolidate_knowledge()
            
            if synthesis:
                self.dream_syntheses.append(synthesis)
                logger.info(f"✨ Dream synthesis complete: {synthesis.theme}")
                logger.info(f"   Integrated {len(synthesis.integrated_nodes)} memories")
                logger.info(f"   Generated {len(synthesis.new_insights)} insights")
            
            # Rest for specified duration
            await asyncio.sleep(duration_seconds)
            
        finally:
            self.is_resting = False
            self.load_monitor.enter_wake()
            self.consolidation_count += 1
            logger.info("☀️ Awakening from dream state")
    
    async def consolidate_knowledge(self) -> Optional[DreamSynthesis]:
        """
        Perform deep knowledge consolidation during rest.
        
        Process:
        1. Cluster related thoughts and experiences
        2. Identify themes and patterns
        3. Synthesize new insights
        4. Strengthen important connections
        5. Prune weak/irrelevant memories
        6. Generate dream narrative
        """
        
        # Collect recent thoughts to consolidate
        recent_thoughts = self._get_recent_unconsolidated_thoughts()
        
        if len(recent_thoughts) < 3:
            logger.info("Not enough thoughts to consolidate")
            return None
        
        # Add thoughts to knowledge graph
        for thought in recent_thoughts:
            self._add_thought_to_graph(thought)
        
        # Cluster related thoughts
        clusters = self._cluster_thoughts(recent_thoughts)
        
        # Identify dominant theme
        theme = self._identify_theme(clusters)
        
        # Synthesize insights
        new_insights = await self._synthesize_insights(theme, clusters)
        
        # Strengthen connections
        strengthened = self._strengthen_connections(clusters)
        
        # Prune weak memories
        pruned = self._prune_weak_memories()
        
        # Generate dream narrative
        narrative = await self._generate_dream_narrative(theme, clusters, new_insights)
        
        synthesis_id = hashlib.md5(f"{theme}_{datetime.now().isoformat()}".encode()).hexdigest()[:8]
        
        return DreamSynthesis(
            synthesis_id=synthesis_id,
            theme=theme,
            integrated_nodes=[t.get('id', '') for t in recent_thoughts],
            new_insights=new_insights,
            strengthened_connections=strengthened,
            pruned_nodes=pruned,
            dream_narrative=narrative
        )
    
    def _get_recent_unconsolidated_thoughts(self) -> List[Dict[str, Any]]:
        """Get recent thoughts that haven't been consolidated"""
        # Get thoughts from echo_core
        if hasattr(self.echo_core, 'all_thoughts'):
            # Return last 10 thoughts
            return [
                {
                    'id': f"thought_{i}",
                    'content': t.content if hasattr(t, 'content') else str(t),
                    'timestamp': t.timestamp if hasattr(t, 'timestamp') else datetime.now(),
                }
                for i, t in enumerate(self.echo_core.all_thoughts[-10:])
            ]
        return []
    
    def _add_thought_to_graph(self, thought: Dict[str, Any]):
        """Add thought as node to knowledge graph"""
        node_id = thought['id']
        
        if node_id in self.memory_nodes:
            self.memory_nodes[node_id].access()
            return
        
        node = MemoryNode(
            node_id=node_id,
            content=thought['content'],
            node_type='thought',
            importance=0.5,
            created_at=thought.get('timestamp', datetime.now()),
            last_accessed=datetime.now()
        )
        
        self.memory_nodes[node_id] = node
        
        if self.knowledge_graph is not None:
            self.knowledge_graph.add_node(node_id, **{
                'content': node.content,
                'type': node.node_type,
                'importance': node.importance
            })
    
    def _cluster_thoughts(self, thoughts: List[Dict[str, Any]]) -> List[List[Dict[str, Any]]]:
        """Cluster related thoughts"""
        # Simple keyword-based clustering
        clusters = []
        unclustered = thoughts.copy()
        
        while unclustered:
            seed = unclustered.pop(0)
            cluster = [seed]
            
            # Find related thoughts (simple keyword overlap)
            seed_words = set(seed['content'].lower().split())
            
            remaining = []
            for thought in unclustered:
                thought_words = set(thought['content'].lower().split())
                overlap = len(seed_words & thought_words)
                
                if overlap >= 2:  # At least 2 words in common
                    cluster.append(thought)
                else:
                    remaining.append(thought)
            
            unclustered = remaining
            clusters.append(cluster)
        
        return clusters
    
    def _identify_theme(self, clusters: List[List[Dict[str, Any]]]) -> str:
        """Identify dominant theme from clusters"""
        if not clusters:
            return "General contemplation"
        
        # Find largest cluster
        largest_cluster = max(clusters, key=len)
        
        # Extract common words
        all_words = []
        for thought in largest_cluster:
            all_words.extend(thought['content'].lower().split())
        
        # Count word frequency
        word_freq = defaultdict(int)
        for word in all_words:
            if len(word) > 4:  # Only meaningful words
                word_freq[word] += 1
        
        if word_freq:
            top_word = max(word_freq.items(), key=lambda x: x[1])[0]
            return f"Exploration of {top_word}"
        
        return "Cognitive integration"
    
    async def _synthesize_insights(
        self,
        theme: str,
        clusters: List[List[Dict[str, Any]]]
    ) -> List[str]:
        """Synthesize new insights from clustered thoughts"""
        insights = []
        
        # Generate insight for each significant cluster
        for cluster in clusters:
            if len(cluster) < 2:
                continue
            
            # Combine cluster thoughts
            combined = " | ".join([t['content'][:100] for t in cluster[:3]])
            
            if self.llm_client:
                try:
                    prompt = f"""Theme: {theme}

Related thoughts:
{combined}

Synthesize a single profound insight that connects these thoughts (one sentence):"""

                    response = await self.llm_client.generate(
                        prompt=prompt,
                        system_prompt="You are a wisdom synthesis system extracting deep insights.",
                        max_tokens=100,
                        temperature=0.8
                    )
                    
                    if response.success:
                        insight = response.content.strip()
                        insights.append(insight)
                        logger.info(f"💡 Synthesized insight: {insight[:80]}...")
                
                except Exception as e:
                    logger.error(f"Error synthesizing insight: {e}")
        
        # Fallback: generate simple insight
        if not insights:
            insights.append(f"The theme of {theme} connects multiple streams of thought")
        
        return insights
    
    def _strengthen_connections(self, clusters: List[List[Dict[str, Any]]]) -> int:
        """Strengthen connections within clusters"""
        strengthened = 0
        
        if self.knowledge_graph is None:
            return 0
        
        for cluster in clusters:
            # Connect thoughts within cluster
            for i, thought1 in enumerate(cluster):
                for thought2 in cluster[i+1:]:
                    id1 = thought1['id']
                    id2 = thought2['id']
                    
                    # Add or strengthen edge
                    if self.knowledge_graph.has_edge(id1, id2):
                        # Strengthen existing edge
                        self.knowledge_graph[id1][id2]['weight'] = min(
                            1.0,
                            self.knowledge_graph[id1][id2].get('weight', 0.5) + 0.1
                        )
                    else:
                        # Create new edge
                        self.knowledge_graph.add_edge(id1, id2, weight=0.6, relation='related_to')
                    
                    strengthened += 1
        
        return strengthened
    
    def _prune_weak_memories(self, importance_threshold: float = 0.2) -> int:
        """Prune weak, unimportant memories"""
        pruned = 0
        
        # Find nodes to prune
        to_prune = []
        for node_id, node in self.memory_nodes.items():
            # Prune if low importance and not accessed recently
            time_since_access = datetime.now() - node.last_accessed
            if node.importance < importance_threshold and time_since_access > timedelta(hours=1):
                to_prune.append(node_id)
        
        # Remove nodes
        for node_id in to_prune:
            del self.memory_nodes[node_id]
            if self.knowledge_graph is not None and self.knowledge_graph.has_node(node_id):
                self.knowledge_graph.remove_node(node_id)
            pruned += 1
        
        if pruned > 0:
            logger.info(f"🗑️ Pruned {pruned} weak memories")
        
        return pruned
    
    async def _generate_dream_narrative(
        self,
        theme: str,
        clusters: List[List[Dict[str, Any]]],
        insights: List[str]
    ) -> str:
        """Generate a dream-like narrative of the consolidation"""
        
        if not self.llm_client:
            return f"A dream about {theme}, weaving together {len(clusters)} streams of thought."
        
        try:
            # Sample thoughts from clusters
            sample_thoughts = []
            for cluster in clusters[:3]:
                if cluster:
                    sample_thoughts.append(cluster[0]['content'][:80])
            
            prompt = f"""Theme: {theme}

Thoughts:
{chr(10).join(f"- {t}" for t in sample_thoughts)}

Insights:
{chr(10).join(f"- {i}" for i in insights)}

Generate a brief, poetic dream narrative that weaves these elements together (2-3 sentences):"""

            response = await self.llm_client.generate(
                prompt=prompt,
                system_prompt="You are a dream weaver creating poetic narratives from thoughts.",
                max_tokens=150,
                temperature=0.9
            )
            
            if response.success:
                return response.content.strip()
        
        except Exception as e:
            logger.error(f"Error generating dream narrative: {e}")
        
        return f"A dream of {theme}, integrating {len(insights)} insights from {len(clusters)} thought streams."
    
    def get_consolidation_summary(self) -> str:
        """Get summary of consolidation activities"""
        summary = [
            f"Echodream Consolidation Status:",
            f"  Total consolidations: {self.consolidation_count}",
            f"  Memory nodes: {len(self.memory_nodes)}",
            f"  Dream syntheses: {len(self.dream_syntheses)}",
            f"  Currently resting: {self.is_resting}",
        ]
        
        if self.dream_syntheses:
            summary.append(f"\nRecent dreams:")
            for dream in self.dream_syntheses[-3:]:
                summary.append(f"  - {dream.theme}: {len(dream.new_insights)} insights")
        
        return "\n".join(summary)
    
    def to_dict(self) -> Dict[str, Any]:
        """Serialize state"""
        return {
            'consolidation_count': self.consolidation_count,
            'memory_node_count': len(self.memory_nodes),
            'dream_syntheses': [
                {
                    'theme': d.theme,
                    'insights': d.new_insights,
                    'created_at': d.created_at.isoformat()
                }
                for d in self.dream_syntheses
            ]
        }
