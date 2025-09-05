# autovibe REST API Design

## API Architecture

The Hub exposes a comprehensive REST API for managing Intelligent Machines, projects, checkpoints, and resources. All endpoints use UUID-based identification and support real-time resource tracking.

### Base URL
```
https://hub.autovibe.local/api/v1
```

### Authentication
All API requests require authentication:
```bash
curl -H "Authorization: Bearer ${AUTOVIBE_API_KEY}" \
     -H "Content-Type: application/json" \
     https://hub.autovibe.local/api/v1/machines
```

## Machine Management Endpoints

### Spawn Intelligent Machine
```http
POST /machines
{
  "type": "claude_code",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "checkpoint_uuid": "checkpoint-abc123-def456",
  "prompt": "Add user authentication system",
  "file_modifications": {
    "config/auth.json": "new OAuth configuration content",
    "src/auth.py": "updated authentication module"
  },
  "resource_constraints": {
    "max_money": 25.0,
    "max_time_minutes": 120,
    "max_api_calls": 50
  }
}

Response:
{
  "machine_uuid": "machine-789xyz-012abc",
  "vm_id": "vm-001",
  "status": "spawning",
  "estimated_resources": {
    "money": 20.0,
    "time_minutes": 90,
    "api_calls": 35
  },
  "spawned_at": "2024-08-24T10:30:00Z"
}
```

### Get Machine Status
```http
GET /machines/{machine_uuid}

Response:
{
  "uuid": "machine-789xyz-012abc",
  "type": "claude_code", 
  "status": "working",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "vm_id": "vm-001",
  "prompt": "Add user authentication system",
  "resource_usage": {
    "money": {"estimated": 25.0, "actual": 18.75, "unit": "CAD"},
    "time": {"estimated": 120, "actual": 95, "unit": "minutes"},
    "api_calls": {"estimated": 50, "actual": 32, "unit": "calls"},
    "compute": {"cpu_hours": 1.2, "memory_gb_hours": 2.8}
  },
  "progress": "Analyzing existing authentication patterns...",
  "last_checkpoint": "checkpoint-def456-789abc",
  "checkpoints_created": 3,
  "spawned_at": "2024-08-24T10:30:00Z",
  "updated_at": "2024-08-24T11:15:00Z"
}
```

### List Machines
```http
GET /machines?status=working&project_uuid={project_uuid}&limit=20

Response:
{
  "machines": [...],
  "total_count": 45,
  "has_more": true,
  "next_offset": 20
}
```

### Terminate Machine
```http
POST /machines/{machine_uuid}/terminate
{
  "reason": "user_requested",
  "create_checkpoint": true
}

Response:
{
  "status": "terminated",
  "final_checkpoint": "checkpoint-final-xyz789",
  "resource_usage": {...},
  "terminated_at": "2024-08-24T11:45:00Z"
}
```

### Request Manual Checkpoint
```http
POST /machines/{machine_uuid}/checkpoint
{
  "context": "before_major_refactor",
  "description": "Safe checkpoint before restructuring authentication"
}

Response:
{
  "checkpoint_uuid": "checkpoint-manual-abc123",
  "created_at": "2024-08-24T11:20:00Z",
  "size_mb": 245
}
```

## Project Management Endpoints

### Create Project
```http
POST /projects
{
  "name": "autovibe-core",
  "greediness": 0.42,
  "description": "Core autovibe system development",
  "resource_budgets": {
    "daily_money": 3.0,
    "daily_time_hours": 8,
    "daily_api_calls": 500,
    "storage_gb": 100
  }
}

Response:
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "autovibe-core",
  "greediness": 0.42,
  "initial_checkpoint": "checkpoint-genesis-000000",
  "created_at": "2024-08-24T10:00:00Z"
}
```

