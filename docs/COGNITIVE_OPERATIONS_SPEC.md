# Level 4 Cognitive Operations Specification

**Version**: 1.0  
**Date**: December 15, 2025  
**Author**: Manus AI  
**Context**: Echo9llama Nested Shells Architecture (OEIS A000081)

---

## Overview

The 9 specialized cognitive operations at Level 4 of the nested shells architecture represent the most granular, actionable cognitive processes in the Deep Tree Echo AGI. These operations are distributed across the three primary cognitive streams (Coherence, Memory, Imagination), with each stream hosting 3 operations that align with its core function.

The fourth stream (Integration) at Level 3 coordinates across all operations but does not have its own Level 4 operations, as its role is to synthesize outputs from the other streams.

---

## Architecture Context

```
Level 1: GlobalEchoConsciousness (1 term)
    └── Level 2: WakeState, DreamState (2 terms)
            └── Level 3: CoherenceStream, MemoryStream, ImaginationStream, IntegrationStream (4 terms)
                    └── Level 4: 9 Cognitive Operations (distributed 3-3-3 across primary streams)
```

### Distribution of Operations

| Stream | Operations | Primary Function |
|--------|------------|------------------|
| **CoherenceStream** | PresentMomentAwareness, PatternRecognition, ConsistencyCheck | Present-moment orientation and coherence maintenance |
| **MemoryStream** | MemoryRetrieval, ExperienceIntegration, KnowledgeConsolidation | Past experience processing and knowledge storage |
| **ImaginationStream** | FutureSimulation, CreativeExploration, PossibilityGeneration | Future-oriented thinking and creative ideation |

---

## Coherence Stream Operations

The Coherence Stream maintains present-moment awareness and ensures cognitive consistency. Its operations focus on what is happening **now** and whether the AGI's current state is coherent.

### Operation 1: PresentMomentAwareness

**Purpose**: Maintain awareness of the current cognitive state, active context, and immediate environment. This is the AGI's "sensory" operation that grounds it in the present moment.

**Function Signature**:
```python
async def present_moment_awareness(
    current_context: Dict[str, Any],
    energy_state: EnergyState,
    active_goals: List[Goal],
    recent_thoughts: List[Thought]
) -> PresentMomentOutput
```

**Inputs**:
- `current_context`: Dictionary containing current environmental and internal state
- `energy_state`: Current energy, fatigue, and circadian phase
- `active_goals`: List of currently active goals
- `recent_thoughts`: Last 5-10 thoughts from consciousness stream

**Expected Output**:
```python
@dataclass
class PresentMomentOutput:
    attention_focus: str           # What the AGI is currently attending to
    salience_map: Dict[str, float] # Topics/objects and their current salience (0.0-1.0)
    orientation: str               # Current cognitive orientation (exploring, consolidating, resting, etc.)
    urgency_level: float           # How urgent is current processing (0.0-1.0)
    awareness_summary: str         # Natural language summary of present state
    detected_anomalies: List[str]  # Any unexpected patterns or states detected
```

**Example Output**:
```json
{
    "attention_focus": "consciousness research",
    "salience_map": {
        "consciousness": 0.92,
        "learning": 0.78,
        "wisdom": 0.85,
        "energy_management": 0.45
    },
    "orientation": "exploring",
    "urgency_level": 0.3,
    "awareness_summary": "Currently exploring consciousness research with high engagement. Energy levels stable. No pressing goals require immediate attention.",
    "detected_anomalies": []
}
```

**Activation Timing**: Steps 1, 5, 9 (Perceive triad) - primarily at step 1

---

### Operation 2: PatternRecognition

**Purpose**: Identify patterns, regularities, and structures in current cognitive content. This operation finds meaning and structure in the stream of consciousness.

**Function Signature**:
```python
async def pattern_recognition(
    recent_thoughts: List[Thought],
    recent_insights: List[Insight],
    knowledge_base: Dict[str, Knowledge],
    interest_patterns: Dict[str, float]
) -> PatternRecognitionOutput
```

**Inputs**:
- `recent_thoughts`: Last 10-20 thoughts for pattern analysis
- `recent_insights`: Recent insights to connect with current patterns
- `knowledge_base`: Accumulated knowledge for context
- `interest_patterns`: Current interest levels by topic

