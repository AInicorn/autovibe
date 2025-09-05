# Hub Architecture - The Orchestrator Machine

## Hub as Intelligent Machine

The Hub is itself an Intelligent Machine that manages other Intelligent Machines. It follows the same lifecycle patterns but with specialized orchestration capabilities.

## Hub VM Structure

### VM Configuration
- **Type**: Standard VM (not container for MVP)
- **Resources**: 4+ cores, 8+ GB RAM, 50+ GB disk
- **Network**: API access + Proxmox management interface access
- **Permissions**: Can spawn/terminate other VMs via Proxmox API

### Internal Stack (Docker Compose)
```yaml
# Conceptual structure - not implementation
services:
  orchestrator:
    # Main Python application
    # Handles API requests, machine spawning, budget management
    
  postgres:
    # Project data, checkpoints metadata, resource tracking
    
  redis:
    # Session state, caching, job queues
    
  nginx:
    # Reverse proxy, SSL termination, rate limiting
```

## Hub Checkpointing System

### Infrastructure-Level Checkpointing
Unlike task machines that checkpoint application state, Hub checkpoints its **entire VM**:

1. **Mark in Database**: Update checkpoint UUID in PostgreSQL
2. **Trigger Snapshot**: Call Proxmox API to create ZFS/PBS snapshot of entire VM
3. **Verify Completion**: Confirm snapshot creation and update database status

### Checkpoint Frequency
- **Production**: Daily automated checkpoints (00:00 UTC)
- **Development**: Manual checkpoints only (auto-checkpointing disabled)
- **Before Major Operations**: Manual checkpoint before spawning experimental Hubs
- **Emergency**: On-demand checkpoints before risky system changes

### Technical Implementation
```python
# Conceptual flow
def create_hub_checkpoint():
    checkpoint_uuid = generate_uuid()
    
    # 1. Mark in database first
    db.execute("""
        INSERT INTO hub_checkpoints (uuid, vm_id, status, created_at)
        VALUES (%s, %s, 'creating', NOW())
    """, checkpoint_uuid, hub_vm_id)
    
    # 2. Trigger Proxmox snapshot
    proxmox_api.create_vm_snapshot(hub_vm_id, checkpoint_uuid)
    
    # 3. Update status on completion
    db.execute("""
        UPDATE hub_checkpoints SET status = 'completed'
        WHERE uuid = %s
    """, checkpoint_uuid)
```

## Multi-Hub Evolution System

### Hub Spawning Strategy
When current Hub shows signs of local optima or performance issues:

1. **Identify Stagnation**: Metrics show declining efficiency, budget waste, or poor machine selection
2. **Select Evolutionary Checkpoint**: Choose checkpoint from ~2 weeks ago as evolutionary base
3. **Define Experimental Parameters**: Different objectives, algorithms, or configurations
4. **Spawn Experimental Hub**: Create new Hub VM from selected checkpoint
5. **Parallel Operation**: Run both Hubs in isolated environments for evaluation period

### Hub Selection Process
After 14-day experimental period:

1. **Performance Comparison**: Compare resource efficiency, project completion rates, budget utilization
2. **A/B Test Results**: Analyze which Hub produces better outcomes
3. **Selection Decision**: Choose superior Hub, terminate inferior one
4. **State Migration**: If experimental Hub wins, migrate active projects and continue evolution

### Hub Types and Specialization

#### Production Hub
- **Focus**: Stability, proven algorithms, conservative resource allocation
- **Checkpointing**: Daily automated snapshots
- **Risk Tolerance**: Low - prioritizes reliability over optimization

#### Development Hub  
- **Focus**: Experimental features, algorithm testing, new machine types
- **Checkpointing**: Manual only - frequent iterations without snapshot overhead
- **Risk Tolerance**: High - allows failure for learning

#### Domain-Specific Hubs
- **Client Hubs**: Dedicated to specific clients with custom configurations
- **Project Hubs**: Specialized for particular project types (web dev, ML, DevOps)
- **Geographic Hubs**: Located in different regions for latency optimization

## Hub Recovery and Rollback

### Self-Recovery Mechanisms
Hubs can detect and recover from failures:

1. **Health Monitoring**: Continuous self-assessment of system health
2. **Automatic Rollback**: If critical failure detected, rollback to previous checkpoint
3. **Progressive Recovery**: Try recent checkpoints first, fall back to older ones if needed
4. **Alert Generation**: Notify administrators of recovery events

### Recovery Decision Logic
```python
def hub_recovery_strategy():
    if critical_failure_detected():
        # Try last 3 checkpoints progressively
        for checkpoint in get_recent_checkpoints(limit=3):
            if attempt_rollback(checkpoint):
                log_recovery_event("success", checkpoint.uuid)
                return
        
        # If recent checkpoints fail, try 1-week old
        stable_checkpoint = get_checkpoint_from_days_ago(7)
        if attempt_rollback(stable_checkpoint):
            log_recovery_event("deep_rollback", stable_checkpoint.uuid)
            alert_administrators("Hub recovered from week-old checkpoint")
```

## Proxmox Integration

### Required Proxmox Permissions
Hub VM needs specific permissions to manage other VMs:
- **VM Creation**: Spawn new VMs from templates
- **VM Control**: Start, stop, restart operations  
- **Snapshot Management**: Create, list, delete VM snapshots
- **Resource Monitoring**: Check VM resource utilization
- **Template Access**: Clone from pre-configured machine templates

### API Integration Points
- **VM Lifecycle**: Create → Configure → Start → Monitor → Stop → Snapshot → Destroy
- **Resource Quotas**: Enforce budget limits through VM resource constraints
- **Snapshot Orchestration**: Coordinate checkpoint creation across multiple VMs
- **Health Monitoring**: Track VM performance and availability

## Hub Security Model

### Isolation Principles
- **Network Segmentation**: Hub VM in management VLAN, task VMs in isolated networks
- **API Authentication**: All Proxmox API calls authenticated and logged
- **Resource Limits**: Hub cannot exceed allocated resource quotas
- **Audit Trail**: All Hub actions logged for security and debugging

### Permission Granularity
- **VM Management**: Can create/destroy VMs within allocated resource pools
- **Snapshot Control**: Can trigger snapshots but not directly access snapshot storage
- **Database Access**: Full control over autovibe database, read-only system tables
- **Network Access**: API endpoints only, no direct internet access for task VMs

This Hub architecture creates a self-managing, self-evolving orchestration system that maintains the same Intelligent Machine principles while providing sophisticated infrastructure management capabilities.