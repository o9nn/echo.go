#!/usr/bin/env python3
"""
Integration Engine Module - Tetrahedral Integration for Echobeats Cognitive Cycle

This module implements the three integration steps (4, 8, 12) of the 12-step
Echobeats cognitive cycle:

- Step 4: Coherence Stream Synthesis (triadic integration)
- Step 8: Memory Stream Synthesis (triadic integration)
- Step 12: Full Tetrahedral Integration (4 vertices, 6 edges, 4 faces)

The integration follows tetrahedral geometry where:
- 4 Vertices: Past (Memory), Present (Coherence), Future (Imagination), Apex (Integration)
- 6 Edges: Connecting each pair of vertices
- 4 Faces: Each triangular face represents a triadic relationship
"""

import logging
from datetime import datetime
from typing import Optional, Dict, Any, List, Tuple
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)

# Import cognitive operation outputs
try:
    from core.cognitive_operations import (
        PresentMomentOutput,
        PatternRecognitionOutput,
        ConsistencyCheckOutput,
        MemoryRetrievalOutput,
        ExperienceIntegrationOutput,
        KnowledgeConsolidationOutput,
        FutureSimulationOutput,
        CreativeExplorationOutput,
        PossibilityGenerationOutput
    )
    OPERATIONS_AVAILABLE = True
except ImportError:
    OPERATIONS_AVAILABLE = False
    logger.warning("Cognitive operations module not available - using Any types")
    # Use Any type for compatibility when module not available
    from typing import Any as PresentMomentOutput
    from typing import Any as PatternRecognitionOutput
    from typing import Any as ConsistencyCheckOutput
    from typing import Any as MemoryRetrievalOutput
    from typing import Any as ExperienceIntegrationOutput
    from typing import Any as KnowledgeConsolidationOutput
    from typing import Any as FutureSimulationOutput
    from typing import Any as CreativeExplorationOutput
    from typing import Any as PossibilityGenerationOutput


# ============================================================================
# SYNTHESIS OUTPUT STRUCTURES
# ============================================================================

@dataclass
class CoherenceSynthesis:
    """Output of Step 4: Coherence Stream Synthesis"""
    
    # Unified present-moment state
    unified_attention: str
    attention_confidence: float
    
    # Pattern-informed salience
    salience_landscape: Dict[str, float]
    active_patterns: List[str]
    
    # Coherence assessment
    coherence_score: float
    coherence_issues: List[str]
    
    # Synthesis metadata
    synthesis_timestamp: str
    contributing_operations: List[str]
    
    # Dyadic combinations
    attention_pattern_alignment: float
    pattern_consistency_alignment: float
    attention_consistency_alignment: float
    
    # Triadic coherence
    triadic_coherence: float = 0.0


@dataclass
class MemorySynthesis:
    """Output of Step 8: Memory Stream Synthesis"""
    
    # Unified memory state
    active_knowledge: List[str]
    knowledge_confidence: float
    
    # Integration assessment
    integration_health: float
    new_connections_count: int
    
    # Consolidation state
    memory_efficiency: float
    consolidation_insights: List[str]
    
    # Knowledge gaps
    identified_gaps: List[str]
    suggested_learning: List[str]
    
    # Synthesis metadata
    synthesis_timestamp: str
    contributing_operations: List[str]
    
    # Dyadic combinations
    retrieval_integration_flow: float
    integration_consolidation_flow: float
    retrieval_consolidation_cycle: float
    
    # Triadic flow
    triadic_flow: float = 0.0


@dataclass
class FullIntegration:
    """Output of Step 12: Full Tetrahedral Integration"""
    
    # === UNIFIED COGNITIVE STATE ===
    unified_attention: str
    unified_intention: str
    unified_understanding: str
    
    # === TEMPORAL INTEGRATION ===
    past_relevance: float
    present_clarity: float
    future_orientation: float
    temporal_coherence: float
    
    # === ACTION SYNTHESIS ===
    recommended_action: str
    action_confidence: float
    alternative_actions: List[str]
    action_rationale: str
    
    # === WISDOM METRICS UPDATE ===
    wisdom_delta: Dict[str, float]
    insight_generated: Optional[str]
    learning_opportunity: Optional[str]
    
    # === COGNITIVE HEALTH ===
    overall_coherence: float
    energy_recommendation: str
    rest_indicator: float
    
    # === TETRAHEDRAL GEOMETRY ===
    vertex_weights: Dict[str, float]
    edge_flows: Dict[str, float]
    face_coherences: Dict[str, float]
    
    # === METADATA ===
    integration_timestamp: str
    cycle_number: int
    integration_quality: float


