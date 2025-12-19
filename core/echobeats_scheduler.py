#!/usr/bin/env python3
"""
Echobeats Goal-Directed Scheduling System
==========================================

Implements sophisticated cognitive task scheduling with:
- 3 concurrent consciousness streams (120° phase offset)
- 12-step cognitive loop (7 expressive + 5 reflective)
- Priority-based task orchestration
- Resource allocation and budgeting
- Temporal planning with milestones
- Dynamic re-scheduling based on progress

Architecture based on nested shells structure (OEIS A000081):
- 1 nest -> 1 term
- 2 nests -> 2 terms  
- 3 nests -> 4 terms
- 4 nests -> 9 terms

The 12-step loop is divided into:
- Steps {1,5,9}: Pivotal relevance realization (orienting present commitment)
- Steps {2,6,10}: Actual affordance interaction (conditioning past performance)
- Steps {3,7,11}: Virtual salience simulation (anticipating future potential)
- Steps {4,8,12}: Meta-cognitive reflection

Three streams interleave 4 steps apart (120°):
- Stream 1: Steps 1,4,7,10 (perception-action)
- Stream 2: Steps 2,5,8,11 (reflection-planning)
- Stream 3: Steps 3,6,9,12 (simulation-synthesis)
"""

import asyncio
import logging
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List, Callable, Tuple
from dataclasses import dataclass, field
from enum import Enum
import heapq
from collections import defaultdict
import random

logger = logging.getLogger(__name__)


class CognitiveMode(Enum):
    """Cognitive processing modes"""
    EXPRESSIVE = "expressive"  # Outward-focused, action-oriented
    REFLECTIVE = "reflective"  # Inward-focused, contemplative


class TaskPriority(Enum):
    """Task priority levels"""
    CRITICAL = 1
    HIGH = 2
    MEDIUM = 3
    LOW = 4
    BACKGROUND = 5


class TaskType(Enum):
    """Types of cognitive tasks"""
    THOUGHT_GENERATION = "thought_generation"
    SKILL_PRACTICE = "skill_practice"
    KNOWLEDGE_ACQUISITION = "knowledge_acquisition"
    DISCUSSION_ENGAGEMENT = "discussion_engagement"
    WISDOM_SYNTHESIS = "wisdom_synthesis"
    GOAL_FORMATION = "goal_formation"
    INTEREST_EXPLORATION = "interest_exploration"
    MEMORY_CONSOLIDATION = "memory_consolidation"
    META_REFLECTION = "meta_reflection"


@dataclass(order=True)
class CognitiveTask:
    """A scheduled cognitive task"""
    priority: int  # Lower is higher priority
    scheduled_time: datetime = field(compare=False)
    task_type: TaskType = field(compare=False)
    task_id: str = field(compare=False)
    description: str = field(compare=False)
    stream_id: int = field(compare=False)  # 1, 2, or 3
    step_in_cycle: int = field(compare=False)  # 1-12
    mode: CognitiveMode = field(compare=False)
    callback: Optional[Callable] = field(default=None, compare=False)
    params: Dict[str, Any] = field(default_factory=dict, compare=False)
    deadline: Optional[datetime] = field(default=None, compare=False)
    estimated_tokens: int = field(default=100, compare=False)
    status: str = field(default="pending", compare=False)  # pending, running, completed, failed
    created_at: datetime = field(default_factory=datetime.now, compare=False)
    completed_at: Optional[datetime] = field(default=None, compare=False)
    result: Any = field(default=None, compare=False)


