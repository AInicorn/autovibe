# AutoVibe API Endpoints Design

## REST API Architecture

AutoVibe exposes a RESTful API for Intelligent Machine management, budget control, and project operations.

### Base URL
```
http://localhost:8000/api/v1
```

### Core Endpoints

#### Intelligent Machine Management
```http
# Create new Intelligent Machine from checkpoint
POST /machines
{
  "type": "CLAUDE_CODE",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "checkpoint_uuid": "checkpoint-550e8400-e29b-41d4",
  "prompt": "Add user authentication system",
  "file_modifications": {}  // Optional file replacements before spawn
}

# Get Intelligent Machine status  
GET /machines/{machine_id}
Response:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "CLAUDE_CODE",
  "status": "WORKING",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "vm_id": "vm-001",
  "prompt": "Add user authentication system",
  "estimated_cost": 25.0,
  "actual_cost": 18.75,
  "progress": "Analyzing existing auth patterns...",
  "spawned_at": "2024-08-23T10:30:00Z"
}

# Terminate Intelligent Machine
POST /machines/{machine_id}/terminate
{
  "reason": "user_requested",
  "save_state": true
}

# List Intelligent Machines
GET /machines?status=WORKING&project_id=1
```

#### Budget Management
```http
# Get budget overview
GET /budget/overview
Response:
{
  "total_biweekly_money_budget": 42.0,
  "total_spent_this_week": 28.50,
  "remaining_budget": 13.50,
  "utilization_percentage": 67.9,
  "projects": [
    {
      "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
      "daily_budget": 2.52,
      "spent_today": 1.85,
      "remaining_today": 0.67
    }
  ]
}

# Get project budget allocation  
GET /budget/projects/{project_uuid}
Response:
{
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "greediness": 0.42,
  "daily_allocation": 2.52,
  "spent_today": 1.85,
  "remaining_today": 0.67,
  "weekly_total": 17.64,
  "spent_this_week": 12.95
}

# Update project greediness
PUT /budget/projects/{project_uuid}
{
  "greediness": 0.5
}

# [POST-MVP] Get resource alerts - removed for MVP
```

#### Project Operations
```http
# Create project
POST /projects
{
  "name": "crypto-trading-bot",
  "greediness": 0.3,
  "description": "Automated cryptocurrency trading system"
}

# Get project details with evolution history
GET /projects/{project_uuid}
Response:
{
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "crypto-trading-bot",
  "greediness": 0.3,
  "current_checkpoint": "checkpoint-abc123",
  "recent_evolutions": [
    {
      "checkpoint_uuid": "checkpoint-xyz789",
      "prompt": "Add user authentication",
      "machine_id": "550e8400-e29b-41d4-a716-446655440000",
      "resource_usage": {"money": 12.50, "time_minutes": 45, "api_calls": 85},
      "timestamp": "2024-08-23T10:30:00Z",
      "status": "COMPLETED"
    }
  ],
  "total_cost": 45.75,
  "created_at": "2024-08-23T09:00:00Z",
  "updated_at": "2024-08-23T12:15:00Z"
}

# Update project
PUT /projects/{project_uuid}
{
  "greediness": 0.4,
  "description": "Updated project description"
}

# List projects with filtering
GET /projects?min_greediness=0.1&status=active
```

#### VM and Snapshot Management
```http
# [RESTRICTED] Get VM status - only accessible by system/orchestrator
GET /vms/{vm_id}
Response:
{
  "vm_id": "vm-001",
  "status": "RUNNING",
  "project_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "machine_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "resource_metrics": {
    "cpu_percent": 45.2,
    "memory_gb": 2.7,
    "disk_gb": 4.6,
    "network_mbps": 5.4
  },
  "uptime_hours": 2.3
}

# Create checkpoint from VM state
POST /vms/{vm_id}/checkpoint
{
  "description": "After implementing authentication",
  "save_point": true
}

# Spawn new machine from checkpoint
POST /projects/{project_uuid}/spawn-from-checkpoint
{
  "checkpoint_uuid": "checkpoint-xyz789",
  "prompt": "Continue from this checkpoint with alternative approach"
}

# List project checkpoints
GET /projects/{project_uuid}/checkpoints

# Get full evolution history (paginated)
GET /projects/{project_uuid}/evolution-history?limit=50&offset=0
Response:
{
  "evolutions": [...],  // Limited list of evolution entries
  "total_count": 245,
  "has_more": true,
  "next_offset": 50
}
```

#### System Status and Analytics
```http
# Get system health
GET /health
Response:
{
  "status": "healthy",
  "active_machines": 3,
  "active_vms": 3,
  "total_daily_spend": 18.75,
  "budget_utilization": 78.5,
  "system_load": 0.42
}

# Get detailed system statistics
GET /stats
Response:
{
  "machines": {
    "total_spawned_today": 12,
    "currently_active": 3,
    "average_runtime_minutes": 45.2,
    "success_rate": 94.5
  },
  "budget": {
    "total_biweekly_money_budget": 42.0,
    "spent_this_week": 28.50,
    "projected_weekly_spend": 35.75,
    "most_expensive_project": "crypto-trading-bot",
    "most_efficient_project": "documentation-updates"
  },
  "performance": {
    "average_cost_per_operation": 8.25,
    "cost_savings_vs_manual": 67.3,
    "project_completion_rate": 89.2
  }
}

# Get cost analytics
GET /analytics/resource-usage?period=week&project_uuid=550e8400-e29b-41d4
```

### Response Format

All endpoints return JSON with consistent structure:
```json
{
  "success": true,
  "data": {...},
  "message": "Operation completed successfully",
  "timestamp": "2024-08-23T10:30:00Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Error Handling
- **400 Bad Request**: Invalid input parameters or malformed JSON
- **401 Unauthorized**: Invalid API key or insufficient permissions
- **404 Not Found**: Resource not found (project, machine, snapshot, etc.)
- **409 Conflict**: Resource conflict (machine already running, insufficient budget)
- **429 Too Many Requests**: Budget exceeded or rate limited
- **500 Internal Server Error**: System error or VM management failure

Error response format:
```json
{
  "success": false,
  "error": {
    "code": "BUDGET_EXCEEDED",
    "message": "Insufficient budget to spawn new Intelligent Machine",
    "details": {
      "required_budget": 25.0,
      "available_budget": 12.50,
      "project_uuid": "550e8400-e29b-41d4-a716-446655440000"
    }
  },
  "timestamp": "2024-08-23T10:30:00Z",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Authentication & Security
- **API Keys**: Required for all endpoints
- **Rate Limiting**: 1000 requests per hour per API key
- **HTTPS Only**: All API communication must be encrypted
- **CORS**: Configurable cross-origin resource sharing

This comprehensive API design provides full control over AutoVibe's Intelligent Machine orchestration capabilities.