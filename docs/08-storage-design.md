# AutoVibe Storage Design

## PostgreSQL Database Architecture

AutoVibe uses PostgreSQL as its primary data store for robust, scalable, and ACID-compliant data management.

## Core Database Schema

### Projects Table
```sql
CREATE TABLE projects (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    greediness DECIMAL(3,2) CHECK (greediness >= 0 AND greediness <= 1),
    current_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    repository_url TEXT,  -- Flexible repository linking
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_projects_greediness ON projects(greediness);
CREATE INDEX idx_projects_metadata ON projects USING GIN(metadata);
```

### Intelligent Machines Table
```sql
CREATE TABLE intelligent_machines (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,  -- CLAUDE_CODE, AIDER, QWEN_CODER
    status VARCHAR(50) NOT NULL,  -- SPAWNING, ACTIVE, WORKING, TERMINATING, TERMINATED
    project_uuid UUID REFERENCES projects(uuid),
    checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    vm_id VARCHAR(255),
    prompt TEXT NOT NULL,
    file_modifications JSONB DEFAULT '{}',
    resource_usage JSONB DEFAULT '{}',  -- {money, time, api_calls, compute}
    progress TEXT,
    spawned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    terminated_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_machines_status ON intelligent_machines(status);
CREATE INDEX idx_machines_project ON intelligent_machines(project_uuid);
CREATE INDEX idx_machines_spawned ON intelligent_machines(spawned_at);
```

### Checkpoints Table (VM Snapshots)
```sql
CREATE TABLE checkpoints (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID REFERENCES projects(uuid),
    parent_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    vm_snapshot_path TEXT NOT NULL,  -- Path to actual VM snapshot file
    file_manifest JSONB NOT NULL,  -- List of files and their hashes
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    size_bytes BIGINT,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_checkpoints_project ON checkpoints(project_uuid);
CREATE INDEX idx_checkpoints_parent ON checkpoints(parent_checkpoint_uuid);
CREATE INDEX idx_checkpoints_created ON checkpoints(created_at);
```

### Resource Usage Table
```sql
CREATE TABLE resource_usage (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID REFERENCES projects(uuid),
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    resource_type VARCHAR(50) NOT NULL,  -- MONEY, TIME, API_CALLS, CPU, MEMORY, DISK
    amount DECIMAL(15,4) NOT NULL,
    unit VARCHAR(20) NOT NULL,  -- CAD, minutes, calls, gb-hours, etc
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_resource_usage_project ON resource_usage(project_uuid);
CREATE INDEX idx_resource_usage_machine ON resource_usage(machine_uuid);
CREATE INDEX idx_resource_usage_timestamp ON resource_usage(timestamp);
CREATE INDEX idx_resource_usage_type ON resource_usage(resource_type);
```

### Budget Allocations Table
```sql
CREATE TABLE budget_allocations (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID REFERENCES projects(uuid),
    resource_type VARCHAR(50) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    allocated_amount DECIMAL(15,4) NOT NULL,
    spent_amount DECIMAL(15,4) DEFAULT 0,
    unit VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_uuid, resource_type, period_start)
);

CREATE INDEX idx_budget_allocations_project ON budget_allocations(project_uuid);
CREATE INDEX idx_budget_allocations_period ON budget_allocations(period_start, period_end);
```

### Evolution History Table
```sql
CREATE TABLE evolution_history (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_uuid UUID REFERENCES projects(uuid),
    from_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    to_checkpoint_uuid UUID REFERENCES checkpoints(uuid),
    machine_uuid UUID REFERENCES intelligent_machines(uuid),
    prompt TEXT NOT NULL,
    resource_usage JSONB NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_evolution_project ON evolution_history(project_uuid);
CREATE INDEX idx_evolution_created ON evolution_history(created_at);
```

## Advanced Storage Features

### Partitioning Strategy
For large-scale deployments, partition tables by time:
```sql
-- Partition resource_usage by month
CREATE TABLE resource_usage_2024_08 PARTITION OF resource_usage
    FOR VALUES FROM ('2024-08-01') TO ('2024-09-01');

-- Partition evolution_history by quarter
CREATE TABLE evolution_history_2024_q3 PARTITION OF evolution_history
    FOR VALUES FROM ('2024-07-01') TO ('2024-10-01');
```

