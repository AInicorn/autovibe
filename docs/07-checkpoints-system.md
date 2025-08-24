# AutoVibe Machine Checkpoints System

## Overview

The checkpoints system is the core mechanism for project evolution in AutoVibe. Each checkpoint represents a complete snapshot of a project's state at a specific point in time, enabling branching evolution paths and reproducible machine spawning.

## Core Concepts

### What is a Checkpoint?

A checkpoint is a complete, immutable snapshot that includes:
- **VM Snapshot**: Full virtual machine state including OS, installed tools, and environment
- **File System State**: All project files and their exact contents
- **Metadata**: Information about when, why, and how the checkpoint was created
- **Parent Reference**: Link to the previous checkpoint in the evolution chain

### Checkpoint vs Traditional Version Control

Unlike Git commits which track file changes, checkpoints capture:
- Entire VM environment state
- Installed dependencies and tools
- System configuration
- Running processes state (optional)
- Complete reproducibility of the execution environment

## Checkpoint Lifecycle

### Creation Process

1. **Trigger Event**: Machine completes its task or manual checkpoint request
2. **State Freeze**: VM execution pauses momentarily
3. **Snapshot Creation**: VM state is captured to disk
4. **File Manifest**: All project files are cataloged with hashes
5. **Metadata Recording**: Checkpoint details stored in database
6. **Parent Linking**: Reference to parent checkpoint established
7. **Checkpoint UUID**: Unique identifier generated

### Checkpoint Composition

A checkpoint is composed of separate, independently managed components:

#### Core Checkpoint Record
```
Checkpoint {
    uuid: "checkpoint-550e8400-e29b-41d4",
    project_uuid: "project-123e4567-e89b", 
    parent_checkpoint_uuid: "checkpoint-parent-uuid",
    created_at: "2024-08-23T10:30:00Z",
    description: "After implementing authentication system",
    components: {
        vm_snapshot_id: "vm-550e8400",
        file_manifest_id: "files-550e8400", 
        metadata_id: "meta-550e8400"
    }
}
```

#### VM Snapshot Component
```
VMSnapshot {
    id: "vm-550e8400",
    checkpoint_uuid: "checkpoint-550e8400-e29b-41d4",
    snapshot_path: "/storage/vm-snapshots/vm-550e8400.qcow2",
    size_bytes: 5368709120,
    compression: "qcow2_compressed"
}
```

#### File Manifest Component  
```
FileManifest {
    id: "files-550e8400",
    checkpoint_uuid: "checkpoint-550e8400-e29b-41d4",
    files: {
        "src/main.py": "sha256:abc123...",
        "config/settings.json": "sha256:def456...",
        // ... all project files
    }
}
```

#### Metadata Component
```
CheckpointMetadata {
    id: "meta-550e8400", 
    checkpoint_uuid: "checkpoint-550e8400-e29b-41d4",
    machine_uuid: "machine-that-created-this",
    prompt: "Add user authentication",
    machine_type: "CLAUDE_CODE",
    resource_usage: {...}
}
```

## Evolution Tree

### Tree Structure

Projects evolve through a tree of checkpoints:

```
Initial Checkpoint (Genesis)
    │
    ├─> Checkpoint A (Added database schema)
    │       │
    │       ├─> Checkpoint B (Implemented API)
    │       │       │
    │       │       └─> Checkpoint C (Added tests)
    │       │
    │       └─> Checkpoint B' (Alternative API design)
    │               │
    │               └─> Checkpoint C' (Different approach)
    │
    └─> Checkpoint A2 (Different architecture)
```

### Branching Strategy

- **Linear Evolution**: Most common path, sequential improvements
- **Experimental Branches**: Try different approaches from same checkpoint
- **Parallel Development**: Multiple machines working from same checkpoint
- **Merge Points**: Combine successful features from different branches (post-MVP)

## Machine Spawning from Checkpoints

### Spawn Process

1. **Select Checkpoint**: Choose starting point in evolution tree
2. **Optional File Modifications**: Apply file replacements before spawn
3. **VM Restoration**: Create new VM from checkpoint snapshot
4. **Machine Initialization**: Start Intelligent Machine in restored environment
5. **Prompt Execution**: Machine begins working on assigned task

