# Proxmox Integration Architecture

## Overview

Proxmox serves as the foundational infrastructure layer for autovibe, providing VM orchestration, snapshot management, and resource control. The Hub communicates with Proxmox APIs to manage the entire Intelligent Machine lifecycle.

## Proxmox API Integration

### Authentication and Permissions
```python
class ProxmoxClient:
    def __init__(self, host, username, password, realm='pve'):
        self.host = host
        self.username = username
        self.realm = realm
        self.api_token = None
        self.session = self._authenticate(password)
    
    def _authenticate(self, password):
        """Authenticate with Proxmox and obtain API ticket"""
        auth_data = {
            'username': f"{self.username}@{self.realm}",
            'password': password
        }
        
        response = requests.post(
            f"https://{self.host}:8006/api2/json/access/ticket",
            data=auth_data,
            verify=False  # SSL cert verification config
        )
        
        if response.status_code == 200:
            ticket_data = response.json()['data']
            self.api_token = ticket_data['ticket']
            return requests.Session()
        else:
            raise ProxmoxAuthenticationError("Failed to authenticate with Proxmox")
```

### Required Proxmox Permissions
Hub VM user needs specific permissions:
- **VM.Allocate**: Create new VMs
- **VM.Config**: Modify VM configuration
- **VM.Console**: Access VM console (for debugging)
- **VM.PowerMgmt**: Start, stop, restart VMs
- **VM.Snapshot**: Create, delete, restore snapshots
- **Datastore.AllocateSpace**: Allocate storage for VMs and snapshots
- **Datastore.Audit**: Read storage usage information

## VM Lifecycle Management

### VM Creation from Templates
```python
def create_vm_from_template(template_id, vm_config):
    """Create new VM by cloning from template"""
    
    clone_params = {
        'newid': get_next_available_vmid(),
        'name': vm_config['name'],
        'target': vm_config.get('target_node', 'proxmox-node-1'),
        'full': 1,  # Full clone, not linked clone
        'storage': vm_config.get('storage', 'local-lvm')
    }
    
    # Clone VM from template
    response = proxmox_api.post(
        f"/nodes/{clone_params['target']}/qemu/{template_id}/clone",
        data=clone_params
    )
    
    new_vm_id = clone_params['newid']
    
    # Configure VM resources
    vm_resources = {
        'cores': vm_config['cpu_cores'],
        'memory': vm_config['memory_mb'],
        'net0': f"virtio,bridge=vmbr0,firewall=1",
        'scsi0': f"{vm_config['storage']}:{vm_config['disk_gb']}"
    }
    
    proxmox_api.put(
        f"/nodes/{clone_params['target']}/qemu/{new_vm_id}/config",
        data=vm_resources
    )
    
    return new_vm_id
```

### VM Template Management
```python
# VM Templates for different machine types
VM_TEMPLATES = {
    'claude_code': {
        'template_id': 100,
        'cpu_cores': 2,
        'memory_mb': 4096,
        'disk_gb': 20,
        'network_config': 'api_access',
        'installed_software': [
            'ubuntu-22.04-base',
            'docker',
            'python3.11',
            'nodejs',
            'claude-api-client',
            'git',
            'autovibe-agent'
        ]
    },
    
    'hub': {
        'template_id': 101,
        'cpu_cores': 4,
        'memory_mb': 8192,
        'disk_gb': 50,
        'network_config': 'management_access',
        'installed_software': [
            'ubuntu-22.04-server',
            'docker',
            'docker-compose',
            'postgresql-client',
            'nginx',
            'proxmox-ve-client',
            'autovibe-hub'
        ]
    }
}
```

## Snapshot and Checkpoint Integration

