# Intelligent Machines Architecture

## What are Intelligent Machines?

Intelligent Machines are AI-powered entities that spawn in isolated environments (VMs for MVP, containers post-MVP) to perform project evolution operations. Each machine has a specific prompt/goal and operates within strict resource constraints.

## Machine Types

### Claude Code Machine
- **Purpose**: Primary machine for all operations (MVP focus)
- **Capabilities**: Complex reasoning, code generation, file operations
- **Best For**: Architecture decisions, feature implementation, debugging

### Future Machine Types (Post-MVP)
- **Aider Machine**: Focused code editing and refactoring
- **Qwen Coder Machine**: Local inference for simple operations
- See POST_MVP.md for details

## Machine Control Script

Each machine is controlled by a simple Python script running as a systemctl service:

### Python Script (`intelligent_machine.py`)
- **Service Management**: Runs as systemd service on each worker VM
- **Claude Integration**: Calls Claude Code with `--param` for single prompt execution
- **State Persistence**: Saves machine state between operations and waits
- **Progress Reporting**: Reports current status back to orchestrator
- **Resource Monitoring**: Tracks API usage, time, and resource consumption

### SystemD Service Configuration
```ini
[Unit]
Description=AutoVibe Intelligent Machine
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/python3 /opt/autovibe/intelligent_machine.py --project-id=%i
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Script Workflow
1. **Initialization**: Load project checkpoint and machine configuration
2. **Prompt Execution**: Execute single Claude Code command with specific parameters
3. **State Save**: Persist machine state after each operation
4. **Wait State**: Idle until next prompt or termination signal
5. **Reporting**: Send progress updates to orchestrator

## Machine Lifecycle

```
SPAWNING → ACTIVE → WORKING → TERMINATING → TERMINATED
     ↓        ↓        ↓         ↓           ↓
   Budget   File    Save     Clean      State
   Check    Ops     State    Exit       Saved
```

### Spawn Process
1. **Resource Verification**: Check if sufficient resources (money, time, API calls) exist
2. **VM Creation**: Launch isolated environment with project checkpoint
3. **Machine Activation**: Initialize Intelligent Machine with prompt/goal
4. **File Replacement**: Apply any file modifications before machine starts
5. **Context Loading**: Load current project files and state

### Operation Phase  
1. **File Analysis**: Machine examines current project state
2. **Planning**: Determine what file operations are needed
3. **Execution**: Perform file replacements/modifications
4. **Validation**: Verify changes meet the prompt/goal

### Termination Process
1. **State Capture**: Save all modified files and project state
2. **Checkpoint Creation**: Create checkpoint for future branching  
3. **Resource Reporting**: Log actual resource usage (immediately, not waiting for termination)
4. **Clean Shutdown**: Terminate machine and deallocate resources

## VM Environment

Each Intelligent Machine operates in:
- **Isolated VM**: Full environment isolation
- **Project Files**: Complete project codebase loaded
- **Development Tools**: Code editors, compilers, testing frameworks
- **Network Access**: Limited, budget-controlled API access
- **State Persistence**: Ability to save/snapshot entire environment

## Resource Integration

- **Pre-spawn Check**: Verify all resource budgets before machine creation
- **Real-time Monitoring**: Track all resource usage (API calls, compute, time, money)
- **Resource Limits**: API-level restrictions based on available resources
- **Usage Reporting**: Detailed breakdown of actual vs estimated resource consumption