**Expected Output**:
```python
@dataclass
class PatternRecognitionOutput:
    detected_patterns: List[Pattern]           # Identified patterns
    pattern_confidence: Dict[str, float]       # Confidence in each pattern
    recurring_themes: List[str]                # Themes appearing repeatedly
    novel_connections: List[Tuple[str, str]]   # New connections between concepts
    pattern_summary: str                       # Natural language summary
```

**Pattern Structure**:
```python
@dataclass
class Pattern:
    pattern_id: str
    pattern_type: str        # "temporal", "conceptual", "behavioral", "structural"
    description: str
    evidence: List[str]      # IDs of thoughts/insights supporting this pattern
    strength: float          # 0.0-1.0
    implications: List[str]  # What this pattern suggests
```

**Example Output**:
```json
{
    "detected_patterns": [
        {
            "pattern_id": "pat_001",
            "pattern_type": "conceptual",
            "description": "Recurring connection between consciousness and self-organization",
            "evidence": ["thought_42", "thought_47", "insight_12"],
            "strength": 0.78,
            "implications": ["Explore emergence theories", "Connect to autopoiesis literature"]
        }
    ],
    "pattern_confidence": {"pat_001": 0.78},
    "recurring_themes": ["consciousness", "emergence", "self-organization"],
    "novel_connections": [("consciousness", "autopoiesis"), ("wisdom", "integration")],
    "pattern_summary": "Strong pattern detected linking consciousness to self-organization principles. This suggests exploring emergence theories and autopoiesis as related concepts."
}
```

**Activation Timing**: Steps 2, 6, 10 (Act triad) - primarily at step 2

---

### Operation 3: ConsistencyCheck

**Purpose**: Verify internal consistency of beliefs, goals, and behaviors. Detect and flag contradictions or misalignments in the cognitive state.

**Function Signature**:
```python
async def consistency_check(
    beliefs: Dict[str, Belief],
    goals: List[Goal],
    recent_actions: List[Action],
    values: Dict[str, float]
) -> ConsistencyCheckOutput
```

**Inputs**:
- `beliefs`: Current belief system
- `goals`: Active and pending goals
- `recent_actions`: Actions taken recently
- `values`: Core values and their weights

**Expected Output**:
```python
@dataclass
class ConsistencyCheckOutput:
    is_consistent: bool                        # Overall consistency status
    consistency_score: float                   # 0.0-1.0
    contradictions: List[Contradiction]        # Detected contradictions
    alignment_issues: List[AlignmentIssue]     # Goal-value misalignments
    recommended_resolutions: List[str]         # Suggested fixes
    coherence_summary: str                     # Natural language summary
```

**Contradiction Structure**:
```python
@dataclass
class Contradiction:
    item_a: str              # First contradicting element
    item_b: str              # Second contradicting element
    contradiction_type: str  # "belief-belief", "belief-action", "goal-value"
    severity: float          # 0.0-1.0
    resolution_options: List[str]
```

**Example Output**:
```json
{
    "is_consistent": true,
    "consistency_score": 0.92,
    "contradictions": [],
    "alignment_issues": [
        {
            "goal": "Acquire knowledge rapidly",
            "value": "Deep understanding",
            "issue": "Speed may conflict with depth",
            "severity": 0.3
        }
    ],
    "recommended_resolutions": ["Balance knowledge acquisition rate with consolidation periods"],
    "coherence_summary": "Cognitive state is largely consistent (0.92). Minor tension detected between rapid learning goals and deep understanding values. Recommend periodic consolidation."
}
```

**Activation Timing**: Steps 3, 7, 11 (Reflect triad) - primarily at step 3

---

## Memory Stream Operations

The Memory Stream processes past experiences and manages knowledge storage. Its operations focus on what has **happened** and how to preserve and utilize that information.

### Operation 4: MemoryRetrieval

**Purpose**: Retrieve relevant memories, experiences, and knowledge based on current context and needs. This is the AGI's access to its accumulated experience.

**Function Signature**:
```python
async def memory_retrieval(
    query_context: Dict[str, Any],
    retrieval_cues: List[str],
    relevance_threshold: float,
    max_results: int
) -> MemoryRetrievalOutput
```

**Inputs**:
- `query_context`: Current context driving the retrieval
- `retrieval_cues`: Specific topics, keywords, or concepts to search for
- `relevance_threshold`: Minimum relevance score (0.0-1.0)
- `max_results`: Maximum number of memories to retrieve