### Snapshot Creation
```python
def create_vm_snapshot(vm_id, snapshot_name, description=""):
    """Create VM snapshot via Proxmox API"""
    
    snapshot_params = {
        'snapname': snapshot_name,
        'description': description,
        'vmstate': 1,  # Include RAM state for full checkpoint
        'include_ram': True
    }
    
    # Initiate snapshot creation
    response = proxmox_api.post(
        f"/nodes/{get_vm_node(vm_id)}/qemu/{vm_id}/snapshot",
        data=snapshot_params
    )
    
    # Monitor snapshot creation progress
    task_id = response.json()['data']
    wait_for_task_completion(task_id)
    
    # Verify snapshot exists
    snapshots = list_vm_snapshots(vm_id)
    if snapshot_name in [s['name'] for s in snapshots]:
        return {
            'snapshot_id': snapshot_name,
            'vm_id': vm_id,
            'created_at': datetime.utcnow(),
            'size_bytes': get_snapshot_size(vm_id, snapshot_name)
        }
    else:
        raise SnapshotCreationError(f"Failed to create snapshot {snapshot_name}")
```

### Snapshot Restoration
```python
def restore_vm_from_snapshot(vm_id, snapshot_name, new_vm_id=None):
    """Restore VM from snapshot, optionally as new VM"""
    
    if new_vm_id:
        # Create new VM from snapshot (branching)
        clone_params = {
            'newid': new_vm_id,
            'name': f"restored-{vm_id}-{snapshot_name}",
            'snapname': snapshot_name,
            'target': get_vm_node(vm_id),
            'full': 1
        }
        
        response = proxmox_api.post(
            f"/nodes/{get_vm_node(vm_id)}/qemu/{vm_id}/clone",
            data=clone_params
        )
        
        return new_vm_id
    else:
        # Restore existing VM to snapshot state
        response = proxmox_api.post(
            f"/nodes/{get_vm_node(vm_id)}/qemu/{vm_id}/snapshot/{snapshot_name}/rollback"
        )
        
        task_id = response.json()['data']
        wait_for_task_completion(task_id)
        
        return vm_id
```

### Snapshot Management
```python
def cleanup_old_snapshots(retention_policy):
    """Clean up snapshots based on retention policy"""
    
    all_vms = get_all_autovibe_vms()
    
    for vm_id in all_vms:
        snapshots = list_vm_snapshots(vm_id)
        
        # Apply retention policy
        for snapshot in snapshots:
            should_delete = apply_retention_policy(snapshot, retention_policy)
            
            if should_delete:
                delete_vm_snapshot(vm_id, snapshot['name'])
                log_snapshot_deletion(vm_id, snapshot['name'], "retention_policy")

def apply_retention_policy(snapshot, policy):
    """Determine if snapshot should be deleted based on policy"""
    age_days = (datetime.utcnow() - snapshot['created_at']).days
    
    # Keep recent snapshots
    if age_days <= policy['keep_daily_days']:
        return False
    
    # Keep weekly snapshots
    if age_days <= policy['keep_weekly_days'] and snapshot['name'].endswith('-weekly'):
        return False
    
    # Keep monthly snapshots  
    if age_days <= policy['keep_monthly_days'] and snapshot['name'].endswith('-monthly'):
        return False
    
    # Keep manually marked important snapshots
    if snapshot.get('important', False):
        return False
    
    return True
```

## Resource Monitoring and Control

### VM Resource Monitoring
```python
def get_vm_resource_usage(vm_id):
    """Get real-time VM resource usage"""
    
    response = proxmox_api.get(
        f"/nodes/{get_vm_node(vm_id)}/qemu/{vm_id}/status/current"
    )
    
    vm_status = response.json()['data']
    
    return {
        'cpu_usage_percent': vm_status.get('cpu', 0) * 100,
        'memory_usage_bytes': vm_status.get('mem', 0),
        'memory_total_bytes': vm_status.get('maxmem', 0),
        'disk_read_bytes': vm_status.get('diskread', 0),
        'disk_write_bytes': vm_status.get('diskwrite', 0),
        'network_in_bytes': vm_status.get('netin', 0),
        'network_out_bytes': vm_status.get('netout', 0),
        'uptime_seconds': vm_status.get('uptime', 0),
        'status': vm_status.get('status', 'unknown')
    }

def monitor_resource_limits(vm_id, resource_limits):
    """Monitor VM and enforce resource limits"""
    
    current_usage = get_vm_resource_usage(vm_id)
    
    violations = []
    
    # Check CPU limit
    if current_usage['cpu_usage_percent'] > resource_limits.get('max_cpu_percent', 90):
        violations.append('cpu_exceeded')
    
    # Check memory limit
    memory_usage_percent = (current_usage['memory_usage_bytes'] / 
                          current_usage['memory_total_bytes']) * 100
    if memory_usage_percent > resource_limits.get('max_memory_percent', 90):
        violations.append('memory_exceeded')
    
    # Check uptime limit
    if current_usage['uptime_seconds'] > resource_limits.get('max_uptime_seconds', 86400):
        violations.append('max_runtime_exceeded')
    
    return violations
```

