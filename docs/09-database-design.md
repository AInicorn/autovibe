# Database Design - PostgreSQL Schema

## Overview

autovibe uses PostgreSQL as its primary data store for metadata, resource tracking, and system state management. The schema is designed for scalability, performance, and comprehensive audit trails.

## Core Tables

### Projects
The Projects table stores project metadata using UUID primary keys, with required name and greediness factor (0-1 decimal with validation). The table tracks project status (default 'active'), description, initial and current checkpoint references, flexible metadata in JSONB format, and timestamps. Indexes optimize queries on greediness (descending), status, and metadata (GIN index). An update trigger automatically maintains the updated_at timestamp.

### Intelligent Machines
```sql
CREATE TABLE intelligent_machines (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- 'claude_code', 'aider', 'qwen_coder', 'hub'
    status VARCHAR(50) NOT NULL DEFAULT 'spawning', -- 'spawning', 'active', 'working', 'checkpointing', 'terminated'
    project_uuid UUID REFERENCES projects(uuid),
    parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    spawned_by_hub_uuid UUID REFERENCES intelligent_machines(uuid),
    vm_id VARCHAR(100),
    vm_node VARCHAR(100),
    prompt TEXT,
    file_modifications JSONB DEFAULT '{}',
    resource_constraints JSONB DEFAULT '{}',
    resource_usage JSONB DEFAULT '{}', -- Real-time resource tracking
    progress TEXT,
    error_message TEXT,
    last_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    checkpoints_created INTEGER DEFAULT 0,
    spawned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    terminated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_machines_status ON intelligent_machines(status);
CREATE INDEX idx_machines_project ON intelligent_machines(project_uuid);
CREATE INDEX idx_machines_spawned_at ON intelligent_machines(spawned_at);
CREATE INDEX idx_machines_type ON intelligent_machines(type);
CREATE INDEX idx_machines_hub ON intelligent_machines(spawned_by_hub_uuid);
```

### Checkpoints
```sql
CREATE TABLE checkpoints (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    created_by_machine_uuid UUID REFERENCES intelligent_machines(uuid),
    checkpoint_type VARCHAR(50) NOT NULL, -- 'genesis', 'automatic', 'manual', 'final', 'hub_daily'
    proxmox_snapshot_id VARCHAR(200),
    vm_id VARCHAR(100),
    file_manifest JSONB, -- {filename: {size, sha256, modified_at}}
    description TEXT,
    context VARCHAR(200), -- 'api_response', 'file_operation', 'manual', etc.
    size_bytes BIGINT,
    compression_ratio DECIMAL(4,2),
    verification_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_checkpoints_project ON checkpoints(project_uuid);
CREATE INDEX idx_checkpoints_parent ON checkpoints(parent_checkpoint_uuid);
CREATE INDEX idx_checkpoints_created_at ON checkpoints(created_at DESC);
CREATE INDEX idx_checkpoints_type ON checkpoints(checkpoint_type);
CREATE INDEX idx_checkpoints_machine ON checkpoints(created_by_machine_uuid);

-- Partial index for active checkpoints
CREATE INDEX idx_checkpoints_active ON checkpoints(project_uuid, created_at DESC) 
    WHERE verification_status = 'verified';
```

### Resource Usage Tracking
```sql
CREATE TABLE resource_usage (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    resource_type VARCHAR(50) NOT NULL, -- 'money', 'time', 'api_calls', 'cpu', 'memory', 'storage', 'network'
    resource_subtype VARCHAR(50), -- 'claude_api', 'vm_runtime', 'storage_snapshot', etc.
    amount DECIMAL(15,6) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- 'CAD', 'USD', 'minutes', 'hours', 'calls', 'gb', 'gb_hours'
    cost_cad DECIMAL(10,4), -- Normalized to CAD for aggregation
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    billing_period DATE, -- For aggregation
    metadata JSONB DEFAULT '{}'
);

-- Partition by billing_period for performance
CREATE INDEX idx_resource_usage_project_period ON resource_usage(project_uuid, billing_period);
CREATE INDEX idx_resource_usage_machine ON resource_usage(machine_uuid);
CREATE INDEX idx_resource_usage_type ON resource_usage(resource_type, resource_subtype);
CREATE INDEX idx_resource_usage_recorded_at ON resource_usage(recorded_at);

-- Monthly partitioning
CREATE TABLE resource_usage_2024_08 PARTITION OF resource_usage
    FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');
```

