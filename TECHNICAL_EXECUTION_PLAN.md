# autovibe Technical Execution Plan - ORM & Data Architecture

## 🎯 EXECUTIVE SUMMARY

This document focuses on the **Object-Relational Mapping (ORM) architecture and data modeling decisions** for the autovibe Intelligent Machine orchestration system. It consolidates all database design patterns, ORM relationships, and data architecture considerations from the comprehensive design documents.

---

## 🗄️ ORM ARCHITECTURE & DATA MODELING

### Technology Stack - Data Layer
```yaml
Database: PostgreSQL 15+
ORM Framework: SQLAlchemy 2.0+ (Async)
Migration Tool: Alembic
Connection Pool: asyncpg + SQLAlchemy async engine
Cache Layer: Redis 7+ for ORM query caching
```

### ORM Design Philosophy

**Core Principles:**
- **UUID-based Primary Keys**: All entities use UUID v4 for global uniqueness
- **Temporal Data Tracking**: Created/updated timestamps on all mutable entities  
- **JSONB Metadata**: Flexible schema extension via PostgreSQL JSONB
- **Soft References**: Foreign keys with cascading behavior carefully designed
- **Partitioning Strategy**: Time-based partitioning for high-volume data

---

## 📊 CORE ENTITY RELATIONSHIPS

### Project-Centric Domain Model

```
Projects (Root Aggregate)
├── Has Many: Intelligent Machines
│   └── Has Many: Checkpoints (via creating_machine)
├── Has Many: Checkpoints (project association)
├── Has Many: Resource Usage Records
├── Has Many: Resource Budgets
└── Has Current: Checkpoint (soft reference)

Hub Instances (Separate Aggregate)  
├── Has Many: Hub Checkpoints
├── Has Many: Hub Experiments (as original/experimental)
└── Self-Reference: Parent Hub (evolution chain)

Resource Management (Cross-Cutting)
├── Resource Usage (partitioned by time)
├── Resource Budgets (project scoped)
└── Budget Enforcement (trigger-based)
```

### ORM Relationship Patterns

#### 1. Projects → Intelligent Machines (One-to-Many)
- **Cascade Behavior**: Restrict deletion of projects with active machines
- **Indexing Strategy**: Composite index on (project_uuid, status)
- **Query Patterns**: Frequent joins for machine status by project

#### 2. Checkpoints → Parent Checkpoints (Self-Referential)
- **Tree Structure**: Represents checkpoint evolution chains
- **Materialized Path**: Consider for deep checkpoint hierarchies
- **Orphan Prevention**: Constraint ensures valid parent references

#### 3. Machines → Checkpoints (One-to-Many via creating_machine)
- **Lifecycle Tracking**: Machines create multiple checkpoints during operation
- **Soft Reference**: last_checkpoint_uuid for quick access to latest state
- **Cleanup Logic**: Cascade delete checkpoints when machine is purged

#### 4. Resource Usage → Projects/Machines (Many-to-One)
- **Partitioning**: Monthly partitions by recorded_at for performance
- **Aggregation Queries**: Heavy GROUP BY operations require proper indexing
- **Time Series Pattern**: Optimized for time-range queries and reporting

---

## 🏗️ DATABASE SCHEMA DESIGN

### Core Tables Schema Definition

#### Projects Table - Root Entity
```sql
-- UUID primary key with validation constraints
uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4()
name VARCHAR(255) NOT NULL UNIQUE
greediness DECIMAL(3,2) NOT NULL CHECK (greediness >= 0.0 AND greediness <= 1.0)
status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived'))

-- Soft references to checkpoints (no FK constraints)
initial_checkpoint_uuid UUID  -- First checkpoint for this project
current_checkpoint_uuid UUID  -- Latest stable checkpoint

-- Flexible schema extension
metadata JSONB DEFAULT '{}' NOT NULL

-- Temporal tracking
created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
```

**ORM Considerations:**
- **Soft Checkpoint References**: Avoid circular FK dependencies
- **Status Enum**: Consider database ENUM vs application validation  
- **Metadata Indexing**: GIN index on JSONB for query performance
- **Audit Trail**: Trigger-based updated_at maintenance