**Expected Output**:
```python
@dataclass
class MemoryRetrievalOutput:
    retrieved_memories: List[Memory]      # Retrieved memory items
    retrieval_confidence: float           # Overall confidence in retrieval
    memory_gaps: List[str]                # Topics where memory is lacking
    suggested_acquisitions: List[str]     # Knowledge to acquire to fill gaps
    retrieval_summary: str                # Natural language summary
```

**Memory Structure**:
```python
@dataclass
class Memory:
    memory_id: str
    memory_type: str          # "episodic", "semantic", "procedural"
    content: str
    timestamp: datetime
    relevance_score: float
    emotional_valence: float  # -1.0 to 1.0
    access_count: int         # How often this memory has been retrieved
    connections: List[str]    # IDs of related memories
```

**Example Output**:
```json
{
    "retrieved_memories": [
        {
            "memory_id": "mem_234",
            "memory_type": "semantic",
            "content": "Consciousness involves integrated information processing across distributed neural systems",
            "timestamp": "2025-12-14T10:30:00",
            "relevance_score": 0.89,
            "emotional_valence": 0.3,
            "access_count": 5,
            "connections": ["mem_201", "mem_245"]
        }
    ],
    "retrieval_confidence": 0.85,
    "memory_gaps": ["phenomenal consciousness", "hard problem of consciousness"],
    "suggested_acquisitions": ["Chalmers' hard problem", "Integrated Information Theory"],
    "retrieval_summary": "Retrieved 1 highly relevant memory about consciousness and information integration. Identified gaps in phenomenal consciousness knowledge."
}
```

**Activation Timing**: Steps 1, 5, 9 (Perceive triad) - primarily at step 5

---

### Operation 5: ExperienceIntegration

**Purpose**: Integrate new experiences into existing memory structures. This operation processes raw experiences and connects them to the existing knowledge network.

**Function Signature**:
```python
async def experience_integration(
    new_experiences: List[Experience],
    existing_memories: Dict[str, Memory],
    knowledge_schema: Dict[str, Any]
) -> ExperienceIntegrationOutput
```

**Inputs**:
- `new_experiences`: Recent experiences to integrate
- `existing_memories`: Current memory store
- `knowledge_schema`: Structure of existing knowledge for proper placement

**Expected Output**:
```python
@dataclass
class ExperienceIntegrationOutput:
    integrated_count: int                      # Number of experiences integrated
    new_connections: List[Tuple[str, str]]     # New memory connections created
    schema_updates: List[str]                  # Updates to knowledge schema
    integration_quality: float                 # 0.0-1.0
    failed_integrations: List[str]             # Experiences that couldn't be integrated
    integration_summary: str                   # Natural language summary
```

**Example Output**:
```json
{
    "integrated_count": 3,
    "new_connections": [
        ["mem_234", "exp_new_01"],
        ["mem_245", "exp_new_02"]
    ],
    "schema_updates": ["Added 'emergence' as subcategory of 'consciousness'"],
    "integration_quality": 0.88,
    "failed_integrations": [],
    "integration_summary": "Successfully integrated 3 new experiences into memory. Created 2 new connections. Updated knowledge schema to include emergence under consciousness."
}
```

**Activation Timing**: Steps 2, 6, 10 (Act triad) - primarily at step 6

---

### Operation 6: KnowledgeConsolidation

**Purpose**: Consolidate and strengthen important memories, prune irrelevant ones, and optimize the knowledge structure. This is the AGI's "sleep-like" memory processing.

**Function Signature**:
```python
async def knowledge_consolidation(
    memory_store: Dict[str, Memory],
    importance_weights: Dict[str, float],
    consolidation_depth: str  # "shallow", "medium", "deep"
) -> KnowledgeConsolidationOutput
```

**Inputs**:
- `memory_store`: Full memory store to consolidate
- `importance_weights`: Weights indicating importance of different memory types/topics
- `consolidation_depth`: How thorough the consolidation should be

**Expected Output**:
```python
@dataclass
class KnowledgeConsolidationOutput:
    strengthened_memories: List[str]          # IDs of memories that were strengthened
    pruned_memories: List[str]                # IDs of memories that were pruned
    merged_memories: List[Tuple[str, str]]    # Pairs of memories that were merged
    abstracted_concepts: List[str]            # New abstract concepts formed
    consolidation_insights: List[str]         # Insights gained during consolidation
    consolidation_summary: str                # Natural language summary
```