### Resource Budgets
```sql
CREATE TABLE resource_budgets (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID NOT NULL REFERENCES projects(uuid),
    resource_type VARCHAR(50) NOT NULL,
    budget_period VARCHAR(50) NOT NULL, -- 'daily', 'weekly', 'monthly', 'total'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    allocated_amount DECIMAL(15,6) NOT NULL,
    used_amount DECIMAL(15,6) DEFAULT 0,
    reserved_amount DECIMAL(15,6) DEFAULT 0, -- For pending operations
    unit VARCHAR(20) NOT NULL,
    auto_reset BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(project_uuid, resource_type, budget_period, period_start)
);

CREATE INDEX idx_budgets_project_period ON resource_budgets(project_uuid, period_start, period_end);
CREATE INDEX idx_budgets_active ON resource_budgets(period_start, period_end) 
    WHERE period_end >= CURRENT_DATE;
```

## Hub Management Tables

### Hub Instances
```sql
CREATE TABLE hub_instances (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL, -- 'primary', 'development', 'experimental', 'domain_specific'
    status VARCHAR(50) NOT NULL, -- 'active', 'inactive', 'experimental', 'retiring'
    vm_id VARCHAR(100) NOT NULL,
    vm_node VARCHAR(100),
    version VARCHAR(50),
    configuration JSONB DEFAULT '{}',
    parent_hub_uuid UUID REFERENCES hub_instances(uuid),
    experiment_id UUID REFERENCES hub_experiments(uuid),
    spawned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_checkpoint_at TIMESTAMP WITH TIME ZONE,
    terminated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_hub_instances_status ON hub_instances(status);
CREATE INDEX idx_hub_instances_type ON hub_instances(type);
```

### Hub Checkpoints
```sql
CREATE TABLE hub_checkpoints (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hub_uuid UUID NOT NULL REFERENCES hub_instances(uuid),
    checkpoint_type VARCHAR(50) NOT NULL, -- 'daily', 'manual', 'pre_experiment', 'recovery'
    proxmox_snapshot_id VARCHAR(200) NOT NULL,
    trigger_reason VARCHAR(200),
    system_metrics JSONB, -- CPU, memory, disk usage at checkpoint time
    active_machines INTEGER DEFAULT 0,
    active_projects INTEGER DEFAULT 0,
    checkpoint_size_gb DECIMAL(10,3),
    verification_status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_hub_checkpoints_hub ON hub_checkpoints(hub_uuid);
CREATE INDEX idx_hub_checkpoints_created_at ON hub_checkpoints(created_at DESC);
CREATE INDEX idx_hub_checkpoints_type ON hub_checkpoints(checkpoint_type);
```

### Hub Experiments
```sql
CREATE TABLE hub_experiments (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    base_checkpoint_uuid UUID NOT NULL REFERENCES hub_checkpoints(uuid),
    original_hub_uuid UUID NOT NULL REFERENCES hub_instances(uuid),
    experimental_hub_uuid UUID REFERENCES hub_instances(uuid),
    configuration_changes JSONB NOT NULL,
    success_metrics JSONB, -- Define what constitutes success
    performance_data JSONB, -- Collected performance metrics
    status VARCHAR(50) DEFAULT 'planned', -- 'planned', 'running', 'completed', 'failed'
    duration_days INTEGER DEFAULT 14,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    result VARCHAR(50), -- 'experimental_wins', 'original_wins', 'inconclusive'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_hub_experiments_status ON hub_experiments(status);
CREATE INDEX idx_hub_experiments_started_at ON hub_experiments(started_at);
```

## Performance and Analytics Views

### Project Resource Summary (Materialized View)
```sql
CREATE MATERIALIZED VIEW project_resource_summary AS
SELECT 
    p.uuid as project_uuid,
    p.name as project_name,
    p.greediness,
    DATE(ru.recorded_at) as usage_date,
    ru.resource_type,
    SUM(ru.amount) as total_amount,
    ru.unit,
    SUM(ru.cost_cad) as total_cost_cad,
    COUNT(*) as transaction_count,
    AVG(ru.amount) as avg_amount
FROM projects p
JOIN resource_usage ru ON p.uuid = ru.project_uuid
WHERE ru.recorded_at >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY p.uuid, p.name, p.greediness, DATE(ru.recorded_at), ru.resource_type, ru.unit;

CREATE UNIQUE INDEX idx_project_summary_unique 
    ON project_resource_summary(project_uuid, usage_date, resource_type);

-- Refresh schedule
CREATE EXTENSION IF NOT EXISTS pg_cron;
SELECT cron.schedule('refresh-project-summary', '0 1 * * *', 'REFRESH MATERIALIZED VIEW project_resource_summary;');
```