### Get Project Details
```http
GET /projects/{project_uuid}

Response:
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "autovibe-core",
  "greediness": 0.42,
  "current_checkpoint": "checkpoint-latest-abc123",
  "resource_usage_today": {
    "money": {"allocated": 3.0, "used": 1.85, "remaining": 1.15},
    "time": {"allocated": 8, "used": 3.2, "remaining": 4.8},
    "api_calls": {"allocated": 500, "used": 142, "remaining": 358}
  },
  "active_machines": 2,
  "total_checkpoints": 127,
  "created_at": "2024-08-24T10:00:00Z",
  "updated_at": "2024-08-24T11:30:00Z"
}
```

### Update Project
```http
PUT /projects/{project_uuid}
{
  "greediness": 0.5,
  "description": "Updated project scope"
}
```

### List Projects
```http
GET /projects?active=true&sort=greediness_desc

Response:
{
  "projects": [
    {
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "name": "autovibe-core", 
      "greediness": 0.42,
      "status": "active",
      "daily_resource_usage": {...}
    }
  ],
  "total_count": 8
}
```

## Checkpoint Management Endpoints

### List Project Checkpoints
```http
GET /projects/{project_uuid}/checkpoints?limit=50&type=manual

Response:
{
  "checkpoints": [
    {
      "uuid": "checkpoint-abc123-def456",
      "created_at": "2024-08-24T11:15:00Z",
      "created_by_machine": "machine-789xyz-012abc",
      "type": "automatic",
      "size_mb": 340,
      "description": "After implementing OAuth integration",
      "file_count": 127
    }
  ],
  "total_count": 89,
  "has_more": true
}
```

### Get Checkpoint Details  
```http
GET /checkpoints/{checkpoint_uuid}

Response:
{
  "uuid": "checkpoint-abc123-def456",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "parent_checkpoint": "checkpoint-parent-xyz789",
  "created_by_machine": "machine-789xyz-012abc",
  "proxmox_snapshot_id": "snap-vm001-abc123",
  "file_manifest": {
    "src/auth.py": {"size": 2048, "sha256": "abc123..."},
    "config/oauth.json": {"size": 512, "sha256": "def456..."}
  },
  "created_at": "2024-08-24T11:15:00Z",
  "size_bytes": 356515840,
  "metadata": {
    "trigger": "api_response",
    "machine_context": "implementing oauth flow"
  }
}
```

### Spawn from Checkpoint
```http
POST /checkpoints/{checkpoint_uuid}/spawn
{
  "machine_type": "claude_code",
  "prompt": "Continue authentication implementation with testing",
  "file_modifications": {}
}

Response:
{
  "machine_uuid": "machine-new-456def",
  "spawned_at": "2024-08-24T12:00:00Z",
  "parent_checkpoint": "checkpoint-abc123-def456"
}
```

### Compare Checkpoints
```http
GET /checkpoints/{checkpoint1_uuid}/diff/{checkpoint2_uuid}

Response:
{
  "files_added": ["tests/test_auth.py", "docs/auth.md"],
  "files_modified": ["src/auth.py", "config/settings.json"],
  "files_deleted": ["old_auth.py"],
  "total_changes": 147,
  "summary": "Added comprehensive testing and documentation for authentication system"
}
```

## Resource Management Endpoints

### Get Resource Overview
```http
GET /resources/overview

Response:
{
  "system_totals": {
    "daily_money_budget": 3.0,
    "daily_money_used": 1.85,
    "active_machines": 5,
    "total_projects": 8
  },
  "resource_utilization": {
    "money": 61.7,      // Percentage used
    "time": 45.2,
    "api_calls": 28.4,
    "compute": 67.3
  },
  "top_resource_consumers": [
    {
      "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
      "project_name": "autovibe-core",
      "daily_money_used": 0.95
    }
  ]
}
```

### Get Project Resource Usage
```http
GET /resources/projects/{project_uuid}?period=last_7_days

Response:
{
  "period": "last_7_days",
  "total_usage": {
    "money": 12.75,
    "time_hours": 25.3,
    "api_calls": 1247,
    "compute_hours": 18.7
  },
  "daily_breakdown": [
    {
      "date": "2024-08-24",
      "money": 1.85,
      "time_hours": 3.2,
      "api_calls": 142,
      "machines_active": 2
    }
  ],
  "efficiency_metrics": {
    "cost_per_checkpoint": 0.34,
    "api_calls_per_operation": 8.7,
    "average_machine_runtime": 45.2
  }
}
```