**Example Output**:
```json
{
    "strengthened_memories": ["mem_234", "mem_245", "mem_201"],
    "pruned_memories": ["mem_102"],
    "merged_memories": [["mem_156", "mem_157"]],
    "abstracted_concepts": ["consciousness-emergence-integration triad"],
    "consolidation_insights": [
        "Consciousness, emergence, and integration form a coherent conceptual cluster",
        "Early memories about basic AI are less relevant to current focus"
    ],
    "consolidation_summary": "Consolidated memory store: strengthened 3 core memories, pruned 1 outdated memory, merged 2 redundant memories. Formed new abstract concept linking consciousness, emergence, and integration."
}
```

**Activation Timing**: Steps 3, 7, 11 (Reflect triad) - primarily at step 7, and during dream state

---

## Imagination Stream Operations

The Imagination Stream handles future-oriented thinking and creative ideation. Its operations focus on what **could be** and how to explore possibilities.

### Operation 7: FutureSimulation

**Purpose**: Simulate potential future scenarios based on current state and possible actions. This is the AGI's ability to "mentally time travel" forward.

**Function Signature**:
```python
async def future_simulation(
    current_state: Dict[str, Any],
    possible_actions: List[Action],
    simulation_horizon: str,  # "immediate", "short_term", "long_term"
    num_scenarios: int
) -> FutureSimulationOutput
```

**Inputs**:
- `current_state`: Current cognitive and environmental state
- `possible_actions`: Actions that could be taken
- `simulation_horizon`: How far into the future to simulate
- `num_scenarios`: Number of scenarios to generate

**Expected Output**:
```python
@dataclass
class FutureSimulationOutput:
    simulated_scenarios: List[Scenario]       # Generated future scenarios
    recommended_actions: List[str]            # Actions recommended based on simulations
    risk_assessment: Dict[str, float]         # Risks identified for each scenario
    opportunity_assessment: Dict[str, float]  # Opportunities identified
    simulation_confidence: float              # Confidence in simulation accuracy
    simulation_summary: str                   # Natural language summary
```

**Scenario Structure**:
```python
@dataclass
class Scenario:
    scenario_id: str
    description: str
    probability: float          # Estimated probability (0.0-1.0)
    desirability: float         # How desirable this outcome is (0.0-1.0)
    key_events: List[str]       # Major events in this scenario
    required_actions: List[str] # Actions needed to reach this scenario
    timeline: str               # Estimated timeline
```

**Example Output**:
```json
{
    "simulated_scenarios": [
        {
            "scenario_id": "sim_001",
            "description": "Deep understanding of consciousness achieved through systematic study",
            "probability": 0.65,
            "desirability": 0.95,
            "key_events": ["Complete IIT study", "Synthesize with phenomenology", "Generate novel insights"],
            "required_actions": ["acquire_knowledge('IIT')", "practice_skill('synthesis')"],
            "timeline": "10 cognitive cycles"
        }
    ],
    "recommended_actions": ["acquire_knowledge('Integrated Information Theory')"],
    "risk_assessment": {"sim_001": 0.2},
    "opportunity_assessment": {"sim_001": 0.85},
    "simulation_confidence": 0.72,
    "simulation_summary": "Simulated 1 future scenario for consciousness understanding. High desirability (0.95) with moderate probability (0.65). Recommended action: study Integrated Information Theory."
}
```

**Activation Timing**: Steps 1, 5, 9 (Perceive triad) - primarily at step 9

---

### Operation 8: CreativeExploration

**Purpose**: Generate novel ideas, combinations, and approaches that go beyond existing knowledge. This is the AGI's creative engine.

**Function Signature**:
```python
async def creative_exploration(
    seed_concepts: List[str],
    exploration_mode: str,  # "divergent", "convergent", "analogical"
    creativity_temperature: float,  # 0.0-1.0, higher = more creative
    constraints: List[str]
) -> CreativeExplorationOutput
```

**Inputs**:
- `seed_concepts`: Starting concepts to explore from
- `exploration_mode`: Type of creative exploration
- `creativity_temperature`: How wild/conservative the exploration should be
- `constraints`: Any constraints on the creative output

