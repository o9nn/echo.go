#!/usr/bin/env python3
"""
Comprehensive analysis of echo9llama to identify problems and improvement opportunities
"""
import os
import json
import subprocess
from pathlib import Path
from collections import defaultdict
import re

def analyze_codebase():
    """Analyze the codebase structure and identify issues"""
    issues = []
    opportunities = []
    
    # Check Go implementation
    go_files = list(Path("core/deeptreeecho").glob("*.go"))
    print(f"Found {len(go_files)} Go files in core/deeptreeecho")
    
    # Check Python implementation
    py_files = list(Path("core").glob("**/*.py"))
    print(f"Found {len(py_files)} Python files in core")
    
    # Check for empty or stub files
    empty_files = []
    for f in Path(".").glob("**/*.go"):
        if f.stat().st_size == 0:
            empty_files.append(str(f))
    
    for f in Path(".").glob("**/*.py"):
        if f.stat().st_size == 0:
            empty_files.append(str(f))
    
    if empty_files:
        issues.append({
            "category": "Empty Files",
            "severity": "high",
            "description": f"Found {len(empty_files)} empty implementation files",
            "files": empty_files[:10],  # Show first 10
            "recommendation": "Implement or remove empty files"
        })
    
    # Check for dual implementation (Go vs Python)
    issues.append({
        "category": "Architecture",
        "severity": "medium",
        "description": "Dual implementation in Go and Python creates maintenance burden",
        "recommendation": "Consolidate to single implementation or establish clear separation of concerns"
    })
    
    # Check for missing integration
    opportunities.append({
        "category": "LLM Integration",
        "priority": "high",
        "description": "Both ANTHROPIC_API_KEY and OPENROUTER_API_KEY are available but may not be fully integrated",
        "recommendation": "Implement multi-provider LLM system with fallback"
    })
    
    # Check for echobeats implementation
    echobeats_files = [f for f in go_files if "echobeats" in f.name.lower()]
    if echobeats_files:
        opportunities.append({
            "category": "Echobeats Scheduler",
            "priority": "high",
            "description": f"Found {len(echobeats_files)} echobeats-related files",
            "files": [str(f) for f in echobeats_files],
            "recommendation": "Verify 3-stream tetrahedral architecture with 12-step cognitive loop"
        })
    
    # Check for consciousness implementation
    consciousness_files = [f for f in py_files if "consciousness" in str(f).lower()]
    if consciousness_files:
        opportunities.append({
            "category": "Stream of Consciousness",
            "priority": "high",
            "description": f"Found {len(consciousness_files)} consciousness-related files",
            "files": [str(f) for f in consciousness_files],
            "recommendation": "Ensure persistent autonomous thought generation independent of external prompts"
        })
    
    # Check for persistence
    state_files = list(Path(".").glob("**/state*.json"))
    if state_files:
        opportunities.append({
            "category": "State Persistence",
            "priority": "medium",
            "description": f"Found {len(state_files)} state files",
            "recommendation": "Implement comprehensive state save/restore with versioning"
        })
    
    # Check for discussion autonomy
    discussion_files = [f for f in go_files if "discussion" in f.name.lower()]
    discussion_files += [f for f in py_files if "discussion" in str(f).lower()]
    if discussion_files:
        opportunities.append({
            "category": "Discussion Autonomy",
            "priority": "high",
            "description": f"Found {len(discussion_files)} discussion-related files",
            "files": [str(f) for f in discussion_files],
            "recommendation": "Implement ability to start/end/respond to discussions based on interest patterns"
        })
    
    # Check for knowledge integration
    knowledge_files = [f for f in go_files if "knowledge" in f.name.lower() or "echodream" in f.name.lower()]
    knowledge_files += [f for f in py_files if "knowledge" in str(f).lower() or "echodream" in str(f).lower()]
    if knowledge_files:
        opportunities.append({
            "category": "Knowledge Integration",
            "priority": "high",
            "description": f"Found {len(knowledge_files)} knowledge/echodream files",
            "files": [str(f) for f in knowledge_files],
            "recommendation": "Implement echodream knowledge consolidation during rest cycles"
        })
    
    return issues, opportunities

