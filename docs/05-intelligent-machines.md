# Intelligent Machines - Unified Architecture

## Core Philosophy

Every component in autovibe is an Intelligent Machine following the same architectural principles:
- **Spawn** from checkpoints in isolated VM environments
- **Execute** specific goals while creating regular checkpoints  
- **Communicate** via defined interfaces and permissions
- **Terminate** cleanly, creating final checkpoints for future evolution

## Machine Categories

### Task Machines
Perform specific development, analysis, or operational tasks:

#### Claude Code Machine
- **Purpose**: Primary machine for complex reasoning, code generation, architecture decisions
- **Capabilities**: Multi-file editing, system design, debugging, documentation
- **Resource Profile**: High API usage, moderate compute, variable runtime
- **Checkpointing**: After each API response, before major operations

#### Aider Machine (Post-MVP)
- **Purpose**: Specialized for focused code editing and refactoring
- **Capabilities**: Single-file changes, test generation, code review
- **Resource Profile**: Moderate API usage, low compute, fast runtime
- **Checkpointing**: After file modifications, between logical steps

#### Qwen Coder Machine (Post-MVP)  
- **Purpose**: Local inference for simple operations, bulk changes
- **Capabilities**: Template generation, documentation, repetitive tasks
- **Resource Profile**: Local compute intensive, no API costs
- **Checkpointing**: After batch operations, on completion

### Hub Machines
Orchestrate and manage other machines:

#### Primary Hub
- **Purpose**: Main orchestrator, API server, resource manager
- **Capabilities**: Machine spawning, budget allocation, checkpoint coordination
- **Resource Profile**: Continuous operation, moderate compute, database intensive
- **Checkpointing**: Daily production snapshots, before major changes

#### Specialized Hubs (Post-MVP)
- **Development Hub**: Experimental features, algorithm testing
- **Domain Hub**: Client-specific or project-specific configurations
- **Geographic Hub**: Regional deployment for latency optimization

## Machine Lifecycle

### Unified Spawn Process
```python
def spawn_intelligent_machine(machine_type, checkpoint_uuid, prompt, options={}):
    # 1. Resource validation
    if not verify_resource_availability(machine_type, options):
        raise InsufficientResourcesError()
    
    # 2. VM creation from checkpoint
    vm_id = proxmox.restore_vm_from_checkpoint(
        checkpoint_uuid=checkpoint_uuid,
        machine_type=machine_type,
        resource_limits=calculate_resource_limits(options)
    )
    
    # 3. Apply file modifications (if specified)
    if options.get('file_modifications'):
        apply_file_changes(vm_id, options['file_modifications'])
    
    # 4. Register machine in system
    machine = register_machine(
        vm_id=vm_id,
        type=machine_type,
        parent_checkpoint=checkpoint_uuid,
        prompt=prompt,
        spawned_by=get_current_hub_id()
    )
    
    # 5. Start machine execution
    start_machine_execution(machine.uuid, prompt)
    
    return machine.uuid
```

### Regular Checkpointing During Operation
All machines create checkpoints at logical intervals:

```python
class IntelligentMachine:
    def __init__(self, machine_uuid):
        self.uuid = machine_uuid
        self.checkpoint_triggers = self._setup_checkpoint_triggers()
    
    def _setup_checkpoint_triggers(self):
        return {
            'api_response': True,      # After each API call
            'file_operation': True,    # After file writes
            'time_interval': 300,      # Every 5 minutes max
            'resource_threshold': 0.1, # Every 10% of budget used
            'user_request': True       # Manual checkpoints
        }
    
    def maybe_create_checkpoint(self, trigger_type):
        if self.should_checkpoint(trigger_type):
            checkpoint_uuid = self.create_checkpoint(trigger_type)
            self.last_checkpoint = checkpoint_uuid
            return checkpoint_uuid
    
    def create_checkpoint(self, context):
        return create_task_checkpoint(
            machine_id=self.uuid,
            context=context,
            file_state=self.capture_file_state(),
            memory_state=self.capture_working_memory()
        )
```

### Clean Termination Process
```python
def terminate_machine(machine_uuid, reason="completed"):
    machine = db.get_machine(machine_uuid)
    
    # 1. Create final checkpoint
    final_checkpoint = create_final_checkpoint(machine_uuid, reason)
    
    # 2. Resource cleanup and reporting
    resource_usage = calculate_final_resource_usage(machine_uuid)
    db.update_machine_resource_usage(machine_uuid, resource_usage)
    
    # 3. VM shutdown and cleanup
    proxmox.shutdown_vm(machine.vm_id)
    proxmox.cleanup_vm_resources(machine.vm_id)
    
    # 4. Update machine status
    db.update_machine_status(machine_uuid, "terminated", final_checkpoint)
    
    # 5. Notify spawning Hub
    notify_hub_of_completion(machine.spawned_by, machine_uuid, final_checkpoint)
```

