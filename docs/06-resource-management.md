# Resource Management and Economic System

## Multi-Resource Budget Framework

autovibe manages multiple resource types simultaneously, not just monetary costs:

- **Money**: API costs, VM runtime, storage, network transfer
- **Time**: Maximum execution duration, queue time, idle time
- **API Calls**: Service-specific quotas and rate limits  
- **Compute**: CPU hours, memory GB-hours, disk operations
- **Storage**: Checkpoint storage, file storage, database storage

## Project Resource Allocation

### Greediness-Based Distribution
Projects receive resources based on their **greediness factor** (0.0-1.0):

Daily resource allocation calculates each project's share based on their greediness factor relative to the total greediness across all projects. Each project receives a proportional allocation of the daily money budget, compute hours, API call quota, and storage quota based on their greediness ratio.

### Example Resource Distribution
Resource distribution example shows system-wide limits: $42 CAD biweekly budget, 24 CPU hours daily across all VMs, 1000 combined daily API calls, and 500GB total storage. Projects receive allocations proportional to their greediness factors - for example, autovibe-core with 0.42 greediness gets $1.26/day, 10 CPU hours/day, and 420 API calls/day, while crypto-profit with 0.1 greediness gets $0.30/day, 2.4 CPU hours/day, and 100 API calls/day.

## Real-Time Resource Tracking

### Resource Usage Monitoring
The ResourceTracker class monitors real-time resource usage for each machine, tracking money spent, execution time, API calls, CPU hours, memory usage, and storage consumption. It records API calls with service details, costs, and token counts, then enforces budget limits after each operation. Compute usage is recorded at 1-minute intervals, calculating both resource consumption and associated VM runtime costs. Budget enforcement checks current usage against allocated limits and raises ResourceExceededException when limits are exceeded.

### Budget Enforcement Strategies

#### Soft Limits (Warnings)
Soft limit checking alerts when resource utilization exceeds a configurable threshold (default 80%). The system compares current usage against allocated budget for each resource type, calculates utilization percentages, and generates warnings containing resource type, utilization level, and remaining allocation for resources approaching limits.

#### Hard Limits (Operation Blocking)
Hard limit enforcement blocks operations that would exceed resource allocations. The system estimates the resource cost of the requested operation, compares it against current usage and available budget, and raises ResourceLimitBlockedException with details about the resource constraint, required amount, available resources, and suggested optimization alternatives.

## Dynamic Resource Reallocation

### Performance-Based Adjustments
Dynamic resource reallocation periodically adjusts project allocations based on performance metrics. The system calculates efficiency scores for each project based on successful operations, total resource costs, and delivered value over the past 7 days. Project greediness factors are gradually adjusted using a weighted approach (80% efficiency-based, 20% current allocation) with bounds checking to ensure values remain between 0.01 and 1.0.

## Resource Cost Calculations

### API Cost Tracking
API cost tracking maintains pricing data for different services: Claude-3.5-Sonnet charges $0.003 per 1K input tokens and $0.015 per 1K output tokens, while Claude-3-Haiku charges $0.00025 per 1K input tokens and $0.00125 per 1K output tokens. The cost calculation function multiplies token counts by the appropriate per-token rates and returns the total cost.

### VM Runtime Cost Calculation
VM runtime cost calculation uses base rates of $0.05 per CPU core hour, $0.01 per GB memory hour, and $0.001 per GB storage hour. The calculation multiplies resource quantities by their respective rates and applies VM type multipliers: standard VMs (1.0x), Hub VMs (1.2x for additional overhead), high memory VMs (1.5x), and compute optimized VMs (1.3x).

## Resource Optimization Strategies

### Machine Selection by Resource Efficiency
Resource-efficient machine selection evaluates available machine types within resource constraints, calculating efficiency scores based on historical success rates divided by total resource usage. The algorithm factors in prompt compatibility and selects the machine type with the highest combined efficiency and compatibility score.

### Resource Usage Prediction
Resource usage prediction analyzes prompt complexity and applies multipliers to baseline usage for the machine type. Complexity-based multipliers range from 0.5x to 1.5x for money costs, 5 to 35 minutes for time duration, 1 to 11 API calls, and 0.1 to 0.6 CPU hours. The system extracts prompt features, calculates complexity scores, and generates predicted resource requirements.

## Emergency Resource Management

### Circuit Breaker Pattern
The ResourceCircuitBreaker implements emergency resource management using the circuit breaker pattern with three states: closed (normal operation), open (blocking operations), and half-open (testing recovery). The breaker trips to open state after 3 consecutive resource failures and attempts reset after a 5-minute timeout. When open, it raises CircuitBreakerOpenError to prevent further resource exhaustion, allowing the system to recover from resource constraints.

This comprehensive resource management system ensures efficient utilization while preventing resource exhaustion and maintaining system stability.