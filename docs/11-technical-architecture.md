# Technical Architecture Overview

## System Architecture

autovibe implements a hub-and-spoke architecture where the Hub (itself an Intelligent Machine) orchestrates other Intelligent Machines in isolated VM environments.

The architecture consists of a central Hub VM containing the Hub Service with orchestration, budget management, and API capabilities, plus a PostgreSQL database for state management, all running under Docker Compose. Task Machine VMs connect to the Hub, each containing specialized agents (like Claude Agent or Aider Agent) along with project files and checkpoints. The underlying Proxmox Infrastructure provides VM management, snapshots, storage, and networking capabilities for the entire system.

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
The Hub VM runs multiple services orchestrated through Docker Compose. The hub-api service runs a FastAPI application handling REST API requests, machine lifecycle management, and resource budget enforcement. The hub-scheduler provides background job processing for checkpointing, Hub evolution management, and resource monitoring. The database service runs PostgreSQL for metadata and state storage plus resource usage tracking. Redis handles session and cache storage, job queue management, and real-time data. Nginx provides reverse proxy functionality, SSL termination, rate limiting, and static file serving. Prometheus collects metrics for resource monitoring and performance tracking. Grafana provides monitoring dashboards, alerting capabilities, and data visualization.

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
Each Intelligent Machine VM contains an agent with several key components. The IntelligentMachineAgent initializes with a unique machine UUID, Hub API client connection, checkpoint manager, resource tracker, and AI service client (such as Anthropic's Claude). The main execution loop processes prompts by tracking resources before and after each operation, executing AI operations through the service client, reporting resource usage to the Hub, creating checkpoints when appropriate, and processing responses to update project files. This architecture ensures comprehensive monitoring and state management throughout the agent's lifecycle.

### Resource Tracking Integration
The ResourceTracker component monitors and reports all resource consumption. For API calls, it calculates costs based on service type and token counts (input and output), updates current usage totals for money and API calls, performs real-time budget limit checks that raise exceptions when exceeded, and reports detailed usage data to the Hub including resource type, service name, costs, token counts, and timestamps. This comprehensive tracking ensures accurate billing and budget enforcement across all machine operations.

## Data Flow Architecture

### Machine Spawning Flow
The machine spawning process follows a seven-step sequence: An API request initiates the process through the Hub API Service, which performs a resource budget check to ensure sufficient allocation. Upon validation, the Hub API requests VM creation through Proxmox, which starts the VM using the specified checkpoint. The VM Agent registers with the Hub upon initialization, triggering a machine status update in the Hub's database. Finally, the VM Agent begins prompt execution to complete the spawning process.

### Checkpointing Flow
The checkpointing process involves seven coordinated steps: The Machine Agent triggers checkpoint creation based on configured criteria, then captures the current file state for manifest generation. The Agent requests VM snapshot creation through Proxmox, which performs the actual VM snapshot operation. The Agent records checkpoint metadata in the database, after which the Hub performs checkpoint verification to ensure integrity. Finally, the Hub applies storage optimization techniques to manage disk space efficiently.

### Resource Tracking Flow
Resource tracking follows a six-step process: A Machine Agent generates resource usage events during operations, then reports this usage data to the Hub API. The Hub API logs the resource information in the database, triggering the Hub Scheduler to perform budget enforcement checks. The Scheduler also handles data aggregation and analytics processing. Finally, the Hub API responds with current budget limit status to inform ongoing operations.

## Security Architecture

### Network Segmentation
The network architecture uses two-tier segmentation for security isolation. The Management Network (vmbr0) serves the Hub VM and Proxmox Management interfaces with full access capabilities. The Machine Network (vmbr1) provides isolated connectivity for Task Machine VMs with restricted API-only access and no direct Internet connectivity, ensuring secure operation while maintaining necessary functionality.

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
The system tracks comprehensive metrics across three categories. Hub metrics include active machines count, API requests per second, checkpoint creation time, and resource budget utilization. Machine metrics monitor machine spawn time, checkpoint frequency, API response time, and resource efficiency. Infrastructure metrics track Proxmox API response time, VM creation time, snapshot creation time, and storage usage. This multi-layered monitoring approach provides visibility into system performance at all levels.

### Health Checks
- **Hub Health**: API responsiveness, database connectivity, Proxmox connectivity
- **Machine Health**: Agent responsiveness, resource usage, checkpoint success
- **Infrastructure Health**: VM status, storage availability, network connectivity

### Alerting Rules
- **Resource Alerts**: Budget threshold violations, resource exhaustion
- **Performance Alerts**: API response time degradation, checkpoint failures
- **System Alerts**: VM failures, database connectivity issues, storage problems

This architecture provides a robust, scalable foundation for the autovibe system while maintaining security, observability, and operational simplicity.