## Communication and Permissions

### Hub-to-Machine Communication
- **Spawning**: Hub creates machines via Proxmox API
- **Monitoring**: Hub queries machine status via VM metrics
- **Termination**: Hub can request clean shutdown or force termination
- **Checkpointing**: Hub can trigger manual checkpoints

### Machine-to-Hub Communication  
- **Status Updates**: Machines report progress, resource usage, errors
- **Resource Requests**: Machines can request additional resources
- **Checkpoint Notifications**: Machines report checkpoint creation
- **Completion Reports**: Machines report final results

### Machine-to-Machine Communication (Post-MVP)
- **File Exchange**: Machines can request files from each other's checkpoints
- **Conversation Sharing**: Access to other machines' conversation history
- **Collaborative Checkpointing**: Coordinated checkpoints across related machines

### Permission Model
```python
# Machine permission matrix
MACHINE_PERMISSIONS = {
    'task_machines': {
        'proxmox_api': [],  # No direct Proxmox access
        'hub_api': ['status_update', 'resource_request', 'checkpoint_notify'],
        'file_system': ['read', 'write'],  # Within VM only
        'network': ['api_endpoints_only']  # No internet, only specific APIs
    },
    
    'hub_machines': {
        'proxmox_api': ['vm_create', 'vm_control', 'snapshot_manage'],
        'hub_api': ['full_access'],
        'file_system': ['read', 'write'],
        'network': ['api_access', 'proxmox_management']
    }
}
```

## Resource Management Integration

### Resource Tracking
Every machine continuously tracks:
- **Money**: API costs, VM runtime costs, storage costs
- **Time**: Execution duration, idle time, queue time  
- **Compute**: CPU usage, memory consumption, disk I/O
- **API Calls**: Service-specific quotas and rate limits

### Budget Enforcement
```python
def check_resource_limits(machine_uuid, operation_type):
    machine = db.get_machine(machine_uuid)
    project = db.get_project(machine.project_uuid)
    
    current_usage = get_current_resource_usage(machine_uuid)
    project_budget = get_project_resource_budget(machine.project_uuid)
    
    # Check multiple resource types
    for resource_type in ['money', 'time', 'api_calls', 'compute']:
        used = current_usage[resource_type]
        allocated = project_budget[resource_type] * project.greediness
        
        if used + estimate_operation_cost(operation_type, resource_type) > allocated:
            raise ResourceLimitExceededError(resource_type, used, allocated)
    
    return True
```

### Machine Selection Algorithm
```python
def select_optimal_machine(prompt, project_uuid, constraints={}):
    # Analyze prompt requirements
    requirements = analyze_prompt_requirements(prompt)
    
    # Get available machine types
    available_machines = get_machine_types_within_budget(project_uuid)
    
    # Score machines by capability match and efficiency
    scored_machines = []
    for machine_type in available_machines:
        capability_score = calculate_capability_match(machine_type, requirements)
        efficiency_score = get_historical_efficiency(machine_type, requirements)
        budget_efficiency = machine_type.cost_per_unit / machine_type.quality_score
        
        total_score = (
            capability_score * 0.4 +
            efficiency_score * 0.4 + 
            budget_efficiency * 0.2
        )
        
        scored_machines.append((machine_type, total_score))
    
    # Return highest scoring machine
    return max(scored_machines, key=lambda x: x[1])[0]
```

## Machine Templates and Configuration

### VM Templates
Each machine type has optimized VM templates:
- **Base Image**: Ubuntu/Debian with common tools
- **Machine-Specific Tools**: Claude API clients, development tools, specialized software
- **Resource Profiles**: CPU, memory, disk sized for machine type
- **Network Configuration**: Appropriate access permissions

### Configuration Management
```python
# Machine configuration profiles
MACHINE_CONFIGS = {
    'claude_code': {
        'vm_template': 'ubuntu-dev-template',
        'cpu_cores': 2,
        'memory_gb': 4,
        'disk_gb': 20,
        'network_profile': 'api_only',
        'installed_tools': ['claude-api', 'git', 'docker', 'node', 'python'],
        'checkpoint_frequency': 'high'  # After each API call
    },
    
    'hub': {
        'vm_template': 'ubuntu-server-template', 
        'cpu_cores': 4,
        'memory_gb': 8,
        'disk_gb': 50,
        'network_profile': 'management',
        'installed_tools': ['docker', 'postgresql', 'nginx', 'proxmox-client'],
        'checkpoint_frequency': 'daily'  # Daily automated
    }
}
```

This unified architecture ensures all components follow the same patterns while allowing specialization for different roles and capabilities.