### Materialized Views for Analytics
```sql
-- Daily resource usage summary
CREATE MATERIALIZED VIEW daily_resource_summary AS
SELECT 
    project_uuid,
    DATE(timestamp) as date,
    resource_type,
    SUM(amount) as total_amount,
    unit,
    COUNT(*) as transaction_count
FROM resource_usage
GROUP BY project_uuid, DATE(timestamp), resource_type, unit;

CREATE UNIQUE INDEX ON daily_resource_summary(project_uuid, date, resource_type);

-- Project evolution tree view
CREATE MATERIALIZED VIEW evolution_tree AS
WITH RECURSIVE tree AS (
    SELECT 
        c.uuid,
        c.project_uuid,
        c.parent_checkpoint_uuid,
        c.description,
        0 as depth,
        ARRAY[c.uuid] as path
    FROM checkpoints c
    WHERE c.parent_checkpoint_uuid IS NULL
    
    UNION ALL
    
    SELECT 
        c.uuid,
        c.project_uuid,
        c.parent_checkpoint_uuid,
        c.description,
        t.depth + 1,
        t.path || c.uuid
    FROM checkpoints c
    JOIN tree t ON c.parent_checkpoint_uuid = t.uuid
)
SELECT * FROM tree;
```

### Triggers for Data Consistency
```sql
-- Update project's updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Cascade resource usage to budget allocations
CREATE OR REPLACE FUNCTION update_budget_spent()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE budget_allocations
    SET spent_amount = spent_amount + NEW.amount
    WHERE project_uuid = NEW.project_uuid
      AND resource_type = NEW.resource_type
      AND NEW.timestamp BETWEEN period_start AND period_end;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cascade_resource_to_budget
    AFTER INSERT ON resource_usage
    FOR EACH ROW
    EXECUTE FUNCTION update_budget_spent();
```

## Connection Management

### Connection Pooling
Use PgBouncer or connection pooling libraries:
- Pool size: 20-50 connections
- Max client connections: 100
- Connection timeout: 30 seconds
- Statement timeout: 120 seconds for regular queries
- Long query timeout: 600 seconds for analytics

### Read Replicas
For scaling read operations:
- Primary: All writes
- Replica 1: Real-time analytics
- Replica 2: Backup and batch processing

## Backup Strategy

### Continuous Archiving
- WAL archiving to S3/object storage
- Point-in-time recovery capability
- Retention: 30 days of WAL files

### Regular Backups
- Daily full backups via pg_dump
- Hourly incremental backups
- Weekly logical backups for long-term storage
- Monthly archives to cold storage

## Performance Optimization

### Indexing Strategy
- B-tree indexes for equality/range queries
- GIN indexes for JSONB fields
- BRIN indexes for time-series data
- Partial indexes for filtered queries

### Query Optimization
- Use EXPLAIN ANALYZE for query planning
- Implement query result caching
- Use prepared statements for frequent queries
- Batch insert operations

### Vacuum and Maintenance
- Auto-vacuum enabled with custom thresholds
- Weekly VACUUM ANALYZE
- Monthly REINDEX for heavily updated tables
- Quarterly full VACUUM

## Migration Strategy

### Schema Versioning
Use migration tools like Flyway or Alembic:
- Version-controlled migrations
- Rollback capabilities
- Pre/post migration hooks

### Zero-Downtime Migrations
- Use CREATE INDEX CONCURRENTLY
- Add columns with defaults as nullable first
- Rename in multiple steps
- Use feature flags for application changes

## Monitoring and Metrics

### Key Metrics to Track
- Connection pool usage
- Query response times
- Table and index sizes
- Vacuum and checkpoint frequency
- Replication lag
- Lock contention

### Alerting Thresholds
- Connection pool > 80% utilized
- Query time > 5 seconds
- Replication lag > 1 minute
- Disk usage > 80%
- Deadlock detection

## Security Considerations

### Access Control
- Role-based access control (RBAC)
- Separate roles for application, analytics, admin
- Row-level security for multi-tenant data
- SSL/TLS for all connections

### Data Encryption
- Encryption at rest using LUKS or cloud provider encryption
- Encryption in transit via SSL/TLS
- Column-level encryption for sensitive data
- Backup encryption

### Audit Logging
- Enable pgaudit extension
- Log all DDL operations
- Log sensitive data access
- Regular audit log review

## Repository Linking

Projects can be flexibly linked to different repositories:
- Repository URL stored in projects table
- Can point to this repo initially
- Can be updated to point to external repos
- Supports multiple repository providers (GitHub, GitLab, Bitbucket)

This comprehensive PostgreSQL design provides a robust foundation for AutoVibe's data management needs while maintaining flexibility for future scaling and optimization.