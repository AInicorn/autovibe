# Proxmox Integration Architecture

## Overview

Proxmox serves as the foundational infrastructure layer for autovibe, providing VM orchestration, snapshot management, and resource control. The Hub communicates with Proxmox APIs to manage the entire Intelligent Machine lifecycle.

## Proxmox API Integration

### Authentication and Permissions
The ProxmoxClient class handles authentication and API communication with Proxmox servers. It initializes with host credentials, authenticates by sending username and password to the Proxmox API ticket endpoint, extracts the API token from successful responses, and establishes an authenticated session. Authentication failures raise ProxmoxAuthenticationError exceptions.

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
VM creation from templates involves a two-step process: First, clone a VM from the specified template using Proxmox API, generating a new VM ID and specifying target node, storage location, and full clone parameters. Second, configure the new VM's resources including CPU cores, memory allocation, network interface with virtio bridge and firewall, and storage disk configuration. The function returns the newly created VM ID.

### VM Template Management
VM template configurations define resource specifications and software packages for different machine types:

**Claude Code Template (ID 100)**: 2 CPU cores, 4GB memory, 20GB disk with API access network configuration. Includes Ubuntu 22.04 base, Docker, Python 3.11, Node.js, Claude API client, Git, and autovibe-agent.

**Hub Template (ID 101)**: 4 CPU cores, 8GB memory, 50GB disk with management access network configuration. Includes Ubuntu 22.04 server, Docker with Docker Compose, PostgreSQL client, Nginx, Proxmox VE client, and autovibe-hub software.

## Snapshot and Checkpoint Integration

### Snapshot Creation
VM snapshot creation initiates a Proxmox API call with snapshot parameters including name, description, and VM state inclusion (with RAM for full checkpoints). The system monitors snapshot creation progress using the returned task ID, waits for completion, verifies the snapshot exists in the VM's snapshot list, and returns snapshot metadata including ID, VM ID, creation timestamp, and size. Failed snapshot creation raises a SnapshotCreationError.

### Snapshot Restoration
VM snapshot restoration operates in two modes: If a new VM ID is provided, the system creates a new VM by cloning from the snapshot with parameters including target node and full clone settings (branching). If no new VM ID is provided, the system restores the existing VM to the snapshot state using Proxmox rollback API, monitors task completion, and returns the appropriate VM ID.

### Snapshot Management
Snapshot cleanup applies retention policies by iterating through all autovibe VMs and their snapshots. The retention policy evaluation considers snapshot age and naming conventions: recent snapshots are kept based on daily retention days, weekly snapshots (ending with '-weekly') are kept for the weekly retention period, monthly snapshots (ending with '-monthly') are kept for the monthly retention period, and manually marked important snapshots are always preserved. Snapshots meeting deletion criteria are removed and logged.

## Resource Monitoring and Control

### VM Resource Monitoring
VM resource monitoring retrieves real-time usage statistics via Proxmox API including CPU usage percentage, memory consumption, disk I/O statistics, network traffic, uptime, and VM status. Resource limit monitoring compares current usage against configured limits (default 90% for CPU and memory, 24 hours for maximum runtime) and returns a list of violations including CPU exceeded, memory exceeded, or maximum runtime exceeded.

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