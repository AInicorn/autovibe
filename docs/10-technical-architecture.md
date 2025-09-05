# Technical Architecture Overview

## System Architecture

autovibe implements a hub-and-spoke architecture where the Hub (itself an Intelligent Machine) orchestrates other Intelligent Machines in isolated VM environments.

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Hub VM        │    │  Task Machine   │    │  Task Machine   │
│                 │    │     VM #1       │    │     VM #2       │
│ ┌─────────────┐ │    │                 │    │                 │
│ │ Hub Service │ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │             │◄├────┤ │Claude Agent │ │    │ │ Aider Agent │ │
│ │ - Orchestr. │ │    │ │             │ │    │ │             │ │
│ │ - Budget    │ │    │ └─────────────┘ │    │ └─────────────┘ │
│ │ - API       │ │    │                 │    │                 │
│ └─────────────┘ │    │ Project Files   │    │ Project Files   │
│                 │    │ + Checkpoints   │    │ + Checkpoints   │
│ ┌─────────────┐ │    └─────────────────┘    └─────────────────┘
│ │ PostgreSQL  │ │
│ │  Database   │ │           ┌─────────────────┐
│ └─────────────┘ │           │   Proxmox       │
│                 │           │  Infrastructure │
│ Docker Compose  │           │                 │
└─────────────────┘           │ - VM Management │
                              │ - Snapshots     │
                              │ - Storage       │
                              │ - Networking    │
                              └─────────────────┘
```

## Technology Stack

### Infrastructure Layer
- **Proxmox VE**: VM orchestration, snapshot management, resource control
- **ZFS/PBS**: Storage backend for snapshots and data deduplication
- **Linux (Ubuntu)**: Base operating system for all VMs
- **Docker/Docker Compose**: Container orchestration within Hub VM

### Application Layer
- **Python 3.11+**: Primary programming language
- **FastAPI**: REST API framework for Hub services
- **PostgreSQL 15**: Primary database for metadata and state
- **Redis**: Caching, session management, and job queues
- **nginx**: Reverse proxy, SSL termination, rate limiting

### Machine Agent Layer
- **Claude API**: Primary AI service integration
- **anthropic-sdk**: Official Anthropic Python SDK
- **autovibe-agent**: Custom agent framework for machine lifecycle
- **Git**: Version control integration
- **Docker**: Containerized tooling within VMs

## Hub Internal Architecture

### Hub VM Components
```yaml
# docker-compose.yml structure
services:
  hub-api:
    # FastAPI application
    # Handles REST API requests
    # Manages machine lifecycle
    # Enforces resource budgets
    
  hub-scheduler:
    # Background job processor  
    # Handles checkpointing
    # Manages Hub evolution
    # Resource monitoring
    
  database:
    # PostgreSQL primary database
    # Metadata and state storage
    # Resource usage tracking
    
  redis:
    # Session and cache storage
    # Job queue management
    # Real-time data
    
  nginx:
    # Reverse proxy
    # SSL termination
    # Rate limiting
    # Static file serving

  prometheus:
    # Metrics collection
    # Resource monitoring
    # Performance tracking
    
  grafana:
    # Monitoring dashboard
    # Alerting
    # Visualization
```

### Hub Service Responsibilities

#### API Service (`hub-api`)
- **Machine Management**: Spawn, monitor, terminate Intelligent Machines
- **Project Management**: Create, update, manage project configurations
- **Checkpoint Coordination**: Trigger and coordinate checkpoint creation
- **Resource Enforcement**: Enforce budgets and resource limits
- **Authentication**: Handle API authentication and authorization

#### Scheduler Service (`hub-scheduler`)
- **Background Jobs**: Process long-running tasks asynchronously
- **Automated Checkpointing**: Create scheduled Hub checkpoints
- **Resource Monitoring**: Track and aggregate resource usage
- **Hub Evolution**: Manage experimental Hub spawning and selection
- **Cleanup Tasks**: Remove expired checkpoints and clean up resources

## Machine Agent Architecture

### Agent Components
Each Intelligent Machine VM contains:

```python
# Machine Agent Structure
class IntelligentMachineAgent:
    def __init__(self):
        self.machine_uuid = get_machine_uuid()
        self.hub_client = HubAPIClient()
        self.checkpoint_manager = CheckpointManager()
        self.resource_tracker = ResourceTracker()
        self.claude_client = AnthropicClient()
        
    async def execute_prompt(self, prompt: str):
        # Main execution loop with checkpointing
        while not self.is_complete():
            # Track resources before operation
            self.resource_tracker.start_operation()
            
            # Execute AI operation
            response = await self.claude_client.complete(prompt)
            
            # Track resources after operation
            usage = self.resource_tracker.end_operation()
            self.hub_client.report_resource_usage(usage)
            
            # Create checkpoint after API call
            if self.should_checkpoint():
                checkpoint = await self.checkpoint_manager.create()
                self.hub_client.report_checkpoint(checkpoint)
            
            # Process response and update files
            self.process_response(response)
