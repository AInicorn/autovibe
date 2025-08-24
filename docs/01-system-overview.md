# AutoVibe System Overview

## What is AutoVibe?

AutoVibe is an Intelligent Machine orchestration system that manages AI-powered entities with budget-aware economic controls. The system spawns Intelligent Machines in isolated VM environments to evolve projects through file operations while maintaining strict budget limits.

## Core Concept: Project Evolution Tree

AutoVibe works on a **project evolution model**:
- **Project State** (files, code, VM snapshot)
- **Prompt/Goal** + **VM State** = **Spawn Intelligent Machine**
- **Intelligent Machine performs file operations** (replacements, modifications)
- **[Post-MVP] Machine returns output files and conversation logs**
- **Machine terminates cleanly** 
- **Save new project state** for next evolution branch

```
Project State A + "Add login" → Intelligent Machine → Project State B
                                                             ↓
Project State B + "Fix bug" → Intelligent Machine → Project State C
```

## System Goals

1. **Budget-First Operations**: Every Intelligent Machine spawn is budget-controlled
2. **Clean VM Isolation**: Each machine runs in isolated, saveable VM environment
3. **File-Based Evolution**: Projects evolve through intelligent file modifications
4. **Economic Efficiency**: Automatic cost tracking and spending limits

## Key Components

- **Budget Manager**: Tracks spending and enforces limits per project/machine
- **VM Orchestrator**: Creates, snapshots, and manages isolated environments  
- **Machine Spawner**: Launches Intelligent Machines with specific prompts/goals
- **Project Manager**: Manages multiple parallel project evolution trees
- **Economic System**: Handles budget allocation based on project "greediness"

## MVP Focus

For initial release, keep it simple:
- **Primary Machine**: Claude Code for all operations
- **File Operations**: Replace/modify files within VM
- **Basic Budget Check**: "Can we afford to spawn this machine?"
- **Manual Prompts**: User provides evolution goal/prompt
- **Git Integration**: Post-MVP (for now, all state in VM)