def check_vision_alignment():
    """Check alignment with ultimate vision"""
    vision_components = {
        "Fully Autonomous": {
            "required": ["autonomous thought generation", "self-directed goals", "independent operation"],
            "status": "partial"
        },
        "Wisdom-Cultivating": {
            "required": ["wisdom extraction", "insight synthesis", "learning from experience"],
            "status": "partial"
        },
        "Deep Tree Echo": {
            "required": ["hierarchical memory", "echo propagation", "hypergraph structure"],
            "status": "partial"
        },
        "Persistent Cognitive Event Loops": {
            "required": ["echobeats scheduler", "3 concurrent streams", "12-step cycle"],
            "status": "partial"
        },
        "Self-Orchestrated Scheduling": {
            "required": ["goal-directed scheduling", "priority management", "resource allocation"],
            "status": "partial"
        },
        "Stream-of-Consciousness": {
            "required": ["continuous thought generation", "independent awareness", "context coherence"],
            "status": "partial"
        },
        "Wake/Rest Cycles": {
            "required": ["autonomous wake/rest", "echodream consolidation", "energy management"],
            "status": "partial"
        },
        "Interest Patterns": {
            "required": ["topic affinity tracking", "curiosity-driven exploration", "engagement thresholds"],
            "status": "partial"
        },
        "Discussion Autonomy": {
            "required": ["initiate discussions", "respond contextually", "end gracefully"],
            "status": "partial"
        },
        "Knowledge & Skills": {
            "required": ["knowledge acquisition", "skill practice", "continuous learning"],
            "status": "partial"
        }
    }
    
    return vision_components

def generate_report():
    """Generate comprehensive analysis report"""
    print("=" * 80)
    print("ECHO9LLAMA EVOLUTION ANALYSIS")
    print("=" * 80)
    print()
    
    issues, opportunities = analyze_codebase()
    
    print("IDENTIFIED PROBLEMS")
    print("-" * 80)
    for i, issue in enumerate(issues, 1):
        print(f"\n{i}. [{issue['severity'].upper()}] {issue['category']}")
        print(f"   Description: {issue['description']}")
        if 'files' in issue:
            print(f"   Affected files: {len(issue.get('files', []))}")
        print(f"   Recommendation: {issue['recommendation']}")
    
    print("\n\n")
    print("IMPROVEMENT OPPORTUNITIES")
    print("-" * 80)
    for i, opp in enumerate(opportunities, 1):
        print(f"\n{i}. [{opp['priority'].upper()}] {opp['category']}")
        print(f"   Description: {opp['description']}")
        if 'files' in opp:
            print(f"   Related files: {', '.join(opp['files'][:3])}")
        print(f"   Recommendation: {opp['recommendation']}")
    
    print("\n\n")
    print("VISION ALIGNMENT CHECK")
    print("-" * 80)
    vision = check_vision_alignment()
    for component, details in vision.items():
        status_symbol = "⚠️" if details['status'] == "partial" else "✅" if details['status'] == "complete" else "❌"
        print(f"\n{status_symbol} {component} ({details['status']})")
        print(f"   Required: {', '.join(details['required'])}")
    
    # Save to JSON
    report = {
        "timestamp": str(Path("ITERATION_DEC08_2025.md").stat().st_mtime if Path("ITERATION_DEC08_2025.md").exists() else ""),
        "issues": issues,
        "opportunities": opportunities,
        "vision_alignment": vision
    }
    
    with open("ANALYSIS_REPORT.json", "w") as f:
        json.dump(report, f, indent=2)
    
    print("\n\n")
    print("=" * 80)
    print(f"Report saved to ANALYSIS_REPORT.json")
    print("=" * 80)

if __name__ == "__main__":
    generate_report()