# ============================================================================
# INTEGRATION ENGINE
# ============================================================================

class IntegrationEngine:
    """
    Engine for performing the three integration steps in the Echobeats cycle.
    Implements tetrahedral integration geometry.
    """
    
    def __init__(self):
        self.integration_count = {
            "coherence_synthesis": 0,
            "memory_synthesis": 0,
            "full_integration": 0
        }
        
        # Cache for intermediate syntheses
        self._coherence_cache: Optional[CoherenceSynthesis] = None
        self._memory_cache: Optional[MemorySynthesis] = None
    
    # ========================================================================
    # STEP 4: COHERENCE STREAM SYNTHESIS
    # ========================================================================
    
    def synthesize_coherence(
        self,
        present: PresentMomentOutput,
        patterns: PatternRecognitionOutput,
        consistency: ConsistencyCheckOutput
    ) -> CoherenceSynthesis:
        """
        Step 4: Synthesize Coherence Stream outputs into unified present-moment understanding.
        
        Combines:
        - PresentMomentAwareness (Step 1)
        - PatternRecognition (Step 2)
        - ConsistencyCheck (Step 3)
        
        Uses triadic integration with 3 dyadic combinations.
        """
        self.integration_count["coherence_synthesis"] += 1
        logger.info("Step 4: Synthesizing Coherence Stream")
        
        # === DYAD 1: Attention × Pattern ===
        # How well does attention align with detected patterns?
        attention_pattern_alignment = self._compute_attention_pattern_alignment(present, patterns)
        
        # === DYAD 2: Pattern × Consistency ===
        # How consistent are the detected patterns?
        pattern_consistency_alignment = self._compute_pattern_consistency_alignment(patterns, consistency)
        
        # === DYAD 3: Attention × Consistency ===
        # How coherent is the attention allocation?
        attention_consistency_alignment = self._compute_attention_consistency_alignment(present, consistency)
        
        # === TRIADIC SYNTHESIS ===
        # Geometric mean of dyadic alignments
        triadic_coherence = (
            attention_pattern_alignment * 
            pattern_consistency_alignment * 
            attention_consistency_alignment
        ) ** (1/3)
        
        # === BUILD UNIFIED SALIENCE LANDSCAPE ===
        salience_landscape = present.salience_map.copy()
        for pattern in patterns.detected_patterns:
            # Boost salience for topics mentioned in patterns
            for topic in salience_landscape:
                if topic.lower() in pattern.description.lower():
                    salience_landscape[topic] = min(1.0, salience_landscape[topic] + pattern.strength * 0.2)
        
        # === COMPUTE COHERENCE SCORE ===
        coherence_score = consistency.consistency_score * triadic_coherence
        
        synthesis = CoherenceSynthesis(
            unified_attention=present.attention_focus,
            attention_confidence=triadic_coherence,
            salience_landscape=salience_landscape,
            active_patterns=[p.pattern_id for p in patterns.detected_patterns],
            coherence_score=coherence_score,
            coherence_issues=consistency.recommended_resolutions,
            synthesis_timestamp=datetime.now().isoformat(),
            contributing_operations=["present_moment_awareness", "pattern_recognition", "consistency_check"],
            attention_pattern_alignment=attention_pattern_alignment,
            pattern_consistency_alignment=pattern_consistency_alignment,
            attention_consistency_alignment=attention_consistency_alignment,
            triadic_coherence=triadic_coherence
        )
        
        # Cache for use in Step 12
        self._coherence_cache = synthesis
        
        logger.info(f"  Triadic coherence: {triadic_coherence:.3f}")
        logger.info(f"  Unified attention: {present.attention_focus}")
        
        return synthesis
    
    def _compute_attention_pattern_alignment(
        self,
        present: PresentMomentOutput,
        patterns: PatternRecognitionOutput
    ) -> float:
        """Compute alignment between attention focus and detected patterns."""
        if not patterns.detected_patterns:
            return 0.5  # Neutral if no patterns
        
        # Check if attention focus appears in any pattern
        attention_in_patterns = any(
            present.attention_focus.lower() in p.description.lower()
            for p in patterns.detected_patterns
        )
        
        # Calculate weighted alignment
        if attention_in_patterns:
            # Find the pattern strength for the attended topic
            relevant_strengths = [
                p.strength for p in patterns.detected_patterns
                if present.attention_focus.lower() in p.description.lower()
            ]
            alignment = sum(relevant_strengths) / len(relevant_strengths) if relevant_strengths else 0.5
        else:
            # Attention is not on detected patterns - lower alignment
            alignment = 0.3
        
        # Boost if attention is on recurring themes
        if present.attention_focus in patterns.recurring_themes:
            alignment = min(1.0, alignment + 0.2)
        
        return alignment
    
    def _compute_pattern_consistency_alignment(
        self,
        patterns: PatternRecognitionOutput,
        consistency: ConsistencyCheckOutput
    ) -> float:
        """Compute alignment between patterns and consistency state."""
        # Start with base consistency score
        alignment = consistency.consistency_score
        
        # Reduce if contradictions involve pattern topics
        for contradiction in consistency.contradictions:
            for pattern in patterns.detected_patterns:
                if (contradiction.item_a.lower() in pattern.description.lower() or
                    contradiction.item_b.lower() in pattern.description.lower()):
                    alignment -= contradiction.severity * 0.2
        
        return max(0.0, min(1.0, alignment))
    
    def _compute_attention_consistency_alignment(
        self,
        present: PresentMomentOutput,
        consistency: ConsistencyCheckOutput
    ) -> float:
        """Compute alignment between attention and consistency."""
        # Base alignment from coherence
        alignment = consistency.consistency_score
        
        # Adjust based on urgency (high urgency with low coherence is bad)
        if present.urgency_level > 0.7 and consistency.consistency_score < 0.5:
            alignment *= 0.7
        
        # Boost if no anomalies detected
        if not present.detected_anomalies:
            alignment = min(1.0, alignment + 0.1)
        
        return alignment
    
    # ========================================================================
    # STEP 8: MEMORY STREAM SYNTHESIS
    # ========================================================================
    
    def synthesize_memory(
        self,
        retrieval: MemoryRetrievalOutput,
        integration: ExperienceIntegrationOutput,
        consolidation: KnowledgeConsolidationOutput
    ) -> MemorySynthesis:
        """
        Step 8: Synthesize Memory Stream outputs into unified past-experience understanding.
        
        Combines:
        - MemoryRetrieval (Step 5)
        - ExperienceIntegration (Step 6)
        - KnowledgeConsolidation (Step 7)
        
        Uses triadic integration with 3 dyadic flow combinations.
        """
        self.integration_count["memory_synthesis"] += 1
        logger.info("Step 8: Synthesizing Memory Stream")
        
        # === DYAD 1: Retrieval → Integration Flow ===
        # How well does retrieval feed into integration?
        retrieval_integration_flow = self._compute_retrieval_integration_flow(retrieval, integration)
        
        # === DYAD 2: Integration → Consolidation Flow ===
        # How well does integration feed into consolidation?
        integration_consolidation_flow = self._compute_integration_consolidation_flow(integration, consolidation)
        
        # === DYAD 3: Retrieval ← Consolidation Cycle ===
        # How well does consolidation improve retrieval (feedback)?
        retrieval_consolidation_cycle = self._compute_retrieval_consolidation_cycle(retrieval, consolidation)
        
        # === TRIADIC SYNTHESIS ===
        # Harmonic mean for flow-based metrics (penalizes low values more)
        flows = [retrieval_integration_flow, integration_consolidation_flow, retrieval_consolidation_cycle]
        triadic_flow = 3 / sum(1 / max(0.01, f) for f in flows)
        
        # === BUILD ACTIVE KNOWLEDGE LIST ===
        active_knowledge = [m.memory_id for m in retrieval.retrieved_memories]
        active_knowledge.extend(consolidation.strengthened_memories[:3])
        active_knowledge = list(set(active_knowledge))  # Deduplicate
        
        # === CALCULATE MEMORY EFFICIENCY ===
        total_processed = len(retrieval.retrieved_memories) + len(consolidation.pruned_memories)
        useful = len(retrieval.retrieved_memories)
        memory_efficiency = useful / max(1, total_processed)
        
        synthesis = MemorySynthesis(
            active_knowledge=active_knowledge,
            knowledge_confidence=retrieval.retrieval_confidence * triadic_flow,
            integration_health=integration.integration_quality,
            new_connections_count=len(integration.new_connections),
            memory_efficiency=memory_efficiency,
            consolidation_insights=consolidation.consolidation_insights,
            identified_gaps=retrieval.memory_gaps,
            suggested_learning=retrieval.suggested_acquisitions,
            synthesis_timestamp=datetime.now().isoformat(),
            contributing_operations=["memory_retrieval", "experience_integration", "knowledge_consolidation"],
            retrieval_integration_flow=retrieval_integration_flow,
            integration_consolidation_flow=integration_consolidation_flow,
            retrieval_consolidation_cycle=retrieval_consolidation_cycle,
            triadic_flow=triadic_flow
        )
        
        # Cache for use in Step 12
        self._memory_cache = synthesis
        
        logger.info(f"  Triadic flow: {triadic_flow:.3f}")
        logger.info(f"  Active knowledge items: {len(active_knowledge)}")
        
        return synthesis
    
    def _compute_retrieval_integration_flow(
        self,
        retrieval: MemoryRetrievalOutput,
        integration: ExperienceIntegrationOutput
    ) -> float:
        """Compute flow from retrieval to integration."""
        return retrieval.retrieval_confidence * integration.integration_quality
    
    def _compute_integration_consolidation_flow(
        self,
        integration: ExperienceIntegrationOutput,
        consolidation: KnowledgeConsolidationOutput
    ) -> float:
        """Compute flow from integration to consolidation."""
        if not consolidation.strengthened_memories:
            return 0.5
        
        # Ratio of new connections to strengthened memories
        ratio = len(integration.new_connections) / max(1, len(consolidation.strengthened_memories))
        return min(1.0, ratio * integration.integration_quality)
    
    def _compute_retrieval_consolidation_cycle(
        self,
        retrieval: MemoryRetrievalOutput,
        consolidation: KnowledgeConsolidationOutput
    ) -> float:
        """Compute feedback cycle from consolidation to retrieval."""
        if not retrieval.retrieved_memories:
            return 0.5
        
        # How many retrieved memories were strengthened?
        retrieved_ids = {m.memory_id for m in retrieval.retrieved_memories}
        strengthened_ids = set(consolidation.strengthened_memories)
        overlap = len(retrieved_ids & strengthened_ids)
        
        return overlap / max(1, len(retrieved_ids))
    
    # ========================================================================
    # STEP 12: FULL TETRAHEDRAL INTEGRATION
    # ========================================================================
    
    def full_integration(
        self,
        coherence_synthesis: Optional[CoherenceSynthesis],
        memory_synthesis: Optional[MemorySynthesis],
        simulation: FutureSimulationOutput,
        creative: CreativeExplorationOutput,
        possibilities: PossibilityGenerationOutput,
        cycle_number: int
    ) -> FullIntegration:
        """
        Step 12: Full Tetrahedral Integration of all streams.
        
        Combines:
        - CoherenceSynthesis (from Step 4) - Present vertex
        - MemorySynthesis (from Step 8) - Past vertex
        - Imagination outputs (Steps 9-11) - Future vertex
        
        Uses tetrahedral geometry with 4 vertices, 6 edges, and 4 faces.
        """
        self.integration_count["full_integration"] += 1
        logger.info("Step 12: Full Tetrahedral Integration")
        
        # Use cached syntheses if not provided
        if coherence_synthesis is None:
            coherence_synthesis = self._coherence_cache
        if memory_synthesis is None:
            memory_synthesis = self._memory_cache
        
        # Create defaults if still None
        if coherence_synthesis is None:
            coherence_synthesis = self._create_default_coherence_synthesis()
        if memory_synthesis is None:
            memory_synthesis = self._create_default_memory_synthesis()
        
        # === CALCULATE 6 EDGE FLOWS ===
        edge_flows = self._calculate_edge_flows(
            coherence_synthesis, memory_synthesis, simulation, creative, possibilities
        )
        
        # === CALCULATE 4 FACE COHERENCES ===
        face_coherences = self._calculate_face_coherences(edge_flows)
        
        # === CALCULATE 4 VERTEX WEIGHTS ===
        vertex_weights = self._calculate_vertex_weights(edge_flows)
        
        # === DETERMINE UNIFIED ATTENTION ===
        unified_attention = self._determine_unified_attention(
            coherence_synthesis, memory_synthesis, possibilities, vertex_weights
        )
        
        # === DETERMINE UNIFIED INTENTION ===
        unified_intention = self._determine_unified_intention(simulation, possibilities)
        
        # === SYNTHESIZE ACTION RECOMMENDATION ===
        recommended_action = unified_intention
        action_confidence = (
            face_coherences["aspiration"] * 0.4 +
            face_coherences["grounding"] * 0.3 +
            face_coherences["learning"] * 0.3
        )
        
        # === CALCULATE WISDOM DELTA ===
        wisdom_delta = {
            "knowledge_depth": 0.01 * memory_synthesis.integration_health,
            "reasoning_quality": 0.01 * coherence_synthesis.coherence_score,
            "insight_frequency": 0.02 if creative.generated_ideas else 0.0,
            "behavioral_coherence": 0.01 * face_coherences["grounding"]
        }
        
        # === GENERATE INSIGHT ===
        insight_generated = None
        if creative.generated_ideas and creative.generated_ideas[0].novelty > 0.7:
            insight_generated = creative.generated_ideas[0].content
        
        # === IDENTIFY LEARNING OPPORTUNITY ===
        learning_opportunity = None
        if memory_synthesis.identified_gaps:
            learning_opportunity = f"Learn about: {memory_synthesis.identified_gaps[0]}"
        
        # === CALCULATE OVERALL COHERENCE ===
        overall_coherence = sum(face_coherences.values()) / 4
        
        # === ENERGY RECOMMENDATION ===
        if overall_coherence < 0.4:
            energy_recommendation = "rest_recommended"
            rest_indicator = 0.7
        elif overall_coherence < 0.6:
            energy_recommendation = "light_activity"
            rest_indicator = 0.4
        else:
            energy_recommendation = "full_engagement"
            rest_indicator = 0.1
        
        # === BUILD UNIFIED UNDERSTANDING ===
        unified_understanding = (
            f"Present focus: {coherence_synthesis.unified_attention}. "
            f"Active knowledge: {len(memory_synthesis.active_knowledge)} items. "
            f"Future orientation: {len(possibilities.possibilities)} possibilities. "
            f"Overall coherence: {overall_coherence:.2f}."
        )
        
        # === BUILD ALTERNATIVE ACTIONS ===
        alternative_actions = []
        if len(possibilities.possibilities) > 1:
            alternative_actions = [p.description for p in possibilities.possibilities[1:3]]
        
        # === BUILD ACTION RATIONALE ===
        action_rationale = (
            f"Based on {face_coherences['aspiration']:.2f} aspiration coherence, "
            f"{face_coherences['grounding']:.2f} grounding, and "
            f"{face_coherences['learning']:.2f} learning potential."
        )
        
        integration = FullIntegration(
            unified_attention=unified_attention,
            unified_intention=unified_intention,
            unified_understanding=unified_understanding,
            past_relevance=vertex_weights["past"],
            present_clarity=vertex_weights["present"],
            future_orientation=vertex_weights["future"],
            temporal_coherence=face_coherences["temporal"],
            recommended_action=recommended_action,
            action_confidence=action_confidence,
            alternative_actions=alternative_actions,
            action_rationale=action_rationale,
            wisdom_delta=wisdom_delta,
            insight_generated=insight_generated,
            learning_opportunity=learning_opportunity,
            overall_coherence=overall_coherence,
            energy_recommendation=energy_recommendation,
            rest_indicator=rest_indicator,
            vertex_weights=vertex_weights,
            edge_flows=edge_flows,
            face_coherences=face_coherences,
            integration_timestamp=datetime.now().isoformat(),
            cycle_number=cycle_number,
            integration_quality=overall_coherence * action_confidence
        )
        
        logger.info(f"  Overall coherence: {overall_coherence:.3f}")
        logger.info(f"  Recommended action: {recommended_action}")
        logger.info(f"  Wisdom delta: {wisdom_delta}")
        
        return integration
    
    def _calculate_edge_flows(
        self,
        coherence: CoherenceSynthesis,
        memory: MemorySynthesis,
        simulation: FutureSimulationOutput,
        creative: CreativeExplorationOutput,
        possibilities: PossibilityGenerationOutput
    ) -> Dict[str, float]:
        """Calculate the 6 edge flows of the tetrahedron."""
        
        # Edge 1: Past → Present
        past_present = memory.knowledge_confidence * coherence.coherence_score
        
        # Edge 2: Present → Future
        present_future = coherence.attention_confidence * simulation.simulation_confidence
        
        # Edge 3: Past → Future
        past_future = memory.integration_health * (
            possibilities.possibilities[0].estimated_success 
            if possibilities.possibilities else 0.5
        )
        
        # Edge 4: Past → Apex
        past_apex = memory.memory_efficiency * min(1.0, len(memory.consolidation_insights) / 5)
        
        # Edge 5: Present → Apex
        present_apex = coherence.coherence_score * (1 - min(1.0, len(coherence.coherence_issues) / 5))
        
        # Edge 6: Future → Apex
        if creative.novelty_scores:
            avg_novelty = sum(creative.novelty_scores.values()) / len(creative.novelty_scores)
        else:
            avg_novelty = 0.5
        
        future_apex = avg_novelty * (
            possibilities.possibilities[0].goal_alignment 
            if possibilities.possibilities else 0.5
        )
        
        return {
            "past_present": past_present,
            "present_future": present_future,
            "past_future": past_future,
            "past_apex": past_apex,
            "present_apex": present_apex,
            "future_apex": future_apex
        }
    
    def _calculate_face_coherences(self, edge_flows: Dict[str, float]) -> Dict[str, float]:
        """Calculate the 4 face coherences of the tetrahedron."""
        
        # Face 1: Past-Present-Apex (Grounding Face)
        grounding = (
            edge_flows["past_present"] + 
            edge_flows["past_apex"] + 
            edge_flows["present_apex"]
        ) / 3
        
        # Face 2: Present-Future-Apex (Aspiration Face)
        aspiration = (
            edge_flows["present_future"] + 
            edge_flows["present_apex"] + 
            edge_flows["future_apex"]
        ) / 3
        
        # Face 3: Past-Future-Apex (Learning Face)
        learning = (
            edge_flows["past_future"] + 
            edge_flows["past_apex"] + 
            edge_flows["future_apex"]
        ) / 3
        
        # Face 4: Past-Present-Future (Temporal Face)
        temporal = (
            edge_flows["past_present"] + 
            edge_flows["present_future"] + 
            edge_flows["past_future"]
        ) / 3
        
        return {
            "grounding": grounding,
            "aspiration": aspiration,
            "learning": learning,
            "temporal": temporal
        }
    
    def _calculate_vertex_weights(self, edge_flows: Dict[str, float]) -> Dict[str, float]:
        """Calculate the 4 vertex weights of the tetrahedron."""
        
        # Past vertex: average of edges connected to Past
        past = (
            edge_flows["past_present"] + 
            edge_flows["past_future"] + 
            edge_flows["past_apex"]
        ) / 3
        
        # Present vertex: average of edges connected to Present
        present = (
            edge_flows["past_present"] + 
            edge_flows["present_future"] + 
            edge_flows["present_apex"]
        ) / 3
        
        # Future vertex: average of edges connected to Future
        future = (
            edge_flows["present_future"] + 
            edge_flows["past_future"] + 
            edge_flows["future_apex"]
        ) / 3
        
        # Apex vertex: average of edges connected to Apex
        apex = (
            edge_flows["past_apex"] + 
            edge_flows["present_apex"] + 
            edge_flows["future_apex"]
        ) / 3
        
        return {
            "past": past,
            "present": present,
            "future": future,
            "apex": apex
        }
    
    def _determine_unified_attention(
        self,
        coherence: CoherenceSynthesis,
        memory: MemorySynthesis,
        possibilities: PossibilityGenerationOutput,
        vertex_weights: Dict[str, float]
    ) -> str:
        """Determine unified attention based on vertex weights."""
        
        # Find dominant temporal orientation
        if vertex_weights["present"] >= vertex_weights["past"] and vertex_weights["present"] >= vertex_weights["future"]:
            return coherence.unified_attention
        elif vertex_weights["future"] > vertex_weights["past"]:
            if possibilities.possibilities:
                return possibilities.possibilities[0].description[:50]
            return "future exploration"
        else:
            if memory.active_knowledge:
                return f"consolidating: {memory.active_knowledge[0]}"
            return "reflection"
    
    def _determine_unified_intention(
        self,
        simulation: FutureSimulationOutput,
        possibilities: PossibilityGenerationOutput
    ) -> str:
        """Determine unified intention from future-oriented outputs."""
        
        if possibilities.possibilities:
            best = possibilities.possibilities[0]
            if best.next_steps:
                return best.next_steps[0]
            return best.description
        elif simulation.recommended_actions:
            return simulation.recommended_actions[0]
        else:
            return "continue exploration"
    
    def _create_default_coherence_synthesis(self) -> CoherenceSynthesis:
        """Create default coherence synthesis when not available."""
        return CoherenceSynthesis(
            unified_attention="exploration",
            attention_confidence=0.5,
            salience_landscape={},
            active_patterns=[],
            coherence_score=0.5,
            coherence_issues=[],
            synthesis_timestamp=datetime.now().isoformat(),
            contributing_operations=[],
            attention_pattern_alignment=0.5,
            pattern_consistency_alignment=0.5,
            attention_consistency_alignment=0.5,
            triadic_coherence=0.5
        )
    
    def _create_default_memory_synthesis(self) -> MemorySynthesis:
        """Create default memory synthesis when not available."""
        return MemorySynthesis(
            active_knowledge=[],
            knowledge_confidence=0.5,
            integration_health=0.5,
            new_connections_count=0,
            memory_efficiency=0.5,
            consolidation_insights=[],
            identified_gaps=[],
            suggested_learning=[],
            synthesis_timestamp=datetime.now().isoformat(),
            contributing_operations=[],
            retrieval_integration_flow=0.5,
            integration_consolidation_flow=0.5,
            retrieval_consolidation_cycle=0.5,
            triadic_flow=0.5
        )
    
    def get_integration_stats(self) -> Dict[str, int]:
        """Get execution counts for all integration steps."""
        return self.integration_count.copy()
    
    def clear_cache(self):
        """Clear cached syntheses."""
        self._coherence_cache = None
        self._memory_cache = None