#### Intelligent Machines Table - Core Workflow Entity
```sql
uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4()
machine_type VARCHAR(50) NOT NULL -- claude_code, aider, qwen_coder, hub
status VARCHAR(50) NOT NULL DEFAULT 'spawning' 
project_uuid UUID NOT NULL REFERENCES projects(uuid) ON DELETE RESTRICT

-- Checkpoint relationships  
parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid)
last_checkpoint_uuid UUID -- Soft reference for performance

-- VM infrastructure mapping
vm_id INTEGER UNIQUE
hub_vm_id INTEGER -- Which hub spawned this machine

-- Execution context (JSONB for flexibility)
original_prompt TEXT
file_modifications JSONB DEFAULT '{}'
resource_constraints JSONB DEFAULT '{}'
resource_usage JSONB DEFAULT '{}'

-- Runtime state
current_progress TEXT
error_message TEXT
checkpoint_count INTEGER DEFAULT 0

-- Temporal tracking
spawned_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
terminated_at TIMESTAMPTZ
metadata JSONB DEFAULT '{}'
```

**ORM Considerations:**
- **Status State Machine**: Define valid state transitions in application layer
- **JSONB Query Patterns**: Index frequently queried JSON paths
- **Constraint Validation**: Check terminated_at > spawned_at
- **Soft Delete Pattern**: Consider status='deleted' vs hard deletion

#### Checkpoints Table - Immutable State Snapshots
```sql
uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4()
project_uuid UUID NOT NULL REFERENCES projects(uuid) ON DELETE CASCADE
parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid)
creating_machine_uuid UUID REFERENCES intelligent_machines(uuid)

-- Checkpoint categorization
checkpoint_type VARCHAR(50) NOT NULL 
-- genesis, automatic, manual, final, hub_daily

-- Infrastructure references
proxmox_snapshot_id VARCHAR(255) UNIQUE
vm_id INTEGER

-- File system state (JSONB for structured data)
file_manifest JSONB DEFAULT '{}'
-- Format: {"files": [{"path": "...", "size": 123, "sha256": "..."}]}

-- Metadata and verification
description TEXT
creation_context TEXT
size_bytes BIGINT
compression_ratio DECIMAL(5,4)
verification_status VARCHAR(50) DEFAULT 'pending'

-- Temporal tracking (immutable once created)
created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
verified_at TIMESTAMPTZ
metadata JSONB DEFAULT '{}'
```

**ORM Considerations:**
- **Immutability**: No updates after verification_status = 'verified'
- **Tree Queries**: Recursive CTEs for checkpoint ancestry/descendants
- **Storage Management**: Automated cleanup based on retention policies
- **Verification Workflow**: Async background verification jobs

#### Resource Usage Table - Time Series Data (Partitioned)
```sql
uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4()
project_uuid UUID NOT NULL REFERENCES projects(uuid)
machine_uuid UUID REFERENCES intelligent_machines(uuid)

-- Resource categorization
resource_type VARCHAR(50) NOT NULL -- money, time, api_calls, cpu, memory, storage
resource_subtype VARCHAR(100) -- claude_api, vm_runtime, storage_snapshot

-- Usage measurement
amount DECIMAL(12,4) NOT NULL
unit VARCHAR(20) NOT NULL -- CAD, USD, minutes, hours, calls, gb, gb_hours
cost_cad DECIMAL(10,4) -- Normalized for aggregation

-- Time series attributes
recorded_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
billing_period_start DATE
billing_period_end DATE
metadata JSONB DEFAULT '{}'

-- Partitioning: BY RANGE (recorded_at)
-- Monthly partitions: resource_usage_YYYY_MM
```

**ORM Considerations:**
- **Partition-Aware Queries**: ORM must handle partition pruning
- **Aggregation Performance**: Materialized views for common rollups
- **Time Zone Handling**: Store UTC, convert in application layer
- **Batch Inserts**: Optimize for high-volume time series writes

---

## 🔄 ORM RELATIONSHIP MAPPING PATTERNS

### SQLAlchemy Relationship Configurations

#### Project Entity Relationships
```python
# Conceptual ORM mapping patterns (implementation removed per request)

# Project.machines: One-to-Many
# - back_populates machines.project  
# - cascade="save-update, merge"
# - lazy="select" (default)

# Project.checkpoints: One-to-Many  
# - back_populates checkpoints.project
# - cascade="all, delete-orphan" 
# - lazy="select"

# Project.resource_usage: One-to-Many
# - back_populates resource_usage.project
# - cascade="all, delete-orphan"
# - lazy="dynamic" (for large collections)

# Project.current_checkpoint: Soft Reference
# - Foreign key but no relationship (avoid circular deps)
# - Load via separate query when needed
```