### Machine Performance Metrics
```sql
CREATE MATERIALIZED VIEW machine_performance_metrics AS
SELECT 
    im.type as machine_type,
    COUNT(*) as total_machines,
    AVG(EXTRACT(EPOCH FROM (im.terminated_at - im.spawned_at))/60) as avg_runtime_minutes,
    COUNT(*) FILTER (WHERE im.status = 'terminated' AND im.error_message IS NULL) as successful_machines,
    ROUND(
        COUNT(*) FILTER (WHERE im.status = 'terminated' AND im.error_message IS NULL)::DECIMAL / COUNT(*) * 100, 
        2
    ) as success_rate_percent,
    AVG((im.resource_usage->>'money')::DECIMAL) as avg_money_cost,
    AVG((im.resource_usage->>'api_calls')::INTEGER) as avg_api_calls,
    AVG(im.checkpoints_created) as avg_checkpoints_created
FROM intelligent_machines im
WHERE im.spawned_at >= CURRENT_DATE - INTERVAL '30 days'
    AND im.type != 'hub'
GROUP BY im.type;

CREATE UNIQUE INDEX idx_machine_perf_type ON machine_performance_metrics(machine_type);
```

## Database Functions and Triggers

### Resource Budget Enforcement
```sql
CREATE OR REPLACE FUNCTION enforce_resource_budget()
RETURNS TRIGGER AS $$
DECLARE
    current_budget RECORD;
    budget_exceeded BOOLEAN := FALSE;
BEGIN
    -- Get current budget for this resource type
    SELECT allocated_amount, used_amount, unit INTO current_budget
    FROM resource_budgets rb
    WHERE rb.project_uuid = NEW.project_uuid
        AND rb.resource_type = NEW.resource_type
        AND CURRENT_DATE BETWEEN rb.period_start AND rb.period_end
        AND rb.budget_period = 'daily';
    
    -- Check if adding this usage would exceed budget
    IF FOUND THEN
        IF (current_budget.used_amount + NEW.amount) > current_budget.allocated_amount THEN
            RAISE EXCEPTION 'Resource budget exceeded for project % resource %: current=%, adding=%, budget=%',
                NEW.project_uuid, NEW.resource_type, current_budget.used_amount, 
                NEW.amount, current_budget.allocated_amount
                USING ERRCODE = 'check_violation';
        END IF;
        
        -- Update budget usage
        UPDATE resource_budgets 
        SET used_amount = used_amount + NEW.amount,
            updated_at = NOW()
        WHERE project_uuid = NEW.project_uuid
            AND resource_type = NEW.resource_type
            AND budget_period = 'daily'
            AND CURRENT_DATE BETWEEN period_start AND period_end;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_budget_before_insert
    BEFORE INSERT ON resource_usage
    FOR EACH ROW
    EXECUTE FUNCTION enforce_resource_budget();
```

### Checkpoint Verification
```sql
CREATE OR REPLACE FUNCTION verify_checkpoint_integrity()
RETURNS TRIGGER AS $$
BEGIN
    -- Schedule asynchronous verification
    INSERT INTO checkpoint_verification_queue (checkpoint_uuid, queued_at)
    VALUES (NEW.uuid, NOW());
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER schedule_checkpoint_verification
    AFTER INSERT ON checkpoints
    FOR EACH ROW
    EXECUTE FUNCTION verify_checkpoint_integrity();
```

## Database Maintenance and Optimization

### Automated Cleanup Jobs
```sql
-- Clean up old resource usage records (keep 1 year)
CREATE OR REPLACE FUNCTION cleanup_old_resource_usage()
RETURNS void AS $$
BEGIN
    DELETE FROM resource_usage 
    WHERE recorded_at < CURRENT_DATE - INTERVAL '1 year';
    
    -- Log cleanup results
    INSERT INTO maintenance_log (operation, affected_rows, performed_at)
    VALUES ('cleanup_resource_usage', ROW_COUNT, NOW());
END;
$$ LANGUAGE plpgsql;

-- Schedule via pg_cron
SELECT cron.schedule('cleanup-resource-usage', '0 2 1 * *', 'SELECT cleanup_old_resource_usage();');
```

### Backup and Recovery Considerations
```sql
-- Critical tables for backup priority
-- 1. projects (small, critical)
-- 2. checkpoints (large, critical)
-- 3. intelligent_machines (medium, important)
-- 4. resource_usage (large, important for billing)
-- 5. hub_instances (small, critical)

-- Recovery point objectives:
-- - Projects: 0 data loss acceptable
-- - Checkpoints: 15 minutes acceptable
-- - Resource usage: 1 hour acceptable (can be reconstructed from logs)
```

This PostgreSQL schema provides robust data management for autovibe with comprehensive audit trails, resource tracking, and performance optimization capabilities.