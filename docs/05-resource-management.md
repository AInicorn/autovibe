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

```python
def calculate_daily_resource_allocation(projects, total_resources):
    total_greediness = sum(p.greediness for p in projects)
    
    allocations = {}
    for project in projects:
        project_share = project.greediness / total_greediness
        
        allocations[project.uuid] = {
            'money': total_resources['daily_money_budget'] * project_share,
            'time_hours': total_resources['daily_compute_hours'] * project_share,
            'api_calls': total_resources['daily_api_quota'] * project_share,
            'storage_gb': total_resources['storage_quota'] * project_share
        }
    
    return allocations
```

### Example Resource Distribution
```yaml
# System-wide resource limits
total_biweekly_money_budget: 42.0  # CAD
total_daily_compute_hours: 24       # CPU hours across all VMs
total_daily_api_calls: 1000         # Combined API quota
total_storage_gb: 500               # Checkpoint and file storage

# Project allocations (greediness-based)
projects:
  autovibe-core:
    greediness: 0.42
    daily_money: 1.26    # 42/14 * 0.42 = $1.26/day
    daily_compute: 10.08  # 24 * 0.42 = 10 CPU hours/day
    daily_api_calls: 420  # 1000 * 0.42 = 420 calls/day
    
  crypto-profit:
    greediness: 0.1
    daily_money: 0.30    # $0.30/day
    daily_compute: 2.4    # 2.4 CPU hours/day
    daily_api_calls: 100  # 100 calls/day
```

## Real-Time Resource Tracking

### Resource Usage Monitoring
```python
class ResourceTracker:
    def __init__(self, machine_uuid):
        self.machine_uuid = machine_uuid
        self.usage_log = []
        self.current_usage = {
            'money': 0.0,
            'time_minutes': 0,
            'api_calls': 0,
            'cpu_hours': 0.0,
            'memory_gb_hours': 0.0,
            'storage_gb': 0.0
        }
    
    def record_api_call(self, service, cost, tokens=None):
        self.current_usage['money'] += cost
        self.current_usage['api_calls'] += 1
        
        self.usage_log.append({
            'timestamp': datetime.utcnow(),
            'type': 'api_call',
            'service': service,
            'cost': cost,
            'tokens': tokens
        })
        
        # Check budget limits after each operation
        self.enforce_budget_limits()
    
    def record_compute_usage(self, cpu_percent, memory_gb):
        interval_hours = 1/60  # 1 minute intervals
        
        self.current_usage['cpu_hours'] += (cpu_percent / 100) * interval_hours
        self.current_usage['memory_gb_hours'] += memory_gb * interval_hours
        
        # Update VM runtime costs
        vm_cost_per_hour = calculate_vm_hourly_cost(cpu_percent, memory_gb)
        self.current_usage['money'] += vm_cost_per_hour * interval_hours
    
    def enforce_budget_limits(self):
        project_budget = get_current_project_budget(self.machine_uuid)
        
        for resource_type, used in self.current_usage.items():
            allocated = project_budget[resource_type]
            
            if used > allocated:
                raise ResourceExceededException(
                    resource=resource_type,
                    used=used,
                    allocated=allocated,
                    machine_uuid=self.machine_uuid
                )
```

### Budget Enforcement Strategies

#### Soft Limits (Warnings)
```python
def check_soft_limits(machine_uuid, threshold=0.8):
    """Alert when approaching resource limits"""
    usage = get_current_resource_usage(machine_uuid)
    budget = get_project_resource_budget(machine_uuid)
    
    warnings = []
    for resource_type, used in usage.items():
        allocated = budget[resource_type]
        utilization = used / allocated if allocated > 0 else 0
        
        if utilization > threshold:
            warnings.append({
                'resource': resource_type,
                'utilization': utilization,
                'remaining': allocated - used
            })
    
    return warnings
```

#### Hard Limits (Operation Blocking)
```python
def enforce_hard_limits(machine_uuid, operation_type):
    """Block operations that would exceed limits"""
    estimated_cost = estimate_operation_resource_cost(operation_type)
    current_usage = get_current_resource_usage(machine_uuid)
    budget = get_project_resource_budget(machine_uuid)
    
    for resource_type, estimated in estimated_cost.items():
        used = current_usage[resource_type]
        allocated = budget[resource_type]
        
        if used + estimated > allocated:
            # Block operation and suggest alternatives
            alternatives = suggest_resource_optimizations(machine_uuid, operation_type)
            
            raise ResourceLimitBlockedException(
                resource=resource_type,
                required=estimated,
                available=allocated - used,
                alternatives=alternatives
            )
```

## Dynamic Resource Reallocation

### Performance-Based Adjustments
```python
def optimize_resource_allocation():
    """Periodically adjust allocations based on performance"""
    projects = db.get_all_projects()
    
    # Calculate efficiency metrics
    efficiency_scores = {}
    for project in projects:
        recent_machines = get_recent_machines(project.uuid, days=7)
        
        efficiency_scores[project.uuid] = calculate_efficiency_score(
            successful_operations=count_successful_operations(recent_machines),
            total_resource_cost=sum_resource_usage(recent_machines),
            value_delivered=assess_project_value(project.uuid)
        )
    
    # Adjust greediness based on efficiency
    total_efficiency = sum(efficiency_scores.values())
    
    for project in projects:
        current_greediness = project.greediness
        efficiency_ratio = efficiency_scores[project.uuid] / total_efficiency
        
        # Gradual adjustment toward efficiency-based allocation
        optimal_greediness = efficiency_ratio * 0.8  # 80% efficiency-based, 20% current
        new_greediness = current_greediness * 0.8 + optimal_greediness * 0.2
        
        # Apply bounds and update
        project.greediness = max(0.01, min(1.0, new_greediness))
        db.update_project_greediness(project.uuid, project.greediness)
```