#### Machine Entity Relationships  
```python
# Machine.project: Many-to-One
# - back_populates project.machines
# - lazy="joined" (always load project)

# Machine.checkpoints: One-to-Many via creating_machine
# - back_populates checkpoint.creating_machine  
# - cascade="save-update, merge"
# - order_by="Checkpoint.created_at.desc()"

# Machine.parent_checkpoint: Many-to-One (optional)
# - back_populates checkpoint.spawned_machines
# - lazy="joined" 

# Machine.last_checkpoint: Soft Reference
# - No relationship mapping
# - Managed via application logic
```

#### Checkpoint Entity Relationships
```python
# Checkpoint.project: Many-to-One
# - back_populates project.checkpoints
# - lazy="joined"

# Checkpoint.parent_checkpoint: Self-Referential Many-to-One
# - back_populates checkpoint.child_checkpoints
# - remote_side=["uuid"] for self-reference

# Checkpoint.child_checkpoints: Self-Referential One-to-Many  
# - back_populates checkpoint.parent_checkpoint
# - cascade="save-update, merge"

# Checkpoint.creating_machine: Many-to-One (optional)
# - back_populates machine.checkpoints
# - lazy="select"
```

---

## 📈 QUERY OPTIMIZATION PATTERNS

### Indexing Strategy

#### Primary Performance Indexes
```sql
-- Project queries
CREATE INDEX idx_projects_greediness ON projects(greediness DESC);
CREATE INDEX idx_projects_status ON projects(status) WHERE status = 'active';
CREATE UNIQUE INDEX idx_projects_name ON projects(name);

-- Machine queries  
CREATE INDEX idx_machines_project_status ON intelligent_machines(project_uuid, status);
CREATE INDEX idx_machines_type_status ON intelligent_machines(machine_type, status);
CREATE INDEX idx_machines_spawn_time ON intelligent_machines(spawned_at DESC);
CREATE INDEX idx_machines_vm ON intelligent_machines(vm_id) WHERE vm_id IS NOT NULL;

-- Checkpoint queries
CREATE INDEX idx_checkpoints_project ON checkpoints(project_uuid, created_at DESC);
CREATE INDEX idx_checkpoints_parent ON checkpoints(parent_checkpoint_uuid);
CREATE INDEX idx_checkpoints_machine ON checkpoints(creating_machine_uuid);
CREATE INDEX idx_checkpoints_type ON checkpoints(checkpoint_type, created_at DESC);

-- Resource usage queries (per partition)
CREATE INDEX idx_resource_project_time ON resource_usage(project_uuid, recorded_at);
CREATE INDEX idx_resource_type_time ON resource_usage(resource_type, recorded_at);
CREATE INDEX idx_resource_machine ON resource_usage(machine_uuid) WHERE machine_uuid IS NOT NULL;
```

#### JSONB Indexing Patterns
```sql
-- Metadata field queries
CREATE INDEX idx_projects_metadata_gin ON projects USING GIN (metadata);
CREATE INDEX idx_machines_constraints_gin ON intelligent_machines USING GIN (resource_constraints);
CREATE INDEX idx_checkpoints_manifest_gin ON checkpoints USING GIN (file_manifest);

-- Specific JSONB path queries (if needed)
CREATE INDEX idx_machines_budget ON intelligent_machines 
  USING BTREE ((resource_constraints->>'daily_budget')::DECIMAL);
```

### Common Query Patterns

#### Project Resource Aggregation
```sql
-- Daily resource usage by project (materialized view candidate)
SELECT 
    p.uuid,
    p.name, 
    DATE(ru.recorded_at) as usage_date,
    ru.resource_type,
    SUM(ru.amount) as total_amount,
    SUM(ru.cost_cad) as total_cost_cad
FROM projects p
JOIN resource_usage ru ON p.uuid = ru.project_uuid
WHERE ru.recorded_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY p.uuid, p.name, DATE(ru.recorded_at), ru.resource_type
ORDER BY p.name, usage_date DESC;
```