```

### Resource Tracking Integration
```python
class ResourceTracker:
    def track_api_call(self, service: str, tokens_in: int, tokens_out: int):
        cost = self.calculate_api_cost(service, tokens_in, tokens_out)
        
        self.current_usage['money'] += cost
        self.current_usage['api_calls'] += 1
        
        # Real-time budget check
        if self.check_budget_limits():
            raise ResourceLimitExceededError()
        
        # Report to Hub
        self.hub_client.report_resource_usage({
            'resource_type': 'api_call',
            'service': service,
            'cost': cost,
            'tokens_in': tokens_in,
            'tokens_out': tokens_out,
            'timestamp': datetime.utcnow()
        })
```

## Data Flow Architecture

### Machine Spawning Flow
```
1. API Request → Hub API Service
2. Hub API → Resource Budget Check
3. Hub API → Proxmox VM Creation
4. Proxmox → VM Started with Checkpoint
5. VM Agent → Registration with Hub
6. Hub → Machine Status Update
7. VM Agent → Begin Prompt Execution
```

### Checkpointing Flow
```
1. Machine Agent → Checkpoint Trigger
2. Machine Agent → File State Capture
3. Machine Agent → Proxmox Snapshot Request
4. Proxmox → VM Snapshot Creation
5. Machine Agent → Metadata to Database
6. Hub → Checkpoint Verification
7. Hub → Storage Optimization
```

### Resource Tracking Flow
```
1. Machine Agent → Resource Usage Event
2. Machine Agent → Hub API Resource Report
3. Hub API → Database Resource Log
4. Hub Scheduler → Budget Enforcement Check
5. Hub Scheduler → Aggregation and Analytics
6. Hub API → Budget Limit Response
```

## Security Architecture

### Network Segmentation
```
┌─────────────────┐
│  Management     │  Hub VM
│   Network       │  Proxmox Management
│  (vmbr0)        │  
└─────────────────┘
         │
┌─────────────────┐
│   Machine       │  Task Machine VMs
│   Network       │  Isolated API Access
│   (vmbr1)       │  No Internet Access
└─────────────────┘
```

### Permission Model
- **Hub VM**: Full Proxmox API access, database access, external APIs
- **Task VMs**: Limited Hub API access, specific external APIs only
- **User API**: Authenticated access to Hub API endpoints
- **Admin API**: Full system management access

### Data Encryption
- **At Rest**: Database encryption, snapshot encryption
- **In Transit**: TLS 1.3 for all API communications
- **Secrets**: Encrypted configuration management
- **Audit**: Comprehensive audit logging

## Scalability Considerations

### Horizontal Scaling
- **Multiple Proxmox Nodes**: Distribute VMs across hardware
- **Database Sharding**: Partition data by project or time
- **API Load Balancing**: Multiple Hub API instances
- **Storage Distribution**: Distributed checkpoint storage

### Performance Optimization
- **Database Indexing**: Optimized queries for frequent operations
- **Connection Pooling**: Efficient database connection management
- **Caching**: Redis caching for frequently accessed data
- **Async Processing**: Non-blocking I/O for all external communications

### Resource Management
- **Dynamic Allocation**: Adjust resources based on demand
- **Predictive Scaling**: Scale resources based on usage patterns
- **Cost Optimization**: Right-size VMs based on workload requirements
- **Storage Optimization**: Compression and deduplication

## Monitoring and Observability

### Metrics Collection
```python
# Key metrics tracked
METRICS = {
    'hub_metrics': [
        'active_machines_count',
        'api_requests_per_second',
        'checkpoint_creation_time',
        'resource_budget_utilization'
    ],
    'machine_metrics': [
        'machine_spawn_time',
        'checkpoint_frequency',
        'api_response_time',
        'resource_efficiency'
    ],
    'infrastructure_metrics': [
        'proxmox_api_response_time',
        'vm_creation_time',
        'snapshot_creation_time',
        'storage_usage'
    ]
}
```

### Health Checks
- **Hub Health**: API responsiveness, database connectivity, Proxmox connectivity
- **Machine Health**: Agent responsiveness, resource usage, checkpoint success
- **Infrastructure Health**: VM status, storage availability, network connectivity

### Alerting Rules
- **Resource Alerts**: Budget threshold violations, resource exhaustion
- **Performance Alerts**: API response time degradation, checkpoint failures
- **System Alerts**: VM failures, database connectivity issues, storage problems

This architecture provides a robust, scalable foundation for the autovibe system while maintaining security, observability, and operational simplicity.