# Database Design - PostgreSQL Schema

## Overview

autovibe uses PostgreSQL as its primary data store for metadata, resource tracking, and system state management. The schema is designed for scalability, performance, and comprehensive audit trails.

## Core Tables

### Projects
The Projects table stores project metadata using UUID primary keys, with required name and greediness factor (0-1 decimal with validation). The table tracks project status (default 'active'), description, initial and current checkpoint references, flexible metadata in JSONB format, and timestamps. Indexes optimize queries on greediness (descending), status, and metadata (GIN index). An update trigger automatically maintains the updated_at timestamp.

### Intelligent Machines
The Intelligent Machines table stores all active and historical machine instances using UUID primary keys. It includes the machine type (claude_code, aider, qwen_coder, hub), current status (spawning, active, working, checkpointing, terminated), and references to the associated project and parent checkpoint. The table tracks VM identification details, original prompt text, file modifications in JSONB format, resource constraints and usage data, current progress status, error messages if any, and checkpoint information. Timestamps record spawning and termination times, with flexible metadata storage. Indexes optimize queries on status, project association, spawn time, machine type, and hub relationships.

### Checkpoints
The Checkpoints table maintains a complete history of all system snapshots using UUID primary keys. Each checkpoint references its project, optional parent checkpoint, and the machine that created it. The checkpoint type indicates its origin (genesis, automatic, manual, final, hub_daily). The table stores the Proxmox snapshot ID, associated VM ID, and a file manifest in JSONB format containing filename mappings to size, SHA256 hash, and modification timestamps. Additional fields include description text, creation context, size in bytes, compression ratio, and verification status. Timestamps track creation and verification times, with flexible metadata storage. Indexes optimize queries on project association, parent relationships, creation time (descending), checkpoint type, and creating machine. A partial index specifically targets verified checkpoints for active project queries.

### Resource Usage Tracking
The Resource Usage table tracks all system resource consumption with UUID primary keys and references to projects and machines. It categorizes resources by type (money, time, api_calls, cpu, memory, storage, network) and optional subtype (claude_api, vm_runtime, storage_snapshot). Each entry records the consumed amount with its unit (CAD, USD, minutes, hours, calls, gb, gb_hours) and normalized cost in CAD for aggregation. Records include timestamps and billing period dates for time-based analysis. The table uses monthly partitioning for performance with appropriate indexes on project-period combinations, machine references, resource types, and recording timestamps. A sample partition structure is established for monthly data separation.

### Resource Budgets
The Resource Budgets table manages spending limits and allocations using UUID primary keys with project references. Each budget entry specifies the resource type and budget period (daily, weekly, monthly, total) with defined start and end dates. The table tracks allocated amounts, current usage, and reserved amounts for pending operations. Units specify the measurement type, and auto-reset controls automatic budget renewal. Timestamps track creation and updates. A unique constraint ensures one budget per project-resource-period combination. Indexes optimize queries on project periods and active budget lookups.

## Hub Management Tables

### Hub Instances
The Hub Instances table tracks all Hub VM instances using UUID primary keys. Each hub has a type (primary, development, experimental, domain_specific) and status (active, inactive, experimental, retiring). The table stores VM identification details, version information, and configuration data in JSONB format. Hub relationships are tracked through parent hub references and experiment associations. Timestamps record spawning, last checkpoint, and termination times, with flexible metadata storage. Indexes optimize queries on status and type for efficient hub management operations.

### Hub Checkpoints
The Hub Checkpoints table maintains snapshots of Hub instances using UUID primary keys with references to hub instances. Each checkpoint specifies its type (daily, manual, pre_experiment, recovery) and stores the Proxmox snapshot ID with trigger reason. System metrics at checkpoint time are stored in JSONB format, including CPU, memory, and disk usage. The table tracks active machines and projects counts, checkpoint size in gigabytes, and verification status. Timestamps record creation and verification times with flexible metadata storage. Indexes optimize queries on hub association, creation time (descending), and checkpoint type.

### Hub Experiments
The Hub Experiments table manages Hub evolution experiments using UUID primary keys. Each experiment has a name and description, references to base checkpoint and original hub, with optional experimental hub assignment. Configuration changes are stored in JSONB format along with success metrics definitions and collected performance data. The experiment status tracks progression (planned, running, completed, failed) with configurable duration in days. Timestamps record start and completion times, with results categorization (experimental_wins, original_wins, inconclusive). Creation timestamps and metadata provide additional tracking capabilities. Indexes optimize queries on status and start times for experiment management.

## Performance and Analytics Views

### Project Resource Summary (Materialized View)
The Project Resource Summary is a materialized view that aggregates resource usage data by project, date, and resource type over the last 90 days. It joins project information with resource usage records to calculate total amounts, costs in CAD, transaction counts, and average amounts per resource type. The view includes project UUID, name, greediness factor, usage date, resource type, total amounts with units, total costs, transaction counts, and average amounts. A unique index ensures data integrity across project, date, and resource type combinations. The view refreshes daily at 1 AM using pg_cron scheduling for up-to-date analytics.

### Machine Performance Metrics
The Machine Performance Metrics is a materialized view that analyzes machine performance over the last 30 days, excluding hub-type machines. It calculates key performance indicators by machine type including total machine count, average runtime in minutes, successful completion count, success rate percentage, average money cost, average API calls, and average checkpoints created per machine. The view filters for successfully terminated machines without error messages to calculate success rates. Performance data is extracted from JSONB resource usage fields. A unique index ensures one entry per machine type for efficient performance tracking and analysis.

## Database Functions and Triggers

### Resource Budget Enforcement
The Resource Budget Enforcement function is a database trigger that validates resource usage against daily budgets before allowing new usage records. The function retrieves the current daily budget for the project and resource type, checks if adding the new usage amount would exceed the allocated budget, and raises an exception if limits are exceeded. When usage is within limits, the function updates the budget's used amount and timestamp. The trigger executes before each insert on the resource usage table, ensuring real-time budget enforcement with detailed error messages including current usage, attempted addition, and budget limits.

### Checkpoint Verification
The Checkpoint Verification function schedules asynchronous integrity verification for newly created checkpoints. When a checkpoint is inserted, the trigger adds an entry to the checkpoint verification queue with the checkpoint UUID and current timestamp. This allows the system to process verification tasks in the background without blocking checkpoint creation, ensuring data integrity while maintaining system performance.

## Database Maintenance and Optimization

### Automated Cleanup Jobs
The Automated Cleanup system includes functions to maintain database performance by removing old data. The cleanup function for resource usage records removes entries older than one year and logs the cleanup operation with affected row counts and timestamp to a maintenance log table. The cleanup is scheduled to run monthly on the first day at 2 AM using pg_cron scheduling, ensuring regular maintenance without manual intervention.

### Backup and Recovery Considerations
The database backup strategy prioritizes tables based on criticality and size. Projects table requires immediate backup due to its critical nature and small size. Checkpoints table, while large, is critical and requires frequent backup with a 15-minute recovery point objective. Intelligent machines table is moderately sized and important for system state. Resource usage table is large but has a one-hour recovery point objective since it can be reconstructed from system logs. Hub instances table is small but critical for system operations. The backup strategy ensures zero data loss for projects while allowing acceptable recovery windows for other components based on their reconstruction capabilities.

This PostgreSQL schema provides robust data management for autovibe with comprehensive audit trails, resource tracking, and performance optimization capabilities.