### Update Project Resource Budget
```http
PUT /resources/projects/{project_uuid}/budget
{
  "daily_money": 4.0,
  "daily_api_calls": 600,
  "max_concurrent_machines": 3
}
```

## Hub Management Endpoints

### Get Hub Status
```http
GET /hub/status

Response:
{
  "hub_uuid": "hub-primary-001",
  "status": "healthy",
  "uptime_hours": 72.3,
  "version": "1.0.0",
  "last_checkpoint": "hub-checkpoint-daily-20240824",
  "active_machines": 5,
  "active_projects": 8,
  "resource_utilization": {
    "cpu_percent": 45.2,
    "memory_percent": 67.8,
    "disk_percent": 23.1
  },
  "api_stats": {
    "requests_per_hour": 247,
    "average_response_time_ms": 85,
    "error_rate_percent": 0.3
  }
}
```

### Create Hub Checkpoint
```http
POST /hub/checkpoint
{
  "type": "manual",
  "reason": "before_experimental_feature",
  "description": "Safe checkpoint before enabling multi-hub evolution"
}

Response:
{
  "checkpoint_uuid": "hub-checkpoint-manual-abc123", 
  "created_at": "2024-08-24T12:00:00Z",
  "estimated_completion": "2024-08-24T12:05:00Z"
}
```

### Spawn Experimental Hub
```http
POST /hub/spawn-experimental
{
  "base_checkpoint": "hub-checkpoint-daily-20240810",
  "experiment_name": "optimized_machine_selection",
  "configuration_changes": {
    "machine_selection_algorithm": "ml_based",
    "checkpoint_frequency": "adaptive",
    "resource_optimization": "aggressive"
  },
  "experiment_duration_days": 14
}

Response:
{
  "experimental_hub_uuid": "hub-experiment-001",
  "experiment_id": "exp-machine-selection-001",
  "estimated_spawn_time": "2024-08-24T12:10:00Z"
}
```

## Monitoring and Analytics Endpoints

### System Health
```http
GET /health

Response:
{
  "status": "healthy",
  "services": {
    "hub": "healthy",
    "database": "healthy", 
    "proxmox": "healthy",
    "api": "healthy"
  },
  "active_machines": 5,
  "resource_utilization": 67.3,
  "last_checkpoint": "2024-08-24T00:00:00Z"
}
```

### Analytics and Metrics
```http
GET /analytics/efficiency?period=last_30_days

Response:
{
  "period": "last_30_days",
  "overall_efficiency": {
    "successful_operations": 245,
    "total_operations": 267,
    "success_rate": 91.8,
    "average_cost_per_success": 2.45
  },
  "machine_performance": {
    "claude_code": {
      "success_rate": 94.2,
      "average_cost": 18.75,
      "average_time_minutes": 42.3
    }
  },
  "resource_trends": [
    {
      "date": "2024-08-24",
      "efficiency_score": 87.3,
      "cost_per_operation": 2.34
    }
  ]
}
```

## Error Responses

### Standard Error Format
```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_LIMIT_EXCEEDED",
    "message": "Insufficient budget to spawn new machine",
    "details": {
      "required_money": 25.0,
      "available_money": 12.50,
      "project_uuid": "550e8400-e29b-41d4-a716-446655440000"
    },
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-08-24T11:30:00Z"
  }
}
```

### HTTP Status Codes
- **200**: Success
- **201**: Created (machines, projects, checkpoints)
- **400**: Bad Request (invalid parameters)
- **401**: Unauthorized (invalid API key)
- **403**: Forbidden (insufficient permissions)
- **404**: Not Found (resource doesn't exist)
- **409**: Conflict (resource limits exceeded)
- **429**: Too Many Requests (rate limited)
- **500**: Internal Server Error (system failure)

This comprehensive API enables full management of the autovibe system while maintaining security, resource controls, and detailed monitoring capabilities.