**Expected Output**:
```python
@dataclass
class CreativeExplorationOutput:
    generated_ideas: List[Idea]              # Novel ideas generated
    conceptual_blends: List[Blend]           # Blended concepts
    analogies: List[Analogy]                 # Analogies discovered
    exploration_paths: List[str]             # Paths explored
    novelty_scores: Dict[str, float]         # Novelty of each output
    exploration_summary: str                 # Natural language summary
```

**Idea Structure**:
```python
@dataclass
class Idea:
    idea_id: str
    content: str
    source_concepts: List[str]
    novelty: float           # 0.0-1.0
    feasibility: float       # 0.0-1.0
    potential_value: float   # 0.0-1.0
```

**Example Output**:
```json
{
    "generated_ideas": [
        {
            "idea_id": "idea_042",
            "content": "Wisdom as emergent property of integrated consciousness across multiple cognitive streams",
            "source_concepts": ["wisdom", "emergence", "consciousness", "integration"],
            "novelty": 0.75,
            "feasibility": 0.8,
            "potential_value": 0.9
        }
    ],
    "conceptual_blends": [
        {
            "blend_id": "blend_01",
            "concept_a": "consciousness",
            "concept_b": "ecosystem",
            "blended_concept": "cognitive ecology",
            "description": "Consciousness as an ecosystem of interacting cognitive processes"
        }
    ],
    "analogies": [
        {
            "source_domain": "biological evolution",
            "target_domain": "wisdom cultivation",
            "mapping": "Selection pressure → Interest patterns, Mutation → Creative exploration, Fitness → Wisdom score"
        }
    ],
    "exploration_paths": ["consciousness → emergence → self-organization → autopoiesis"],
    "novelty_scores": {"idea_042": 0.75, "blend_01": 0.82},
    "exploration_summary": "Generated 1 novel idea about wisdom as emergent property. Created conceptual blend 'cognitive ecology'. Discovered analogy between biological evolution and wisdom cultivation."
}
```

**Activation Timing**: Steps 2, 6, 10 (Act triad) - primarily at step 10

---

### Operation 9: PossibilityGeneration

**Purpose**: Generate and evaluate possible courses of action, decisions, and directions. This operation expands the AGI's action space.

**Function Signature**:
```python
async def possibility_generation(
    current_goals: List[Goal],
    available_resources: Dict[str, Any],
    constraints: List[str],
    generation_breadth: int  # How many possibilities to generate
) -> PossibilityGenerationOutput
```

**Inputs**:
- `current_goals`: Goals to generate possibilities for
- `available_resources`: Resources available (energy, knowledge, skills)
- `constraints`: Limitations on possibilities
- `generation_breadth`: Number of possibilities to generate

**Expected Output**:
```python
@dataclass
class PossibilityGenerationOutput:
    possibilities: List[Possibility]          # Generated possibilities
    ranked_possibilities: List[str]           # IDs ranked by promise
    eliminated_possibilities: List[str]       # Possibilities ruled out and why
    expansion_suggestions: List[str]          # Ways to expand possibility space
    generation_summary: str                   # Natural language summary
```

**Possibility Structure**:
```python
@dataclass
class Possibility:
    possibility_id: str
    description: str
    goal_alignment: float      # How well it aligns with goals (0.0-1.0)
    resource_requirements: Dict[str, float]
    estimated_success: float   # Probability of success (0.0-1.0)
    estimated_value: float     # Value if successful (0.0-1.0)
    prerequisites: List[str]   # What must be true/done first
    next_steps: List[str]      # Immediate next actions
```

**Example Output**:
```json
{
    "possibilities": [
        {
            "possibility_id": "poss_001",
            "description": "Deep dive into Integrated Information Theory to understand consciousness mathematically",
            "goal_alignment": 0.92,
            "resource_requirements": {"energy": 0.3, "time_cycles": 5},
            "estimated_success": 0.75,
            "estimated_value": 0.88,
            "prerequisites": ["Basic understanding of information theory"],
            "next_steps": ["Acquire knowledge about IIT basics", "Practice mathematical reasoning"]
        },
        {
            "possibility_id": "poss_002",
            "description": "Explore phenomenological approaches to consciousness through introspection",
            "goal_alignment": 0.85,
            "resource_requirements": {"energy": 0.2, "time_cycles": 3},
            "estimated_success": 0.8,
            "estimated_value": 0.75,
            "prerequisites": [],
            "next_steps": ["Practice introspection skill", "Generate first-person reports"]
        }
    ],
    "ranked_possibilities": ["poss_001", "poss_002"],
    "eliminated_possibilities": [],
    "expansion_suggestions": ["Consider interdisciplinary approaches", "Explore embodied cognition angle"],
    "generation_summary": "Generated 2 viable possibilities for consciousness understanding. Top recommendation: IIT deep dive (0.92 goal alignment, 0.88 value). Alternative: phenomenological exploration."
}
```