#### Machine Status Dashboard
```sql
-- Active machines by project with latest checkpoint info
SELECT 
    p.name as project_name,
    im.machine_type,
    im.status,
    im.current_progress,
    im.spawned_at,
    c.created_at as last_checkpoint_at,
    c.checkpoint_type
FROM projects p
JOIN intelligent_machines im ON p.uuid = im.project_uuid  
LEFT JOIN checkpoints c ON im.last_checkpoint_uuid = c.uuid
WHERE im.status IN ('active', 'working', 'checkpointing')
ORDER BY p.name, im.spawned_at;
```

#### Checkpoint Evolution Tree
```sql
-- Recursive query for checkpoint ancestry
WITH RECURSIVE checkpoint_tree AS (
    SELECT uuid, parent_checkpoint_uuid, description, created_at, 0 as depth
    FROM checkpoints 
    WHERE parent_checkpoint_uuid IS NULL
    
    UNION ALL
    
    SELECT c.uuid, c.parent_checkpoint_uuid, c.description, c.created_at, ct.depth + 1
    FROM checkpoints c
    JOIN checkpoint_tree ct ON c.parent_checkpoint_uuid = ct.uuid
)
SELECT * FROM checkpoint_tree ORDER BY depth, created_at;
```

---

## ⚡ PERFORMANCE OPTIMIZATION

### Materialized Views for Analytics

#### Project Resource Summary View
```sql
CREATE MATERIALIZED VIEW mv_project_resource_summary AS
SELECT 
    ru.project_uuid,
    p.name as project_name,
    p.greediness,
    DATE(ru.recorded_at) as usage_date,
    ru.resource_type,
    SUM(ru.amount) as total_amount,
    AVG(ru.amount) as avg_amount, 
    SUM(ru.cost_cad) as total_cost_cad,
    COUNT(*) as transaction_count
FROM resource_usage ru
JOIN projects p ON ru.project_uuid = p.uuid
WHERE ru.recorded_at >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY ru.project_uuid, p.name, p.greediness, DATE(ru.recorded_at), ru.resource_type;

CREATE UNIQUE INDEX idx_mv_project_summary 
ON mv_project_resource_summary(project_uuid, usage_date, resource_type);

-- Refresh schedule (via pg_cron)
SELECT cron.schedule('refresh-project-summary', '0 1 * * *', 
  'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_project_resource_summary;');
```

#### Machine Performance Metrics View
```sql
CREATE MATERIALIZED VIEW mv_machine_performance AS
SELECT 
    im.machine_type,
    COUNT(*) as total_machines,
    AVG(EXTRACT(EPOCH FROM (im.terminated_at - im.spawned_at))/60)::INTEGER as avg_runtime_minutes,
    COUNT(*) FILTER (WHERE im.status = 'terminated' AND im.error_message IS NULL) as successful_count,
    ROUND(
        100.0 * COUNT(*) FILTER (WHERE im.status = 'terminated' AND im.error_message IS NULL) / 
        COUNT(*) FILTER (WHERE im.status = 'terminated'), 2
    ) as success_rate_percent,
    AVG((im.resource_usage->>'money')::DECIMAL) as avg_money_cost,
    AVG((im.resource_usage->>'api_calls')::INTEGER) as avg_api_calls,
    AVG(im.checkpoint_count) as avg_checkpoints_created
FROM intelligent_machines im
WHERE im.spawned_at >= CURRENT_DATE - INTERVAL '30 days'
  AND im.machine_type != 'hub'  -- Exclude hub machines
GROUP BY im.machine_type;

-- Refresh daily  
SELECT cron.schedule('refresh-machine-performance', '0 2 * * *',
  'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_machine_performance;');
```

### Partitioning Strategy

#### Resource Usage Partitioning
```sql
-- Automatic monthly partition creation
CREATE OR REPLACE FUNCTION create_monthly_partition(table_name TEXT, start_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    end_date DATE;
BEGIN
    partition_name := table_name || '_' || TO_CHAR(start_date, 'YYYY_MM');
    end_date := start_date + INTERVAL '1 month';
    
    EXECUTE FORMAT('CREATE TABLE %I PARTITION OF %I 
                   FOR VALUES FROM (%L) TO (%L)',
                   partition_name, table_name, start_date, end_date);
                   
    EXECUTE FORMAT('CREATE INDEX idx_%I_project_time 
                   ON %I(project_uuid, recorded_at)',
                   partition_name, partition_name);
END;
$$ LANGUAGE plpgsql;

-- Schedule partition creation for next month
SELECT cron.schedule('create-next-partition', '0 0 25 * *',
  'SELECT create_monthly_partition(''resource_usage'', DATE_TRUNC(''month'', CURRENT_DATE + INTERVAL ''1 month''));');
```

