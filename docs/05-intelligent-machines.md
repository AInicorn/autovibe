# Intelligent Machines - Unified Architecture

## Core Philosophy

Every component in autovibe is an Intelligent Machine following the same architectural principles:
- **Spawn** from checkpoints in isolated VM environments
- **Execute** specific goals while creating regular checkpoints  
- **Communicate** via defined interfaces and permissions
- **Terminate** cleanly, creating final checkpoints for future evolution

## Machine Categories

### Task Machines
Perform specific development, analysis, or operational tasks:

#### Claude Code Machine
- **Purpose**: Primary machine for complex reasoning, code generation, architecture decisions
- **Capabilities**: Multi-file editing, system design, debugging, documentation
- **Resource Profile**: High API usage, moderate compute, variable runtime
- **Checkpointing**: After each API response, before major operations

#### Aider Machine (Post-MVP)
- **Purpose**: Specialized for focused code editing and refactoring
- **Capabilities**: Single-file changes, test generation, code review
- **Resource Profile**: Moderate API usage, low compute, fast runtime
- **Checkpointing**: After file modifications, between logical steps

#### Qwen Coder Machine (Post-MVP)  
- **Purpose**: Local inference for simple operations, bulk changes
- **Capabilities**: Template generation, documentation, repetitive tasks
- **Resource Profile**: Local compute intensive, no API costs
- **Checkpointing**: After batch operations, on completion

### Hub Machines
Orchestrate and manage other machines:

#### Primary Hub
- **Purpose**: Main orchestrator, API server, resource manager
- **Capabilities**: Machine spawning, budget allocation, checkpoint coordination
- **Resource Profile**: Continuous operation, moderate compute, database intensive
- **Checkpointing**: Daily production snapshots, before major changes

#### Specialized Hubs (Post-MVP)
- **Development Hub**: Experimental features, algorithm testing
- **Domain Hub**: Client-specific or project-specific configurations
- **Geographic Hub**: Regional deployment for latency optimization

## Machine Lifecycle

### Unified Spawn Process
The unified spawn process involves five steps: First, validate that sufficient resources are available for the requested machine type and options. Second, create a VM by restoring from the specified checkpoint with calculated resource limits. Third, apply any file modifications if specified in the options. Fourth, register the machine in the system database with VM ID, type, parent checkpoint, prompt, and spawning Hub ID. Finally, start machine execution with the provided prompt and return the machine UUID.

### Regular Checkpointing During Operation
All machines create checkpoints at logical intervals:

The Intelligent Machine base class implements checkpointing behavior with configurable triggers: after API responses, after file operations, at 5-minute time intervals, every 10% of budget usage, and on user request. The class provides methods to evaluate whether checkpointing should occur and creates checkpoints with captured file state and working memory context when triggered.

### Clean Termination Process
The clean termination process involves five steps: First, create a final checkpoint capturing the machine's end state and termination reason. Second, calculate and record final resource usage for billing and analysis. Third, shut down the VM and clean up associated Proxmox resources. Fourth, update the machine's status to terminated in the database with the final checkpoint reference. Finally, notify the spawning Hub of completion with the machine UUID and final checkpoint details.

## Communication and Permissions

### Hub-to-Machine Communication
- **Spawning**: Hub creates machines via Proxmox API
- **Monitoring**: Hub queries machine status via VM metrics
- **Termination**: Hub can request clean shutdown or force termination
- **Checkpointing**: Hub can trigger manual checkpoints

### Machine-to-Hub Communication  
- **Status Updates**: Machines report progress, resource usage, errors
- **Resource Requests**: Machines can request additional resources
- **Checkpoint Notifications**: Machines report checkpoint creation
- **Completion Reports**: Machines report final results

### Machine-to-Machine Communication (Post-MVP)
- **File Exchange**: Machines can request files from each other's checkpoints
- **Conversation Sharing**: Access to other machines' conversation history
- **Collaborative Checkpointing**: Coordinated checkpoints across related machines

### Permission Model
The permission model defines different access levels for machine types:

**Task Machines**: No direct Proxmox access, limited Hub API access (status updates, resource requests, checkpoint notifications), file system read/write within VM only, and network access restricted to specific API endpoints only.

**Hub Machines**: Full Proxmox API access (VM creation, control, snapshot management), complete Hub API access, full file system permissions, and network access to both APIs and Proxmox management interfaces.

## Resource Management Integration

### Resource Tracking
Every machine continuously tracks:
- **Money**: API costs, VM runtime costs, storage costs
- **Time**: Execution duration, idle time, queue time  
- **Compute**: CPU usage, memory consumption, disk I/O
- **API Calls**: Service-specific quotas and rate limits

### Budget Enforcement
Budget enforcement checks resource limits before operations by retrieving machine and project data, calculating current usage and allocated budgets. For each resource type (money, time, API calls, compute), the system verifies that current usage plus estimated operation cost doesn't exceed the allocated budget (adjusted by project greediness factor). If limits are exceeded, a ResourceLimitExceededError is raised with details about the resource type and usage levels.

### Machine Selection Algorithm
The machine selection algorithm analyzes prompt requirements and identifies available machine types within the project budget. Each machine type is scored based on three factors: capability match with prompt requirements (40% weight), historical efficiency for similar tasks (40% weight), and budget efficiency calculated as cost per unit divided by quality score (20% weight). The algorithm returns the machine type with the highest total score.

## Machine Templates and Configuration

### VM Templates
Each machine type has optimized VM templates:
- **Base Image**: Ubuntu/Debian with common tools
- **Machine-Specific Tools**: Claude API clients, development tools, specialized software
- **Resource Profiles**: CPU, memory, disk sized for machine type
- **Network Configuration**: Appropriate access permissions

### Configuration Management
Machine configuration profiles define resource and tool requirements for different machine types:

**Claude Code Machines**: Use Ubuntu development template with 2 CPU cores, 4GB memory, 20GB disk, API-only network profile, and tools including Claude API client, Git, Docker, Node.js, and Python. Checkpointing occurs at high frequency (after each API call).

**Hub Machines**: Use Ubuntu server template with 4 CPU cores, 8GB memory, 50GB disk, management network profile, and tools including Docker, PostgreSQL, Nginx, and Proxmox client. Checkpointing occurs daily for production stability.

This unified architecture ensures all components follow the same patterns while allowing specialization for different roles and capabilities.