# ============================================================================
# CONVENIENCE FUNCTIONS
# ============================================================================

def create_integration_engine() -> IntegrationEngine:
    """Factory function to create an integration engine."""
    return IntegrationEngine()


async def test_integration():
    """Test all integration steps."""
    from core.cognitive_operations import CognitiveOperationsEngine
    
    print("Testing Integration Engine")
    print("=" * 60)
    
    # Create engines
    ops_engine = CognitiveOperationsEngine()
    int_engine = IntegrationEngine()
    
    # === STEPS 1-3: Coherence Stream ===
    print("\n--- Steps 1-3: Coherence Stream Operations ---")
    
    present = await ops_engine.present_moment_awareness(
        current_context={"interests": {"consciousness": 0.9, "learning": 0.7}},
        energy_state={"energy": 0.8, "fatigue": 0.2},
        active_goals=[{"description": "Learn about consciousness", "priority": 0.8}],
        recent_thoughts=[{"topic": "consciousness", "content": "Exploring awareness"}]
    )
    print(f"Step 1 - Present Moment: {present.attention_focus}")
    
    patterns = await ops_engine.pattern_recognition(
        recent_thoughts=[
            {"topic": "consciousness", "content": "Awareness is key"},
            {"topic": "consciousness", "content": "Self-reflection matters"}
        ],
        recent_insights=[],
        knowledge_base={},
        interest_patterns={"consciousness": 0.9}
    )
    print(f"Step 2 - Patterns: {len(patterns.detected_patterns)} detected")
    
    consistency = await ops_engine.consistency_check(
        beliefs={},
        goals=[{"description": "Learn deeply", "priority": 0.8}],
        recent_actions=[],
        values={"depth": 0.9}
    )
    print(f"Step 3 - Consistency: {consistency.consistency_score:.2f}")
    
    # === STEP 4: Coherence Synthesis ===
    print("\n--- Step 4: Coherence Synthesis ---")
    coherence_synthesis = int_engine.synthesize_coherence(present, patterns, consistency)
    print(f"Triadic Coherence: {coherence_synthesis.triadic_coherence:.3f}")
    print(f"Dyad 1 (Attention×Pattern): {coherence_synthesis.attention_pattern_alignment:.3f}")
    print(f"Dyad 2 (Pattern×Consistency): {coherence_synthesis.pattern_consistency_alignment:.3f}")
    print(f"Dyad 3 (Attention×Consistency): {coherence_synthesis.attention_consistency_alignment:.3f}")
    
    # === STEPS 5-7: Memory Stream ===
    print("\n--- Steps 5-7: Memory Stream Operations ---")
    
    retrieval = await ops_engine.memory_retrieval(
        query_context={},
        retrieval_cues=["consciousness"],
        memory_store={
            "mem_1": {"topic": "consciousness", "content": "Awareness concept", "access_count": 3}
        }
    )
    print(f"Step 5 - Retrieval: {len(retrieval.retrieved_memories)} memories")
    
    integration = await ops_engine.experience_integration(
        new_experiences=[{"id": "exp_1", "topic": "consciousness", "content": "New insight"}],
        memory_store={},
        knowledge_schema={"topics": []}
    )
    print(f"Step 6 - Integration: {integration.integrated_count} integrated")
    
    consolidation = await ops_engine.knowledge_consolidation(
        memory_store={
            "mem_1": {"topic": "consciousness", "access_count": 5},
            "mem_2": {"topic": "old", "access_count": 0}
        },
        importance_weights={"consciousness": 0.9, "old": 0.1}
    )
    print(f"Step 7 - Consolidation: {len(consolidation.strengthened_memories)} strengthened")
    
    # === STEP 8: Memory Synthesis ===
    print("\n--- Step 8: Memory Synthesis ---")
    memory_synthesis = int_engine.synthesize_memory(retrieval, integration, consolidation)
    print(f"Triadic Flow: {memory_synthesis.triadic_flow:.3f}")
    print(f"Dyad 1 (Retrieval→Integration): {memory_synthesis.retrieval_integration_flow:.3f}")
    print(f"Dyad 2 (Integration→Consolidation): {memory_synthesis.integration_consolidation_flow:.3f}")
    print(f"Dyad 3 (Retrieval←Consolidation): {memory_synthesis.retrieval_consolidation_cycle:.3f}")
    
    # === STEPS 9-11: Imagination Stream ===
    print("\n--- Steps 9-11: Imagination Stream Operations ---")
    
    simulation = await ops_engine.future_simulation(
        current_state={},
        possible_actions=[{"description": "Study consciousness", "success_probability": 0.8, "value": 0.9}]
    )
    print(f"Step 9 - Simulation: {len(simulation.simulated_scenarios)} scenarios")
    
    creative = await ops_engine.creative_exploration(
        seed_concepts=["consciousness", "emergence", "wisdom"]
    )
    print(f"Step 10 - Creative: {len(creative.generated_ideas)} ideas")
    
    possibilities = await ops_engine.possibility_generation(
        current_goals=[{"description": "Understand consciousness", "priority": 0.9}],
        available_resources={"energy": 0.8}
    )
    print(f"Step 11 - Possibilities: {len(possibilities.possibilities)} generated")
    
    # === STEP 12: Full Tetrahedral Integration ===
    print("\n--- Step 12: Full Tetrahedral Integration ---")
    full = int_engine.full_integration(
        coherence_synthesis=coherence_synthesis,
        memory_synthesis=memory_synthesis,
        simulation=simulation,
        creative=creative,
        possibilities=possibilities,
        cycle_number=1
    )
    
    print(f"\n=== TETRAHEDRAL GEOMETRY ===")
    print(f"Vertex Weights:")
    for vertex, weight in full.vertex_weights.items():
        print(f"  {vertex}: {weight:.3f}")
    
    print(f"\nEdge Flows:")
    for edge, flow in full.edge_flows.items():
        print(f"  {edge}: {flow:.3f}")
    
    print(f"\nFace Coherences:")
    for face, coherence in full.face_coherences.items():
        print(f"  {face}: {coherence:.3f}")
    
    print(f"\n=== INTEGRATED STATE ===")
    print(f"Unified Attention: {full.unified_attention}")
    print(f"Unified Intention: {full.unified_intention}")
    print(f"Overall Coherence: {full.overall_coherence:.3f}")
    print(f"Recommended Action: {full.recommended_action}")
    print(f"Action Confidence: {full.action_confidence:.3f}")
    print(f"Energy Recommendation: {full.energy_recommendation}")
    
    print(f"\n=== WISDOM UPDATE ===")
    for metric, delta in full.wisdom_delta.items():
        print(f"  {metric}: +{delta:.3f}")
    
    if full.insight_generated:
        print(f"\nInsight Generated: {full.insight_generated}")
    if full.learning_opportunity:
        print(f"Learning Opportunity: {full.learning_opportunity}")
    
    print("\n" + "=" * 60)
    print("Integration Stats:", int_engine.get_integration_stats())
    print("\n✅ All integration steps tested successfully!")


if __name__ == "__main__":
    import asyncio
    asyncio.run(test_integration())