## Resource Cost Calculations

### API Cost Tracking
```python
API_COSTS = {
    'claude-3-5-sonnet': {
        'input_per_token': 0.000003,   # $0.003 per 1K tokens
        'output_per_token': 0.000015,  # $0.015 per 1K tokens
    },
    'claude-3-haiku': {
        'input_per_token': 0.00000025, # $0.00025 per 1K tokens
        'output_per_token': 0.00000125, # $0.00125 per 1K tokens
    }
}

def calculate_api_cost(service, input_tokens, output_tokens):
    costs = API_COSTS.get(service, {})
    
    input_cost = input_tokens * costs.get('input_per_token', 0)
    output_cost = output_tokens * costs.get('output_per_token', 0)
    
    return input_cost + output_cost
```

### VM Runtime Cost Calculation
```python
def calculate_vm_hourly_cost(cpu_cores, memory_gb, storage_gb, vm_type='standard'):
    """Calculate VM costs based on resource usage"""
    
    # Base costs per hour (example pricing)
    RESOURCE_COSTS = {
        'cpu_core_hour': 0.05,     # $0.05 per core hour
        'memory_gb_hour': 0.01,    # $0.01 per GB hour
        'storage_gb_hour': 0.001,  # $0.001 per GB hour
    }
    
    hourly_cost = (
        cpu_cores * RESOURCE_COSTS['cpu_core_hour'] +
        memory_gb * RESOURCE_COSTS['memory_gb_hour'] +
        storage_gb * RESOURCE_COSTS['storage_gb_hour']
    )
    
    # Apply VM type multiplier
    VM_TYPE_MULTIPLIERS = {
        'standard': 1.0,
        'hub': 1.2,      # Hub VMs have additional overhead
        'high_memory': 1.5,
        'compute_optimized': 1.3
    }
    
    return hourly_cost * VM_TYPE_MULTIPLIERS.get(vm_type, 1.0)
```

## Resource Optimization Strategies

### Machine Selection by Resource Efficiency
```python
def select_resource_efficient_machine(prompt, project_uuid, resource_constraints):
    """Choose machine type that maximizes output per resource unit"""
    
    available_machines = get_machine_types_within_constraints(resource_constraints)
    
    scores = []
    for machine_type in available_machines:
        # Historical performance data
        avg_success_rate = get_average_success_rate(machine_type)
        avg_resource_usage = get_average_resource_usage(machine_type)
        
        # Calculate efficiency score
        efficiency = avg_success_rate / sum(avg_resource_usage.values())
        
        # Factor in prompt compatibility
        prompt_match = calculate_prompt_compatibility(prompt, machine_type)
        
        total_score = efficiency * prompt_match
        scores.append((machine_type, total_score))
    
    return max(scores, key=lambda x: x[1])[0]
```

### Resource Usage Prediction
```python
def predict_operation_resource_usage(operation_type, prompt, machine_type):
    """Predict resource requirements based on historical data"""
    
    # Analyze prompt complexity
    prompt_features = extract_prompt_features(prompt)
    complexity_score = assess_prompt_complexity(prompt_features)
    
    # Get baseline usage for machine type
    baseline_usage = get_baseline_resource_usage(machine_type)
    
    # Apply complexity multipliers
    COMPLEXITY_MULTIPLIERS = {
        'money': lambda c: 0.5 + c * 0.5,      # 0.5x to 1.5x based on complexity
        'time_minutes': lambda c: 5 + c * 30,   # 5 to 35 minutes
        'api_calls': lambda c: 1 + c * 10,      # 1 to 11 API calls
        'cpu_hours': lambda c: 0.1 + c * 0.5    # 0.1 to 0.6 CPU hours
    }
    
    predicted_usage = {}
    for resource_type, multiplier_func in COMPLEXITY_MULTIPLIERS.items():
        base_usage = baseline_usage[resource_type]
        predicted_usage[resource_type] = base_usage * multiplier_func(complexity_score)
    
    return predicted_usage
```

## Emergency Resource Management

### Circuit Breaker Pattern
```python
class ResourceCircuitBreaker:
    def __init__(self, resource_type, threshold_multiplier=1.5):
        self.resource_type = resource_type
        self.threshold_multiplier = threshold_multiplier
        self.failure_count = 0
        self.last_failure_time = None
        self.state = 'closed'  # closed, open, half_open
    
    def call(self, operation):
        if self.state == 'open':
            if self._should_attempt_reset():
                self.state = 'half_open'
            else:
                raise CircuitBreakerOpenError(f"{self.resource_type} circuit breaker is open")
        
        try:
            result = operation()
            self._on_success()
            return result
        except ResourceExceededException as e:
            self._on_failure()
            raise
    
    def _on_failure(self):
        self.failure_count += 1
        self.last_failure_time = datetime.utcnow()
        
        if self.failure_count >= 3:  # Trip after 3 failures
            self.state = 'open'
    
    def _should_attempt_reset(self):
        return (datetime.utcnow() - self.last_failure_time).seconds > 300  # 5 minute timeout
```

This comprehensive resource management system ensures efficient utilization while preventing resource exhaustion and maintaining system stability.