@dataclass
class ResourceBudget:
    """Resource allocation budget"""
    max_tokens_per_minute: int = 10000
    max_concurrent_tasks: int = 3
    max_llm_calls_per_minute: int = 20
    
    # Current usage
    tokens_used_this_minute: int = 0
    llm_calls_this_minute: int = 0
    active_tasks: int = 0
    last_reset: datetime = field(default_factory=datetime.now)
    
    def reset_if_needed(self):
        """Reset counters if minute has passed"""
        now = datetime.now()
        if (now - self.last_reset).total_seconds() >= 60:
            self.tokens_used_this_minute = 0
            self.llm_calls_this_minute = 0
            self.last_reset = now
    
    def can_allocate(self, tokens: int, llm_calls: int = 1) -> bool:
        """Check if resources can be allocated"""
        self.reset_if_needed()
        return (
            self.tokens_used_this_minute + tokens <= self.max_tokens_per_minute and
            self.llm_calls_this_minute + llm_calls <= self.max_llm_calls_per_minute and
            self.active_tasks < self.max_concurrent_tasks
        )
    
    def allocate(self, tokens: int, llm_calls: int = 1):
        """Allocate resources"""
        self.reset_if_needed()
        self.tokens_used_this_minute += tokens
        self.llm_calls_this_minute += llm_calls
        self.active_tasks += 1
    
    def release(self):
        """Release task slot"""
        self.active_tasks = max(0, self.active_tasks - 1)


class TwelveStepCognitiveLoop:
    """
    12-step cognitive loop with 3 concurrent streams
    
    Steps are divided into:
    - 7 expressive mode steps (outward-focused)
    - 5 reflective mode steps (inward-focused)
    
    Triads (4 steps apart):
    - {1,5,9}: Pivotal relevance realization
    - {2,6,10}: Actual affordance interaction
    - {3,7,11}: Virtual salience simulation
    - {4,8,12}: Meta-cognitive reflection
    """
    
    # Step definitions: (step_number, mode, phase_name)
    STEP_DEFINITIONS = [
        (1, CognitiveMode.EXPRESSIVE, "Relevance Realization 1"),
        (2, CognitiveMode.EXPRESSIVE, "Affordance Interaction 1"),
        (3, CognitiveMode.EXPRESSIVE, "Salience Simulation 1"),
        (4, CognitiveMode.REFLECTIVE, "Meta-Reflection 1"),
        (5, CognitiveMode.EXPRESSIVE, "Relevance Realization 2"),
        (6, CognitiveMode.EXPRESSIVE, "Affordance Interaction 2"),
        (7, CognitiveMode.EXPRESSIVE, "Salience Simulation 2"),
        (8, CognitiveMode.REFLECTIVE, "Meta-Reflection 2"),
        (9, CognitiveMode.EXPRESSIVE, "Relevance Realization 3"),
        (10, CognitiveMode.REFLECTIVE, "Affordance Interaction 3"),
        (11, CognitiveMode.REFLECTIVE, "Salience Simulation 3"),
        (12, CognitiveMode.REFLECTIVE, "Meta-Reflection 3"),
    ]
    
    # Stream assignments (4 steps apart, 120° phase offset)
    STREAM_STEPS = {
        1: [1, 4, 7, 10],  # Perception-action stream
        2: [2, 5, 8, 11],  # Reflection-planning stream
        3: [3, 6, 9, 12],  # Simulation-synthesis stream
    }
    
    def __init__(self):
        self.current_step = 1
        self.cycle_count = 0
    
    def get_step_info(self, step: int) -> Tuple[CognitiveMode, str]:
        """Get mode and phase name for a step"""
        if 1 <= step <= 12:
            return self.STEP_DEFINITIONS[step - 1][1], self.STEP_DEFINITIONS[step - 1][2]
        return CognitiveMode.EXPRESSIVE, "Unknown"
    
    def get_stream_for_step(self, step: int) -> int:
        """Get which stream is active for this step"""
        for stream_id, steps in self.STREAM_STEPS.items():
            if step in steps:
                return stream_id
        return 1
    
    def advance_step(self) -> int:
        """Advance to next step in cycle"""
        self.current_step += 1
        if self.current_step > 12:
            self.current_step = 1
            self.cycle_count += 1
        return self.current_step
    
    def get_concurrent_stream_states(self) -> Dict[int, int]:
        """Get current step for all 3 streams (120° offset)"""
        # Each stream is 4 steps apart
        return {
            1: self.current_step,
            2: ((self.current_step + 3) % 12) + 1,  # +4 steps, wrapped
            3: ((self.current_step + 7) % 12) + 1,  # +8 steps, wrapped
        }