### File Modification System

Before spawning a machine, files can be replaced:

```
spawn_machine({
    checkpoint_uuid: "checkpoint-550e8400",
    prompt: "Refactor authentication to use OAuth",
    file_modifications: {
        "config/auth.json": "new OAuth configuration content",
        "src/auth.py": "updated authentication module"
    }
})
```

This allows:
- Configuration changes without new checkpoints
- Quick iterations on specific files
- A/B testing different approaches
- Hot-fixing issues before machine spawn

## Storage and Performance

### Storage Strategy

- **Incremental Snapshots**: Only store differences from parent
- **Compression**: QCOW2 format with compression
- **Deduplication**: Shared base layers for common environments
- **Tiered Storage**: Recent checkpoints on fast SSD, older on HDD

### Performance Optimization

- **Copy-on-Write**: Efficient snapshot creation
- **Lazy Loading**: Load VM state on-demand
- **Parallel Restoration**: Multiple VMs from same checkpoint
- **Checkpoint Caching**: Keep frequently used checkpoints in memory

## Checkpoint Management

### Retention Policy

- **Active Checkpoints**: All checkpoints in main evolution path
- **Experimental Branches**: Keep for 30 days unless marked important
- **Failed Evolutions**: Keep for debugging (7 days default)
- **Archive Mode**: Compress and store long-term

### Pruning Strategy

Remove checkpoints that:
- Lead to dead ends with no value
- Are intermediate states with no branches
- Exceed retention policy limits
- Have been superseded by better alternatives

### Checkpoint Operations

- **List**: Show all checkpoints for a project
- **Diff**: Compare files between checkpoints
- **Restore**: Spawn new machine from checkpoint
- **Delete**: Remove checkpoint and update tree
- **Export**: Package checkpoint for sharing
- **Import**: Load external checkpoint

## Integration with Project Evolution

### Automatic Checkpointing

Checkpoints are created automatically when:
- Intelligent Machine completes its task successfully
- Significant milestone is reached
- Manual checkpoint request via API
- Before risky operations (optional)

### Evolution Tracking

Each evolution step is recorded:
- From checkpoint UUID
- To checkpoint UUID
- Prompt that drove the evolution
- Resource usage for the operation
- Success/failure status
- Error messages if failed

## API Integration

### Checkpoint Endpoints

```
# List project checkpoints
GET /projects/{project_uuid}/checkpoints

# Get checkpoint details
GET /checkpoints/{checkpoint_uuid}

# Create checkpoint from current state
POST /projects/{project_uuid}/checkpoints

# Spawn machine from checkpoint
POST /machines
{
    checkpoint_uuid: "...",
    prompt: "...",
    file_modifications: {}
}

# Compare checkpoints
GET /checkpoints/{uuid1}/diff/{uuid2}

# Delete checkpoint
DELETE /checkpoints/{checkpoint_uuid}
```

## Best Practices

### When to Create Checkpoints

- After completing major features
- Before experimental changes
- At stable, working states
- After successful test runs

### Checkpoint Naming

Use descriptive names that indicate:
- What was accomplished
- Why the checkpoint was created
- Whether it's a stable state

### Branch Management

- Keep main evolution path clean
- Name experimental branches clearly
- Document failed attempts for learning
- Merge successful experiments back

## Future Enhancements (Post-MVP)

### Advanced Features

- **Checkpoint Merging**: Combine features from multiple branches
- **Distributed Checkpoints**: Store across multiple nodes
- **Incremental File Tracking**: Git-like file versioning within checkpoints
- **Checkpoint Templates**: Reusable base checkpoints
- **Cross-Project Checkpoints**: Share common states

### Performance Improvements

- **Predictive Caching**: Pre-load likely next checkpoints
- **Checkpoint Streaming**: Start VMs before full restoration
- **P2P Checkpoint Sharing**: Distributed checkpoint storage
- **Smart Compression**: AI-optimized compression strategies

This checkpoints system provides the foundation for AutoVibe's unique project evolution approach, enabling reproducible, branchable, and efficient development workflows.