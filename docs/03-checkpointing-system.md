# Comprehensive Checkpointing System

## Philosophy: Regular Checkpointing, Not Just Termination

autovibe uses **fine-grained checkpointing** throughout machine operation, not just at completion. This enables precise recovery, branching, and optimization.

## Checkpointing Strategies by Machine Type

### Task Machines (Claude Code, Aider, etc.)

#### When to Checkpoint
- **Between API Messages**: After each Claude API response
- **After File Operations**: Successful file writes, project structure changes
- **Before Risky Operations**: Destructive refactoring, major architectural changes
- **At Natural Breakpoints**: End of functions, completion of user stories
- **Resource Thresholds**: Every N API calls or M minutes of runtime
- **User-Triggered**: Manual checkpoint requests via API

#### Application-Level Checkpointing
```python
# Conceptual checkpoint creation for task machines
def create_task_checkpoint(machine_id, context="automatic"):
    checkpoint_data = {
        "machine_id": machine_id,
        "file_system": capture_filesystem_state(),
        "conversation_history": get_recent_messages(),
        "working_memory": extract_machine_context(),
        "resource_usage": get_current_usage_stats(),
        "context": context
    }
    
    # 1. Create VM snapshot via Proxmox
    snapshot_id = proxmox.create_vm_snapshot(machine_id)
    
    # 2. Store metadata in database
    checkpoint_uuid = db.create_checkpoint(checkpoint_data, snapshot_id)
    
    return checkpoint_uuid
```

### Hub Machines

#### When to Checkpoint
- **Daily Production**: Automated daily snapshots (00:00 UTC)
- **Development Disabled**: No automatic checkpointing in dev environments
- **Before Evolution**: Manual checkpoint before spawning experimental Hubs
- **Configuration Changes**: Before major system configuration updates
- **Recovery Points**: After successful completion of complex operations

#### Infrastructure-Level Checkpointing
```python
# Hub checkpointing at VM level
def create_hub_checkpoint():
    checkpoint_uuid = generate_uuid()
    
    # Mark creation in database
    hub_checkpoint = {
        "uuid": checkpoint_uuid,
        "hub_vm_id": current_hub_vm_id,
        "type": "infrastructure",
        "trigger": determine_checkpoint_trigger(),
        "status": "creating"
    }
    
    db.create_hub_checkpoint(hub_checkpoint)
    
    # Trigger Proxmox VM snapshot
    proxmox_result = proxmox_api.snapshot_vm(
        vm_id=current_hub_vm_id,
        snapshot_name=checkpoint_uuid,
        description=f"Hub checkpoint {checkpoint_uuid}"
    )
    
    # Update status on completion
    db.update_checkpoint_status(checkpoint_uuid, "completed")
```

## Checkpoint Storage and Management

### Storage Architecture
- **VM Snapshots**: Stored in ZFS/PBS via Proxmox
- **Metadata**: PostgreSQL database with checkpoint references
- **File Manifests**: Detailed file listings and checksums
- **Incremental Storage**: ZFS deduplication for efficient storage

### Database Schema (Conceptual)
```sql
-- Checkpoint metadata
CREATE TABLE checkpoints (
    uuid UUID PRIMARY KEY,
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    checkpoint_type VARCHAR(50), -- 'task', 'hub', 'emergency'
    proxmox_snapshot_id VARCHAR(100),
    file_manifest JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'completed',
    size_bytes BIGINT,
    metadata JSONB
);

-- Hub-specific checkpoints
CREATE TABLE hub_checkpoints (
    uuid UUID PRIMARY KEY,
    hub_vm_id VARCHAR(50),
    trigger_reason VARCHAR(100), -- 'daily', 'manual', 'pre_evolution'
    system_metrics JSONB,
    active_machines INTEGER,
    checkpoint_size_gb DECIMAL(10,2)
);
```

### Checkpoint Retention Policies

#### Task Machine Checkpoints
- **Recent Checkpoints**: Keep all checkpoints from last 7 days
- **Daily Snapshots**: Keep one checkpoint per day for last 30 days  
- **Weekly Archives**: Keep weekly checkpoints for 6 months
- **Milestone Checkpoints**: Manually marked important checkpoints kept indefinitely

#### Hub Checkpoints
- **Daily Production**: Keep daily snapshots for 30 days
- **Weekly Archives**: Keep weekly snapshots for 1 year
- **Evolutionary Bases**: Keep checkpoints used for Hub evolution experiments
- **Recovery Points**: Keep checkpoints that were successfully used for recovery

## Checkpoint Recovery and Branching

### Task Machine Recovery
```python
def restore_task_machine(checkpoint_uuid):
    checkpoint = db.get_checkpoint(checkpoint_uuid)
    
    # 1. Create new VM from checkpoint
    new_vm_id = proxmox.restore_vm_from_snapshot(
        snapshot_id=checkpoint.proxmox_snapshot_id,
        vm_name=f"restored-{checkpoint_uuid[:8]}"
    )
    
    # 2. Update machine registry
    machine = create_machine_record(
        vm_id=new_vm_id,
        parent_checkpoint=checkpoint_uuid,
        status="restored"
    )
    
    # 3. Resume or branch operation
    return machine
```