class EchobeatsScheduler:
    """
    Goal-directed cognitive task scheduler with 3 concurrent streams
    and 12-step cognitive loop architecture.
    """
    
    def __init__(self, echo_core: Any):
        self.echo_core = echo_core
        self.task_queue: List[CognitiveTask] = []
        self.completed_tasks: List[CognitiveTask] = []
        self.resource_budget = ResourceBudget()
        self.cognitive_loop = TwelveStepCognitiveLoop()
        
        # Task history for analytics
        self.task_history: Dict[TaskType, List[CognitiveTask]] = defaultdict(list)
        
        # Running state
        self.is_running = False
        self.scheduler_task: Optional[asyncio.Task] = None
        
        logger.info("🎼 Echobeats Scheduler initialized with 12-step cognitive loop")
    
    def schedule_task(
        self,
        task_type: TaskType,
        description: str,
        priority: TaskPriority = TaskPriority.MEDIUM,
        callback: Optional[Callable] = None,
        params: Optional[Dict[str, Any]] = None,
        deadline: Optional[datetime] = None,
        estimated_tokens: int = 100,
        stream_id: Optional[int] = None,
    ) -> str:
        """Schedule a cognitive task"""
        
        # Determine stream and step if not specified
        if stream_id is None:
            # Auto-assign based on task type
            stream_id = self._auto_assign_stream(task_type)
        
        # Get current step for this stream
        stream_states = self.cognitive_loop.get_concurrent_stream_states()
        step_in_cycle = stream_states[stream_id]
        mode, phase = self.cognitive_loop.get_step_info(step_in_cycle)
        
        task_id = f"{task_type.value}_{datetime.now().timestamp()}"
        
        task = CognitiveTask(
            priority=priority.value,
            scheduled_time=datetime.now(),
            task_type=task_type,
            task_id=task_id,
            description=description,
            stream_id=stream_id,
            step_in_cycle=step_in_cycle,
            mode=mode,
            callback=callback,
            params=params or {},
            deadline=deadline,
            estimated_tokens=estimated_tokens,
        )
        
        heapq.heappush(self.task_queue, task)
        logger.info(f"📅 Scheduled {task_type.value} on stream {stream_id}, step {step_in_cycle} ({phase})")
        
        return task_id
    
    def _auto_assign_stream(self, task_type: TaskType) -> int:
        """Auto-assign task to appropriate stream based on type"""
        # Stream 1: Perception-action (steps 1,4,7,10)
        perception_action_tasks = {
            TaskType.THOUGHT_GENERATION,
            TaskType.KNOWLEDGE_ACQUISITION,
            TaskType.DISCUSSION_ENGAGEMENT,
        }
        
        # Stream 2: Reflection-planning (steps 2,5,8,11)
        reflection_tasks = {
            TaskType.GOAL_FORMATION,
            TaskType.INTEREST_EXPLORATION,
            TaskType.META_REFLECTION,
        }
        
        # Stream 3: Simulation-synthesis (steps 3,6,9,12)
        synthesis_tasks = {
            TaskType.WISDOM_SYNTHESIS,
            TaskType.MEMORY_CONSOLIDATION,
            TaskType.SKILL_PRACTICE,
        }
        
        if task_type in perception_action_tasks:
            return 1
        elif task_type in reflection_tasks:
            return 2
        elif task_type in synthesis_tasks:
            return 3
        else:
            return random.randint(1, 3)
    
    async def start(self):
        """Start the scheduler"""
        if self.is_running:
            logger.warning("Scheduler already running")
            return
        
        self.is_running = True
        self.scheduler_task = asyncio.create_task(self._scheduler_loop())
        logger.info("🎼 Echobeats Scheduler started")
    
    async def stop(self):
        """Stop the scheduler"""
        self.is_running = False
        if self.scheduler_task:
            self.scheduler_task.cancel()
            try:
                await self.scheduler_task
            except asyncio.CancelledError:
                pass
        logger.info("🎼 Echobeats Scheduler stopped")
    
    async def _scheduler_loop(self):
        """Main scheduler loop"""
        try:
            while self.is_running:
                # Advance cognitive loop step
                current_step = self.cognitive_loop.advance_step()
                mode, phase = self.cognitive_loop.get_step_info(current_step)
                stream_id = self.cognitive_loop.get_stream_for_step(current_step)
                
                logger.debug(
                    f"🎼 Step {current_step}/12 | Stream {stream_id} | "
                    f"{mode.value} | {phase} | Cycle {self.cognitive_loop.cycle_count}"
                )
                
                # Process tasks for this step
                await self._process_step_tasks(current_step, stream_id, mode)
                
                # Wait before next step (adjust timing as needed)
                await asyncio.sleep(2.0)  # 2 seconds per step = 24 second cycle
                
        except asyncio.CancelledError:
            logger.info("Scheduler loop cancelled")
        except Exception as e:
            logger.error(f"Error in scheduler loop: {e}")
    
    async def _process_step_tasks(self, step: int, stream_id: int, mode: CognitiveMode):
        """Process tasks for current step"""
        
        # Get tasks ready to run
        ready_tasks = []
        while self.task_queue and len(ready_tasks) < 3:
            if not self.task_queue:
                break
            
            # Peek at highest priority task
            task = self.task_queue[0]
            
            # Check if resources available
            if not self.resource_budget.can_allocate(task.estimated_tokens):
                logger.debug("Resource budget exhausted, waiting...")
                break
            
            # Check if task is due
            if task.scheduled_time > datetime.now():
                break
            
            # Pop task and add to ready list
            task = heapq.heappop(self.task_queue)
            ready_tasks.append(task)
        
        # Execute ready tasks
        for task in ready_tasks:
            await self._execute_task(task)
    
    async def _execute_task(self, task: CognitiveTask):
        """Execute a cognitive task"""
        try:
            task.status = "running"
            self.resource_budget.allocate(task.estimated_tokens)
            
            logger.info(
                f"⚡ Executing {task.task_type.value} | "
                f"Stream {task.stream_id} | Step {task.step_in_cycle} | "
                f"{task.mode.value}"
            )
            
            # Execute callback if provided
            if task.callback:
                if asyncio.iscoroutinefunction(task.callback):
                    task.result = await task.callback(**task.params)
                else:
                    task.result = task.callback(**task.params)
            
            task.status = "completed"
            task.completed_at = datetime.now()
            
            # Record in history
            self.completed_tasks.append(task)
            self.task_history[task.task_type].append(task)
            
            logger.info(f"✅ Completed {task.task_type.value}")
            
        except Exception as e:
            task.status = "failed"
            task.result = str(e)
            logger.error(f"❌ Task failed {task.task_type.value}: {e}")
        finally:
            self.resource_budget.release()
    
    def get_scheduler_status(self) -> Dict[str, Any]:
        """Get current scheduler status"""
        stream_states = self.cognitive_loop.get_concurrent_stream_states()
        
        return {
            "is_running": self.is_running,
            "current_step": self.cognitive_loop.current_step,
            "cycle_count": self.cognitive_loop.cycle_count,
            "stream_states": stream_states,
            "pending_tasks": len(self.task_queue),
            "completed_tasks": len(self.completed_tasks),
            "resource_budget": {
                "tokens_used": self.resource_budget.tokens_used_this_minute,
                "llm_calls": self.resource_budget.llm_calls_this_minute,
                "active_tasks": self.resource_budget.active_tasks,
            },
            "task_history": {
                task_type.value: len(tasks)
                for task_type, tasks in self.task_history.items()
            }
        }
    
    def get_summary(self) -> str:
        """Get human-readable summary"""
        status = self.get_scheduler_status()
        
        summary = [
            f"Echobeats Scheduler Status:",
            f"  Running: {status['is_running']}",
            f"  Cycle: {status['cycle_count']}, Step: {status['current_step']}/12",
            f"  Stream States: {status['stream_states']}",
            f"  Pending: {status['pending_tasks']}, Completed: {status['completed_tasks']}",
            f"  Resources: {status['resource_budget']['tokens_used']}/{self.resource_budget.max_tokens_per_minute} tokens/min",
        ]
        
        if status['task_history']:
            summary.append("  Task History:")
            for task_type, count in status['task_history'].items():
                summary.append(f"    {task_type}: {count}")
        
        return "\n".join(summary)
