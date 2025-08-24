# AutoVibe VM Architecture

## VM-Based Isolation

Each Intelligent Machine operates in a completely isolated virtual machine environment, enabling clean state management and evolution tracking.

## VM Lifecycle

### VM Creation Process
1. **Base Image**: Start from clean development environment template
2. **Project Loading**: Mount/copy current project files into VM
3. **Tool Installation**: Install required development tools (compilers, IDEs, etc.)
4. **Network Configuration**: Limited internet access for API calls only
5. **Monitoring Setup**: Resource usage and cost tracking agents

### VM Operation Phase
1. **Machine Spawn**: Launch Intelligent Machine process within VM
2. **File System Access**: Machine can read/write all project files  
3. **Resource Monitoring**: Track CPU, memory, disk, network usage
4. **Cost Accumulation**: Monitor API calls and resource consumption

### VM Snapshot & Termination
1. **State Capture**: Freeze entire VM state including file system
2. **Snapshot Creation**: Save complete VM image for future restoration
3. **Resource Cleanup**: Deallocate VM resources
4. **Cost Reporting**: Final billing and budget updates

## VM Types & Specifications

### Standard Development VM (MVP)
**Specifications:**
- CPU: 2 cores
- Memory: 4GB RAM
- Disk: 20GB SSD
- Network: API access only
- Resource cost: Tracked per actual usage

**Use Cases:**
- Claude Code Intelligent Machines
- General development tasks
- File operations and code generation

### Lightweight VM (Post-MVP)
See POST_MVP.md for lightweight VM specifications and use cases.

## Project Evolution Through VM Snapshots

### Evolution Tree Structure
```
Project Genesis VM
        ↓
    [Snapshot A] ← "Add login feature" 
        ↓
    [Snapshot B] ← "Fix security bug"
        ↓               ↘
    [Snapshot C]    [Snapshot B'] ← "Try different approach"
```

### Snapshot Management
- **Incremental Snapshots**: Only store changes from previous state
- **Branching**: Create multiple evolution paths from same snapshot
- **Rollback**: Restore to any previous snapshot state
- **Pruning**: Remove unused evolution branches to save storage

## VM Security & Isolation

### Network Configuration
- **Full Internet Access**: VMs have complete internet access (required for development)
- **Traffic Interception**: All traffic routed through mitmproxy for logging
- **Traffic Storage**: Complete HTTP/HTTPS capture for analysis and billing
- **No Restrictions**: Firewall controls deferred to post-MVP

### File System Isolation  
- **Sandboxed Environment**: No access to host system
- **Project Scope**: Only project files accessible to Intelligent Machine
- **Temporary Storage**: Separate temp space for processing

### Resource Limits
- **CPU Throttling**: Prevent runaway processes
- **Memory Caps**: Hard limits to prevent system impact
- **Disk Quotas**: Control storage usage per project
- **Time Limits**: Maximum runtime before automatic termination

## Cost Optimization

### VM Lifecycle Management
- Immediate shutdown after Intelligent Machine termination
- No idle VMs consuming resources
- Snapshot compression to reduce storage costs

### Resource Scaling (Post-MVP)
- **Auto-scaling**: Create VMs only when needed
- **Shared Resources**: Multiple small tasks in single VM
- **Preemptible VMs**: Use cheaper, interruptible instances when possible

This VM architecture ensures complete isolation while enabling sophisticated project evolution tracking.