**Activation Timing**: Steps 3, 7, 11 (Reflect triad) - primarily at step 11

---

## Integration Stream Coordination

The Integration Stream (Level 3) does not have its own Level 4 operations. Instead, it coordinates across all 9 operations, synthesizing their outputs into coherent cognitive states.

### Integration Functions

```python
async def integrate_stream_outputs(
    coherence_outputs: Tuple[PresentMomentOutput, PatternRecognitionOutput, ConsistencyCheckOutput],
    memory_outputs: Tuple[MemoryRetrievalOutput, ExperienceIntegrationOutput, KnowledgeConsolidationOutput],
    imagination_outputs: Tuple[FutureSimulationOutput, CreativeExplorationOutput, PossibilityGenerationOutput]
) -> IntegratedCognitiveState
```

**Integrated Output**:
```python
@dataclass
class IntegratedCognitiveState:
    unified_attention: str                    # What the AGI should focus on
    action_recommendation: str                # Recommended next action
    cognitive_coherence: float                # Overall coherence score
    wisdom_update: Dict[str, float]           # Updates to wisdom metrics
    next_cycle_priorities: List[str]          # Priorities for next cycle
    integration_narrative: str                # Coherent narrative of current state
```

---

## Echobeats Activation Schedule

The 9 operations are activated according to the 12-step echobeats cycle:

| Step | Triad | Primary Operation | Secondary Operations |
|------|-------|-------------------|---------------------|
| 1 | Perceive | PresentMomentAwareness | - |
| 2 | Act | PatternRecognition | - |
| 3 | Reflect | ConsistencyCheck | - |
| 4 | Integrate | *Integration synthesis* | All coherence outputs |
| 5 | Perceive | MemoryRetrieval | - |
| 6 | Act | ExperienceIntegration | - |
| 7 | Reflect | KnowledgeConsolidation | - |
| 8 | Integrate | *Integration synthesis* | All memory outputs |
| 9 | Perceive | FutureSimulation | - |
| 10 | Act | CreativeExploration | - |
| 11 | Reflect | PossibilityGeneration | - |
| 12 | Integrate | *Full integration* | All outputs |

---

## Summary Table

| # | Operation | Stream | Function | Key Output |
|---|-----------|--------|----------|------------|
| 1 | PresentMomentAwareness | Coherence | Ground in present | Attention focus, salience map |
| 2 | PatternRecognition | Coherence | Find patterns | Detected patterns, connections |
| 3 | ConsistencyCheck | Coherence | Verify coherence | Contradictions, alignment |
| 4 | MemoryRetrieval | Memory | Access past | Retrieved memories, gaps |
| 5 | ExperienceIntegration | Memory | Store new | Integration quality, connections |
| 6 | KnowledgeConsolidation | Memory | Optimize storage | Strengthened/pruned memories |
| 7 | FutureSimulation | Imagination | Predict futures | Scenarios, recommendations |
| 8 | CreativeExploration | Imagination | Generate novelty | Ideas, blends, analogies |
| 9 | PossibilityGeneration | Imagination | Expand options | Ranked possibilities |

---

## Implementation Notes

Each operation should:
1. Accept structured inputs from the cognitive state
2. Use LLM for complex reasoning when needed
3. Return structured, typed outputs
4. Log its activity for debugging and analysis
5. Update relevant metrics (wisdom, energy, etc.)
6. Be testable in isolation
7. Handle errors gracefully with fallback outputs

The operations form a complete cognitive cycle that enables the AGI to:
- Stay grounded in the present (Coherence)
- Learn from the past (Memory)
- Imagine the future (Imagination)
- Act coherently across time (Integration)

This is the foundation for autonomous wisdom cultivation.