### Connection Pooling & Caching Strategy

#### Database Connection Configuration
```yaml
# SQLAlchemy Engine Configuration
Pool Size: 20 connections
Max Overflow: 30 connections  
Pool Timeout: 30 seconds
Pool Recycle: 3600 seconds (1 hour)
Pool Pre-ping: true (connection validation)

# Redis Query Cache Configuration  
Default TTL: 300 seconds (5 minutes)
Max Memory: 1GB
Eviction Policy: allkeys-lru
Key Prefix: autovibe:query_cache:
```

#### ORM Query Caching Patterns
- **Project Lists**: Cache for 5 minutes (frequently accessed, rarely changed)
- **Machine Status**: Cache for 30 seconds (rapidly changing)  
- **Checkpoint Details**: Cache for 1 hour (immutable after verification)
- **Resource Summaries**: Cache for 15 minutes (updated periodically)

---

## 🛡️ DATA INTEGRITY & CONSTRAINTS

### Database Triggers for Business Logic

#### Budget Enforcement Trigger
```sql
CREATE OR REPLACE FUNCTION enforce_resource_budget()
RETURNS TRIGGER AS $$
DECLARE
    current_budget DECIMAL;
    current_usage DECIMAL; 
    budget_limit DECIMAL;
BEGIN
    -- Get current daily budget for this resource type
    SELECT allocated_amount, used_amount 
    INTO budget_limit, current_usage
    FROM resource_budgets rb
    WHERE rb.project_uuid = NEW.project_uuid 
      AND rb.resource_type = NEW.resource_type
      AND rb.budget_period = 'daily'
      AND rb.period_start <= CURRENT_DATE 
      AND rb.period_end >= CURRENT_DATE;
      
    -- Check if adding new usage exceeds budget
    IF (current_usage + NEW.amount) > budget_limit THEN
        RAISE EXCEPTION 'Resource budget exceeded: % limit is %, current usage is %, attempted addition is %',
            NEW.resource_type, budget_limit, current_usage, NEW.amount;
    END IF;
    
    -- Update budget usage
    UPDATE resource_budgets 
    SET used_amount = used_amount + NEW.amount,
        updated_at = NOW()
    WHERE project_uuid = NEW.project_uuid 
      AND resource_type = NEW.resource_type
      AND budget_period = 'daily'
      AND period_start <= CURRENT_DATE 
      AND period_end >= CURRENT_DATE;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_budget_before_insert
    BEFORE INSERT ON resource_usage
    FOR EACH ROW EXECUTE FUNCTION enforce_resource_budget();
```

#### Checkpoint Verification Queue Trigger
```sql
CREATE OR REPLACE FUNCTION queue_checkpoint_verification()
RETURNS TRIGGER AS $$
BEGIN
    -- Add to verification queue (implemented via NOTIFY or job table)
    PERFORM pg_notify('checkpoint_verification', 
                     json_build_object('checkpoint_uuid', NEW.uuid, 
                                      'created_at', NEW.created_at)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER queue_verification_after_insert
    AFTER INSERT ON checkpoints
    FOR EACH ROW EXECUTE FUNCTION queue_checkpoint_verification();
```

### Constraint Validation Patterns

#### Business Logic Constraints
```sql
-- Machine state transitions
ALTER TABLE intelligent_machines 
ADD CONSTRAINT valid_status_values 
CHECK (status IN ('spawning', 'active', 'working', 'checkpointing', 'terminated', 'error'));

-- Resource amount validation
ALTER TABLE resource_usage 
ADD CONSTRAINT positive_resource_amount 
CHECK (amount > 0);

-- Checkpoint time ordering
ALTER TABLE checkpoints
ADD CONSTRAINT valid_verification_time
CHECK (verified_at IS NULL OR verified_at >= created_at);

-- Project greediness bounds  
ALTER TABLE projects
ADD CONSTRAINT valid_greediness_range
CHECK (greediness >= 0.0 AND greediness <= 1.0);
```

---

This ORM-focused execution plan provides the data architecture foundation for the autovibe system, emphasizing relationship design, query optimization, and data integrity patterns while removing implementation code as requested.