### Hub Recovery Process
```python
def recover_hub_from_checkpoint(checkpoint_uuid, recovery_reason="manual"):
    hub_checkpoint = db.get_hub_checkpoint(checkpoint_uuid)
    
    # 1. Create recovery log entry
    recovery_log = {
        "from_checkpoint": checkpoint_uuid,
        "reason": recovery_reason,
        "initiated_at": datetime.utcnow(),
        "status": "in_progress"
    }
    db.log_hub_recovery(recovery_log)
    
    # 2. Restore Hub VM from snapshot
    restored_vm_id = proxmox.restore_hub_vm(hub_checkpoint.proxmox_snapshot_id)
    
    # 3. Update DNS/load balancing to point to restored Hub
    update_hub_routing(restored_vm_id)
    
    # 4. Verify system health
    if verify_hub_health(restored_vm_id):
        db.update_recovery_status(recovery_log.id, "completed")
        alert_administrators(f"Hub recovered successfully from {checkpoint_uuid}")
    else:
        db.update_recovery_status(recovery_log.id, "failed")
        # Try older checkpoint or alert for manual intervention
```

## Multi-Hub Evolutionary Checkpointing

### Experimental Hub Creation
When spawning experimental Hubs:

1. **Select Evolutionary Base**: Choose checkpoint from ~2 weeks ago
2. **Define Variations**: Configuration changes, algorithm tweaks, new features
3. **Create Branch Point**: Mark this as evolutionary experiment in database
4. **Spawn with Modifications**: Apply variations during restoration process
5. **Track Performance**: Monitor both original and experimental Hubs

### Evolutionary Selection Process
```python
def hub_evolutionary_selection(experiment_duration_days=14):
    active_experiments = db.get_active_hub_experiments()
    
    for experiment in active_experiments:
        if experiment.runtime_days >= experiment_duration_days:
            # Compare performance metrics
            original_metrics = get_hub_performance(experiment.original_hub_id)
            experimental_metrics = get_hub_performance(experiment.experimental_hub_id)
            
            # Selection criteria: efficiency, budget utilization, success rate
            if experimental_metrics.overall_score > original_metrics.overall_score:
                # Experimental Hub wins - promote it
                promote_experimental_hub(experiment.experimental_hub_id)
                retire_hub(experiment.original_hub_id)
                
                db.record_evolutionary_success(experiment.id)
            else:
                # Original Hub wins - terminate experiment
                retire_hub(experiment.experimental_hub_id)
                db.record_evolutionary_failure(experiment.id)
```

## Emergency Recovery Procedures

### Cascade Recovery System
If Hub fails completely:

1. **Detect Failure**: Health monitoring detects Hub unresponsiveness
2. **Automatic Recovery**: Attempt restoration from most recent checkpoint
3. **Progressive Fallback**: If recent checkpoint fails, try progressively older ones
4. **Deep Recovery**: Fall back to stable checkpoint from 1+ weeks ago
5. **Manual Intervention**: Alert administrators if all automatic recovery fails

### Data Integrity Verification
```python
def verify_checkpoint_integrity(checkpoint_uuid):
    checkpoint = db.get_checkpoint(checkpoint_uuid)
    
    # 1. Verify database consistency
    if not db.verify_checkpoint_metadata(checkpoint_uuid):
        return False, "Database metadata inconsistent"
    
    # 2. Verify Proxmox snapshot exists
    if not proxmox.snapshot_exists(checkpoint.proxmox_snapshot_id):
        return False, "Proxmox snapshot missing"
    
    # 3. Verify file manifest matches snapshot contents
    if checkpoint.file_manifest:
        if not verify_file_manifest(checkpoint):
            return False, "File manifest verification failed"
    
    return True, "Checkpoint integrity verified"
```

## Checkpoint Performance Optimization

### Incremental Checkpointing
- **ZFS Copy-on-Write**: Only changed blocks stored in snapshots
- **File-Level Deduplication**: Identical files across checkpoints shared
- **Compression**: Snapshot data compressed for storage efficiency
- **Background Processing**: Checkpoint creation doesn't block machine operation

### Cleanup and Maintenance
```python
# Automated checkpoint maintenance
def cleanup_expired_checkpoints():
    # Apply retention policies
    expired_checkpoints = db.get_expired_checkpoints()
    
    for checkpoint in expired_checkpoints:
        # Remove Proxmox snapshot
        proxmox.delete_snapshot(checkpoint.proxmox_snapshot_id)
        
        # Remove database metadata
        db.delete_checkpoint(checkpoint.uuid)
        
        # Update storage usage statistics
        update_storage_metrics()
```

This comprehensive checkpointing system enables fine-grained recovery, experimental branching, and self-healing capabilities across all system components.