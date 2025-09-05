# autovibe System Overview

## What is autovibe?

autovibe is an Intelligent Machine orchestration system where AI-powered entities manage project evolution through regular checkpointing and budget-aware resource allocation.

## Core Philosophy: Everything is an Intelligent Machine

**Intelligent Machines** are AI entities that:
- Spawn from checkpoints in isolated VM environments
- Execute specific goals/prompts
- Create regular checkpoints during operation
- Terminate cleanly and save new project state

This includes:
- **Task Machines**: Claude Code, Aider, Qwen Coder performing development work
- **Hub Machines**: Orchestrator systems managing other machines and resources

## Project Evolution Model

autovibe operates on a **checkpoint-driven evolution model**:

1. **Project State** (files, code, VM snapshot stored as checkpoint)
2. **Goal/Prompt** + **Checkpoint** = **Spawn Intelligent Machine**
3. **Machine operates** with regular checkpointing between logical steps
4. **Machine terminates** creating final checkpoint
5. **New checkpoint** becomes basis for next evolution

The evolution process follows a sequential pattern: Starting with Checkpoint A and a goal to "Add authentication", an Intelligent Machine is spawned, performs the work, and creates Checkpoint B. Then Checkpoint B becomes the foundation for the next evolution cycle where another goal like "Add testing" spawns a new Machine that creates Checkpoint C, and so forth.

## Hub as Intelligent Machine

The **Hub** itself is an Intelligent Machine running in a VM with:
- **Docker Compose stack**: Python orchestrator + PostgreSQL + services
- **Regular checkpointing**: Daily snapshots via Proxmox/ZFS
- **Self-evolution**: Can checkpoint and evolve itself
- **Multi-Hub spawning**: Can create other Hub instances for different environments, experimental testing, client isolation, and performance optimization

### Hub Responsibilities
- Budget management and resource allocation
- Intelligent Machine spawning and lifecycle management
- Checkpoint orchestration and storage
- API endpoint provision
- Multi-project coordination

## Multi-Hub Architecture

Hubs can spawn other Hub instances:
- **Production Hub**: Stable operations
- **Development Hub**: Experimental features  
- **Domain Hubs**: Project-specific or client-specific
- **Evolutionary Hubs**: Testing different optimization approaches

### Hub Selection and Evolution
- When Hub performance stagnates (local optima), spawn new Hub from 2-week old checkpoint
- Run experimental approaches in parallel for 14 days
- Select better-performing Hub, terminate underperformer
- Continuous Hub evolution and optimization

## Regular Checkpointing Strategy

Unlike traditional systems that only checkpoint at completion, autovibe uses **fine-grained checkpointing**:

### For Task Machines
- Between Claude API messages/responses
- After significant file operations
- Before risky operations
- At natural breakpoints in workflows

### For Hub Machines  
- Daily in production environments
- Disabled by default in development
- Before major configuration changes
- Before spawning experimental Hubs

## Resource Management

Every operation is governed by **resource budgets**:
- **Money**: API costs, VM costs, storage costs
- **Time**: Maximum runtime limits
- **Compute**: CPU, memory, disk quotas
- **API Calls**: Service-specific rate limits

Budget allocation uses **greediness factors** (0.0-1.0) for priority-based resource distribution.

## Technical Foundation

- **VM Orchestration**: Proxmox for VM management and snapshotting
- **Storage**: ZFS/PBS for checkpoint storage and deduplication  
- **Database**: PostgreSQL for metadata, resource tracking, and system state
- **API**: RESTful endpoints for machine management and monitoring
- **Permissions**: Fine-grained access control for machine-to-infrastructure communication

## MVP Focus

Initial implementation prioritizes:
- **Single Hub**: One orchestrator Hub managing all operations
- **VM-only**: All machines run in VMs (containers post-MVP)
- **Claude Code Primary**: Focus on one machine type initially  
- **Manual Prompts**: User-provided evolution goals
- **Basic Budgeting**: Simple resource tracking and limits

This architecture creates a self-managing, self-evolving system where every component follows the same Intelligent Machine lifecycle patterns.