### Resource Quota Enforcement
```python
def enforce_vm_resource_quotas(project_uuid):
    """Enforce project-level resource quotas"""
    
    project_vms = get_project_vms(project_uuid)
    project_budget = get_project_resource_budget(project_uuid)
    
    total_usage = {
        'cpu_hours': 0,
        'memory_gb_hours': 0,
        'storage_gb': 0,
        'network_gb': 0
    }
    
    # Calculate total project resource usage
    for vm_id in project_vms:
        vm_usage = get_vm_resource_usage(vm_id)
        uptime_hours = vm_usage['uptime_seconds'] / 3600
        
        total_usage['cpu_hours'] += (vm_usage['cpu_usage_percent'] / 100) * uptime_hours
        total_usage['memory_gb_hours'] += (vm_usage['memory_usage_bytes'] / 1e9) * uptime_hours
        total_usage['network_gb'] += (vm_usage['network_in_bytes'] + vm_usage['network_out_bytes']) / 1e9
    
    # Check against quotas
    for resource_type, used in total_usage.items():
        quota = project_budget.get(f"{resource_type}_quota", float('inf'))
        
        if used > quota:
            # Take enforcement action
            handle_quota_violation(project_uuid, resource_type, used, quota)

def handle_quota_violation(project_uuid, resource_type, used, quota):
    """Handle resource quota violations"""
    
    violation_severity = (used - quota) / quota
    
    if violation_severity < 0.1:  # 10% over quota
        # Soft warning
        log_quota_warning(project_uuid, resource_type, used, quota)
    
    elif violation_severity < 0.5:  # 50% over quota
        # Throttle new VM creation
        set_project_vm_creation_throttle(project_uuid, True)
        alert_project_admins(project_uuid, f"Resource quota exceeded: {resource_type}")
    
    else:  # Severe violation
        # Emergency shutdown of least critical VMs
        emergency_shutdown_non_critical_vms(project_uuid)
        alert_system_administrators(f"Emergency quota enforcement: {project_uuid}")
```

## Network Security and Isolation

### Network Configuration
```python
# Network profiles for different machine types
NETWORK_PROFILES = {
    'api_access': {
        'bridge': 'vmbr1',  # Isolated bridge for API-only access
        'firewall': True,
        'allowed_destinations': [
            'api.anthropic.com:443',
            'api.openai.com:443',
            'api.github.com:443',
            'hub.autovibe.local:8000'  # Hub API access
        ],
        'blocked_destinations': ['0.0.0.0/0']  # Block all other destinations
    },
    
    'management_access': {
        'bridge': 'vmbr0',  # Main management bridge
        'firewall': True,
        'allowed_destinations': [
            'proxmox.local:8006',    # Proxmox management
            'postgres.local:5432',   # Database access
            'api.anthropic.com:443', # API access for Hub
            '10.0.0.0/8'            # Internal network
        ]
    }
}

def configure_vm_network(vm_id, network_profile):
    """Configure VM network based on security profile"""
    
    profile = NETWORK_PROFILES[network_profile]
    
    # Configure network interface
    net_config = f"virtio,bridge={profile['bridge']}"
    if profile['firewall']:
        net_config += ",firewall=1"
    
    proxmox_api.put(
        f"/nodes/{get_vm_node(vm_id)}/qemu/{vm_id}/config",
        data={'net0': net_config}
    )
    
    # Apply firewall rules
    apply_firewall_rules(vm_id, profile)
```

This Proxmox integration provides the foundation for secure, scalable VM orchestration with comprehensive snapshot management and resource control.