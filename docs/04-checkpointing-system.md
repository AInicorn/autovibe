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

#### Application-Level Checkpointing (Post-MVP)
Task machine checkpointing captures comprehensive state information:

1. **State Collection**: Gather machine ID, filesystem state, conversation history, working memory context, and resource usage statistics
2. **VM Snapshot Creation**: Trigger Proxmox API to create VM snapshot of the entire machine state
3. **Metadata Storage**: Store checkpoint data and snapshot reference in database using ORM operations
4. **UUID Return**: Provide unique checkpoint identifier for future reference and recovery operations

### Hub Machines

#### When to Checkpoint
- **Daily Production**: Automated daily snapshots (00:00 UTC)
- **Development Disabled**: No automatic checkpointing in dev environments
- **Before Evolution**: Manual checkpoint before spawning experimental Hubs
- **Configuration Changes**: Before major system configuration updates
- **Recovery Points**: After successful completion of complex operations

#### Infrastructure-Level Checkpointing
Hub checkpointing process operates at the infrastructure level:

1. **UUID Generation**: Create unique checkpoint identifier for tracking
2. **Database Marking**: Record checkpoint creation with Hub VM ID, type "infrastructure", trigger context, and "creating" status
3. **Proxmox Snapshot**: Call Proxmox API to create named VM snapshot with checkpoint UUID and descriptive metadata
4. **Status Update**: Mark checkpoint as "completed" in database once Proxmox confirms snapshot creation

## Checkpoint Storage and Management

### Storage Architecture
- **VM Snapshots**: Stored in ZFS/PBS via Proxmox
- **Metadata**: PostgreSQL database with checkpoint references
- **File Manifests**: Detailed file listings and checksums
- **Incremental Storage**: ZFS deduplication for efficient storage

### Database Schema Design
Checkpoint metadata is stored using ORM models with the following key entities:

**Main Checkpoints Table**: Contains UUID primary keys, machine references, parent checkpoint relationships, checkpoint types (task/hub/emergency), Proxmox snapshot references, file manifests as JSON, creation timestamps, status tracking, size information, and flexible metadata storage.

**Hub Checkpoints Table**: Specialized for Hub-level checkpoints with Hub VM identifiers, trigger reasons (daily/manual/pre-evolution), system metrics as JSON, active machine counts, and checkpoint size tracking in GB.

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
Task machine recovery involves a three-step process: First, retrieve checkpoint metadata using ORM operations, then create a new VM by restoring from the Proxmox snapshot with a generated name based on the checkpoint UUID. Second, update the machine registry with the new VM ID, parent checkpoint reference, and restored status. Finally, return the machine record for resuming or branching operations.

### Hub Recovery Process
Hub recovery follows a four-step process: First, retrieve hub checkpoint data and create a recovery log entry with the source checkpoint UUID, recovery reason, timestamp, and in-progress status. Second, restore the Hub VM from the Proxmox snapshot using the stored snapshot ID. Third, update DNS and load balancing configuration to route traffic to the restored Hub instance. Finally, verify system health and update recovery status to either completed (with administrator notification) or failed (triggering fallback to older checkpoints or manual intervention).

## Multi-Hub Evolutionary Checkpointing

### Experimental Hub Creation
When spawning experimental Hubs:

1. **Select Evolutionary Base**: Choose checkpoint from ~2 weeks ago
2. **Define Variations**: Configuration changes, algorithm tweaks, new features
3. **Create Branch Point**: Mark this as evolutionary experiment in database
4. **Spawn with Modifications**: Apply variations during restoration process
5. **Track Performance**: Monitor both original and experimental Hubs

### Evolutionary Selection Process
The evolutionary selection process operates with a configurable experiment duration (default 14 days). The system retrieves all active Hub experiments and evaluates those that have run for the full duration. Performance metrics are collected from both original and experimental Hubs, comparing efficiency, budget utilization, and success rates. If the experimental Hub shows superior overall performance, it gets promoted while the original Hub is retired and the evolutionary success is recorded. Otherwise, the experimental Hub is terminated, the original continues operation, and the evolutionary failure is logged.

## Emergency Recovery Procedures

### Cascade Recovery System
If Hub fails completely:

1. **Detect Failure**: Health monitoring detects Hub unresponsiveness
2. **Automatic Recovery**: Attempt restoration from most recent checkpoint
3. **Progressive Fallback**: If recent checkpoint fails, try progressively older ones
4. **Deep Recovery**: Fall back to stable checkpoint from 1+ weeks ago
5. **Manual Intervention**: Alert administrators if all automatic recovery fails

### Data Integrity Verification
Data integrity verification performs a three-step validation process: First, verify database metadata consistency for the checkpoint record. Second, confirm that the referenced Proxmox snapshot actually exists in storage. Third, if a file manifest is present, verify that the manifest matches the actual snapshot contents. The verification returns success status with appropriate error messages for any validation failures.

## Checkpoint Performance Optimization

### Incremental Checkpointing
- **ZFS Copy-on-Write**: Only changed blocks stored in snapshots
- **File-Level Deduplication**: Identical files across checkpoints shared
- **Compression**: Snapshot data compressed for storage efficiency
- **Background Processing**: Checkpoint creation doesn't block machine operation

### Cleanup and Maintenance
Automated checkpoint maintenance applies retention policies by identifying expired checkpoints through ORM queries. For each expired checkpoint, the system removes the associated Proxmox snapshot, deletes the database metadata record, and updates storage usage statistics to maintain accurate capacity tracking.

This comprehensive checkpointing system enables fine-grained recovery, experimental branching, and self-healing capabilities across all system components.